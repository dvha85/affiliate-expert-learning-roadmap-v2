package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CanaryGrant struct {
	GrantID                string   `json:"grant_id"`
	GrantVersion           string   `json:"grant_version"`
	PolicyVersion          string   `json:"policy_version"`
	ApprovalRef            string   `json:"approval_ref"`
	ApprovedBy             string   `json:"approved_by"`
	ApproverID             string   `json:"approver_id"`
	ApprovedAt             string   `json:"approved_at"`
	ValidFrom              string   `json:"valid_from"`
	ExpiresAt              string   `json:"expires_at"`
	AllowedRiskClasses     []string `json:"allowed_risk_classes"`
	AllowedActionTypes     []string `json:"allowed_action_types"`
	AllowedHosts           []string `json:"allowed_hosts"`
	ExecutorIDs            []string `json:"executor_ids"`
	MaxExecutionsTotal     int      `json:"max_executions_total"`
	MaxExecutionsPerWindow int      `json:"max_executions_per_window"`
	WindowSeconds          int      `json:"window_seconds"`
	MaxCostMinorTotal      int64    `json:"max_cost_minor_total"`
	Currency               string   `json:"currency"`
	MaxPendingOutcomes     int      `json:"max_pending_outcomes"`
	KillSwitchRequired     bool     `json:"kill_switch_required"`
	CorrelationID          string   `json:"correlation_id"`
	HashVersion            string   `json:"hash_version"`
	GrantHash              string   `json:"grant_hash"`
}

type CanaryCostBound struct {
	CostBoundID   string `json:"cost_bound_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	MaxCostMinor  int64  `json:"max_cost_minor"`
	Currency      string `json:"currency"`
	SourceRef     string `json:"source_ref"`
	ObservedAt    string `json:"observed_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	HashVersion   string `json:"hash_version"`
	CostBoundHash string `json:"cost_bound_hash"`
}

type CanaryOutcomeLink struct {
	OutcomeID   string `json:"outcome_id"`
	ExecutionID string `json:"execution_id"`
	ObservedAt  string `json:"observed_at"`
}

type CanaryLedger struct {
	GrantID                   string              `json:"grant_id"`
	GrantVersion              string              `json:"grant_version"`
	GrantHash                 string              `json:"grant_hash"`
	WindowStartedAt           string              `json:"window_started_at"`
	ExecutionsTotal           int                 `json:"executions_total"`
	ExecutionsInWindow        int                 `json:"executions_in_window"`
	CostMinorTotal            int64               `json:"cost_minor_total"`
	PendingOutcomes           int                 `json:"pending_outcomes"`
	PendingExecutionIDs       []string            `json:"pending_execution_ids"`
	SuccessfulIdempotencyKeys []string            `json:"successful_idempotency_keys"`
	OutcomeLinks              []CanaryOutcomeLink `json:"outcome_links"`
	ReconciliationRequired    bool                `json:"reconciliation_required"`
	LastExecutionAt           string              `json:"last_execution_at,omitempty"`
	UpdatedAt                 string              `json:"updated_at"`
}

type CanaryGateDecision struct {
	GateID                    string `json:"gate_id"`
	GrantID                   string `json:"grant_id"`
	GrantVersion              string `json:"grant_version"`
	GrantHash                 string `json:"grant_hash"`
	IntentID                  string `json:"intent_id"`
	IntentHash                string `json:"intent_hash"`
	PolicyVersion             string `json:"policy_version"`
	RiskClass                 string `json:"risk_class"`
	CostBoundID               string `json:"cost_bound_id"`
	CostBoundHash             string `json:"cost_bound_hash"`
	CostBoundMinor            int64  `json:"cost_bound_minor"`
	Decision                  string `json:"decision"`
	Reason                    string `json:"reason"`
	EvaluatedAt               string `json:"evaluated_at"`
	ExecutionsTotalBefore     int    `json:"executions_total_before"`
	ExecutionsInWindowBefore  int    `json:"executions_in_window_before"`
	CostMinorTotalBefore      int64  `json:"cost_minor_total_before"`
	PendingOutcomesBefore     int    `json:"pending_outcomes_before"`
	PerActionApprovalRequired bool   `json:"per_action_approval_required"`
	ExecutionAuthorized       bool   `json:"execution_authorized"`
}

type CanaryExecutionAuthorization struct {
	AuthorizationID      string `json:"authorization_id"`
	IntentID             string `json:"intent_id"`
	IntentHash           string `json:"intent_hash"`
	PolicyVersion        string `json:"policy_version"`
	CanaryGrantID        string `json:"canary_grant_id"`
	CanaryGrantVersion   string `json:"canary_grant_version"`
	CanaryGrantHash      string `json:"canary_grant_hash"`
	CanaryGateID         string `json:"canary_gate_id"`
	CanaryCostBoundID    string `json:"canary_cost_bound_id"`
	CanaryCostBoundHash  string `json:"canary_cost_bound_hash"`
	CanaryCostBoundMinor int64  `json:"canary_cost_bound_minor"`
	ExecutorID           string `json:"executor_id"`
	AuthorizedAt         string `json:"authorized_at"`
	ExpiresAt            string `json:"expires_at"`
	IdempotencyKey       string `json:"idempotency_key"`
	CorrelationID        string `json:"correlation_id"`
	ExecutionMode        string `json:"execution_mode"`
	ExecutionAuthorized  bool   `json:"execution_authorized"`
}

