//go:build linux || darwin

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestM11JournalFIFOFailsClosedBeforeRuntimeRead(t *testing.T) {
	for name, journalPath := range map[string]func(string) string{
		"failed":  m11FailedExecutionJournalPath,
		"unknown": m11UnknownStopJournalPath,
		"outcome": m11OutcomeJournalPath,
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
				t.Fatalf("init failed: code=%d response=%+v", code, response)
			}
			path := journalPath(dir)
			if err := syscall.Mkfifo(path, 0600); err != nil {
				t.Fatal(err)
			}
			if err := m11JournalRecoveryRequired(dir); err == nil || !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("FIFO journal was not rejected before recovery: %v", err)
			}
			if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed for FIFO: code=%d response=%+v", code, response)
			}
			if _, err := readM11Journal(filepath.Clean(path)); err == nil || !strings.Contains(err.Error(), "not a regular file") {
				t.Fatalf("FIFO journal reached reader: %v", err)
			}
		})
	}
}

func TestM10ExecutionJournalFIFOFailsClosedBeforeRuntimeRead(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	path := m10ExecutionJournalPath(dir)
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m10ExecutionJournalRecoveryRequired(dir); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("FIFO M10 journal was not rejected before recovery: %v", err)
	}
	if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("status did not fail closed for M10 FIFO: code=%d response=%+v", code, response)
	}
	if _, err := readM10ExecutionJournal(filepath.Clean(path)); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("FIFO M10 journal reached reader: %v", err)
	}
}

// The recovery-presence check intentionally only needs Lstat: it fails closed
// as soon as a journal exists. The parser, however, must not follow a same-byte
// replacement after its Lstat/open check, because a journal is an authority
// replay plan. Exercise the real M10/M11 readers with the stable-reader seam.
func TestRecoveryJournalReadersRejectSymlinkSwapAfterOpen(t *testing.T) {
	readers := map[string]func(string) ([]byte, error){
		"m10": readM10ExecutionJournal,
		"m11": readM11Journal,
	}
	for name, reader := range readers {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "recovery-journal.json")
			contents := []byte(`{"version":"fixture-journal/v1"}`)
			if err := os.WriteFile(path, contents, 0600); err != nil {
				t.Fatal(err)
			}
			external := filepath.Join(dir, "same-byte-external.json")
			if err := os.WriteFile(external, contents, 0600); err != nil {
				t.Fatal(err)
			}
			swapped := false
			stableRegularFileReadHook = func(openedPath string) error {
				if filepath.Clean(openedPath) != filepath.Clean(path) || swapped {
					return nil
				}
				swapped = true
				if err := os.Remove(path); err != nil {
					return err
				}
				return os.Symlink(external, path)
			}
			t.Cleanup(func() { stableRegularFileReadHook = nil })

			if _, err := reader(path); err == nil || !strings.Contains(err.Error(), "changed while reading stable regular file") {
				t.Fatalf("recovery journal reader accepted a same-byte symlink swap: %v", err)
			}
			if !swapped {
				t.Fatal("recovery journal reader did not reach stable-reader swap seam")
			}
		})
	}
}
