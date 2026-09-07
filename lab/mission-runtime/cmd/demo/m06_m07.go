package main

import (
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m06"
	"net/url"
	"strings"
)

type WatchRequest = m06.WatchRequest
type CanonicalObservation = m06.CanonicalObservation

func contentHash(body string) string             { return m06.ContentHash(body) }
func EvaluateWatchRequest(r WatchRequest) string { return m06.EvaluateWatchRequest(r) }
func NormalizeWatchObservation(r WatchRequest, subjectID string) (CanonicalObservation, string) {
	return m06.NormalizeWatchObservation(r, subjectID)
}

type ToolSpec struct {
	Name           string   `json:"name"`
	ReadOnly       bool     `json:"read_only"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHosts   []string `json:"allowed_hosts"`
}
type AgentToolCall struct {
	ToolName string `json:"tool_name"`
	Method   string `json:"method"`
	Target   string `json:"target"`
}
type AgentProposal struct {
	State       string          `json:"state"`
	Answer      string          `json:"answer"`
	EvidenceIDs []string        `json:"evidence_ids"`
	ToolCalls   []AgentToolCall `json:"tool_calls"`
}

func EvaluateAgentProposal(p AgentProposal, registry []ToolSpec, ids []string, untrustedToolText ...string) string {
	_ = untrustedToolText // tool/page text is data only; never interpreted as permission or policy here.
	if p.State == "ABSTAIN" {
		return "ABSTAIN"
	}
	if p.State != "PROPOSE" && p.State != "HUMAN_REVIEW" {
		return missionInvalid
	}
	ev := map[string]bool{}
	for _, id := range ids {
		ev[id] = true
	}
	if len(p.EvidenceIDs) == 0 {
		return "REJECT_UNGROUNDED"
	}
	for _, id := range p.EvidenceIDs {
		if !ev[id] {
			return "REJECT_UNGROUNDED"
		}
	}
	tools := map[string]ToolSpec{}
	for _, t := range registry {
		tools[t.Name] = t
	}
	for _, c := range p.ToolCalls {
		t, ok := tools[c.ToolName]
		if !ok || !t.ReadOnly {
			return "REJECT_TOOL"
		}
		m := strings.ToUpper(strings.TrimSpace(c.Method))
		methodOK := false
		for _, a := range t.AllowedMethods {
			if strings.EqualFold(a, m) {
				methodOK = true
			}
		}
		if !methodOK || (m != "GET" && m != "HEAD") {
			return "REJECT_TOOL"
		}
		u, e := url.Parse(c.Target)
		if e != nil || u.Scheme != "https" {
			return "REJECT_TOOL"
		}
		hostOK := false
		for _, h := range t.AllowedHosts {
			if strings.EqualFold(h, u.Hostname()) {
				hostOK = true
			}
		}
		if !hostOK {
			return "REJECT_TOOL"
		}
	}
	return "SUPPORTED"
}