type CanaryExecutionRecord struct {
	ExecutionID          string `json:"execution_id"`
	AuthorizationID      string `json:"authorization_id"`
	CanaryGrantID        string `json:"canary_grant_id"`
	CanaryGrantVersion   string `json:"canary_grant_version"`
	CanaryGrantHash      string `json:"canary_grant_hash"`
	CanaryGateID         string `json:"canary_gate_id"`
	CanaryCostBoundID    string `json:"canary_cost_bound_id"`
	CanaryCostBoundHash  string `json:"canary_cost_bound_hash"`
	CanaryCostBoundMinor int64  `json:"canary_cost_bound_minor"`
	IntentID             string `json:"intent_id"`
	IntentHash           string `json:"intent_hash"`
	ExecutorID           string `json:"executor_id"`
	IdempotencyKey       string `json:"idempotency_key"`
	AttemptedAt          string `json:"attempted_at"`
	Status               string `json:"status"`
	SideEffectState      string `json:"side_effect_state"`
	ExternalRef          string `json:"external_ref,omitempty"`
	Error                string `json:"error,omitempty"`
	CorrelationID        string `json:"correlation_id"`
}

type M10State struct {
	Intent        ShadowActionIntent             `json:"intent"`
	Policy        ShadowPolicyDecision           `json:"policy"`
	Grant         CanaryGrant                    `json:"grant"`
	CostBound     CanaryCostBound                `json:"cost_bound"`
	Ledger        CanaryLedger                   `json:"ledger"`
	Authorization *CanaryExecutionAuthorization `json:"authorization,omitempty"`
	Execution     *CanaryExecutionRecord         `json:"execution,omitempty"`
}

type M10Context struct {
	Now                 string
	KillSwitch          bool
	RevokedGrantIDs     []string
	ApprovedGrantHashes map[string]string
	TrustedCostBounds   map[string]string
	Executor            ExecutorProfile
	AllowedExecutorIDs  []string
	PolicyContext       ShadowPolicyContext
}

type canaryGrantHashPayload struct {
	GrantID                string   `json:"grant_id"`
	GrantVersion           string   `json:"grant_version"`
	PolicyVersion          string   `json:"policy_version"`
	ApprovalRef            string   `json:"approval_ref"`
	ApprovedBy             string   `json:"approved_by"`
	ApproverID             string   `json:"approver_id"`
	ApprovedAt             string   `json:"approved_at"`
	ValidFrom              string   `json:"valid_from"`
	ExpiresAt              string   `json:"expires_at"`
	AllowedRiskClasses     []string `json:"allowed_risk_classes"`
	AllowedActionTypes     []string `json:"allowed_action_types"`
	AllowedHosts           []string `json:"allowed_hosts"`
	ExecutorIDs            []string `json:"executor_ids"`
	MaxExecutionsTotal     int      `json:"max_executions_total"`
	MaxExecutionsPerWindow int      `json:"max_executions_per_window"`
	WindowSeconds          int      `json:"window_seconds"`
	MaxCostMinorTotal      int64    `json:"max_cost_minor_total"`
	Currency               string   `json:"currency"`
	MaxPendingOutcomes     int      `json:"max_pending_outcomes"`
	KillSwitchRequired     bool     `json:"kill_switch_required"`
	CorrelationID          string   `json:"correlation_id"`
	HashVersion            string   `json:"hash_version"`
}

type canaryCostHashPayload struct {
	CostBoundID   string `json:"cost_bound_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	MaxCostMinor  int64  `json:"max_cost_minor"`
	Currency      string `json:"currency"`
	SourceRef     string `json:"source_ref"`
	ObservedAt    string `json:"observed_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	HashVersion   string `json:"hash_version"`
}

func sortedCopy(xs []string) []string {
	out := append([]string(nil), xs...)
	sort.Strings(out)
	return out
}

