package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func m11ArtifactRegistryPath(dir string) string { return filepath.Join(dir, "m11-artifacts.jsonl") }

// m11RegistryAppendFault is a test-only seam for deterministic failure drills.
// Production leaves it nil; it is deliberately not controlled by CLI input or
// environment variables.
var m11RegistryAppendFault func(phase string, entry corem11.ArtifactEntry) error

func m11AppendFault(phase string, entry corem11.ArtifactEntry) error {
	if m11RegistryAppendFault == nil {
		return nil
	}
	return m11RegistryAppendFault(phase, entry)
}

func loadM11ArtifactRegistry(dir string) ([]corem11.ArtifactEntry, error) {
	raw, err := os.ReadFile(m11ArtifactRegistryPath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entries := []corem11.ArtifactEntry{}
	seen := map[string]string{}
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, err := corem11.ValidateArtifactEntry(line)
		if err != nil {
			return nil, fmt.Errorf("invalid M11 artifact registry entry: %w", err)
		}
		key := entry.ArtifactKind + "\x00" + entry.ArtifactID
		if prior, exists := seen[key]; exists {
			if prior == entry.ContentHash {
				return nil, fmt.Errorf("duplicate M11 artifact registry entry")
			}
			return nil, fmt.Errorf("M11 artifact ID reused with different content")
		}
		seen[key] = entry.ContentHash
		entries = append(entries, entry)
	}
	if err := corem11.ValidateArtifactGraph(entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func registerM11Artifact(dir, kind string, raw []byte) (corem11.ArtifactEntry, string, error) {
	entry, err := corem11.NewArtifactEntry(kind, raw)
	if err != nil {
		return entry, "", err
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return entry, "", err
	}
	for _, registered := range entries {
		if registered.ArtifactKind != entry.ArtifactKind || registered.ArtifactID != entry.ArtifactID {
			continue
		}
		if registered.ContentHash == entry.ContentHash {
			return entry, appendDuplicate, nil
		}
		return entry, "", fmt.Errorf("M11 artifact ID reused with different content")
	}
	if err := corem11.ValidateArtifactGraph(append(entries, entry)); err != nil {
		return entry, "", err
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return entry, "", err
	}
	f, err := os.OpenFile(m11ArtifactRegistryPath(dir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return entry, "", err
	}
	if err := m11AppendFault("before_write", entry); err != nil {
		_ = f.Close()
		return entry, "", err
	}
	_, err = f.Write(append(line, '\n'))
	faultErr := m11AppendFault("after_write", entry)
	if syncErr := f.Sync(); err == nil {
		err = syncErr
	}
	if faultErr == nil {
		faultErr = m11AppendFault("after_sync", entry)
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil && faultErr != nil {
		err = faultErr
	}
	if err != nil {
		return entry, "", err
	}
	return entry, appendAdded, nil
}

func resolveM11Artifact(dir, kind, id, hash string) (corem11.ArtifactEntry, error) {
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ArtifactEntry{}, err
	}
	for _, entry := range entries {
		if entry.ArtifactKind == kind && entry.ArtifactID == id && (hash == "" || hash == entry.ContentHash) {
			return entry, nil
		}
	}
	return corem11.ArtifactEntry{}, fmt.Errorf("M11 artifact does not resolve")
}

func activateM11Lease(dir, leaseID, activatedAt string) (corem11.ProductionActivationRecord, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionActivationRecord{}, "", err
	}
	if state.Stop {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ProductionActivationRecord{}, "", err
	}
	var lease *corem11.ProductionLease
	approvalExists := false
	for _, entry := range entries {
		profile := ""
		if entry.ArtifactKind == corem11.ArtifactKindLease {
			profile = "lease"
		}
		if entry.ArtifactKind == corem11.ArtifactKindLeaseApproval {
			profile = "approval"
		}
		if profile == "" {
			continue
		}
		value, status := corem11.DecodeArtifact(profile, entry.Artifact)
		if status != corem11.Valid {
			return corem11.ProductionActivationRecord{}, "", fmt.Errorf("invalid M11 registry artifact")
		}
		if candidate, ok := value.(*corem11.ProductionLease); ok && candidate.LeaseID == leaseID {
			copied := *candidate
			lease = &copied
		}
		if candidate, ok := value.(*corem11.ProductionLeaseApproval); ok && candidate.LeaseID == leaseID {
			approvalExists = true
		}
	}
	if lease == nil || !approvalExists {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("activation requires registered production lease and approval")
	}
	activated, errActivated := time.Parse(time.RFC3339, activatedAt)
	validFrom, errValid := time.Parse(time.RFC3339, lease.ValidFrom)
	expires, errExpires := time.Parse(time.RFC3339, lease.ExpiresAt)
	if errActivated != nil || errValid != nil || errExpires != nil || activated.Before(validFrom) || !activated.Before(expires) {
		return corem11.ProductionActivationRecord{}, "", fmt.Errorf("activation time is outside the registered lease")
	}
	record := corem11.ProductionActivationRecord{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ActivatedAt: activatedAt}
	raw, err := json.Marshal(record)
	if err != nil {
		return record, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindActivation, raw)
	return record, status, err
}

func initializeM11Ledger(dir, leaseID, initializedAt string) (corem11.ProductionLedger, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	if state.Stop {
		return corem11.ProductionLedger{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	var lease *corem11.ProductionLease
	var activation *corem11.ProductionActivationRecord
	for _, entry := range entries {
		profile := ""
		if entry.ArtifactKind == corem11.ArtifactKindLease {
			profile = "lease"
		}
		if entry.ArtifactKind == corem11.ArtifactKindActivation {
			profile = "activation"
		}
		if profile == "" {
			continue
		}
		value, status := corem11.DecodeArtifact(profile, entry.Artifact)
		if status != corem11.Valid {
			return corem11.ProductionLedger{}, "", fmt.Errorf("invalid M11 registry artifact")
		}
		if candidate, ok := value.(*corem11.ProductionLease); ok && candidate.LeaseID == leaseID {
			copied := *candidate
			lease = &copied
		}
		if candidate, ok := value.(*corem11.ProductionActivationRecord); ok && candidate.LeaseID == leaseID {
			copied := *candidate
			activation = &copied
		}
	}
	if lease == nil || activation == nil || activation.LeaseVersion != lease.LeaseVersion || activation.LeaseHash != lease.LeaseHash {
		return corem11.ProductionLedger{}, "", fmt.Errorf("ledger initialization requires registered activation")
	}
	now, errNow := time.Parse(time.RFC3339, initializedAt)
	activated, errActivated := time.Parse(time.RFC3339, activation.ActivatedAt)
	expires, errExpires := time.Parse(time.RFC3339, lease.ExpiresAt)
	if errNow != nil || errActivated != nil || errExpires != nil || now.Before(activated) || !now.Before(expires) {
		return corem11.ProductionLedger{}, "", fmt.Errorf("ledger initialization time is outside active lease")
	}
	ledger := corem11.ProductionLedger{LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, ControlMode: "NORMAL", WindowStartedAt: initializedAt, PendingExecutionIDs: []string{}, SuccessfulIdempotencyKeys: []string{}, OutcomeLinks: []corem11.ProductionOutcomeLink{}, ReconciliationResolutionIDs: []string{}, UpdatedAt: initializedAt}
	raw, err := json.Marshal(ledger)
	if err != nil {
		return ledger, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, raw)
	return ledger, status, err
}

func m11ArtifactValue(dir, kind, id string) (any, error) {
	entry, err := resolveM11Artifact(dir, kind, id, "")
	if err != nil {
		return nil, err
	}
	profiles := map[string]string{corem11.ArtifactKindLease: "lease", corem11.ArtifactKindLeaseApproval: "approval", corem11.ArtifactKindHealth: "health", corem11.ArtifactKindCostBound: "cost", corem11.ArtifactKindLedger: "ledger", corem11.ArtifactKindGate: "gate", corem11.ArtifactKindAuthorization: "authorization", corem11.ArtifactKindExecution: "execution", corem11.ArtifactKindActivation: "activation", corem11.ArtifactKindReconciliation: "resolution", corem11.ArtifactKindEvaluation: "evaluation", corem11.ArtifactKindCycle: "cycle"}
	profile := profiles[kind]
	if profile == "" {
		return nil, fmt.Errorf("unsupported M11 gate artifact")
	}
	value, status := corem11.DecodeArtifact(profile, entry.Artifact)
	if status != corem11.Valid {
		return nil, fmt.Errorf("invalid registered M11 artifact")
	}
	return value, nil
}

func m11Allowed(values []string, value string) bool {
	for _, candidate := range values {
		if strings.EqualFold(candidate, value) {
			return true
		}
	}
	return false
}

func evaluateM11Gate(dir, leaseID, healthID, costID, ledgerID, evaluatedAt string) (corem11.ProductionGateDecision, string, error) {
	state, err := loadMissionState(dir)
	if err != nil || state.Intent == nil || state.Policy == nil {
		return corem11.ProductionGateDecision{}, "", fmt.Errorf("M11 gate requires persisted intent and policy")
	}
	leaseValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLease, leaseID)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	lease := leaseValue.(*corem11.ProductionLease)
	healthValue, err := m11ArtifactValue(dir, corem11.ArtifactKindHealth, healthID)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	health := healthValue.(*corem11.ProductionHealthSnapshot)
	costValue, err := m11ArtifactValue(dir, corem11.ArtifactKindCostBound, costID)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	cost := costValue.(corem10.TrustedCostBound)
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, ledgerID)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	activationValue, err := m11ArtifactValue(dir, corem11.ArtifactKindActivation, lease.LeaseID+"/"+lease.LeaseVersion)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	activation := activationValue.(*corem11.ProductionActivationRecord)
	if _, err := m11ArtifactValue(dir, corem11.ArtifactKindLeaseApproval, lease.ApprovalRef); err != nil {
		return corem11.ProductionGateDecision{}, "", err
	}
	now, err := time.Parse(time.RFC3339, evaluatedAt)
	if err != nil {
		return corem11.ProductionGateDecision{}, "", fmt.Errorf("invalid evaluated_at")
	}
	gate := corem11.ProductionGateDecision{GateID: "prod-gate-" + lease.LeaseID + "-" + state.Intent.IntentID, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: state.Policy.PolicyVersion, RiskClass: state.Policy.RiskClass, HealthSnapshotID: health.SnapshotID, HealthSnapshotHash: health.SnapshotHash, CostBoundID: cost.CostBoundID, CostBoundHash: cost.CostBoundHash, CostBoundMinor: cost.MaxCostMinor, Decision: "DENY", Reason: "INVALID_PRODUCTION_STATE", EvaluatedAt: evaluatedAt, ExecutionsTotalBefore: ledger.ExecutionsTotal, ExecutionsInWindowBefore: ledger.ExecutionsInWindow, CostMinorTotalBefore: ledger.CostMinorTotal, PendingOutcomesBefore: ledger.PendingOutcomes, ExecutionAuthorized: false}
	decision := func(kind, reason string) (corem11.ProductionGateDecision, string, error) {
		gate.Decision, gate.Reason = kind, reason
		raw, e := json.Marshal(gate)
		if e != nil {
			return gate, "", e
		}
		_, status, e := registerM11Artifact(dir, corem11.ArtifactKindGate, raw)
		return gate, status, e
	}
	if state.Stop {
		return decision("STOP", "DURABLE_STOP")
	}
	validFrom, e1 := time.Parse(time.RFC3339, lease.ValidFrom)
	expires, e2 := time.Parse(time.RFC3339, lease.ExpiresAt)
	if e1 != nil || e2 != nil || now.Before(validFrom) || !now.Before(expires) || activation.LeaseHash != lease.LeaseHash {
		return decision("DENY", "LEASE_INACTIVE")
	}
	if ledger.LeaseHash != lease.LeaseHash || ledger.ControlMode == "STOPPED" {
		return decision("STOP", "STICKY_STOP")
	}
	if ledger.ReconciliationRequired {
		return decision("STOP", "RECONCILIATION_REQUIRED")
	}
	if state.Policy.IntentID != state.Intent.IntentID || state.Policy.IntentHash != state.Intent.IntentHash || state.Policy.PolicyVersion != lease.PolicyVersion || !m11Allowed(lease.AllowedRiskClasses, state.Policy.RiskClass) || !m11Allowed(lease.AllowedActionTypes, state.Intent.ActionType) || !m11Allowed(lease.AllowedHosts, strings.Split(strings.TrimPrefix(state.Intent.Target, "https://"), "/")[0]) {
		return decision("DENY", "POLICY_OR_SCOPE_MISMATCH")
	}
	if state.Policy.RiskClass != "RISK0" || state.Policy.Decision != "ALLOW" {
		return decision("REQUIRE_APPROVAL", "RISK_NOT_PRODUCTION_ELIGIBLE")
	}
	if status := corem10.ValidFor(cost, state.Intent.IntentID, state.Intent.IntentHash, state.Intent.CorrelationID, lease.Currency, now); status != "VALID" {
		return decision("DENY", status)
	}
	if ledger.ExecutionsTotal >= lease.MaxExecutionsTotal || ledger.ExecutionsInWindow >= lease.MaxExecutionsPerWindow || ledger.PendingOutcomes >= lease.MaxPendingOutcomes || cost.MaxCostMinor > lease.MaxCostMinorTotal-ledger.CostMinorTotal {
		return decision("DENY", "BUDGET_EXCEEDED")
	}
	observed, observedErr := time.Parse(time.RFC3339, health.ObservedAt)
	if observedErr != nil || observed.After(now) || health.LeaseHash != lease.LeaseHash {
		return decision("DENY", "HEALTH_MISMATCH")
	}
	if now.Unix()-observed.Unix() > int64(lease.MaxHealthSnapshotAgeSeconds) {
		return decision("DEGRADE", "HEALTH_STALE")
	}
	if health.ComplianceAlertCount > 0 {
		return decision("STOP", "COMPLIANCE_ALERT")
	}
	if health.ReconciliationRequired || health.ConsecutiveFailures >= lease.MaxConsecutiveFailures || health.OldestPendingOutcomeAgeSeconds > lease.MaxOutcomeAgeSeconds {
		return decision("STOP", "HEALTH_SAFETY_BLOCK")
	}
	if !health.TelemetryComplete || health.DependencyState != "HEALTHY" {
		return decision("DEGRADE", "HEALTH_DEGRADED")
	}
	return decision("ALLOW_PRODUCTION", "PRODUCTION_ELIGIBLE")
}

func authorizeM11Production(dir, leaseID, gateID, executorID, authorizedAt string) (corem11.ProductionExecutionAuthorization, string, error) {
	state, err := loadMissionState(dir)
	if err != nil || state.Intent == nil || state.Policy == nil {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("M11 authorization requires persisted intent and policy")
	}
	if state.Stop {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	now, err := time.Parse(time.RFC3339, authorizedAt)
	if err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("invalid authorized_at")
	}
	if err := missionAuthorityActive(state, now); err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", err
	}
	leaseValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLease, leaseID)
	if err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", err
	}
	lease := leaseValue.(*corem11.ProductionLease)
	gateValue, err := m11ArtifactValue(dir, corem11.ArtifactKindGate, gateID)
	if err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", err
	}
	gate := gateValue.(*corem11.ProductionGateDecision)
	if gate.Decision != "ALLOW_PRODUCTION" || gate.LeaseID != lease.LeaseID || gate.LeaseVersion != lease.LeaseVersion || gate.LeaseHash != lease.LeaseHash || gate.IntentID != state.Intent.IntentID || gate.IntentHash != state.Intent.IntentHash || gate.PolicyVersion != state.Policy.PolicyVersion || !m11Allowed(lease.ExecutorIDs, executorID) {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("production gate or executor does not authorize this request")
	}
	healthValue, err := m11ArtifactValue(dir, corem11.ArtifactKindHealth, gate.HealthSnapshotID)
	if err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", err
	}
	health := healthValue.(*corem11.ProductionHealthSnapshot)
	costValue, err := m11ArtifactValue(dir, corem11.ArtifactKindCostBound, gate.CostBoundID)
	if err != nil {
		return corem11.ProductionExecutionAuthorization{}, "", err
	}
	cost := costValue.(corem10.TrustedCostBound)
	if health.SnapshotHash != gate.HealthSnapshotHash || cost.CostBoundHash != gate.CostBoundHash || cost.MaxCostMinor != gate.CostBoundMinor || corem10.ValidFor(cost, state.Intent.IntentID, state.Intent.IntentHash, state.Intent.CorrelationID, lease.Currency, now) != "VALID" {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("production gate dependencies no longer resolve")
	}
	limits := []string{lease.ExpiresAt, state.Intent.ExpiresAt, state.Approval.ExpiresAt, cost.ExpiresAt}
	expires := time.Time{}
	for _, raw := range limits {
		parsed, parseErr := time.Parse(time.RFC3339, raw)
		if parseErr != nil {
			return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("authorization expiry is invalid")
		}
		if expires.IsZero() || parsed.Before(expires) {
			expires = parsed
		}
	}
	if !expires.After(now) {
		return corem11.ProductionExecutionAuthorization{}, "", fmt.Errorf("authorization would already be expired")
	}
	authorization := corem11.ProductionExecutionAuthorization{AuthorizationID: "prod-auth-" + gate.GateID + "-" + executorID, IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: state.Policy.PolicyVersion, ProductionLeaseID: lease.LeaseID, ProductionLeaseVersion: lease.LeaseVersion, ProductionLeaseHash: lease.LeaseHash, ProductionGateID: gate.GateID, ProductionHealthSnapshotID: health.SnapshotID, ProductionHealthSnapshotHash: health.SnapshotHash, ProductionCostBoundID: cost.CostBoundID, ProductionCostBoundHash: cost.CostBoundHash, ProductionCostBoundMinor: cost.MaxCostMinor, ExecutorID: executorID, AuthorizedAt: authorizedAt, ExpiresAt: expires.Format(time.RFC3339), IdempotencyKey: state.Intent.IdempotencyKey, CorrelationID: state.Intent.CorrelationID, ExecutionMode: "GOVERNED_PRODUCTION", ExecutionAuthorized: true}
	raw, err := json.Marshal(authorization)
	if err != nil {
		return authorization, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindAuthorization, raw)
	return authorization, status, err
}

