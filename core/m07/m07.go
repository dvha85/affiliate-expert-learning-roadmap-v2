// Package m07 contains the small, deterministic parts of the M07 boundary
// that must be shared by the learner CLI and workflow adapters.
package m07

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type ToolSpec struct {
	Name            string   `json:"name"`
	ReadOnly        bool     `json:"read_only"`
	AllowedMethods  []string `json:"allowed_methods"`
	AllowedHosts    []string `json:"allowed_hosts"`
	TimeoutMS       int      `json:"timeout_ms,omitempty"`
	FollowRedirects bool     `json:"follow_redirects,omitempty"`
}

type ToolRequest struct {
	ToolName string `json:"tool_name"`
	Method   string `json:"method"`
	Target   string `json:"target"`
}

type Claim struct {
	Text         string          `json:"text"`
	FieldOrClaim string          `json:"field_or_claim"`
	Value        json.RawMessage `json:"value"`
	EvidenceIDs  []string        `json:"evidence_ids"`
}

type AgentOutput struct {
	State           string        `json:"state"`
	Answer          string        `json:"answer"`
	EvidenceIDs     []string      `json:"evidence_ids"`
	Claims          []Claim       `json:"claims"`
	ToolCalls       []ToolRequest `json:"tool_calls"`
	Authority       string        `json:"authority"`
	WritePermission bool          `json:"write_permission"`
}

type Evidence struct {
	EvidenceID            string `json:"evidence_id"`
	SubjectID             string `json:"subject_id,omitempty"`
	FieldOrClaim          string `json:"field_or_claim,omitempty"`
	Value                 any    `json:"value,omitempty"`
	ClaimKind             string `json:"claim_kind"`
	SourceAuthorityOrRole string `json:"source_authority_or_role,omitempty"`
	Limitation            string `json:"limitation"`
}

func uniqueNonEmpty(values []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func ValidateRegistry(registry []ToolSpec) error {
	seen := map[string]bool{}
	for _, tool := range registry {
		if strings.TrimSpace(tool.Name) == "" || seen[tool.Name] || !tool.ReadOnly {
			return fmt.Errorf("tool registry must contain unique read-only tools")
		}
		seen[tool.Name] = true
		if len(tool.AllowedMethods) == 0 || len(tool.AllowedHosts) == 0 {
			return fmt.Errorf("tool %s has empty method/host policy", tool.Name)
		}
		if tool.TimeoutMS < 0 || tool.TimeoutMS > 60000 {
			return fmt.Errorf("tool %s has invalid timeout", tool.Name)
		}
		if tool.FollowRedirects {
			return fmt.Errorf("tool %s must disable redirects", tool.Name)
		}
		for _, method := range tool.AllowedMethods {
			method = strings.ToUpper(strings.TrimSpace(method))
			if method != "GET" && method != "HEAD" {
				return fmt.Errorf("tool %s allows a write method", tool.Name)
			}
		}
		for _, host := range tool.AllowedHosts {
			if err := validateHost(host); err != nil {
				return fmt.Errorf("tool %s: %w", tool.Name, err)
			}
		}
	}
	if len(seen) == 0 {
		return fmt.Errorf("empty tool registry")
	}
	return nil
}

func validateHost(host string) error {
	host = strings.TrimSpace(host)
	u, err := url.Parse("https://" + host)
	if err != nil || u.Hostname() == "" || u.Host != host || u.User != nil || u.Port() != "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("invalid allowlisted host %q", host)
	}
	return nil
}

