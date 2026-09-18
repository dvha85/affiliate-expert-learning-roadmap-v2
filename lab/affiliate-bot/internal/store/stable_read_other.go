//go:build !windows

package store

import "os"

func openStableRegularFileForRead(path string) (*os.File, error) {
	return os.Open(path)
}
