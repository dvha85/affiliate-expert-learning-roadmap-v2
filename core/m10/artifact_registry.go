package m10

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

const (
	ArtifactKindCanaryGrant            = "CANARY_GRANT"
	ArtifactKindTrustedCostBound       = "TRUSTED_COST_BOUND"
	ArtifactKindCanaryGate             = "CANARY_GATE"
	ArtifactKindExecutionAuthorization = "EXECUTION_AUTHORIZATION"
	ArtifactKindExecutionRecord        = "EXECUTION_RECORD"
)

// ArtifactEntry is a compact, immutable registry envelope. ContentHash is a
// digest of the canonical serialized artifact, distinct from any domain hash
// carried inside that artifact.
type ArtifactEntry struct {
	ArtifactKind string          `json:"artifact_kind"`
	ArtifactID   string          `json:"artifact_id"`
	ContentHash  string          `json:"content_hash"`
	Artifact     json.RawMessage `json:"artifact"`
}

func artifactContentHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func canonicalArtifact(kind string, raw []byte) (string, []byte, error) {
	switch kind {
	case ArtifactKindCanaryGrant:
		grant, status := DecodeCanaryGrant(raw)
		if status != "VALID" {
			return "", nil, fmt.Errorf("canary grant: %s", status)
		}
		canonical, err := json.Marshal(grant)
		return grant.GrantID, canonical, err
	case ArtifactKindTrustedCostBound:
		bound, status := DecodeTrustedCostBound(raw)
		if status != "VALID" {
			return "", nil, fmt.Errorf("trusted cost bound: %s", status)
		}
		canonical, err := json.Marshal(bound)
		return bound.CostBoundID, canonical, err
	case ArtifactKindCanaryGate:
		gate, err := ValidateCanaryGateDecision(raw)
		if err != nil {
			return "", nil, err
		}
		canonical, err := json.Marshal(gate)
		return gate.GateID, canonical, err
	case ArtifactKindExecutionAuthorization:
		authorization, err := ValidateExecutionAuthorization(raw)
		if err != nil {
			return "", nil, err
		}
		canonical, err := json.Marshal(authorization)
		return authorization.AuthorizationID, canonical, err
	case ArtifactKindExecutionRecord:
		record, err := ValidateExecutionRecord(raw)
		if err != nil {
			return "", nil, err
		}
		canonical, err := json.Marshal(record)
		return record.ExecutionID, canonical, err
	default:
		return "", nil, fmt.Errorf("unsupported M10 artifact kind %q", kind)
	}
}

func NewArtifactEntry(kind string, raw []byte) (ArtifactEntry, error) {
	id, canonical, err := canonicalArtifact(kind, raw)
	if err != nil {
		return ArtifactEntry{}, err
	}
	return ArtifactEntry{ArtifactKind: kind, ArtifactID: id, ContentHash: artifactContentHash(canonical), Artifact: canonical}, nil
}

func ValidateArtifactEntry(raw []byte) (ArtifactEntry, error) {
	var entry ArtifactEntry
	if err := contracts.DecodeStrict(raw, &entry); err != nil {
		return entry, err
	}
	expected, err := NewArtifactEntry(entry.ArtifactKind, entry.Artifact)
	if err != nil {
		return ArtifactEntry{}, err
	}
	if entry.ArtifactID != expected.ArtifactID || entry.ContentHash != expected.ContentHash || !bytes.Equal(entry.Artifact, expected.Artifact) {
		return ArtifactEntry{}, fmt.Errorf("M10 artifact registry entry integrity mismatch")
	}
	return entry, nil
}

