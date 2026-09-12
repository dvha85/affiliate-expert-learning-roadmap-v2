package store

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONLDoesNotAcknowledgeBeforeSyncBoundary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	appendLineFault = func(phase string) error {
		if phase == "after_write" {
			return errors.New("injected before-sync interruption")
		}
		return nil
	}
	t.Cleanup(func() { appendLineFault = nil })
	if err := (JSONL{}).AppendLine(path, []byte(`{"id":"one"}`)); err == nil {
		t.Fatal("pre-sync fault was acknowledged as success")
	}
	appendLineFault = nil
	if err := (JSONL{}).AppendLine(path, []byte(`{"id":"two"}`)); err != nil {
		t.Fatalf("append after failed acknowledgement did not retry: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || !bytes.HasSuffix(raw, []byte("\n")) || !bytes.Contains(raw, []byte(`{"id":"two"}`)) {
		t.Fatalf("post-failure JSONL framing invalid: %q err=%v", raw, err)
	}
}

func TestJSONLAppendAndRead(t *testing.T) {
	p := filepath.Join(t.TempDir(), "history.jsonl")
	s := JSONL{}
	record := []byte(`{"id":"one"}`)
	if e := s.AppendLine(p, record); e != nil {
		t.Fatal(e)
	}
	if e := s.AppendLine(p, []byte(`{"id":"two"}`)); e != nil {
		t.Fatal(e)
	}
	f, e := s.Open(p)
	if e != nil {
		t.Fatal(e)
	}
	b, e := io.ReadAll(f)
	f.Close()
	if e != nil {
		t.Fatal(e)
	}
	if string(b) != "{\"id\":\"one\"}\n{\"id\":\"two\"}\n" || string(record) != `{"id":"one"}` {
		t.Fatal(string(b))
	}
	for _, bad := range [][]byte{nil, []byte("{}\n{}"), []byte("{}\r"), bytes.Repeat([]byte("x"), MaxHistoryRecordBytes+1)} {
		if e = s.AppendLine(p, bad); e == nil {
			t.Fatal("accepted invalid frame")
		}
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(b, after) {
		t.Fatal("rejected frame changed file")
	}
}

func TestJSONLOpenRejectsSameByteSymlinkReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	external := filepath.Join(dir, "external-history.jsonl")
	original := []byte(`{"id":"same-byte"}` + "\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, original, 0600); err != nil {
		t.Fatal(err)
	}
	openPathHook = func(got string) error {
		if got != path {
			t.Fatalf("hook path = %q, want %q", got, path)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { openPathHook = nil })
	if _, err := (JSONL{}).Open(path); err == nil {
		t.Fatal("reader followed a same-byte symlink replacement")
	}
	after, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatalf("reader mutated external target: %q", after)
	}
}

func TestJSONLReaderReportsSameByteReplacementAfterOpen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	external := filepath.Join(dir, "external-history.jsonl")
	original := []byte(`{"id":"same-byte"}` + "\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, original, 0600); err != nil {
		t.Fatal(err)
	}
	readPathHook = func(got string) error {
		if got != path {
			t.Fatalf("hook path = %q, want %q", got, path)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { readPathHook = nil })
	reader, err := (JSONL{}).Open(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, original) {
		t.Fatalf("descriptor bytes = %q, want %q", data, original)
	}
	if err := reader.Close(); err == nil {
		t.Fatal("post-open same-byte replacement was acknowledged")
	}
}

func TestOversizedRecordDoesNotCreateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if err := (JSONL{}).AppendLine(path, bytes.Repeat([]byte("x"), MaxHistoryRecordBytes+1)); err == nil {
		t.Fatal("oversize accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("file created", err)
	}
}

func TestJSONLDoesNotCreateParentOrTruncate(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "missing", "history.jsonl")
	s := JSONL{}
	if e := s.AppendLine(p, []byte(`{}`)); e == nil {
		t.Fatal("created parent")
	}
	if _, e := os.Stat(filepath.Dir(p)); !os.IsNotExist(e) {
		t.Fatal(e)
	}
	block := filepath.Join(dir, "block")
	if e := os.WriteFile(block, []byte("keep"), 0600); e != nil {
		t.Fatal(e)
	}
	if e := s.AppendLine(filepath.Join(block, "history.jsonl"), []byte(`{}`)); e == nil {
		t.Fatal("accepted invalid path")
	}
	b, _ := os.ReadFile(block)
	if string(b) != "keep" {
		t.Fatal("truncated blocker")
	}
}

func TestJSONLAppendRejectsSameByteSymlinkReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	external := filepath.Join(dir, "external-history.jsonl")
	original := []byte(`{"id":"same-byte"}` + "\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(external, original, 0600); err != nil {
		t.Fatal(err)
	}
	appendLinePathHook = func(got string) error {
		if got != path {
			t.Fatalf("hook path = %q, want %q", got, path)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { appendLinePathHook = nil })
	if err := (JSONL{}).AppendLine(path, []byte(`{"id":"must-not-append"}`)); err == nil {
		t.Fatal("append followed a same-byte symlink replacement")
	}
	after, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatalf("external target changed: %q", after)
	}
}

func TestJSONLAppendRejectsSymlinkCreatedAfterAbsentCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.jsonl")
	external := filepath.Join(dir, "external-history.jsonl")
	original := []byte(`{"id":"external"}` + "\n")
	if err := os.WriteFile(external, original, 0600); err != nil {
		t.Fatal(err)
	}
	appendLinePathHook = func(got string) error {
		if got != path {
			t.Fatalf("hook path = %q, want %q", got, path)
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { appendLinePathHook = nil })
	if err := (JSONL{}).AppendLine(path, []byte(`{"id":"must-not-create"}`)); err == nil {
		t.Fatal("append followed a symlink created after the absent check")
	}
	after, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatalf("external target changed: %q", after)
	}
}
