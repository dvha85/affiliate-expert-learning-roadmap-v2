//go:build !aix && !android && !darwin && !dragonfly && !freebsd && !hurd && !illumos && !ios && !linux && !netbsd && !openbsd && !solaris

package main

import "os"

// Platforms without the POSIX descriptor traversal keep the existing
// regular-file and O_EXCL checks. Their native lock/path hardening remains a
// separate compatibility boundary until a platform-specific stable parent
// primitive is available.
func openStableRegularFileForAppendPath(path string, newFile bool) (*os.File, error) {
	if stableRegularFileAppendHook != nil {
		if err := stableRegularFileAppendHook(path); err != nil {
			return nil, err
		}
	}
	flags := os.O_WRONLY | os.O_APPEND
	if newFile {
		flags |= os.O_CREATE | os.O_EXCL
	}
	return os.OpenFile(path, flags, 0600)
}
