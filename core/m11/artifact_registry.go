package m11

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

const (
	ArtifactKindLease             = "PRODUCTION_LEASE"
	ArtifactKindLeaseApproval     = "PRODUCTION_LEASE_APPROVAL"
	ArtifactKindHealth            = "PRODUCTION_HEALTH_SNAPSHOT"
	ArtifactKindCostBound         = "TRUSTED_COST_BOUND"
	ArtifactKindLedger            = "PRODUCTION_LEDGER"
	ArtifactKindGate              = "PRODUCTION_GATE"
	ArtifactKindAuthorization     = "PRODUCTION_EXECUTION_AUTHORIZATION"
	ArtifactKindExecution         = "PRODUCTION_EXECUTION_RECORD"
	ArtifactKindActivation        = "PRODUCTION_ACTIVATION"
	ArtifactKindReconciliation    = "PRODUCTION_RECONCILIATION"
	ArtifactKindRecoveryAdmission = "PRODUCTION_RECOVERY_ADMISSION"
	ArtifactKindEvaluation        = "PRODUCTION_OUTCOME_EVALUATION"
	ArtifactKindCycle             = "PRODUCTION_CYCLE"
)

// ArtifactEntry is an append-only, canonical M11 lifecycle artifact envelope.
// Its integrity digest is distinct from hashes carried by individual artifacts.
type ArtifactEntry struct {
	ArtifactKind string          `json:"artifact_kind"`
	ArtifactID   string          `json:"artifact_id"`
	ContentHash  string          `json:"content_hash"`
	Artifact     json.RawMessage `json:"artifact"`
}

func kindProfile(kind string) string {
	switch kind {
	case ArtifactKindLease:
		return "lease"
	case ArtifactKindLeaseApproval:
		return "approval"
	case ArtifactKindHealth:
		return "health"
	case ArtifactKindCostBound:
		return "cost"
	case ArtifactKindLedger:
		return "ledger"
	case ArtifactKindGate:
		return "gate"
	case ArtifactKindAuthorization:
		return "authorization"
	case ArtifactKindExecution:
		return "execution"
	case ArtifactKindActivation:
		return "activation"
	case ArtifactKindReconciliation:
		return "resolution"
	case ArtifactKindRecoveryAdmission:
		return "recovery_admission"
	case ArtifactKindEvaluation:
		return "evaluation"
	case ArtifactKindCycle:
		return "cycle"
	default:
		return ""
	}
}

func contentHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func NewArtifactEntry(kind string, raw []byte) (ArtifactEntry, error) {
	profile := kindProfile(kind)
	if profile == "" {
		return ArtifactEntry{}, fmt.Errorf("unsupported M11 artifact kind %q", kind)
	}
	id, canonical, status := CanonicalArtifact(profile, raw)
	if status != Valid {
		return ArtifactEntry{}, fmt.Errorf("M11 %s: %s", profile, status)
	}
	return ArtifactEntry{ArtifactKind: kind, ArtifactID: id, ContentHash: contentHash(canonical), Artifact: canonical}, nil
}

func ValidateArtifactEntry(raw []byte) (ArtifactEntry, error) {
	var entry ArtifactEntry
	if err := contracts.DecodeStrict(raw, &entry); err != nil {
		return entry, err
	}
	expected, err := NewArtifactEntry(entry.ArtifactKind, entry.Artifact)
	if err != nil {
		return entry, err
	}
	if entry.ArtifactID != expected.ArtifactID || entry.ContentHash != expected.ContentHash || !bytes.Equal(entry.Artifact, expected.Artifact) {
		return entry, fmt.Errorf("M11 artifact registry entry integrity mismatch")
	}
	return entry, nil
}

