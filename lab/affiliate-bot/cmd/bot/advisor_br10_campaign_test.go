package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestBR10CampaignFrozenContext(t *testing.T) {
	c, err := buildBR10AdvisorFixture(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got := advisorContextDigest(c); got != br10CampaignContextSHA256 {
		t.Fatalf("fixture digest %s", got)
	}
}

func TestBR10CampaignDeepSeekLoopback(t *testing.T) {
	t.Setenv("DEEPSEEK_API_KEY", "fixture-key-not-real")
	path := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(path); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if _, err := os.Stat(filepath.Join(path, "attempt-001.json")); err != nil {
			t.Error("request before reservation", err)
		}
		content := `{"state":"HUMAN_REVIEW","recommendation":"Cần người đánh giá.","reason":"Outcome còn pending.","evidence_ids":["br11-outcome"],"unknowns":["Chưa có kết quả đo"],"write_tool_requested":false}`
		_ = json.NewEncoder(w).Encode(map[string]any{"model": deepSeekModel, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": content}}}})
	}))
	defer server.Close()
	p := newDeepSeekProvider()
	p.endpoint = server.URL
	n, status, err := runBR10RecordedCampaignAttempt(context.Background(), path, t.TempDir(), p)
	if err != nil || n != 1 || status != "SUPPORTED" || calls != 1 {
		t.Fatal(n, status, err, calls)
	}
	results, err := readCampaignResults(path)
	if err != nil || len(results) != 1 {
		t.Fatal(results, err)
	}
	r := results[0]
	if r.AdvisorOutput == nil || r.Usage == nil || r.EstimatedMicroUSD == nil || *r.EstimatedMicroUSD != 110 || r.Provider != p.identity() {
		t.Fatal(r)
	}
}

func TestBR10CampaignPreservesLegacyAndBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runRecordedCampaignAttempt(context.Background(), path, mockAdvisorProvider{}); err != nil {
		t.Fatal(err)
	}
	old, err := os.ReadFile(filepath.Join(path, "result-001.json"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 2; i <= 6; i++ {
		n, status, err := runBR10RecordedCampaignAttempt(context.Background(), path, t.TempDir(), mockAdvisorProvider{})
		if err != nil || n != i || status != "SUPPORTED" {
			t.Fatal(n, status, err)
		}
	}
	now, err := os.ReadFile(filepath.Join(path, "result-001.json"))
	if err != nil || string(now) != string(old) {
		t.Fatal("legacy rewritten", err)
	}
	report, err := readCampaignReport(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Results) != 6 || report.Results[0].Version != "br11-result/v1" || report.Results[1].AdvisorOutput == nil {
		t.Fatal(report)
	}
	if _, _, err := runBR10RecordedCampaignAttempt(context.Background(), path, t.TempDir(), mockAdvisorProvider{}); err == nil {
		t.Fatal("budget reset")
	}
}

func TestBR10CampaignRejectsArtifactTampering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(path); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runBR10RecordedCampaignAttempt(context.Background(), path, t.TempDir(), mockAdvisorProvider{}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(path, "result-001.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"context", "rehashed", "missing", "advice", "status", "unknown-abstain", "legacy"} {
		t.Run(kind, func(t *testing.T) {
			var r campaignResult
			if err := json.Unmarshal(raw, &r); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "context":
				r.Context.Config.Question = "changed"
			case "rehashed":
				r.Context.Config.Question = "changed"
				r.ContextSHA256 = advisorContextDigest(*r.Context)
			case "missing":
				r.AdvisorOutput = nil
			case "advice":
				r.AdvisorOutput.State = "ADVISE"
			case "status":
				r.Status = "PROVIDER_ERROR"
			case "unknown-abstain":
				r.Status = "ABSTAIN"
				r.AdvisorOutput.State = "ABSTAIN"
				r.AdvisorOutput.EvidenceIDs = []string{"orphan"}
			case "legacy":
				r.Version = "br11-result/v1"
				r.ContextSHA256 = campaignContextDigest()
			}
			if validateCampaignResult(r) == nil {
				t.Fatal("tampered result accepted")
			}
		})
	}
}

type rejectedBR10Provider struct{}

func (rejectedBR10Provider) identity() providerIdentity { return (mockAdvisorProvider{}).identity() }
func (rejectedBR10Provider) generate(context.Context, advisorContext) ([]byte, error) {
	return []byte(`{"secret":"must-not-store"}`), nil
}

func TestBR10CampaignRejectedOutputNotStored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(path); err != nil {
		t.Fatal(err)
	}
	n, status, err := runBR10RecordedCampaignAttempt(context.Background(), path, t.TempDir(), rejectedBR10Provider{})
	if err != nil || n != 1 || status == "SUPPORTED" {
		t.Fatal(n, status, err)
	}
	results, err := readCampaignResults(path)
	if err != nil || len(results) != 1 || results[0].AdvisorOutput != nil {
		t.Fatal(results, err)
	}
	if n, err := reserveAdvisorAttempt(path); err != nil || n != 2 {
		t.Fatal("rejected attempt refunded", n, err)
	}
}
