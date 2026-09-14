package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

type faultHistory struct {
	raw               []byte
	readErr, closeErr error
	writeErr          error
	appendThenError   bool
	writes            int
}

type closeErrorReader struct {
	*bytes.Reader
	err error
}

func (r closeErrorReader) Close() error { return r.err }

func (s *faultHistory) Open(string) (io.ReadCloser, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	if s.closeErr != nil {
		return closeErrorReader{Reader: bytes.NewReader(s.raw), err: s.closeErr}, nil
	}
	return io.NopCloser(bytes.NewReader(s.raw)), nil
}
func (s *faultHistory) AppendLine(_ string, b []byte) error {
	s.writes++
	if s.appendThenError {
		s.raw = append(s.raw, append(b, '\n')...)
		return s.writeErr
	}
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

	// The command-side implementation must distinguish a failed ACK from a
	// failed append when the same canonical record is already replayable.
	visible := &faultHistory{appendThenError: true, writeErr: errors.New("lost acknowledgement")}
	visibleRecord, e := NewHistoryRecord("r-visible", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o-visible", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if e != nil {
		t.Fatal(e)
	}
	if status, e := appendHistoryWith(visible, "visible", visibleRecord); status != appendPublished || !isPublishedAppendUncertainty(e) {
		t.Fatalf("visible append uncertainty was hidden: status=%s err=%v", status, e)
	}
	visibleRecords, e := loadHistoryWith(visible, "visible")
	if e != nil || len(visibleRecords) != 1 || !reflect.DeepEqual(visibleRecords[0], visibleRecord) {
		t.Fatalf("visible history record was not replayable: records=%+v err=%v", visibleRecords, e)
	}
	if status, e := appendHistoryWith(visible, "visible", visibleRecord); e != nil || status != appendDuplicate || visible.writes != 1 {
		t.Fatalf("exact retry after visible uncertainty failed: status=%s err=%v writes=%d", status, e, visible.writes)
	}
}

func TestWatcherHistoryHandoffDisclosesVisibleAppendUncertainty(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	input := filepath.Join(dir, "record.json")
	record, err := NewHistoryRecord("watcher-visible-append", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("watcher-visible-observation", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	canonicalHistoryAppend = func(_ store.History, path string, candidate HistoryRecord) (string, error) {
		if status, err := appendHistoryWith(store.JSONL{}, path, candidate); err != nil || status != appendAdded {
			return status, err
		}
		return appendPublished, &publishedAppendUncertainty{cause: errors.New("injected acknowledgement loss after history append")}
	}
	t.Cleanup(func() { canonicalHistoryAppend = appendHistoryWith })
	var stdout, stderr bytes.Buffer
	if code := runWatcherHistoryHandoff([]string{history, input}, &stdout, &stderr); code == 0 {
		t.Fatalf("visible append uncertainty returned success: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope["status"] != appendPublished || envelope["canonical_history_ack"] != false || envelope["artifact"] == nil {
		t.Fatalf("watcher hid visible append uncertainty: %+v", envelope)
	}
	if records, err := LoadHistory(history); err != nil || len(records) != 1 || !reflect.DeepEqual(records[0], record) {
		t.Fatalf("watcher-visible record was not canonical: records=%+v err=%v", records, err)
	}
	canonicalHistoryAppend = appendHistoryWith
	stdout.Reset()
	stderr.Reset()
	if code := runWatcherHistoryHandoff([]string{history, input}, &stdout, &stderr); code != 0 {
		t.Fatalf("exact retry failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil || envelope["status"] != appendDuplicate || envelope["canonical_history_ack"] != true {
		t.Fatalf("exact retry did not acknowledge canonical record: envelope=%+v err=%v", envelope, err)
	}
}

func TestLoadHistoryPropagatesCloseFailure(t *testing.T) {
	record, err := NewHistoryRecord("r-close", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o-close", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	storage := &faultHistory{raw: append(raw, '\n'), closeErr: errors.New("history changed while reading")}
	if records, err := loadHistoryWith(storage, "history.jsonl"); err == nil || records != nil {
		t.Fatalf("close failure was acknowledged: records=%v err=%v", records, err)
	}
}
