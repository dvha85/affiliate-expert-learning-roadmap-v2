package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"
)

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
}

type intentHashPayload struct {
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
	IntentMode          string         `json:"intent_mode"`
	ExecutionAuthorized bool           `json:"execution_authorized"`
}

func ComputeShadowIntentHash(i ShadowActionIntent) string {
	evidence := append([]string(nil), i.EvidenceIDs...)
	sort.Strings(evidence)
	payload := intentHashPayload{
		IntentID: i.IntentID, DecisionID: i.DecisionID, EvidenceIDs: evidence,
		ActionType: strings.ToUpper(strings.TrimSpace(i.ActionType)), Target: i.Target,
		Parameters: i.Parameters, ProposedBy: i.ProposedBy, ProposalRef: i.ProposalRef,
		CreatedAt: i.CreatedAt, ExpiresAt: i.ExpiresAt, CorrelationID: i.CorrelationID,
		IdempotencyKey: i.IdempotencyKey, IntentMode: i.IntentMode, ExecutionAuthorized: i.ExecutionAuthorized,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func SealShadowActionIntent(i ShadowActionIntent) ShadowActionIntent {
	i.IntentMode = "PROPOSAL_ONLY"
	i.ExecutionAuthorized = false
	i.ActionType = strings.ToUpper(strings.TrimSpace(i.ActionType))
	i.IntentHash = ComputeShadowIntentHash(i)
	return i
}

func stringSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[strings.TrimSpace(value)] = true
	}
	return out
}

func allowedHost(target string, hosts []string) bool {
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return false
	}
	for _, host := range hosts {
		if strings.EqualFold(strings.TrimSpace(host), u.Hostname()) {
			return true
		}
	}
	return false
}

func EvaluateShadowPolicy(i ShadowActionIntent, ctx ShadowPolicyContext) ShadowPolicyDecision {
	decision := ShadowPolicyDecision{
		PolicyVersion: ctx.PolicyVersion, IntentID: i.IntentID, IntentHash: i.IntentHash,
		Decision: "DENY", RiskClass: "RISK2", Reason: "INVALID_INTENT",
		PolicyReviewRequired: false, PolicyMode: "NON_AUTHORIZING", ExecutionAuthorized: false, PolicyCheckedAt: ctx.Now,
	}

	if strings.TrimSpace(ctx.PolicyVersion) == "" {
		decision.Reason = "POLICY_UNAVAILABLE"
		return decision
	}
	if strings.TrimSpace(i.IntentID) == "" || strings.TrimSpace(i.DecisionID) == "" ||
		strings.TrimSpace(i.ActionType) == "" || strings.TrimSpace(i.Target) == "" ||
		strings.TrimSpace(i.CorrelationID) == "" || strings.TrimSpace(i.IdempotencyKey) == "" ||
		(i.ProposedBy != "human" && i.ProposedBy != "agent") {
		return decision
	}
	if i.IntentMode != "PROPOSAL_ONLY" || i.ExecutionAuthorized {
		decision.Reason = "INTENT_AUTHORITY_FORBIDDEN"
		return decision
	}
	if i.IntentHash == "" || i.IntentHash != ComputeShadowIntentHash(i) {
		decision.Reason = "TAMPERED_INTENT"
		return decision
	}

	risk, ok := ctx.ActionRisk[i.ActionType]
	if !ok || (risk != "RISK0" && risk != "RISK1" && risk != "RISK2") {
		decision.Reason = "UNKNOWN_ACTION_POLICY"
		return decision
	}
	decision.RiskClass = risk

	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	created, errCreated := time.Parse(time.RFC3339, i.CreatedAt)
	expires, errExpires := time.Parse(time.RFC3339, i.ExpiresAt)
	if errNow != nil || errCreated != nil || errExpires != nil || !expires.After(created) {
		decision.Reason = "INVALID_TIME_BINDING"
		return decision
	}
	if created.After(now) {
		decision.Decision = "WAIT"
		decision.Reason = "INTENT_NOT_YET_VALID"
		return decision
	}
	if !expires.After(now) {
		decision.Reason = "EXPIRED_INTENT"
		return decision
	}

	if !stringSet(ctx.KnownDecisionIDs)[i.DecisionID] {
		decision.Decision = "GET_MORE_DATA"
		decision.Reason = "MISSING_DECISION_LINK"
		return decision
	}
	knownEvidence := stringSet(ctx.KnownEvidenceIDs)
	if len(i.EvidenceIDs) == 0 {
		decision.Decision = "GET_MORE_DATA"
		decision.Reason = "MISSING_EVIDENCE_LINK"
		return decision
	}
	for _, id := range i.EvidenceIDs {
		if !knownEvidence[id] {
			decision.Decision = "GET_MORE_DATA"
			decision.Reason = "MISSING_EVIDENCE_LINK"
			return decision
		}
	}
	if i.ProposedBy == "agent" {
		if strings.TrimSpace(i.ProposalRef) == "" || !stringSet(ctx.KnownProposalIDs)[i.ProposalRef] {
			decision.Decision = "GET_MORE_DATA"
			decision.Reason = "MISSING_PROPOSAL_LINK"
			return decision
		}
	}
	if !allowedHost(i.Target, ctx.AllowedHosts) {
		decision.Reason = "TARGET_NOT_ALLOWED"
		return decision
	}

	if previousHash, seen := ctx.SeenIdempotency[i.IdempotencyKey]; seen {
		if previousHash == i.IntentHash {
			decision.Decision = "WAIT"
			decision.Reason = "DUPLICATE_INTENT"
			return decision
		}
		decision.Reason = "IDEMPOTENCY_COLLISION"
		return decision
	}

	switch risk {
	case "RISK0":
		decision.Decision = "ALLOW"
		decision.Reason = "SHADOW_POLICY_ALLOW"
	case "RISK1":
		decision.Decision = "HUMAN_REVIEW"
		decision.Reason = "RISK1_REQUIRES_REVIEW"
		decision.PolicyReviewRequired = true
	case "RISK2":
		decision.Decision = "HUMAN_REVIEW"
		decision.Reason = "RISK2_REQUIRES_REVIEW"
		decision.PolicyReviewRequired = true
	}
	return decision
}

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
		ActionRisk: map[string]string{"PREPARE_LOCAL_DRAFT": "RISK0", "UPDATE_DRAFT": "RISK1", "PUBLISH_CONTENT": "RISK2"},
		SeenIdempotency: map[string]string{},
	})
}
