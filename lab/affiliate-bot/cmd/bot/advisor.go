package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m04"
)

type advisorConfig struct {
	DecisionID  string `json:"decision_id"`
	Question    string `json:"question"`
	AsOf        string `json:"as_of"`
	MaxAgeHours *int   `json:"max_age_hours"`
}
type advisorContext struct {
	Version    string                `json:"version"`
	Config     advisorConfig         `json:"config"`
	Evidence   []m04.AdvisorEvidence `json:"evidence"`
	Payload    map[string]any        `json:"payload"`
	Limitation string                `json:"limitation"`
}

func buildAdvisorContext(historyPath, actionsPath, outcomesPath string, c advisorConfig) (advisorContext, error) {
	ctx := advisorContext{Version: "br11a-context/v1", Config: c, Payload: map[string]any{}, Evidence: []m04.AdvisorEvidence{}, Limitation: "Local caller-declared evidence, not trusted truth. Mock only; no business recommendation or execution. Outcome snapshots are not additive."}
	if strings.TrimSpace(c.DecisionID) == "" || strings.TrimSpace(c.Question) == "" || c.MaxAgeHours == nil || *c.MaxAgeHours < 0 || *c.MaxAgeHours > 8760 {
		return ctx, fmt.Errorf("decision/question and max_age_hours in [0,8760] required")
	}
	if _, err := time.Parse(time.RFC3339, c.AsOf); err != nil {
		return ctx, err
	}
	history, err := LoadHistory(historyPath)
	if err != nil {
		return ctx, err
	}
	if err := resolveActionDecision(history, c.DecisionID); err != nil {
		return ctx, err
	}
	actions, err := loadActions(actionsPath, history)
	if err != nil {
		return ctx, err
	}
	outcomes, err := loadOutcomes(outcomesPath, actions)
	if err != nil {
		return ctx, err
	}
	add := func(id, at, source string, payload any) error {
		if _, ok := ctx.Payload[id]; ok {
			return fmt.Errorf("ambiguous cross-kind evidence_id %q", id)
		}
		ctx.Payload[id] = payload
		ctx.Evidence = append(ctx.Evidence, m04.AdvisorEvidence{EvidenceID: id, ObservedAt: at, SourceRef: source})
		return nil
	}
	for _, r := range history {
		if r.RecordedResult.DecisionID == c.DecisionID {
			if err := add(r.RecordID, r.AsOf, "history:"+r.RecordID, r); err != nil {
				return ctx, err
			}
			for _, o := range r.Observations {
				source := o.SourceRef
				if source == "" {
					source = o.SourceURL
				}
				if err := add(o.ObservationID, o.ObservedAt, source, o); err != nil {
					return ctx, err
				}
				raw, _ := json.Marshal(o)
				fields, err := m00.SourceFields(raw)
				if err != nil {
					return ctx, err
				}
				for _, f := range fields {
					source := f.SourceRef
					if source == "" {
						source = f.SourceURL
					}
					if err := add(f.ObservationID, f.ObservedAt, source, f); err != nil {
						return ctx, err
					}
				}
			}
		}
	}
	selected := map[string]bool{}
	for _, a := range actions {
		if a.DecisionID == c.DecisionID {
			selected[a.ActionID] = true
			if err := add(a.ActionID, a.PerformedAt, "human-record:"+a.ActionID, a); err != nil {
				return ctx, err
			}
		}
	}
	for _, o := range outcomes {
		if selected[o.EffectRef.EffectID] {
			if err := add(o.OutcomeID, o.ObservedAt, o.SourceRef, o); err != nil {
				return ctx, err
			}
		}
	}
	return ctx, nil
}

// Pure deterministic mock: it requests human inspection, never chooses offers.
func mockAdvisor(ctx advisorContext) []byte {
	ids := []string{}
	for _, e := range ctx.Evidence {
		ids = append(ids, e.EvidenceID)
	}
	output := m04.AdvisorOutput{State: "HUMAN_REVIEW", Recommendation: "Kiểm tra thủ công bằng chứng và các snapshot trước khi kết luận.", Reason: "Mock offline chỉ kiểm đường context và liên kết, không đánh giá hiệu quả affiliate.", EvidenceIDs: ids, Unknowns: []string{"Nguồn và thời gian chưa được xác thực; chưa có live provider."}, WriteToolRequested: false}
	raw, _ := json.Marshal(output)
	return raw
}

func checkAdvisorResponse(raw []byte, ctx advisorContext) (m04.AdvisorOutput, string) {
	output, status := m04.DecodeAdvisorOutput(raw)
	if status != "VALID" {
		return output, status
	}
	// Check every context item, even if the response omits stale evidence or abstains.
	all := output
	all.State = "HUMAN_REVIEW"
	all.EvidenceIDs = []string{}
	for _, e := range ctx.Evidence {
		all.EvidenceIDs = append(all.EvidenceIDs, e.EvidenceID)
	}
	if status := m04.EvaluateAdvisorOutput(all, ctx.Evidence, ctx.Config.AsOf, *ctx.Config.MaxAgeHours); status != "SUPPORTED" {
		return output, status
	}
	status = m04.EvaluateAdvisorOutput(output, ctx.Evidence, ctx.Config.AsOf, *ctx.Config.MaxAgeHours)
	if status == "SUPPORTED" || status == "ABSTAIN" {
		if err := contracts.ValidateRaw("advisor-output.schema.json", raw); err != nil {
			return output, "INVALID_SCHEMA"
		}
	}
	return output, status
}

func runAdvisor(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "fixture-run" {
		return runAdvisorFixture(args, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "campaign-init" {
		return runCampaignInitCLI(args, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "campaign-report" {
		return runCampaignReportCLI(args, stdout, stderr)
	}
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "advisor mock", "provider": "mock/v1", "status": status, "execution_permitted": false}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 5 || args[0] != "mock" {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot advisor mock HISTORY ACTIONS OUTCOMES CONFIG.json"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	raw, err := os.ReadFile(args[4])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	var config advisorConfig
	if err := contracts.DecodeStrict(raw, &config); err != nil {
		return emit("CONFIG_ERROR", nil, err, 1)
	}
	ctx, err := buildAdvisorContext(args[1], args[2], args[3], config)
	if err != nil {
		return emit("CONTEXT_ERROR", nil, err, 1)
	}
	provider := mockAdvisorProvider{}
	output, status := evaluateAdvisorProvider(context.Background(), provider, ctx)
	if status != "SUPPORTED" && status != "ABSTAIN" {
		return emit(status, nil, fmt.Errorf("advisor rejected: %s", status), 1)
	}
	return emit(status, map[string]any{"advisor_output": output, "context": ctx, "provenance": provider.identity()}, nil, 0)
}
