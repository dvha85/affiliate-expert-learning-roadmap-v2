//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

var managedLockBeforeOpenHook func()

// Windows uses an exclusive native file handle. The kernel releases the
// handle when the owning process exits, so a crashed writer cannot leave a
// stale directory claim that blocks recovery. Ancestor validation remains a
// bounded pathname preflight; full multi-host/distributed locking is out of
// scope for this local runtime gate.
func acquireManagedPathLock(path string) (func(), error) {
	f, err := openWindowsRegularFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, windows.OPEN_ALWAYS)
	if err != nil {
		if errors.Is(err, windows.ERROR_SHARING_VIOLATION) || errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, errManagedPathLockBusy
		}
		return nil, err
	}
	return func() { _ = f.Close() }, nil
}

// openWindowsRegularFile opens a final regular file without following a
// final reparse point. It keeps the existing caller-owned ancestor preflight;
// Windows lacks a direct openat equivalent in this boundary.
func openWindowsRegularFile(path string, access, share, creation uint32) (*os.File, error) {
	clean := filepath.Clean(path)
	if err := validateWindowsPathParent(clean); err != nil {
		return nil, err
	}
	if managedLockBeforeOpenHook != nil && access == windows.GENERIC_READ|windows.GENERIC_WRITE && share == 0 && creation == windows.OPEN_ALWAYS {
		managedLockBeforeOpenHook()
	}
	name, err := windows.UTF16PtrFromString(clean)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(name, access, share, nil, creation, windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		_ = windows.CloseHandle(h)
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("%s is a reparse point", clean)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("%s is a directory", clean)
	}
	f := os.NewFile(uintptr(h), clean)
	if f == nil {
		_ = windows.CloseHandle(h)
		return nil, errors.New("windows file handle could not be wrapped")
	}
	return f, nil
}

func validateWindowsPathParent(path string) error {
	for candidate := filepath.Dir(filepath.Clean(path)); ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("managed path parent contains a symlink or non-directory")
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(candidate)
		if next == candidate {
			return errors.New("managed path parent has no existing directory ancestor")
		}
	}
}