// ValidateArtifactGraph makes each governed-canary artifact resolve to its
// exact immutable parents. It is non-authorizing: mutable ledger/reservation
// ownership remains with the runtime adapter.
func ValidateArtifactGraph(entries []ArtifactEntry) error {
	grants := map[string]CanaryGrant{}
	bounds := map[string]TrustedCostBound{}
	gates := map[string]CanaryGateDecision{}
	authorizations := map[string]ExecutionAuthorization{}
	records := []ExecutionRecord{}
	executionAuthorizations := map[string]string{}
	seenEntries := map[string]bool{}
	for _, entry := range entries {
		expected, err := NewArtifactEntry(entry.ArtifactKind, entry.Artifact)
		if err != nil {
			return fmt.Errorf("invalid registered M10 artifact: %w", err)
		}
		if entry.ArtifactID != expected.ArtifactID || entry.ContentHash != expected.ContentHash || !bytes.Equal(entry.Artifact, expected.Artifact) {
			return fmt.Errorf("M10 artifact registry entry integrity mismatch")
		}
		entryKey := entry.ArtifactKind + "\x00" + entry.ArtifactID
		if seenEntries[entryKey] {
			return fmt.Errorf("duplicate M10 artifact in registry graph")
		}
		seenEntries[entryKey] = true
		switch entry.ArtifactKind {
		case ArtifactKindCanaryGrant:
			grant, status := DecodeCanaryGrant(entry.Artifact)
			if status != "VALID" {
				return fmt.Errorf("invalid registered canary grant")
			}
			grants[grant.GrantID] = grant
		case ArtifactKindTrustedCostBound:
			bound, status := DecodeTrustedCostBound(entry.Artifact)
			if status != "VALID" {
				return fmt.Errorf("invalid registered trusted cost bound")
			}
			bounds[bound.CostBoundID] = bound
		case ArtifactKindCanaryGate:
			gate, err := ValidateCanaryGateDecision(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered canary gate: %w", err)
			}
			gates[gate.GateID] = gate
		case ArtifactKindExecutionAuthorization:
			authorization, err := ValidateExecutionAuthorization(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered execution authorization: %w", err)
			}
			authorizations[authorization.AuthorizationID] = authorization
		case ArtifactKindExecutionRecord:
			record, err := ValidateExecutionRecord(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered execution record: %w", err)
			}
			records = append(records, record)
		default:
			return fmt.Errorf("unsupported registered M10 artifact kind")
		}
	}
	for _, gate := range gates {
		grant, grantOK := grants[gate.GrantID]
		bound, boundOK := bounds[gate.CostBoundID]
		if !grantOK || !boundOK || grant.GrantVersion != gate.GrantVersion || grant.GrantHash != gate.GrantHash || bound.CostBoundHash != gate.CostBoundHash || bound.MaxCostMinor != gate.CostBoundMinor || bound.IntentID != gate.IntentID || bound.IntentHash != gate.IntentHash || grant.PolicyVersion != gate.PolicyVersion {
			return fmt.Errorf("canary gate has an orphaned or mismatched registry link")
		}
		expectedGateID := canaryGateID(CanaryGateInput{
			Grant: grant, CostBound: bound, IntentID: gate.IntentID, IntentHash: gate.IntentHash, PolicyVersion: gate.PolicyVersion,
			Now: gate.EvaluatedAt,
			Ledger: CanaryLedgerSnapshot{
				ExecutionsTotal: gate.ExecutionsTotalBefore, ExecutionsInWindow: gate.ExecutionsInWindowBefore,
				CostMinorTotal: gate.CostMinorTotalBefore, PendingOutcomes: gate.PendingOutcomesBefore,
			},
		})
		if gate.GateID != expectedGateID {
			return fmt.Errorf("canary gate has a non-canonical gate ID")
		}
	}
	for _, authorization := range authorizations {
		grant, grantOK := grants[authorization.CanaryGrantID]
		gate, gateOK := gates[authorization.CanaryGateID]
		bound, boundOK := bounds[authorization.CanaryCostBoundID]
		if !grantOK || !gateOK || !boundOK || gate.Decision != "ALLOW_CANARY" || gate.PerActionApprovalRequired || grant.GrantVersion != authorization.CanaryGrantVersion || grant.GrantHash != authorization.CanaryGrantHash || gate.GrantID != authorization.CanaryGrantID || gate.GrantVersion != authorization.CanaryGrantVersion || gate.GrantHash != authorization.CanaryGrantHash || gate.IntentID != authorization.IntentID || gate.IntentHash != authorization.IntentHash || gate.PolicyVersion != authorization.PolicyVersion || gate.CostBoundID != authorization.CanaryCostBoundID || gate.CostBoundHash != authorization.CanaryCostBoundHash || gate.CostBoundMinor != authorization.CanaryCostBoundMinor || bound.CostBoundHash != authorization.CanaryCostBoundHash || bound.MaxCostMinor != authorization.CanaryCostBoundMinor {
			return fmt.Errorf("execution authorization has an orphaned or mismatched registry link")
		}
		if err := validateCanaryAuthorizationParents(authorization, grant, gate, bound); err != nil {
			return fmt.Errorf("execution authorization parent graph is invalid: %w", err)
		}
		expectedAuthorizationID := authorizationID(CanaryAuthorizationInput{
			Gate: gate, IntentID: authorization.IntentID, IntentHash: authorization.IntentHash, ExecutorID: authorization.ExecutorID,
			AuthorizedAt: authorization.AuthorizedAt, IdempotencyKey: authorization.IdempotencyKey,
		})
		if authorization.AuthorizationID != expectedAuthorizationID {
			return fmt.Errorf("execution authorization has a non-canonical authorization ID")
		}
	}
	for _, record := range records {
		authorization, exists := authorizations[record.AuthorizationID]
		attemptedAt, attemptedErr := time.Parse(time.RFC3339, record.AttemptedAt)
		authorizedAt, authorizedErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
		expiresAt, expiresErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
		if !exists || attemptedErr != nil || authorizedErr != nil || expiresErr != nil || attemptedAt.Before(authorizedAt) || !attemptedAt.Before(expiresAt) || authorization.IntentID != record.IntentID || authorization.IntentHash != record.IntentHash || authorization.ExecutorID != record.ExecutorID || authorization.IdempotencyKey != record.IdempotencyKey || authorization.CorrelationID != record.CorrelationID || authorization.CanaryGrantID != record.CanaryGrantID || authorization.CanaryGrantVersion != record.CanaryGrantVersion || authorization.CanaryGrantHash != record.CanaryGrantHash || authorization.CanaryGateID != record.CanaryGateID || authorization.CanaryCostBoundID != record.CanaryCostBoundID || authorization.CanaryCostBoundHash != record.CanaryCostBoundHash || authorization.CanaryCostBoundMinor != record.CanaryCostBoundMinor {
			return fmt.Errorf("execution record has an orphaned or mismatched registry link")
		}
		if record.ExecutionID != terminalExecutionID(authorization, record.AttemptedAt, record.Status, record.Error) {
			return fmt.Errorf("execution record has a non-canonical execution ID")
		}
		if priorExecutionID, found := executionAuthorizations[record.AuthorizationID]; found && priorExecutionID != record.ExecutionID {
			return fmt.Errorf("execution authorization has more than one terminal record")
		}
		executionAuthorizations[record.AuthorizationID] = record.ExecutionID
	}
	return nil
}

