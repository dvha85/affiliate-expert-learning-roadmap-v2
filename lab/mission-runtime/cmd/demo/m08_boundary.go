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

func DecodeM08Intent(raw []byte) (ShadowActionIntent, string) {
	var i ShadowActionIntent
	if contracts.ValidateRaw("action-intent.schema.json", raw) != nil || contracts.DecodeStrict(raw, &i) != nil {
		return i, "INVALID_SCHEMA"
	}
	// Preserve arbitrary JSON numbers in parameters rather than rounding float64.
	value, err := contracts.Decode(raw)
	if err != nil {
		return i, "INVALID_SCHEMA"
	}
	i.Parameters = value.(map[string]any)["parameters"].(map[string]any)
	return i, missionValid
}
func DecodeM08Policy(raw []byte) (ShadowPolicyDecision, string) {
	var p ShadowPolicyDecision
	if contracts.ValidateRaw("policy-decision.schema.json", raw) != nil || contracts.DecodeStrict(raw, &p) != nil {
		return p, "INVALID_SCHEMA"
	}
	return p, missionValid
}
func DecodeM08Context(raw []byte) (ShadowPolicyContext, string) {
	var ctx ShadowPolicyContext
	value, err := contracts.Decode(raw)
	if err != nil {
		return ctx, "INVALID_CONTEXT"
	}
	object, ok := value.(map[string]any)
	if !ok {
		return ctx, "INVALID_CONTEXT"
	}
	for _, key := range []string{"policy_version", "now", "known_decision_ids", "known_evidence_ids", "known_proposal_ids", "allowed_hosts", "action_risk", "seen_idempotency"} {
		if object[key] == nil {
			return ctx, "INVALID_CONTEXT"
		}
	}
	for _, key := range []string{"known_decision_ids", "known_evidence_ids", "known_proposal_ids", "allowed_hosts"} {
		items, ok := object[key].([]any)
		if !ok {
			return ctx, "INVALID_CONTEXT"
		}
		seen := map[string]bool{}
		for _, item := range items {
			s, ok := item.(string)
			if !ok || strings.TrimSpace(s) == "" || s != strings.TrimSpace(s) || seen[s] {
				return ctx, "INVALID_CONTEXT"
			}
			seen[s] = true
		}
	}
	for _, key := range []string{"action_risk", "seen_idempotency"} {
		items, ok := object[key].(map[string]any)
		if !ok {
			return ctx, "INVALID_CONTEXT"
		}
		for k, v := range items {
			s, ok := v.(string)
			if !ok || strings.TrimSpace(k) == "" || strings.TrimSpace(s) == "" {
				return ctx, "INVALID_CONTEXT"
			}
			if key == "action_risk" && s != "RISK0" && s != "RISK1" && s != "RISK2" {
				return ctx, "INVALID_CONTEXT"
			}
		}
	}
	if contracts.DecodeStrict(raw, &ctx) != nil || strings.TrimSpace(ctx.PolicyVersion) == "" {
		return ctx, "INVALID_CONTEXT"
	}
	if _, err := time.Parse(time.RFC3339, ctx.Now); err != nil {
		return ctx, "INVALID_CONTEXT"
	}
	return ctx, missionValid
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
