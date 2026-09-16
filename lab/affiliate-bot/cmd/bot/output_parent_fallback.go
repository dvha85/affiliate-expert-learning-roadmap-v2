//go:build !aix && !android && !darwin && !dragonfly && !freebsd && !hurd && !illumos && !ios && !linux && !netbsd && !openbsd && !solaris && !windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var outputParentBeforeOpenHook func()

func ensureOutputParentPlatform(parent string) error {
	for candidate := parent; ; candidate = filepath.Dir(candidate) {
		info, err := os.Lstat(candidate)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("backup/restore target parent contains a symlink or non-directory")
			}
			break
		}
		if !os.IsNotExist(err) {
			return err
		}
		next := filepath.Dir(candidate)
		if next == candidate {
			return fmt.Errorf("backup/restore target parent has no existing directory ancestor")
		}
	}
	if outputParentBeforeOpenHook != nil {
		outputParentBeforeOpenHook()
	}
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(parent)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("backup/restore target parent must be a non-symlink directory")
	}
	return nil
}
