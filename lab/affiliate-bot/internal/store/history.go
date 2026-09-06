// Package store owns local filesystem access, not history validation or authority.
package store

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// History is the I/O seam. The application validates history and conflicts before
// AppendLine. This is not a concurrent transaction or a trusted state service.
type History interface {
	Open(path string) (io.ReadCloser, error)
	AppendLine(path string, record []byte) error
}

type JSONL struct{}

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
	closeErr := f.Close()
	if e != nil {
		return e
	}
	if n != len(line) {
		return io.ErrShortWrite
	}
	return closeErr
}
