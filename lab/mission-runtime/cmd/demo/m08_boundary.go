package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
)

func DecodeM08Intent(raw []byte) (ShadowActionIntent, string) {
	decoded, status := corem08.DecodeIntent(raw)
	if status != "VALID" {
		return ShadowActionIntent{}, status
	}
	i := ShadowActionIntent{IntentID: decoded.IntentID, DecisionID: decoded.DecisionID, EvidenceIDs: decoded.EvidenceIDs, ActionType: decoded.ActionType, Target: decoded.Target, Parameters: decoded.Parameters, ProposedBy: decoded.ProposedBy, ProposalRef: decoded.ProposalRef, CreatedAt: decoded.CreatedAt, ExpiresAt: decoded.ExpiresAt, CorrelationID: decoded.CorrelationID, IdempotencyKey: decoded.IdempotencyKey, IntentHash: decoded.IntentHash, IntentMode: decoded.IntentMode, ExecutionAuthorized: decoded.ExecutionAuthorized}
	i.ShadowOnly = i.IntentMode == "PROPOSAL_ONLY" && !i.ExecutionAuthorized
	i.DryRun = i.ShadowOnly
	return i, missionValid
}
func DecodeM08Policy(raw []byte) (ShadowPolicyDecision, string) {
	decoded, status := corem08.DecodePolicy(raw)
	if status != "VALID" {
		return ShadowPolicyDecision{}, "INVALID_SCHEMA"
	}
	return ShadowPolicyDecision{PolicyVersion: decoded.PolicyVersion, IntentID: decoded.IntentID, IntentHash: decoded.IntentHash, Decision: decoded.Decision, RiskClass: decoded.RiskClass, Reason: decoded.Reason, PolicyReviewRequired: decoded.PolicyReviewRequired, PolicyMode: decoded.PolicyMode, ExecutionAuthorized: decoded.ExecutionAuthorized, PolicyCheckedAt: decoded.PolicyCheckedAt, ApprovalRequired: decoded.PolicyReviewRequired, ShadowOnly: true}, missionValid
}
func DecodeM08Context(raw []byte) (ShadowPolicyContext, string) {
	ctx, status := corem08.DecodePolicyContext(raw)
	if status != "VALID" {
		return ShadowPolicyContext{}, "INVALID_CONTEXT"
	}
	return ShadowPolicyContext(ctx), missionValid
}

func runM08Check(w io.Writer, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: demo m08-check INTENT.json CONTEXT.json")
	}
	raw, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	contextRaw, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	intent, state := DecodeM08Intent(raw)
	if state != missionValid {
		return fmt.Errorf("M08 intent: %s", state)
	}
	ctx, state := DecodeM08Context(contextRaw)
	if state != missionValid {
		return fmt.Errorf("M08 context: %s", state)
	}
	// Never SealShadowActionIntent here: submitted hash/mode/authority are evidence.
	policy := EvaluateShadowPolicy(intent, ctx)
	policyRaw, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	if _, state := DecodeM08Policy(policyRaw); state != missionValid {
		return fmt.Errorf("M08 policy output: %s", state)
	}
	intentRaw, err := json.Marshal(intent)
	if err != nil {
		return err
	}
	if _, state := DecodeM08Intent(intentRaw); state != missionValid {
		return fmt.Errorf("M08 intent output: %s", state)
	}
	return json.NewEncoder(w).Encode(struct {
		Intent ShadowActionIntent   `json:"intent"`
		Policy ShadowPolicyDecision `json:"policy"`
	}{intent, policy})
}
