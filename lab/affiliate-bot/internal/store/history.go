// Package store owns local filesystem access, not history validation or authority.
package store

import (
	"bytes"
	"fmt"
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

func appendLineFailure(phase string) error {
	if appendLineFault == nil {
		return nil
	}
	return appendLineFault(phase)
}

// MaxHistoryRecordBytes is the JSON payload limit, excluding LF/CRLF framing.
// Reader and writer share this bound; rejection occurs before opening a file.
const MaxHistoryRecordBytes = 1 << 20

type stableJSONLReader struct {
	file   *os.File
	path   string
	opened fs.FileInfo
}

func (r *stableJSONLReader) Read(p []byte) (int, error) { return r.file.Read(p) }

func (r *stableJSONLReader) Close() error {
	closeErr := r.file.Close()
	after, err := os.Lstat(r.path)
	if err != nil {
		return err
	}
	if !after.Mode().IsRegular() || !os.SameFile(r.opened, after) {
		return fmt.Errorf("history path changed while reading")
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
	return &stableJSONLReader{file: f, path: path, opened: opened}, nil
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
