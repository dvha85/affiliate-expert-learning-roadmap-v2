//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

var managedLockBeforeOpenHook func()

// acquireManagedPathLock uses a kernel advisory lock rather than a stale
// marker directory. Closing the descriptor, including implicit close on
// process exit, releases the lock while preserving the harmless control file.
func acquireManagedPathLock(path string) (func(), error) {
	if err := validateManagedLockParent(path); err != nil {
		return nil, err
	}
	canonicalPath, err := canonicalManagedLockPath(path)
	if err != nil {
		return nil, err
	}
	if managedLockBeforeOpenHook != nil {
		managedLockBeforeOpenHook()
	}
	f, err := openManagedLockFile(canonicalPath)
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

// canonicalManagedLockPath resolves only the already-existing prefix. macOS
// exposes /var as a system symlink, while missing output directories must stay
// missing until the final O_CREAT/openat operation. The unresolved suffix is
// retained so descriptor traversal still rejects a component swapped to a
// symlink after this resolution and before the open.
func canonicalManagedLockPath(path string) (string, error) {
	clean := filepath.Clean(path)
	parent, name := filepath.Dir(clean), filepath.Base(clean)
	missing := make([]string, 0)
	candidate := parent
	for {
		_, err := os.Lstat(candidate)
		if err == nil {
			resolved, resolveErr := filepath.EvalSymlinks(candidate)
			if resolveErr != nil {
				return "", resolveErr
			}
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return filepath.Join(resolved, name), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		next := filepath.Dir(candidate)
		if next == candidate {
			return "", errors.New("managed lock parent has no existing directory ancestor")
		}
		missing = append(missing, filepath.Base(candidate))
		candidate = next
	}
}

// validateManagedLockParent rejects symlinked or non-directory ancestors
// before the lock pathname is opened. Missing ancestors are allowed so
// the caller can create a lock beside a not-yet-created output, but the first
// existing ancestor must still be a real directory. This is a bounded local
// preflight; concurrent ancestor replacement is handled by the
// descriptor-pinned directory-handle/openat path below on supported POSIX
// systems.
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

// openManagedLockFile pins every existing parent directory before opening the
// final lock entry. A pathname component replaced by a symlink after the
// preflight is therefore rejected by O_NOFOLLOW instead of being resolved
// through an external tree. The parent descriptor is kept until Openat has
// opened the lock entry; this closes the preflight-to-open race for each
// traversed component on supported POSIX systems.
func openManagedLockFile(path string) (*os.File, error) {
	clean := filepath.Clean(path)
	parent, name := filepath.Dir(clean), filepath.Base(clean)
	directory, err := openManagedDirectory(parent)
	if err != nil {
		return nil, fmt.Errorf("open managed lock parent %q: %w", parent, err)
	}
	defer unix.Close(directory)
	fd, err := unix.Openat(directory, name, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, fmt.Errorf("open managed lock entry %q: %w", name, err)
	}
	file := os.NewFile(uintptr(fd), clean)
	if file == nil {
		_ = unix.Close(fd)
		return nil, errors.New("managed lock file descriptor could not be wrapped")
	}
	return file, nil
}

func openManagedDirectory(path string) (int, error) {
	clean := filepath.Clean(path)
	start := "."
	components := strings.Split(clean, string(filepath.Separator))
	if filepath.IsAbs(clean) {
		start = string(filepath.Separator)
		components = strings.Split(strings.TrimPrefix(clean, string(filepath.Separator)), string(filepath.Separator))
	}
	fd, err := unix.Open(start, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, fmt.Errorf("open managed lock traversal root %q: %w", start, err)
	}
	for _, component := range components {
		if component == "" || component == "." {
			continue
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, fmt.Errorf("open managed lock directory component %q: %w", component, openErr)
		}
		fd = next
	}
	return fd, nil
}