func ComputeCanaryGrantHash(g CanaryGrant) string {
	payload := canaryGrantHashPayload{
		GrantID: g.GrantID, GrantVersion: g.GrantVersion, PolicyVersion: g.PolicyVersion,
		ApprovalRef: g.ApprovalRef, ApprovedBy: g.ApprovedBy, ApproverID: g.ApproverID,
		ApprovedAt: g.ApprovedAt, ValidFrom: g.ValidFrom, ExpiresAt: g.ExpiresAt,
		AllowedRiskClasses: sortedCopy(g.AllowedRiskClasses), AllowedActionTypes: sortedCopy(g.AllowedActionTypes),
		AllowedHosts: sortedCopy(g.AllowedHosts), ExecutorIDs: sortedCopy(g.ExecutorIDs),
		MaxExecutionsTotal: g.MaxExecutionsTotal, MaxExecutionsPerWindow: g.MaxExecutionsPerWindow,
		WindowSeconds: g.WindowSeconds, MaxCostMinorTotal: g.MaxCostMinorTotal, Currency: g.Currency,
		MaxPendingOutcomes: g.MaxPendingOutcomes, KillSwitchRequired: g.KillSwitchRequired,
		CorrelationID: g.CorrelationID, HashVersion: g.HashVersion,
	}
	raw, err := json.Marshal(payload)
	if err != nil { return "" }
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func SealCanaryGrant(g CanaryGrant) CanaryGrant {
	if g.HashVersion == "" { g.HashVersion = "go-json-v1" }
	g.Currency = strings.ToUpper(strings.TrimSpace(g.Currency))
	g.GrantHash = ComputeCanaryGrantHash(g)
	return g
}

func ComputeCanaryCostBoundHash(c CanaryCostBound) string {
	payload := canaryCostHashPayload{
		CostBoundID: c.CostBoundID, IntentID: c.IntentID, IntentHash: c.IntentHash,
		MaxCostMinor: c.MaxCostMinor, Currency: strings.ToUpper(strings.TrimSpace(c.Currency)),
		SourceRef: c.SourceRef, ObservedAt: c.ObservedAt, ExpiresAt: c.ExpiresAt,
		CorrelationID: c.CorrelationID, HashVersion: c.HashVersion,
	}
	raw, err := json.Marshal(payload)
	if err != nil { return "" }
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func SealCanaryCostBound(c CanaryCostBound) CanaryCostBound {
	if c.HashVersion == "" { c.HashVersion = "go-json-v1" }
	c.Currency = strings.ToUpper(strings.TrimSpace(c.Currency))
	c.CostBoundHash = ComputeCanaryCostBoundHash(c)
	return c
}

func hasWildcard(xs []string) bool {
	for _, x := range xs { if strings.TrimSpace(x) == "*" { return true } }
	return false
}

func normalizedCanaryLedger(l CanaryLedger, g CanaryGrant, now time.Time) (CanaryLedger, bool) {
	if l.GrantID != g.GrantID || l.GrantVersion != g.GrantVersion || l.GrantHash != g.GrantHash ||
		l.ExecutionsTotal < 0 || l.ExecutionsInWindow < 0 || l.CostMinorTotal < 0 || l.PendingOutcomes < 0 ||
		l.PendingOutcomes != len(l.PendingExecutionIDs) {
		return l, false
	}
	started, err := time.Parse(time.RFC3339, l.WindowStartedAt)
	if err != nil || now.Before(started) { return l, false }
	if now.Sub(started) >= time.Duration(g.WindowSeconds)*time.Second {
		l.WindowStartedAt = now.Format(time.RFC3339)
		l.ExecutionsInWindow = 0
	}
	return l, true
}

func newCanaryGate(state M10State, ctx M10Context) CanaryGateDecision {
	return CanaryGateDecision{
		GateID: "gate-" + state.Grant.GrantID + "-" + state.Intent.IntentID,
		GrantID: state.Grant.GrantID, GrantVersion: state.Grant.GrantVersion, GrantHash: state.Grant.GrantHash,
		IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: state.Policy.PolicyVersion,
		RiskClass: state.Policy.RiskClass, CostBoundID: state.CostBound.CostBoundID, CostBoundHash: state.CostBound.CostBoundHash,
		CostBoundMinor: state.CostBound.MaxCostMinor, Decision: "DENY", Reason: "INVALID_CANARY_STATE", EvaluatedAt: ctx.Now,
		ExecutionsTotalBefore: state.Ledger.ExecutionsTotal, ExecutionsInWindowBefore: state.Ledger.ExecutionsInWindow,
		CostMinorTotalBefore: state.Ledger.CostMinorTotal, PendingOutcomesBefore: state.Ledger.PendingOutcomes,
		ExecutionAuthorized: false,
	}
}

func requireApproval(g *CanaryGateDecision, reason string) CanaryGateDecision {
	g.Decision = "REQUIRE_APPROVAL"; g.Reason = reason; g.PerActionApprovalRequired = true; return *g
}

func EvaluateCanaryGate(state M10State, ctx M10Context) CanaryGateDecision {
	g := newCanaryGate(state, ctx)
	i, p, grant, costBound := state.Intent, state.Policy, state.Grant, state.CostBound
	if i.IntentHash == "" || i.IntentHash != ComputeShadowIntentHash(i) || !i.ShadowOnly || !i.DryRun {
		g.Reason = "INVALID_INTENT_BINDING"; return g
	}
	if grant.HashVersion != "go-json-v1" || grant.GrantHash == "" || grant.GrantHash != ComputeCanaryGrantHash(grant) {
		g.Reason = "TAMPERED_GRANT"; return g
	}
	if grant.ApprovedBy != "human" || strings.TrimSpace(grant.ApproverID) == "" || strings.TrimSpace(grant.ApprovalRef) == "" ||
		strings.TrimSpace(grant.GrantID) == "" || strings.TrimSpace(grant.GrantVersion) == "" || !grant.KillSwitchRequired ||
		grant.MaxExecutionsTotal < 1 || grant.MaxExecutionsPerWindow < 1 || grant.MaxExecutionsPerWindow > grant.MaxExecutionsTotal ||
		grant.WindowSeconds < 60 || grant.MaxCostMinorTotal < 0 || grant.MaxPendingOutcomes < 1 ||
		len(grant.AllowedRiskClasses) == 0 || len(grant.AllowedActionTypes) == 0 || len(grant.AllowedHosts) == 0 || len(grant.ExecutorIDs) == 0 ||
		hasWildcard(grant.AllowedActionTypes) || hasWildcard(grant.AllowedHosts) || hasWildcard(grant.ExecutorIDs) {
		g.Reason = "INVALID_GRANT"; return g
	}
	for _, risk := range grant.AllowedRiskClasses { if risk != "RISK0" && risk != "RISK1" { g.Reason = "INVALID_GRANT"; return g } }
	if len(grant.Currency) != 3 || grant.Currency != strings.ToUpper(grant.Currency) { g.Reason = "INVALID_GRANT"; return g }
	if ctx.ApprovedGrantHashes == nil || ctx.TrustedCostBounds == nil { g.Reason = "GOVERNANCE_INPUTS_UNAVAILABLE"; return g }
	approvedHash, ok := ctx.ApprovedGrantHashes[grant.ApprovalRef]
	if !ok { g.Reason = "GRANT_PROVENANCE_MISSING"; return g }
	if approvedHash != grant.GrantHash { g.Reason = "GRANT_APPROVAL_MISMATCH"; return g }
	if containsFold(ctx.RevokedGrantIDs, grant.GrantID) { g.Reason = "GRANT_REVOKED"; return g }
	if ctx.KillSwitch { g.Reason = "KILL_SWITCH_ACTIVE"; return g }

	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	approvedAt, errApproved := time.Parse(time.RFC3339, grant.ApprovedAt)
	validFrom, errValid := time.Parse(time.RFC3339, grant.ValidFrom)
	grantExpires, errGrantExpires := time.Parse(time.RFC3339, grant.ExpiresAt)
	if errNow != nil || errApproved != nil || errValid != nil || errGrantExpires != nil || approvedAt.After(validFrom) || !grantExpires.After(validFrom) {
		g.Reason = "INVALID_GRANT_TIME"; return g
	}
	if now.Before(validFrom) { g.Decision = "WAIT"; g.Reason = "GRANT_NOT_YET_ACTIVE"; return g }
	if !grantExpires.After(now) { return requireApproval(&g, "GRANT_EXPIRED") }

	policyChecked, errPolicyChecked := time.Parse(time.RFC3339, p.PolicyCheckedAt)
	if errPolicyChecked != nil || policyChecked.After(now) || p.PolicyVersion == "" || p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || !p.ShadowOnly || p.ExecutionAuthorized {
		g.Reason = "INVALID_POLICY_STATE"; return g
	}
	freshCtx := ctx.PolicyContext; freshCtx.Now = ctx.Now
	fresh := EvaluateShadowPolicy(i, freshCtx)
	if fresh.Reason == "POLICY_UNAVAILABLE" || !policyEquivalentForExecution(p, fresh) { g.Reason = "POLICY_REVALIDATION_FAILED"; return g }
	g.RiskClass = fresh.RiskClass; g.PolicyVersion = fresh.PolicyVersion
	if grant.PolicyVersion != fresh.PolicyVersion { g.Reason = "GRANT_POLICY_MISMATCH"; return g }
	if fresh.RiskClass == "RISK2" { return requireApproval(&g, "RISK2_PER_ACTION_APPROVAL_REQUIRED") }
	if fresh.RiskClass == "RISK0" && fresh.Decision != "ALLOW" { g.Reason = "POLICY_STATE_NOT_CANARY_ELIGIBLE"; return g }
	if fresh.RiskClass == "RISK1" && (fresh.Decision != "HUMAN_REVIEW" || fresh.Reason != "RISK1_REQUIRES_REVIEW") { g.Reason = "POLICY_STATE_NOT_CANARY_ELIGIBLE"; return g }
	if fresh.RiskClass != "RISK0" && fresh.RiskClass != "RISK1" { g.Reason = "UNKNOWN_RISK_CLASS"; return g }
	if !containsFold(grant.AllowedRiskClasses, fresh.RiskClass) { return requireApproval(&g, "RISK_NOT_DELEGATED") }
	if !containsFold(grant.AllowedActionTypes, i.ActionType) || !allowedHost(i.Target, grant.AllowedHosts) { return requireApproval(&g, "SCOPE_NOT_DELEGATED") }
	if !containsFold(grant.ExecutorIDs, ctx.Executor.ExecutorID) || !containsFold(ctx.AllowedExecutorIDs, ctx.Executor.ExecutorID) || !executorAllows(ctx.Executor, i) {
		g.Reason = "EXECUTOR_NOT_ALLOWED"; return g
	}

	if strings.TrimSpace(costBound.CostBoundID) == "" { g.Reason = "COST_BOUND_MISSING"; return g }
	if costBound.HashVersion != "go-json-v1" || costBound.CostBoundHash == "" || costBound.CostBoundHash != ComputeCanaryCostBoundHash(costBound) {
		g.Reason = "TAMPERED_COST_BOUND"; return g
	}
	if costBound.IntentID != i.IntentID || costBound.IntentHash != i.IntentHash || costBound.CorrelationID != i.CorrelationID ||
		costBound.MaxCostMinor < 0 || costBound.Currency != grant.Currency || strings.TrimSpace(costBound.SourceRef) == "" {
		g.Reason = "COST_BOUND_MISMATCH"; return g
	}
	trustedCostHash, ok := ctx.TrustedCostBounds[costBound.CostBoundID]
	if !ok || trustedCostHash != costBound.CostBoundHash { g.Reason = "COST_BOUND_UNTRUSTED"; return g }
	observedAt, errObserved := time.Parse(time.RFC3339, costBound.ObservedAt)
	costExpires, errCostExpires := time.Parse(time.RFC3339, costBound.ExpiresAt)
	if errObserved != nil || errCostExpires != nil || observedAt.After(now) || !costExpires.After(observedAt) || !costExpires.After(now) {
		g.Reason = "COST_BOUND_EXPIRED"; return g
	}
	g.CostBoundID = costBound.CostBoundID; g.CostBoundHash = costBound.CostBoundHash; g.CostBoundMinor = costBound.MaxCostMinor

	ledger, ok := normalizedCanaryLedger(state.Ledger, grant, now)
	if !ok { g.Reason = "LEDGER_MISMATCH"; return g }
	g.ExecutionsTotalBefore = ledger.ExecutionsTotal; g.ExecutionsInWindowBefore = ledger.ExecutionsInWindow
	g.CostMinorTotalBefore = ledger.CostMinorTotal; g.PendingOutcomesBefore = ledger.PendingOutcomes
	if ledger.ReconciliationRequired { g.Decision = "WAIT"; g.Reason = "RECONCILIATION_REQUIRED"; return g }
	if containsFold(ledger.SuccessfulIdempotencyKeys, i.IdempotencyKey) { g.Decision = "WAIT"; g.Reason = "DUPLICATE_EXECUTION"; return g }
	if ledger.ExecutionsTotal >= grant.MaxExecutionsTotal { return requireApproval(&g, "CANARY_TOTAL_BUDGET_EXHAUSTED") }
	if ledger.ExecutionsInWindow >= grant.MaxExecutionsPerWindow { g.Decision = "WAIT"; g.Reason = "RATE_LIMIT_REACHED"; return g }
	if ledger.CostMinorTotal+costBound.MaxCostMinor > grant.MaxCostMinorTotal { return requireApproval(&g, "CANARY_COST_BUDGET_EXHAUSTED") }
	if ledger.PendingOutcomes >= grant.MaxPendingOutcomes { g.Decision = "WAIT"; g.Reason = "OUTCOME_BACKPRESSURE"; return g }
	g.Decision = "ALLOW_CANARY"; g.Reason = "CANARY_ELIGIBLE"; g.PerActionApprovalRequired = false
	return g
}

func AuthorizeCanary(state M10State, ctx M10Context) (CanaryExecutionAuthorization, CanaryGateDecision, string) {
	gate := EvaluateCanaryGate(state, ctx)
	if gate.Decision != "ALLOW_CANARY" { return CanaryExecutionAuthorization{}, gate, gate.Decision }
	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	intentExpires, errIntent := time.Parse(time.RFC3339, state.Intent.ExpiresAt)
	grantExpires, errGrant := time.Parse(time.RFC3339, state.Grant.ExpiresAt)
	costExpires, errCost := time.Parse(time.RFC3339, state.CostBound.ExpiresAt)
	if errNow != nil || errIntent != nil || errGrant != nil || errCost != nil || !intentExpires.After(now) || !grantExpires.After(now) || !costExpires.After(now) {
		gate.Decision = "DENY"; gate.Reason = "AUTHORIZATION_TIME_INVALID"; return CanaryExecutionAuthorization{}, gate, gate.Decision
	}
	expires := minTime(minTime(intentExpires, grantExpires), costExpires)
	auth := CanaryExecutionAuthorization{
		AuthorizationID: "canary-auth-" + state.Grant.GrantID + "-" + state.Intent.IntentID,
		IntentID: state.Intent.IntentID, IntentHash: state.Intent.IntentHash, PolicyVersion: state.Policy.PolicyVersion,
		CanaryGrantID: state.Grant.GrantID, CanaryGrantVersion: state.Grant.GrantVersion, CanaryGrantHash: state.Grant.GrantHash,
		CanaryGateID: gate.GateID, CanaryCostBoundID: state.CostBound.CostBoundID, CanaryCostBoundHash: state.CostBound.CostBoundHash,
		CanaryCostBoundMinor: state.CostBound.MaxCostMinor, ExecutorID: ctx.Executor.ExecutorID, AuthorizedAt: ctx.Now,
		ExpiresAt: expires.Format(time.RFC3339), IdempotencyKey: state.Intent.IdempotencyKey, CorrelationID: state.Intent.CorrelationID,
		ExecutionMode: "GOVERNED_CANARY", ExecutionAuthorized: true,
	}
	return auth, gate, "AUTHORIZED"
}

func canaryLedgerPath(dir, grantID, version string) string {
	sum := sha256.Sum256([]byte(grantID + "\x00" + version)); return filepath.Join(dir, "canary-ledger-"+hex.EncodeToString(sum[:])+".json")
}
func canaryLockPath(dir, grantID, version string) string {
	sum := sha256.Sum256([]byte("lock\x00" + grantID + "\x00" + version)); return filepath.Join(dir, "canary-lock-"+hex.EncodeToString(sum[:]))
}
func persistCanaryLedger(path string, ledger CanaryLedger) error {
	b, err := json.MarshalIndent(ledger, "", "  "); if err != nil { return err }
	tmp := path + ".tmp"; if err = os.WriteFile(tmp, b, 0600); err != nil { return err }; return os.Rename(tmp, path)
}
func loadExistingCanaryLedger(path string) (CanaryLedger, error) {
	var ledger CanaryLedger; b, err := os.ReadFile(path); if err != nil { return ledger, err }; err = json.Unmarshal(b, &ledger); return ledger, err
}
func ensureInitialCanaryLedger(path string, ledger CanaryLedger) error {
	b, err := json.MarshalIndent(ledger, "", "  "); if err != nil { return err }
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); if err != nil { return err }
	if _, err = f.Write(b); err != nil { _ = f.Close(); return err }; if err = f.Sync(); err != nil { _ = f.Close(); return err }; return f.Close()
}

func canaryAuthorizationMatches(got, expected CanaryExecutionAuthorization) bool { return got == expected }
func canaryExecutionRecord(state M10State, auth CanaryExecutionAuthorization, now, status, effectState, externalRef, message string) CanaryExecutionRecord {
	return CanaryExecutionRecord{
		ExecutionID: "exec-" + auth.AuthorizationID, AuthorizationID: auth.AuthorizationID,
		CanaryGrantID: auth.CanaryGrantID, CanaryGrantVersion: auth.CanaryGrantVersion, CanaryGrantHash: auth.CanaryGrantHash,
		CanaryGateID: auth.CanaryGateID, CanaryCostBoundID: auth.CanaryCostBoundID, CanaryCostBoundHash: auth.CanaryCostBoundHash,
		CanaryCostBoundMinor: auth.CanaryCostBoundMinor, IntentID: auth.IntentID, IntentHash: auth.IntentHash, ExecutorID: auth.ExecutorID,
		IdempotencyKey: auth.IdempotencyKey, AttemptedAt: now, Status: status, SideEffectState: effectState,
		ExternalRef: externalRef, Error: message, CorrelationID: state.Intent.CorrelationID,
	}
}

func ExecuteCanaryLocalSandbox(state *M10State, auth CanaryExecutionAuthorization, ctx M10Context, dir string) (CanaryExecutionRecord, string) {
	if state == nil || state.Authorization == nil || *state.Authorization != auth || !auth.ExecutionAuthorized || auth.ExecutionMode != "GOVERNED_CANARY" {
		return CanaryExecutionRecord{}, "DENY_AUTHORIZATION"
	}
	if ctx.KillSwitch { return CanaryExecutionRecord{}, "DENY_KILL_SWITCH" }
	if err := os.MkdirAll(dir, 0755); err != nil { return CanaryExecutionRecord{}, "EXECUTION_FAILED" }
	lockPath := canaryLockPath(dir, state.Grant.GrantID, state.Grant.GrantVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { if os.IsExist(err) { return CanaryExecutionRecord{}, "WAIT_CANARY_LOCK" }; return CanaryExecutionRecord{}, "EXECUTION_FAILED" }
	_ = lock.Close(); defer os.Remove(lockPath)

	ledgerPath := canaryLedgerPath(dir, state.Grant.GrantID, state.Grant.GrantVersion)
	ledger, err := loadExistingCanaryLedger(ledgerPath)
	if os.IsNotExist(err) {
		if state.Ledger.ExecutionsTotal != 0 || state.Ledger.ExecutionsInWindow != 0 || state.Ledger.CostMinorTotal != 0 || state.Ledger.PendingOutcomes != 0 ||
			len(state.Ledger.SuccessfulIdempotencyKeys) != 0 || len(state.Ledger.PendingExecutionIDs) != 0 || len(state.Ledger.OutcomeLinks) != 0 || state.Ledger.ReconciliationRequired {
			return CanaryExecutionRecord{}, "WAIT_LEDGER_MISSING"
		}
		if err = ensureInitialCanaryLedger(ledgerPath, state.Ledger); err != nil { return CanaryExecutionRecord{}, "WAIT_RECONCILIATION" }
		ledger = state.Ledger
	} else if err != nil { return CanaryExecutionRecord{}, "WAIT_RECONCILIATION" }
	state.Ledger = ledger
	expected, _, status := AuthorizeCanary(*state, ctx)
	if status != "AUTHORIZED" { return CanaryExecutionRecord{}, status }
	if !canaryAuthorizationMatches(auth, expected) { return CanaryExecutionRecord{}, "DENY_AUTHORIZATION" }
	now, errNow := time.Parse(time.RFC3339, ctx.Now); authorizedAt, errAuthorized := time.Parse(time.RFC3339, auth.AuthorizedAt); expiresAt, errExpires := time.Parse(time.RFC3339, auth.ExpiresAt)
	if errNow != nil || errAuthorized != nil || errExpires != nil || authorizedAt.After(now) || !expiresAt.After(now) { return CanaryExecutionRecord{}, "DENY_EXPIRED_AUTHORIZATION" }

	marker := sandboxIdempotencyPath(dir, auth.IdempotencyKey)
	if _, err = os.Stat(marker); err == nil {
		if containsFold(state.Ledger.SuccessfulIdempotencyKeys, auth.IdempotencyKey) { return CanaryExecutionRecord{}, "WAIT_ALREADY_EXECUTED" }
		state.Ledger.ReconciliationRequired = true; state.Ledger.UpdatedAt = ctx.Now; _ = persistCanaryLedger(ledgerPath, state.Ledger)
		rec := canaryExecutionRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, "idempotency marker exists without durable successful ledger entry")
		state.Execution = &rec; return rec, "WAIT_RECONCILIATION"
	} else if !os.IsNotExist(err) { return CanaryExecutionRecord{}, "WAIT_RECONCILIATION" }

	payload := map[string]any{
		"intent_id": state.Intent.IntentID, "intent_hash": state.Intent.IntentHash, "action_type": state.Intent.ActionType,
		"target": state.Intent.Target, "parameters": state.Intent.Parameters, "canary_grant_id": auth.CanaryGrantID,
		"canary_grant_hash": auth.CanaryGrantHash, "canary_gate_id": auth.CanaryGateID,
		"cost_bound_id": auth.CanaryCostBoundID, "cost_bound_hash": auth.CanaryCostBoundHash, "cost_bound_minor": auth.CanaryCostBoundMinor,
		"execution_mode": auth.ExecutionMode, "idempotency_key": auth.IdempotencyKey,
	}
	b, err := json.MarshalIndent(payload, "", "  "); if err != nil { return CanaryExecutionRecord{}, "EXECUTION_FAILED" }
	f, err := os.OpenFile(marker, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); if err != nil { return CanaryExecutionRecord{}, "WAIT_RECONCILIATION" }
	uncertain := func(message string) (CanaryExecutionRecord, string) {
		_ = f.Close(); state.Ledger.ReconciliationRequired = true; state.Ledger.UpdatedAt = ctx.Now; _ = persistCanaryLedger(ledgerPath, state.Ledger)
		rec := canaryExecutionRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, message); state.Execution = &rec; return rec, "WAIT_RECONCILIATION"
	}
	if _, err = f.Write(b); err != nil { return uncertain(err.Error()) }
	if err = f.Sync(); err != nil { return uncertain(err.Error()) }
	if err = f.Close(); err != nil {
		state.Ledger.ReconciliationRequired = true; state.Ledger.UpdatedAt = ctx.Now; _ = persistCanaryLedger(ledgerPath, state.Ledger)
		rec := canaryExecutionRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, err.Error()); state.Execution = &rec; return rec, "WAIT_RECONCILIATION"
	}

	normalized, ok := normalizedCanaryLedger(state.Ledger, state.Grant, now); if !ok { return uncertain("ledger became invalid before durable update") }
	normalized.ExecutionsTotal++; normalized.ExecutionsInWindow++; normalized.CostMinorTotal += auth.CanaryCostBoundMinor; normalized.PendingOutcomes++
	normalized.PendingExecutionIDs = append(normalized.PendingExecutionIDs, "exec-"+auth.AuthorizationID)
	normalized.SuccessfulIdempotencyKeys = append(normalized.SuccessfulIdempotencyKeys, auth.IdempotencyKey)
	normalized.LastExecutionAt = ctx.Now; normalized.UpdatedAt = ctx.Now
	if err = persistCanaryLedger(ledgerPath, normalized); err != nil {
		state.Ledger.ReconciliationRequired = true; state.Ledger.UpdatedAt = ctx.Now
		rec := canaryExecutionRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", marker, "side effect written but ledger persistence failed"); state.Execution = &rec; return rec, "WAIT_RECONCILIATION"
	}
	state.Ledger = normalized
	rec := canaryExecutionRecord(*state, auth, ctx.Now, "SUCCEEDED", "PERFORMED", marker, ""); state.Execution = &rec; return rec, "EXECUTED"
}

