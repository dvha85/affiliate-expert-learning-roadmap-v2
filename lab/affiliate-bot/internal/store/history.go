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
	f, e := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
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
	d, e := os.Open(filepath.Dir(path))
	if e != nil {
		return e
	}
	e = d.Sync()
	closeErr := d.Close()
	if e != nil {
		return e
	}
	return closeErr
}
