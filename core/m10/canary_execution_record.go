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

// ExecutionRecord is an immutable audit artifact. This learner implementation
// only records a cancelled canary operation; it never invokes an executor.
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

func cancelledExecutionID(in CancelledExecutionInput) string {
	raw, _ := json.Marshal([]string{in.Authorization.AuthorizationID, in.AttemptedAt, in.Reason})
	sum := sha256.Sum256(raw)
	return "canary-exec-" + hex.EncodeToString(sum[:])
}

// CancelCanaryExecution creates a terminal audit record. It deliberately does
// not reserve budget, mutate a ledger, or contact the named executor.
func CancelCanaryExecution(in CancelledExecutionInput) (ExecutionRecord, error) {
	authorizationRaw, err := json.Marshal(in.Authorization)
	if err != nil {
		return ExecutionRecord{}, err
	}
	authorization, err := ValidateExecutionAuthorization(authorizationRaw)
	if err != nil {
		return ExecutionRecord{}, fmt.Errorf("invalid execution authorization: %w", err)
	}
	attempted, attemptedErr := time.Parse(time.RFC3339, in.AttemptedAt)
	authorized, authorizedErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
	if attemptedErr != nil || authorizedErr != nil || attempted.Before(authorized) {
		return ExecutionRecord{}, fmt.Errorf("cancellation timestamp is invalid")
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		return ExecutionRecord{}, fmt.Errorf("cancellation reason is required")
	}
	return ExecutionRecord{
		ExecutionID:     cancelledExecutionID(CancelledExecutionInput{Authorization: authorization, AttemptedAt: in.AttemptedAt, Reason: reason}),
		AuthorizationID: authorization.AuthorizationID, CanaryGrantID: authorization.CanaryGrantID,
		CanaryGrantVersion: authorization.CanaryGrantVersion, CanaryGrantHash: authorization.CanaryGrantHash,
		CanaryGateID: authorization.CanaryGateID, CanaryCostBoundID: authorization.CanaryCostBoundID,
		CanaryCostBoundHash: authorization.CanaryCostBoundHash, CanaryCostBoundMinor: authorization.CanaryCostBoundMinor,
		IntentID: authorization.IntentID, IntentHash: authorization.IntentHash, ExecutorID: authorization.ExecutorID,
		IdempotencyKey: authorization.IdempotencyKey, AttemptedAt: in.AttemptedAt, Status: "CANCELLED",
		SideEffectState: "NOT_PERFORMED", Error: reason, CorrelationID: authorization.CorrelationID,
	}, nil
}

func ValidateExecutionRecord(raw []byte) (ExecutionRecord, error) {
	var record ExecutionRecord
	if err := contracts.ValidateRaw("execution-record.schema.json", raw); err != nil {
		return record, err
	}
	if err := contracts.DecodeStrict(raw, &record); err != nil {
		return ExecutionRecord{}, err
	}
	if record.Status != "CANCELLED" || record.SideEffectState != "NOT_PERFORMED" || record.AuthorizationID == "" || record.CanaryGrantID == "" || record.CanaryGrantVersion == "" || record.CanaryGrantHash == "" || record.CanaryGateID == "" || record.CanaryCostBoundID == "" || record.CanaryCostBoundHash == "" || record.Error == "" {
		return ExecutionRecord{}, fmt.Errorf("invalid cancelled governed-canary execution record")
	}
	return record, nil
}
