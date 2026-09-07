package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

const campaignPriceVersion = "deepseek-flash-peak-cache-miss-2026-09-07"

// Metadata only. Never persist raw provider text, secrets, or rejected output.
type campaignResult struct {
	Version            string           `json:"version"`
	Attempt            int              `json:"attempt"`
	Provider           providerIdentity `json:"provider"`
	ContextSHA256      string           `json:"context_sha256"`
	Status             string           `json:"status"`
	Usage              *deepSeekUsage   `json:"usage"`
	PriceVersion       string           `json:"price_version"`
	EstimatedMicroUSD  *int64           `json:"estimated_microusd"`
	ExecutionPermitted bool             `json:"execution_permitted"`
}

func estimateCampaignUsage(u *deepSeekUsage) (*int64, error) {
	if u == nil {
		return nil, nil
	}
	if u.PromptTokens == nil || u.CompletionTokens == nil || u.TotalTokens == nil || *u.PromptTokens < 0 || *u.PromptTokens > 1000000 || *u.CompletionTokens < 0 || *u.CompletionTokens > 1024 || *u.TotalTokens != *u.PromptTokens+*u.CompletionTokens {
		return nil, errors.New("invalid campaign usage")
	}
	// USD/MTok 0.44 input, 1.32 output; round UP to whole microUSD.
	n := (int64(*u.PromptTokens)*44 + int64(*u.CompletionTokens)*132 + 99) / 100
	return &n, nil
}

func validateCampaignResult(r campaignResult) error {
	if r.Version != "br11-result/v1" || r.Attempt < 1 || r.Attempt > 6 || r.ExecutionPermitted || r.PriceVersion != campaignPriceVersion {
		return errors.New("invalid result metadata")
	}
	if r.Provider != (mockAdvisorProvider{}).identity() && r.Provider != newDeepSeekProvider().identity() {
		return errors.New("invalid result provider")
	}
	switch r.Status {
	case "SUPPORTED", "ABSTAIN", "ABSTAIN_STALE", "ABSTAIN_FUTURE", "INVALID", "INVALID_SCHEMA", "REJECT_UNGROUNDED", "REJECT_WRITE_REQUEST", "REJECT_ADVICE", "PROVIDER_ERROR", "CONFIG_ERROR":
	default:
		return errors.New("invalid result status")
	}
	h, err := hex.DecodeString(r.ContextSHA256)
	if err != nil || len(h) != 32 {
		return errors.New("invalid context digest")
	}
	cost, err := estimateCampaignUsage(r.Usage)
	if err != nil {
		return err
	}
	if (cost == nil) != (r.EstimatedMicroUSD == nil) || (cost != nil && *cost != *r.EstimatedMicroUSD) {
		return errors.New("result cost mismatch")
	}
	if r.Provider.Provider == "mock/v1" && r.Usage != nil {
		return errors.New("mock cannot report billed usage")
	}
	return nil
}

// Called under the same campaign lock as reservation/result writers.
func readCampaignResults(path string) ([]campaignResult, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	results := []campaignResult{}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "result-") {
			continue
		}
		raw, err := readCampaignFile(filepath.Join(path, entry.Name()), 4096)
		if err != nil {
			return nil, err
		}
		var r campaignResult
		if contracts.DecodeStrict(raw, &r) != nil || validateCampaignResult(r) != nil || entry.Name() != fmt.Sprintf("result-%03d.json", r.Attempt) {
			return nil, errors.New("corrupt campaign result")
		}
		canonical, _ := json.Marshal(r)
		if string(canonical) != string(raw) {
			return nil, errors.New("noncanonical campaign result")
		}
		reservation, err := readCampaignFile(filepath.Join(path, fmt.Sprintf("attempt-%03d.json", r.Attempt)), 128)
		want, _ := json.Marshal(campaignReservationRecord{r.Attempt, campaignReservation})
		if err != nil || string(reservation) != string(want) {
			return nil, errors.New("orphan campaign result")
		}
		results = append(results, r)
	}
	return results, nil
}

func persistCampaignResult(path string, r campaignResult) error {
	if err := validateCampaignResult(r); err != nil {
		return err
	}
	lock := filepath.Join(path, "lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		return errors.New("campaign locked")
	}
	defer os.Remove(lock)
	if _, err := readCampaignResults(path); err != nil {
		return err
	}
	rawReservation, err := readCampaignFile(filepath.Join(path, fmt.Sprintf("attempt-%03d.json", r.Attempt)), 128)
	want, _ := json.Marshal(campaignReservationRecord{r.Attempt, campaignReservation})
	if err != nil || string(rawReservation) != string(want) {
		return errors.New("missing reservation")
	}
	raw, _ := json.Marshal(r)
	f, err := os.OpenFile(filepath.Join(path, fmt.Sprintf("result-%03d.json", r.Attempt)), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	closed := f.Close()
	if err != nil {
		return err
	}
	if closed != nil {
		return closed
	}
	return syncCampaignDir(path)
}

// Internal only. This uses the ledger fixture, not yet the BR-10 live runner.
func runRecordedCampaignAttempt(ctx context.Context, path string, p advisorProvider) (int, string, error) {
	if p.identity() != (mockAdvisorProvider{}).identity() && p.identity() != newDeepSeekProvider().identity() {
		return 0, "CONFIG_ERROR", errors.New("unsupported campaign provider")
	}
	n, status, err := runCampaignAttempt(ctx, path, p)
	if err != nil {
		return n, status, err
	}
	raw, _ := json.Marshal(campaignFixture())
	digest := sha256.Sum256(raw)
	r := campaignResult{Version: "br11-result/v1", Attempt: n, Provider: p.identity(), ContextSHA256: hex.EncodeToString(digest[:]), Status: status, PriceVersion: campaignPriceVersion}
	if ds, ok := p.(*deepSeekProvider); ok && ds.usage.PromptTokens != nil {
		u := ds.usage
		r.Usage = &u
	}
	r.EstimatedMicroUSD, err = estimateCampaignUsage(r.Usage)
	if err != nil {
		return n, "RESULT_ERROR", err
	}
	if err = persistCampaignResult(path, r); err != nil {
		return n, "RESULT_ERROR", err
	}
	return n, status, nil
}
