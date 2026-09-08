package m10

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// ExecutionRecord is an immutable audit artifact. The learner profile records
// only terminal no-side-effect outcomes; it never invokes an executor.
type ExecutionRecord struct {
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
	Error                string `json:"error,omitempty"`
	CorrelationID        string `json:"correlation_id"`
}

type CancelledExecutionInput struct {
	Authorization ExecutionAuthorization
	AttemptedAt   string
	Reason        string
}

type FailedExecutionInput struct {
	Authorization ExecutionAuthorization
	AttemptedAt   string
	Reason        string
}

func terminalExecutionID(authorization ExecutionAuthorization, attemptedAt, status, reason string) string {
	raw, _ := json.Marshal([]string{authorization.AuthorizationID, attemptedAt, status, reason})
	sum := sha256.Sum256(raw)
	return "canary-exec-" + hex.EncodeToString(sum[:])
}

func terminalNoSideEffectExecution(authorization ExecutionAuthorization, attemptedAt, status, reason string) (ExecutionRecord, error) {
	authorizationRaw, err := json.Marshal(authorization)
	if err != nil {
		return ExecutionRecord{}, err
	}
	authorization, err = ValidateExecutionAuthorization(authorizationRaw)
	if err != nil {
		return ExecutionRecord{}, fmt.Errorf("invalid execution authorization: %w", err)
	}
	attempted, attemptedErr := time.Parse(time.RFC3339, attemptedAt)
	authorized, authorizedErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
	if attemptedErr != nil || authorizedErr != nil || attempted.Before(authorized) {
		return ExecutionRecord{}, fmt.Errorf("execution timestamp is invalid")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ExecutionRecord{}, fmt.Errorf("execution reason is required")
	}
	return ExecutionRecord{
		ExecutionID:     terminalExecutionID(authorization, attemptedAt, status, reason),
		AuthorizationID: authorization.AuthorizationID, CanaryGrantID: authorization.CanaryGrantID,
		CanaryGrantVersion: authorization.CanaryGrantVersion, CanaryGrantHash: authorization.CanaryGrantHash,
		CanaryGateID: authorization.CanaryGateID, CanaryCostBoundID: authorization.CanaryCostBoundID,
		CanaryCostBoundHash: authorization.CanaryCostBoundHash, CanaryCostBoundMinor: authorization.CanaryCostBoundMinor,
		IntentID: authorization.IntentID, IntentHash: authorization.IntentHash, ExecutorID: authorization.ExecutorID,
		IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: attemptedAt, Status: status,
		SideEffectState: "NOT_PERFORMED", Error: reason, CorrelationID: authorization.CorrelationID,
	}, nil
}

// CancelCanaryExecution creates a terminal audit record. It deliberately does
// not reserve budget, mutate a ledger, or contact the named executor.
func CancelCanaryExecution(in CancelledExecutionInput) (ExecutionRecord, error) {
	return terminalNoSideEffectExecution(in.Authorization, in.AttemptedAt, "CANCELLED", in.Reason)
}

// FailCanaryExecutionFixture is a deterministic local stub. It models a
// failed attempt before an executor was contacted and cannot represent a live
// execution or a performed side effect.
func FailCanaryExecutionFixture(in FailedExecutionInput) (ExecutionRecord, error) {
	return terminalNoSideEffectExecution(in.Authorization, in.AttemptedAt, "FAILED", in.Reason)
}

func ValidateExecutionRecord(raw []byte) (ExecutionRecord, error) {
	var record ExecutionRecord
	if err := contracts.ValidateRaw("execution-record.schema.json", raw); err != nil {
		return record, err
	}
	if err := contracts.DecodeStrict(raw, &record); err != nil {
		return ExecutionRecord{}, err
	}
	if (record.Status != "CANCELLED" && record.Status != "FAILED") || record.SideEffectState != "NOT_PERFORMED" || record.AuthorizationID == "" || record.CanaryGrantID == "" || record.CanaryGrantVersion == "" || record.CanaryGrantHash == "" || record.CanaryGateID == "" || record.CanaryCostBoundID == "" || record.CanaryCostBoundHash == "" || record.Error == "" {
		return ExecutionRecord{}, fmt.Errorf("invalid terminal no-side-effect governed-canary execution record")
	}
	return record, nil
}
