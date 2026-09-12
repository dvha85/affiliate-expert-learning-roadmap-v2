// Package store owns local filesystem access, not history validation or authority.
package store

import (
	"bytes"
	"fmt"
	"io"
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

func appendLineFailure(phase string) error {
	if appendLineFault == nil {
		return nil
	}
	return appendLineFault(phase)
}

// MaxHistoryRecordBytes is the JSON payload limit, excluding LF/CRLF framing.
// Reader and writer share this bound; rejection occurs before opening a file.
const MaxHistoryRecordBytes = 1 << 20

func (JSONL) Open(path string) (io.ReadCloser, error) { return os.Open(path) }

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