func reserveM11Authorization(dir, authorizationID, ledgerID, reservedAt string) (corem11.ProductionLedger, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	if state.Stop {
		return corem11.ProductionLedger{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	now, err := time.Parse(time.RFC3339, reservedAt)
	if err != nil {
		return corem11.ProductionLedger{}, "", fmt.Errorf("invalid reserved_at")
	}
	authValue, err := m11ArtifactValue(dir, corem11.ArtifactKindAuthorization, authorizationID)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	auth := authValue.(*corem11.ProductionExecutionAuthorization)
	expires, err := time.Parse(time.RFC3339, auth.ExpiresAt)
	if err != nil || !expires.After(now) || !auth.ExecutionAuthorized || auth.ExecutionMode != "GOVERNED_PRODUCTION" {
		return corem11.ProductionLedger{}, "", fmt.Errorf("production authorization is inactive")
	}
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, ledgerID)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	if ledger.LeaseID != auth.ProductionLeaseID || ledger.LeaseVersion != auth.ProductionLeaseVersion || ledger.LeaseHash != auth.ProductionLeaseHash || ledger.ControlMode != "NORMAL" || ledger.ReconciliationRequired || ledger.ExecutionsTotal < 0 || ledger.CostMinorTotal < 0 {
		return corem11.ProductionLedger{}, "", fmt.Errorf("production ledger is not reservable")
	}
	leaseValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLease, auth.ProductionLeaseID)
	if err != nil {
		return corem11.ProductionLedger{}, "", err
	}
	lease := leaseValue.(*corem11.ProductionLease)
	executionID := "prod-exec-" + auth.AuthorizationID
	for _, id := range ledger.PendingExecutionIDs {
		if id == executionID {
			return corem11.ProductionLedger{}, "", fmt.Errorf("production authorization already reserved")
		}
	}
	if ledger.ExecutionsTotal >= lease.MaxExecutionsTotal || ledger.ExecutionsInWindow >= lease.MaxExecutionsPerWindow || ledger.PendingOutcomes >= lease.MaxPendingOutcomes || auth.ProductionCostBoundMinor > lease.MaxCostMinorTotal-ledger.CostMinorTotal {
		return corem11.ProductionLedger{}, "", fmt.Errorf("production budget exhausted")
	}
	next := *ledger
	next.ExecutionsTotal++
	next.ExecutionsInWindow++
	next.CostMinorTotal += auth.ProductionCostBoundMinor
	next.PendingOutcomes++
	next.PendingExecutionIDs = append(append([]string(nil), ledger.PendingExecutionIDs...), executionID)
	next.UpdatedAt = reservedAt
	raw, err := json.Marshal(next)
	if err != nil {
		return next, "", err
	}
	// The caller supplies an immutable predecessor ledger. Checking that single
	// artifact is insufficient: a second request could reuse it with a different
	// timestamp and reserve the same authorization twice. Search the canonical
	// registry before append, while still recognizing an exact retry.
	expected, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, raw)
	if err != nil {
		return next, "", err
	}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return next, "", err
	}
	for _, entry := range entries {
		if entry.ArtifactKind != corem11.ArtifactKindLedger {
			continue
		}
		value, decodeStatus := corem11.DecodeArtifact("ledger", entry.Artifact)
		if decodeStatus != corem11.Valid {
			return next, "", fmt.Errorf("invalid registered production ledger")
		}
		for _, pendingID := range value.(*corem11.ProductionLedger).PendingExecutionIDs {
			if pendingID != executionID {
				continue
			}
			if entry.ArtifactID == expected.ArtifactID && entry.ContentHash == expected.ContentHash {
				return next, appendDuplicate, nil
			}
			return next, "", fmt.Errorf("production authorization already reserved")
		}
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, raw)
	return next, status, err
}

