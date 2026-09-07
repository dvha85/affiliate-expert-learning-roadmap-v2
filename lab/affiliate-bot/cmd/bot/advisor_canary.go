package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Only the first reservation is allowed by this CLI. Further paid attempts
// require a separate reviewed change after inspecting the first canary.
func runCampaignCanary(ctx context.Context, path string, p *deepSeekProvider) (int, string, error) {
	key := os.Getenv("DEEPSEEK_API_KEY")
	if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "\r\n") {
		return 0, "CONFIG_ERROR", errors.New("DEEPSEEK_API_KEY unavailable or invalid; never paste it into logs")
	}
	if err := validateCampaignPath(path); err != nil {
		return 0, "PATH_ERROR", err
	}
	// Outside ledger so its strict entry whitelist stays unchanged. Held across
	// preflight and request: concurrent CLI processes cannot both pass first-run.
	app := filepath.Dir(path)
	lock := filepath.Join(app, "deepseek-br11-canary.lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		return 0, "LOCKED", errors.New("canary locked; inspect before recovery, do not retry automatically")
	}
	defer os.Remove(lock)
	if err := syncCampaignDir(app); err != nil {
		return 0, "IO_ERROR", err
	}
	r, err := readCampaignReport(path)
	if err != nil {
		return 0, "REPORT_ERROR", err
	}
	if r.Attempts != 0 {
		return 0, "REVIEW_REQUIRED", errors.New("campaign already has a reservation; review report and billing, never reset or retry canary")
	}
	bundle, err := os.MkdirTemp(app, "br11-canary-fixture-")
	if err != nil {
		return 0, "IO_ERROR", err
	}
	return runBR10RecordedCampaignAttempt(ctx, path, bundle, p)
}

func runCampaignCanaryCLI(args []string, stdout, stderr io.Writer) int {
	emit := func(n int, status string, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		if err := json.NewEncoder(stdout).Encode(map[string]any{"command": "advisor campaign-canary", "status": status, "attempt": n, "execution_permitted": false}); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 1 {
		return emit(0, "USAGE_ERROR", errors.New("usage: bot advisor campaign-canary (no path, endpoint, model or input arguments)"), 2)
	}
	path, err := advisorCampaignPath()
	if err != nil {
		return emit(0, "CONFIG_ERROR", err, 1)
	}
	n, status, err := runCampaignCanary(context.Background(), path, newDeepSeekProvider())
	code := 1
	if err == nil && (status == "SUPPORTED" || status == "ABSTAIN") {
		code = 0
	}
	return emit(n, status, err, code)
}
