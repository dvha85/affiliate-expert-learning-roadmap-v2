package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// Serialize empty runtime collections canonically without mutating the ledger.
// This does not relax the raw input schema or wire the persistence reader.
func (l ProductionLedger) MarshalJSON() ([]byte, error) {
	type wire ProductionLedger
	v := wire(l)
	if v.PendingExecutionIDs == nil {
		v.PendingExecutionIDs = []string{}
	}
	if v.SuccessfulIdempotencyKeys == nil {
		v.SuccessfulIdempotencyKeys = []string{}
	}
	if v.OutcomeLinks == nil {
		v.OutcomeLinks = []ProductionOutcomeLink{}
	}
	if v.ReconciliationResolutionIDs == nil {
		v.ReconciliationResolutionIDs = []string{}
	}
	return json.Marshal(v)
}

// DecodeM11Artifact checks one untrusted file, never a production permission.
func DecodeM11Artifact(kind string, raw []byte) (any, string) {
	if kind == "cost" {
		return DecodeM10Artifact("cost", raw)
	}
	var schema string
	var v any
	switch kind {
	case "lease":
		schema, v = "production-lease.schema.json", &ProductionLease{}
	case "approval":
		schema, v = "production-lease-approval.schema.json", &ProductionLeaseApproval{}
	case "health":
		schema, v = "production-health-snapshot.schema.json", &ProductionHealthSnapshot{}
	case "ledger":
		schema, v = "production-ledger.schema.json", &ProductionLedger{}
	case "gate":
		schema, v = "production-gate-decision.schema.json", &ProductionGateDecision{}
	case "authorization":
		schema, v = "execution-authorization.schema.json", &ProductionExecutionAuthorization{}
	case "execution":
		schema, v = "execution-record.schema.json", &ProductionExecutionRecord{}
	case "activation":
		schema, v = "production-activation-record.schema.json", &productionActivationRecord{}
	case "resolution":
		schema, v = "production-reconciliation-resolution.schema.json", &ProductionReconciliationResolution{}
	case "cycle":
		schema, v = "production-cycle-record.schema.json", &ProductionCycleRecord{}
	default:
		return nil, "INVALID_PROFILE"
	}
	if contracts.ValidateRaw(schema, raw) != nil || contracts.DecodeStrict(raw, v) != nil {
		return nil, "INVALID_SCHEMA"
	}
	before := func(a, b string) bool {
		x, e := time.Parse(time.RFC3339, a)
		y, f := time.Parse(time.RFC3339, b)
		return e == nil && f == nil && x.Before(y)
	}
	switch x := v.(type) {
	case *ProductionLease:
		if x.LeaseHash != ComputeProductionLeaseHash(*x) {
			return nil, "TAMPERED_LEASE"
		}
		if before(x.ValidFrom, x.ReviewedAt) || !before(x.ValidFrom, x.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
		if x.MaxExecutionsPerWindow > x.MaxExecutionsTotal || hasWildcard(x.AllowedActionTypes) || hasWildcard(x.ExecutorIDs) {
			return nil, "INVALID_LEASE"
		}
	case *ProductionHealthSnapshot:
		if x.SnapshotHash != ComputeProductionHealthHash(*x) {
			return nil, "TAMPERED_HEALTH"
		}
	case *ProductionExecutionAuthorization:
		if x.ExecutionMode != "GOVERNED_PRODUCTION" {
			return nil, "INVALID_PROFILE"
		}
		if !before(x.AuthorizedAt, x.ExpiresAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *ProductionCycleRecord:
		if before(x.ClosedAt, x.OpenedAt) {
			return nil, "INVALID_TIME_BINDING"
		}
	case *ProductionLedger:
		if x.ExecutionsInWindow > x.ExecutionsTotal || x.PendingOutcomes != len(x.PendingExecutionIDs) || x.PendingOutcomes > x.ExecutionsTotal || len(x.SuccessfulIdempotencyKeys) > x.ExecutionsTotal {
			return nil, "INVALID_LEDGER"
		}
		if before(x.UpdatedAt, x.WindowStartedAt) || x.LastExecutionAt != "" && before(x.UpdatedAt, x.LastExecutionAt) || x.LastOutcomeAt != "" && before(x.UpdatedAt, x.LastOutcomeAt) {
			return nil, "INVALID_TIME_BINDING"
		}
		outcomes, executions := map[string]bool{}, map[string]bool{}
		for _, link := range x.OutcomeLinks {
			if outcomes[link.OutcomeID] || executions[link.ExecutionID] {
				return nil, "INVALID_LEDGER"
			}
			outcomes[link.OutcomeID], executions[link.ExecutionID] = true, true
			if before(x.UpdatedAt, link.ObservedAt) {
				return nil, "INVALID_TIME_BINDING"
			}
		}
		for _, id := range x.PendingExecutionIDs {
			if executions[id] {
				return nil, "INVALID_LEDGER"
			}
		}
	}
	return v, missionValid
}

func runM11Check(w io.Writer, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: demo m11-check lease|approval|health|cost|ledger|gate|authorization|execution|activation|resolution|cycle FILE.json")
	}
	b, e := os.ReadFile(args[1])
	if e != nil {
		return e
	}
	if _, s := DecodeM11Artifact(args[0], b); s != missionValid {
		return fmt.Errorf("M11 artifact audit: %s", s)
	}
	// Same non-authorizing summary contract as the single-file M10 audit.
	return json.NewEncoder(w).Encode(M10CheckSummary{Result: "ARTIFACT_VALID_UNVERIFIED", ArtifactType: args[0]})
}
