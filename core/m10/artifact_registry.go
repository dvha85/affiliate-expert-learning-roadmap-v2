package m10

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

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
