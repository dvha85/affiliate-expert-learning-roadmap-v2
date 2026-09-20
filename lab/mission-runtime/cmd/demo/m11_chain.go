package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem03 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem05 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

type M11ChainSummary struct {
	Result                  string `json:"result"`
	Profile                 string `json:"profile"`
	LeaseID                 string `json:"lease_id"`
	ExecutionID             string `json:"execution_id"`
	ProvenanceAuthenticated bool   `json:"provenance_authenticated"`
	ExecutionPermitted      bool   `json:"execution_permitted"`
	ResumePermitted         bool   `json:"resume_permitted"`
}

// The bundle is a local audit transport, not a trusted state or canonical artifact.
func CheckM11Chain(raw []byte) (M11ChainSummary, string) {
	fail := func(s string) (M11ChainSummary, string) { return M11ChainSummary{}, s }
	if _, e := contracts.Decode(raw); e != nil {
		return fail("INVALID_SCHEMA")
	}
	var bundle map[string]json.RawMessage
	if json.Unmarshal(raw, &bundle) != nil || bundle == nil {
		return fail("INVALID_SCHEMA")
	}
	var profile string
	if json.Unmarshal(bundle["profile"], &profile) != nil || (profile != "closed_cycle" && profile != "resolved_stop") {
		return fail("INVALID_PROFILE")
	}
	kinds := map[string]string{"lease": "lease", "approval": "approval", "health": "health", "cost": "cost", "pre_ledger": "ledger", "post_ledger": "ledger", "gate": "gate", "authorization": "authorization", "execution": "execution", "activation": "activation"}
	allowed := map[string]bool{"profile": true, "intent": true, "policy": true}
	for k := range kinds {
		allowed[k] = true
	}
	if profile == "closed_cycle" {
		kinds["cycle"] = "cycle"
		allowed["cycle"] = true
		allowed["outcome"] = true
		allowed["evaluation"] = true
		allowed["proposal"] = true
		allowed["review"] = true
	} else {
		kinds["resolution"] = "resolution"
		allowed["resolution"] = true
		kinds["stop_ledger"] = "ledger"
		allowed["stop_ledger"] = true
	}
	for k := range bundle {
		if !allowed[k] {
			return fail("INVALID_PROFILE")
		}
	}
	i, s := DecodeM08Intent(bundle["intent"])
	if s != missionValid {
		return fail(s)
	}
	p, s := DecodeM08Policy(bundle["policy"])
	if s != missionValid {
		return fail(s)
	}
	v := map[string]any{}
	for k, kind := range kinds {
		x, s := DecodeM11Artifact(kind, bundle[k])
		if s != missionValid {
			return fail(s)
		}
		v[k] = x
	}
	var out OutcomeRecord
	var ev EvaluationRecord
	var prop *ImprovementProposal
	var review *ReviewRecord
	if profile == "closed_cycle" {
		out, s = DecodeM03Outcome(bundle["outcome"])
		if s != missionValid {
			return fail(s)
		}
		ev, s = DecodeM05Evaluation(bundle["evaluation"])
		if s != missionValid {
			return fail(s)
		}
		if b, ok := bundle["proposal"]; ok {
			x, s := DecodeM05Proposal(b)
			if s != missionValid {
				return fail(s)
			}
			prop = &x
		}
		if b, ok := bundle["review"]; ok {
			x, s := DecodeM05Review(b)
			if s != missionValid {
				return fail(s)
			}
			review = &x
		}
		if review != nil && prop == nil {
			return fail("BROKEN_LINK")
		}
	}
	l := v["lease"].(*ProductionLease)
	ap := v["approval"].(*ProductionLeaseApproval)
	h := v["health"].(*ProductionHealthSnapshot)
	c := v["cost"].(*CanaryCostBound)
	pre := v["pre_ledger"].(*ProductionLedger)
	post := v["post_ledger"].(*ProductionLedger)
	g := v["gate"].(*ProductionGateDecision)
	a := v["authorization"].(*ProductionExecutionAuthorization)
	r := v["execution"].(*ProductionExecutionRecord)
	activation := v["activation"].(*productionActivationRecord)
	toCore := func(in any, out any) bool {
		raw, err := json.Marshal(in)
		return err == nil && json.Unmarshal(raw, out) == nil
	}
	// DecodeM08Intent deliberately keeps parameters as json.Number so the
	// canonical intent hash can distinguish integers beyond float64's 2^53
	// precision boundary. Do not marshal the intent through map[string]any
	// again here: that conversion would round a valid large integer before the
	// shared historical validator recomputes the hash.
	ci := coreIntent(i)
	var cp corem08.PolicyDecision
	var cl corem11.ProductionLease
	var ca corem11.ProductionLeaseApproval
	var ch corem11.ProductionHealthSnapshot
	var cc corem10.TrustedCostBound
	var cpre, cpost corem11.ProductionLedger
	var cg corem11.ProductionGateDecision
	var cauth corem11.ProductionExecutionAuthorization
	var cexec corem11.ProductionExecutionRecord
	var cactivation corem11.ProductionActivationRecord
	if !toCore(p, &cp) || !toCore(*l, &cl) || !toCore(*ap, &ca) || !toCore(*h, &ch) || !toCore(*c, &cc) || !toCore(*pre, &cpre) || !toCore(*post, &cpost) || !toCore(*g, &cg) || !toCore(*a, &cauth) || !toCore(*r, &cexec) || !toCore(*activation, &cactivation) {
		return fail("INVALID_SCHEMA")
	}
	chain := corem11.HistoricalChain{Profile: profile, Intent: ci, Policy: cp, Lease: cl, Approval: ca, Health: ch, Cost: cc, PreLedger: cpre, PostLedger: cpost, Gate: cg, Authorization: cauth, Execution: cexec, Activation: cactivation}
	if profile == "closed_cycle" {
		var co corem03.OutcomeRecord
		var ce corem05.EvaluationRecord
		if !toCore(out, &co) || !toCore(ev, &ce) {
			return fail("INVALID_SCHEMA")
		}
		chain.Outcome, chain.Evaluation = &co, &ce
		if prop != nil {
			var cpv corem05.ImprovementProposal
			if !toCore(*prop, &cpv) {
				return fail("INVALID_SCHEMA")
			}
			chain.Proposal = &cpv
		}
		if review != nil {
			var cr corem05.ReviewRecord
			if !toCore(*review, &cr) {
				return fail("INVALID_SCHEMA")
			}
			chain.Review = &cr
		}
		var ccycle corem11.ProductionCycleRecord
		if !toCore(*v["cycle"].(*ProductionCycleRecord), &ccycle) {
			return fail("INVALID_SCHEMA")
		}
		chain.Cycle = &ccycle
	} else {
		var cres corem11.ProductionReconciliationResolution
		var cstop corem11.ProductionLedger
		if !toCore(*v["resolution"].(*ProductionReconciliationResolution), &cres) || !toCore(*v["stop_ledger"].(*ProductionLedger), &cstop) {
			return fail("INVALID_SCHEMA")
		}
		chain.Resolution, chain.StopLedger = &cres, &cstop
	}
	if status := corem11.ValidateHistoricalChain(chain); status != corem11.Valid && status != "REVIEW_REQUIRED" {
		return fail(status)
	}
	return M11ChainSummary{Result: "CONSISTENT_UNVERIFIED", Profile: profile, LeaseID: l.LeaseID, ExecutionID: r.ExecutionID}, missionValid
}

func runM11ChainCheck(w io.Writer, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: demo m11-chain-check BUNDLE.json")
	}
	b, e := os.ReadFile(args[0])
	if e != nil {
		return e
	}
	s, status := CheckM11Chain(b)
	if status != missionValid {
		return fmt.Errorf("M11 chain audit: %s", status)
	}
	return json.NewEncoder(w).Encode(s)
}
