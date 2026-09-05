package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

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
	return o, validateAdvisorFields(o)
}

func validateAdvisorFields(o AdvisorOutput) string {
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

func runAdvisorCheck(w io.Writer, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("usage: demo advisor-check OUTPUT.json EVIDENCE.json AS_OF MAX_AGE_HOURS")
	}
	raw, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	o, state := DecodeAdvisorOutput(raw)
	if state != missionValid {
		return fmt.Errorf("advisor output: %s", state)
	}
	evidenceRaw, err := os.ReadFile(args[1])
	if err != nil {
		return err
	}
	var evidence []AdvisorEvidence
	if err := json.Unmarshal(evidenceRaw, &evidence); err != nil {
		return fmt.Errorf("invalid evidence JSON: %w", err)
	}
	maxAge, err := strconv.Atoi(args[3])
	if err != nil || maxAge < 0 {
		return fmt.Errorf("MAX_AGE_HOURS must be a non-negative integer")
	}
	state = EvaluateAdvisorOutput(o, evidence, args[2], maxAge)
	return json.NewEncoder(w).Encode(struct {
		Output AdvisorOutput `json:"advisor_output"`
		Result string        `json:"result"`
	}{o, state})
}
