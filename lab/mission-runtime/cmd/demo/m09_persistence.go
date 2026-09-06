package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// DecodeM09State validates the persistence envelope and every present artifact
// before typed decoders can restore compatibility aliases. It grants no authority.
func DecodeM09State(raw []byte) (M09State, error) {
	fail := func(reason string) (M09State, error) {
		return M09State{}, fmt.Errorf("M09 persisted state: %s", reason)
	}
	v, e := contracts.Decode(raw)
	if e != nil {
		return fail("INVALID_JSON")
	}
	object, ok := v.(map[string]any)
	if !ok {
		return fail("INVALID_ENVELOPE")
	}
	schemas := map[string]string{"intent": "action-intent.schema.json", "policy": "policy-decision.schema.json", "approval": "approval-record.schema.json", "authorization": "execution-authorization.schema.json", "execution": "execution-record.schema.json"}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return fail("INVALID_ENVELOPE")
	}
	if object["intent"] == nil || object["policy"] == nil {
		return fail("MISSING_REQUIRED_ARTIFACT")
	}
	for key, value := range object {
		if schema, ok := schemas[key]; ok {
			if contracts.ValidateRaw(schema, fields[key]) != nil {
				return fail("INVALID_SCHEMA: " + key)
			}
			continue
		}
		if key != "consumed_approval_ids" && key != "succeeded_idempotency" {
			return fail("UNKNOWN_FIELD")
		}
		markers, ok := value.(map[string]any)
		if !ok {
			return fail("INVALID_MARKERS")
		}
		for id, flag := range markers {
			b, ok := flag.(bool)
			if !ok || !b || strings.TrimSpace(id) == "" {
				return fail("INVALID_MARKERS")
			}
		}
	}
	s := M09State{}
	i, status := DecodeM08Intent(fields["intent"])
	if status != missionValid {
		return fail(status)
	}
	s.Intent = i
	p, status := DecodeM08Policy(fields["policy"])
	if status != missionValid {
		return fail(status)
	}
	s.Policy = p
	if b, ok := fields["approval"]; ok {
		a, status := DecodeM09Approval(b)
		if status != missionValid {
			return fail(status)
		}
		s.Approval = &a
	}
	if b, ok := fields["authorization"]; ok {
		a, status := DecodeM09Authorization(b)
		if status != missionValid {
			return fail(status)
		}
		s.Authorization = &a
	}
	if b, ok := fields["execution"]; ok {
		r, status := DecodeM09Execution(b)
		if status != missionValid {
			return fail(status)
		}
		s.Execution = &r
	}
	// Legacy omitempty maps are accepted only as empty sets; no file is rewritten.
	s.ConsumedApprovalIDs = map[string]bool{}
	s.SucceededIdempotency = map[string]bool{}
	for key, dst := range map[string]map[string]bool{"consumed_approval_ids": s.ConsumedApprovalIDs, "succeeded_idempotency": s.SucceededIdempotency} {
		if entries, ok := object[key].(map[string]any); ok {
			for id := range entries {
				dst[id] = true
			}
		}
	}
	if i.IntentHash != ComputeShadowIntentHash(i) {
		return fail("TAMPERED_INTENT")
	}
	if p.IntentID != i.IntentID || p.IntentHash != i.IntentHash {
		return fail("BROKEN_POLICY_LINK")
	}
	if a := s.Approval; a != nil {
		if a.IntentID != i.IntentID || a.IntentHash != i.IntentHash || a.PolicyVersion != p.PolicyVersion || a.CorrelationID != i.CorrelationID {
			return fail("BROKEN_APPROVAL_LINK")
		}
	}
	if a := s.Authorization; a != nil {
		if s.Approval == nil || a.ApprovalID != s.Approval.ApprovalID || a.IntentID != i.IntentID || a.IntentHash != i.IntentHash || a.PolicyVersion != p.PolicyVersion || a.CorrelationID != i.CorrelationID || a.IdempotencyKey != i.IdempotencyKey {
			return fail("BROKEN_AUTHORIZATION_LINK")
		}
	}
	if r := s.Execution; r != nil {
		a := s.Authorization
		if a == nil || r.AuthorizationID != a.AuthorizationID || r.ApprovalID != a.ApprovalID || r.IntentID != i.IntentID || r.IntentHash != i.IntentHash || r.ExecutorID != a.ExecutorID || r.IdempotencyKey != i.IdempotencyKey || r.CorrelationID != i.CorrelationID {
			return fail("BROKEN_EXECUTION_LINK")
		}
		if (r.SideEffectState == "PERFORMED" || r.SideEffectState == "UNKNOWN") && !s.ConsumedApprovalIDs[r.ApprovalID] {
			return fail("MISSING_CONSUMED_MARKER")
		}
		if r.Status == "SUCCEEDED" && !s.SucceededIdempotency[r.IdempotencyKey] {
			return fail("MISSING_SUCCESS_MARKER")
		}
	}
	return s, nil
}

func encodeM09State(s M09State) ([]byte, error) {
	// New snapshots always emit marker objects, including empty ones.
	type snapshot struct {
		Intent        ShadowActionIntent      `json:"intent"`
		Policy        ShadowPolicyDecision    `json:"policy"`
		Approval      *ApprovalRecord         `json:"approval,omitempty"`
		Authorization *ExecutionAuthorization `json:"authorization,omitempty"`
		Execution     *ExecutionRecord        `json:"execution,omitempty"`
		Consumed      map[string]bool         `json:"consumed_approval_ids"`
		Succeeded     map[string]bool         `json:"succeeded_idempotency"`
	}
	consumed, succeeded := s.ConsumedApprovalIDs, s.SucceededIdempotency
	if consumed == nil {
		consumed = map[string]bool{}
	}
	if succeeded == nil {
		succeeded = map[string]bool{}
	}
	b, e := json.MarshalIndent(snapshot{s.Intent, s.Policy, s.Approval, s.Authorization, s.Execution, consumed, succeeded}, "", "  ")
	if e != nil {
		return nil, e
	}
	if _, e = DecodeM09State(b); e != nil {
		return nil, e
	}
	return b, nil
}
