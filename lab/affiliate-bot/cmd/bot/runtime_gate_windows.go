//go:build windows

package main

import "os"

// Windows keeps the existing cooperative directory claim. The POSIX
// regression is intentionally not advertised for this fallback until a
// native auto-releasing Windows lock is implemented.
func acquireManagedPathLock(path string) (func(), error) {
	if err := os.Mkdir(path, 0700); err != nil {
		if os.IsExist(err) {
			return nil, errManagedPathLockBusy
		}
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}