func recordFailedM11Execution(dir, authorizationID, reservationLedgerID, attemptedAt, reason string) (corem11.ProductionExecutionRecord, corem11.ProductionLedger, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	if state.Stop {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	now, err := time.Parse(time.RFC3339, attemptedAt)
	if err != nil || strings.TrimSpace(reason) == "" {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("invalid fixture execution input")
	}
	authValue, err := m11ArtifactValue(dir, corem11.ArtifactKindAuthorization, authorizationID)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	auth := authValue.(*corem11.ProductionExecutionAuthorization)
	expires, err := time.Parse(time.RFC3339, auth.ExpiresAt)
	if err != nil || !expires.After(now) {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("production authorization is expired")
	}
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, reservationLedgerID)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	executionID := "prod-exec-" + auth.AuthorizationID
	pending := false
	for _, id := range ledger.PendingExecutionIDs {
		pending = pending || id == executionID
	}
	if !pending || ledger.PendingOutcomes < 1 || ledger.LeaseID != auth.ProductionLeaseID {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("execution has no governed reservation")
	}
	record := corem11.ProductionExecutionRecord{ExecutionID: executionID, AuthorizationID: auth.AuthorizationID, ProductionLeaseID: auth.ProductionLeaseID, ProductionLeaseVersion: auth.ProductionLeaseVersion, ProductionLeaseHash: auth.ProductionLeaseHash, ProductionGateID: auth.ProductionGateID, ProductionHealthSnapshotID: auth.ProductionHealthSnapshotID, ProductionHealthSnapshotHash: auth.ProductionHealthSnapshotHash, ProductionCostBoundID: auth.ProductionCostBoundID, ProductionCostBoundHash: auth.ProductionCostBoundHash, ProductionCostBoundMinor: auth.ProductionCostBoundMinor, IntentID: auth.IntentID, IntentHash: auth.IntentHash, ExecutorID: auth.ExecutorID, IdempotencyKey: auth.IdempotencyKey, AttemptedAt: attemptedAt, Status: "FAILED", SideEffectState: "NOT_PERFORMED", Error: reason, CorrelationID: auth.CorrelationID}
	raw, err := json.Marshal(record)
	if err != nil {
		return record, corem11.ProductionLedger{}, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindExecution, raw)
	if err != nil {
		return record, corem11.ProductionLedger{}, status, err
	}
	next := *ledger
	next.ConsecutiveFailures++
	next.LastExecutionAt = attemptedAt
	next.UpdatedAt = attemptedAt
	ledgerRaw, err := json.Marshal(next)
	if err != nil {
		return record, next, status, err
	}
	_, ledgerStatus, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		return record, next, ledgerStatus, err
	}
	return record, next, status, nil
}

