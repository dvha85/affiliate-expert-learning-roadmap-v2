package main

import (
	"encoding/json"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

type WatchRequest = m06.WatchRequest
type CanonicalObservation = m06.CanonicalObservation

func contentHash(body string) string             { return m06.ContentHash(body) }
func EvaluateWatchRequest(r WatchRequest) string { return m06.EvaluateWatchRequest(r) }
func NormalizeWatchObservation(r WatchRequest, subjectID string) (CanonicalObservation, string) {
	return m06.NormalizeWatchObservation(r, subjectID)
}

type ToolSpec = corem07.ToolSpec
type AgentToolCall = corem07.ToolRequest
type AgentProposal = corem07.AgentOutput

func demoM07Proposal() AgentProposal {
	claim := corem07.Claim{FieldOrClaim: "fixture", Value: json.RawMessage(`"fixture"`), EvidenceIDs: []string{"e1"}}
	claim.Text = `fixture="fixture" [evidence:e1]`
	return AgentProposal{
		State: "HUMAN_REVIEW", Answer: claim.Text, Claims: []corem07.Claim{claim},
		EvidenceIDs: []string{"e1"}, ToolCalls: []AgentToolCall{}, Authority: "A2-RO", WritePermission: false,
	}
}

func EvaluateAgentProposal(p AgentProposal, registry []ToolSpec, ids []string, untrustedToolText ...string) string {
	_ = untrustedToolText // tool/page text is data only; never interpreted as permission or policy here.
	if p.State == "ABSTAIN" && len(p.ToolCalls) == 0 {
		return "ABSTAIN"
	}
	for _, call := range p.ToolCalls {
		if err := corem07.ValidateToolRequest(call, registry); err != nil {
			return "REJECT_TOOL"
		}
		// A valid request still needs an adapter-registered trace before model
		// grounding. This harness has only the portable request, not a trace.
		return "REJECT_TOOL"
	}
	evidence := make([]corem07.Evidence, 0, len(ids))
	for _, id := range ids {
		evidence = append(evidence, corem07.Evidence{EvidenceID: id, FieldOrClaim: "fixture", Value: "fixture", ClaimKind: "unknown", Limitation: "fixture context"})
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return missionInvalid
	}
	if _, err := corem07.ValidateAgentOutput(raw, evidence, registry); err != nil {
		for _, id := range p.EvidenceIDs {
			found := false
			for _, available := range ids {
				if id == available {
					found = true
					break
				}
			}
			if !found {
				return "REJECT_UNGROUNDED"
			}
		}
		return "INVALID_SCHEMA"
	}
	return "SUPPORTED"
}
