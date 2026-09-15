//go:build !aix && !android && !darwin && !dragonfly && !freebsd && !hurd && !illumos && !ios && !linux && !netbsd && !openbsd && !solaris && !windows

package main

import "os"

func acquireManagedPathLock(path string) (func(), error) {
	if err := os.Mkdir(path, 0700); err != nil {
		if os.IsExist(err) {
			return nil, errManagedPathLockBusy
		}
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}