// recordUnknownM11Execution is the only learner fixture that can model an
// indeterminate external effect. It fails closed: the mission STOP marker and
// immutable stopped ledger are written before returning the record. A later
// reconciliation can establish facts, but never reactivates this lease.
func recordUnknownM11Execution(dir, authorizationID, reservationLedgerID, attemptedAt, reason string) (corem11.ProductionExecutionRecord, corem11.ProductionLedger, string, error) {
	state, err := loadMissionState(dir)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	if state.Stop {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	now, err := time.Parse(time.RFC3339, attemptedAt)
	if err != nil || strings.TrimSpace(reason) == "" {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("invalid unknown execution input")
	}
	authValue, err := m11ArtifactValue(dir, corem11.ArtifactKindAuthorization, authorizationID)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	auth := authValue.(*corem11.ProductionExecutionAuthorization)
	expires, err := time.Parse(time.RFC3339, auth.ExpiresAt)
	if err != nil || !expires.After(now) || !auth.ExecutionAuthorized || auth.ExecutionMode != "GOVERNED_PRODUCTION" {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("production authorization is inactive")
	}
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, reservationLedgerID)
	if err != nil {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	executionID := "prod-exec-" + auth.AuthorizationID
	pending := false
	for _, id := range ledger.PendingExecutionIDs {
		pending = pending || id == executionID
	}
	if !pending || ledger.PendingOutcomes < 1 || ledger.LeaseID != auth.ProductionLeaseID || ledger.ControlMode != "NORMAL" || ledger.ReconciliationRequired {
		return corem11.ProductionExecutionRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("execution has no governed reservation")
	}
	record := corem11.ProductionExecutionRecord{ExecutionID: executionID, AuthorizationID: auth.AuthorizationID, ProductionLeaseID: auth.ProductionLeaseID, ProductionLeaseVersion: auth.ProductionLeaseVersion, ProductionLeaseHash: auth.ProductionLeaseHash, ProductionGateID: auth.ProductionGateID, ProductionHealthSnapshotID: auth.ProductionHealthSnapshotID, ProductionHealthSnapshotHash: auth.ProductionHealthSnapshotHash, ProductionCostBoundID: auth.ProductionCostBoundID, ProductionCostBoundHash: auth.ProductionCostBoundHash, ProductionCostBoundMinor: auth.ProductionCostBoundMinor, IntentID: auth.IntentID, IntentHash: auth.IntentHash, ExecutorID: auth.ExecutorID, IdempotencyKey: auth.IdempotencyKey, AttemptedAt: attemptedAt, Status: "RECONCILIATION_REQUIRED", SideEffectState: "UNKNOWN", Error: reason, CorrelationID: auth.CorrelationID}
	raw, err := json.Marshal(record)
	if err != nil {
		return record, corem11.ProductionLedger{}, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindExecution, raw)
	if err != nil {
		return record, corem11.ProductionLedger{}, status, err
	}
	next := *ledger
	next.ControlMode = "STOPPED"
	next.StopReason = "RECONCILIATION_REQUIRED"
	next.ReconciliationRequired = true
	next.LastExecutionAt = attemptedAt
	next.UpdatedAt = attemptedAt
	ledgerRaw, err := json.Marshal(next)
	if err != nil {
		return record, next, status, err
	}
	if _, ledgerStatus, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, ledgerRaw); err != nil {
		return record, next, ledgerStatus, err
	}
	state.Stop, state.StopReason = true, "RECONCILIATION_REQUIRED"
	if err := saveMissionState(dir, state); err != nil {
		return record, next, status, err
	}
	if err := writeJSONAtomic(filepath.Join(dir, "STOP"), map[string]any{"active": true, "reason": state.StopReason}); err != nil {
		return record, next, status, err
	}
	return record, next, status, nil
}

