package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
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

	second := lease
	second.LeaseID, second.ApprovalRef, second.CorrelationID = "uncertain-lease", "uncertain-approval", "uncertain-correlation"
	second.LeaseHash = corem11.ComputeProductionLeaseHash(second)
	secondRaw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	artifactRegistryPublishFailure = func(path string) error {
		if filepath.Clean(path) == filepath.Clean(m11ArtifactRegistryPath(dir)) {
			return errors.New("injected registry parent sync failure")
		}
		return nil
	}
	if entry, status, err := registerM11Artifact(dir, corem11.ArtifactKindLease, secondRaw); err == nil || status != appendAdded {
		t.Fatalf("registry publication uncertainty was not surfaced: entry=%+v status=%s err=%v", entry, status, err)
	} else {
		var uncertain *immutableArtifactPublishUncertainError
		if !errors.As(err, &uncertain) || entry.ArtifactID != second.LeaseID {
			t.Fatalf("registry publication did not retain resolvable immutable entry: entry=%+v err=%v", entry, err)
		}
	}
	entries, err = loadM11ArtifactRegistry(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("uncertain M11 registry append was not retained: entries=%d err=%v", len(entries), err)
	}
	artifactRegistryPublishFailure = nil
	if _, status, err := registerM11Artifact(dir, corem11.ArtifactKindLease, secondRaw); err != nil || status != appendDuplicate {
		t.Fatalf("uncertain M11 registry append did not exact-retry: status=%s err=%v", status, err)
	}
}

func TestM11RegistryAppendRejectsSameByteNameReplacement(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	makeLease := func(id string) ([]byte, error) {
		lease := corem11.ProductionLease{LeaseID: id, LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "approval-" + id, ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "corr-" + id, HashVersion: "go-json-v1"}
		lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
		return json.Marshal(lease)
	}
	first, err := makeLease("lease-append-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLease, first); err != nil {
		t.Fatal(err)
	}
	registry := m11ArtifactRegistryPath(dir)
	original, err := os.ReadFile(registry)
	if err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "same-byte-external-registry.jsonl")
	if err := os.WriteFile(external, original, 0600); err != nil {
		t.Fatal(err)
	}
	second, err := makeLease("lease-append-2")
	if err != nil {
		t.Fatal(err)
	}
	swapped := false
	stableRegularFileAppendHook = func(path string) error {
		if path != registry {
			return nil
		}
		swapped = true
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { stableRegularFileAppendHook = nil })
	if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLease, second); err == nil {
		t.Fatal("same-byte M11 registry replacement reached append")
	}
	if !swapped {
		t.Fatal("M11 append replacement seam was not reached")
	}
	after, err := os.ReadFile(external)
	if err != nil || string(after) != string(original) {
		t.Fatalf("external registry was changed: err=%v", err)
	}
}

func TestM11RegistryLoaderRejectsSchemaValidOrphanBeforeRuntimeUse(t *testing.T) {
	dir := t.TempDir()
	approval := corem11.ProductionLeaseApproval{
		ApprovalID: "orphan-approval", LeaseID: "missing-lease", LeaseVersion: "v1",
		LeaseHash:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		PromotionReviewRef: "fixture-review", SourceCanaryGrantID: "fixture-grant",
		SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"},
		ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z",
		Decision: "APPROVE_PRODUCTION_LEASE",
	}
	raw, err := json.Marshal(approval)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLeaseApproval, raw)
	if err != nil {
		t.Fatal(err)
	}
	line, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(dir), append(line, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11ArtifactRegistry(dir); err == nil || !strings.Contains(err.Error(), "orphaned") {
		t.Fatalf("schema-valid orphan approval reached runtime loader: %v", err)
	}
}

