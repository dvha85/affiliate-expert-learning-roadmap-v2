//go:build windows

package store

// Windows does not expose a portable directory-handle fsync through Go's
// os.File API. File contents are still flushed by AppendLine's file Sync;
// pathname identity checks remain the acknowledgement guard.
func syncDirectory(string) error {
	return nil
}
