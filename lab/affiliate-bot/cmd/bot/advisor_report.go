package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Stable per OS-user configuration root; never derived from cwd/repo or a CLI
// path argument. Changing the OS profile/root is outside this local guarantee.
func advisorCampaignPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil || !filepath.IsAbs(root) {
		return "", errors.New("user config root unavailable")
	}
	return filepath.Join(root, "affiliate-expert-learning-roadmap-v2", "deepseek-br11-v1"), nil
}

type campaignReport struct {
	Version                      string           `json:"version"`
	Attempts                     int              `json:"attempts"`
	ReservedMicroUSD             int              `json:"reserved_microusd"`
	RemainingReservationMicroUSD int              `json:"remaining_reservation_microusd"`
	EstimatedKnownMicroUSD       int64            `json:"estimated_known_microusd"`
	MissingResultAttempts        []int            `json:"missing_result_attempts"`
	UnknownUsageAttempts         []int            `json:"unknown_usage_attempts"`
	Results                      []campaignResult `json:"results"`
	InvoiceReconciled            bool             `json:"invoice_reconciled"`
	ExecutionPermitted           bool             `json:"execution_permitted"`
}

func readCampaignReport(path string) (campaignReport, error) {
	r := campaignReport{Version: "br11-report/v1", MissingResultAttempts: []int{}, UnknownUsageAttempts: []int{}}
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return r, errors.New("campaign missing or invalid")
	}
	lock := filepath.Join(path, "lock")
	if err = os.Mkdir(lock, 0700); err != nil {
		return r, errors.New("campaign locked")
	}
	defer os.Remove(lock)
	r.Results, err = readCampaignResults(path)
	if err != nil {
		return r, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return r, err
	}
	for _, entry := range entries {
		if entry.Name() == "lock" || entry.Name() == "manifest" {
			continue
		}
		if !entry.Type().IsRegular() {
			return r, errors.New("campaign entry invalid")
		}
		if strings.HasPrefix(entry.Name(), "result-") {
			continue
		}
		r.Attempts++
	}
	if r.Attempts > 6 {
		return r, errors.New("campaign exceeds reservation cap")
	}
	for n := 1; n <= r.Attempts; n++ {
		want, _ := json.Marshal(campaignReservationRecord{n, campaignReservation})
		raw, err := readCampaignFile(filepath.Join(path, fmt.Sprintf("attempt-%03d.json", n)), int64(len(want)))
		if err != nil || string(raw) != string(want) {
			return r, errors.New("invalid reservation sequence")
		}
	}
	seen := map[int]bool{}
	for _, result := range r.Results {
		seen[result.Attempt] = true
		if result.EstimatedMicroUSD == nil {
			r.UnknownUsageAttempts = append(r.UnknownUsageAttempts, result.Attempt)
		} else {
			r.EstimatedKnownMicroUSD += *result.EstimatedMicroUSD
		}
	}
	for n := 1; n <= r.Attempts; n++ {
		if !seen[n] {
			r.MissingResultAttempts = append(r.MissingResultAttempts, n)
		}
	}
	r.ReservedMicroUSD = r.Attempts * campaignReservation
	r.RemainingReservationMicroUSD = 3000000 - r.ReservedMicroUSD
	return r, nil
}

func runCampaignReportCLI(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		envelope := map[string]any{"command": "advisor campaign-report", "status": status, "execution_permitted": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 1 {
		return emit("USAGE_ERROR", nil, errors.New("usage: bot advisor campaign-report (no path argument)"), 2)
	}
	path, err := advisorCampaignPath()
	if err != nil {
		return emit("CONFIG_ERROR", nil, err, 1)
	}
	if err := validateCampaignPath(path); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	r, err := readCampaignReport(path)
	if err != nil {
		return emit("REPORT_ERROR", nil, err, 1)
	}
	return emit("OK", r, nil, 0)
}
