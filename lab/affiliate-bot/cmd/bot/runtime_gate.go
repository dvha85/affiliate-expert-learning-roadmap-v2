package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// runtimeGateName serializes every managed runtime mutation with backup
// capture. It is a directory because mkdir is an atomic cross-process claim
// on the local filesystem. A stale gate fails closed; no command removes it
// automatically or guesses whether a prior writer completed.
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
	path := runtimeGatePath(dir)
	if err := os.Mkdir(path, 0700); err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("runtime is busy or has an unrecovered writer; explicit recovery required")
		}
		return nil, err
	}
	return func() { _ = os.Remove(path) }, nil
}

func acquireHistoryRuntimeGate(historyPath string) (func(), error) {
	return acquireRuntimeGate(filepath.Dir(historyPath))
}
