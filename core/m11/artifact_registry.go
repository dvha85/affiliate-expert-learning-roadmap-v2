package m11

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

const (
	ArtifactKindLease          = "PRODUCTION_LEASE"
	ArtifactKindLeaseApproval  = "PRODUCTION_LEASE_APPROVAL"
	ArtifactKindHealth         = "PRODUCTION_HEALTH_SNAPSHOT"
	ArtifactKindCostBound      = "TRUSTED_COST_BOUND"
	ArtifactKindLedger         = "PRODUCTION_LEDGER"
	ArtifactKindGate           = "PRODUCTION_GATE"
	ArtifactKindAuthorization  = "PRODUCTION_EXECUTION_AUTHORIZATION"
	ArtifactKindExecution      = "PRODUCTION_EXECUTION_RECORD"
	ArtifactKindActivation     = "PRODUCTION_ACTIVATION"
	ArtifactKindReconciliation = "PRODUCTION_RECONCILIATION"
	ArtifactKindCycle          = "PRODUCTION_CYCLE"
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
			if !ok || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash {
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
		case *ProductionActivationRecord:
			lease, ok := leases[x.LeaseID]
			if !ok || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash {
				return fmt.Errorf("production activation has an orphaned lease link")
			}
		case *ProductionGateDecision:
			lease, leaseOK := leases[x.LeaseID]
			snapshot, healthOK := health[x.HealthSnapshotID]
			bound, costOK := costs[x.CostBoundID]
			if !leaseOK || !healthOK || !costOK || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || lease.PolicyVersion != x.PolicyVersion || snapshot.SnapshotHash != x.HealthSnapshotHash || bound.CostBoundHash != x.CostBoundHash || bound.MaxCostMinor != x.CostBoundMinor || bound.IntentID != x.IntentID || bound.IntentHash != x.IntentHash {
				return fmt.Errorf("production gate has an orphaned or mismatched link")
			}
			gates[x.GateID] = *x
		case *ProductionExecutionAuthorization:
			lease, leaseOK := leases[x.ProductionLeaseID]
			gate, gateOK := gates[x.ProductionGateID]
			snapshot, healthOK := health[x.ProductionHealthSnapshotID]
			bound, costOK := costs[x.ProductionCostBoundID]
			if !leaseOK || !gateOK || !healthOK || !costOK || lease.LeaseVersion != x.ProductionLeaseVersion || lease.LeaseHash != x.ProductionLeaseHash || gate.IntentID != x.IntentID || gate.IntentHash != x.IntentHash || gate.PolicyVersion != x.PolicyVersion || snapshot.SnapshotHash != x.ProductionHealthSnapshotHash || bound.CostBoundHash != x.ProductionCostBoundHash || bound.MaxCostMinor != x.ProductionCostBoundMinor || x.ExecutionMode != "GOVERNED_PRODUCTION" || !x.ExecutionAuthorized {
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
		case *ProductionCycleRecord:
			lease, leaseOK := leases[x.LeaseID]
			gate, gateOK := gates[x.GateID]
			auth, authOK := authorizations[x.AuthorizationID]
			execution, executionOK := executions[x.ExecutionID]
			if !leaseOK || !gateOK || !authOK || !executionOK || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || gate.IntentID != x.IntentID || gate.IntentHash != x.IntentHash || auth.AuthorizationID != x.AuthorizationID || execution.ExecutionID != x.ExecutionID || execution.CorrelationID != x.CorrelationID {
				return fmt.Errorf("production cycle has an orphaned or mismatched link")
			}
		}
	}
	return nil
}