// reconcileM11Execution requires an already registered, human-authored
// resolution. It preserves durable STOP: the old lease can never resume.
func reconcileM11Execution(dir, resolutionID, ledgerID string) (corem11.ProductionReconciliationResolution, corem11.ProductionLedger, string, error) {
	state, err := loadMissionState(dir)
	if err != nil || !state.Stop {
		return corem11.ProductionReconciliationResolution{}, corem11.ProductionLedger{}, "", fmt.Errorf("reconciliation requires durable STOP")
	}
	value, err := m11ArtifactValue(dir, corem11.ArtifactKindReconciliation, resolutionID)
	if err != nil {
		return corem11.ProductionReconciliationResolution{}, corem11.ProductionLedger{}, "", err
	}
	resolution := value.(*corem11.ProductionReconciliationResolution)
	if resolution.ResolvedBy != "human" || strings.TrimSpace(resolution.ResolverID) == "" || strings.TrimSpace(resolution.Reason) == "" || resolution.EffectState != "NOT_PERFORMED" {
		return *resolution, corem11.ProductionLedger{}, "", fmt.Errorf("reconciliation requires a human NOT_PERFORMED resolution")
	}
	resolvedAt, err := time.Parse(time.RFC3339, resolution.ResolvedAt)
	if err != nil {
		return *resolution, corem11.ProductionLedger{}, "", fmt.Errorf("invalid reconciliation time")
	}
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, ledgerID)
	if err != nil {
		return *resolution, corem11.ProductionLedger{}, "", err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	executionValue, err := m11ArtifactValue(dir, corem11.ArtifactKindExecution, resolution.ExecutionID)
	if err != nil {
		return *resolution, corem11.ProductionLedger{}, "", err
	}
	execution := executionValue.(*corem11.ProductionExecutionRecord)
	if ledger.ControlMode != "STOPPED" || !ledger.ReconciliationRequired || execution.Status != "RECONCILIATION_REQUIRED" || execution.SideEffectState != "UNKNOWN" || ledger.LeaseID != resolution.LeaseID || ledger.LeaseVersion != resolution.LeaseVersion || ledger.LeaseHash != resolution.LeaseHash || execution.ProductionLeaseID != resolution.LeaseID || resolvedAt.Before(mustM11Time(execution.AttemptedAt)) {
		return *resolution, corem11.ProductionLedger{}, "", fmt.Errorf("reconciliation does not bind the stopped unknown execution")
	}
	for _, id := range ledger.ReconciliationResolutionIDs {
		if id == resolution.ResolutionID {
			return *resolution, *ledger, appendDuplicate, nil
		}
	}
	next := *ledger
	next.ReconciliationRequired = false
	next.StopReason = "RECOVERY_REVIEW_REQUIRED"
	next.ReconciliationResolutionIDs = append(append([]string(nil), ledger.ReconciliationResolutionIDs...), resolution.ResolutionID)
	next.UpdatedAt = resolution.ResolvedAt
	raw, err := json.Marshal(next)
	if err != nil {
		return *resolution, next, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, raw)
	return *resolution, next, status, err
}

