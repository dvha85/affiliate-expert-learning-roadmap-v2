//go:build windows

package main

// Windows does not expose a portable directory-handle fsync through Go's
// os.File API. File contents are still flushed by the caller's file Sync;
// pathname identity and replacement checks remain the acknowledgement guard.
func syncDirectoryPlatform(string) error {
	return nil
}