func TestM11RegistryLoaderRejectsOrphanReconciliationLedgerLink(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	entries, err := readM11ArtifactRegistry(fixture.dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 || entries[len(entries)-1].ArtifactKind != corem11.ArtifactKindLedger {
		t.Fatalf("fixture registry does not end with a ledger: %d entries", len(entries))
	}
	value, status := corem11.DecodeArtifact("ledger", entries[len(entries)-1].Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode ledger: %s", status)
	}
	ledger := *value.(*corem11.ProductionLedger)
	ledger.ControlMode = "STOPPED"
	ledger.StopReason = "RECOVERY_REVIEW_REQUIRED"
	ledger.ReconciliationRequired = false
	ledger.ReconciliationResolutionIDs = []string{"missing-resolution"}
	ledgerRaw, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	entries[len(entries)-1], err = corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	contents := []byte{}
	for _, entry := range entries {
		line, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		contents = append(contents, line...)
		contents = append(contents, '\n')
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(fixture.dir), contents, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11ArtifactRegistry(fixture.dir); err == nil || !strings.Contains(err.Error(), "reconciliation") {
		t.Fatalf("schema-valid orphan reconciliation link reached runtime loader: %v", err)
	}
}

func TestM11RegistryLoaderRejectsMixedCorrelationLineage(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	entries, err := readM11ArtifactRegistry(fixture.dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 8 {
		t.Fatalf("fixture registry is unexpectedly short: %d", len(entries))
	}
	healthValue, status := corem11.DecodeArtifact("health", entries[2].Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode health: %s", status)
	}
	costValue, status := corem11.DecodeArtifact("cost", entries[3].Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode cost: %s", status)
	}
	gateValue, status := corem11.DecodeArtifact("gate", entries[6].Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode gate: %s", status)
	}
	authValue, status := corem11.DecodeArtifact("authorization", entries[7].Artifact)
	if status != corem11.Valid {
		t.Fatalf("decode authorization: %s", status)
	}
	health := *healthValue.(*corem11.ProductionHealthSnapshot)
	foreignCost := costValue.(corem10.TrustedCostBound)
	foreignCost.CorrelationID = "foreign-correlation"
	foreignCost.CostBoundHash = corem10.ComputeTrustedCostBoundHash(foreignCost)
	foreignGate := *gateValue.(*corem11.ProductionGateDecision)
	foreignGate.CostBoundHash = foreignCost.CostBoundHash
	ledgerEntry := entries[4]
	lease := fixture.lease
	foreignGate.GateID = corem11.ComputeProductionGateID(lease, foreignGate.IntentID, foreignGate.IntentHash, health, foreignCost, ledgerEntry, foreignGate.EvaluatedAt)
	foreignAuth := *authValue.(*corem11.ProductionExecutionAuthorization)
	foreignAuth.ProductionGateID = foreignGate.GateID
	foreignAuth.ProductionCostBoundHash = foreignCost.CostBoundHash
	foreignAuth.CorrelationID = foreignCost.CorrelationID
	foreignAuth.AuthorizationID = corem11.ComputeProductionAuthorizationID(foreignGate.GateID, foreignAuth.ExecutorID, foreignAuth.AuthorizedAt)

	// Rewrite only a disposable registry with checksum-valid replacements. The
	// old loader accepted this shape because all copied IDs/hashes were updated;
	// the shared runtime loader must now reject the mixed correlation root.
	mixed := append([]corem11.ArtifactEntry(nil), entries[:8]...)
	raw, err := json.Marshal(foreignCost)
	if err != nil {
		t.Fatal(err)
	}
	mixed[3], err = corem11.NewArtifactEntry(corem11.ArtifactKindCostBound, raw)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(foreignGate)
	if err != nil {
		t.Fatal(err)
	}
	mixed[6], err = corem11.NewArtifactEntry(corem11.ArtifactKindGate, raw)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(foreignAuth)
	if err != nil {
		t.Fatal(err)
	}
	mixed[7], err = corem11.NewArtifactEntry(corem11.ArtifactKindAuthorization, raw)
	if err != nil {
		t.Fatal(err)
	}
	contents := []byte{}
	for _, entry := range mixed {
		line, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		contents = append(contents, line...)
		contents = append(contents, '\n')
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(fixture.dir), contents, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11ArtifactRegistry(fixture.dir); err == nil {
		t.Fatal("runtime loader accepted a mixed M11 correlation lineage")
	}
}

func TestM11RegistryLoaderRejectsEvaluationWithMismatchedEvidence(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init evaluation fixture: code=%d response=%+v", code, response)
	}
	journal, predecessor := setupM11OutcomeJournalFixture(t, dir)
	raw, err := json.Marshal(journal.Outcome)
	if err != nil {
		t.Fatal(err)
	}
	outcome, _, status, err := recordM11FixtureOutcome(dir, predecessor.ArtifactID, raw)
	if err != nil || status != appendAdded {
		t.Fatalf("record fixture outcome: status=%s err=%v", status, err)
	}
	if _, status, err := evaluateM11FixtureOutcome(dir, outcome.OutcomeID, "loader-evidence-evaluation", "2026-09-08T00:00:03Z"); err != nil || status != appendAdded {
		t.Fatalf("record fixture evaluation: status=%s err=%v", status, err)
	}

	entries, err := readM11ArtifactRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	evaluationIndex := -1
	for index, entry := range entries {
		if entry.ArtifactKind == corem11.ArtifactKindEvaluation {
			evaluationIndex = index
			break
		}
	}
	if evaluationIndex < 0 {
		t.Fatal("fixture evaluation was not recorded")
	}
	value, decodeStatus := corem11.DecodeArtifact("evaluation", entries[evaluationIndex].Artifact)
	if decodeStatus != corem11.Valid {
		t.Fatalf("decode evaluation: %s", decodeStatus)
	}
	evaluation := *value.(*corem11.ProductionOutcomeEvaluation)
	evaluation.EvidenceIDs = []string{"unrelated-evidence"}
	evaluationRaw, err := json.Marshal(evaluation)
	if err != nil {
		t.Fatal(err)
	}
	entries[evaluationIndex], err = corem11.NewArtifactEntry(corem11.ArtifactKindEvaluation, evaluationRaw)
	if err != nil {
		t.Fatal(err)
	}

	contents := []byte{}
	for _, entry := range entries {
		line, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		contents = append(contents, line...)
		contents = append(contents, '\n')
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(dir), contents, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11ArtifactRegistry(dir); err == nil || !strings.Contains(err.Error(), "outcome evaluation") {
		t.Fatalf("checksum-valid evaluation with mismatched evidence reached runtime loader: %v", err)
	}
}

func TestM11RegistryLoaderRejectsRecoveryAdmissionWithoutRecordedApproval(t *testing.T) {
	dir := t.TempDir()
	lease := corem11.ProductionLease{LeaseID: "admission-lease", LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "admission-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "admission-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	admission := corem11.ProductionRecoveryAdmission{RecoveryAdmissionID: "missing-approval-admission", PriorRuntimeDir: "/runtime/old", PriorLeaseID: "old-lease", PriorLeaseVersion: "v1", PriorLeaseHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", PriorApprovalID: "old-approval", ResolutionID: "old-resolution", NewRuntimeID: "runtime-new", NewRuntimeDir: "/runtime/new", NewLeaseID: lease.LeaseID, NewLeaseVersion: lease.LeaseVersion, NewLeaseHash: lease.LeaseHash, NewApprovalID: lease.ApprovalRef, ReviewedBy: "human", ReviewerID: "new-reviewer", ReviewedAt: "2026-09-08T00:00:01Z", ExecutionPermitted: false}
	entries := []corem11.ArtifactEntry{}
	for _, artifact := range []struct {
		kind  string
		value any
	}{
		{corem11.ArtifactKindLease, lease},
		{corem11.ArtifactKindRecoveryAdmission, admission},
	} {
		raw, err := json.Marshal(artifact.value)
		if err != nil {
			t.Fatal(err)
		}
		entry, err := corem11.NewArtifactEntry(artifact.kind, raw)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	lines := []byte{}
	for _, entry := range entries {
		line, err := json.Marshal(entry)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, append(line, '\n')...)
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(dir), lines, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM11ArtifactRegistry(dir); err == nil || !strings.Contains(err.Error(), "orphaned") {
		t.Fatalf("recovery admission without approval reached runtime loader: %v", err)
	}
}

func TestM11JournalSymlinkFailsClosedBeforeRecoveryOrMutation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	external := filepath.Join(t.TempDir(), "untrusted-journal.json")
	if err := os.WriteFile(external, []byte(`{"version":"m11-unknown-stop-journal/v1"}`), 0600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	binary := buildMissionBinary(t)
	for _, journal := range []string{m11ManualStopJournalPath(dir), m11OutcomeJournalPath(dir), m11FailedExecutionJournalPath(dir), m11UnknownStopJournalPath(dir)} {
		if err := os.Symlink(external, journal); err != nil {
			t.Fatal(err)
		}
		if err := m11JournalRecoveryRequired(dir); err == nil {
			t.Fatal("symlinked M11 journal was not rejected before recovery")
		}
		if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("status did not fail closed for %s: code=%d response=%+v", filepath.Base(journal), code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread-input.json")); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("writer reached input handling for %s: code=%d response=%+v", filepath.Base(journal), code, response)
		}
		if _, err := os.Stat(m11ArtifactRegistryPath(dir)); !os.IsNotExist(err) {
			t.Fatalf("symlinked journal caused registry mutation: %v", err)
		}
		after, err := os.ReadFile(external)
		if err != nil || string(after) != string(before) {
			t.Fatalf("recovery touched external journal target: err=%v before=%q after=%q", err, before, after)
		}
		if err := os.Remove(journal); err != nil {
			t.Fatal(err)
		}
	}
}

func TestM11RecoveryAdmissionFailsClosedWhileOldRuntimeGateIsHeld(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	if code, response := missionCall(t, "init", oldDir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init old runtime: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "init", newDir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init new runtime: code=%d response=%+v", code, response)
	}
	release, err := acquireRuntimeGate(oldDir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	_, _, err = admitM11Recovery(newDir, oldDir, filepath.Join(t.TempDir(), "unread-handoff.json"), nil)
	if err == nil || !strings.Contains(err.Error(), "runtime is busy") {
		t.Fatalf("recovery admission read an old runtime while its writer gate was held: %v", err)
	}
	if _, err := os.Stat(m11ArtifactRegistryPath(newDir)); !os.IsNotExist(err) {
		t.Fatalf("blocked recovery admission mutated new runtime: %v", err)
	}
}

func TestM11RecoveryExportFailsClosedWhileRuntimeGateIsHeld(t *testing.T) {
	dir, output := t.TempDir(), filepath.Join(t.TempDir(), "handoff.json")
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init runtime: code=%d response=%+v", code, response)
	}
	release, err := acquireRuntimeGate(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if code, response := missionCall(t, "m11-recovery-export", dir, "unread-resolution", "unread-ledger", output); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("recovery export read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("blocked recovery export wrote a handoff: %v", err)
	}
}

func TestM11ReadCommandsFailClosedWhileRuntimeGateIsHeld(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init runtime: code=%d response=%+v", code, response)
	}
	_, ledgerEntry := setupM11OutcomeJournalFixture(t, dir)
	release, err := acquireRuntimeGate(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if code, response := missionCall(t, "m11-resolve", dir, corem11.ArtifactKindLedger, ledgerEntry.ArtifactID); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("M11 resolver read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "status", dir); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("mission status read a runtime while its writer gate was held: code=%d response=%+v", code, response)
	}
}

// A journal may survive a process exit after it has been synced but before the
// corresponding registry/state transition is complete. A fresh Bot process
// must expose only RECOVERY_REQUIRED until its locked writer prelude replays
// the exact journal; a read-only status call must never perform that replay.
func TestM11InterruptedJournalFreshProcessFailsClosedThenLockedWriterRecovers(t *testing.T) {
	binary := buildMissionBinary(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	t.Run("unknown_stop", func(t *testing.T) {
		fixture := newM11UnknownStopFixture(t)
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
				return errors.New("injected interrupted UNKNOWN ledger write")
			}
			return nil
		}
		if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fresh-process fixture"); err == nil {
			t.Fatal("interrupted UNKNOWN transition was not surfaced")
		}
		m11RegistryAppendFault = nil
		if code, response := missionBinaryCall(t, binary, "mission", "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed UNKNOWN transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", fixture.dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "STOPPED" {
			t.Fatalf("locked writer did not recover then retain STOP: code=%d response=%+v", code, response)
		}
		assertM11UnknownStopRecovered(t, fixture, attemptedAt, "fresh-process fixture")
	})
	t.Run("failed_execution", func(t *testing.T) {
		fixture := newM11UnknownStopFixture(t)
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
				return errors.New("injected interrupted FAILED ledger write")
			}
			return nil
		}
		if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fresh-process fixture"); err == nil {
			t.Fatal("interrupted FAILED transition was not surfaced")
		}
		m11RegistryAppendFault = nil
		if code, response := missionBinaryCall(t, binary, "mission", "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed FAILED transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", fixture.dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "INPUT_ERROR" {
			t.Fatalf("locked writer did not recover FAILED journal before input handling: code=%d response=%+v", code, response)
		}
		if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
			t.Fatalf("fresh-process FAILED recovery left journal: %v", err)
		}
		_, ledger, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
		if err != nil || ledger.ConsecutiveFailures != 1 || ledger.LastExecutionAt != attemptedAt {
			t.Fatalf("fresh-process FAILED recovery did not preserve exact ledger: ledger=%+v err=%v", ledger, err)
		}
	})
	t.Run("fixture_outcome", func(t *testing.T) {
		dir := t.TempDir()
		if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
			t.Fatalf("init failed: code=%d response=%+v", code, response)
		}
		journal, _ := setupM11OutcomeJournalFixture(t, dir)
		if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
			t.Fatal(err)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "status", dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
			t.Fatalf("fresh process exposed outcome transition: code=%d response=%+v", code, response)
		}
		if code, response := missionBinaryCall(t, binary, "mission", "m11-register", dir, corem11.ArtifactKindLease, filepath.Join(t.TempDir(), "unread.json")); code == 0 || response["status"] != "INPUT_ERROR" {
			t.Fatalf("locked writer did not recover outcome journal before input handling: code=%d response=%+v", code, response)
		}
		if _, err := os.Stat(m11OutcomeJournalPath(dir)); !os.IsNotExist(err) {
			t.Fatalf("fresh-process outcome recovery left journal: %v", err)
		}
		outcomes, err := loadM11FixtureOutcomes(dir)
		if err != nil || len(outcomes) != 1 || outcomes[0].OutcomeID != journal.Outcome.OutcomeID {
			t.Fatalf("fresh-process outcome recovery was not exact: outcomes=%+v err=%v", outcomes, err)
		}
	})
}

// TestM11ProcessTerminationChild is a test-binary-only entrypoint. It kills
// the child at a selected M11 journal boundary, exercising the real process
// boundary rather than only an in-process error return.
func TestM11ProcessTerminationChild(t *testing.T) {
	if os.Getenv("GO_WANT_M11_PROCESS_TERMINATION") != "1" {
		return
	}
	mode := os.Getenv("GO_M11_PROCESS_TERMINATION_MODE")
	if mode != "exit" && mode != "kill" && mode != "after-ledger-kill" && mode != "outcome-after-append-kill" {
		os.Exit(2)
	}
	separator := -1
	for index, value := range os.Args {
		if value == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		os.Exit(2)
	}
	terminate := func() {
		if mode == "kill" || mode == "after-ledger-kill" || mode == "outcome-after-append-kill" {
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				os.Exit(98)
			}
			if err := process.Kill(); err != nil {
				os.Exit(98)
			}
			// Keep the test child at the injected boundary if signal delivery is
			// deferred until the next scheduling point. SIGKILL is non-catchable;
			// this loop is only a defensive fallback against crossing the cleanup
			// code before the kernel terminates the process.
			for {
				runtime.Gosched()
			}
		}
		os.Exit(97)
	}
	if mode == "outcome-after-append-kill" {
		m11OutcomeAppendFault = func(phase string) error {
			if phase == "after_append" {
				terminate()
			}
			return nil
		}
	} else if mode == "after-ledger-kill" {
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if phase == "after_sync" && entry.ArtifactKind == corem11.ArtifactKindLedger {
				terminate()
			}
			return nil
		}
	} else {
		m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
			if phase != "before_write" || entry.ArtifactKind != corem11.ArtifactKindLedger {
				return nil
			}
			terminate()
			return nil
		}
	}
	os.Exit(runMissionCommand(os.Args[separator+1:], os.Stdout, os.Stderr))
}

