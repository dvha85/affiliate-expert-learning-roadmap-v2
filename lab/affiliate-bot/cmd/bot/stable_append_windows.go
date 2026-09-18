//go:build windows

package main

import (
	"fmt"
	"io"
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
	// CreateFile starts the file pointer at offset zero; unlike POSIX
	// O_APPEND, GENERIC_WRITE alone does not append. Seek while the handle is
	// exclusive for writers so canonical JSONL registries keep their history.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("seek stable append target: %w", err)
	}
	return f, nil
}
