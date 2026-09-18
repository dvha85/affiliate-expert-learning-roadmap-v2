//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// openPinnedWindowsDirectory opens a directory without following the final
// reparse point and without granting delete sharing. The handle therefore
// both identifies the directory we inspected and prevents a concurrent
// rename/replacement while a child pathname is opened or created.
func openPinnedWindowsDirectory(path string) (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(filepath.Clean(path))
	if err != nil {
		return 0, err
	}
	h, err := windows.CreateFile(
		name,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return 0, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		_ = windows.CloseHandle(h)
		return 0, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		_ = windows.CloseHandle(h)
		return 0, fmt.Errorf("%s is a reparse point", path)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		_ = windows.CloseHandle(h)
		return 0, fmt.Errorf("%s is not a directory", path)
	}
	return h, nil
}

func windowsExistingDirectoryPrefix(path string) (string, []string, error) {
	missing := make([]string, 0)
	for candidate := filepath.Clean(path); ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return "", nil, fmt.Errorf("Windows path contains a symlink or non-directory")
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
			return "", nil, errors.New("Windows path has no existing directory ancestor")
		}
		missing = append(missing, filepath.Base(candidate))
	}
}

// pinWindowsDirectoryChain walks a path one directory at a time. Existing
// ancestors are opened with OPEN_REPARSE_POINT and no delete sharing. Missing
// components may be created one at a time, then pinned before the next
// component is created. This is the conservative Windows substitute for
// POSIX openat/mkdirat: pathname traversal is not exposed after the chain is
// pinned, and a reparse replacement is rejected before an external tree can
// be followed.
func pinWindowsDirectoryChain(path string, createMissing bool) ([]windows.Handle, error) {
	existing, missing, err := windowsExistingDirectoryPrefix(path)
	if err != nil {
		return nil, err
	}
	if !createMissing && len(missing) != 0 {
		return nil, fmt.Errorf("Windows path parent does not exist")
	}
	handles := make([]windows.Handle, 0, len(missing)+1)
	closeOnError := func(err error) ([]windows.Handle, error) {
		for i := len(handles) - 1; i >= 0; i-- {
			_ = windows.CloseHandle(handles[i])
		}
		return nil, err
	}
	current := existing
	first, err := openPinnedWindowsDirectory(current)
	if err != nil {
		return nil, fmt.Errorf("open Windows directory %q: %w", current, err)
	}
	handles = append(handles, first)
	for _, component := range missing {
		if component == "" || component == "." {
			continue
		}
		next := filepath.Join(current, component)
		if err := os.Mkdir(next, 0700); err != nil && !os.IsExist(err) {
			return closeOnError(fmt.Errorf("create Windows directory %q: %w", next, err))
		}
		h, err := openPinnedWindowsDirectory(next)
		if err != nil {
			return closeOnError(fmt.Errorf("open Windows directory %q: %w", next, err))
		}
		handles = append(handles, h)
		current = next
	}
	return handles, nil
}

func closeWindowsDirectoryChain(handles []windows.Handle) {
	for i := len(handles) - 1; i >= 0; i-- {
		_ = windows.CloseHandle(handles[i])
	}
}