func TestM11ProcessKillAfterJournalBeforeLedgerAppendLeavesJournalForFreshRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	for _, operation := range []string{"failed", "unknown"} {
		t.Run(operation, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			attemptedAt := "2026-09-08T00:00:02Z"
			args := []string{}
			if operation == "failed" {
				args = append(args, "m11-record-failed", fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "process kill failed fixture")
			} else {
				args = append(args, "m11-record-unknown", fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "process kill unknown fixture")
			}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM11ProcessTerminationChild$", "--"}, args...)...)
			command.Env = append(os.Environ(), "GO_WANT_M11_PROCESS_TERMINATION=1", "GO_M11_PROCESS_TERMINATION_MODE=kill")
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			if err := command.Run(); err == nil {
				t.Fatal("M11 child unexpectedly completed after injected SIGKILL")
			} else {
				exited, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
				if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
					t.Fatalf("M11 child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
				}
			}
			var journalPath string
			if operation == "failed" {
				journalPath = m11FailedExecutionJournalPath(fixture.dir)
			} else {
				journalPath = m11UnknownStopJournalPath(fixture.dir)
			}
			if _, err := os.Stat(journalPath); err != nil {
				entries, readErr := os.ReadDir(fixture.dir)
				names := make([]string, 0, len(entries))
				for _, entry := range entries {
					names = append(names, entry.Name())
				}
				t.Fatalf("process kill did not leave the recovery journal: %v entries=%v read_err=%v", err, names, readErr)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh process exposed the interrupted M11 transition: code=%d response=%+v", code, response)
			}
			missingInput := filepath.Join(t.TempDir(), "missing-artifact.json")
			if code, response := missionCall(t, "m11-register", fixture.dir, corem11.ArtifactKindLease, missingInput); operation == "failed" {
				if code == 0 || response["status"] != "INPUT_ERROR" {
					t.Fatalf("locked writer did not recover FAILED journal before input handling: code=%d response=%+v", code, response)
				}
			} else if code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("locked writer did not recover UNKNOWN STOP journal before retaining STOP: code=%d response=%+v", code, response)
			}
			if operation == "failed" {
				if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
					t.Fatalf("recovered FAILED journal remains: %v", err)
				}
				_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
				if err != nil || head.ConsecutiveFailures != 1 || head.LastExecutionAt != attemptedAt || head.UpdatedAt != attemptedAt {
					t.Fatalf("recovered FAILED ledger mismatch: ledger=%+v err=%v", head, err)
				}
			} else {
				assertM11UnknownStopRecovered(t, fixture, attemptedAt, "process kill unknown fixture")
			}
		})
	}
}

// TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery exercises
// the later side of the M11 execution transition. The ledger append and its
// directory sync are visible, but the child dies before journal cleanup. A
// fresh reader must still fail closed, and a locked writer must recognize the
// exact execution/ledger pair without appending either side twice.
func TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("SIGKILL process-boundary proof is not portable on Windows")
	}
	for _, operation := range []string{"failed", "unknown"} {
		t.Run(operation, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			attemptedAt := "2026-09-08T00:00:02Z"
			args := []string{}
			if operation == "failed" {
				args = append(args, "m11-record-failed", fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "post-ledger process kill failed fixture")
			} else {
				args = append(args, "m11-record-unknown", fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "post-ledger process kill unknown fixture")
			}
			command := exec.Command(os.Args[0], append([]string{"-test.run=^TestM11ProcessTerminationChild$", "--"}, args...)...)
			command.Env = append(os.Environ(), "GO_WANT_M11_PROCESS_TERMINATION=1", "GO_M11_PROCESS_TERMINATION_MODE=after-ledger-kill")
			var childStdout, childStderr bytes.Buffer
			command.Stdout, command.Stderr = &childStdout, &childStderr
			if err := command.Run(); err == nil {
				t.Fatal("M11 child unexpectedly completed after injected post-ledger SIGKILL")
			} else {
				exited, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				status, ok := exited.ProcessState.Sys().(syscall.WaitStatus)
				if !ok || !status.Signaled() || status.Signal() != syscall.SIGKILL {
					t.Fatalf("M11 post-ledger child did not receive SIGKILL: %v stdout=%q stderr=%q", err, childStdout.String(), childStderr.String())
				}
			}
			var journalPath string
			if operation == "failed" {
				journalPath = m11FailedExecutionJournalPath(fixture.dir)
			} else {
				journalPath = m11UnknownStopJournalPath(fixture.dir)
			}
			if _, err := os.Stat(journalPath); err != nil {
				t.Fatalf("post-ledger process kill did not leave the recovery journal: %v", err)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("fresh process exposed the interrupted post-ledger transition: code=%d response=%+v", code, response)
			}
			// The child has already appended the execution and expected ledger.
			// Assert the visible state before recovery so this test cannot pass by
			// merely exercising the earlier pre-append branch.
			entries, err := loadM11ArtifactRegistry(fixture.dir)
			if err != nil {
				t.Fatalf("load post-ledger registry: %v", err)
			}
			if count := countM11Artifacts(entries, corem11.ArtifactKindExecution); count != 1 {
				t.Fatalf("post-ledger process kill left %d execution artifacts, want 1", count)
			}
			if count := countM11Artifacts(entries, corem11.ArtifactKindLedger); count != 3 {
				t.Fatalf("post-ledger process kill left %d ledger artifacts, want gate, predecessor and transition", count)
			}
			missingInput := filepath.Join(t.TempDir(), "missing-artifact.json")
			if code, response := missionCall(t, "m11-register", fixture.dir, corem11.ArtifactKindLease, missingInput); operation == "failed" {
				if code == 0 || response["status"] != "INPUT_ERROR" {
					t.Fatalf("locked writer did not recover post-ledger FAILED journal before input handling: code=%d response=%+v", code, response)
				}
			} else if code == 0 || response["status"] != "STOPPED" {
				t.Fatalf("locked writer did not recover post-ledger UNKNOWN STOP journal before retaining STOP: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(journalPath); !os.IsNotExist(err) {
				t.Fatalf("post-ledger recovery journal remains: %v", err)
			}
			entries, err = loadM11ArtifactRegistry(fixture.dir)
			if err != nil {
				t.Fatalf("reload post-ledger registry: %v", err)
			}
			if count := countM11Artifacts(entries, corem11.ArtifactKindExecution); count != 1 {
				t.Fatalf("post-ledger recovery duplicated execution artifact: count=%d", count)
			}
			if count := countM11Artifacts(entries, corem11.ArtifactKindLedger); count != 3 {
				t.Fatalf("post-ledger recovery duplicated ledger transition: count=%d", count)
			}
			if operation == "failed" {
				_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
				if err != nil || head.ConsecutiveFailures != 1 || head.LastExecutionAt != attemptedAt || head.UpdatedAt != attemptedAt {
					t.Fatalf("recovered post-ledger FAILED ledger mismatch: ledger=%+v err=%v", head, err)
				}
			} else {
				assertM11UnknownStopRecovered(t, fixture, attemptedAt, "post-ledger process kill unknown fixture")
			}
		})
	}
}

