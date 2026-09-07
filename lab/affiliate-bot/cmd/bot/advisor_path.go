package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Check the existing chain without following symlinks. The application-owned
// portion must be private. Hostile concurrent path replacement is out of scope.
func validateCampaignPath(path string) error {
	if !strings.HasPrefix(filepath.Base(path), "deepseek-") {
		return errors.New("invalid campaign directory name")
	}
	return validatePrivateDirectoryChain(path, filepath.Dir(path))
}

func validatePrivateDirectoryChain(path, privateParent string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("campaign path must be absolute and clean")
	}
	for current := path; ; current = filepath.Dir(current) {
		info, err := os.Lstat(current)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("campaign path missing, symlinked or not a directory")
		}
		if current == path || current == privateParent {
			if info.Mode().Perm()&0077 != 0 {
				return errors.New("campaign application directories must be private")
			}
			if err := campaignOwnerCheck(info); err != nil {
				return err
			}
		}
		if filepath.Dir(current) == current {
			break
		}
	}
	return nil
}
