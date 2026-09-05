package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func m08BoundaryFixture() (ShadowActionIntent, ShadowPolicyContext) {
	i := SealShadowActionIntent(ShadowActionIntent{IntentID: "syn-i", DecisionID: "syn-d", EvidenceIDs: []string{"syn-e"}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-03T01:00:00Z", ExpiresAt: "2026-09-03T03:00:00Z", CorrelationID: "syn-c", IdempotencyKey: "syn-k"})
	c := ShadowPolicyContext{PolicyVersion: "test-v1", Now: "2026-09-03T02:00:00Z", KnownDecisionIDs: []string{"syn-d"}, KnownEvidenceIDs: []string{"syn-e"}, KnownProposalIDs: []string{}, AllowedHosts: []string{"example.com"}, ActionRisk: map[string]string{"DRAFT": "RISK0"}, SeenIdempotency: map[string]string{}}
	return i, c
}
func TestM08RawSchema(t *testing.T) {
	i, c := m08BoundaryFixture()
	p := EvaluateShadowPolicy(i, c)
	for _, item := range []struct {
		value  any
		decode func([]byte) string
	}{{i, func(b []byte) string { _, s := DecodeM08Intent(b); return s }}, {p, func(b []byte) string { _, s := DecodeM08Policy(b); return s }}} {
		raw, _ := json.Marshal(item.value)
		var fields map[string]json.RawMessage
		json.Unmarshal(raw, &fields)
		for key := range fields {
			for _, mode := range []string{"missing", "null"} {
				copy := map[string]json.RawMessage{}
				for k, v := range fields {
					copy[k] = v
				}
				if mode == "missing" {
					delete(copy, key)
				} else {
					copy[key] = json.RawMessage(`null`)
				}
				bad, _ := json.Marshal(copy)
				if s := item.decode(bad); s != "INVALID_SCHEMA" {
					t.Fatal(key, mode, s)
				}
			}
		}
		for _, bad := range []string{string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"extra":1,`, 1), strings.Replace(string(raw), `"execution_authorized":false`, `"execution_authorized":true`, 1), strings.Replace(string(raw), `"intent_id":`, `"intent_id":"other","intent_id":`, 1)} {
			if item.decode([]byte(bad)) != "INVALID_SCHEMA" {
				t.Fatal("bad schema accepted")
			}
		}
	}
	i.ProposedBy = "agent"
	raw, _ := json.Marshal(i)
	if _, s := DecodeM08Intent(raw); s != "INVALID_SCHEMA" {
		t.Fatal("agent missing proposal accepted")
	}
	p.Decision = "HUMAN_REVIEW"
	p.PolicyReviewRequired = false
	raw, _ = json.Marshal(p)
	if _, s := DecodeM08Policy(raw); s != "INVALID_SCHEMA" {
		t.Fatal("review inconsistency")
	}
}
func TestM08BoundaryPolicy(t *testing.T) {
	for _, mode := range []string{"allow", "risk1", "risk2", "missing_decision", "missing_evidence", "expired", "tamper", "duplicate", "collision"} {
		t.Run(mode, func(t *testing.T) {
			i, c := m08BoundaryFixture()
			want := "SHADOW_POLICY_ALLOW"
			switch mode {
			case "risk1":
				c.ActionRisk["DRAFT"] = "RISK1"
				want = "RISK1_REQUIRES_REVIEW"
			case "risk2":
				c.ActionRisk["DRAFT"] = "RISK2"
				want = "RISK2_REQUIRES_REVIEW"
			case "missing_decision":
				c.KnownDecisionIDs = []string{}
				want = "MISSING_DECISION_LINK"
			case "missing_evidence":
				c.KnownEvidenceIDs = []string{}
				want = "MISSING_EVIDENCE_LINK"
			case "expired":
				c.Now = i.ExpiresAt
				want = "EXPIRED_INTENT"
			case "tamper":
				i.Target = "https://example.com/changed"
				want = "TAMPERED_INTENT"
			case "duplicate":
				c.SeenIdempotency[i.IdempotencyKey] = i.IntentHash
				want = "DUPLICATE_INTENT"
			case "collision":
				c.SeenIdempotency[i.IdempotencyKey] = "other"
				want = "IDEMPOTENCY_COLLISION"
			}
			raw, _ := json.Marshal(i)
			decoded, s := DecodeM08Intent(raw)
			if s != missionValid {
				t.Fatal(s)
			}
			p := EvaluateShadowPolicy(decoded, c)
			if p.Reason != want || p.ExecutionAuthorized || p.PolicyMode != "NON_AUTHORIZING" {
				t.Fatal(p)
			}
			out, _ := json.Marshal(p)
			if _, s := DecodeM08Policy(out); s != missionValid {
				t.Fatal(s)
			}
		})
	}
}
func TestM08ParameterNumbers(t *testing.T) {
	i, _ := m08BoundaryFixture()
	i.Parameters = map[string]any{"id": json.Number("9007199254740993")}
	i = SealShadowActionIntent(i)
	raw, _ := json.Marshal(i)
	decoded, s := DecodeM08Intent(raw)
	if s != missionValid || ComputeShadowIntentHash(decoded) != i.IntentHash {
		t.Fatal("number rounded", s)
	}
}
func TestM08CLI(t *testing.T) {
	i, c := m08BoundaryFixture()
	dir := t.TempDir()
	a, b := filepath.Join(dir, "intent.json"), filepath.Join(dir, "context.json")
	raw, _ := json.Marshal(i)
	ctx, _ := json.Marshal(c)
	os.WriteFile(a, raw, 0600)
	os.WriteFile(b, ctx, 0600)
	var out bytes.Buffer
	if err := runM08Check(&out, []string{a, b}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"decision":"ALLOW"`)) || bytes.Contains(out.Bytes(), []byte(`"execution_authorized":true`)) {
		t.Fatal("authority")
	}
	if err := runM08Check(m03BrokenWriter{}, []string{a, b}); err == nil {
		t.Fatal("writer")
	}
	for _, bad := range []string{`{}`, strings.Replace(string(raw), `false`, `true`, 1)} {
		os.WriteFile(a, []byte(bad), 0600)
		out.Reset()
		if err := runM08Check(&out, []string{a, b}); err == nil || out.Len() != 0 {
			t.Fatal("bad input accepted")
		}
	}
	os.WriteFile(a, raw, 0600)
	bad := strings.Replace(string(ctx), `"known_decision_ids":["syn-d"]`, `"known_decision_ids":[null]`, 1)
	os.WriteFile(b, []byte(bad), 0600)
	out.Reset()
	if err := runM08Check(&out, []string{a, b}); err == nil || out.Len() != 0 {
		t.Fatal("bad context")
	}
	got, _ := os.ReadFile(b)
	if string(got) != bad {
		t.Fatal("input modified")
	}
	if err := runM08Check(&out, nil); err == nil {
		t.Fatal("args")
	}
}