func countM11Artifacts(entries []corem11.ArtifactEntry, kind string) int {
	count := 0
	for _, entry := range entries {
		if entry.ArtifactKind == kind {
			count++
		}
	}
	return count
}

type m11UnknownStopFixture struct {
	dir           string
	lease         corem11.ProductionLease
	authorization corem11.ProductionExecutionAuthorization
	ledgerEntry   corem11.ArtifactEntry
}

func newM11UnknownStopFixture(t *testing.T) m11UnknownStopFixture {
	t.Helper()
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	lease := corem11.ProductionLease{LeaseID: "unknown-journal-lease", LeaseVersion: "v1", PolicyVersion: "policy-1", ApprovalRef: "unknown-journal-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "unknown-journal-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	approval := corem11.ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	health := corem11.ProductionHealthSnapshot{SnapshotID: "unknown-journal-health", LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ObservedAt: "2026-09-08T00:00:00Z", SourceRefs: []string{"fixture:health"}, DependencyState: "HEALTHY", TelemetryComplete: true, HashVersion: "go-json-v1"}
	health.SnapshotHash = corem11.ComputeProductionHealthHash(health)
	cost := corem10.TrustedCostBound{CostBoundID: "unknown-journal-cost", IntentID: "unknown-journal-intent", IntentHash: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", MaxCostMinor: 10, Currency: "USD", SourceRef: "fixture:cost", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: lease.CorrelationID, HashVersion: "go-json-v1"}
	cost.CostBoundHash = corem10.ComputeTrustedCostBoundHash(cost)
	executionID := "pending-identity-not-yet-derived"
	ledger := corem11.ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: "2026-09-08T00:00:00Z", ExecutionsTotal: 1, ExecutionsInWindow: 1, CostMinorTotal: 10, PendingOutcomes: 1, PendingExecutionIDs: []string{executionID}, SuccessfulIdempotencyKeys: []string{}, OutcomeLinks: []corem11.ProductionOutcomeLink{}, ReconciliationResolutionIDs: []string{}, UpdatedAt: "2026-09-08T00:00:01Z"}
	ledgerRaw, err := json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ledgerEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	gateLedger := ledger
	gateLedger.ExecutionsTotal, gateLedger.ExecutionsInWindow, gateLedger.CostMinorTotal, gateLedger.PendingOutcomes = 0, 0, 0, 0
	gateLedger.PendingExecutionIDs = []string{}
	gateLedger.UpdatedAt = "2026-09-08T00:00:00Z"
	gateLedgerRaw, err := json.Marshal(gateLedger)
	if err != nil {
		t.Fatal(err)
	}
	gateLedgerEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, gateLedgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	activation := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: "2026-09-08T00:00:00Z"}
	gate := corem11.ProductionGateDecision{LedgerArtifactID: gateLedgerEntry.ArtifactID, LedgerContentHash: gateLedgerEntry.ContentHash, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, RiskClass: "RISK0", HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "ALLOW_PRODUCTION", Reason: "fixture", EvaluatedAt: "2026-09-08T00:00:00Z"}
	gate.GateID = corem11.ComputeProductionGateID(lease, gate.IntentID, gate.IntentHash, health, cost, gateLedgerEntry, gate.EvaluatedAt)
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: corem11.ComputeProductionAuthorizationID(gate.GateID, "fixture_stub", "2026-09-08T00:00:00Z"), IntentID: cost.IntentID, IntentHash: cost.IntentHash, PolicyVersion: lease.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: "fixture_stub", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T00:01:00Z", IdempotencyKey: "unknown-journal-key", CorrelationID: cost.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	executionID = corem11.ComputeProductionExecutionID(authorization.AuthorizationID)
	ledger.PendingExecutionIDs = []string{executionID}
	ledgerRaw, err = json.Marshal(ledger)
	if err != nil {
		t.Fatal(err)
	}
	ledgerEntry, err = corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		kind  string
		value any
	}{{corem11.ArtifactKindLease, lease}, {corem11.ArtifactKindLeaseApproval, approval}, {corem11.ArtifactKindHealth, health}, {corem11.ArtifactKindCostBound, cost}, {corem11.ArtifactKindLedger, gateLedger}, {corem11.ArtifactKindActivation, activation}, {corem11.ArtifactKindGate, gate}, {corem11.ArtifactKindAuthorization, authorization}, {corem11.ArtifactKindLedger, ledger}} {
		registerM11TestArtifact(t, dir, item.kind, item.value)
	}
	return m11UnknownStopFixture{dir: dir, lease: lease, authorization: authorization, ledgerEntry: ledgerEntry}
}

func TestBackupRestoreRejectsExecutionMissingReservationLedger(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	execution := corem11.ProductionExecutionRecord{
		ExecutionID:                  corem11.ComputeProductionExecutionID(fixture.authorization.AuthorizationID),
		AuthorizationID:              fixture.authorization.AuthorizationID,
		ProductionLeaseID:            fixture.authorization.ProductionLeaseID,
		ProductionLeaseVersion:       fixture.authorization.ProductionLeaseVersion,
		ProductionLeaseHash:          fixture.authorization.ProductionLeaseHash,
		ProductionGateID:             fixture.authorization.ProductionGateID,
		ProductionHealthSnapshotID:   fixture.authorization.ProductionHealthSnapshotID,
		ProductionHealthSnapshotHash: fixture.authorization.ProductionHealthSnapshotHash,
		ProductionCostBoundID:        fixture.authorization.ProductionCostBoundID,
		ProductionCostBoundHash:      fixture.authorization.ProductionCostBoundHash,
		ProductionCostBoundMinor:     fixture.authorization.ProductionCostBoundMinor,
		IntentID:                     fixture.authorization.IntentID,
		IntentHash:                   fixture.authorization.IntentHash,
		ExecutorID:                   fixture.authorization.ExecutorID,
		IdempotencyKey:               fixture.authorization.IdempotencyKey,
		AttemptedAt:                  "2026-09-08T00:00:02Z",
		Status:                       "CANCELLED",
		SideEffectState:              "NOT_PERFORMED",
		CorrelationID:                fixture.authorization.CorrelationID,
	}
	registerM11TestArtifact(t, fixture.dir, corem11.ArtifactKindExecution, execution)
	if err := validateM11BackupGraph(fixture.dir); err != nil {
		t.Fatalf("valid reserved execution was rejected: %v", err)
	}

	raw, err := os.ReadFile(m11ArtifactRegistryPath(fixture.dir))
	if err != nil {
		t.Fatal(err)
	}
	lines := make([][]byte, 0)
	mutated := false
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, entryErr := corem11.ValidateArtifactEntry(line)
		if entryErr != nil {
			t.Fatal(entryErr)
		}
		if entry.ArtifactKind == corem11.ArtifactKindLedger && entry.ArtifactID == fixture.ledgerEntry.ArtifactID {
			value, status := corem11.DecodeArtifact("ledger", entry.Artifact)
			if status != corem11.Valid {
				t.Fatalf("fixture ledger is invalid: %s", status)
			}
			ledger := *value.(*corem11.ProductionLedger)
			ledger.PendingExecutionIDs = []string{}
			ledger.PendingOutcomes = 0
			ledgerRaw, marshalErr := json.Marshal(ledger)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			replacement, replacementErr := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
			if replacementErr != nil {
				t.Fatal(replacementErr)
			}
			line, err = json.Marshal(replacement)
			if err != nil {
				t.Fatal(err)
			}
			mutated = true
		}
		lines = append(lines, line)
	}
	if !mutated {
		t.Fatal("reservation ledger fixture was not found")
	}
	if err := os.WriteFile(m11ArtifactRegistryPath(fixture.dir), append(bytes.Join(lines, []byte{'\n'}), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := validateM11BackupGraph(fixture.dir); err == nil || !strings.Contains(err.Error(), "orphaned from its restored reservation ledger") {
		t.Fatalf("M11 reservation-ledger guard: checksum-valid orphan execution was accepted: %v", err)
	}
}

func TestM11UnknownStopJournalRecoversAfterStoppedLedgerWriteFailure(t *testing.T) {
	for _, phase := range []string{"before_write", "after_write", "after_sync"} {
		faultPhase := phase
		t.Run(phase, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
			m11RegistryAppendFault = func(registryPhase string, entry corem11.ArtifactEntry) error {
				if entry.ArtifactKind == corem11.ArtifactKindLedger && registryPhase == faultPhase {
					return errors.New("injected stopped-ledger write failure")
				}
				return nil
			}
			t.Cleanup(func() { m11RegistryAppendFault = nil })
			if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("stopped-ledger fault was not surfaced")
			}
			if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); err != nil {
				t.Fatalf("unknown STOP journal was removed before recovery: %v", err)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on unknown STOP journal: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-resolve", fixture.dir, corem11.ArtifactKindLedger, fixture.ledgerEntry.ArtifactID); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 resolver exposed an UNKNOWN-to-STOP partial transition: code=%d response=%+v", code, response)
			}
			exportPath := filepath.Join(fixture.dir, "partial-recovery-handoff.json")
			if code, response := missionCall(t, "m11-recovery-export", fixture.dir, "missing-resolution", fixture.ledgerEntry.ArtifactID, exportPath); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 recovery export exposed a partial transition: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(exportPath); !os.IsNotExist(err) {
				t.Fatalf("M11 recovery export wrote during pending journal recovery: %v", err)
			}
			admissionRuntime := t.TempDir()
			if code, response := missionCall(t, "init", admissionRuntime); code != 0 || response["status"] != "INITIALIZED" {
				t.Fatalf("init admission runtime: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-recovery-admit", admissionRuntime, fixture.dir, filepath.Join(t.TempDir(), "missing-handoff.json"), filepath.Join(t.TempDir(), "missing-admission.json")); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 recovery admission read an interrupted old runtime: code=%d response=%+v", code, response)
			}
			if _, err := os.Stat(m11ArtifactRegistryPath(admissionRuntime)); !os.IsNotExist(err) {
				t.Fatalf("M11 recovery admission wrote during pending old journal recovery: %v", err)
			}
			state, err := loadMissionState(fixture.dir)
			if faultPhase == "before_write" && (err != nil || state.Stop) {
				t.Fatalf("pre-write fault marked mission stopped: state=%+v err=%v", state, err)
			}
			if faultPhase == "after_write" && (err == nil || state.Stop) {
				t.Fatalf("post-write fault did not fail closed at the stopped-ledger recovery boundary: state=%+v err=%v", state, err)
			}
			m11RegistryAppendFault = nil
			if err := recoverM11UnknownStopJournal(fixture.dir); err != nil {
				t.Fatalf("unknown STOP journal recovery failed: %v", err)
			}
			assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
		})
	}
}

