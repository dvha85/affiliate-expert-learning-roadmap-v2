package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

func linkedProposal(raw []byte, evaluations []m05.EvaluationRecord) (m05.ImprovementProposal, error) {
	p, status := m05.DecodeM05Proposal(raw)
	if status != "VALID" || m05.EvaluateImprovementProposal(p) != "REVIEW_REQUIRED" {
		return p, fmt.Errorf("invalid proposal schema or semantics")
	}
	if strings.TrimSpace(p.CurrentVersion) != p.CurrentVersion || strings.TrimSpace(p.ProposedVersion) != p.ProposedVersion {
		return p, fmt.Errorf("version must not contain surrounding whitespace")
	}
	// Learner importer requires an explicit risk assessment, stricter than the
	// shared schema's optional risks field. It is not proof the risks are accurate.
	if len(p.Risks) == 0 {
		return p, fmt.Errorf("explicit risks required")
	}
	for _, risk := range p.Risks {
		if strings.TrimSpace(risk) == "" {
			return p, fmt.Errorf("blank risk")
		}
	}
	seen := map[string]bool{}
	for _, id := range p.EvaluationIDs {
		if strings.TrimSpace(id) == "" || seen[id] {
			return p, fmt.Errorf("duplicate or empty evaluation ID")
		}
		seen[id] = true
		count := 0
		for _, e := range evaluations {
			if e.EvaluationID == id {
				count++
			}
		}
		if count != 1 {
			return p, fmt.Errorf("evaluation must resolve exactly once")
		}
	}
	if m05.ValidateProposalEvaluationLink(p, evaluations) != "VALID" {
		return p, fmt.Errorf("proposal evaluation link invalid")
	}
	return p, nil
}

func linkedReview(raw []byte, proposals []m05.ImprovementProposal, evaluations []m05.EvaluationRecord) (m05.ReviewRecord, error) {
	r, status := m05.DecodeM05Review(raw)
	if status != "VALID" || strings.TrimSpace(r.ReviewID) == "" || strings.TrimSpace(r.Reason) == "" {
		return r, fmt.Errorf("invalid human review")
	}
	var selected *m05.ImprovementProposal
	for i := range proposals {
		if proposals[i].ProposalID == r.ProposalID {
			if selected != nil {
				return r, fmt.Errorf("ambiguous proposal")
			}
			selected = &proposals[i]
		}
	}
	if selected == nil || m05.ValidateReviewRecord(r, *selected) != "VALID" {
		return r, fmt.Errorf("review proposal link invalid")
	}
	reviewed, err := time.Parse(time.RFC3339, r.ReviewedAt)
	if err != nil {
		return r, err
	}
	for _, id := range selected.EvaluationIDs {
		count := 0
		for _, e := range evaluations {
			if e.EvaluationID == id {
				count++
				at, err := time.Parse(time.RFC3339, e.EvaluatedAt)
				if err != nil || reviewed.Before(at) {
					return r, fmt.Errorf("review before evaluation")
				}
			}
		}
		if count != 1 {
			return r, fmt.Errorf("evaluation must resolve exactly once")
		}
	}
	return r, nil
}

// Single-writer JSONL, matching BR-12b ownership and framing; no repair/reset.
func loadImprovementRecords[T any](path string, decode func([]byte) (T, error), id func(T) string) ([]T, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("store must be regular, not a symlink")
	}
	f, err := (store.JSONL{}).Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) > 0 && data[len(data)-1] != '\n' {
			return 0, nil, fmt.Errorf("incomplete JSONL framing")
		}
		return bufio.ScanLines(data, atEOF)
	})
	values := []T{}
	seen := map[string]bool{}
	for scanner.Scan() {
		if len(scanner.Bytes()) > store.MaxHistoryRecordBytes {
			return nil, fmt.Errorf("record too large")
		}
		value, err := decode(scanner.Bytes())
		if err != nil {
			return nil, err
		}
		if seen[id(value)] {
			return nil, fmt.Errorf("duplicate record ID")
		}
		seen[id(value)] = true
		values = append(values, value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func importImprovement[T any](path, input string, existing []T, decode func([]byte) (T, error), id func(T) string) (string, any, error) {
	raw, err := readCampaignFile(input, store.MaxHistoryRecordBytes)
	if err != nil {
		return "INPUT_ERROR", nil, err
	}
	value, err := decode(raw)
	if err != nil {
		return "INVALID_RECORD", nil, err
	}
	for _, old := range existing {
		if id(old) == id(value) {
			if reflect.DeepEqual(old, value) {
				return "EXACT_DUPLICATE", value, nil
			}
			return "CONFLICT", nil, fmt.Errorf("record ID reused with different content")
		}
	}
	raw, err = json.Marshal(value)
	if err != nil {
		return "INVALID_RECORD", nil, err
	}
	if err := (store.JSONL{}).AppendLine(path, raw); err != nil {
		return "STORE_ERROR", nil, err
	}
	return "APPENDED", value, nil
}

func runImprovementStore(kind string, args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		envelope := map[string]any{"command": kind, "status": status, "execution_permitted": false, "auto_apply": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	count := 6
	if kind == "review" {
		count = 7
	}
	if (kind != "proposal" && kind != "review") || len(args) == 0 || (args[0] != "import" && args[0] != "list") || (args[0] == "list" && len(args) != count) || (args[0] == "import" && len(args) != count+1) {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot proposal import|list HISTORY ACTIONS OUTCOMES EVALUATIONS PROPOSALS [INPUT]; bot review import|list HISTORY ACTIONS OUTCOMES EVALUATIONS PROPOSALS REVIEWS [INPUT] (INPUT only for import)"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	h, err := LoadHistory(args[1])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	a, err := loadActions(args[2], h)
	if err != nil {
		return emit("ACTION_STORE_ERROR", nil, err, 1)
	}
	o, err := loadOutcomes(args[3], a)
	if err != nil {
		return emit("OUTCOME_STORE_ERROR", nil, err, 1)
	}
	e, err := loadEvaluations(args[4], h, a, o)
	if err != nil {
		return emit("EVALUATION_STORE_ERROR", nil, err, 1)
	}
	decodeP := func(raw []byte) (m05.ImprovementProposal, error) { return linkedProposal(raw, e) }
	idP := func(p m05.ImprovementProposal) string { return p.ProposalID }
	proposals, err := loadImprovementRecords(args[5], decodeP, idP)
	if err != nil && !(kind == "proposal" && args[0] == "import" && os.IsNotExist(err)) {
		return emit("PROPOSAL_STORE_ERROR", nil, err, 1)
	}
	if kind == "proposal" {
		if args[0] == "list" {
			return emit("VALID", proposals, nil, 0)
		}
		status, artifact, err := importImprovement(args[5], args[6], proposals, decodeP, idP)
		code := 0
		if err != nil {
			code = 1
		}
		return emit(status, artifact, err, code)
	}
	decodeR := func(raw []byte) (m05.ReviewRecord, error) { return linkedReview(raw, proposals, e) }
	idR := func(r m05.ReviewRecord) string { return r.ReviewID }
	reviews, err := loadImprovementRecords(args[6], decodeR, idR)
	if err != nil && !(args[0] == "import" && os.IsNotExist(err)) {
		return emit("REVIEW_STORE_ERROR", nil, err, 1)
	}
	if args[0] == "list" {
		return emit("VALID", reviews, nil, 0)
	}
	status, artifact, err := importImprovement(args[6], args[7], reviews, decodeR, idR)
	code := 0
	if err != nil {
		code = 1
	}
	return emit(status, artifact, err, code)
}
