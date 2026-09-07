package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Caller supplies the OS config root, never a user CLI path. Existing data is
// never repaired, chmodded, reset or removed on failure.
func initializeFixedCampaign(root string) error {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return errors.New("invalid config root")
	}
	// Verify existing ancestors before creating the application's directory.
	for p := root; ; p = filepath.Dir(p) {
		info, err := os.Lstat(p)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("config root missing or symlinked")
		}
		if p == root {
			if err := campaignOwnerCheck(info); err != nil {
				return err
			}
			if info.Mode().Perm()&0022 != 0 {
				return errors.New("config root writable by others")
			}
		}
		if filepath.Dir(p) == p {
			break
		}
	}
	app := filepath.Join(root, "affiliate-expert-learning-roadmap-v2")
	if err := os.Mkdir(app, 0700); err != nil && !os.IsExist(err) {
		return err
	}
	if err := validatePrivateDirectoryChain(app, app); err != nil {
		return err
	}
	if err := syncCampaignDir(root); err != nil {
		return err
	}
	path := filepath.Join(app, "deepseek-br11-v1")
	if err := initAdvisorCampaign(path); err != nil {
		return err
	}
	return validateCampaignPath(path)
}

func runCampaignInitCLI(args []string, stdout, stderr io.Writer) int {
	status, code := "INITIALIZED", 0
	var err error
	if len(args) != 1 {
		status, code, err = "USAGE_ERROR", 2, errors.New("usage: bot advisor campaign-init (no path argument)")
	} else {
		var root string
		root, err = os.UserConfigDir()
		if err == nil {
			err = initializeFixedCampaign(root)
		}
		if err != nil {
			status, code = "INIT_ERROR", 1
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	if err := json.NewEncoder(stdout).Encode(map[string]any{"command": "advisor campaign-init", "status": status, "execution_permitted": false}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return code
}