func TestM11UnknownStopJournalRecoversAfterDurableStopWriteFailure(t *testing.T) {
	for _, tc := range []struct {
		name           string
		failureAt      int
		stateCommitted bool
	}{
		{name: "mission_state", failureAt: 2},
		{name: "stop_marker", failureAt: 3, stateCommitted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
			atomicWrites := 0
			missionStateWriteFault = func(phase string) error {
				if phase != "before_rename" {
					return nil
				}
				atomicWrites++ // journal, mission-state after stopped ledger, then STOP marker
				if atomicWrites == tc.failureAt {
					return errors.New("injected durable STOP write failure")
				}
				return nil
			}
			t.Cleanup(func() { missionStateWriteFault = nil })
			if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("durable STOP write fault was not surfaced")
			}
			if atomicWrites != tc.failureAt {
				t.Fatalf("fault did not reach expected durable STOP write: atomic writes=%d", atomicWrites)
			}
			if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); err != nil {
				t.Fatalf("unknown STOP journal was removed before recovery: %v", err)
			}
			state, err := loadMissionState(fixture.dir)
			if tc.stateCommitted && (err != nil || !state.Stop) {
				t.Fatalf("STOP marker fault did not retain stopped mission state: state=%+v err=%v", state, err)
			}
			if !tc.stateCommitted && err == nil {
				t.Fatalf("mission-state fault did not fail closed before STOP: state=%+v", state)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on durable STOP write fault: code=%d response=%+v", code, response)
			}
			missionStateWriteFault = nil
			if err := recoverM11UnknownStopJournal(fixture.dir); err != nil {
				t.Fatalf("unknown STOP journal did not recover durable STOP write failure: %v", err)
			}
			assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
		})
	}
}

func TestM11FailedExecutionJournalRecoversAfterLedgerWriteFailure(t *testing.T) {
	for _, phase := range []string{"before_write", "after_write", "after_sync"} {
		faultPhase := phase
		t.Run(phase, func(t *testing.T) {
			fixture := newM11UnknownStopFixture(t)
			const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture execution failure"
			m11RegistryAppendFault = func(registryPhase string, entry corem11.ArtifactEntry) error {
				if entry.ArtifactKind == corem11.ArtifactKindLedger && registryPhase == faultPhase {
					return errors.New("injected failed-execution ledger write failure")
				}
				return nil
			}
			t.Cleanup(func() { m11RegistryAppendFault = nil })
			if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
				t.Fatal("failed-execution ledger fault was not surfaced")
			}
			if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); err != nil {
				t.Fatalf("failed execution journal was removed before recovery: %v", err)
			}
			if code, response := missionCall(t, "status", fixture.dir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("status did not fail closed on failed-execution journal: code=%d response=%+v", code, response)
			}
			if code, response := missionCall(t, "m11-resolve", fixture.dir, corem11.ArtifactKindLedger, fixture.ledgerEntry.ArtifactID); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
				t.Fatalf("M11 resolver exposed a FAILED partial transition: code=%d response=%+v", code, response)
			}
			m11RegistryAppendFault = nil
			if err := recoverM11FailedExecutionJournal(fixture.dir); err != nil {
				t.Fatalf("failed-execution journal recovery failed: %v", err)
			}
			if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
				t.Fatalf("failed execution journal remains after recovery: %v", err)
			}
			_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
			if err != nil || head.ConsecutiveFailures != 1 || head.LastExecutionAt != attemptedAt || head.UpdatedAt != attemptedAt {
				t.Fatalf("failed-execution ledger transition was not recovered: ledger=%+v err=%v", head, err)
			}
			record, ledger, status, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason)
			if err != nil || status != appendDuplicate || record.Status != "FAILED" || ledger.UpdatedAt != attemptedAt || ledger.ConsecutiveFailures != 1 {
				t.Fatalf("failed-execution retry was not exact after recovery: record=%+v ledger=%+v status=%s err=%v", record, ledger, status, err)
			}
		})
	}
}

