package main

import corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"

type ShadowActionIntent struct {
	IntentID            string         `json:"intent_id"`
	DecisionID          string         `json:"decision_id"`
	EvidenceIDs         []string       `json:"evidence_ids"`
	ActionType          string         `json:"action_type"`
	Target              string         `json:"target"`
	Parameters          map[string]any `json:"parameters"`
	ProposedBy          string         `json:"proposed_by"`
	ProposalRef         string         `json:"proposal_ref,omitempty"`
	CreatedAt           string         `json:"created_at"`
	ExpiresAt           string         `json:"expires_at"`
	CorrelationID       string         `json:"correlation_id"`
	IdempotencyKey      string         `json:"idempotency_key"`
	IntentHash          string         `json:"intent_hash"`
	IntentMode          string         `json:"intent_mode"`
	ExecutionAuthorized bool           `json:"execution_authorized"`
	// Internal compatibility aliases for M09-M11 runtime code. They are not canonical JSON fields.
	ShadowOnly bool `json:"-"`
	DryRun     bool `json:"-"`
}

type ShadowPolicyContext struct {
	PolicyVersion    string            `json:"policy_version"`
	Now              string            `json:"now"`
	KnownDecisionIDs []string          `json:"known_decision_ids"`
	KnownEvidenceIDs []string          `json:"known_evidence_ids"`
	KnownProposalIDs []string          `json:"known_proposal_ids"`
	AllowedHosts     []string          `json:"allowed_hosts"`
	ActionRisk       map[string]string `json:"action_risk"`
	SeenIdempotency  map[string]string `json:"seen_idempotency"`
}

type ShadowPolicyDecision struct {
	PolicyVersion        string `json:"policy_version"`
	IntentID             string `json:"intent_id"`
	IntentHash           string `json:"intent_hash"`
	Decision             string `json:"decision"`
	RiskClass            string `json:"risk_class"`
	Reason               string `json:"reason"`
	PolicyReviewRequired bool   `json:"policy_review_required"`
	PolicyMode           string `json:"policy_mode"`
	ExecutionAuthorized  bool   `json:"execution_authorized"`
	PolicyCheckedAt      string `json:"policy_checked_at"`
	// Internal compatibility aliases for M09-M11 runtime code. They are not canonical JSON fields.
	ApprovalRequired bool `json:"-"`
	ShadowOnly       bool `json:"-"`
}

func coreIntent(i ShadowActionIntent) corem08.Intent {
	return corem08.Intent{IntentID: i.IntentID, DecisionID: i.DecisionID, EvidenceIDs: i.EvidenceIDs, ActionType: i.ActionType, Target: i.Target, Parameters: i.Parameters, ProposedBy: i.ProposedBy, ProposalRef: i.ProposalRef, CreatedAt: i.CreatedAt, ExpiresAt: i.ExpiresAt, CorrelationID: i.CorrelationID, IdempotencyKey: i.IdempotencyKey, IntentHash: i.IntentHash, IntentMode: i.IntentMode, ExecutionAuthorized: i.ExecutionAuthorized}
}

func ComputeShadowIntentHash(i ShadowActionIntent) string {
	return corem08.ComputeIntentHash(coreIntent(i))
}

func SealShadowActionIntent(i ShadowActionIntent) ShadowActionIntent {
	sealed := corem08.SealIntent(coreIntent(i))
	i.IntentMode, i.ExecutionAuthorized, i.ActionType, i.IntentHash = sealed.IntentMode, sealed.ExecutionAuthorized, sealed.ActionType, sealed.IntentHash
	i.ShadowOnly, i.DryRun = true, true
	return i
}

func EvaluateShadowPolicy(i ShadowActionIntent, ctx ShadowPolicyContext) ShadowPolicyDecision {
	p := corem08.EvaluatePolicy(coreIntent(i), corem08.PolicyContext(ctx))
	return ShadowPolicyDecision{PolicyVersion: p.PolicyVersion, IntentID: p.IntentID, IntentHash: p.IntentHash, Decision: p.Decision, RiskClass: p.RiskClass, Reason: p.Reason, PolicyReviewRequired: p.PolicyReviewRequired, PolicyMode: p.PolicyMode, ExecutionAuthorized: p.ExecutionAuthorized, PolicyCheckedAt: p.PolicyCheckedAt, ApprovalRequired: p.PolicyReviewRequired, ShadowOnly: true}
}

func allowedHost(target string, hosts []string) bool { return corem08.AllowedHost(target, hosts) }

func demoM08Decision() ShadowPolicyDecision {
	intent := SealShadowActionIntent(ShadowActionIntent{
		IntentID: "intent-1", DecisionID: "decision-1", EvidenceIDs: []string{"e1"},
		ActionType: "PREPARE_LOCAL_DRAFT", Target: "https://example.com/draft",
		Parameters: map[string]any{"title": "draft only"}, ProposedBy: "agent", ProposalRef: "proposal-1",
		CreatedAt: "2026-09-03T01:00:00Z", ExpiresAt: "2026-09-03T03:00:00Z",
		CorrelationID: "corr-1", IdempotencyKey: "idem-1",
	})
	return EvaluateShadowPolicy(intent, ShadowPolicyContext{
		PolicyVersion: "m08-v1", Now: "2026-09-03T02:00:00Z",
		KnownDecisionIDs: []string{"decision-1"}, KnownEvidenceIDs: []string{"e1"}, KnownProposalIDs: []string{"proposal-1"}, AllowedHosts: []string{"example.com"},
		ActionRisk:      map[string]string{"PREPARE_LOCAL_DRAFT": "RISK0", "UPDATE_DRAFT": "RISK1", "PUBLISH_CONTENT": "RISK2"},
		SeenIdempotency: map[string]string{},
	})
}
