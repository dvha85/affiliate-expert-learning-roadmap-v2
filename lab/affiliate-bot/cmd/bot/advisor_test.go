package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAdvisorConfigStrictAndMockReadOnly(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "h")
	actions := filepath.Join(dir, "a")
	outcomes := filepath.Join(dir, "o")
	obs, e := loadHistoryObservations("../../data/m02-sample-observations.json")
	if e != nil {
		t.Fatal(e)
	}
	r, e := NewHistoryRecord("d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", obs)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AppendHistory(history, r); e != nil {
		t.Fatal(e)
	}
	action := m03.HumanActionRecord{ActionID: "a", DecisionID: "d", ActionType: "synthetic", Target: "fixture", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	raw, _ := json.Marshal(action)
	ap := filepath.Join(dir, "action")
	if e = os.WriteFile(ap, raw, 0600); e != nil {
		t.Fatal(e)
	}
	if runActionStore([]string{"record", history, actions, ap}, io.Discard, io.Discard) != 0 {
		t.Fatal("action")
	}
	op := filepath.Join(dir, "outcome")
	o := m03.OutcomeRecord{OutcomeID: "o", EffectRef: m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: "a"}, ObservedAt: "2026-09-05T00:00:00Z", Status: "PENDING", Metrics: map[string]float64{}, SourceRef: "fixture"}
	raw, _ = json.Marshal(o)
	if e = os.WriteFile(op, raw, 0600); e != nil {
		t.Fatal(e)
	}
	if runOutcomeStore([]string{"import", history, actions, outcomes, op}, io.Discard, io.Discard) != 0 {
		t.Fatal("outcome")
	}
	cfg := advisorConfig{DecisionID: "d", Question: "q", AsOf: "2026-09-06T00:00:00Z", MaxAgeHours: ptr(8760)}
	ctx, e := buildAdvisorContext(history, actions, outcomes, cfg)
	if e != nil {
		t.Fatal(e)
	}
	if status := func() string { _, s := checkAdvisorResponse(mockAdvisor(ctx), ctx); return s }(); status != "SUPPORTED" {
		t.Fatal(status)
	}
	t.Run("DeepSeek_BR10_chain", func(t *testing.T) {
		t.Setenv("DEEPSEEK_API_KEY", "synthetic-test-key-only")
		paths := []string{history, actions, outcomes}
		before := make([][]byte, len(paths))
		for i, path := range paths {
			var err error
			before[i], err = os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, tc := range []struct {
			name, content, want string
			age                 int
			asOf                string
			calls               int
		}{
			{"valid", `{"state":"HUMAN_REVIEW","recommendation":"Kiểm tra thủ công","reason":"Snapshot đang pending, chưa kết luận hiệu quả","evidence_ids":["d","a","o"],"unknowns":["Chưa xác thực nguồn"],"write_tool_requested":false}`, "SUPPORTED", 8760, cfg.AsOf, 1},
			{"orphan", `{"state":"HUMAN_REVIEW","reason":"review","evidence_ids":["invented"],"unknowns":[],"write_tool_requested":false}`, "REJECT_UNGROUNDED", 8760, cfg.AsOf, 1},
			{"empty reason", `{"state":"ABSTAIN","reason":"","evidence_ids":[],"unknowns":[],"write_tool_requested":false}`, "INVALID_SCHEMA", 8760, cfg.AsOf, 1},
			{"malformed", `{`, "INVALID_SCHEMA", 8760, cfg.AsOf, 1},
			{"write", `{"state":"ABSTAIN","reason":"review","evidence_ids":[],"unknowns":[],"write_tool_requested":true}`, "REJECT_WRITE_REQUEST", 8760, cfg.AsOf, 1},
			{"stale", `{}`, "ABSTAIN_STALE", 1, cfg.AsOf, 0},
			{"future", `{}`, "ABSTAIN_FUTURE", 8760, "2026-09-02T00:00:00Z", 0},
		} {
			t.Run(tc.name, func(t *testing.T) {
				calls := 0
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					var request struct {
						Messages []struct{ Role, Content string }
					}
					if err := json.NewDecoder(r.Body).Decode(&request); err != nil || len(request.Messages) != 2 {
						t.Error("request envelope")
						w.WriteHeader(400)
						return
					}
					var sent advisorContext
					if err := json.Unmarshal([]byte(request.Messages[1].Content), &sent); err != nil {
						t.Error(err)
					}
					for _, id := range []string{"d", "obs-a-1", "obs-b-1", "a", "o"} {
						if _, ok := sent.Payload[id]; !ok {
							t.Errorf("missing upstream payload %s", id)
						}
					}
					if sent.Config.DecisionID != "d" || sent.Version != "br11a-context/v1" {
						t.Error("context lineage")
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"model": deepSeekModel, "choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": tc.content}}}, "usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}})
				}))
				defer s.Close()
				p := newDeepSeekProvider()
				p.endpoint = s.URL
				c := ctx
				c.Config.MaxAgeHours = ptr(tc.age)
				c.Config.AsOf = tc.asOf
				out, status := evaluateAdvisorProvider(context.Background(), p, c)
				if status != tc.want || calls != tc.calls {
					t.Fatal(status, calls)
				}
				if status != "SUPPORTED" && out.Recommendation != "" {
					t.Fatal("rejected artifact escaped")
				}
			})
		}
		for i, path := range paths {
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before[i], after) {
				t.Fatal("upstream changed", path, err)
			}
		}
	})
	var bad advisorConfig
	if e := contracts.DecodeStrict([]byte(`{"decision_id":"d","question":"q","as_of":"x","max_age_hours":1,"extra":true}`), &bad); e == nil {
		t.Fatal("unknown config accepted")
	}
}
func ptr(v int) *int { return &v }
