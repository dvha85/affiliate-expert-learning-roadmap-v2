package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/canonical"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

type m07Context = canonical.Context

const maxM07PortableInputBytes int64 = 1 << 20

// readM07PortableInput is the boundary for caller-supplied registry, model,
// and tool-result files. They remain untrusted portable input until the M07
// validators accept their content, so a post-open pathname swap must fail
// before a tool evidence artifact or grounded proposal can be persisted.
func readM07PortableInput(path string) ([]byte, error) {
	b, _, err := readStableRegularFileLimit(path, maxM07PortableInputBytes)
	return b, err
}

func m07EvidenceContext(record HistoryRecord) (m07Context, error) {
	inputs := make([]canonical.ObservationInput, 0, len(record.Observations))
	for _, observation := range record.Observations {
		raw, err := json.Marshal(observation)
		if err != nil {
			return m07Context{}, err
		}
		inputs = append(inputs, canonical.ObservationInput{
			Raw: raw, ObservationID: observation.ObservationID, SubjectID: observation.SubjectID,
			ClaimKind: observation.ClaimKind, SourceAuthorityOrRole: observation.SourceAuthorityOrRole,
			Limitation: observation.Limitation, Value: observation,
		})
	}
	return canonical.BuildEvidenceContext(record.RecordID, record.RecordedResult.DecisionID, record.RecordedResult.EvidenceIDs, inputs)
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
	raw, err := readM07PortableInput(path)
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

// m07PersistenceStatus keeps an adapter/CLI acknowledgement fail-closed when
// an immutable artifact became visible but its parent directory sync could not
// be confirmed. A normal persistence error means no ACK; this distinct status
// also tells a caller to resolve or exact-retry the deterministic artifact
// rather than assuming it can safely create a different one.
func m07PersistenceStatus(err error) string {
	var uncertain *immutableArtifactPublishUncertainError
	if errors.As(err, &uncertain) {
		return "PUBLISHED_RECOVERY_REQUIRED"
	}
	return "PERSISTENCE_ERROR"
}

func runM07(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil && m07PersistenceStatus(err) == "PUBLISHED_RECOVERY_REQUIRED" {
			status = m07PersistenceStatus(err)
		}
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
		raw, err := readM07PortableInput(args[4])
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
			artifact := any(nil)
			if m07PersistenceStatus(err) == "PUBLISHED_RECOVERY_REQUIRED" {
				// The immutable trace is already visible and has been validated;
				// disclose its deterministic recovery identity without treating it
				// as an acknowledgement that a downstream caller may consume.
				artifact = map[string]any{"registered": registered, "evidence": registered.Evidence()}
			}
			return emit("PERSISTENCE_ERROR", artifact, err, 1)
		}
		return emit(status, map[string]any{"registered": registered, "evidence": registered.Evidence()}, nil, 0)
	case "register-proposal":
		raw, err := readM07PortableInput(args[3])
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
			toolRaw, err := readM07PortableInput(args[6])
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
			artifact := any(nil)
			if m07PersistenceStatus(err) == "PUBLISHED_RECOVERY_REQUIRED" {
				// Proposal identity is deterministic from the already validated raw
				// output. It supports exact recovery/retry but is never an ACK.
				artifact = proposal
			}
			return emit("PERSISTENCE_ERROR", artifact, err, 1)
		}
		return emit(status, proposal, nil, 0)
	case "validate":
		raw, err := readM07PortableInput(args[3])
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
			toolRaw, err := readM07PortableInput(args[5])
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
