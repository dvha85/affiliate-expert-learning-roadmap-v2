// Package store owns local filesystem access, not history validation or authority.
package store

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// History is the I/O seam. The application validates history and conflicts before
// AppendLine. This is not a concurrent transaction or a trusted state service.
type History interface {
	Open(path string) (io.ReadCloser, error)
	AppendLine(path string, record []byte) error
}

type JSONL struct{}

// appendLineFault is an in-package test seam for the acknowledge-after-sync
// boundary. It is never configurable by callers.
var appendLineFault func(phase string) error

// appendLinePathHook is a test-only seam for replacing a name after its
// pre-open check. It is intentionally package-private: callers cannot choose
// an append target or interfere with the write boundary in production.
var appendLinePathHook func(path string) error

// openPathHook is a test-only seam for replacing a history name after the
// reader has inspected it but before os.Open. Production callers cannot use
// it to select or mutate a source path.
var openPathHook func(path string) error

// readPathHook runs only in store tests after a descriptor has been bound to
// the inspected inode and before that descriptor is returned to a caller.
// The stable reader checks the name again on Close, so this lets a regression
// prove that a post-open replacement is surfaced to callers that complete a
// parse and check Close.
var readPathHook func(path string) error

// portableInputReadHook is an in-package test seam for the generic portable
// reader. Production callers cannot select a post-open mutation point.
var portableInputReadHook func(path string) error

func appendLineFailure(phase string) error {
	if appendLineFault == nil {
		return nil
	}
	return appendLineFault(phase)
}

// MaxHistoryRecordBytes is the JSON payload limit, excluding LF/CRLF framing.
// Reader and writer share this bound; rejection occurs before opening a file.
const MaxHistoryRecordBytes = 1 << 20

// MaxPortableInputBytes bounds caller-supplied learner files while preserving
// the M03/M04 semantic one-MiB store contract, which intentionally tests an
// over-limit record after it has been decoded.
const MaxPortableInputBytes int64 = 16 << 20

// ReadPortableInput reads a caller-supplied file only while its pathname stays
// bound to one regular inode. It also rereads the opened descriptor before
// acknowledgement, so same-inode rewrites cannot silently change the bytes a
// CLI validates. This is a local filesystem guard, not a multi-host snapshot.
func ReadPortableInput(path string) ([]byte, error) {
	return ReadStableRegularFileLimit(path, MaxPortableInputBytes)
}

// ReadStableRegularFileLimit is the bounded byte-oriented counterpart of the
// JSONL reader. It is for portable input rather than canonical history; the
// caller owns semantic decoding and status mapping.
func ReadStableRegularFileLimit(path string, limit int64) ([]byte, error) {
	if limit < -1 {
		return nil, fmt.Errorf("invalid stable regular file limit")
	}
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || limit >= 0 && before.Size() > limit {
		return nil, fmt.Errorf("portable input is not an allowed regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, statErr := f.Stat()
	if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) || opened.Size() != before.Size() || limit >= 0 && opened.Size() > limit {
		_ = f.Close()
		if statErr != nil {
			return nil, statErr
		}
		return nil, fmt.Errorf("portable input changed while opening")
	}
	if portableInputReadHook != nil {
		if err := portableInputReadHook(path); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	read := func() ([]byte, error) {
		reader := io.Reader(f)
		if limit >= 0 {
			reader = io.LimitReader(f, limit+1)
		}
		value, err := io.ReadAll(reader)
		if err != nil || limit >= 0 && int64(len(value)) > limit {
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("portable input exceeds stable regular file limit")
		}
		return value, nil
	}
	first, err := read()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()
		return nil, err
	}
	second, err := read()
	closeErr := f.Close()
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(first, second) {
		return nil, fmt.Errorf("portable input content changed while reading")
	}
	if closeErr != nil {
		return nil, closeErr
	}
	after, err := os.Lstat(path)
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) || after.Size() != opened.Size() || limit >= 0 && after.Size() > limit {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("portable input changed while reading")
	}
	return first, nil
}

type stableJSONLReader struct {
	file       *os.File
	path       string
	opened     fs.FileInfo
	snapshot   hash.Hash
	reachedEOF bool
	bytesRead  int64
	lastByte   byte
}

func (r *stableJSONLReader) Read(p []byte) (int, error) {
	n, err := r.file.Read(p)
	if n > 0 {
		_, _ = r.snapshot.Write(p[:n])
		r.bytesRead += int64(n)
		r.lastByte = p[n-1]
	}
	if err == io.EOF {
		r.reachedEOF = true
		// AppendLine acknowledges only a complete LF-delimited record.  Treat a
		// non-empty file without its final LF as a possibly interrupted append,
		// not as a valid final JSON value that downstream decoders may use.
		if r.bytesRead > 0 && r.lastByte != '\n' {
			return n, fmt.Errorf("JSONL store has incomplete final line framing")
		}
	}
	return n, err
}

