//go:build windows

package main

import (
	"fmt"
	"path/filepath"
)

var outputParentBeforeOpenHook func()

func ensureOutputParentPlatform(parent string) error {
	clean := filepath.Clean(parent)
	if _, _, err := windowsExistingDirectoryPrefix(clean); err != nil {
		return fmt.Errorf("backup/restore target parent preflight: %w", err)
	}
	if outputParentBeforeOpenHook != nil {
		outputParentBeforeOpenHook()
	}
	handles, err := pinWindowsDirectoryChain(clean, true)
	if err != nil {
		return fmt.Errorf("backup/restore target parent traversal: %w", err)
	}
	closeWindowsDirectoryChain(handles)
	return nil
}
