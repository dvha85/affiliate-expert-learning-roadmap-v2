package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCampaignResultRejectsWrongContextAndManifest(t *testing.T) {
	for _, kind := range []string{"context", "manifest"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "campaign")
			if err := initAdvisorCampaign(path); err != nil {
				t.Fatal(err)
			}
			if _, _, err := runRecordedCampaignAttempt(context.Background(), path, mockAdvisorProvider{}); err != nil {
				t.Fatal(err)
			}
			results, err := readCampaignResults(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := reserveAdvisorAttempt(path); err != nil {
				t.Fatal(err)
			}
			if kind == "context" {
				results[0].ContextSHA256 = strings.Repeat("0", 64)
				raw, _ := json.Marshal(results[0])
				if err := os.WriteFile(filepath.Join(path, "result-001.json"), raw, 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.WriteFile(filepath.Join(path, "manifest"), []byte("wrong campaign"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := readCampaignResults(path); err == nil {
				t.Fatal("invalid provenance accepted")
			}
			if _, err := reserveAdvisorAttempt(path); err == nil {
				t.Fatal("invalid provenance allowed next attempt")
			}
			results[0].Attempt = 2
			if err := persistCampaignResult(path, results[0]); err == nil {
				t.Fatal("invalid campaign accepted write")
			}
		})
	}
}

func TestCampaignResultRestartAndNoRefund(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 6; i++ {
		n, status, err := runRecordedCampaignAttempt(context.Background(), p, mockAdvisorProvider{})
		if err != nil || n != i || status != "SUPPORTED" {
			t.Fatal(n, status, err)
		}
		results, err := readCampaignResults(p)
		if err != nil || len(results) != i {
			t.Fatal(results, err)
		}
		if results[i-1].Usage != nil || results[i-1].EstimatedMicroUSD != nil || results[i-1].ExecutionPermitted {
			t.Fatal("fabricated cost/permission")
		}
	}
	if _, _, err := runRecordedCampaignAttempt(context.Background(), p, mockAdvisorProvider{}); err == nil {
		t.Fatal("refunded reservations")
	}
}
func TestCampaignResultCorruptionAndOrphan(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runRecordedCampaignAttempt(context.Background(), p, mockAdvisorProvider{}); err != nil {
		t.Fatal(err)
	}
	rs, err := readCampaignResults(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := persistCampaignResult(p, rs[0]); err == nil {
		t.Fatal("overwrote immutable result")
	}
	r := rs[0]
	r.Attempt = 2
	if err := persistCampaignResult(p, r); err == nil {
		t.Fatal("orphan persisted")
	}
	raw, _ := json.Marshal(r)
	if err := os.WriteFile(filepath.Join(p, "result-002.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := reserveAdvisorAttempt(p); err == nil {
		t.Fatal("orphan did not block next request")
	}
}
func TestCampaignMissingResultKeepsReservation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	if _, err := reserveAdvisorAttempt(p); err != nil {
		t.Fatal(err)
	}
	n, _, err := runRecordedCampaignAttempt(context.Background(), p, mockAdvisorProvider{})
	if err != nil || n != 2 {
		t.Fatal(n, err)
	}
	rs, err := readCampaignResults(p)
	if err != nil || len(rs) != 1 || rs[0].Attempt != 2 {
		t.Fatal(rs, err)
	}
}
func TestCampaignUsageIntegerEstimate(t *testing.T) {
	for _, tc := range []struct {
		in, out int
		want    int64
	}{{0, 0, 0}, {1, 0, 1}, {100, 50, 110}, {1000000, 1024, 441352}} {
		u := deepSeekUsage{ptr(tc.in), ptr(tc.out), ptr(tc.in + tc.out)}
		cost, err := estimateCampaignUsage(&u)
		if err != nil || *cost != tc.want {
			t.Fatal(cost, err)
		}
	}
	u := deepSeekUsage{ptr(-1), ptr(0), ptr(-1)}
	if _, err := estimateCampaignUsage(&u); err == nil {
		t.Fatal("negative usage")
	}
	u = deepSeekUsage{ptr(10), ptr(0), ptr(11)}
	if _, err := estimateCampaignUsage(&u); err == nil {
		t.Fatal("mismatched usage")
	}
	if cost, err := estimateCampaignUsage(nil); cost != nil || err != nil {
		t.Fatal("unknown is not zero")
	}
}
