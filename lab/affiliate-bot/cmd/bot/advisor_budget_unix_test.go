//go:build linux || darwin

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestCampaignRejectsFIFO(t *testing.T) {
	for _, name := range []string{"manifest", "attempt-001.json"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "campaign")
			if err := initAdvisorCampaign(path); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(path, name)
			if name == "manifest" {
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
			}
			if err := syscall.Mkfifo(target, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := reserveAdvisorAttempt(path); err == nil {
				t.Fatal("FIFO accepted")
			}
		})
	}
}
