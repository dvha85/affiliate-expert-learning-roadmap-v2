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

const validAdvisorJSON = `{"state":"ADVISE","reason":"weak ranking","evidence_ids":["e1"],"unknowns":[],"write_tool_requested":false}`

func TestAdvisorSchemaContractMatchesBoundary(t *testing.T) {
	path := filepath.Join(filepath.Dir(evalPath(t, "M04-grounded-ai-advisor")), "..", "..", "contracts", "advisor-output.schema.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var actual, expected map[string]any
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatal(err)
	}
	const contract = `{"type":"object","required":["state","reason","evidence_ids","unknowns","write_tool_requested"],"properties":{"state":{"enum":["ADVISE","HUMAN_REVIEW","ABSTAIN"]},"recommendation":{"type":"string"},"reason":{"type":"string","minLength":1},"evidence_ids":{"type":"array","items":{"type":"string","minLength":1},"uniqueItems":true},"unknowns":{"type":"array","items":{"type":"string"}},"write_tool_requested":{"const":false}},"additionalProperties":false}`
	if err := json.Unmarshal([]byte(contract), &expected); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"$schema", "$id", "title", "description"} {
		delete(actual, key)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("AdvisorOutput schema changed: review raw/typed boundary and mutation tests before updating this contract snapshot")
	}
}

func TestAdvisorRejectsEmptyReason(t *testing.T) {
	o := AdvisorOutput{State: "ADVISE", EvidenceIDs: []string{"e1"}, Unknowns: []string{}}
	ev := []AdvisorEvidence{{EvidenceID: "e1", ObservedAt: "2026-09-03T00:00:00Z", SourceRef: "synthetic"}}
	if got := EvaluateAdvisorOutput(o, ev, "2026-09-03T01:00:00Z", 24); got != "INVALID_SCHEMA" {
		t.Fatalf("empty reason must fail before SUPPORTED: got %s", got)
	}
}

func TestAdvisorJSONBoundary(t *testing.T) {
	cases := []struct{ name, raw, want string }{
		{"valid", validAdvisorJSON, missionValid},
		{"reason_empty", strings.Replace(validAdvisorJSON, "weak ranking", "", 1), "INVALID_SCHEMA"},
		{"reason_whitespace", strings.Replace(validAdvisorJSON, "weak ranking", "   ", 1), "INVALID_SCHEMA"},
		{"unknown_field", strings.Replace(validAdvisorJSON, `"state":`, `"extra":true,"state":`, 1), "INVALID_SCHEMA"},
		{"case_sensitive", strings.Replace(validAdvisorJSON, `"state":`, `"State":`, 1), "INVALID_SCHEMA"},
		{"duplicate_key", strings.Replace(validAdvisorJSON, `"state":`, `"state":"ABSTAIN","state":`, 1), "INVALID_SCHEMA"},
		{"duplicate_ids", strings.Replace(validAdvisorJSON, `["e1"]`, `["e1","e1"]`, 1), "INVALID_SCHEMA"},
		{"empty_id", strings.Replace(validAdvisorJSON, `["e1"]`, `[""]`, 1), "INVALID_SCHEMA"},
		{"wrong_enum", strings.Replace(validAdvisorJSON, "ADVISE", "SUPPORTED", 1), "INVALID_SCHEMA"},
		{"null_item", strings.Replace(validAdvisorJSON, `"unknowns":[]`, `"unknowns":[null]`, 1), "INVALID_SCHEMA"},
		{"number_item", strings.Replace(validAdvisorJSON, `"unknowns":[]`, `"unknowns":[3]`, 1), "INVALID_SCHEMA"},
		{"wrong_array_type", strings.Replace(validAdvisorJSON, `"unknowns":[]`, `"unknowns":{}`, 1), "INVALID_SCHEMA"},
		{"write_forbidden", strings.Replace(validAdvisorJSON, "false", "true", 1), "REJECT_WRITE_REQUEST"},
		{"wrong_bool_type", strings.Replace(validAdvisorJSON, "false", `"false"`, 1), "INVALID_SCHEMA"},
		{"null_recommendation", strings.Replace(validAdvisorJSON, `"state":`, `"recommendation":null,"state":`, 1), "INVALID_SCHEMA"},
		{"trailing_object", validAdvisorJSON + `{}`, "INVALID_SCHEMA"},
		{"truncated", validAdvisorJSON[:20], "INVALID_SCHEMA"},
		{"array_root", `[]`, "INVALID_SCHEMA"},
		{"null_root", `null`, "INVALID_SCHEMA"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, got := DecodeAdvisorOutput([]byte(tc.raw))
			if got != tc.want {
				t.Fatalf("want %s got %s", tc.want, got)
			}
		})
	}
	for _, field := range []string{"state", "reason", "evidence_ids", "unknowns", "write_tool_requested"} {
		for _, mode := range []string{"missing", "null"} {
			t.Run(field+"_"+mode, func(t *testing.T) {
				var object map[string]any
				if err := json.Unmarshal([]byte(validAdvisorJSON), &object); err != nil {
					t.Fatal(err)
				}
				if mode == "missing" {
					delete(object, field)
				} else {
					object[field] = nil
				}
				raw, err := json.Marshal(object)
				if err != nil {
					t.Fatal(err)
				}
				if _, got := DecodeAdvisorOutput(raw); got != "INVALID_SCHEMA" {
					t.Fatalf("got %s", got)
				}
			})
		}
	}
}

