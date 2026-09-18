//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// openStableRegularFileForAppendPath opens an append target relative to a
// descriptor-pinned parent directory. The shared caller has already checked
// whether the final name exists; keeping the parent descriptor alive through
// the final open closes the preflight-to-open escape where a missing registry
// could otherwise be created below a replacement symlink directory.
func openStableRegularFileForAppendPath(path string, newFile bool) (*os.File, error) {
	clean := filepath.Clean(path)
	parentPath, name := filepath.Dir(clean), filepath.Base(clean)
	// macOS exposes /var as the system alias /private/var. Resolve that
	// platform-owned alias before descriptor traversal; caller-controlled
	// symlinked ancestors remain rejected by openManagedDirectory.
	parentFD, err := openManagedDirectory(outputParentOpenPath(parentPath))
	if err != nil {
		return nil, fmt.Errorf("open append parent %q: %w", parentPath, err)
	}
	parent := os.NewFile(uintptr(parentFD), parentPath)
	if parent == nil {
		_ = unix.Close(parentFD)
		return nil, errors.New("append parent descriptor could not be wrapped")
	}
	defer parent.Close()
	openedParent, err := parent.Stat()
	if err != nil {
		return nil, err
	}
	if !openedParent.IsDir() {
		return nil, fmt.Errorf("append parent %q is not a directory", parentPath)
	}
	if stableRegularFileAppendHook != nil {
		if err := stableRegularFileAppendHook(path); err != nil {
			return nil, err
		}
	}
	currentParent, err := os.Lstat(parentPath)
	if err != nil {
		return nil, err
	}
	if !currentParent.IsDir() || currentParent.Mode()&os.ModeSymlink != 0 || !os.SameFile(openedParent, currentParent) {
		return nil, fmt.Errorf("append parent %q changed while opening", parentPath)
	}

	flags := unix.O_WRONLY | unix.O_APPEND | unix.O_CLOEXEC | unix.O_NOFOLLOW
	if newFile {
		flags |= unix.O_CREAT | unix.O_EXCL
	}
	fd, err := unix.Openat(int(parent.Fd()), name, flags, 0600)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), clean)
	if f == nil {
		_ = unix.Close(fd)
		return nil, errors.New("append file descriptor could not be wrapped")
	}
	return f, nil
}