func TestM11ReconcileRejectsSecondResolutionForSameUnknownExecution(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	execution, _, status, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fixture timeout")
	if err != nil || status != appendAdded {
		t.Fatalf("create stopped UNKNOWN fixture: execution=%+v status=%s err=%v", execution, status, err)
	}
	stoppedEntry, _, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve stopped ledger: %v", err)
	}
	first := corem11.ProductionReconciliationResolution{ResolutionID: "resolution-first", LeaseID: fixture.lease.LeaseID, LeaseVersion: fixture.lease.LeaseVersion, LeaseHash: fixture.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:03Z", EffectState: "NOT_PERFORMED", Reason: "human reviewed fixture timeout"}
	registerM11TestArtifact(t, fixture.dir, corem11.ArtifactKindReconciliation, first)
	if resolution, ledger, status, err := reconcileM11Execution(fixture.dir, first.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendAdded || len(ledger.ReconciliationResolutionIDs) != 1 || ledger.ReconciliationResolutionIDs[0] != first.ResolutionID || resolution.ResolutionID != first.ResolutionID {
		t.Fatalf("first reconciliation did not advance the stopped ledger: resolution=%+v ledger=%+v status=%s err=%v", resolution, ledger, status, err)
	}
	if _, ledger, status, err := reconcileM11Execution(fixture.dir, first.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendDuplicate || len(ledger.ReconciliationResolutionIDs) != 1 || ledger.ReconciliationResolutionIDs[0] != first.ResolutionID {
		t.Fatalf("exact reconciliation retry did not resolve from the new head: ledger=%+v status=%s err=%v", ledger, status, err)
	}
	second := corem11.ProductionReconciliationResolution{ResolutionID: "resolution-stale", LeaseID: fixture.lease.LeaseID, LeaseVersion: fixture.lease.LeaseVersion, LeaseHash: fixture.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-2", ResolvedAt: "2026-09-08T00:00:04Z", EffectState: "NOT_PERFORMED", Reason: "attempt stale fork"}
	secondRaw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM11Artifact(fixture.dir, corem11.ArtifactKindReconciliation, secondRaw); err == nil {
		t.Fatal("second reconciliation resolution for one UNKNOWN execution was registered")
	}
	_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil || len(head.ReconciliationResolutionIDs) != 1 || head.ReconciliationResolutionIDs[0] != first.ResolutionID || head.UpdatedAt != first.ResolvedAt {
		t.Fatalf("rejected reconciliation changed the current ledger head: ledger=%+v err=%v", head, err)
	}
}

func TestMissionM11ReconcileDisclosesRegistryPublishUncertainty(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	execution, _, status, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, "fixture timeout")
	if err != nil || status != appendAdded {
		t.Fatalf("create stopped UNKNOWN fixture: execution=%+v status=%s err=%v", execution, status, err)
	}
	stoppedEntry, _, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve stopped ledger: %v", err)
	}
	resolution := corem11.ProductionReconciliationResolution{ResolutionID: "uncertain-reconciliation", LeaseID: fixture.lease.LeaseID, LeaseVersion: fixture.lease.LeaseVersion, LeaseHash: fixture.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:03Z", EffectState: "NOT_PERFORMED", Reason: "human reviewed fixture timeout"}
	registerM11TestArtifact(t, fixture.dir, corem11.ArtifactKindReconciliation, resolution)
	artifactRegistryPublishFailure = func(path string) error {
		if filepath.Clean(path) == filepath.Clean(m11ArtifactRegistryPath(fixture.dir)) {
			return errors.New("injected registry parent sync failure")
		}
		return nil
	}
	t.Cleanup(func() { artifactRegistryPublishFailure = nil })
	if code, response := missionCall(t, "m11-reconcile", fixture.dir, resolution.ResolutionID, stoppedEntry.ArtifactID); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("reconciliation uncertainty was not disclosed: code=%d response=%+v", code, response)
	} else {
		artifact, ok := response["artifact"].(map[string]any)
		if !ok {
			t.Fatalf("reconciliation uncertainty omitted visible transition: %+v", response)
		}
		returnedResolution, hasResolution := artifact["resolution"].(map[string]any)
		returnedLedger, hasLedger := artifact["stopped_ledger"].(map[string]any)
		if !hasResolution || !hasLedger || returnedResolution["resolution_id"] != resolution.ResolutionID || returnedLedger["reconciliation_required"] != false {
			t.Fatalf("reconciliation uncertainty did not disclose the canonical transition: %+v", artifact)
		}
	}
	if _, ledger, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID); err != nil || len(ledger.ReconciliationResolutionIDs) != 1 || ledger.ReconciliationResolutionIDs[0] != resolution.ResolutionID {
		t.Fatalf("uncertain reconciliation ledger was not retained as head: ledger=%+v err=%v", ledger, err)
	}
	artifactRegistryPublishFailure = nil
	if code, response := missionCall(t, "m11-reconcile", fixture.dir, resolution.ResolutionID, stoppedEntry.ArtifactID); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("reconciliation exact retry failed: code=%d response=%+v", code, response)
	}
}