func TestAdvisorCheckSerializesValidatedArtifact(t *testing.T) {
	dir := t.TempDir()
	outputPath, evidencePath := filepath.Join(dir, "output.json"), filepath.Join(dir, "evidence.json")
	if err := os.WriteFile(outputPath, []byte(validAdvisorJSON), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidencePath, []byte(`[{"evidence_id":"e1","observed_at":"2026-09-03T00:00:00Z","source_ref":"synthetic"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	args := []string{outputPath, evidencePath, "2026-09-03T01:00:00Z", "24"}
	if err := runAdvisorCheck(&out, args); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Output json.RawMessage `json:"advisor_output"`
		Result string          `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Result != "SUPPORTED" {
		t.Fatalf("got %s", envelope.Result)
	}
	if _, state := DecodeAdvisorOutput(envelope.Output); state != missionValid {
		t.Fatalf("emitted artifact is invalid: %s", state)
	}
	if err := os.WriteFile(outputPath, []byte(strings.Replace(validAdvisorJSON, "weak ranking", "", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runAdvisorCheck(&out, args); err == nil || out.Len() != 0 {
		t.Fatal("malformed output must fail without emitting a success envelope")
	}
}

func TestAdvisorStructuralChecksDoNotReplaceSemantics(t *testing.T) {
	o, state := DecodeAdvisorOutput([]byte(validAdvisorJSON))
	if state != missionValid {
		t.Fatal(state)
	}
	cases := []struct{ name, at, source, want string }{
		{"fresh", "2026-09-03T00:00:00Z", "synthetic", "SUPPORTED"},
		{"stale", "2026-08-01T00:00:00Z", "synthetic", "ABSTAIN_STALE"},
		{"future", "2026-09-04T00:00:00Z", "synthetic", "ABSTAIN_FUTURE"},
		{"bad_time", "invalid", "synthetic", missionInvalid},
		{"empty_source", "2026-09-03T00:00:00Z", "", missionInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev := []AdvisorEvidence{{EvidenceID: "e1", ObservedAt: tc.at, SourceRef: tc.source}}
			if got := EvaluateAdvisorOutput(o, ev, "2026-09-03T01:00:00Z", 24); got != tc.want {
				t.Fatalf("want %s got %s", tc.want, got)
			}
		})
	}
	ev := []AdvisorEvidence{{EvidenceID: "e1", ObservedAt: "2026-09-03T00:00:00Z", SourceRef: "synthetic"}}
	if got := EvaluateAdvisorOutput(o, append(ev, ev[0]), "2026-09-03T01:00:00Z", 24); got != missionInvalid {
		t.Fatal("duplicate evidence silently overwrote source")
	}
	if got := EvaluateAdvisorOutput(o, nil, "2026-09-03T01:00:00Z", 24); got != "REJECT_UNGROUNDED" {
		t.Fatal(got)
	}
}
