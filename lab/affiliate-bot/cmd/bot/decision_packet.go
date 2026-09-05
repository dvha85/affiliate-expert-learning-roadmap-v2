package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
)

// DecisionContext is supplied explicitly by the learner, never invented by ranking.
type DecisionContext struct {
	Question        string   `json:"question"`
	SupportedFacts  []string `json:"supported_facts"`
	Assumptions     []string `json:"assumptions"`
	Unknowns        []string `json:"unknowns"`
	NextMeasurement string   `json:"next_measurement"`
}

type DecisionPacket struct {
	DecisionID      string   `json:"decision_id"`
	Question        string   `json:"question"`
	EvidenceIDs     []string `json:"evidence_ids"`
	State           string   `json:"state"`
	Reason          string   `json:"reason"`
	SupportedFacts  []string `json:"supported_facts"`
	Assumptions     []string `json:"assumptions"`
	Unknowns        []string `json:"unknowns"`
	MissingEvidence []string `json:"missing_evidence"`
	NextMeasurement string   `json:"next_measurement"`
	Action          any      `json:"action"`
}

func copyStrings(items []string) []string { return append([]string{}, items...) }

func DecodeDecisionContext(raw []byte) (DecisionContext, error) {
	var context DecisionContext
	value, err := contracts.Decode(raw)
	if err != nil {
		return context, err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return context, fmt.Errorf("context must be an object")
	}
	for _, name := range []string{"question", "supported_facts", "assumptions", "unknowns", "next_measurement"} {
		if object[name] == nil {
			return context, fmt.Errorf("context field %s is required and non-null", name)
		}
	}
	for _, name := range []string{"supported_facts", "assumptions", "unknowns"} {
		items, ok := object[name].([]any)
		if !ok {
			return context, fmt.Errorf("context %s must be an array", name)
		}
		for _, item := range items {
			if _, ok := item.(string); !ok {
				return context, fmt.Errorf("context %s must contain strings", name)
			}
		}
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&context); err != nil {
		return context, err
	}
	if len(object) != 5 {
		return context, fmt.Errorf("context has unknown fields")
	}
	return context, nil
}

func HistoryDecisionPacket(record HistoryRecord, context DecisionContext) (DecisionPacket, error) {
	var packet DecisionPacket
	// Refuse drift and unknown formulas; never export an unverified recorded result.
	if report := Replay(record); report.State != replayMatch {
		return packet, fmt.Errorf("cannot export decision: replay=%s: %s", report.State, report.Reason)
	}
	if strings.TrimSpace(context.Question) == "" || context.SupportedFacts == nil || context.Assumptions == nil || context.Unknowns == nil {
		return packet, fmt.Errorf("question and explicit context arrays are required")
	}
	r := record.RecordedResult
	packet = DecisionPacket{
		DecisionID: r.DecisionID, Question: context.Question, EvidenceIDs: copyStrings(r.EvidenceIDs),
		State: r.State, Reason: strings.Join(r.Reasons, "\n"), SupportedFacts: copyStrings(context.SupportedFacts),
		Assumptions: copyStrings(context.Assumptions), Unknowns: copyStrings(context.Unknowns),
		MissingEvidence: copyStrings(r.MissingEvidence), NextMeasurement: context.NextMeasurement, Action: nil,
	}
	raw, err := json.Marshal(packet)
	if err != nil {
		return DecisionPacket{}, err
	}
	if err := contracts.ValidateRaw("decision-packet.schema.json", raw); err != nil {
		return DecisionPacket{}, fmt.Errorf("invalid DecisionPacket: %w", err)
	}
	return packet, nil
}

func exportHistoryDecision(w io.Writer, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: history decision <history.jsonl> <record_id> <context.json>")
	}
	records, err := LoadHistory(args[0])
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(args[2])
	if err != nil {
		return err
	}
	context, err := DecodeDecisionContext(raw)
	if err != nil {
		return err
	}
	var selected *HistoryRecord
	for i := range records {
		if records[i].RecordID == args[1] {
			if selected != nil {
				return fmt.Errorf("ambiguous duplicate record_id %q", args[1])
			}
			selected = &records[i]
		}
	}
	if selected == nil {
		return fmt.Errorf("record_id %q not found", args[1])
	}
	packet, err := HistoryDecisionPacket(*selected, context)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(packet)
}
