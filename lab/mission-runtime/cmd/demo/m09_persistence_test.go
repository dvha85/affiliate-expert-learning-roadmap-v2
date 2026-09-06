package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func persistedM09(t *testing.T) []byte {
	t.Helper()
	s, c := baseM09()
	a, status := AuthorizeM09(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &a
	b, e := encodeM09State(s)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestM09PersistenceRejectsRawMutations(t *testing.T) {
	for _, tc := range []struct {
		artifact, key string
		value         any
		remove        bool
	}{
		{"intent", "execution_authorized", nil, true}, {"intent", "execution_authorized", nil, false}, {"intent", "shadow_only", true, false},
		{"intent", "target", "https://example.com/tampered", false}, {"policy", "intent_id", "wrong", false},
		{"approval", "one_time", nil, true}, {"approval", "approved_by", "agent", false}, {"approval", "intent_id", "wrong", false},
		{"authorization", "execution_mode", "GOVERNED_CANARY", false}, {"authorization", "idempotency_key", "wrong", false},
	} {
		t.Run(tc.artifact+"/"+tc.key, func(t *testing.T) {
			var envelope map[string]json.RawMessage
			json.Unmarshal(persistedM09(t), &envelope)
			var fields map[string]json.RawMessage
			json.Unmarshal(envelope[tc.artifact], &fields)
			if tc.remove {
				delete(fields, tc.key)
			} else {
				fields[tc.key], _ = json.Marshal(tc.value)
			}
			envelope[tc.artifact], _ = json.Marshal(fields)
			bad, _ := json.Marshal(envelope)
			p := filepath.Join(t.TempDir(), "state.json")
			if e := os.WriteFile(p, bad, 0600); e != nil {
				t.Fatal(e)
			}
			s, e := LoadM09State(p)
			if e == nil || !reflect.DeepEqual(s, M09State{}) {
				t.Fatal("invalid file exposed state", e, s)
			}
			got, _ := os.ReadFile(p)
			if !bytes.Equal(got, bad) {
				t.Fatal("loader rewrote input")
			}
		})
	}
	raw := persistedM09(t)
	for _, bad := range []string{`null`, `[]`, `{}`, string(raw) + ` {}`, strings.Replace(string(raw), `{`, `{"extra":true,`, 1), strings.Replace(string(raw), `{`, `{"intent":{},`, 1), strings.Replace(string(raw), `"intent_id":`, `"intent_id":"duplicate","intent_id":`, 1)} {
		if s, e := DecodeM09State([]byte(bad)); e == nil || !reflect.DeepEqual(s, M09State{}) {
			t.Fatal("ambiguous JSON accepted", e)
		}
	}
	for _, key := range []string{"intent", "policy", "approval", "authorization", "execution", "consumed_approval_ids", "succeeded_idempotency"} {
		var envelope map[string]json.RawMessage
		json.Unmarshal(raw, &envelope)
		envelope[key] = json.RawMessage(`null`)
		b, _ := json.Marshal(envelope)
		if _, e := DecodeM09State(b); e == nil {
			t.Fatal("null accepted", key)
		}
	}
	for _, value := range []string{`{"id":false}`, `{"id":1}`, `{" ":true}`, `[]`} {
		var envelope map[string]json.RawMessage
		json.Unmarshal(raw, &envelope)
		envelope["consumed_approval_ids"] = json.RawMessage(value)
		b, _ := json.Marshal(envelope)
		if _, e := DecodeM09State(b); e == nil {
			t.Fatal("bad markers", value)
		}
	}
}

func TestM09PersistenceExactNumbersAndResume(t *testing.T) {
	s, c := baseM09()
	s.Intent.Parameters = map[string]any{"large": json.Number("9007199254740993"), "nested": []any{json.Number("1.0000000000000001"), json.Number("1e20")}}
	s.Intent = SealShadowActionIntent(s.Intent)
	s.Policy = EvaluateShadowPolicy(s.Intent, c.PolicyContext)
	s.Approval.IntentHash = s.Intent.IntentHash
	p := filepath.Join(t.TempDir(), "state.json")
	if e := PersistM09State(p, s); e != nil {
		t.Fatal(e)
	}
	loaded, e := LoadM09State(p)
	if e != nil {
		t.Fatal(e)
	}
	if loaded.Intent.IntentHash != ComputeShadowIntentHash(loaded.Intent) || !reflect.DeepEqual(s.Intent.Parameters, loaded.Intent.Parameters) {
		t.Fatal("number/hash changed")
	}
	if !loaded.Intent.ShadowOnly || !loaded.Intent.DryRun || !loaded.Policy.ShadowOnly {
		t.Fatal("validated aliases missing")
	}
	if _, status := AuthorizeM09(loaded, c); status != "AUTHORIZED" {
		t.Fatal(status)
	}
	before, _ := os.ReadFile(p)
	if e := PersistM09State(p, loaded); e != nil {
		t.Fatal(e)
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("roundtrip unstable")
	}
	// An absent approval is a valid pending state, not a required live chain.
	s.Approval = nil
	if e := PersistM09State(p, s); e != nil {
		t.Fatal(e)
	}
	loaded, e = LoadM09State(p)
	if e != nil {
		t.Fatal(e)
	}
	if _, status := AuthorizeM09(loaded, c); status != "WAIT_APPROVAL" {
		t.Fatal(status)
	}
}

func TestM09PersistenceReplayMarkers(t *testing.T) {
	s, c := baseM09()
	a, status := AuthorizeM09(s, c)
	if status != "AUTHORIZED" {
		t.Fatal(status)
	}
	s.Authorization = &a
	if _, status := ExecuteLocalSandbox(&s, a, c, t.TempDir()); status != "EXECUTED" {
		t.Fatal(status)
	}
	s.ConsumedApprovalIDs["older-approval"] = true
	s.SucceededIdempotency["older-key"] = true
	p := filepath.Join(t.TempDir(), "state.json")
	if e := PersistM09State(p, s); e != nil {
		t.Fatal(e)
	}
	loaded, e := LoadM09State(p)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(s.ConsumedApprovalIDs, loaded.ConsumedApprovalIDs) || !reflect.DeepEqual(s.SucceededIdempotency, loaded.SucceededIdempotency) {
		t.Fatal("lost history")
	}
	if _, status := AuthorizeM09(loaded, c); status != "WAIT_ALREADY_EXECUTED" {
		t.Fatal(status)
	}
	raw, _ := os.ReadFile(p)
	for _, key := range []string{"consumed_approval_ids", "succeeded_idempotency"} {
		var env map[string]json.RawMessage
		json.Unmarshal(raw, &env)
		delete(env, key)
		b, _ := json.Marshal(env)
		if _, e := DecodeM09State(b); e == nil {
			t.Fatal("stripped marker accepted", key)
		}
	}
	// Legacy empty maps were omitted by the old writer; retain pending compatibility.
	pending, _ := baseM09()
	legacy, _ := json.Marshal(pending)
	if _, e := DecodeM09State(legacy); e != nil {
		t.Fatal(e)
	}
}

func TestM09PersistenceWriterRejectsBeforeIO(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "state.json")
	valid, _ := baseM09()
	if e := PersistM09State(p, valid); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(p)
	bad := valid
	bad.Intent.IntentMode = ""
	if e := PersistM09State(p, bad); e == nil {
		t.Fatal("invalid output written")
	}
	after, _ := os.ReadFile(p)
	if !bytes.Equal(before, after) {
		t.Fatal("overwrote valid snapshot")
	}
	newPath := filepath.Join(dir, "must-not-exist", "state.json")
	if e := PersistM09State(newPath, bad); e == nil {
		t.Fatal("invalid accepted")
	}
	if _, e := os.Stat(filepath.Dir(newPath)); !os.IsNotExist(e) {
		t.Fatal("invalid state created directory")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal("temporary output leaked")
	}
	if s, e := LoadM09State(filepath.Join(dir, "missing")); e == nil || !reflect.DeepEqual(s, M09State{}) {
		t.Fatal("missing file")
	}
}
