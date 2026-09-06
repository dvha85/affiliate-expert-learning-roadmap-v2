package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

type faultHistory struct {
	raw               []byte
	readErr, writeErr error
	writes            int
}

func (s *faultHistory) Open(string) (io.ReadCloser, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	return io.NopCloser(bytes.NewReader(s.raw)), nil
}
func (s *faultHistory) AppendLine(_ string, b []byte) error {
	s.writes++
	if s.writeErr != nil {
		return s.writeErr
	}
	s.raw = append(s.raw, append(b, '\n')...)
	return nil
}

func TestHistoryStoreSeamFailClosed(t *testing.T) {
	r, e := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o1", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		name    string
		raw     []byte
		readErr error
	}{{"read", []byte("keep"), errors.New("read failed")}, {"corrupt", []byte("{}\n"), nil}} {
		s := &faultHistory{raw: append([]byte(nil), tc.raw...), readErr: tc.readErr}
		if status, e := appendHistoryWith(s, "explicit", r); e == nil || status != "" {
			t.Fatal(tc.name, status, e)
		}
		if s.writes != 0 || !bytes.Equal(s.raw, tc.raw) {
			t.Fatal("reset on read failure")
		}
	}
	s := &faultHistory{readErr: os.ErrNotExist, writeErr: errors.New("write failed")}
	if status, e := appendHistoryWith(s, "explicit", r); e == nil || status != "" || s.writes != 1 {
		t.Fatal(status, e)
	}
	s = &faultHistory{}
	if status, e := appendHistoryWith(s, "explicit", r); e != nil || status != appendAdded {
		t.Fatal(status, e)
	}
	before := append([]byte(nil), s.raw...)
	if status, e := appendHistoryWith(s, "explicit", r); e != nil || status != appendDuplicate || s.writes != 1 {
		t.Fatal(status, e)
	}
	loaded, e := loadHistoryWith(s, "explicit")
	if e != nil || len(loaded) != 1 || Replay(loaded[0]).State != replayMatch {
		t.Fatal(loaded, e)
	}
	r.RecordedResult.Reasons = append(r.RecordedResult.Reasons, "conflict")
	if _, e := appendHistoryWith(s, "explicit", r); e == nil {
		t.Fatal("conflict accepted")
	}
	if !bytes.Equal(before, s.raw) || s.writes != 1 {
		t.Fatal("conflict modified history")
	}
	// Existing validated prefix must survive an injected append failure.
	loaded[0].RecordID = "r2"
	s.writeErr = errors.New("append failed")
	if status, e := appendHistoryWith(s, "explicit", loaded[0]); e == nil || status != "" {
		t.Fatal(status, e)
	}
	if !bytes.Equal(before, s.raw) {
		t.Fatal("lost prefix")
	}
}
