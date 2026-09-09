package m10

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

// CanaryGrant is the immutable delegation artifact used before any canary
// budget may be reserved. It is deliberately separate from mutable ledger
// counters owned by the learner runtime.
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

func sortedGrantValues(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func ComputeCanaryGrantHash(g CanaryGrant) string {
	raw, err := json.Marshal(canaryGrantHashPayload{
		GrantID: g.GrantID, GrantVersion: g.GrantVersion, PolicyVersion: g.PolicyVersion,
		ApprovalRef: g.ApprovalRef, ApprovedBy: g.ApprovedBy, ApproverID: g.ApproverID,
		ApprovedAt: g.ApprovedAt, ValidFrom: g.ValidFrom, ExpiresAt: g.ExpiresAt,
		AllowedRiskClasses: sortedGrantValues(g.AllowedRiskClasses), AllowedActionTypes: sortedGrantValues(g.AllowedActionTypes),
		AllowedHosts: sortedGrantValues(g.AllowedHosts), ExecutorIDs: sortedGrantValues(g.ExecutorIDs),
		MaxExecutionsTotal: g.MaxExecutionsTotal, MaxExecutionsPerWindow: g.MaxExecutionsPerWindow,
		WindowSeconds: g.WindowSeconds, MaxCostMinorTotal: g.MaxCostMinorTotal, Currency: strings.ToUpper(strings.TrimSpace(g.Currency)),
		MaxPendingOutcomes: g.MaxPendingOutcomes, KillSwitchRequired: g.KillSwitchRequired,
		CorrelationID: g.CorrelationID, HashVersion: g.HashVersion,
	})
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func DecodeCanaryGrant(raw []byte) (CanaryGrant, string) {
	var grant CanaryGrant
	if contracts.ValidateRaw("canary-grant.schema.json", raw) != nil || contracts.DecodeStrict(raw, &grant) != nil {
		return grant, "INVALID_SCHEMA"
	}
	if grant.HashVersion != "go-json-v1" || grant.GrantHash != ComputeCanaryGrantHash(grant) {
		return grant, "TAMPERED_GRANT"
	}
	approved, approvedErr := time.Parse(time.RFC3339, grant.ApprovedAt)
	validFrom, fromErr := time.Parse(time.RFC3339, grant.ValidFrom)
	expires, expiresErr := time.Parse(time.RFC3339, grant.ExpiresAt)
	if approvedErr != nil || fromErr != nil || expiresErr != nil || approved.After(validFrom) || !expires.After(validFrom) || grant.MaxExecutionsPerWindow > grant.MaxExecutionsTotal {
		return grant, "INVALID_TIME_OR_LIMITS"
	}
	return grant, "VALID"
}

func containsGrantValue(values []string, want string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(want)) {
			return true
		}
	}
	return false
}

// ValidCanaryGrantFor rechecks every binding at the point a mutable learner
// state accepts a delegation. It does not authorize execution.
func ValidCanaryGrantFor(g CanaryGrant, intentID, intentHash, policyVersion, approvalID, approverID, correlationID, riskClass, actionType, target string, now time.Time) string {
	if g.IntentionalPlaceholder() {
		return "INVALID_GRANT"
	}
	if g.PolicyVersion != policyVersion || g.ApprovalRef != approvalID || g.ApproverID != approverID || g.CorrelationID != correlationID || !containsGrantValue(g.AllowedRiskClasses, riskClass) || !containsGrantValue(g.AllowedActionTypes, actionType) {
		return "GRANT_BINDING_MISMATCH"
	}
	parsed, err := url.Parse(target)
	if err != nil || parsed.Scheme != "https" || !containsGrantValue(g.AllowedHosts, parsed.Hostname()) {
		return "GRANT_SCOPE_MISMATCH"
	}
	validFrom, fromErr := time.Parse(time.RFC3339, g.ValidFrom)
	expires, expiresErr := time.Parse(time.RFC3339, g.ExpiresAt)
	if fromErr != nil || expiresErr != nil || now.Before(validFrom) || !expires.After(now) {
		return "GRANT_EXPIRED_OR_INACTIVE"
	}
	// The canonical grant schema has no intent fields. The remaining intent
	// binding is carried by its approval/policy/correlation linkage above; keep
	// these parameters explicit so callers cannot accidentally omit the check.
	if strings.TrimSpace(intentID) == "" || strings.TrimSpace(intentHash) == "" {
		return "GRANT_BINDING_MISMATCH"
	}
	return "VALID"
}

func (g CanaryGrant) IntentionalPlaceholder() bool {
	return g.ApprovedBy != "human" || !g.KillSwitchRequired || strings.TrimSpace(g.GrantID) == "" || strings.TrimSpace(g.GrantVersion) == "" || strings.TrimSpace(g.Currency) == ""
}
