//go:build linux || darwin

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFixedCampaignInitNoReset(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := initializeFixedCampaign(root); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "affiliate-expert-learning-roadmap-v2", "deepseek-br11-v1")
	if _, err := reserveAdvisorAttempt(path); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(path, "attempt-001.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := initializeFixedCampaign(root); err == nil {
		t.Fatal("reset accepted")
	}
	after, err := os.ReadFile(filepath.Join(path, "attempt-001.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("reservation changed")
	}
	if code := runAdvisor([]string{"campaign-init", root}, io.Discard, io.Discard); code != 2 {
		t.Fatal(code)
	}
}
func TestFixedCampaignInitUnsafeParent(t *testing.T) {
	for _, kind := range []string{"symlink", "permissions", "partial"} {
		t.Run(kind, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			app := filepath.Join(root, "affiliate-expert-learning-roadmap-v2")
			if kind == "symlink" {
				if err := os.Symlink(root, app); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(app, 0700); err != nil {
					t.Fatal(err)
				}
				if kind == "permissions" {
					if err := os.Chmod(app, 0777); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.Mkdir(filepath.Join(app, "deepseek-br11-v1"), 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := initializeFixedCampaign(root); err == nil {
				t.Fatal("unsafe init accepted")
			}
		})
	}
}