func TestMissionM11RecoveryAdmissionDisclosesRegistryPublishUncertainty(t *testing.T) {
	oldRuntime := newM11UnknownStopFixture(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	execution, _, status, err := recordUnknownM11Execution(oldRuntime.dir, oldRuntime.authorization.AuthorizationID, oldRuntime.ledgerEntry.ArtifactID, attemptedAt, "fixture timeout")
	if err != nil || status != appendAdded {
		t.Fatalf("create stopped UNKNOWN fixture: execution=%+v status=%s err=%v", execution, status, err)
	}
	stoppedEntry, _, err := m11LedgerHead(oldRuntime.dir, oldRuntime.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve stopped ledger: %v", err)
	}
	resolution := corem11.ProductionReconciliationResolution{ResolutionID: "admission-uncertain-resolution", LeaseID: oldRuntime.lease.LeaseID, LeaseVersion: oldRuntime.lease.LeaseVersion, LeaseHash: oldRuntime.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:03Z", EffectState: "NOT_PERFORMED", Reason: "human reviewed fixture timeout"}
	registerM11TestArtifact(t, oldRuntime.dir, corem11.ArtifactKindReconciliation, resolution)
	if _, _, status, err := reconcileM11Execution(oldRuntime.dir, resolution.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendAdded {
		t.Fatalf("reconcile old runtime: status=%s err=%v", status, err)
	}
	_, reviewedLedger, err := m11LedgerHead(oldRuntime.dir, oldRuntime.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve reviewed stopped ledger: %v", err)
	}
	reviewedRaw, err := json.Marshal(reviewedLedger)
	if err != nil {
		t.Fatal(err)
	}
	reviewedEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, reviewedRaw)
	if err != nil {
		t.Fatal(err)
	}
	handoffPath := filepath.Join(t.TempDir(), "recovery-handoff.json")
	artifactWriteFault = func(phase string) error {
		if phase == "after_publish_before_parent_sync" {
			return errors.New("injected handoff parent sync failure")
		}
		return nil
	}
	t.Cleanup(func() { artifactWriteFault = nil })
	if code, response := missionCall(t, "m11-recovery-export", oldRuntime.dir, resolution.ResolutionID, reviewedEntry.ArtifactID, handoffPath); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("recovery handoff uncertainty was not disclosed: code=%d response=%+v", code, response)
	} else if artifact, ok := response["artifact"].(map[string]any); !ok || artifact["resolution_id"] != resolution.ResolutionID || artifact["execution_permitted"] != false {
		t.Fatalf("recovery handoff uncertainty omitted visible non-authorizing handoff: %+v", response)
	}
	if _, err := os.Stat(handoffPath); err != nil {
		t.Fatalf("recovery handoff was not visible after unconfirmed publish: %v", err)
	}
	artifactWriteFault = nil
	if code, response := missionCall(t, "m11-recovery-export", oldRuntime.dir, resolution.ResolutionID, reviewedEntry.ArtifactID, handoffPath); code != 0 || response["status"] != appendDuplicate {
		t.Fatalf("recovery handoff exact retry failed: code=%d response=%+v", code, response)
	}

	newRuntime := t.TempDir()
	if code, response := missionCall(t, "init", newRuntime); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init new runtime: code=%d response=%+v", code, response)
	}
	newLease := oldRuntime.lease
	newLease.LeaseID = "admission-uncertain-new-lease"
	newLease.ApprovalRef = "admission-uncertain-new-approval"
	newLease.ReviewedAt = "2026-09-08T00:00:04Z"
	newLease.ValidFrom = "2026-09-08T00:00:04Z"
	newLease.CorrelationID = "admission-uncertain-new-correlation"
	newLease.LeaseHash = corem11.ComputeProductionLeaseHash(newLease)
	newApproval := corem11.ProductionLeaseApproval{ApprovalID: newLease.ApprovalRef, LeaseID: newLease.LeaseID, LeaseVersion: newLease.LeaseVersion, LeaseHash: newLease.LeaseHash, PromotionReviewRef: newLease.PromotionReviewRef, SourceCanaryGrantID: newLease.SourceCanaryGrantID, SourceCanaryGrantVersion: newLease.SourceCanaryGrantVersion, SourceCanaryGrantHash: newLease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: newLease.ReviewerID, ReviewedAt: newLease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	registerM11TestArtifact(t, newRuntime, corem11.ArtifactKindLease, newLease)
	registerM11TestArtifact(t, newRuntime, corem11.ArtifactKindLeaseApproval, newApproval)
	if code, response := missionCall(t, "m11-activate", newRuntime, newLease.LeaseID, "2026-09-08T00:00:05Z"); code != 0 || response["status"] != appendAdded {
		t.Fatalf("activate new runtime: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-ledger-init", newRuntime, newLease.LeaseID, "2026-09-08T00:00:05Z"); code != 0 || response["status"] != appendAdded {
		t.Fatalf("initialize new runtime ledger: code=%d response=%+v", code, response)
	}
	oldAbs, newAbs, err := m11DistinctRuntimeDirs(oldRuntime.dir, newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	admission := corem11.ProductionRecoveryAdmission{RecoveryAdmissionID: "uncertain-recovery-admission", PriorRuntimeDir: oldAbs, PriorLeaseID: oldRuntime.lease.LeaseID, PriorLeaseVersion: oldRuntime.lease.LeaseVersion, PriorLeaseHash: oldRuntime.lease.LeaseHash, PriorApprovalID: oldRuntime.lease.ApprovalRef, ResolutionID: resolution.ResolutionID, NewRuntimeID: "admission-uncertain-runtime", NewRuntimeDir: newAbs, NewLeaseID: newLease.LeaseID, NewLeaseVersion: newLease.LeaseVersion, NewLeaseHash: newLease.LeaseHash, NewApprovalID: newApproval.ApprovalID, ReviewedBy: "human", ReviewerID: "reviewer-2", ReviewedAt: "2026-09-08T00:00:06Z", ExecutionPermitted: false}
	admissionRaw, err := json.Marshal(admission)
	if err != nil {
		t.Fatal(err)
	}
	admissionPath := filepath.Join(t.TempDir(), "recovery-admission.json")
	if err := writeJSONAtomic(admissionPath, json.RawMessage(admissionRaw)); err != nil {
		t.Fatal(err)
	}
	// Handoff is portable input, not a runtime artifact. A duplicate key with
	// the same apparent value must be rejected before it can collapse through
	// a generic JSON map and register a recovery admission.
	originalHandoff, err := os.ReadFile(handoffPath)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Replace(originalHandoff, []byte(`"profile": "M11_RECOVERY_HANDOFF/v1"`), []byte(`"profile": "M11_RECOVERY_HANDOFF/v1", "profile": "M11_RECOVERY_HANDOFF/v1"`), 1)
	if bytes.Equal(duplicate, originalHandoff) {
		t.Fatal("could not construct duplicate-key recovery handoff")
	}
	registryBefore, err := os.ReadFile(m11ArtifactRegistryPath(newRuntime))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(handoffPath, duplicate, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-recovery-admit", newRuntime, oldRuntime.dir, handoffPath, admissionPath); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("duplicate-key recovery handoff reached admission: code=%d response=%+v", code, response)
	}
	registryAfter, err := os.ReadFile(m11ArtifactRegistryPath(newRuntime))
	if err != nil || !bytes.Equal(registryBefore, registryAfter) {
		t.Fatalf("duplicate-key recovery handoff mutated new registry: err=%v before=%q after=%q", err, registryBefore, registryAfter)
	}
	if err := os.WriteFile(handoffPath, originalHandoff, 0600); err != nil {
		t.Fatal(err)
	}
	artifactRegistryPublishFailure = func(path string) error {
		if filepath.Clean(path) == filepath.Clean(m11ArtifactRegistryPath(newRuntime)) {
			return errors.New("injected registry parent sync failure")
		}
		return nil
	}
	t.Cleanup(func() { artifactRegistryPublishFailure = nil })
	if code, response := missionCall(t, "m11-recovery-admit", newRuntime, oldRuntime.dir, handoffPath, admissionPath); code == 0 || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" {
		t.Fatalf("recovery admission uncertainty was not disclosed: code=%d response=%+v", code, response)
	} else if artifact, ok := response["artifact"].(map[string]any); !ok || artifact["recovery_admission_id"] != admission.RecoveryAdmissionID || artifact["execution_permitted"] != false {
		t.Fatalf("recovery admission uncertainty omitted the visible non-authorizing artifact: %+v", response)
	}
	if code, response := missionCall(t, "m11-resolve", newRuntime, corem11.ArtifactKindRecoveryAdmission, admission.RecoveryAdmissionID); code != 0 || response["status"] != "RESOLVED" {
		t.Fatalf("uncertain recovery admission was not resolvable: code=%d response=%+v", code, response)
	}
	artifactRegistryPublishFailure = nil
	if code, response := missionCall(t, "m11-recovery-admit", newRuntime, oldRuntime.dir, handoffPath, admissionPath); code != 0 || response["status"] != appendDuplicate {
		t.Fatalf("recovery admission exact retry failed: code=%d response=%+v", code, response)
	}
}

// Recovery admission is an immutable link, but it still shares the new
// runtime registry with every other writer. Run real Bot processes against
// one new runtime to prove concurrent admission cannot append the same link
// twice or let a second caller observe a partially published graph.
func TestMissionM11RecoveryAdmissionSingleWriterAcrossBotProcesses(t *testing.T) {
	oldRuntime := newM11UnknownStopFixture(t)
	const attemptedAt = "2026-09-08T00:00:02Z"
	execution, _, status, err := recordUnknownM11Execution(oldRuntime.dir, oldRuntime.authorization.AuthorizationID, oldRuntime.ledgerEntry.ArtifactID, attemptedAt, "concurrent recovery fixture")
	if err != nil || status != appendAdded {
		t.Fatalf("create stopped UNKNOWN fixture: status=%s err=%v", status, err)
	}
	stoppedEntry, _, err := m11LedgerHead(oldRuntime.dir, oldRuntime.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve stopped ledger: %v", err)
	}
	resolution := corem11.ProductionReconciliationResolution{ResolutionID: "concurrent-recovery-resolution", LeaseID: oldRuntime.lease.LeaseID, LeaseVersion: oldRuntime.lease.LeaseVersion, LeaseHash: oldRuntime.lease.LeaseHash, ExecutionID: execution.ExecutionID, ResolvedBy: "human", ResolverID: "reviewer-1", ResolvedAt: "2026-09-08T00:00:03Z", EffectState: "NOT_PERFORMED", Reason: "human reviewed concurrent fixture timeout"}
	registerM11TestArtifact(t, oldRuntime.dir, corem11.ArtifactKindReconciliation, resolution)
	if _, _, status, err := reconcileM11Execution(oldRuntime.dir, resolution.ResolutionID, stoppedEntry.ArtifactID); err != nil || status != appendAdded {
		t.Fatalf("reconcile old runtime: status=%s err=%v", status, err)
	}
	_, reviewedLedger, err := m11LedgerHead(oldRuntime.dir, oldRuntime.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve reviewed stopped ledger: %v", err)
	}
	reviewedRaw, err := json.Marshal(reviewedLedger)
	if err != nil {
		t.Fatal(err)
	}
	reviewedEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, reviewedRaw)
	if err != nil {
		t.Fatal(err)
	}
	handoff, err := m11RecoveryHandoff(oldRuntime.dir, resolution.ResolutionID, reviewedEntry.ArtifactID)
	if err != nil {
		t.Fatalf("build recovery handoff: %v", err)
	}
	handoffPath := filepath.Join(t.TempDir(), "concurrent-recovery-handoff.json")
	handoffRaw, err := json.Marshal(handoff)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(handoffPath, json.RawMessage(handoffRaw)); err != nil {
		t.Fatal(err)
	}

	newRuntime := t.TempDir()
	if code, response := missionCall(t, "init", newRuntime); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init new runtime: code=%d response=%+v", code, response)
	}
	newLease := oldRuntime.lease
	newLease.LeaseID = "concurrent-recovery-new-lease"
	newLease.ApprovalRef = "concurrent-recovery-new-approval"
	newLease.ReviewedAt = "2026-09-08T00:00:04Z"
	newLease.ValidFrom = "2026-09-08T00:00:04Z"
	newLease.CorrelationID = "concurrent-recovery-new-correlation"
	newLease.LeaseHash = corem11.ComputeProductionLeaseHash(newLease)
	newApproval := corem11.ProductionLeaseApproval{ApprovalID: newLease.ApprovalRef, LeaseID: newLease.LeaseID, LeaseVersion: newLease.LeaseVersion, LeaseHash: newLease.LeaseHash, PromotionReviewRef: newLease.PromotionReviewRef, SourceCanaryGrantID: newLease.SourceCanaryGrantID, SourceCanaryGrantVersion: newLease.SourceCanaryGrantVersion, SourceCanaryGrantHash: newLease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: newLease.ReviewerID, ReviewedAt: newLease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	registerM11TestArtifact(t, newRuntime, corem11.ArtifactKindLease, newLease)
	registerM11TestArtifact(t, newRuntime, corem11.ArtifactKindLeaseApproval, newApproval)
	if code, response := missionCall(t, "m11-activate", newRuntime, newLease.LeaseID, "2026-09-08T00:00:05Z"); code != 0 || response["status"] != appendAdded {
		t.Fatalf("activate new runtime: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-ledger-init", newRuntime, newLease.LeaseID, "2026-09-08T00:00:05Z"); code != 0 || response["status"] != appendAdded {
		t.Fatalf("initialize new runtime ledger: code=%d response=%+v", code, response)
	}
	oldAbs, newAbs, err := m11DistinctRuntimeDirs(oldRuntime.dir, newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	admission := corem11.ProductionRecoveryAdmission{RecoveryAdmissionID: "concurrent-recovery-admission", PriorRuntimeDir: oldAbs, PriorLeaseID: oldRuntime.lease.LeaseID, PriorLeaseVersion: oldRuntime.lease.LeaseVersion, PriorLeaseHash: oldRuntime.lease.LeaseHash, PriorApprovalID: oldRuntime.lease.ApprovalRef, ResolutionID: resolution.ResolutionID, NewRuntimeID: "concurrent-recovery-runtime-v2", NewRuntimeDir: newAbs, NewLeaseID: newLease.LeaseID, NewLeaseVersion: newLease.LeaseVersion, NewLeaseHash: newLease.LeaseHash, NewApprovalID: newApproval.ApprovalID, ReviewedBy: "human", ReviewerID: "reviewer-2", ReviewedAt: "2026-09-08T00:00:06Z", ExecutionPermitted: false}
	admissionRaw, err := json.Marshal(admission)
	if err != nil {
		t.Fatal(err)
	}
	admissionPath := filepath.Join(t.TempDir(), "concurrent-recovery-admission.json")
	if err := writeJSONAtomic(admissionPath, json.RawMessage(admissionRaw)); err != nil {
		t.Fatal(err)
	}

	binary := buildMissionBinary(t)
	const contenders = 8
	root := t.TempDir()
	readyDir := filepath.Join(root, "recovery-ready")
	startPath := filepath.Join(root, "recovery-start")
	if err := os.Mkdir(readyDir, 0700); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results := make(chan missionProcessResult, contenders)
	for index := 0; index < contenders; index++ {
		command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMissionProcessBarrierHelper$", "--", binary, "mission", "m11-recovery-admit", newRuntime, oldRuntime.dir, handoffPath, admissionPath)
		command.Env = append(os.Environ(), "GO_WANT_MISSION_BARRIER=1", "GO_MISSION_READY_DIR="+readyDir, "GO_MISSION_START_PATH="+startPath)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		id := "recovery-contender-" + strconv.Itoa(index)
		go func(id string, command *exec.Cmd, stdout, stderr *bytes.Buffer) {
			err := command.Wait()
			code := 0
			if exited, ok := err.(*exec.ExitError); ok {
				code = exited.ExitCode()
			}
			results <- missionProcessResult{id: id, code: code, stdout: stdout.String(), stderr: stderr.String(), err: err}
		}(id, command, &stdout, &stderr)
	}
	waitForMissionContenders(t, readyDir, contenders)
	if err := os.WriteFile(startPath, []byte("start"), 0600); err != nil {
		t.Fatal(err)
	}
	appended := 0
	for index := 0; index < contenders; index++ {
		result := <-results
		var response map[string]any
		if err := json.Unmarshal([]byte(result.stdout), &response); err != nil {
			t.Fatalf("recovery contender %s emitted invalid response %q: %v (stderr: %s, command error: %v)", result.id, result.stdout, err, result.stderr, result.err)
		}
		status, _ := response["status"].(string)
		switch status {
		case appendAdded:
			if result.code != 0 {
				t.Fatalf("successful recovery contender %s returned exit %d: %s", result.id, result.code, result.stderr)
			}
			appended++
		case appendDuplicate, "BUSY":
			if status == appendDuplicate && result.code != 0 {
				t.Fatalf("exact recovery contender %s returned exit %d: %s", result.id, result.code, result.stderr)
			}
			if status == "BUSY" && result.code == 0 {
				t.Fatalf("busy recovery contender %s returned success: %+v", result.id, response)
			}
		default:
			t.Fatalf("recovery contender %s returned unexpected status %q: response=%+v stderr=%s command error=%v", result.id, status, response, result.stderr, result.err)
		}
	}
	if appended != 1 {
		t.Fatalf("concurrent recovery admitted %d immutable links across %d Bot processes", appended, contenders)
	}
	entries, err := loadM11ArtifactRegistry(newRuntime)
	if err != nil {
		t.Fatal(err)
	}
	admissions := 0
	for _, entry := range entries {
		if entry.ArtifactKind == corem11.ArtifactKindRecoveryAdmission {
			admissions++
		}
	}
	if admissions != 1 {
		t.Fatalf("new runtime persisted %d recovery admissions, want exactly one", admissions)
	}
	if code, response := missionBinaryCall(t, binary, "mission", "m11-recovery-admit", newRuntime, oldRuntime.dir, handoffPath, admissionPath); code != 0 || response["status"] != appendDuplicate {
		t.Fatalf("fresh exact retry did not resolve to duplicate: code=%d response=%+v", code, response)
	}
}

func TestBackupCreateRecoversPendingM11UnknownStopJournal(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture timeout"
	m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
		if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
			return errors.New("injected stopped-ledger write failure")
		}
		return nil
	}
	if _, _, _, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
		t.Fatal("stopped-ledger fault was not surfaced")
	}
	m11RegistryAppendFault = nil
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	if _, err := buildBR10AdvisorFixture(fixture.dir); err != nil {
		t.Fatalf("build backup history fixture: %v", err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(fixture.dir, name)); err != nil {
			t.Fatalf("remove fixture-only backup input %s: %v", name, err)
		}
	}
	backup, restored := filepath.Join(t.TempDir(), "backup"), filepath.Join(t.TempDir(), "restored")
	if code, response := backupCall(t, "create", fixture.dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup create did not recover pending UNKNOWN STOP journal: code=%d response=%+v", code, response)
	}
	assertM11UnknownStopRecovered(t, fixture, attemptedAt, reason)
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("recovered UNKNOWN STOP backup did not restore: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "status", restored); code != 0 || response["artifact"].(map[string]any)["stop"] != true {
		t.Fatalf("restored UNKNOWN STOP runtime was not durably stopped: code=%d response=%+v", code, response)
	}
}

