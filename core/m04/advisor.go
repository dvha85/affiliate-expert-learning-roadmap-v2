package m04

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"
)

const (
	missionValid   = "VALID"
	missionInvalid = "INVALID"
)

type AdvisorEvidence struct {
	EvidenceID string `json:"evidence_id"`
	ObservedAt string `json:"observed_at"`
	SourceRef  string `json:"source_ref"`
}
type AdvisorOutput struct {
	State              string   `json:"state"`
	Recommendation     string   `json:"recommendation"`
	Reason             string   `json:"reason"`
	EvidenceIDs        []string `json:"evidence_ids"`
	Unknowns           []string `json:"unknowns"`
	WriteToolRequested bool     `json:"write_tool_requested"`
}

func EvaluateAdvisorOutput(o AdvisorOutput, ev []AdvisorEvidence, asOf string, maxAgeHours int) string {
	if state := ValidateAdvisorFields(o); state != missionValid {
		return state
	}
	if o.State == "ABSTAIN" {
		return "ABSTAIN"
	}
	if o.State != "ADVISE" && o.State != "HUMAN_REVIEW" {
		return missionInvalid
	}
	if len(o.EvidenceIDs) == 0 {
		return "REJECT_UNGROUNDED"
	}
	now, e := time.Parse(time.RFC3339, asOf)
	if e != nil {
		return missionInvalid
	}
	idx := map[string]AdvisorEvidence{}
	for _, x := range ev {
		if strings.TrimSpace(x.EvidenceID) == "" || strings.TrimSpace(x.SourceRef) == "" {
			return missionInvalid
		}
		if _, exists := idx[x.EvidenceID]; exists {
			return missionInvalid
		}
		idx[x.EvidenceID] = x
	}
	for _, id := range o.EvidenceIDs {
		x, ok := idx[id]
		if !ok {
			return "REJECT_UNGROUNDED"
		}
		at, e := time.Parse(time.RFC3339, x.ObservedAt)
		if e != nil {
			return missionInvalid
		}
		if at.After(now) {
			return "ABSTAIN_FUTURE"
		}
		if maxAgeHours >= 0 && now.Sub(at) > time.Duration(maxAgeHours)*time.Hour {
			return "ABSTAIN_STALE"
		}
	}
	return "SUPPORTED"
}

// DecodeAdvisorOutput checks the original bytes before Go zero values can hide
// absent fields, null values, duplicate keys or unknown properties.
// This is a contract-specific validator, not a general JSON Schema engine.
func DecodeAdvisorOutput(raw []byte) (AdvisorOutput, string) {
	var o AdvisorOutput
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return o, "INVALID_SCHEMA"
	}
	fields := map[string]json.RawMessage{}
	allowed := map[string]bool{"state": true, "recommendation": true, "reason": true, "evidence_ids": true, "unknowns": true, "write_tool_requested": true}
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return o, "INVALID_SCHEMA"
		}
		name, ok := key.(string)
		if !ok || !allowed[name] {
			return o, "INVALID_SCHEMA"
		}
		if _, exists := fields[name]; exists {
			return o, "INVALID_SCHEMA"
		}
		var value json.RawMessage
		if d.Decode(&value) != nil || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return o, "INVALID_SCHEMA"
		}
		fields[name] = value
	}
	if _, err := d.Token(); err != nil {
		return o, "INVALID_SCHEMA"
	}
	if _, err := d.Token(); err != io.EOF {
		return o, "INVALID_SCHEMA"
	}
	for _, name := range []string{"state", "reason", "evidence_ids", "unknowns", "write_tool_requested"} {
		if _, ok := fields[name]; !ok {
			return o, "INVALID_SCHEMA"
		}
	}
	for _, name := range []string{"evidence_ids", "unknowns"} {
		var items []json.RawMessage
		if json.Unmarshal(fields[name], &items) != nil {
			return o, "INVALID_SCHEMA"
		}
		for _, item := range items {
			var value string
			if bytes.Equal(bytes.TrimSpace(item), []byte("null")) || json.Unmarshal(item, &value) != nil {
				return o, "INVALID_SCHEMA"
			}
		}
	}
	if json.Unmarshal(raw, &o) != nil {
		return AdvisorOutput{}, "INVALID_SCHEMA"
	}
	return o, ValidateAdvisorFields(o)
}

func ValidateAdvisorFields(o AdvisorOutput) string {
	if o.WriteToolRequested {
		return "REJECT_WRITE_REQUEST"
	}
	if o.State != "ADVISE" && o.State != "HUMAN_REVIEW" && o.State != "ABSTAIN" {
		return "INVALID_SCHEMA"
	}
	if strings.TrimSpace(o.Reason) == "" || o.EvidenceIDs == nil || o.Unknowns == nil {
		return "INVALID_SCHEMA"
	}
	seen := map[string]bool{}
	for _, id := range o.EvidenceIDs {
		if strings.TrimSpace(id) == "" || seen[id] {
			return "INVALID_SCHEMA"
		}
		seen[id] = true
	}
	return missionValid
}
