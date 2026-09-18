//go:build windows

package store

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

func openStableRegularFileForRead(path string) (*os.File, error) {
	clean := filepath.Clean(path)
	parents, err := pinWindowsReadParent(filepath.Dir(clean))
	if err != nil {
		return nil, err
	}
	defer closeWindowsReadParents(parents)
	name, err := windows.UTF16PtrFromString(clean)
	if err != nil {
		return nil, err
	}
	h, err := windows.CreateFile(
		name,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return nil, err
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(h, &info); err != nil {
		_ = windows.CloseHandle(h)
		return nil, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("history path is not a regular file")
	}
	f := os.NewFile(uintptr(h), clean)
	if f == nil {
		_ = windows.CloseHandle(h)
		return nil, fmt.Errorf("history file handle could not be wrapped")
	}
	return f, nil
}