func TestBackupCreateRecoversPendingM11FailedExecutionJournal(t *testing.T) {
	fixture := newM11UnknownStopFixture(t)
	if _, err := buildBR10AdvisorFixture(fixture.dir); err != nil {
		t.Fatalf("build backup history fixture: %v", err)
	}
	const attemptedAt, reason = "2026-09-08T00:00:02Z", "fixture execution failure"
	m11RegistryAppendFault = func(phase string, entry corem11.ArtifactEntry) error {
		if entry.ArtifactKind == corem11.ArtifactKindLedger && phase == "before_write" {
			return errors.New("injected failed-execution ledger write failure")
		}
		return nil
	}
	if _, _, _, err := recordFailedM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err == nil {
		t.Fatal("failed-execution ledger fault was not surfaced")
	}
	m11RegistryAppendFault = nil
	t.Cleanup(func() { m11RegistryAppendFault = nil })
	backup, restored := filepath.Join(t.TempDir(), "backup"), filepath.Join(t.TempDir(), "restored")
	if code, response := backupCall(t, "create", fixture.dir, backup); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a recovered FAILED execution without an outcome: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11FailedExecutionJournalPath(fixture.dir)); !os.IsNotExist(err) {
		t.Fatalf("backup did not recover failed-execution journal: %v", err)
	}
	postEntry, _, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil {
		t.Fatalf("resolve recovered post-execution ledger: %v", err)
	}
	outcome := m03.OutcomeRecord{OutcomeID: "backup-recovered-failed-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: "prod-exec-" + fixture.authorization.AuthorizationID}, ObservedAt: "2026-09-08T00:00:03Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m11-outcome/backup-recovery"}
	outcomeInput := filepath.Join(fixture.dir, "m11-outcome-input.json")
	if err := writeFixtureJSON(outcomeInput, outcome); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-outcome", fixture.dir, outcomeInput, postEntry.ArtifactID); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("record recovered FAILED outcome: code=%d response=%+v", code, response)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json", "m11-outcome-input.json"} {
		if err := os.Remove(filepath.Join(fixture.dir, name)); err != nil {
			t.Fatalf("remove fixture-only backup input %s: %v", name, err)
		}
	}
	if code, response := backupCall(t, "create", fixture.dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup did not persist recovered FAILED execution: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("recovered FAILED execution backup did not restore: code=%d response=%+v", code, response)
	}
}

func assertM11UnknownStopRecovered(t *testing.T, fixture m11UnknownStopFixture, attemptedAt, reason string) {
	t.Helper()
	if _, err := os.Stat(m11UnknownStopJournalPath(fixture.dir)); !os.IsNotExist(err) {
		t.Fatalf("unknown STOP journal remains after recovery: %v", err)
	}
	state, err := loadMissionState(fixture.dir)
	if err != nil || !state.Stop || state.StopReason != "RECONCILIATION_REQUIRED" {
		t.Fatalf("recovered mission is not durably stopped: state=%+v err=%v", state, err)
	}
	_, head, err := m11LedgerHead(fixture.dir, fixture.lease.LeaseID)
	if err != nil || head.ControlMode != "STOPPED" || !head.ReconciliationRequired {
		t.Fatalf("recovered stopped ledger is invalid: ledger=%+v err=%v", head, err)
	}
	if _, _, status, err := recordUnknownM11Execution(fixture.dir, fixture.authorization.AuthorizationID, fixture.ledgerEntry.ArtifactID, attemptedAt, reason); err != nil || status != appendDuplicate {
		t.Fatalf("unknown retry was not exact duplicate: status=%s err=%v", status, err)
	}
}
