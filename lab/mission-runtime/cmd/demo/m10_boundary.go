package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

type CanaryGrantApproval struct {
	ApprovalRef  string `json:"approval_ref"`
	GrantID      string `json:"grant_id"`
	GrantVersion string `json:"grant_version"`
	GrantHash    string `json:"grant_hash"`
	Decision     string `json:"decision"`
	ApprovedBy   string `json:"approved_by"`
	ApproverID   string `json:"approver_id"`
	ApprovedAt   string `json:"approved_at"`
}

// MarshalJSON preserves the canonical array shape for an empty runtime ledger.
// Raw input still goes through schema validation and must not contain null arrays.
func (l CanaryLedger) MarshalJSON() ([]byte, error) {
	type wire CanaryLedger
	v := wire(l)
	if v.PendingExecutionIDs == nil {
		v.PendingExecutionIDs = []string{}
	}
	if v.SuccessfulIdempotencyKeys == nil {
		v.SuccessfulIdempotencyKeys = []string{}
	}
	if v.OutcomeLinks == nil {
		v.OutcomeLinks = []CanaryOutcomeLink{}
	}
	return json.Marshal(v)
}

// DecodeM10Artifact validates one file, not a trusted runtime state or a chain.
// No defaults, seals, authority aliases, or ledger normalization are applied.
func DecodeM10Artifact(kind string, raw []byte) (any, string) {
	var schema string
	var value any
	switch kind {
	case "grant":
		grant, status := corem10.DecodeCanaryGrant(raw)
		if status != "VALID" {
			return nil, status
		}
		value = (*CanaryGrant)(&grant)
	case "approval":
		schema, value = "canary-grant-approval.schema.json", &CanaryGrantApproval{}
	case "cost":
		bound, status := corem10.DecodeTrustedCostBound(raw)
		if status != "VALID" {
			return nil, status
		}
		value = (*CanaryCostBound)(&bound)
	case "ledger":
		schema, value = "canary-ledger.schema.json", &CanaryLedger{}
	case "gate":
		gate, err := corem10.ValidateCanaryGateDecision(raw)
		if err != nil {
			return nil, "INVALID_SCHEMA"
		}
		value = (*CanaryGateDecision)(&gate)
	case "authorization":
		authorization, err := corem10.ValidateExecutionAuthorization(raw)
		if err != nil {
			return nil, "INVALID_SCHEMA"
		}
		value = (*CanaryExecutionAuthorization)(&authorization)
	case "execution":
		record, err := corem10.DecodeCanaryExecutionRecord(raw)
		if err != nil {
			return nil, "INVALID_SCHEMA"
		}
		value = (*CanaryExecutionRecord)(&record)
	default:
		return nil, "INVALID_PROFILE"
	}
	if schema != "" && (contracts.ValidateRaw(schema, raw) != nil || contracts.DecodeStrict(raw, value) != nil) {
		return nil, "INVALID_SCHEMA"
	}
	before := func(a, b string) bool {
		x, e1 := time.Parse(time.RFC3339, a)
		y, e2 := time.Parse(time.RFC3339, b)
		return e1 == nil && e2 == nil && x.Before(y)
	}
	switch v := value.(type) {
	case *CanaryGrant:
		if v.GrantHash != ComputeCanaryGrantHash(*v) {
			return nil, "TAMPERED_GRANT"
		}
		if before(v.ValidFrom, v.ApprovedAt) || !before(v.ValidFrom, v.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *CanaryCostBound:
		if v.CostBoundHash != ComputeCanaryCostBoundHash(*v) {
			return nil, "TAMPERED_COST_BOUND"
		}
		if !before(v.ObservedAt, v.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *CanaryExecutionAuthorization:
		if v.ExecutionMode != "GOVERNED_CANARY" {
			return nil, "INVALID_PROFILE"
		}
		if !before(v.AuthorizedAt, v.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *CanaryLedger:
		if v.ExecutionsInWindow > v.ExecutionsTotal || v.PendingOutcomes != len(v.PendingExecutionIDs) || v.PendingOutcomes > v.ExecutionsTotal || len(v.SuccessfulIdempotencyKeys) > v.ExecutionsTotal {
			return nil, "INVALID_LEDGER"
		}
		if before(v.UpdatedAt, v.WindowStartedAt) || (v.LastExecutionAt != "" && before(v.UpdatedAt, v.LastExecutionAt)) {
			return nil, "INVALID_TIME_BINDING"
		}
		outcomes, executions := map[string]bool{}, map[string]bool{}
		for _, link := range v.OutcomeLinks {
			if outcomes[link.OutcomeID] || executions[link.ExecutionID] {
				return nil, "INVALID_LEDGER"
			}
			outcomes[link.OutcomeID], executions[link.ExecutionID] = true, true
			if before(v.UpdatedAt, link.ObservedAt) {
				return nil, "INVALID_TIME_BINDING"
			}
		}
		for _, id := range v.PendingExecutionIDs {
			if executions[id] {
				return nil, "INVALID_LEDGER"
			}
		}
	}
	return value, missionValid
}

type M10CheckSummary struct {
	Result                  string `json:"result"`
	ArtifactType            string `json:"artifact_type"`
	ProvenanceAuthenticated bool   `json:"provenance_authenticated"`
	ChainValidated          bool   `json:"chain_validated"`
	ExecutionPermitted      bool   `json:"execution_permitted"`
}

func runM10Check(w io.Writer, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: demo m10-check grant|approval|cost|ledger|gate|authorization|execution FILE.json")
	}
	raw, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	if _, state := DecodeM10Artifact(args[0], raw); state != missionValid {
		return fmt.Errorf("M10 artifact audit: %s", state)
	}
	// Never emit an authorization or call the canary gate/executor from file input.
	return json.NewEncoder(w).Encode(M10CheckSummary{Result: "ARTIFACT_VALID_UNVERIFIED", ArtifactType: args[0]})
}
