package main

import (
	"encoding/json"
	"errors"
	"testing"

	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func TestM11RegistryAfterWriteFailureRecoversAsExactDuplicate(t *testing.T) {
	dir := t.TempDir()
	lease := corem11.ProductionLease{LeaseID: "fault-lease", LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "fault-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "fault-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	raw, err := json.Marshal(lease)
	if err != nil {
		t.Fatal(err)
	}
	m11RegistryAppendFault = func(phase string, _ corem11.ArtifactEntry) error {
		if phase == "after_write" {
			return errors.New("injected after-write failure")
		}
		return nil
	}
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLease, raw); err == nil {
		t.Fatal("after-write failure was not surfaced")
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("written artifact was not recoverable: entries=%d err=%v", len(entries), err)
	}
	m11RegistryAppendFault = nil
	if _, status, err := registerM11Artifact(dir, corem11.ArtifactKindLease, raw); err != nil || status != appendDuplicate {
		t.Fatalf("retry must be exact duplicate: status=%s err=%v", status, err)
	}
}