// validateCanaryAuthorizationParents rechecks the issuance invariants that
// remain available after the original intent/policy context has been reduced
// to registry artifacts. A checksum-valid envelope is not enough: the graph
// must still describe an actually eligible gate, an in-scope executor and a
// coherent historical time window. This deliberately does not compare against
// the current wall clock, so backup/restore and historical audits remain
// deterministic.
func validateCanaryAuthorizationParents(authorization ExecutionAuthorization, grant CanaryGrant, gate CanaryGateDecision, bound TrustedCostBound) error {
	if gate.Decision != "ALLOW_CANARY" || gate.Reason != "CANARY_ELIGIBLE" || gate.ExecutionAuthorized || gate.PerActionApprovalRequired || gate.RiskClass != "RISK0" {
		return fmt.Errorf("gate does not prove canary eligibility")
	}
	if grant.IntentionalPlaceholder() || !containsGrantValue(grant.AllowedRiskClasses, gate.RiskClass) || !containsGrantValue(grant.ExecutorIDs, authorization.ExecutorID) {
		return fmt.Errorf("grant does not delegate the authorized risk or executor")
	}
	if grant.CorrelationID != authorization.CorrelationID || bound.CorrelationID != authorization.CorrelationID || bound.IntentID != authorization.IntentID || bound.IntentHash != authorization.IntentHash {
		return fmt.Errorf("authorization correlation or cost binding is invalid")
	}
	if authorization.ExecutionMode != "GOVERNED_CANARY" || !authorization.ExecutionAuthorized {
		return fmt.Errorf("authorization is not a governed canary capability")
	}
	gateAt, gateErr := time.Parse(time.RFC3339, gate.EvaluatedAt)
	authorizedAt, authorizedErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
	expiresAt, expiresErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
	validFrom, validErr := time.Parse(time.RFC3339, grant.ValidFrom)
	grantExpires, grantExpiryErr := time.Parse(time.RFC3339, grant.ExpiresAt)
	costExpires, costExpiryErr := time.Parse(time.RFC3339, bound.ExpiresAt)
	if gateErr != nil || authorizedErr != nil || expiresErr != nil || validErr != nil || grantExpiryErr != nil || costExpiryErr != nil {
		return fmt.Errorf("authorization chronology is invalid")
	}
	if gateAt.Before(validFrom) || !gateAt.Before(grantExpires) || authorizedAt.Before(gateAt) || !expiresAt.After(authorizedAt) || expiresAt.After(grantExpires) || expiresAt.After(costExpires) {
		return fmt.Errorf("authorization chronology is invalid")
	}
	if status := ValidFor(bound, authorization.IntentID, authorization.IntentHash, authorization.CorrelationID, grant.Currency, gateAt); status != "VALID" {
		return fmt.Errorf("cost bound is not valid at gate evaluation: %s", status)
	}
	return nil
}
