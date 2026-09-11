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

// ValidateApproval checks an approval against the immutable M08 proposal-only
// intent and its non-authorizing policy result at a runtime-owned time. It
// grants no execution authority and does not mutate caller state.
func ValidateApproval(intent m08.Intent, policy m08.PolicyDecision, approval ApprovalRecord, now time.Time) string {
	if intent.IntentHash == "" || intent.IntentHash != m08.ComputeIntentHash(intent) || intent.IntentMode != "PROPOSAL_ONLY" || intent.ExecutionAuthorized {
		return InvalidIntent
	}
	if policy.IntentID != intent.IntentID || policy.IntentHash != intent.IntentHash || policy.PolicyVersion == "" ||
		(policy.Decision != "ALLOW" && policy.Decision != "HUMAN_REVIEW") || policy.PolicyMode != "NON_AUTHORIZING" || policy.ExecutionAuthorized {
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
