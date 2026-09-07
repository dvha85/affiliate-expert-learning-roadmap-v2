package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func evaluationFixture(t *testing.T) ([]string, string) {
	t.Helper()
	dir := t.TempDir()
	if _, err := buildBR10AdvisorFixture(dir); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, "evaluation-config.json")
	raw := `{"evaluation_id":"eval-1","decision_id":"br11-decision","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"br11-action"},"outcome_ids":["br11-outcome"],"evaluated_at":"2026-09-06T00:00:00Z"}`
	if err := os.WriteFile(config, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	return []string{"create", filepath.Join(dir, "history.jsonl"), filepath.Join(dir, "actions.jsonl"), filepath.Join(dir, "outcomes.jsonl"), filepath.Join(dir, "evaluations.jsonl"), config}, raw
}
func evaluationRun(t *testing.T, args []string, want string) map[string]json.RawMessage {
	t.Helper()
	var out, diagnostic bytes.Buffer
	code := runEvaluationStore(args, &out, &diagnostic)
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err, out.String(), diagnostic.String())
	}
	var status string
	_ = json.Unmarshal(envelope["status"], &status)
	if status != want {
		t.Fatal(code, status, want, diagnostic.String())
	}
	if string(envelope["execution_permitted"]) != "false" || string(envelope["auto_apply"]) != "false" {
		t.Fatal(out.String())
	}
	success := want == "APPENDED" || want == "EXACT_DUPLICATE" || want == "VALID"
	if success != (code == 0) {
		t.Fatal(code, status)
	}
	if !success && envelope["artifact"] != nil {
		t.Fatal("error artifact leaked")
	}
	return envelope
}

func TestEvaluationPersistenceAndConflict(t *testing.T) {
	args, raw := evaluationFixture(t)
	originals := map[string][]byte{}
	for _, p := range args[1:4] {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		originals[p] = b
	}
	env := evaluationRun(t, args, "APPENDED")
	if !bytes.Contains(env["artifact"], []byte(`"result":"INCONCLUSIVE"`)) {
		t.Fatal(string(env["artifact"]))
	}
	saved, err := os.ReadFile(args[4])
	if err != nil {
		t.Fatal(err)
	}
	evaluationRun(t, args, "EXACT_DUPLICATE")
	evaluationRun(t, append([]string{"list"}, args[1:5]...), "VALID")
	if err := os.WriteFile(args[5], []byte(strings.Replace(raw, "2026-09-06", "2026-09-07", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	evaluationRun(t, args, "CONFLICT")
	now, err := os.ReadFile(args[4])
	if err != nil || !bytes.Equal(saved, now) {
		t.Fatal("duplicate/conflict changed bytes", err)
	}
	for p, b := range originals {
		now, err := os.ReadFile(p)
		if err != nil || !bytes.Equal(b, now) {
			t.Fatal("upstream changed", p, err)
		}
	}
	// Mutated success claims cannot be accepted on restart.
	if err := os.WriteFile(args[4], bytes.Replace(saved, []byte("INCONCLUSIVE"), []byte("SUPPORTED"), 1), 0600); err != nil {
		t.Fatal(err)
	}
	evaluationRun(t, append([]string{"list"}, args[1:5]...), "STORE_ERROR")
	evaluationRun(t, args, "STORE_ERROR")
}

func TestEvaluationRejectsInvalidChain(t *testing.T) {
	for _, tc := range []struct{ name, old, new, status string }{
		{"decision", `"br11-decision"`, `"missing"`, "EVALUATION_ERROR"},
		{"action", `"br11-action"`, `"missing"`, "EVALUATION_ERROR"},
		{"outcome", `"br11-outcome"`, `"missing"`, "EVALUATION_ERROR"},
		{"early", "2026-09-06", "2026-09-04", "EVALUATION_ERROR"},
		{"machine", "HUMAN_ACTION", "MACHINE_EXECUTION", "EVALUATION_ERROR"},
		{"duplicate", `["br11-outcome"]`, `["br11-outcome","br11-outcome"]`, "EVALUATION_ERROR"},
		{"result-injection", `"evaluation_id":`, `"result":"SUPPORTED","evaluation_id":`, "CONFIG_ERROR"},
		{"duplicate-key", `"evaluation_id":`, `"evaluation_id":"other","evaluation_id":`, "CONFIG_ERROR"},
		{"missing-time", `"evaluated_at":"2026-09-06T00:00:00Z"`, `"evaluated_at":null`, "EVALUATION_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args, raw := evaluationFixture(t)
			if err := os.WriteFile(args[5], []byte(strings.Replace(raw, tc.old, tc.new, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			evaluationRun(t, args, tc.status)
			if _, err := os.Lstat(args[4]); !os.IsNotExist(err) {
				t.Fatal("rejected create wrote store", err)
			}
		})
	}
}

func TestEvaluationZeroAndBrokenUpstream(t *testing.T) {
	args, _ := evaluationFixture(t)
	raw, err := os.ReadFile(args[3])
	if err != nil {
		t.Fatal(err)
	}
	raw = bytes.Replace(raw, []byte(`"status":"PENDING"`), []byte(`"status":"NO_OBSERVED_OUTCOME"`), 1)
	raw = bytes.Replace(raw, []byte(`"metrics":{}`), []byte(`"metrics":{"valid_orders":0}`), 1)
	if err := os.WriteFile(args[3], raw, 0600); err != nil {
		t.Fatal(err)
	}
	env := evaluationRun(t, args, "APPENDED")
	if !bytes.Contains(env["artifact"], []byte(`"result":"INCONCLUSIVE"`)) {
		t.Fatal(string(env["artifact"]))
	}
	if err := os.WriteFile(args[3], nil, 0600); err != nil {
		t.Fatal(err)
	}
	evaluationRun(t, append([]string{"list"}, args[1:5]...), "STORE_ERROR")
}

func TestEvaluationPathGuards(t *testing.T) {
	args, _ := evaluationFixture(t)
	bad := append([]string(nil), args...)
	bad[4] = args[3]
	evaluationRun(t, bad, "PATH_ERROR")
	if err := os.Link(args[3], args[4]); err != nil {
		t.Fatal(err)
	}
	evaluationRun(t, args, "PATH_ERROR")
}

func TestEvaluationStoreFramingAndDuplicate(t *testing.T) {
	for _, kind := range []string{"partial-line", "duplicate-id"} {
		t.Run(kind, func(t *testing.T) {
			args, _ := evaluationFixture(t)
			evaluationRun(t, args, "APPENDED")
			raw, err := os.ReadFile(args[4])
			if err != nil {
				t.Fatal(err)
			}
			if kind == "partial-line" {
				raw = bytes.TrimSuffix(raw, []byte("\n"))
			} else {
				raw = append(raw, raw...)
			}
			if err := os.WriteFile(args[4], raw, 0600); err != nil {
				t.Fatal(err)
			}
			evaluationRun(t, args, "STORE_ERROR")
			evaluationRun(t, append([]string{"list"}, args[1:5]...), "STORE_ERROR")
		})
	}
}