// ValidateArtifactGraph makes every registered M11 lifecycle link resolve to
// exact, immutable parent artifacts. It deliberately does not authorize a run.
func ValidateArtifactGraph(entries []ArtifactEntry) error {
	leases := map[string]ProductionLease{}
	health := map[string]ProductionHealthSnapshot{}
	costs := map[string]corem10.TrustedCostBound{}
	gates := map[string]ProductionGateDecision{}
	authorizations := map[string]ProductionExecutionAuthorization{}
	executions := map[string]ProductionExecutionRecord{}
	evaluations := map[string]ProductionOutcomeEvaluation{}
	ledgerHeads := map[string]ProductionLedger{}
	ledgerEntries := map[string]ArtifactEntry{}
	ledgers := []ProductionLedger{}
	activations := map[string]ProductionActivationRecord{}
	for _, entry := range entries {
		profile := kindProfile(entry.ArtifactKind)
		value, status := DecodeArtifact(profile, entry.Artifact)
		if status != Valid {
			return fmt.Errorf("invalid registered M11 artifact")
		}
		switch x := value.(type) {
		case *ProductionLease:
			leases[x.LeaseID] = *x
		case *ProductionLeaseApproval:
			lease, ok := leases[x.LeaseID]
			if !ok || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || lease.ApprovalRef != x.ApprovalID || lease.PromotionReviewRef != x.PromotionReviewRef || lease.SourceCanaryGrantID != x.SourceCanaryGrantID || lease.SourceCanaryGrantVersion != x.SourceCanaryGrantVersion || lease.SourceCanaryGrantHash != x.SourceCanaryGrantHash || lease.ReviewerID != x.ReviewerID || lease.ReviewedAt != x.ReviewedAt {
				return fmt.Errorf("production lease approval has an orphaned or mismatched link")
			}
		case *ProductionHealthSnapshot:
			lease, ok := leases[x.LeaseID]
			observedAt, observedErr := time.Parse(time.RFC3339, x.ObservedAt)
			validFrom, validFromErr := time.Parse(time.RFC3339, lease.ValidFrom)
			expiresAt, expiresErr := time.Parse(time.RFC3339, lease.ExpiresAt)
			if !ok || observedErr != nil || validFromErr != nil || expiresErr != nil || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || observedAt.Before(validFrom) || !observedAt.Before(expiresAt) {
				return fmt.Errorf("production health has an orphaned lease link")
			}
			health[x.SnapshotID] = *x
		case corem10.TrustedCostBound:
			costs[x.CostBoundID] = x
		case *ProductionLedger:
			lease, ok := leases[x.LeaseID]
			if !ok || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash {
				return fmt.Errorf("production ledger has an orphaned lease link")
			}
			if previous, exists := ledgerHeads[x.LeaseID]; exists {
				previousAt, previousErr := time.Parse(time.RFC3339, previous.UpdatedAt)
				currentAt, currentErr := time.Parse(time.RFC3339, x.UpdatedAt)
				if previousErr != nil || currentErr != nil || !currentAt.After(previousAt) || x.ExecutionsTotal < previous.ExecutionsTotal || x.CostMinorTotal < previous.CostMinorTotal {
					return fmt.Errorf("production ledger is not a monotonic lease history")
				}
			}
			ledgerHeads[x.LeaseID] = *x
			ledgerEntries[entry.ArtifactID] = entry
			ledgers = append(ledgers, *x)
		case *ProductionActivationRecord:
			lease, ok := leases[x.LeaseID]
			activatedAt, activatedErr := time.Parse(time.RFC3339, x.ActivatedAt)
			validFrom, validFromErr := time.Parse(time.RFC3339, lease.ValidFrom)
			expiresAt, expiresErr := time.Parse(time.RFC3339, lease.ExpiresAt)
			if !ok || activatedErr != nil || validFromErr != nil || expiresErr != nil || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || activatedAt.Before(validFrom) || !activatedAt.Before(expiresAt) {
				return fmt.Errorf("production activation has an orphaned lease link")
			}
			activations[x.LeaseID] = *x
		case *ProductionGateDecision:
			lease, leaseOK := leases[x.LeaseID]
			snapshot, healthOK := health[x.HealthSnapshotID]
			bound, costOK := costs[x.CostBoundID]
			ledger, ledgerOK := ledgerEntries[x.LedgerArtifactID]
			evaluatedAt, evaluatedErr := time.Parse(time.RFC3339, x.EvaluatedAt)
			validFrom, validFromErr := time.Parse(time.RFC3339, lease.ValidFrom)
			expiresAt, expiresErr := time.Parse(time.RFC3339, lease.ExpiresAt)
			if !leaseOK || !healthOK || !costOK || !ledgerOK || evaluatedErr != nil || validFromErr != nil || expiresErr != nil || ledger.ArtifactKind != ArtifactKindLedger || ledger.ContentHash != x.LedgerContentHash || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || lease.PolicyVersion != x.PolicyVersion || snapshot.SnapshotHash != x.HealthSnapshotHash || bound.CostBoundHash != x.CostBoundHash || bound.MaxCostMinor != x.CostBoundMinor || bound.IntentID != x.IntentID || bound.IntentHash != x.IntentHash || evaluatedAt.Before(validFrom) || !evaluatedAt.Before(expiresAt) {
				return fmt.Errorf("production gate has an orphaned or mismatched link")
			}
			gates[x.GateID] = *x
		case *ProductionExecutionAuthorization:
			lease, leaseOK := leases[x.ProductionLeaseID]
			gate, gateOK := gates[x.ProductionGateID]
			snapshot, healthOK := health[x.ProductionHealthSnapshotID]
			bound, costOK := costs[x.ProductionCostBoundID]
			authorizedAt, authorizedErr := time.Parse(time.RFC3339, x.AuthorizedAt)
			authorizationExpiresAt, authorizationExpiryErr := time.Parse(time.RFC3339, x.ExpiresAt)
			validFrom, validFromErr := time.Parse(time.RFC3339, lease.ValidFrom)
			leaseExpiresAt, leaseExpiryErr := time.Parse(time.RFC3339, lease.ExpiresAt)
			if !leaseOK || !gateOK || !healthOK || !costOK || authorizedErr != nil || authorizationExpiryErr != nil || validFromErr != nil || leaseExpiryErr != nil || lease.LeaseVersion != x.ProductionLeaseVersion || lease.LeaseHash != x.ProductionLeaseHash || gate.IntentID != x.IntentID || gate.IntentHash != x.IntentHash || gate.PolicyVersion != x.PolicyVersion || snapshot.SnapshotHash != x.ProductionHealthSnapshotHash || bound.CostBoundHash != x.ProductionCostBoundHash || bound.MaxCostMinor != x.ProductionCostBoundMinor || authorizedAt.Before(validFrom) || !authorizedAt.Before(leaseExpiresAt) || authorizationExpiresAt.After(leaseExpiresAt) || x.ExecutionMode != "GOVERNED_PRODUCTION" || !x.ExecutionAuthorized {
				return fmt.Errorf("production authorization has an orphaned or mismatched link")
			}
			authorizations[x.AuthorizationID] = *x
		case *ProductionExecutionRecord:
			auth, ok := authorizations[x.AuthorizationID]
			if !ok || auth.IntentID != x.IntentID || auth.IntentHash != x.IntentHash || auth.ExecutorID != x.ExecutorID || auth.IdempotencyKey != x.IdempotencyKey || auth.CorrelationID != x.CorrelationID || auth.ProductionLeaseID != x.ProductionLeaseID || auth.ProductionLeaseVersion != x.ProductionLeaseVersion || auth.ProductionLeaseHash != x.ProductionLeaseHash || auth.ProductionGateID != x.ProductionGateID || auth.ProductionHealthSnapshotID != x.ProductionHealthSnapshotID || auth.ProductionHealthSnapshotHash != x.ProductionHealthSnapshotHash || auth.ProductionCostBoundID != x.ProductionCostBoundID || auth.ProductionCostBoundHash != x.ProductionCostBoundHash || auth.ProductionCostBoundMinor != x.ProductionCostBoundMinor {
				return fmt.Errorf("production execution has an orphaned or mismatched link")
			}
			executions[x.ExecutionID] = *x
		case *ProductionReconciliationResolution:
			lease, leaseOK := leases[x.LeaseID]
			execution, executionOK := executions[x.ExecutionID]
			if !leaseOK || !executionOK || execution.ProductionLeaseID != x.LeaseID || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash {
				return fmt.Errorf("production reconciliation has an orphaned or mismatched link")
			}
		case *ProductionRecoveryAdmission:
			lease, leaseOK := leases[x.NewLeaseID]
			if !leaseOK || lease.LeaseVersion != x.NewLeaseVersion || lease.LeaseHash != x.NewLeaseHash || lease.ApprovalRef != x.NewApprovalID || x.ExecutionPermitted {
				return fmt.Errorf("production recovery admission has an orphaned or mismatched link")
			}
		case *ProductionOutcomeEvaluation:
			lease, leaseOK := leases[x.LeaseID]
			execution, executionOK := executions[x.ExecutionID]
			if !leaseOK || !executionOK || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || execution.ProductionLeaseID != x.LeaseID || execution.ProductionLeaseVersion != x.LeaseVersion || execution.ProductionLeaseHash != x.LeaseHash {
				return fmt.Errorf("production outcome evaluation has an orphaned or mismatched link")
			}
			evaluations[x.EvaluationID] = *x
		case *ProductionCycleRecord:
			lease, leaseOK := leases[x.LeaseID]
			gate, gateOK := gates[x.GateID]
			auth, authOK := authorizations[x.AuthorizationID]
			execution, executionOK := executions[x.ExecutionID]
			evaluation, evaluationOK := evaluations[x.EvaluationID]
			if !leaseOK || !gateOK || !authOK || !executionOK || !evaluationOK || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || gate.IntentID != x.IntentID || gate.IntentHash != x.IntentHash || auth.AuthorizationID != x.AuthorizationID || auth.ProductionGateID != x.GateID || execution.AuthorizationID != x.AuthorizationID || execution.ProductionGateID != x.GateID || execution.ExecutionID != x.ExecutionID || execution.AttemptedAt != x.OpenedAt || execution.CorrelationID != x.CorrelationID || evaluation.LeaseID != x.LeaseID || evaluation.ExecutionID != x.ExecutionID || evaluation.OutcomeID != x.OutcomeID {
				return fmt.Errorf("production cycle has an orphaned or mismatched link")
			}
		}
	}
	// Registry entries are append-only, but historical snapshots are validated
	// as a graph. Check the activation relation after decoding all entries so a
	// valid older inventory order cannot hide a ledger that began before its
	// lease was admitted.
	for _, ledger := range ledgers {
		activation, ok := activations[ledger.LeaseID]
		// An append-only registry may contain a historical lease draft before its
		// activation record is appended. Dedicated learner commands never create
		// that interim shape, and complete backup snapshots separately require an
		// activation; do not make loading the append transition itself impossible.
		if !ok {
			continue
		}
		windowStartedAt, windowErr := time.Parse(time.RFC3339, ledger.WindowStartedAt)
		updatedAt, updatedErr := time.Parse(time.RFC3339, ledger.UpdatedAt)
		activatedAt, activationErr := time.Parse(time.RFC3339, activation.ActivatedAt)
		if activation.LeaseVersion != ledger.LeaseVersion || activation.LeaseHash != ledger.LeaseHash || windowErr != nil || updatedErr != nil || activationErr != nil || windowStartedAt.Before(activatedAt) || updatedAt.Before(activatedAt) {
			return fmt.Errorf("production ledger predates its activation")
		}
	}
	for _, snapshot := range health {
		activation, ok := activations[snapshot.LeaseID]
		// Draft registry history can predate an eventual activation entry. Once
		// activation exists, however, health is evidence for the active runtime
		// and cannot have been observed before that admission boundary.
		if !ok {
			continue
		}
		observedAt, observedErr := time.Parse(time.RFC3339, snapshot.ObservedAt)
		activatedAt, activatedErr := time.Parse(time.RFC3339, activation.ActivatedAt)
		if observedErr != nil || activatedErr != nil || activation.LeaseVersion != snapshot.LeaseVersion || activation.LeaseHash != snapshot.LeaseHash || observedAt.Before(activatedAt) {
			return fmt.Errorf("production health predates its activation")
		}
	}
	return nil
}