func mustM11Time(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// m11RecoveryHandoff is deliberately read-only. It proves that a stopped lease
// was reconciled by a human, but it carries no authority to resume it. A new
// runtime must still bind a separately reviewed lease through normal commands.
func m11RecoveryHandoff(dir, resolutionID, ledgerID string) (map[string]any, error) {
	state, err := loadMissionState(dir)
	if err != nil || !state.Stop {
		return nil, fmt.Errorf("recovery handoff requires durable STOP")
	}
	resolutionValue, err := m11ArtifactValue(dir, corem11.ArtifactKindReconciliation, resolutionID)
	if err != nil {
		return nil, err
	}
	resolution := resolutionValue.(*corem11.ProductionReconciliationResolution)
	ledgerValue, err := m11ArtifactValue(dir, corem11.ArtifactKindLedger, ledgerID)
	if err != nil {
		return nil, err
	}
	ledger := ledgerValue.(*corem11.ProductionLedger)
	if ledger.ControlMode != "STOPPED" || ledger.StopReason != "RECOVERY_REVIEW_REQUIRED" || ledger.ReconciliationRequired || ledger.LeaseID != resolution.LeaseID || ledger.LeaseVersion != resolution.LeaseVersion || ledger.LeaseHash != resolution.LeaseHash {
		return nil, fmt.Errorf("recovery handoff does not bind a reviewed stopped ledger")
	}
	linked := false
	for _, id := range ledger.ReconciliationResolutionIDs {
		linked = linked || id == resolution.ResolutionID
	}
	if !linked {
		return nil, fmt.Errorf("recovery resolution is not linked from stopped ledger")
	}
	return map[string]any{"profile": "M11_RECOVERY_HANDOFF/v1", "prior_lease_id": ledger.LeaseID, "prior_lease_version": ledger.LeaseVersion, "prior_lease_hash": ledger.LeaseHash, "resolution_id": resolution.ResolutionID, "execution_id": resolution.ExecutionID, "resolved_by": resolution.ResolvedBy, "resolver_id": resolution.ResolverID, "resolved_at": resolution.ResolvedAt, "effect_state": resolution.EffectState, "prior_stop_reason": ledger.StopReason, "requires_new_runtime": true, "requires_new_lease": true, "execution_permitted": false}, nil
}

// evaluateM11FixtureOutcome is deliberately narrow: it only evaluates the
// learner's CANCELLED/NOT_PERFORMED fixture outcome. It records an audit link,
// never a business result, performance claim, authority grant, or lease change.
func evaluateM11FixtureOutcome(dir, outcomeID, evaluationID, evaluatedAt string) (corem11.ProductionOutcomeEvaluation, string, error) {
	if strings.TrimSpace(outcomeID) == "" || strings.TrimSpace(evaluationID) == "" {
		return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("outcome_id and evaluation_id are required")
	}
	evaluated, err := time.Parse(time.RFC3339, evaluatedAt)
	if err != nil {
		return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("invalid evaluated_at")
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return corem11.ProductionOutcomeEvaluation{}, "", err
	}
	var outcome *m03.OutcomeRecord
	for index := range outcomes {
		if outcomes[index].OutcomeID == outcomeID {
			copied := outcomes[index]
			outcome = &copied
		}
	}
	if outcome == nil {
		return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("M11 fixture outcome does not resolve")
	}
	observed, err := time.Parse(time.RFC3339, outcome.ObservedAt)
	if err != nil || evaluated.Before(observed) {
		return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("evaluation precedes fixture outcome")
	}
	executionValue, err := m11ArtifactValue(dir, corem11.ArtifactKindExecution, outcome.EffectRef.EffectID)
	if err != nil {
		return corem11.ProductionOutcomeEvaluation{}, "", err
	}
	execution := executionValue.(*corem11.ProductionExecutionRecord)
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return corem11.ProductionOutcomeEvaluation{}, "", err
	}
	for _, entry := range entries {
		if entry.ArtifactKind != corem11.ArtifactKindEvaluation {
			continue
		}
		value, status := corem11.DecodeArtifact("evaluation", entry.Artifact)
		if status != corem11.Valid {
			return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("invalid M11 outcome evaluation")
		}
		prior := value.(*corem11.ProductionOutcomeEvaluation)
		if prior.EvaluationID == evaluationID {
			candidate := corem11.ProductionOutcomeEvaluation{EvaluationID: evaluationID, LeaseID: execution.ProductionLeaseID, LeaseVersion: execution.ProductionLeaseVersion, LeaseHash: execution.ProductionLeaseHash, ExecutionID: execution.ExecutionID, OutcomeID: outcome.OutcomeID, EvaluatedAt: evaluatedAt, Result: "FIXTURE_NO_SIDE_EFFECT", EvidenceIDs: []string{outcome.OutcomeID}, Limitations: []string{"offline fixture: proves only a recorded NOT_PERFORMED outcome; it is not a business outcome"}, SourceProfile: "OFFLINE_FIXTURE"}
			if reflect.DeepEqual(*prior, candidate) {
				return candidate, appendDuplicate, nil
			}
			return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("evaluation_id reused with different content")
		}
		if prior.OutcomeID == outcome.OutcomeID || prior.ExecutionID == execution.ExecutionID {
			return corem11.ProductionOutcomeEvaluation{}, "", fmt.Errorf("M11 fixture outcome already evaluated")
		}
	}
	evaluation := corem11.ProductionOutcomeEvaluation{EvaluationID: evaluationID, LeaseID: execution.ProductionLeaseID, LeaseVersion: execution.ProductionLeaseVersion, LeaseHash: execution.ProductionLeaseHash, ExecutionID: execution.ExecutionID, OutcomeID: outcome.OutcomeID, EvaluatedAt: evaluatedAt, Result: "FIXTURE_NO_SIDE_EFFECT", EvidenceIDs: []string{outcome.OutcomeID}, Limitations: []string{"offline fixture: proves only a recorded NOT_PERFORMED outcome; it is not a business outcome"}, SourceProfile: "OFFLINE_FIXTURE"}
	raw, err := json.Marshal(evaluation)
	if err != nil {
		return evaluation, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindEvaluation, raw)
	return evaluation, status, err
}

