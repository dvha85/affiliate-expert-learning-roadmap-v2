//go:build linux || darwin

package main

import (
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
