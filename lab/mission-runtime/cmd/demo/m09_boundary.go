package main

import (
	"encoding/json"
	"fmt"
	corem09 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m09"
	"io"
	"os"
)

func DecodeM09Approval(raw []byte) (ApprovalRecord, string) {
	a, status := corem09.DecodeApproval(raw)
	if status != corem09.Valid {
		return ApprovalRecord{}, "INVALID_SCHEMA"
	}
	return ApprovalRecord(a), missionValid
}
func DecodeM09Authorization(raw []byte) (ExecutionAuthorization, string) {
	a, status := corem09.DecodeAuthorization(raw)
	if status != corem09.Valid {
		return ExecutionAuthorization{}, status
	}
	return a, missionValid
}
func DecodeM09Execution(raw []byte) (ExecutionRecord, string) {
	r, status := corem09.DecodeExecution(raw)
	if status != corem09.Valid {
		return ExecutionRecord{}, status
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
	if state := corem09.ValidateHistoricalChain(coreIntent(i), coreM09Policy(p), corem09.ApprovalRecord(a), corem09.ExecutionAuthorization(auth), corem09.ExecutionRecord(r)); state != corem09.Valid {
		return M09CheckSummary{}, state
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
