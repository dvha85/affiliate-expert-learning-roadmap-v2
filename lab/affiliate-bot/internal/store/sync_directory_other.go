//go:build !windows

package store

import "os"

func syncDirectory(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	closeErr := d.Close()
	if err != nil {
		return err
	}
	return closeErr
}
