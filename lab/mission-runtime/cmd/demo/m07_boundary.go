package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	"io"
	"os"
	"strings"
)

func uniqueM07(values []string) bool {
	seen := map[string]bool{}
	for _, v := range values {
		if strings.TrimSpace(v) == "" || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

func DecodeM07Registry(raw []byte) ([]ToolSpec, string) {
	var registry []ToolSpec
	if contracts.ValidateRaw("tool-registry.schema.json", raw) != nil || contracts.DecodeStrict(raw, &registry) != nil {
		return nil, "INVALID_SCHEMA"
	}
	if len(registry) == 0 {
		return nil, "INVALID_SCHEMA"
	}
	if err := corem07.ValidateRegistry(registry); err != nil {
		return nil, "INVALID_REGISTRY"
	}
	return registry, missionValid
}

// DecodeM07Proposal is the mission-runtime view of the shared M07 output
// decoder. It intentionally has the same required shape as the learner Bot.
func DecodeM07Proposal(raw []byte) (AgentProposal, string) {
	p, err := corem07.DecodeAgentOutput(raw)
	if err != nil {
		return AgentProposal{}, "INVALID_SCHEMA"
	}
	return p, missionValid
}

func CheckM07Files(proposalRaw, registryRaw, idsRaw []byte) (AgentProposal, []ToolSpec, string) {
	registry, state := DecodeM07Registry(registryRaw)
	if state != missionValid {
		return AgentProposal{}, nil, state
	}
	p, state := DecodeM07Proposal(proposalRaw)
	if state != missionValid {
		return AgentProposal{}, nil, state
	}
	value, err := contracts.Decode(idsRaw)
	if err != nil {
		return AgentProposal{}, nil, "INVALID_SCHEMA"
	}
	array, ok := value.([]any)
	if !ok {
		return AgentProposal{}, nil, "INVALID_SCHEMA"
	}
	ids := []string{}
	for _, v := range array {
		id, ok := v.(string)
		if !ok {
			return AgentProposal{}, nil, "INVALID_SCHEMA"
		}
		ids = append(ids, id)
	}
	if !uniqueM07(ids) {
		return AgentProposal{}, nil, "INVALID_CONTEXT"
	}
	available := map[string]bool{}
	for _, id := range ids {
		available[id] = true
	}
	for _, id := range p.EvidenceIDs {
		if !available[id] {
			return AgentProposal{}, nil, "REJECT_UNGROUNDED"
		}
	}
	for _, call := range p.ToolCalls {
		if err := corem07.ValidateToolRequest(call, registry); err != nil {
			return AgentProposal{}, nil, "REJECT_TOOL"
		}
		// This command has no registered adapter trace, so a model-declared
		// request cannot be handed off as grounded output.
		return AgentProposal{}, nil, "REJECT_TOOL"
	}
	state = EvaluateAgentProposal(p, registry, ids)
	return p, registry, state
}

func runM07Check(w io.Writer, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: demo m07-check PROPOSAL.json REGISTRY.json EVIDENCE-IDS.json")
	}
	raw := make([][]byte, 3)
	for i, path := range args {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw[i] = b
	}
	p, registry, state := CheckM07Files(raw[0], raw[1], raw[2])
	if state != "SUPPORTED" && state != "ABSTAIN" {
		return fmt.Errorf("M07 check: %s", state)
	}
	pRaw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if _, s := DecodeM07Proposal(pRaw); s != missionValid {
		return fmt.Errorf("M07 output: %s", s)
	}
	rRaw, err := json.Marshal(registry)
	if err != nil {
		return err
	}
	if _, s := DecodeM07Registry(rRaw); s != missionValid {
		return fmt.Errorf("M07 registry output: %s", s)
	}
	return json.NewEncoder(w).Encode(struct {
		Proposal            AgentProposal `json:"proposal"`
		Registry            []ToolSpec    `json:"registry"`
		Result              string        `json:"result"`
		ExecutionAuthorized bool          `json:"execution_authorized"`
	}{p, registry, state, false})
}
