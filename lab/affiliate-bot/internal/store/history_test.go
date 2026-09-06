package store

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

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
