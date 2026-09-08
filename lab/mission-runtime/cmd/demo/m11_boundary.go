package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
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
	var v any
	switch kind {
	case "lease":
		v = &ProductionLease{}
	case "approval":
		v = &ProductionLeaseApproval{}
	case "health":
		v = &ProductionHealthSnapshot{}
	case "ledger":
		v = &ProductionLedger{}
	case "gate":
		v = &ProductionGateDecision{}
	case "authorization":
		v = &ProductionExecutionAuthorization{}
	case "execution":
		v = &ProductionExecutionRecord{}
	case "activation":
		v = &productionActivationRecord{}
	case "resolution":
		v = &ProductionReconciliationResolution{}
	case "cycle":
		v = &ProductionCycleRecord{}
	default:
		return nil, "INVALID_PROFILE"
	}
	if _, status := corem11.DecodeArtifact(kind, raw); status != corem11.Valid {
		return nil, status
	}
	if json.Unmarshal(raw, v) != nil {
		return nil, "INVALID_SCHEMA"
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
