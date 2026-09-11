// Package m08 provides the canonical proposal-only intent hash and policy
// evaluation shared by the learner CLI and the mission runtime harness.
package m08

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

type Intent struct {
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

type PolicyContext struct {
	PolicyVersion    string            `json:"policy_version"`
	Now              string            `json:"now"`
	KnownDecisionIDs []string          `json:"known_decision_ids"`
	KnownEvidenceIDs []string          `json:"known_evidence_ids"`
	KnownProposalIDs []string          `json:"known_proposal_ids"`
	AllowedHosts     []string          `json:"allowed_hosts"`
	ActionRisk       map[string]string `json:"action_risk"`
	SeenIdempotency  map[string]string `json:"seen_idempotency"`
}

type PolicyDecision struct {
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

// DecodeIntent validates original bytes before typed decoding, so duplicate,
// unknown, case-variant fields and null required fields cannot be hidden by a
// Go struct conversion. Parameters are restored from contracts.Decode to keep
// json.Number values exact for hashing.
func DecodeIntent(raw []byte) (Intent, string) {
	var i Intent
	if contracts.ValidateRaw("action-intent.schema.json", raw) != nil || contracts.DecodeStrict(raw, &i) != nil {
		return i, "INVALID_SCHEMA"
	}
	v, err := contracts.Decode(raw)
	if err != nil {
		return i, "INVALID_SCHEMA"
	}
	params, ok := v.(map[string]any)["parameters"].(map[string]any)
	if !ok {
		return i, "INVALID_SCHEMA"
	}
	i.Parameters = params
	return i, "VALID"
}

func DecodePolicy(raw []byte) (PolicyDecision, string) {
	var p PolicyDecision
	if contracts.ValidateRaw("policy-decision.schema.json", raw) != nil || contracts.DecodeStrict(raw, &p) != nil {
		return p, "INVALID_SCHEMA"
	}
	return p, "VALID"
}

// ValidatePolicyForIntent checks the semantic invariants which are available
// from the two immutable artifacts alone. It intentionally does not replace
// EvaluatePolicy: a caller that still owns its PolicyContext must re-evaluate
// it before granting any authority. This boundary prevents a stored policy
// from using a decision/risk combination that the canonical evaluator could
// never emit for a successful proposal.
func ValidatePolicyForIntent(i Intent, p PolicyDecision) string {
	if i.IntentHash == "" || i.IntentHash != ComputeIntentHash(i) || i.IntentMode != "PROPOSAL_ONLY" || i.ExecutionAuthorized {
		return "INVALID_INTENT"
	}
	if p.PolicyVersion == "" || p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || p.PolicyMode != "NON_AUTHORIZING" || p.ExecutionAuthorized {
		return "INVALID_POLICY"
	}
	created, createdErr := time.Parse(time.RFC3339, i.CreatedAt)
	expires, expiryErr := time.Parse(time.RFC3339, i.ExpiresAt)
	checked, checkedErr := time.Parse(time.RFC3339, p.PolicyCheckedAt)
	if createdErr != nil || expiryErr != nil || checkedErr != nil || !expires.After(created) || checked.Before(created) || !checked.Before(expires) {
		return "INVALID_TIME_BINDING"
	}
	if p.RiskClass != "RISK0" && p.RiskClass != "RISK1" && p.RiskClass != "RISK2" {
		return "INVALID_POLICY"
	}
	switch p.Decision {
	case "ALLOW":
		if p.RiskClass != "RISK0" || p.PolicyReviewRequired {
			return "INVALID_POLICY"
		}
	case "HUMAN_REVIEW":
		if (p.RiskClass != "RISK1" && p.RiskClass != "RISK2") || !p.PolicyReviewRequired {
			return "INVALID_POLICY"
		}
	case "DENY", "WAIT", "GET_MORE_DATA":
		if p.PolicyReviewRequired {
			return "INVALID_POLICY"
		}
	default:
		return "INVALID_POLICY"
	}
	return "VALID"
}

type hashPayload struct {
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

func ComputeIntentHash(i Intent) string {
	ids := append([]string(nil), i.EvidenceIDs...)
	sort.Strings(ids)
	b, err := json.Marshal(hashPayload{i.IntentID, i.DecisionID, ids, strings.ToUpper(strings.TrimSpace(i.ActionType)), i.Target, i.Parameters, i.ProposedBy, i.ProposalRef, i.CreatedAt, i.ExpiresAt, i.CorrelationID, i.IdempotencyKey, i.IntentMode, i.ExecutionAuthorized})
	if err != nil {
		return ""
	}
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}

func SealIntent(i Intent) Intent {
	i.IntentMode = "PROPOSAL_ONLY"
	i.ExecutionAuthorized = false
	i.ActionType = strings.ToUpper(strings.TrimSpace(i.ActionType))
	i.IntentHash = ComputeIntentHash(i)
	return i
}

func set(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[strings.TrimSpace(value)] = true
	}
	return out
}

func AllowedHost(target string, hosts []string) bool {
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.Port() != "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	for _, host := range hosts {
		if strings.EqualFold(strings.TrimSpace(host), u.Hostname()) {
			return true
		}
	}
	return false
}

// EvaluatePolicy is non-authorizing. It returns a closed state when any link
// or dependency is absent instead of allowing the request optimistically.
func EvaluatePolicy(i Intent, ctx PolicyContext) PolicyDecision {
	p := PolicyDecision{PolicyVersion: ctx.PolicyVersion, IntentID: i.IntentID, IntentHash: i.IntentHash, Decision: "DENY", RiskClass: "RISK2", Reason: "INVALID_INTENT", PolicyMode: "NON_AUTHORIZING", PolicyCheckedAt: ctx.Now}
	if strings.TrimSpace(ctx.PolicyVersion) == "" {
		p.Reason = "POLICY_UNAVAILABLE"
		return p
	}
	if strings.TrimSpace(i.IntentID) == "" || strings.TrimSpace(i.DecisionID) == "" || strings.TrimSpace(i.ActionType) == "" || strings.TrimSpace(i.Target) == "" || strings.TrimSpace(i.CorrelationID) == "" || strings.TrimSpace(i.IdempotencyKey) == "" || i.Parameters == nil || (i.ProposedBy != "human" && i.ProposedBy != "agent") {
		return p
	}
	if i.IntentMode != "PROPOSAL_ONLY" || i.ExecutionAuthorized {
		p.Reason = "INTENT_AUTHORITY_FORBIDDEN"
		return p
	}
	if i.IntentHash == "" || i.IntentHash != ComputeIntentHash(i) {
		p.Reason = "TAMPERED_INTENT"
		return p
	}
	risk, ok := ctx.ActionRisk[i.ActionType]
	if !ok || (risk != "RISK0" && risk != "RISK1" && risk != "RISK2") {
		p.Reason = "UNKNOWN_ACTION_POLICY"
		return p
	}
	p.RiskClass = risk
	now, nowErr := time.Parse(time.RFC3339, ctx.Now)
	created, createdErr := time.Parse(time.RFC3339, i.CreatedAt)
	expires, expiresErr := time.Parse(time.RFC3339, i.ExpiresAt)
	if nowErr != nil || createdErr != nil || expiresErr != nil || !expires.After(created) {
		p.Reason = "INVALID_TIME_BINDING"
		return p
	}
	if created.After(now) {
		p.Decision = "WAIT"
		p.Reason = "INTENT_NOT_YET_VALID"
		return p
	}
	if !expires.After(now) {
		p.Reason = "EXPIRED_INTENT"
		return p
	}
	if !set(ctx.KnownDecisionIDs)[i.DecisionID] {
		p.Decision = "GET_MORE_DATA"
		p.Reason = "MISSING_DECISION_LINK"
		return p
	}
	knownEvidence := set(ctx.KnownEvidenceIDs)
	if len(i.EvidenceIDs) == 0 {
		p.Decision = "GET_MORE_DATA"
		p.Reason = "MISSING_EVIDENCE_LINK"
		return p
	}
	for _, id := range i.EvidenceIDs {
		if !knownEvidence[id] {
			p.Decision = "GET_MORE_DATA"
			p.Reason = "MISSING_EVIDENCE_LINK"
			return p
		}
	}
	if i.ProposedBy == "agent" && (strings.TrimSpace(i.ProposalRef) == "" || !set(ctx.KnownProposalIDs)[i.ProposalRef]) {
		p.Decision = "GET_MORE_DATA"
		p.Reason = "MISSING_PROPOSAL_LINK"
		return p
	}
	if !AllowedHost(i.Target, ctx.AllowedHosts) {
		p.Reason = "TARGET_NOT_ALLOWED"
		return p
	}
	if previous, seen := ctx.SeenIdempotency[i.IdempotencyKey]; seen {
		if previous == i.IntentHash {
			p.Decision = "WAIT"
			p.Reason = "DUPLICATE_INTENT"
		} else {
			p.Reason = "IDEMPOTENCY_COLLISION"
		}
		return p
	}
	switch risk {
	case "RISK0":
		p.Decision, p.Reason = "ALLOW", "SHADOW_POLICY_ALLOW"
	case "RISK1":
		p.Decision, p.Reason, p.PolicyReviewRequired = "HUMAN_REVIEW", "RISK1_REQUIRES_REVIEW", true
	case "RISK2":
		p.Decision, p.Reason, p.PolicyReviewRequired = "HUMAN_REVIEW", "RISK2_REQUIRES_REVIEW", true
	}
	return p
}
