package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var errManagedPathLockBusy = errors.New("managed path lock is busy")

// runtimeGateName serializes every managed runtime mutation with backup
// capture. On POSIX the path is an advisory OS lock, so the kernel releases it
// when a writer process exits. The path itself is retained as a regular file;
// it is control metadata and is excluded from backup inventory.
const runtimeGateName = ".runtime-gate.lock"

func runtimeGatePath(dir string) string { return filepath.Join(dir, runtimeGateName) }

func acquireRuntimeGate(dir string) (func(), error) {
	info, err := os.Lstat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("runtime root must be a non-symlink directory")
	}
	release, err := acquireManagedPathLock(runtimeGatePath(dir))
	if err != nil {
		if errors.Is(err, errManagedPathLockBusy) {
			return nil, fmt.Errorf("runtime is busy or has an active writer")
		}
		return nil, err
	}
	return release, nil
}

func acquireHistoryRuntimeGate(historyPath string) (func(), error) {
	return acquireRuntimeGate(filepath.Dir(historyPath))
}
