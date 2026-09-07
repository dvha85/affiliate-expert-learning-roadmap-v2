package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const proposalInput = `{"proposal_id":"p1","evaluation_ids":["eval-1"],"current_version":"v1","proposed_version":"v2","change_summary":"Làm rõ pending","expected_benefit":"Tránh kết luận sớm","risks":["Chưa kiểm chứng"],"rollback":"Khôi phục v1 thủ công","auto_apply":false}`
const reviewInput = `{"review_id":"r1","proposal_id":"p1","reviewed_by":"human","reviewed_at":"2026-09-07T00:00:00Z","decision":"APPROVE_FOR_MANUAL_CHANGE","reason":"Chỉ đồng ý thay đổi thủ công có test"}`

func improvementFixture(t *testing.T) ([]string, []string) {
	t.Helper()
	e, _ := evaluationFixture(t)
	evaluationRun(t, e, "APPENDED")
	dir := filepath.Dir(e[4])
	p := append([]string{"import"}, e[1:5]...)
	p = append(p, filepath.Join(dir, "proposals.jsonl"), filepath.Join(dir, "proposal.json"))
	r := append([]string{"import"}, p[1:6]...)
	r = append(r, filepath.Join(dir, "reviews.jsonl"), filepath.Join(dir, "review.json"))
	for path, raw := range map[string]string{p[6]: proposalInput, r[7]: reviewInput} {
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return p, r
}
func improvementRun(t *testing.T, kind string, args []string, want string) {
	t.Helper()
	var out, diagnostic bytes.Buffer
	code := runImprovementStore(kind, args, &out, &diagnostic)
	var env map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatal(err, out.String())
	}
	var status string
	_ = json.Unmarshal(env["status"], &status)
	if status != want {
		t.Fatal(code, status, want, diagnostic.String())
	}
	if string(env["auto_apply"]) != "false" || string(env["execution_permitted"]) != "false" {
		t.Fatal(out.String())
	}
	success := want == "APPENDED" || want == "EXACT_DUPLICATE" || want == "VALID"
	if success != (code == 0) || (!success && env["artifact"] != nil) {
		t.Fatal(code, out.String())
	}
}

func TestImprovementPersistenceAndAuthority(t *testing.T) {
	p, r := improvementFixture(t)
	original := map[string][]byte{}
	for _, path := range p[1:5] {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		original[path] = b
	}
	for _, item := range []struct {
		kind string
		args []string
	}{{"proposal", p}, {"review", r}} {
		improvementRun(t, item.kind, item.args, "APPENDED")
		path := item.args[len(item.args)-2]
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		improvementRun(t, item.kind, item.args, "EXACT_DUPLICATE")
		improvementRun(t, item.kind, append([]string{"list"}, item.args[1:len(item.args)-1]...), "VALID")
		now, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(raw, now) {
			t.Fatal("duplicate rewrote store", err)
		}
	}
	if err := os.WriteFile(r[7], []byte(strings.Replace(reviewInput, "APPROVE_FOR_MANUAL_CHANGE", "REJECT", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	improvementRun(t, "review", r, "CONFLICT")
	if err := os.WriteFile(p[6], []byte(strings.Replace(proposalInput, "v2", "v3", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	improvementRun(t, "proposal", p, "CONFLICT")
	for path, raw := range original {
		now, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(raw, now) {
			t.Fatal("upstream mutated", path, err)
		}
	}
}

func TestImprovementRejectsBadInput(t *testing.T) {
	for _, tc := range []struct{ kind, old, new string }{
		{"proposal", "eval-1", "orphan"}, {"proposal", `["eval-1"]`, `["eval-1","eval-1"]`},
		{"proposal", "false", "true"}, {"proposal", `"v2"`, `"v1"`}, {"proposal", `"v2"`, `" v1 "`},
		{"proposal", `["Chưa kiểm chứng"]`, `[]`}, {"proposal", "Khôi phục v1 thủ công", " "},
		{"proposal", `"proposal_id":`, `"proposal_id":"other","proposal_id":`},
		{"review", "p1", "orphan"}, {"review", "human", "agent"}, {"review", "2026-09-07", "2026-09-05"},
		{"review", "Chỉ đồng ý thay đổi thủ công có test", " "}, {"review", "APPROVE_FOR_MANUAL_CHANGE", "EXECUTE"},
		{"review", `"review_id":`, `"extra":true,"review_id":`},
	} {
		t.Run(tc.kind+tc.old+tc.new, func(t *testing.T) {
			p, r := improvementFixture(t)
			args, raw := p, proposalInput
			if tc.kind == "review" {
				improvementRun(t, "proposal", p, "APPENDED")
				args, raw = r, reviewInput
			}
			if err := os.WriteFile(args[len(args)-1], []byte(strings.Replace(raw, tc.old, tc.new, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			improvementRun(t, tc.kind, args, "INVALID_RECORD")
			if _, err := os.Lstat(args[len(args)-2]); !os.IsNotExist(err) {
				t.Fatal("invalid input created store", err)
			}
		})
	}
}

func TestImprovementCorruptionAndOrphans(t *testing.T) {
	for _, kind := range []string{"proposal", "review"} {
		for _, mutation := range []string{"duplicate", "partial", "orphan"} {
			t.Run(kind+mutation, func(t *testing.T) {
				p, r := improvementFixture(t)
				improvementRun(t, "proposal", p, "APPENDED")
				improvementRun(t, "review", r, "APPENDED")
				path := p[5]
				if kind == "review" {
					path = r[6]
				}
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				switch mutation {
				case "duplicate":
					raw = append(raw, raw...)
				case "partial":
					raw = bytes.TrimSuffix(raw, []byte("\n"))
				case "orphan":
					if kind == "proposal" {
						raw = bytes.Replace(raw, []byte("eval-1"), []byte("missing"), 1)
					} else {
						raw = bytes.Replace(raw, []byte("p1"), []byte("missing"), 1)
					}
				}
				if err := os.WriteFile(path, raw, 0600); err != nil {
					t.Fatal(err)
				}
				status := "PROPOSAL_STORE_ERROR"
				if kind == "review" {
					status = "REVIEW_STORE_ERROR"
				}
				improvementRun(t, "review", append([]string{"list"}, r[1:7]...), status)
				improvementRun(t, "review", r, status)
			})
		}
	}
}

func TestImprovementAliasesAndUpstream(t *testing.T) {
	p, r := improvementFixture(t)
	bad := append([]string(nil), p...)
	bad[5] = p[4]
	improvementRun(t, "proposal", bad, "PATH_ERROR")
	improvementRun(t, "proposal", p, "APPENDED")
	improvementRun(t, "review", r, "APPENDED")
	if err := os.WriteFile(p[4], nil, 0600); err != nil {
		t.Fatal(err)
	}
	improvementRun(t, "review", append([]string{"list"}, r[1:7]...), "PROPOSAL_STORE_ERROR")
}
