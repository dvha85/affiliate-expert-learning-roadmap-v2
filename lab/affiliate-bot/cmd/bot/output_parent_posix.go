//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"
)

// outputParentBeforeOpenHook is test-only and runs after the existing-prefix
// preflight but before descriptor traversal. It lets the real backup/restore
// path prove that a swapped ancestor is rejected rather than followed.
var outputParentBeforeOpenHook func()

func ensureOutputParentPlatform(parent string) error {
	existing, missing, err := outputParentExistingPrefix(parent)
	if err != nil {
		return err
	}
	if outputParentBeforeOpenHook != nil {
		outputParentBeforeOpenHook()
	}
	directory, err := openManagedDirectory(outputParentOpenPath(existing))
	if err != nil {
		return fmt.Errorf("open output parent %q: %w", existing, err)
	}
	defer unix.Close(directory)
	for _, component := range missing {
		if component == "" || component == "." {
			continue
		}
		if err := unix.Mkdirat(directory, component, 0700); err != nil && !errors.Is(err, unix.EEXIST) {
			return fmt.Errorf("create output parent component %q: %w", component, err)
		}
		next, err := unix.Openat(directory, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if err != nil {
			return fmt.Errorf("open output parent component %q: %w", component, err)
		}
		_ = unix.Close(directory)
		directory = next
	}
	return nil
}

// macOS exposes /var as a system symlink to /private/var. Resolve only that
// known platform alias for descriptor traversal; caller-selected symlinked
// ancestors are rejected by outputParentExistingPrefix before this point.
func outputParentOpenPath(existing string) string {
	clean := filepath.Clean(existing)
	if runtime.GOOS != "darwin" || (clean != "/var" && !strings.HasPrefix(clean, "/var/")) {
		return clean
	}
	resolved, err := filepath.EvalSymlinks("/var")
	if err != nil || resolved != "/private/var" {
		return clean
	}
	if clean == "/var" {
		return resolved
	}
	return filepath.Join(resolved, strings.TrimPrefix(clean, "/var/"))
}

func outputParentExistingPrefix(parent string) (string, []string, error) {
	missing := []string{}
	for candidate := parent; ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return "", nil, fmt.Errorf("backup/restore target parent contains a symlink or non-directory")
			}
			for left, right := 0, len(missing)-1; left < right; left, right = left+1, right-1 {
				missing[left], missing[right] = missing[right], missing[left]
			}
			return candidate, missing, nil
		}
		if !os.IsNotExist(err) {
			return "", nil, err
		}
		next := filepath.Dir(candidate)
		if next == candidate {
			return "", nil, fmt.Errorf("backup/restore target parent has no existing directory ancestor")
		}
		missing = append(missing, filepath.Base(candidate))
	}
}
