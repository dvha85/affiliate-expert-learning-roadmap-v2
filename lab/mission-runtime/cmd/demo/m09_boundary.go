package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"io"
	"os"
	"strings"
	"time"
)

func DecodeM09Approval(raw []byte) (ApprovalRecord, string) {
	var a ApprovalRecord
	if contracts.ValidateRaw("approval-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &a) != nil {
		return ApprovalRecord{}, "INVALID_SCHEMA"
	}
	return a, missionValid
}
func DecodeM09Authorization(raw []byte) (ExecutionAuthorization, string) {
	var a ExecutionAuthorization
	if contracts.ValidateRaw("execution-authorization.schema.json", raw) != nil || contracts.DecodeStrict(raw, &a) != nil {
		return ExecutionAuthorization{}, "INVALID_SCHEMA"
	}
	if a.ExecutionMode != "APPROVED_LIVE" {
		return ExecutionAuthorization{}, "INVALID_PROFILE"
	}
	return a, missionValid
}
func DecodeM09Execution(raw []byte) (ExecutionRecord, string) {
	var r ExecutionRecord
	if contracts.ValidateRaw("execution-record.schema.json", raw) != nil || contracts.DecodeStrict(raw, &r) != nil {
		return ExecutionRecord{}, "INVALID_SCHEMA"
	}
	return r, missionValid
}

type M09CheckSummary struct {
	Result                string `json:"result"`
	IntentID              string `json:"intent_id"`
	ApprovalID            string `json:"approval_id"`
	AuthorizationID       string `json:"authorization_id"`
	ExecutionID           string `json:"execution_id"`
	ExecutionStatus       string `json:"execution_status"`
	SideEffectState       string `json:"side_effect_state"`
	ApprovalAuthenticated bool   `json:"approval_authenticated"`
	ExecutionPermitted    bool   `json:"execution_permitted"`
}

// CheckM09Chain audits supplied historical files. It must never authorize,
// consume approval, fill compatibility aliases, or invoke an executor.
func CheckM09Chain(intentRaw, policyRaw, approvalRaw, authorizationRaw, executionRaw []byte) (M09CheckSummary, string) {
	i, state := DecodeM08Intent(intentRaw)
	if state != missionValid {
		return M09CheckSummary{}, state
	}
	p, state := DecodeM08Policy(policyRaw)
	if state != missionValid {
		return M09CheckSummary{}, state
	}
	a, state := DecodeM09Approval(approvalRaw)
	if state != missionValid {
		return M09CheckSummary{}, state
	}
	auth, state := DecodeM09Authorization(authorizationRaw)
	if state != missionValid {
		return M09CheckSummary{}, state
	}
	r, state := DecodeM09Execution(executionRaw)
	if state != missionValid {
		return M09CheckSummary{}, state
	}
	for _, v := range []string{i.IntentID, i.DecisionID, i.CorrelationID, i.IdempotencyKey, p.PolicyVersion, a.ApprovalID, a.ApproverID, auth.AuthorizationID, auth.ExecutorID, r.ExecutionID} {
		if strings.TrimSpace(v) == "" {
			return M09CheckSummary{}, missionInvalid
		}
	}
	if i.IntentHash != ComputeShadowIntentHash(i) {
		return M09CheckSummary{}, "TAMPERED_INTENT"
	}
	if p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || a.IntentID != i.IntentID || a.IntentHash != i.IntentHash || a.PolicyVersion != p.PolicyVersion || a.CorrelationID != i.CorrelationID {
		return M09CheckSummary{}, "BROKEN_LINK"
	}
	if a.Decision != "APPROVE" {
		return M09CheckSummary{}, "REJECTED_APPROVAL"
	}
	if p.Decision != "ALLOW" && p.Decision != "HUMAN_REVIEW" {
		return M09CheckSummary{}, "INVALID_POLICY_STATE"
	}
	if auth.ApprovalID != a.ApprovalID || auth.IntentID != i.IntentID || auth.IntentHash != i.IntentHash || auth.PolicyVersion != p.PolicyVersion || auth.CorrelationID != i.CorrelationID || auth.IdempotencyKey != i.IdempotencyKey {
		return M09CheckSummary{}, "BROKEN_LINK"
	}
	if r.AuthorizationID != auth.AuthorizationID || r.ApprovalID != a.ApprovalID || r.IntentID != i.IntentID || r.IntentHash != i.IntentHash || r.ExecutorID != auth.ExecutorID || r.IdempotencyKey != i.IdempotencyKey || r.CorrelationID != i.CorrelationID {
		return M09CheckSummary{}, "BROKEN_LINK"
	}
	times := make([]time.Time, 8)
	for n, value := range []string{i.CreatedAt, i.ExpiresAt, p.PolicyCheckedAt, a.ApprovedAt, a.ExpiresAt, auth.AuthorizedAt, auth.ExpiresAt, r.AttemptedAt} {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return M09CheckSummary{}, "INVALID_TIME_BINDING"
		}
		times[n] = parsed
	}
	created, intentEnd, checked, approved, approvalEnd, authorized, authEnd, attempted := times[0], times[1], times[2], times[3], times[4], times[5], times[6], times[7]
	if !intentEnd.After(created) || checked.Before(created) || approved.Before(checked) || !approvalEnd.After(approved) || authorized.Before(approved) || !authEnd.After(authorized) || authEnd.After(intentEnd) || authEnd.After(approvalEnd) || attempted.Before(authorized) {
		return M09CheckSummary{}, "INVALID_TIME_BINDING"
	}
	// Failed/unknown attempts may be recorded after expiry; a claimed performed
	// effect must have been attempted within the authorization window.
	if r.SideEffectState == "PERFORMED" && !attempted.Before(authEnd) {
		return M09CheckSummary{}, "EXPIRED_AUTHORIZATION"
	}
	return M09CheckSummary{Result: "CONSISTENT_UNVERIFIED", IntentID: i.IntentID, ApprovalID: a.ApprovalID, AuthorizationID: auth.AuthorizationID, ExecutionID: r.ExecutionID, ExecutionStatus: r.Status, SideEffectState: r.SideEffectState, ApprovalAuthenticated: false, ExecutionPermitted: false}, missionValid
}

func runM09Check(w io.Writer, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: demo m09-check INTENT.json POLICY.json APPROVAL.json AUTHORIZATION.json EXECUTION.json")
	}
	raw := make([][]byte, 5)
	for n, path := range args {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw[n] = data
	}
	summary, state := CheckM09Chain(raw[0], raw[1], raw[2], raw[3], raw[4])
	if state != missionValid {
		return fmt.Errorf("M09 audit: %s", state)
	}
	// Only emit a non-authorizing summary, never a freshly issued authorization.
	return json.NewEncoder(w).Encode(summary)
}
