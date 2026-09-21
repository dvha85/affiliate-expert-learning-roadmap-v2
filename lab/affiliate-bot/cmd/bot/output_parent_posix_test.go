//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOutputParentOpenPathResolvesOnlyMacOSSystemVarAlias(t *testing.T) {
	input := filepath.Join(string(filepath.Separator), "var", "folders", "fixture")
	got := outputParentOpenPath(input)
	if runtime.GOOS == "darwin" {
		want := filepath.Join(string(filepath.Separator), "private", "var", "folders", "fixture")
		if got != want {
			t.Fatalf("macOS /var alias was not canonicalized: got %q want %q", got, want)
		}
		return
	}
	if got != input {
		t.Fatalf("non-macOS path was unexpectedly rewritten: got %q want %q", got, input)
	}
}

func TestEnsureOutputParentClosesTraversalDescriptors(t *testing.T) {
	countFDs := func() int {
		entries, err := os.ReadDir("/dev/fd")
		if err != nil {
			t.Skipf("descriptor inventory unavailable: %v", err)
		}
		return len(entries)
	}
	before := countFDs()
	for i := 0; i < 20; i++ {
		parent := filepath.Join(t.TempDir(), "one", "two", "three")
		if err := ensureOutputParentPlatform(parent); err != nil {
			t.Fatal(err)
		}
	}
	after := countFDs()
	if after > before+2 {
		t.Fatalf("output-parent traversal leaked descriptors: before=%d after=%d", before, after)
	}
}