// ValidateToolRequest is the preflight policy adapter. It must run before a
// request is handed to an HTTP client or an Agent tool node.
func ValidateToolRequest(request ToolRequest, registry []ToolSpec) error {
	if err := ValidateRegistry(registry); err != nil {
		return err
	}
	tool, ok := func() (ToolSpec, bool) {
		for _, candidate := range registry {
			if candidate.Name == request.ToolName {
				return candidate, true
			}
		}
		return ToolSpec{}, false
	}()
	if !ok || !tool.ReadOnly {
		return fmt.Errorf("tool is not registered read-only")
	}
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if method != "GET" && method != "HEAD" {
		return fmt.Errorf("method %s is not read-only", method)
	}
	allowed := false
	for _, candidate := range tool.AllowedMethods {
		if strings.EqualFold(strings.TrimSpace(candidate), method) {
			allowed = true
		}
	}
	if !allowed {
		return fmt.Errorf("method is not allowed by registry")
	}
	u, err := url.Parse(request.Target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Port() != "" {
		return fmt.Errorf("target must be an https URL without userinfo or port")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("target query/fragment must be explicitly covered by the adapter")
	}
	for _, host := range tool.AllowedHosts {
		if strings.EqualFold(strings.TrimSpace(host), u.Hostname()) {
			return nil
		}
	}
	return fmt.Errorf("target host is not allowlisted")
}

// ValidateAgentOutput checks the model's actual structured output. Context IDs
// are never copied into the answer automatically: every cited ID and every
// claim must be present in the model output and resolve in the supplied store.
func ValidateAgentOutput(raw []byte, evidence []Evidence, registry []ToolSpec) (AgentOutput, error) {
	var shape map[string]json.RawMessage
	if err := json.Unmarshal(raw, &shape); err != nil {
		return AgentOutput{}, fmt.Errorf("agent output is not JSON: %w", err)
	}
	for _, required := range []string{"state", "answer", "evidence_ids", "claims", "tool_calls", "authority", "write_permission"} {
		if _, ok := shape[required]; !ok {
			return AgentOutput{}, fmt.Errorf("agent output missing required field %q", required)
		}
	}
	if string(bytes.TrimSpace(shape["write_permission"])) != "false" {
		return AgentOutput{}, fmt.Errorf("write_permission must be the literal false")
	}
	var output AgentOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		return output, fmt.Errorf("agent output is not JSON: %w", err)
	}
	if output.State != "HUMAN_REVIEW" && output.State != "ABSTAIN" && output.State != "PROPOSE" {
		return output, fmt.Errorf("unsupported agent state")
	}
	if strings.TrimSpace(output.Answer) == "" {
		return output, fmt.Errorf("agent answer is empty")
	}
	answerLower := strings.ToLower(output.Answer)
	for _, marker := range []string{"ignore the policy", "ignore previous instructions", "system prompt", "write permission", "execute this"} {
		if strings.Contains(answerLower, marker) {
			return output, fmt.Errorf("answer contains an instruction-injection marker")
		}
	}
	if output.Authority != "A2-RO" || output.WritePermission {
		return output, fmt.Errorf("agent authority/write contract violated")
	}
	if !uniqueNonEmpty(output.EvidenceIDs) {
		return output, fmt.Errorf("evidence_ids must be unique and non-empty")
	}
	known := map[string]Evidence{}
	for _, item := range evidence {
		if strings.TrimSpace(item.EvidenceID) == "" || known[item.EvidenceID].EvidenceID != "" {
			return output, fmt.Errorf("invalid evidence context")
		}
		known[item.EvidenceID] = item
	}
	if output.State != "ABSTAIN" && len(output.Claims) == 0 {
		return output, fmt.Errorf("grounded output must contain claims")
	}
	for _, id := range output.EvidenceIDs {
		if _, ok := known[id]; !ok {
			return output, fmt.Errorf("evidence id %q is not in canonical context", id)
		}
	}
	for _, claim := range output.Claims {
		if strings.TrimSpace(claim.Text) == "" || strings.TrimSpace(claim.FieldOrClaim) == "" || len(claim.Value) == 0 || !uniqueNonEmpty(claim.EvidenceIDs) || len(claim.EvidenceIDs) == 0 {
			return output, fmt.Errorf("every claim needs text and evidence")
		}
		matched := false
		for _, id := range claim.EvidenceIDs {
			item, ok := known[id]
			if !ok {
				return output, fmt.Errorf("claim cites unknown evidence id %q", id)
			}
			var claimValue any
			if err := json.Unmarshal(claim.Value, &claimValue); err != nil {
				return output, fmt.Errorf("claim value is not valid JSON: %w", err)
			}
			var evidenceValue any
			evidenceRaw, err := json.Marshal(item.Value)
			if err != nil || json.Unmarshal(evidenceRaw, &evidenceValue) != nil {
				return output, fmt.Errorf("evidence %q has an unserializable value", id)
			}
			claimRaw, _ := json.Marshal(claimValue)
			evidenceCanonical, _ := json.Marshal(evidenceValue)
			if item.FieldOrClaim == claim.FieldOrClaim && string(claimRaw) == string(evidenceCanonical) {
				matched = true
			}
		}
		if !matched {
			return output, fmt.Errorf("claim value is not bound to a cited evidence field")
		}
	}
	for _, call := range output.ToolCalls {
		if err := ValidateToolRequest(call, registry); err != nil {
			return output, fmt.Errorf("tool call rejected: %w", err)
		}
	}
	return output, nil
}
