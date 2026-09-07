package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

// Frozen digest of buildBR10AdvisorFixture. A changed fixture requires a new
// version, not silently accepting arbitrary context under an existing ledger.
const br10CampaignContextSHA256 = "9260065e0ccfe2789ed0b5b4309513444fa2fdcd9777448b7a24745838674fda"

func advisorContextDigest(c advisorContext) string {
	raw, _ := json.Marshal(c)
	var normalized any
	_ = json.Unmarshal(raw, &normalized)
	raw, _ = json.Marshal(normalized)
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

func validateCampaignArtifact(r campaignResult) error {
	if r.Version == "br11-result/v1" {
		if r.ContextSHA256 != campaignContextDigest() || r.Context != nil || r.AdvisorOutput != nil {
			return errors.New("invalid legacy context")
		}
		return nil
	}
	if r.Context == nil || r.ContextSHA256 != br10CampaignContextSHA256 || advisorContextDigest(*r.Context) != r.ContextSHA256 {
		return errors.New("invalid BR10 fixture context")
	}
	accepted := r.Status == "SUPPORTED" || r.Status == "ABSTAIN"
	if !accepted {
		if r.AdvisorOutput != nil {
			return errors.New("rejected output must not be persisted")
		}
		return nil
	}
	if r.AdvisorOutput == nil {
		return errors.New("missing accepted output")
	}
	raw, err := json.Marshal(r.AdvisorOutput)
	if err != nil {
		return err
	}
	_, status := checkAdvisorResponse(raw, *r.Context)
	if status != r.Status || r.AdvisorOutput.State == "ADVISE" {
		return errors.New("output/status mismatch")
	}
	known := map[string]bool{}
	for _, e := range r.Context.Evidence {
		known[e.EvidenceID] = true
	}
	for _, id := range r.AdvisorOutput.EvidenceIDs {
		if !known[id] {
			return errors.New("unknown output evidence")
		}
	}
	return nil
}

// Internal only: use a fresh BR10 bundle, never reset the existing ledger.
// Tests use temporary campaigns and mock/loopback providers, not live credit.
func runBR10RecordedCampaignAttempt(ctx context.Context, path, bundle string, p advisorProvider) (int, string, error) {
	if p.identity() != (mockAdvisorProvider{}).identity() && p.identity() != newDeepSeekProvider().identity() {
		return 0, "CONFIG_ERROR", errors.New("unsupported campaign provider")
	}
	c, err := buildBR10AdvisorFixture(bundle)
	if err != nil || advisorContextDigest(c) != br10CampaignContextSHA256 {
		return 0, "FIXTURE_ERROR", errors.New("BR10 fixture build or version mismatch")
	}
	// Normalize interface payloads before persistence: typed structs and decoded
	// maps must produce identical canonical bytes after restart.
	rawContext, _ := json.Marshal(c)
	if err := json.Unmarshal(rawContext, &c); err != nil {
		return 0, "FIXTURE_ERROR", err
	}
	n, err := reserveAdvisorAttempt(path)
	if err != nil {
		return 0, "BUDGET_ERROR", err
	}
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	out, status := evaluateAdvisorProvider(deadline, p, c)
	r := campaignResult{Version: "br11-result/v2", Attempt: n, Provider: p.identity(), ContextSHA256: br10CampaignContextSHA256, Context: &c, Status: status, PriceVersion: campaignPriceVersion}
	if status == "SUPPORTED" || status == "ABSTAIN" {
		r.AdvisorOutput = &out
	}
	if ds, ok := p.(*deepSeekProvider); ok && ds.usage.PromptTokens != nil {
		u := ds.usage
		r.Usage = &u
	}
	r.EstimatedMicroUSD, err = estimateCampaignUsage(r.Usage)
	if err == nil {
		err = persistCampaignResult(path, r)
	}
	if err != nil {
		return n, "RESULT_ERROR", err
	}
	return n, status, nil
}
