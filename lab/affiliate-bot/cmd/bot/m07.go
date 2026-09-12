package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
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

func appendRegisteredToolEvidence(ctx m07Context, raw []byte, registry []corem07.ToolSpec) (m07Context, error) {
	registered, err := corem07.ValidateRegisteredToolResult(raw, registry, ctx.RecordID)
	if err != nil {
		return m07Context{}, err
	}
	evidence := registered.Evidence()
	for _, existing := range ctx.Evidence {
		if existing.EvidenceID == evidence.EvidenceID {
			return m07Context{}, fmt.Errorf("registered tool evidence id already exists in context")
		}
	}
	ctx.EvidenceIDs = append(ctx.EvidenceIDs, evidence.EvidenceID)
	ctx.Evidence = append(ctx.Evidence, evidence)
	return ctx, nil
}

func loadM07Registry(path string) ([]corem07.ToolSpec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var registry []corem07.ToolSpec
	// The registry is executable policy, not permissive configuration. Reject
	// duplicate, aliased, unknown, or trailing fields before ValidateRegistry
	// decides whether the requested host and method are allowed.
	if err := contracts.DecodeStrict(raw, &registry); err != nil {
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
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot m07 context HISTORY RECORD_ID | bot m07 register-tool-result HISTORY RECORD_ID REGISTRY TOOL_RESULT OUTPUT | bot m07 register-proposal HISTORY RECORD_ID MODEL_OUTPUT REGISTRY OUTPUT [REGISTERED_TOOL_RESULT] | bot m07 validate HISTORY RECORD_ID MODEL_OUTPUT REGISTRY [REGISTERED_TOOL_RESULT]"), 2)
	}
	if (args[0] == "context" && len(args) != 3) || (args[0] == "register-tool-result" && len(args) != 6) || (args[0] == "register-proposal" && len(args) != 6 && len(args) != 7) || (args[0] == "validate" && len(args) != 5 && len(args) != 6) || (args[0] != "context" && args[0] != "register-tool-result" && args[0] != "register-proposal" && args[0] != "validate") {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot m07 context HISTORY RECORD_ID | bot m07 register-tool-result HISTORY RECORD_ID REGISTRY TOOL_RESULT OUTPUT | bot m07 register-proposal HISTORY RECORD_ID MODEL_OUTPUT REGISTRY OUTPUT [REGISTERED_TOOL_RESULT] | bot m07 validate HISTORY RECORD_ID MODEL_OUTPUT REGISTRY [REGISTERED_TOOL_RESULT]"), 2)
	}
	// Reject every output/input alias before reading any model, registry, tool or
	// history bytes. The output is immutable, but this preflight also makes the
	// safety boundary independent of which input happens to parse successfully.
	switch args[0] {
	case "register-tool-result":
		if err := distinctPaths(args[1], args[3], args[4], args[5]); err != nil {
			return emit("PATH_CONFLICT", nil, err, 1)
		}
	case "register-proposal":
		paths := []string{args[1], args[3], args[4], args[5]}
		if len(args) == 7 {
			paths = append(paths, args[6])
		}
		if err := distinctPaths(paths...); err != nil {
			return emit("PATH_CONFLICT", nil, err, 1)
		}
	}
	// Canonical history is the source of every M07 evidence context. Hold the
	// same local gate as AppendHistory before resolving it, so a watcher cannot
	// append a partially observed record while this command creates, validates,
	// or persists an M07 artifact from that context.
	release, err := acquireHistoryRuntimeGate(args[1])
	if err != nil {
		return emit("BUSY", nil, err, 1)
	}
	defer release()
	record, err := resolveCanonicalRecord(args[1], args[2])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	ctx, err := m07EvidenceContext(record)
	if err != nil {
		return emit("CONTEXT_ERROR", nil, err, 1)
	}
	switch args[0] {
	case "context":
		return emit("VALID", ctx, nil, 0)
	case "register-tool-result":
		registry, err := loadM07Registry(args[3])
		if err != nil {
			return emit("REGISTRY_ERROR", nil, err, 1)
		}
		raw, err := os.ReadFile(args[4])
		if err != nil {
			return emit("TOOL_RESULT_ERROR", nil, err, 1)
		}
		registered, err := corem07.RegisterToolResult(raw, registry)
		if err != nil {
			return emit("TOOL_RESULT_REJECTED", nil, err, 1)
		}
		if registered.RecordID != record.RecordID {
			return emit("TOOL_RESULT_REJECTED", nil, fmt.Errorf("tool result record_id does not match canonical record"), 1)
		}
		status, err := writeNewJSON(args[5], registered)
		if err != nil {
			return emit("PERSISTENCE_ERROR", nil, err, 1)
		}
		return emit(status, map[string]any{"registered": registered, "evidence": registered.Evidence()}, nil, 0)
	case "register-proposal":
		raw, err := os.ReadFile(args[3])
		if err != nil {
			return emit("OUTPUT_ERROR", nil, err, 1)
		}
		text := strings.TrimSpace(string(raw))
		if strings.HasPrefix(text, "```json") && strings.HasSuffix(text, "```") {
			text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "```json"), "```"))
		}
		registry, err := loadM07Registry(args[4])
		if err != nil {
			return emit("REGISTRY_ERROR", nil, err, 1)
		}
		if len(args) == 7 {
			toolRaw, err := os.ReadFile(args[6])
			if err != nil {
				return emit("TOOL_RESULT_ERROR", nil, err, 1)
			}
			ctx, err = appendRegisteredToolEvidence(ctx, toolRaw, registry)
			if err != nil {
				return emit("TOOL_RESULT_REJECTED", nil, err, 1)
			}
		}
		proposal, err := corem07.RegisterAgentProposal([]byte(text), ctx.Evidence, registry, record.RecordID)
		if err != nil {
			return emit("PROPOSAL_REJECTED", nil, err, 1)
		}
		status, err := writeNewJSON(args[5], proposal)
		if err != nil {
			return emit("PERSISTENCE_ERROR", nil, err, 1)
		}
		return emit(status, proposal, nil, 0)
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
		if len(args) == 6 {
			toolRaw, err := os.ReadFile(args[5])
			if err != nil {
				return emit("TOOL_RESULT_ERROR", nil, err, 1)
			}
			ctx, err = appendRegisteredToolEvidence(ctx, toolRaw, registry)
			if err != nil {
				return emit("TOOL_RESULT_REJECTED", nil, err, 1)
			}
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
