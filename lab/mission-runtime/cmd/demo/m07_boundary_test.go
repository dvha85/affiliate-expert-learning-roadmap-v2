package main

import (
	"bytes"
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const m07P = `{"state":"PROPOSE","answer":"fixture","evidence_ids":["e1"],"tool_calls":[{"tool_name":"http","method":"GET","target":"https://example.com/a"}]}`
const m07R = `[{"name":"http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"]}]`

func TestM07RawBoundary(t *testing.T) {
	for _, tc := range []struct{ name, p, r, ids, want string }{
		{"normal", m07P, m07R, `["e1"]`, "SUPPORTED"},
		{"missing_context", m07P, m07R, `[]`, "REJECT_UNGROUNDED"},
		{"null_context", m07P, m07R, `null`, "INVALID_SCHEMA"},
		{"duplicate_context", m07P, m07R, `["e1","e1"]`, "INVALID_CONTEXT"},
		{"empty_registry", m07P, `[]`, `["e1"]`, "INVALID_SCHEMA"},
		{"write_registry", m07P, strings.Replace(m07R, `true`, `false`, 1), `["e1"]`, "INVALID_SCHEMA"},
		{"duplicate_tool", m07P, "[" + strings.Trim(m07R, "[]") + "," + strings.Trim(m07R, "[]") + "]", `["e1"]`, "INVALID_REGISTRY"},
		{"post", strings.Replace(m07P, `GET`, `POST`, 1), m07R, `["e1"]`, "REJECT_TOOL"},
		{"abstain_post", strings.Replace(strings.Replace(m07P, `PROPOSE`, `ABSTAIN`, 1), `GET`, `POST`, 1), m07R, `["e1"]`, "REJECT_TOOL"},
		{"host", strings.Replace(m07P, `example.com/a`, `evil.invalid/a`, 1), m07R, `["e1"]`, "REJECT_TOOL"},
		{"port", strings.Replace(m07P, `example.com/a`, `example.com:8443/a`, 1), m07R, `["e1"]`, "REJECT_TOOL"},
		{"userinfo", strings.Replace(m07P, `example.com/a`, `user@example.com/a`, 1), m07R, `["e1"]`, "REJECT_TOOL"},
		{"duplicate_key", strings.Replace(m07P, `"state":`, `"state":"ABSTAIN","state":`, 1), m07R, `["e1"]`, "INVALID_SCHEMA"},
		{"null_call", strings.Replace(m07P, `[{"tool_name":"http","method":"GET","target":"https://example.com/a"}]`, `[null]`, 1), m07R, `["e1"]`, "INVALID_SCHEMA"},
		{"unknown_field", strings.Replace(m07P, `{`, `{"extra":true,`, 1), m07R, `["e1"]`, "INVALID_SCHEMA"},
		{"null_ids", strings.Replace(m07P, `["e1"]`, `null`, 1), m07R, `["e1"]`, "INVALID_SCHEMA"},
		{"trailing", m07P + ` {}`, m07R, `["e1"]`, "INVALID_SCHEMA"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, s := CheckM07Files([]byte(tc.p), []byte(tc.r), []byte(tc.ids))
			if s != tc.want {
				t.Fatal(s, tc.want)
			}
		})
	}
	for _, source := range []struct {
		raw      string
		registry bool
	}{{m07P, false}, {strings.Trim(m07R, "[]"), true}} {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(source.raw), &fields); err != nil {
			t.Fatal(err)
		}
		for key := range fields {
			for _, mutation := range []string{"missing", "null"} {
				copy := map[string]json.RawMessage{}
				for k, v := range fields {
					copy[k] = v
				}
				if mutation == "missing" {
					delete(copy, key)
				} else {
					copy[key] = json.RawMessage(`null`)
				}
				raw, _ := json.Marshal(copy)
				var s string
				if source.registry {
					_, s = DecodeM07Registry(append(append([]byte("["), raw...), ']'))
				} else {
					_, s = DecodeM07Proposal(raw)
				}
				if s != "INVALID_SCHEMA" {
					t.Fatal(key, mutation, s)
				}
			}
		}
	}
}

func TestM07CLI(t *testing.T) {
	dir := t.TempDir()
	paths := []string{}
	inputs := []string{m07P, m07R, `["e1"]`}
	for i, name := range []string{"proposal", "registry", "ids"} {
		path := filepath.Join(dir, name+".json")
		if err := os.WriteFile(path, []byte(inputs[i]), 0600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	var out bytes.Buffer
	if err := runM07Check(&out, paths); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Proposal   AgentProposal   `json:"proposal"`
		Registry   json.RawMessage `json:"registry"`
		Result     string          `json:"result"`
		Authorized bool            `json:"execution_authorized"`
	}
	if err := contracts.DecodeStrict(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Authorized || envelope.Result != "SUPPORTED" {
		t.Fatal("authority")
	}
	if err := contracts.ValidateRaw("tool-registry.schema.json", envelope.Registry); err != nil {
		t.Fatal(err)
	}
	if err := runM07Check(m03BrokenWriter{}, paths); err == nil {
		t.Fatal("writer failure lost")
	}
	for i, path := range paths {
		raw, _ := os.ReadFile(path)
		if string(raw) != inputs[i] {
			t.Fatal("input changed")
		}
	}
	bad := strings.Replace(m07P, `GET`, `POST`, 1)
	if err := os.WriteFile(paths[0], []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := runM07Check(&out, paths); err == nil || out.Len() != 0 {
		t.Fatal("rejection emitted output")
	}
	out.Reset()
	if err := runM07Check(&out, nil); err == nil {
		t.Fatal("args")
	}
	if err := runM07Check(&out, []string{"../../testdata/m07-proposal.json", "../../testdata/m07-registry.json", "../../testdata/m07-evidence-ids.json"}); err != nil {
		t.Fatal(err)
	}
}

func TestM07RawEval(t *testing.T) {
	type tc struct {
		ID       string          `json:"case_id"`
		Proposal json.RawMessage `json:"proposal"`
		Registry json.RawMessage `json:"registry"`
		IDs      json.RawMessage `json:"available_evidence_ids"`
		Expected string          `json:"expected"`
	}
	for _, c := range loadCases[tc](t, "M07-readonly-evidence-agent") {
		t.Run(c.ID, func(t *testing.T) {
			want := c.Expected
			if c.ID == "M07-E02-unknown-tool" || c.ID == "M07-E03-write-tool-rejected" || c.ID == "M07-E04-hallucinated-evidence" {
				want = "INVALID_SCHEMA"
			}
			_, _, s := CheckM07Files(c.Proposal, c.Registry, c.IDs)
			if s != want {
				t.Fatal(s, want)
			}
		})
	}
}