// closeM11FixtureCycle ties the fixture evaluation back to the original
// canonical history and M11 authority lineage. It is audit-only and cannot
// reactivate a lease, clear STOP, or authorize another execution.
func closeM11FixtureCycle(dir, cycleID, evaluationID, closedAt string) (corem11.ProductionCycleRecord, string, error) {
	state, err := loadMissionState(dir)
	if err != nil || state.Intent == nil || state.Policy == nil {
		return corem11.ProductionCycleRecord{}, "", fmt.Errorf("cycle requires persisted intent and policy")
	}
	closed, err := time.Parse(time.RFC3339, closedAt)
	if err != nil || strings.TrimSpace(cycleID) == "" {
		return corem11.ProductionCycleRecord{}, "", fmt.Errorf("invalid cycle close input")
	}
	evaluationValue, err := m11ArtifactValue(dir, corem11.ArtifactKindEvaluation, evaluationID)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	evaluation := evaluationValue.(*corem11.ProductionOutcomeEvaluation)
	evaluated, err := time.Parse(time.RFC3339, evaluation.EvaluatedAt)
	if err != nil || closed.Before(evaluated) {
		return corem11.ProductionCycleRecord{}, "", fmt.Errorf("cycle closes before evaluation")
	}
	executionValue, err := m11ArtifactValue(dir, corem11.ArtifactKindExecution, evaluation.ExecutionID)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	execution := executionValue.(*corem11.ProductionExecutionRecord)
	authorizedValue, err := m11ArtifactValue(dir, corem11.ArtifactKindAuthorization, execution.AuthorizationID)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	authorization := authorizedValue.(*corem11.ProductionExecutionAuthorization)
	gateValue, err := m11ArtifactValue(dir, corem11.ArtifactKindGate, execution.ProductionGateID)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	gate := gateValue.(*corem11.ProductionGateDecision)
	if evaluation.LeaseID != execution.ProductionLeaseID || evaluation.LeaseVersion != execution.ProductionLeaseVersion || evaluation.LeaseHash != execution.ProductionLeaseHash || authorization.IntentID != state.Intent.IntentID || authorization.IntentHash != state.Intent.IntentHash || gate.IntentID != state.Intent.IntentID || gate.IntentHash != state.Intent.IntentHash {
		return corem11.ProductionCycleRecord{}, "", fmt.Errorf("cycle has mismatched lineage")
	}
	if _, err := m11ArtifactValue(dir, corem11.ArtifactKindLease, evaluation.LeaseID); err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	matchedOutcome := false
	for _, outcome := range outcomes {
		matchedOutcome = matchedOutcome || outcome.OutcomeID == evaluation.OutcomeID && outcome.EffectRef.EffectID == execution.ExecutionID
	}
	if !matchedOutcome {
		return corem11.ProductionCycleRecord{}, "", fmt.Errorf("cycle evaluation outcome does not resolve")
	}
	record, err := resolveCanonicalRecord(filepath.Join(dir, "history.jsonl"), state.Intent.DecisionID)
	if err != nil {
		return corem11.ProductionCycleRecord{}, "", err
	}
	observationIDs := make([]string, 0, len(record.Observations))
	for _, observation := range record.Observations {
		observationIDs = append(observationIDs, observation.ObservationID)
	}
	cycle := corem11.ProductionCycleRecord{CycleID: cycleID, LeaseID: evaluation.LeaseID, LeaseVersion: evaluation.LeaseVersion, LeaseHash: evaluation.LeaseHash, ObservationIDs: observationIDs, DecisionID: record.RecordID, IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, GateID: gate.GateID, AuthorizationID: authorization.AuthorizationID, ExecutionID: execution.ExecutionID, OutcomeID: evaluation.OutcomeID, EvaluationID: evaluation.EvaluationID, Status: "CLOSED", OpenedAt: execution.AttemptedAt, ClosedAt: closedAt, CorrelationID: execution.CorrelationID}
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return cycle, "", err
	}
	for _, entry := range entries {
		if entry.ArtifactKind != corem11.ArtifactKindCycle {
			continue
		}
		value, status := corem11.DecodeArtifact("cycle", entry.Artifact)
		if status != corem11.Valid {
			return cycle, "", fmt.Errorf("invalid production cycle")
		}
		prior := value.(*corem11.ProductionCycleRecord)
		if prior.CycleID == cycleID {
			if reflect.DeepEqual(*prior, cycle) {
				return cycle, appendDuplicate, nil
			}
			return cycle, "", fmt.Errorf("cycle_id reused with different content")
		}
		if prior.EvaluationID == cycle.EvaluationID || prior.ExecutionID == cycle.ExecutionID {
			return cycle, "", fmt.Errorf("M11 execution already has a closed cycle")
		}
	}
	raw, err := json.Marshal(cycle)
	if err != nil {
		return cycle, "", err
	}
	_, status, err := registerM11Artifact(dir, corem11.ArtifactKindCycle, raw)
	return cycle, status, err
}
