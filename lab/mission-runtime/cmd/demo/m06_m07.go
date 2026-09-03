package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"time"
)

type WatchRequest struct {
	Method        string   `json:"method"`
	URL           string   `json:"url"`
	AllowHosts    []string `json:"allow_hosts"`
	ObservedAt    string   `json:"observed_at"`
	CorrelationID string   `json:"correlation_id"`
	Body          string   `json:"body"`
	PreviousHash  string   `json:"previous_hash"`
}

func contentHash(body string) string {
	s := sha256.Sum256([]byte(body))
	return hex.EncodeToString(s[:])
}
func EvaluateWatchRequest(r WatchRequest) string {
	m := strings.ToUpper(strings.TrimSpace(r.Method))
	if m != "GET" && m != "HEAD" {
		return "REJECT_WRITE_METHOD"
	}
	u, e := url.Parse(r.URL)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" {
		return "REJECT_SOURCE"
	}
	allowed := false
	for _, h := range r.AllowHosts {
		if strings.EqualFold(strings.TrimSpace(h), u.Hostname()) {
			allowed = true
		}
	}
	if !allowed {
		return "REJECT_SOURCE"
	}
	if _, e := time.Parse(time.RFC3339, r.ObservedAt); e != nil || strings.TrimSpace(r.CorrelationID) == "" {
		return missionInvalid
	}
	current := contentHash(r.Body)
	if r.PreviousHash == "" {
		return "NEW"
	}
	if r.PreviousHash == current {
		return "UNCHANGED"
	}
	return "CHANGED"
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

func EvaluateAgentProposal(p AgentProposal, registry []ToolSpec, ids []string) string {
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