func RecordCanaryOutcome(state *M10State, dir, outcomeID, executionID, observedAt string) string {
	if state == nil || strings.TrimSpace(outcomeID) == "" || strings.TrimSpace(executionID) == "" { return "DENY_OUTCOME" }
	if err := os.MkdirAll(dir, 0755); err != nil { return "WAIT_RECONCILIATION" }
	lockPath := canaryLockPath(dir, state.Grant.GrantID, state.Grant.GrantVersion)
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600); if err != nil { return "WAIT_CANARY_LOCK" }
	_ = lock.Close(); defer os.Remove(lockPath)
	ledgerPath := canaryLedgerPath(dir, state.Grant.GrantID, state.Grant.GrantVersion)
	ledger, err := loadExistingCanaryLedger(ledgerPath); if err != nil || ledger.ReconciliationRequired || ledger.PendingOutcomes < 1 { return "WAIT_RECONCILIATION" }
	if ledger.GrantHash != state.Grant.GrantHash { return "WAIT_RECONCILIATION" }
	for _, link := range ledger.OutcomeLinks { if link.OutcomeID == outcomeID { return "DUPLICATE_OUTCOME" } }
	idx := -1; for i, id := range ledger.PendingExecutionIDs { if id == executionID { idx = i; break } }
	if idx < 0 { return "DENY_OUTCOME_LINK" }
	if _, err := time.Parse(time.RFC3339, observedAt); err != nil { return "DENY_OUTCOME" }
	ledger.PendingExecutionIDs = append(ledger.PendingExecutionIDs[:idx], ledger.PendingExecutionIDs[idx+1:]...)
	ledger.PendingOutcomes--; ledger.OutcomeLinks = append(ledger.OutcomeLinks, CanaryOutcomeLink{OutcomeID: outcomeID, ExecutionID: executionID, ObservedAt: observedAt}); ledger.UpdatedAt = observedAt
	if err = persistCanaryLedger(ledgerPath, ledger); err != nil { return "WAIT_RECONCILIATION" }
	state.Ledger = ledger; return "OUTCOME_RECORDED"
}

