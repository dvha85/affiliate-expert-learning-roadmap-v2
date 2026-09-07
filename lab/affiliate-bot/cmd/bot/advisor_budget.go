package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m04"
)

const campaignManifest = "deepseek-fixture-campaign/v1\nmax_attempts=100\nmax_microusd=3000000\nreservation_microusd=500000\n"

// Deliberately conservative: at most six reservations until cost accounting is
// separately reviewed. This is not a claim about current provider pricing.
const campaignReservation = 500000

// Trusted local directory; reject existing special files before opening them.
// This does not defend against a hostile process swapping paths concurrently.
func readCampaignFile(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("campaign file type or size invalid")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errors.New("campaign file read invalid")
	}
	return raw, nil
}

func syncCampaignDir(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

// Explicit setup only. Never recreate a missing ledger on the request path.
func initAdvisorCampaign(path string) error {
	if err := os.Mkdir(path, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(path, "manifest"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString(campaignManifest)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err = syncCampaignDir(path); err != nil {
		return err
	}
	return syncCampaignDir(filepath.Dir(path))
}

type campaignReservationRecord struct {
	Attempt          int `json:"attempt"`
	ReservedMicroUSD int `json:"reserved_microusd"`
}

func reserveAdvisorAttempt(path string) (int, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return 0, errors.New("campaign missing or invalid")
	}
	lock := filepath.Join(path, "lock")
	if err = os.Mkdir(lock, 0700); err != nil {
		return 0, errors.New("campaign locked; inspect interrupted run, do not reset")
	}
	defer os.Remove(lock)
	manifest, err := readCampaignFile(filepath.Join(path, "manifest"), int64(len(campaignManifest)))
	if err != nil || string(manifest) != campaignManifest {
		return 0, errors.New("campaign manifest invalid")
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.Name() == "lock" {
			continue
		}
		if !entry.Type().IsRegular() {
			return 0, errors.New("campaign entry invalid")
		}
		if entry.Name() == "manifest" {
			continue
		}
		count++
	}
	for n := 1; n <= count; n++ {
		want, _ := json.Marshal(campaignReservationRecord{n, campaignReservation})
		raw, err := readCampaignFile(filepath.Join(path, fmt.Sprintf("attempt-%03d.json", n)), int64(len(want)))
		if err != nil || string(raw) != string(want) {
			return 0, errors.New("campaign reservation corrupt or non-contiguous")
		}
	}
	if count >= 100 || count >= 3000000/campaignReservation {
		return 0, errors.New("campaign budget exhausted")
	}
	n := count + 1
	raw, _ := json.Marshal(campaignReservationRecord{n, campaignReservation})
	f, err := os.OpenFile(filepath.Join(path, fmt.Sprintf("attempt-%03d.json", n)), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return 0, err
	}
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return 0, err
	}
	if closeErr != nil {
		return 0, closeErr
	}
	if err = syncCampaignDir(path); err != nil {
		return 0, err
	}
	return n, nil
}

func campaignFixture() advisorContext {
	maxAge := 24
	return advisorContext{Version: "br11a-context/v1", Config: advisorConfig{DecisionID: "synthetic-decision", Question: "Nêu giới hạn của bằng chứng giả lập; không đề xuất thực thi.", AsOf: "2026-09-07T00:00:00Z", MaxAgeHours: &maxAge}, Evidence: []m04.AdvisorEvidence{{EvidenceID: "synthetic-evidence", ObservedAt: "2026-09-07T00:00:00Z", SourceRef: "fixture:br11b2"}}, Payload: map[string]any{"synthetic-evidence": map[string]any{"status": "PENDING", "metrics": map[string]any{}}}, Limitation: "Synthetic fixture only. No real affiliate results or business evidence."}
}

// Internal runner only: no arbitrary input/context and no CLI/live activation.
// A failed call consumes its reservation. No retry and no refund.
func runCampaignAttempt(ctx context.Context, path string, p advisorProvider) (int, string, error) {
	n, err := reserveAdvisorAttempt(path)
	if err != nil {
		return 0, "BUDGET_ERROR", err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, status := evaluateAdvisorProvider(ctx, p, campaignFixture())
	return n, status, nil
}
