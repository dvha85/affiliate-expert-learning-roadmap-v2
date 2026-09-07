//go:build linux || darwin

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCampaignPathGuard(t *testing.T) {
	// Resolve the OS temp root (macOS /var is a symlink) before testing the
	// strict application path; no production path is normalized this way.
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	parent := filepath.Join(root, "app")
	path := filepath.Join(parent, "deepseek-br11-v1")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := validateCampaignPath(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0755); err != nil {
		t.Fatal(err)
	}
	if err := validateCampaignPath(path); err == nil {
		t.Fatal("public app dir accepted")
	}
	if err := os.Chmod(parent, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(parent, alias); err != nil {
		t.Fatal(err)
	}
	if err := validateCampaignPath(filepath.Join(alias, "deepseek-br11-v1")); err == nil {
		t.Fatal("ancestor symlink accepted")
	}
	if err := validateCampaignPath("relative/deepseek-br11-v1"); err == nil {
		t.Fatal("relative accepted")
	}
	if err := validateCampaignPath(filepath.Join(parent, "deepseek-missing")); err == nil {
		t.Fatal("missing accepted")
	}
}
