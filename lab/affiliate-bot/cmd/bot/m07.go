package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m00"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

type m07Context struct {
	RecordID    string             `json:"record_id"`
	DecisionID  string             `json:"decision_id"`
	EvidenceIDs []string           `json:"evidence_ids"`
	Evidence    []corem07.Evidence `json:"evidence"`
	Authority   string             `json:"authority"`
}

func m07EvidenceContext(record HistoryRecord) (m07Context, error) {
	ctx := m07Context{RecordID: record.RecordID, DecisionID: record.RecordID, Authority: "canonical_history_store"}
	ctx.EvidenceIDs = append(ctx.EvidenceIDs, record.RecordedResult.EvidenceIDs...)
	for _, observation := range record.Observations {
		ctx.Evidence = append(ctx.Evidence, corem07.Evidence{
			EvidenceID: observation.ObservationID, SubjectID: observation.SubjectID,
			FieldOrClaim: "product snapshot", Value: observation, ClaimKind: observation.ClaimKind,
			SourceAuthorityOrRole: observation.SourceAuthorityOrRole, Limitation: observation.Limitation,
		})
		raw, err := json.Marshal(observation)
		if err != nil {
			return m07Context{}, err
		}
		fields, err := m00.SourceFields(raw)
		if err != nil {
			return m07Context{}, err
		}
		for _, field := range fields {
			ctx.EvidenceIDs = append(ctx.EvidenceIDs, field.ObservationID)
			ctx.Evidence = append(ctx.Evidence, corem07.Evidence{
				EvidenceID: field.ObservationID, SubjectID: field.SubjectID,
				FieldOrClaim: field.Field, Value: field.Value, ClaimKind: field.ClaimKind,
				SourceAuthorityOrRole: field.Role, Limitation: field.Limitation,
			})
		}
	}
	return ctx, nil
}

func loadM07Registry(path string) ([]corem07.ToolSpec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var registry []corem07.ToolSpec
	if err := json.Unmarshal(raw, &registry); err != nil {
		return nil, err
	}
	if err := corem07.ValidateRegistry(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

func runM07(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "m07", "status": status, "execution_permitted": false}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) < 1 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot m07 context HISTORY RECORD_ID | bot m07 validate HISTORY RECORD_ID MODEL_OUTPUT REGISTRY"), 2)
	}
	if (args[0] == "context" && len(args) != 3) || (args[0] == "validate" && len(args) != 5) || (args[0] != "context" && args[0] != "validate") {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot m07 context HISTORY RECORD_ID | bot m07 validate HISTORY RECORD_ID MODEL_OUTPUT REGISTRY"), 2)
	}
	history, err := LoadHistory(args[1])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	var record *HistoryRecord
	matching := 0
	for i := range history {
		if history[i].RecordID == args[2] {
			copy := history[i]
			record = &copy
			matching++
		}
	}
	if record == nil {
		return emit("NOT_FOUND", nil, fmt.Errorf("decision %s not found in canonical history", args[2]), 1)
	}
	if matching != 1 {
		return emit("HISTORY_ERROR", nil, fmt.Errorf("decision %s resolves to %d records", args[2], matching), 1)
	}
	if replay := Replay(*record); replay.State != replayMatch {
		return emit("HISTORY_ERROR", nil, fmt.Errorf("decision %s is not replay-stable: %s", args[2], replay.State), 1)
	}
	ctx, err := m07EvidenceContext(*record)
	if err != nil {
		return emit("CONTEXT_ERROR", nil, err, 1)
	}
	switch args[0] {
	case "context":
		return emit("VALID", ctx, nil, 0)
	case "validate":
		raw, err := os.ReadFile(args[3])
		if err != nil {
			return emit("OUTPUT_ERROR", nil, err, 1)
		}
		// The model adapter may return a fenced JSON object; strip only the
		// explicit fence, never repair or invent fields.
		text := strings.TrimSpace(string(raw))
		if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
			text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "```json"), "```"))
		}
		registry, err := loadM07Registry(args[4])
		if err != nil {
			return emit("REGISTRY_ERROR", nil, err, 1)
		}
		output, err := corem07.ValidateAgentOutput([]byte(text), ctx.Evidence, registry)
		if err != nil {
			return emit("ABSTAIN", nil, err, 1)
		}
		return emit("VALID", output, nil, 0)
	default:
		return emit("USAGE_ERROR", nil, fmt.Errorf("unknown m07 operation %q", args[0]), 2)
	}
}
