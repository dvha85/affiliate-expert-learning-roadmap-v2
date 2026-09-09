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

// ExecutionAuthorization is a capability artifact. Its creation has no side
// effect; a separate executor must still resolve it again before acting.
type ExecutionAuthorization struct {
	AuthorizationID      string `json:"authorization_id"`
	IntentID             string `json:"intent_id"`
	IntentHash           string `json:"intent_hash"`
	PolicyVersion        string `json:"policy_version"`
	CanaryGrantID        string `json:"canary_grant_id"`
	CanaryGrantVersion   string `json:"canary_grant_version"`
	CanaryGrantHash      string `json:"canary_grant_hash"`
	CanaryGateID         string `json:"canary_gate_id"`
	CanaryCostBoundID    string `json:"canary_cost_bound_id"`
	CanaryCostBoundHash  string `json:"canary_cost_bound_hash"`
	CanaryCostBoundMinor int64  `json:"canary_cost_bound_minor"`
	ExecutorID           string `json:"executor_id"`
	AuthorizedAt         string `json:"authorized_at"`
	ExpiresAt            string `json:"expires_at"`
	IdempotencyKey       string `json:"idempotency_key"`
	CorrelationID        string `json:"correlation_id"`
	ExecutionMode        string `json:"execution_mode"`
	ExecutionAuthorized  bool   `json:"execution_authorized"`
}

type CanaryAuthorizationInput struct {
	Gate            CanaryGateDecision
	Grant           CanaryGrant
	CostBound       TrustedCostBound
	IntentID        string
	IntentHash      string
	PolicyVersion   string
	IdempotencyKey  string
	CorrelationID   string
	IntentExpiresAt string
	ExecutorID      string
	AuthorizedAt    string
}

func authorizationID(in CanaryAuthorizationInput) string {
	raw, _ := json.Marshal([]string{in.Gate.GateID, in.IntentID, in.IntentHash, in.ExecutorID, in.AuthorizedAt, in.IdempotencyKey})
	sum := sha256.Sum256(raw)
	return "canary-auth-" + hex.EncodeToString(sum[:])
}

func minAuthorizationExpiry(values ...time.Time) time.Time {
	result := values[0]
	for _, value := range values[1:] {
		if value.Before(result) {
			result = value
		}
	}
	return result
}

// AuthorizeCanary checks the persisted, non-authorizing gate's exact links
// before minting a short-lived governed-canary capability. It does not reserve
// budget, write a ledger, or invoke an executor.
func AuthorizeCanary(in CanaryAuthorizationInput) (ExecutionAuthorization, error) {
	if in.Gate.Decision != "ALLOW_CANARY" || in.Gate.ExecutionAuthorized || in.Gate.PerActionApprovalRequired || in.Gate.GrantID != in.Grant.GrantID || in.Gate.GrantVersion != in.Grant.GrantVersion || in.Gate.GrantHash != in.Grant.GrantHash || in.Gate.CostBoundID != in.CostBound.CostBoundID || in.Gate.CostBoundHash != in.CostBound.CostBoundHash || in.Gate.CostBoundMinor != in.CostBound.MaxCostMinor || in.Gate.IntentID != in.IntentID || in.Gate.IntentHash != in.IntentHash || in.Gate.PolicyVersion != in.PolicyVersion {
		return ExecutionAuthorization{}, fmt.Errorf("canary gate binding is invalid")
	}
	if !containsGrantValue(in.Grant.ExecutorIDs, in.ExecutorID) || strings.TrimSpace(in.IdempotencyKey) == "" || strings.TrimSpace(in.CorrelationID) == "" {
		return ExecutionAuthorization{}, fmt.Errorf("executor or idempotency binding is invalid")
	}
	authorized, authorizedErr := time.Parse(time.RFC3339, in.AuthorizedAt)
	intentExpiry, intentErr := time.Parse(time.RFC3339, in.IntentExpiresAt)
	grantExpiry, grantErr := time.Parse(time.RFC3339, in.Grant.ExpiresAt)
	costExpiry, costErr := time.Parse(time.RFC3339, in.CostBound.ExpiresAt)
	if authorizedErr != nil || intentErr != nil || grantErr != nil || costErr != nil {
		return ExecutionAuthorization{}, fmt.Errorf("authorization time is invalid")
	}
	expires := minAuthorizationExpiry(intentExpiry, grantExpiry, costExpiry)
	if !expires.After(authorized) {
		return ExecutionAuthorization{}, fmt.Errorf("authorization inputs have expired")
	}
	return ExecutionAuthorization{
		AuthorizationID: authorizationID(in), IntentID: in.IntentID, IntentHash: in.IntentHash, PolicyVersion: in.PolicyVersion,
		CanaryGrantID: in.Grant.GrantID, CanaryGrantVersion: in.Grant.GrantVersion, CanaryGrantHash: in.Grant.GrantHash,
		CanaryGateID: in.Gate.GateID, CanaryCostBoundID: in.CostBound.CostBoundID, CanaryCostBoundHash: in.CostBound.CostBoundHash,
		CanaryCostBoundMinor: in.CostBound.MaxCostMinor, ExecutorID: in.ExecutorID, AuthorizedAt: in.AuthorizedAt,
		ExpiresAt: expires.UTC().Format(time.RFC3339Nano), IdempotencyKey: in.IdempotencyKey, CorrelationID: in.CorrelationID,
		ExecutionMode: "GOVERNED_CANARY", ExecutionAuthorized: true,
	}, nil
}

func ValidateExecutionAuthorization(raw []byte) (ExecutionAuthorization, error) {
	var authorization ExecutionAuthorization
	if err := contracts.ValidateRaw("execution-authorization.schema.json", raw); err != nil {
		return authorization, err
	}
	if err := contracts.DecodeStrict(raw, &authorization); err != nil || !authorization.ExecutionAuthorized || authorization.ExecutionMode != "GOVERNED_CANARY" {
		return ExecutionAuthorization{}, fmt.Errorf("invalid governed canary authorization")
	}
	return authorization, nil
}
