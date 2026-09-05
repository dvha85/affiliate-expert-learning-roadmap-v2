package main

import (
	"encoding/json"
	"fmt"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"io"
	"net/url"
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
	names := []string{}
	for _, tool := range registry {
		names = append(names, tool.Name)
		normalized := []string{}
		for _, host := range tool.AllowedHosts {
			u, err := url.Parse("https://" + host)
			if err != nil || u.Hostname() == "" || u.Host != host || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.Port() != "" || strings.TrimSpace(host) != host {
				return nil, "INVALID_REGISTRY"
			}
			normalized = append(normalized, strings.ToLower(host))
		}
		if !uniqueM07(normalized) {
			return nil, "INVALID_REGISTRY"
		}
	}
	if !uniqueM07(names) {
		return nil, "INVALID_REGISTRY"
	}
	return registry, missionValid
}

// AgentProposal has no canonical schema. This explicit local shape is not
// AdvisorOutput and must not be advertised as canonical-schema conformance.
func DecodeM07Proposal(raw []byte) (AgentProposal, string) {
	var p AgentProposal
	v, err := contracts.Decode(raw)
	if err != nil {
		return p, "INVALID_SCHEMA"
	}
	object, ok := v.(map[string]any)
	if !ok {
		return p, "INVALID_SCHEMA"
	}
	for _, key := range []string{"state", "answer", "evidence_ids", "tool_calls"} {
		if object[key] == nil {
			return p, "INVALID_SCHEMA"
		}
	}
	ids, ok := object["evidence_ids"].([]any)
	if !ok {
		return p, "INVALID_SCHEMA"
	}
	for _, id := range ids {
		if _, ok := id.(string); !ok {
			return p, "INVALID_SCHEMA"
		}
	}
	calls, ok := object["tool_calls"].([]any)
	if !ok {
		return p, "INVALID_SCHEMA"
	}
	for _, call := range calls {
		item, ok := call.(map[string]any)
		if !ok {
			return p, "INVALID_SCHEMA"
		}
		for _, key := range []string{"tool_name", "method", "target"} {
			value, ok := item[key].(string)
			if !ok || strings.TrimSpace(value) == "" {
				return p, "INVALID_SCHEMA"
			}
		}
	}
	if contracts.DecodeStrict(raw, &p) != nil {
		return AgentProposal{}, "INVALID_SCHEMA"
	}
	if (p.State != "ABSTAIN" && p.State != "PROPOSE" && p.State != "HUMAN_REVIEW") || strings.TrimSpace(p.Answer) == "" || !uniqueM07(p.EvidenceIDs) {
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
	// Enforce calls before the legacy ABSTAIN early return, without inventing IDs.
	tools := map[string]ToolSpec{}
	for _, tool := range registry {
		tools[tool.Name] = tool
	}
	for _, call := range p.ToolCalls {
		tool, exists := tools[call.ToolName]
		if !exists {
			return AgentProposal{}, nil, "REJECT_TOOL"
		}
		method := strings.ToUpper(strings.TrimSpace(call.Method))
		allowed := false
		for _, m := range tool.AllowedMethods {
			if method == m {
				allowed = true
			}
		}
		u, err := url.Parse(call.Target)
		if !allowed || err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Port() != "" {
			return AgentProposal{}, nil, "REJECT_TOOL"
		}
		hostOK := false
		for _, host := range tool.AllowedHosts {
			if strings.EqualFold(host, u.Hostname()) {
				hostOK = true
			}
		}
		if !hostOK {
			return AgentProposal{}, nil, "REJECT_TOOL"
		}
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
