//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// Windows opens the append target with a native handle and refuses a final
// reparse point. The caller-owned parent preflight remains deliberately
// bounded because Windows has no direct openat equivalent here.
func openStableRegularFileForAppendPath(path string, newFile bool) (*os.File, error) {
	if stableRegularFileAppendHook != nil {
		if err := stableRegularFileAppendHook(path); err != nil {
			return nil, err
		}
	}
	creation := uint32(windows.OPEN_EXISTING)
	if newFile {
		creation = windows.CREATE_NEW
	}
	f, err := openWindowsRegularFile(filepath.Clean(path), windows.GENERIC_READ|windows.GENERIC_WRITE, windows.FILE_SHARE_READ, creation)
	if err != nil {
		return nil, fmt.Errorf("open stable append target: %w", err)
	}
	return f, nil
}