// RequireCompleteJSONLFraming applies the same append boundary to JSONL bytes
// that were already read through another stable-file primitive.  Runtime M10
// and M11 fixture stores use this form because they validate an immutable byte
// snapshot before splitting it into records.
func RequireCompleteJSONLFraming(raw []byte) error {
	if len(raw) != 0 && raw[len(raw)-1] != '\n' {
		return fmt.Errorf("JSONL store has incomplete final line framing")
	}
	return nil
}

// verifySnapshot detects an in-place rewrite of the descriptor that supplied
// the fully-consumed JSONL bytes. It deliberately compares the opened
// descriptor, not a second pathname open: a pathname may have been replaced
// and is checked separately below. This is a local consistency check, not an
// atomic filesystem snapshot against an uncooperative writer.
func (r *stableJSONLReader) verifySnapshot() error {
	if !r.reachedEOF {
		return fmt.Errorf("history reader closed before complete snapshot")
	}
	if _, err := r.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	current := sha256.New()
	if _, err := io.Copy(current, r.file); err != nil {
		return err
	}
	if !bytes.Equal(r.snapshot.Sum(nil), current.Sum(nil)) {
		return fmt.Errorf("history content changed while reading")
	}
	return nil
}

func (r *stableJSONLReader) Close() error {
	snapshotErr := r.verifySnapshot()
	closeErr := r.file.Close()
	after, err := os.Lstat(r.path)
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || !os.SameFile(r.opened, after) {
		return fmt.Errorf("history path changed while reading")
	}
	if snapshotErr != nil {
		return snapshotErr
	}
	return closeErr
}

func (JSONL) Open(path string) (io.ReadCloser, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() {
		return nil, fmt.Errorf("history path is not a regular file")
	}
	if openPathHook != nil {
		if err := openPathHook(path); err != nil {
			return nil, err
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, statErr := f.Stat()
	if statErr != nil || !opened.Mode().IsRegular() || !os.SameFile(before, opened) {
		_ = f.Close()
		if statErr != nil {
			return nil, statErr
		}
		return nil, fmt.Errorf("history path changed while opening")
	}
	if readPathHook != nil {
		if err := readPathHook(path); err != nil {
			_ = f.Close()
			return nil, err
		}
	}
	return &stableJSONLReader{file: f, path: path, opened: opened, snapshot: sha256.New()}, nil
}

func (JSONL) AppendLine(path string, record []byte) error {
	if len(record) > MaxHistoryRecordBytes {
		return fmt.Errorf("history record exceeds %d bytes", MaxHistoryRecordBytes)
	}
	if len(record) == 0 || bytes.ContainsAny(record, "\r\n") {
		return fmt.Errorf("history record must be one nonempty JSON line")
	}
	before, e := os.Lstat(path)
	newFile := os.IsNotExist(e)
	if e != nil && !newFile {
		return e
	}
	if !newFile && !before.Mode().IsRegular() {
		return fmt.Errorf("history path is not a regular file")
	}
	if appendLinePathHook != nil {
		if e := appendLinePathHook(path); e != nil {
			return e
		}
	}
	flags := os.O_APPEND | os.O_WRONLY
	if newFile {
		// Do not follow a name that appeared after the absent check.
		flags |= os.O_CREATE | os.O_EXCL
	}
	f, e := os.OpenFile(path, flags, 0600)
	if e != nil {
		return e
	}
	opened, statErr := f.Stat()
	if statErr != nil || !opened.Mode().IsRegular() || (!newFile && !os.SameFile(before, opened)) {
		_ = f.Close()
		if statErr != nil {
			return statErr
		}
		return fmt.Errorf("history path changed while opening for append")
	}
	line := make([]byte, len(record)+1)
	copy(line, record)
	line[len(record)] = '\n'
	n, e := f.Write(line)
	if e != nil {
		_ = f.Close()
		return e
	}
	if n != len(line) {
		_ = f.Close()
		return io.ErrShortWrite
	}
	if e = appendLineFailure("after_write"); e != nil {
		_ = f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		_ = f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	after, e := os.Lstat(path)
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		if e != nil {
			return e
		}
		return fmt.Errorf("history path changed while appending")
	}
	d, e := os.Open(filepath.Dir(path))
	if e != nil {
		return e
	}
	e = d.Sync()
	closeErr := d.Close()
	if e != nil {
		return e
	}
	if closeErr != nil {
		return closeErr
	}
	// Directory sync is part of the acknowledgement boundary. Recheck the
	// name afterwards so callers do not receive success for a replaced path.
	after, e = os.Lstat(path)
	if e != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) {
		if e != nil {
			return e
		}
		return fmt.Errorf("history path changed before append acknowledgement")
	}
	return nil
}
