package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func assertRegistryAppendRejectsParentSwapBeforeOpenat(t *testing.T, parent, registry string, register func() error) {
	t.Helper()
	external := filepath.Join(t.TempDir(), "external")
	if err := os.Mkdir(external, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(external, "sentinel")
	if err := os.WriteFile(sentinel, []byte("keep\n"), 0600); err != nil {
		t.Fatal(err)
	}
	moved := parent + "-moved"
	stableRegularFileAppendHook = func(path string) error {
		if path != registry {
			return nil
		}
		if err := os.Rename(parent, moved); err != nil {
			return err
		}
		return os.Symlink(external, parent)
	}
	t.Cleanup(func() { stableRegularFileAppendHook = nil })
	if err := register(); err == nil {
		t.Fatal("parent replacement reached an append target")
	}
	after, err := os.ReadFile(sentinel)
	if err != nil || !bytes.Equal(after, []byte("keep\n")) {
		t.Fatalf("external directory was changed: err=%v contents=%q", err, after)
	}
	if _, err := os.Stat(filepath.Join(external, filepath.Base(registry))); !os.IsNotExist(err) {
		t.Fatalf("registry was created below the external parent: err=%v", err)
	}
}

func TestM11RegistryAppendRejectsParentSwapBeforeOpenat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("descriptor-pinned parent traversal is POSIX-only")
	}
	root := t.TempDir()
	parent := filepath.Join(root, "runtime")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	lease := corem11.ProductionLease{LeaseID: "parent-swap-lease", LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "parent-swap-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "parent-swap-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	raw, err := json.Marshal(lease)
	if err != nil {
		t.Fatal(err)
	}
	registry := m11ArtifactRegistryPath(parent)
	assertRegistryAppendRejectsParentSwapBeforeOpenat(t, parent, registry, func() error {
		_, _, err := registerM11Artifact(parent, corem11.ArtifactKindLease, raw)
		return err
	})
}

func TestM10RegistryAppendRejectsParentSwapBeforeOpenat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("descriptor-pinned parent traversal is POSIX-only")
	}
	root := t.TempDir()
	parent := filepath.Join(root, "runtime")
	if err := os.Mkdir(parent, 0700); err != nil {
		t.Fatal(err)
	}
	bound := corem10.TrustedCostBound{CostBoundID: "parent-swap-cost", IntentID: "intent-parent-swap", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:registry", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: "parent-swap-correlation", HashVersion: "go-json-v1"}
	bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
	raw, err := json.Marshal(bound)
	if err != nil {
		t.Fatal(err)
	}
	registry := m10ArtifactRegistryPath(parent)
	assertRegistryAppendRejectsParentSwapBeforeOpenat(t, parent, registry, func() error {
		_, _, err := registerM10Artifact(parent, corem10.ArtifactKindTrustedCostBound, raw)
		return err
	})
}
