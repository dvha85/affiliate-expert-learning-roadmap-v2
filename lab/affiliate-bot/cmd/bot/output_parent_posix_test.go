//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
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
