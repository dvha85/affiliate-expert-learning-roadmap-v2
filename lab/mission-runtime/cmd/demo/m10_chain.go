package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

type M10ChainSummary struct {
	Result                  string `json:"result"`
	IntentID                string `json:"intent_id"`
	GrantID                 string `json:"grant_id"`
	ExecutionID             string `json:"execution_id"`
	ProvenanceAuthenticated bool   `json:"provenance_authenticated"`
	ExecutionPermitted      bool   `json:"execution_permitted"`
}

// CheckM10Chain audits a historical ALLOW_CANARY issuance chain with a
// pre-gate ledger snapshot. Decoding stays in this harness, while the
// cross-artifact validator is the shared core implementation.
func CheckM10Chain(raw [][]byte) (M10ChainSummary, string) {
	fail := func(status string) (M10ChainSummary, string) { return M10ChainSummary{}, status }
	if len(raw) != 9 {
		return fail("INVALID_INPUT")
	}
	i, status := DecodeM08Intent(raw[0])
	if status != missionValid {
		return fail(status)
	}
	p, status := DecodeM08Policy(raw[1])
	if status != missionValid {
		return fail(status)
	}
	values := make([]any, 7)
	for n, kind := range []string{"grant", "approval", "cost", "ledger", "gate", "authorization", "execution"} {
		value, state := DecodeM10Artifact(kind, raw[n+2])
		if state != missionValid {
			return fail(state)
		}
		values[n] = value
	}
	grant := values[0].(*CanaryGrant)
	approval := values[1].(*CanaryGrantApproval)
	cost := values[2].(*CanaryCostBound)
	ledger := values[3].(*CanaryLedger)
	gate := values[4].(*CanaryGateDecision)
	authorization := values[5].(*CanaryExecutionAuthorization)
	execution := values[6].(*CanaryExecutionRecord)
	links := make([]corem10.CanaryOutcomeLink, len(ledger.OutcomeLinks))
	for index, link := range ledger.OutcomeLinks {
		links[index] = corem10.CanaryOutcomeLink{OutcomeID: link.OutcomeID, ExecutionID: link.ExecutionID, ObservedAt: link.ObservedAt}
	}
	coreLedger := corem10.CanaryLedger{
		GrantID: ledger.GrantID, GrantVersion: ledger.GrantVersion, GrantHash: ledger.GrantHash,
		WindowStartedAt: ledger.WindowStartedAt, ExecutionsTotal: ledger.ExecutionsTotal,
		ExecutionsInWindow: ledger.ExecutionsInWindow, CostMinorTotal: ledger.CostMinorTotal,
		PendingOutcomes: ledger.PendingOutcomes, PendingExecutionIDs: ledger.PendingExecutionIDs,
		SuccessfulIdempotencyKeys: ledger.SuccessfulIdempotencyKeys, OutcomeLinks: links,
		ReconciliationRequired: ledger.ReconciliationRequired, LastExecutionAt: ledger.LastExecutionAt,
		UpdatedAt: ledger.UpdatedAt,
	}
	state := corem10.ValidateHistoricalChain(
		coreIntent(i), coreM10ChainPolicy(p), corem10.CanaryGrant(*grant),
		corem10.CanaryGrantApproval(*approval), corem10.TrustedCostBound(*cost),
		coreLedger, corem10.CanaryGateDecision(*gate), corem10.ExecutionAuthorization(*authorization),
		corem10.ExecutionRecord(*execution),
	)
	if state != "VALID" {
		return fail(state)
	}
	return M10ChainSummary{Result: "CONSISTENT_UNVERIFIED", IntentID: i.IntentID, GrantID: grant.GrantID, ExecutionID: execution.ExecutionID}, missionValid
}

func coreM10ChainPolicy(p ShadowPolicyDecision) corem08.PolicyDecision {
	return corem08.PolicyDecision{PolicyVersion: p.PolicyVersion, IntentID: p.IntentID, IntentHash: p.IntentHash, Decision: p.Decision, RiskClass: p.RiskClass, Reason: p.Reason, PolicyReviewRequired: p.PolicyReviewRequired, PolicyMode: p.PolicyMode, ExecutionAuthorized: p.ExecutionAuthorized, PolicyCheckedAt: p.PolicyCheckedAt}
}

func runM10ChainCheck(w io.Writer, args []string) error {
	if len(args) != 9 {
		return fmt.Errorf("usage: demo m10-chain-check INTENT POLICY GRANT APPROVAL COST PRE_GATE_LEDGER GATE AUTHORIZATION EXECUTION")
	}
	raw := make([][]byte, 9)
	for n, path := range args {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw[n] = data
	}
	summary, status := CheckM10Chain(raw)
	if status != missionValid {
		return fmt.Errorf("M10 chain audit: %s", status)
	}
	return json.NewEncoder(w).Encode(summary)
}
