//go:build !windows

package main

import "os"

func openStableRegularFileForRead(path string) (*os.File, error) {
	return os.Open(path)
}
