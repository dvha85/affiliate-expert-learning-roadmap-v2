// Package m09 provides the strict human-approval boundary shared by the
// learner runtime and the mission harness. It is intentionally
// non-authorizing: callers still own executor, grant and ledger decisions.
package m09

import (
	"fmt"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
)

const (
	Valid                = "VALID"
	InvalidSchema        = "INVALID_SCHEMA"
	InvalidIntent        = "INVALID_INTENT"
	InvalidPolicy        = "INVALID_POLICY_STATE"
	ApprovalMismatch     = "APPROVAL_MISMATCH"
	RejectedApproval     = "REJECTED_APPROVAL"
	InvalidApprover      = "INVALID_APPROVER"
	InvalidTimeBinding   = "INVALID_TIME_BINDING"
	ExpiredIntent        = "EXPIRED_INTENT"
	ExpiredApproval      = "EXPIRED_APPROVAL"
	ApprovalBeforePolicy = "APPROVAL_BEFORE_POLICY"
	InvalidProfile       = "INVALID_PROFILE"
)

// ApprovalRecord is the canonical, human-only, one-time M09 approval.
type ApprovalRecord struct {
	ApprovalID    string `json:"approval_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	PolicyVersion string `json:"policy_version"`
	Decision      string `json:"decision"`
	ApprovedBy    string `json:"approved_by"`
	ApproverID    string `json:"approver_id"`
	ApprovedAt    string `json:"approved_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	OneTime       bool   `json:"one_time"`
}

// DecodeApproval validates raw bytes before typed decoding. In particular it
// rejects duplicate or unknown fields instead of letting encoding/json choose
// a last duplicate value.
func DecodeApproval(raw []byte) (ApprovalRecord, string) {
	var approval ApprovalRecord
	if contracts.ValidateRaw("approval-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &approval) != nil {
		return ApprovalRecord{}, InvalidSchema
	}
	return approval, Valid
}

// ExecutionAuthorization is the M09 human-approval capability shape. M10 and
// M11 have separate governed capability types; keeping this type here makes
// the approved-live decoder explicit instead of silently accepting another
// execution profile at the M09 boundary.
type ExecutionAuthorization struct {
	AuthorizationID     string `json:"authorization_id"`
	IntentID            string `json:"intent_id"`
	IntentHash          string `json:"intent_hash"`
	PolicyVersion       string `json:"policy_version"`
	ApprovalID          string `json:"approval_id"`
	ExecutorID          string `json:"executor_id"`
	AuthorizedAt        string `json:"authorized_at"`
	ExpiresAt           string `json:"expires_at"`
	IdempotencyKey      string `json:"idempotency_key"`
	CorrelationID       string `json:"correlation_id"`
	ExecutionMode       string `json:"execution_mode"`
	ExecutionAuthorized bool   `json:"execution_authorized"`
}

// ExecutionRecord is the M09 audit record shape. It is decoded strictly but
// cross-artifact chain binding remains the responsibility of the caller.
type ExecutionRecord struct {
	ExecutionID     string `json:"execution_id"`
	AuthorizationID string `json:"authorization_id"`
	ApprovalID      string `json:"approval_id"`
	IntentID        string `json:"intent_id"`
	IntentHash      string `json:"intent_hash"`
	ExecutorID      string `json:"executor_id"`
	IdempotencyKey  string `json:"idempotency_key"`
	AttemptedAt     string `json:"attempted_at"`
	Status          string `json:"status"`
	SideEffectState string `json:"side_effect_state"`
	ExternalRef     string `json:"external_ref,omitempty"`
	Error           string `json:"error,omitempty"`
	CorrelationID   string `json:"correlation_id"`
}

// DecodeAuthorization validates an M09 capability before typed decoding. A
// schema-valid canary or production capability is not an M09 approval result.
func DecodeAuthorization(raw []byte) (ExecutionAuthorization, string) {
	var authorization ExecutionAuthorization
	if contracts.ValidateRaw("execution-authorization.schema.json", raw) != nil || contracts.DecodeStrict(raw, &authorization) != nil {
		return ExecutionAuthorization{}, InvalidSchema
	}
	if authorization.ExecutionMode != "APPROVED_LIVE" || !authorization.ExecutionAuthorized {
		return ExecutionAuthorization{}, InvalidProfile
	}
	return authorization, Valid
}

// DecodeExecution validates an M09 execution record before typed decoding.
// Cross-artifact links and time windows are checked by the caller's chain
// validator because the record alone does not contain the referenced intent.
func DecodeExecution(raw []byte) (ExecutionRecord, string) {
	var record ExecutionRecord
	if contracts.ValidateRaw("execution-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &record) != nil {
		return ExecutionRecord{}, InvalidSchema
	}
	return record, Valid
}

// ValidateApproval checks an approval against the immutable M08 proposal-only
// intent and its non-authorizing policy result at a runtime-owned time. It
// grants no execution authority and does not mutate caller state.
func ValidateApproval(intent m08.Intent, policy m08.PolicyDecision, approval ApprovalRecord, now time.Time) string {
	if intent.IntentHash == "" || intent.IntentHash != m08.ComputeIntentHash(intent) || intent.IntentMode != "PROPOSAL_ONLY" || intent.ExecutionAuthorized {
		return InvalidIntent
	}
	if status := m08.ValidatePolicyForIntent(intent, policy); status != "VALID" ||
		(policy.Decision != "ALLOW" && policy.Decision != "HUMAN_REVIEW") {
		return InvalidPolicy
	}
	if approval.Decision == "REJECT" {
		return RejectedApproval
	}
	if approval.Decision != "APPROVE" || approval.ApprovedBy != "human" || strings.TrimSpace(approval.ApproverID) == "" || !approval.OneTime {
		return InvalidApprover
	}
	if approval.ApprovalID == "" || approval.IntentID != intent.IntentID || approval.IntentHash != intent.IntentHash || approval.PolicyVersion != policy.PolicyVersion || approval.CorrelationID != intent.CorrelationID {
		return ApprovalMismatch
	}
	created, createdErr := time.Parse(time.RFC3339, intent.CreatedAt)
	intentExpires, intentExpiryErr := time.Parse(time.RFC3339, intent.ExpiresAt)
	policyChecked, policyErr := time.Parse(time.RFC3339, policy.PolicyCheckedAt)
	approved, approvedErr := time.Parse(time.RFC3339, approval.ApprovedAt)
	approvalExpires, approvalExpiryErr := time.Parse(time.RFC3339, approval.ExpiresAt)
	if createdErr != nil || intentExpiryErr != nil || policyErr != nil || approvedErr != nil || approvalExpiryErr != nil ||
		!intentExpires.After(created) || policyChecked.Before(created) || policyChecked.After(now) || approved.After(now) || !approvalExpires.After(approved) {
		return InvalidTimeBinding
	}
	if !intentExpires.After(now) {
		return ExpiredIntent
	}
	if !approvalExpires.After(now) {
		return ExpiredApproval
	}
	if approved.Before(policyChecked) {
		return ApprovalBeforePolicy
	}
	return Valid
}

// Error turns a status into a stable caller-facing error without making the
// status itself an authority decision.
func Error(status string) error {
	return fmt.Errorf("M09 approval: %s", status)
}
