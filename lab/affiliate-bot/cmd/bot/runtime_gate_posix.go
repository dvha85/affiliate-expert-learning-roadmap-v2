//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// acquireManagedPathLock uses a kernel advisory lock rather than a stale
// marker directory. Closing the descriptor, including implicit close on
// process exit, releases the lock while preserving the harmless control file.
func acquireManagedPathLock(path string) (func(), error) {
	if err := validateManagedLockParent(path); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, errManagedPathLockBusy
		}
		return nil, err
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

// validateManagedLockParent rejects symlinked or non-directory ancestors
// before OpenFile resolves the lock pathname. Missing ancestors are allowed so
// the caller can create a lock beside a not-yet-created output, but the first
// existing ancestor must still be a real directory. This is a bounded local
// preflight; a concurrent ancestor replacement still needs OS-specific
// directory-handle/openat hardening and is outside this helper.
func validateManagedLockParent(path string) error {
	for candidate := filepath.Dir(filepath.Clean(path)); ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return errors.New("managed lock parent contains a symlink or non-directory")
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(candidate)
		if next == candidate {
			return errors.New("managed lock parent has no existing directory ancestor")
		}
	}
}
