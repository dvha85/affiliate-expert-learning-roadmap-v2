package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCampaignReportUnknownAndNoMutation(t *testing.T) {
	p := filepath.Join(t.TempDir(), "campaign")
	if err := initAdvisorCampaign(p); err != nil {
		t.Fatal(err)
	}
	if _, err := reserveAdvisorAttempt(p); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runRecordedCampaignAttempt(context.Background(), p, mockAdvisorProvider{}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(p, "attempt-001.json"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := readCampaignReport(p)
	if err != nil {
		t.Fatal(err)
	}
	if r.Attempts != 2 || r.ReservedMicroUSD != 1000000 || r.RemainingReservationMicroUSD != 2000000 || len(r.MissingResultAttempts) != 1 || r.MissingResultAttempts[0] != 1 || len(r.UnknownUsageAttempts) != 1 || r.UnknownUsageAttempts[0] != 2 || r.InvoiceReconciled || r.ExecutionPermitted {
		t.Fatal(r)
	}
	after, err := os.ReadFile(filepath.Join(p, "attempt-001.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("mutated reservation")
	}
	if _, err := os.Stat(filepath.Join(p, "lock")); !os.IsNotExist(err) {
		t.Fatal("lock not released")
	}
	if err := os.WriteFile(filepath.Join(p, "unexpected"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readCampaignReport(p); err == nil {
		t.Fatal("unknown file accepted")
	}
}
func TestCampaignReportNoImplicitInitOrPathArgument(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing")
	if _, err := readCampaignReport(p); err == nil {
		t.Fatal("missing accepted")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("created missing campaign")
	}
	if code := runAdvisor([]string{"campaign-report", p}, io.Discard, io.Discard); code != 2 {
		t.Fatal(code)
	}
	root, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := advisorCampaignPath()
	if err != nil || got != filepath.Join(root, "affiliate-expert-learning-roadmap-v2", "deepseek-br11-v1") {
		t.Fatal(got, err)
	}
}