func baseDemoM10() (M10State, M10Context) {
	intent := SealShadowActionIntent(ShadowActionIntent{IntentID:"i10",DecisionID:"d10",EvidenceIDs:[]string{"e10"},ActionType:"PREPARE_LOCAL_DRAFT",Target:"https://example.com/draft",Parameters:map[string]any{"title":"canary"},ProposedBy:"human",CreatedAt:"2026-09-03T07:00:00Z",ExpiresAt:"2026-09-03T11:00:00Z",CorrelationID:"c10",IdempotencyKey:"k10"})
	policyCtx := ShadowPolicyContext{PolicyVersion:"m10-v1",Now:"2026-09-03T07:05:00Z",KnownDecisionIDs:[]string{"d10"},KnownEvidenceIDs:[]string{"e10"},AllowedHosts:[]string{"example.com"},ActionRisk:map[string]string{"PREPARE_LOCAL_DRAFT":"RISK0","UPDATE_DRAFT":"RISK1","PUBLISH_CONTENT":"RISK2"},SeenIdempotency:map[string]string{}}
	policy := EvaluateShadowPolicy(intent, policyCtx)
	grant := SealCanaryGrant(CanaryGrant{GrantID:"g10",GrantVersion:"v1",PolicyVersion:"m10-v1",ApprovalRef:"human-grant-approval-10",ApprovedBy:"human",ApproverID:"learner",ApprovedAt:"2026-09-03T07:10:00Z",ValidFrom:"2026-09-03T07:15:00Z",ExpiresAt:"2026-09-03T10:00:00Z",AllowedRiskClasses:[]string{"RISK0"},AllowedActionTypes:[]string{"PREPARE_LOCAL_DRAFT"},AllowedHosts:[]string{"example.com"},ExecutorIDs:[]string{"local_sandbox"},MaxExecutionsTotal:1,MaxExecutionsPerWindow:1,WindowSeconds:3600,MaxCostMinorTotal:0,Currency:"USD",MaxPendingOutcomes:1,KillSwitchRequired:true,CorrelationID:"grant-c10"})
	cost := SealCanaryCostBound(CanaryCostBound{CostBoundID:"cost10",IntentID:intent.IntentID,IntentHash:intent.IntentHash,MaxCostMinor:0,Currency:"USD",SourceRef:"deterministic-cost-registry",ObservedAt:"2026-09-03T07:20:00Z",ExpiresAt:"2026-09-03T09:00:00Z",CorrelationID:intent.CorrelationID})
	ledger := CanaryLedger{GrantID:grant.GrantID,GrantVersion:grant.GrantVersion,GrantHash:grant.GrantHash,WindowStartedAt:"2026-09-03T07:15:00Z",UpdatedAt:"2026-09-03T07:15:00Z"}
	state := M10State{Intent:intent,Policy:policy,Grant:grant,CostBound:cost,Ledger:ledger}
	ctx := M10Context{Now:"2026-09-03T08:00:00Z",ApprovedGrantHashes:map[string]string{grant.ApprovalRef:grant.GrantHash},TrustedCostBounds:map[string]string{cost.CostBoundID:cost.CostBoundHash},Executor:ExecutorProfile{ExecutorID:"local_sandbox",AllowedActionTypes:[]string{"PREPARE_LOCAL_DRAFT"},AllowedHosts:[]string{"example.com"}},AllowedExecutorIDs:[]string{"local_sandbox"},PolicyContext:policyCtx}
	return state, ctx
}

func demoM10() (map[string]any, error) {
	state, ctx := baseDemoM10(); auth, gate, status := AuthorizeCanary(state, ctx); if status != "AUTHORIZED" { return nil, errors.New(status) }
	state.Authorization = &auth; dir, err := os.MkdirTemp("", "m10-canary-"); if err != nil { return nil, err }
	rec, execStatus := ExecuteCanaryLocalSandbox(&state, auth, ctx, dir); if execStatus != "EXECUTED" { return nil, errors.New(execStatus) }
	return map[string]any{"gate":gate,"authorization":auth,"execution":rec,"ledger":state.Ledger,"authority":"GOVERNED_CANARY"}, nil
}
