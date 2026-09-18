//go:build windows

package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// pinWindowsReadParent opens every existing parent component without
// following a reparse point and without delete sharing. This prevents a
// concurrent ancestor replacement while the final reader handle is opened.
// Windows still lacks a portable openat equivalent, so this is a conservative
// local pathname boundary rather than a distributed or power-loss guarantee.
func pinWindowsReadParent(path string) ([]windows.Handle, error) {
	existing, missing, err := windowsReadExistingPrefix(path)
	if err != nil {
		return nil, err
	}
	if len(missing) != 0 {
		return nil, fmt.Errorf("history path parent does not exist")
	}
	name, err := windows.UTF16PtrFromString(filepath.Clean(existing))
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("open history parent %q: %w", existing, err)
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		_ = windows.CloseHandle(h)
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("history parent %q is a reparse point", existing)
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("history parent %q is not a directory", existing)
	}
	return []windows.Handle{h}, nil
}

func windowsReadExistingPrefix(path string) (string, []string, error) {
	missing := make([]string, 0)
	for candidate := filepath.Clean(path); ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return "", nil, fmt.Errorf("history parent contains a symlink or non-directory")
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
			return "", nil, errors.New("history path has no existing directory ancestor")
		}
		missing = append(missing, filepath.Base(candidate))
	}
}

func closeWindowsReadParents(handles []windows.Handle) {
	for i := len(handles) - 1; i >= 0; i-- {
		_ = windows.CloseHandle(handles[i])
	}
}
