package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func assertAdvisorAppendRejectsParentSwapBeforeOpenat(t *testing.T, parent, target string, write func() error) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("descriptor-pinned parent traversal is POSIX-only")
	}
	external := t.TempDir()
	sentinel := filepath.Join(external, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	moved := parent + "-moved"
	stableRegularFileAppendHook = func(path string) error {
		if filepath.Clean(path) != filepath.Clean(target) {
			return nil
		}
		if err := os.Rename(parent, moved); err != nil {
			return err
		}
		return os.Symlink(external, parent)
	}
	t.Cleanup(func() { stableRegularFileAppendHook = nil })
	if err := write(); err == nil {
		t.Fatal("parent replacement reached an append target")
	}
	after, err := os.ReadFile(sentinel)
	if err != nil || !bytes.Equal(after, []byte("keep\n")) {
		t.Fatalf("external directory was changed: err=%v contents=%q", err, after)
	}
	if _, err := os.Stat(filepath.Join(external, filepath.Base(target))); !os.IsNotExist(err) {
		t.Fatalf("append target was created below the external parent: err=%v", err)
	}
}

func TestAdvisorCampaignInitRejectsParentSwapBeforeManifest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("descriptor-pinned parent traversal is POSIX-only")
	}
	root := t.TempDir()
	parent := filepath.Join(root, "campaign")
	external := t.TempDir()
	sentinel := filepath.Join(external, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(parent, "manifest")
	moved := parent + "-moved"
	stableRegularFileAppendHook = func(path string) error {
		if filepath.Clean(path) != filepath.Clean(target) {
			return nil
		}
		if err := os.Rename(parent, moved); err != nil {
			return err
		}
		return os.Symlink(external, parent)
	}
	t.Cleanup(func() { stableRegularFileAppendHook = nil })
	if err := initAdvisorCampaign(parent); err == nil {
		t.Fatal("parent replacement reached the campaign manifest")
	}
	if _, err := os.Stat(filepath.Join(external, "manifest")); !os.IsNotExist(err) {
		t.Fatalf("manifest was created below the external parent: err=%v", err)
	}
	after, err := os.ReadFile(sentinel)
	if err != nil || !bytes.Equal(after, []byte("keep\n")) {
		t.Fatalf("external directory was changed: err=%v contents=%q", err, after)
	}
}

func TestAdvisorReservationRejectsParentSwapBeforeOpenat(t *testing.T) {
	parent := t.TempDir()
	if err := initAdvisorCampaign(parent); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(parent, "attempt-001.json")
	assertAdvisorAppendRejectsParentSwapBeforeOpenat(t, parent, target, func() error {
		_, err := reserveAdvisorAttempt(parent)
		return err
	})
}

func TestAdvisorResultRejectsParentSwapBeforeOpenat(t *testing.T) {
	parent := t.TempDir()
	if err := initAdvisorCampaign(parent); err != nil {
		t.Fatal(err)
	}
	if attempt, err := reserveAdvisorAttempt(parent); err != nil || attempt != 1 {
		t.Fatalf("reserve campaign attempt: attempt=%d err=%v", attempt, err)
	}
	candidate := campaignResult{
		Version:       "br11-result/v1",
		Attempt:       1,
		Provider:      (mockAdvisorProvider{}).identity(),
		ContextSHA256: campaignContextDigest(),
		Status:        "ABSTAIN",
		PriceVersion:  campaignPriceVersion,
	}
	target := filepath.Join(parent, "result-001.json")
	assertAdvisorAppendRejectsParentSwapBeforeOpenat(t, parent, target, func() error {
		return persistCampaignResult(parent, candidate)
	})
}

func TestAdvisorFixtureWriterRejectsParentSwapBeforeOpenat(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "fixture.json")
	assertAdvisorAppendRejectsParentSwapBeforeOpenat(t, parent, target, func() error {
		return writeFixtureJSON(target, map[string]string{"status": "fixture"})
	})
}

func TestBackupStagingWriterRejectsParentSwapBeforeOpenat(t *testing.T) {
	parent := t.TempDir()
	target := filepath.Join(parent, "staging.json")
	assertAdvisorAppendRejectsParentSwapBeforeOpenat(t, parent, target, func() error {
		return writeBackupStagingFile(parent, target, []byte("staging"))
	})
}
