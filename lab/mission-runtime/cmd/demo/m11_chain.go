package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
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
	// All fields below have already passed canonical raw boundaries. Compare exact IDs.
	fields := map[string]map[string]json.RawMessage{}
	for k, b := range bundle {
		if k != "profile" {
			var m map[string]json.RawMessage
			json.Unmarshal(b, &m)
			fields[k] = m
		}
	}
	str := func(path string) string {
		parts := strings.SplitN(path, ".", 2)
		var value string
		json.Unmarshal(fields[parts[0]][parts[1]], &value)
		return value
	}
	match := func(source string, targets ...string) bool {
		want := str(source)
		if strings.TrimSpace(want) == "" {
			return false
		}
		for _, target := range targets {
			if str(target) != want {
				return false
			}
		}
		return true
	}
	for _, field := range []string{"lease_id", "lease_version", "lease_hash"} {
		targets := []string{}
		for _, k := range []string{"approval", "health", "pre_ledger", "post_ledger", "gate", "activation"} {
			targets = append(targets, k+"."+field)
		}
		for _, k := range []string{"authorization", "execution"} {
			targets = append(targets, k+".production_"+field)
		}
		if profile == "closed_cycle" {
			targets = append(targets, "cycle."+field)
		} else {
			targets = append(targets, "resolution."+field, "stop_ledger."+field)
		}
		if !match("lease."+field, targets...) {
			return fail("BROKEN_LINK")
		}
	}
	for _, f := range []string{"promotion_review_ref", "source_canary_grant_id", "source_canary_grant_version", "source_canary_grant_hash", "reviewer_id"} {
		if !match("lease."+f, "approval."+f) {
			return fail("BROKEN_LINK")
		}
	}
	if !match("lease.approval_ref", "approval.approval_id") || !match("intent.intent_id", "policy.intent_id", "cost.intent_id", "gate.intent_id", "authorization.intent_id", "execution.intent_id") || !match("intent.intent_hash", "policy.intent_hash", "cost.intent_hash", "gate.intent_hash", "authorization.intent_hash", "execution.intent_hash") || !match("policy.policy_version", "lease.policy_version", "gate.policy_version", "authorization.policy_version") || !match("intent.correlation_id", "cost.correlation_id", "authorization.correlation_id", "execution.correlation_id") || !match("intent.idempotency_key", "authorization.idempotency_key", "execution.idempotency_key") || !match("gate.gate_id", "authorization.production_gate_id", "execution.production_gate_id") || !match("health.snapshot_id", "gate.health_snapshot_id", "authorization.production_health_snapshot_id", "execution.production_health_snapshot_id") || !match("health.snapshot_hash", "gate.health_snapshot_hash", "authorization.production_health_snapshot_hash", "execution.production_health_snapshot_hash") || !match("cost.cost_bound_id", "gate.cost_bound_id", "authorization.production_cost_bound_id", "execution.production_cost_bound_id") || !match("cost.cost_bound_hash", "gate.cost_bound_hash", "authorization.production_cost_bound_hash", "execution.production_cost_bound_hash") || !match("authorization.authorization_id", "execution.authorization_id") || !match("authorization.executor_id", "execution.executor_id") {
		return fail("BROKEN_LINK")
	}
	if i.IntentHash != ComputeShadowIntentHash(i) {
		return fail("TAMPERED_INTENT")
	}
	if c.Currency != l.Currency || c.MaxCostMinor != g.CostBoundMinor || c.MaxCostMinor != a.ProductionCostBoundMinor || c.MaxCostMinor != r.ProductionCostBoundMinor || g.RiskClass != p.RiskClass {
		return fail("BROKEN_LINK")
	}
	if g.Decision != "ALLOW_PRODUCTION" || g.Reason != "PRODUCTION_ELIGIBLE" || !(p.RiskClass == "RISK0" && p.Decision == "ALLOW" || p.RiskClass == "RISK1" && p.Decision == "HUMAN_REVIEW" && p.Reason == "RISK1_REQUIRES_REVIEW") {
		return fail("INVALID_GATE_STATE")
	}
	if !containsFold(l.AllowedRiskClasses, p.RiskClass) || !containsFold(ap.ValidatedRiskClasses, p.RiskClass) || !containsFold(l.AllowedActionTypes, i.ActionType) || !allowedHost(i.Target, l.AllowedHosts) || !containsFold(l.ExecutorIDs, a.ExecutorID) {
		return fail("SCOPE_NOT_DELEGATED")
	}
	if pre.ControlMode != "NORMAL" || pre.StopReason != "" || pre.ReconciliationRequired || len(pre.ReconciliationResolutionIDs) != 0 || pre.ConsecutiveFailures >= l.MaxConsecutiveFailures || h.ReconciliationRequired || h.ComplianceAlertCount > 0 || h.ConsecutiveFailures >= l.MaxConsecutiveFailures || h.OldestPendingOutcomeAgeSeconds > l.MaxOutcomeAgeSeconds || h.DependencyState != "HEALTHY" || !h.TelemetryComplete {
		return fail("SAFETY_BLOCKED")
	}
	// RFC3339 schema already checked every required timestamp; no current clock used.
	at := func(x string) time.Time { t, _ := time.Parse(time.RFC3339, x); return t }
	now, start, end := at(g.EvaluatedAt), at(l.ValidFrom), at(a.ExpiresAt)
	if !at(ap.ReviewedAt).Equal(at(l.ReviewedAt)) || now.Before(start) || !now.Before(at(l.ExpiresAt)) || !now.Before(at(i.ExpiresAt)) || !at(i.ExpiresAt).After(at(i.CreatedAt)) || at(p.PolicyCheckedAt).Before(at(i.CreatedAt)) || at(p.PolicyCheckedAt).After(now) || at(c.ObservedAt).After(now) || !now.Before(at(c.ExpiresAt)) || at(h.ObservedAt).After(now) || !at(a.AuthorizedAt).Equal(now) || end.After(at(l.ExpiresAt)) || end.After(at(i.ExpiresAt)) || end.After(at(c.ExpiresAt)) || at(r.AttemptedAt).Before(now) || !at(r.AttemptedAt).Before(end) || at(activation.ActivatedAt).Before(start) || at(activation.ActivatedAt).After(now) || at(pre.WindowStartedAt).Before(start) || at(pre.UpdatedAt).After(now) || at(pre.WindowStartedAt).After(now) || at(post.UpdatedAt).Before(at(r.AttemptedAt)) {
		return fail("INVALID_TIME_BINDING")
	}
	// Compare age to integer seconds without overflowing time.Duration.
	ageTooOld := func(t time.Time, seconds int) bool {
		d := now.Unix() - t.Unix()
		return d > int64(seconds) || d == int64(seconds) && now.Nanosecond() > t.Nanosecond()
	}
	if ageTooOld(at(h.ObservedAt), l.MaxHealthSnapshotAgeSeconds) {
		return fail("STALE_HEALTH")
	}
	window := at(pre.WindowStartedAt)
	elapsed := now.Unix() - window.Unix()
	if now.Nanosecond() < window.Nanosecond() {
		elapsed--
	}
	count := pre.ExecutionsInWindow
	if elapsed >= int64(l.WindowSeconds) {
		count = 0
		window = now
	}
	if g.ExecutionsTotalBefore != pre.ExecutionsTotal || g.ExecutionsInWindowBefore != count || g.CostMinorTotalBefore != pre.CostMinorTotal || g.PendingOutcomesBefore != pre.PendingOutcomes {
		return fail("LEDGER_SNAPSHOT_MISMATCH")
	}
	if containsFold(pre.SuccessfulIdempotencyKeys, i.IdempotencyKey) || containsFold(pre.PendingExecutionIDs, r.ExecutionID) {
		return fail("REPLAYED_EXECUTION")
	}
	for _, link := range pre.OutcomeLinks {
		if link.ExecutionID == r.ExecutionID {
			return fail("REPLAYED_EXECUTION")
		}
	}
	if pre.ExecutionsTotal >= l.MaxExecutionsTotal || count >= l.MaxExecutionsPerWindow || pre.PendingOutcomes >= l.MaxPendingOutcomes || pre.CostMinorTotal > l.MaxCostMinorTotal || c.MaxCostMinor > l.MaxCostMinorTotal-pre.CostMinorTotal {
		return fail("BUDGET_EXCEEDED")
	}
	// A single execution transition cannot erase history or reset budget.
	if profile == "closed_cycle" && (post.ExecutionsTotal != pre.ExecutionsTotal+1 || post.ExecutionsInWindow != count+1 || post.CostMinorTotal != pre.CostMinorTotal+c.MaxCostMinor || !at(post.WindowStartedAt).Equal(window) || !at(post.LastExecutionAt).Equal(at(r.AttemptedAt))) {
		return fail("INVALID_LEDGER_TRANSITION")
	}
	if profile == "resolved_stop" {
		x := v["resolution"].(*ProductionReconciliationResolution)
		stop := v["stop_ledger"].(*ProductionLedger)
		expectedStop := *pre
		expectedStop.ControlMode = "STOPPED"
		expectedStop.StopReason = "RECONCILIATION_REQUIRED"
		expectedStop.ReconciliationRequired = true
		expectedStop.UpdatedAt = r.AttemptedAt
		if !reflect.DeepEqual(*stop, expectedStop) {
			return fail("INVALID_STOP_TRANSITION")
		}
		expectedPost := *stop
		expectedPost.ReconciliationRequired = false
		expectedPost.StopReason = "RECOVERY_REVIEW_REQUIRED"
		expectedPost.UpdatedAt = x.ResolvedAt
		expectedPost.ReconciliationResolutionIDs = []string{x.ResolutionID}
		if x.EffectState == "PERFORMED" {
			expectedPost.ExecutionsTotal++
			expectedPost.ExecutionsInWindow++
			expectedPost.CostMinorTotal += c.MaxCostMinor
			expectedPost.PendingOutcomes++
			expectedPost.PendingExecutionIDs = append(append([]string{}, pre.PendingExecutionIDs...), r.ExecutionID)
			expectedPost.SuccessfulIdempotencyKeys = append(append([]string{}, pre.SuccessfulIdempotencyKeys...), i.IdempotencyKey)
		}
		if x.ExecutionID != r.ExecutionID || r.Status != "RECONCILIATION_REQUIRED" || r.SideEffectState != "UNKNOWN" || at(x.ResolvedAt).Before(at(r.AttemptedAt)) || !reflect.DeepEqual(*post, expectedPost) {
			return fail("INVALID_STOP_TRANSITION")
		}
	} else {
		cycle := v["cycle"].(*ProductionCycleRecord)
		state := M11State{Intent: i, Policy: p, Lease: *l, Health: *h, CostBound: *c, Ledger: *pre}
		if s := ValidateCanonicalProductionClosedCycle(*cycle, state, *g, *a, *r, out, ev, prop, review); s != missionValid && s != "REVIEW_REQUIRED" {
			return fail(s)
		}
		// Existing legacy validator uses case-folded membership; audit uses exact sets.
		if !reflect.DeepEqual(cycle.ObservationIDs, i.EvidenceIDs) || !reflect.DeepEqual(ev.OutcomeIDs, []string{out.OutcomeID}) || out.Status != "VALID" || at(cycle.OpenedAt).After(now) || at(cycle.ClosedAt).Before(at(ev.EvaluatedAt)) || !at(post.UpdatedAt).Equal(at(out.ObservedAt)) {
			return fail("BROKEN_CYCLE")
		}
		for _, id := range ev.EvidenceIDs {
			found := false
			for _, known := range i.EvidenceIDs {
				if id == known {
					found = true
				}
			}
			if !found {
				return fail("BROKEN_CYCLE")
			}
		}
		if prop != nil && (!reflect.DeepEqual(prop.EvaluationIDs, []string{ev.EvaluationID}) || prop.AutoApply) {
			return fail("BROKEN_CYCLE")
		}
		if review != nil && (at(review.ReviewedAt).Before(at(ev.EvaluatedAt)) || at(review.ReviewedAt).After(at(cycle.ClosedAt))) {
			return fail("INVALID_TIME_BINDING")
		}
		links := append([]ProductionOutcomeLink{}, pre.OutcomeLinks...)
		links = append(links, ProductionOutcomeLink{OutcomeID: out.OutcomeID, ExecutionID: r.ExecutionID, ObservedAt: out.ObservedAt})
		keys := append([]string{}, pre.SuccessfulIdempotencyKeys...)
		keys = append(keys, i.IdempotencyKey)
		if post.ControlMode != "NORMAL" || post.StopReason != "" || post.ReconciliationRequired || post.ConsecutiveFailures != 0 || post.PendingOutcomes != pre.PendingOutcomes || !reflect.DeepEqual(post.PendingExecutionIDs, pre.PendingExecutionIDs) || !reflect.DeepEqual(post.OutcomeLinks, links) || !reflect.DeepEqual(post.SuccessfulIdempotencyKeys, keys) || !reflect.DeepEqual(post.ReconciliationResolutionIDs, pre.ReconciliationResolutionIDs) || !at(post.LastOutcomeAt).Equal(at(out.ObservedAt)) {
			return fail("INVALID_LEDGER_TRANSITION")
		}
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
