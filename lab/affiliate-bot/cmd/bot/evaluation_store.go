package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

type evaluationConfig struct {
	EvaluationID string        `json:"evaluation_id"`
	DecisionID   string        `json:"decision_id"`
	EffectRef    m03.EffectRef `json:"effect_ref"`
	OutcomeIDs   []string      `json:"outcome_ids"`
	EvaluatedAt  string        `json:"evaluated_at"`
}

// Conservative v1 producer: observations without a reviewed comparison protocol
// never establish effectiveness. Outcome snapshots are not summed.
func buildEvaluation(c evaluationConfig, history []HistoryRecord, actions []m03.HumanActionRecord, outcomes []m03.OutcomeRecord) (m05.EvaluationRecord, error) {
	bad := func(s string) (m05.EvaluationRecord, error) {
		return m05.EvaluationRecord{}, fmt.Errorf("evaluation: %s", s)
	}
	if strings.TrimSpace(c.EvaluationID) == "" || len(c.OutcomeIDs) == 0 {
		return bad("evaluation_id and outcome_ids required")
	}
	if err := resolveActionDecision(history, c.DecisionID); err != nil {
		return bad("decision does not resolve/replay")
	}
	if c.EffectRef.EffectKind != "HUMAN_ACTION" {
		return bad("only stored human action supported")
	}
	var action *m03.HumanActionRecord
	for i := range actions {
		if actions[i].ActionID == c.EffectRef.EffectID {
			if action != nil {
				return bad("ambiguous action")
			}
			action = &actions[i]
		}
	}
	if action == nil || action.DecisionID != c.DecisionID {
		return bad("orphan or mismatched action")
	}
	at, err := time.Parse(time.RFC3339, c.EvaluatedAt)
	if err != nil {
		return bad("invalid evaluated_at")
	}
	selected := []m03.OutcomeRecord{}
	seen := map[string]bool{}
	for _, id := range c.OutcomeIDs {
		if strings.TrimSpace(id) == "" || seen[id] {
			return bad("empty or duplicate outcome ID")
		}
		seen[id] = true
		count := 0
		for _, o := range outcomes {
			if o.OutcomeID == id {
				count++
				if m03.ValidateActionOutcomeLink(*action, o) != "VALID" {
					return bad("outcome effect mismatch")
				}
				observed, err := time.Parse(time.RFC3339, o.ObservedAt)
				if err != nil || at.Before(observed) {
					return bad("evaluation before outcome")
				}
				selected = append(selected, o)
			}
		}
		if count != 1 {
			return bad("outcome must resolve exactly once")
		}
	}
	ids := append([]string(nil), c.OutcomeIDs...)
	sort.Strings(ids)
	e := m05.EvaluationRecord{EvaluationID: c.EvaluationID, DecisionID: c.DecisionID, EffectRef: c.EffectRef, OutcomeIDs: ids, EvaluatedAt: c.EvaluatedAt, Result: "INCONCLUSIVE", EvidenceIDs: append([]string(nil), ids...), Limitations: []string{"br12-evaluation/v1: chưa có protocol so sánh và tiêu chí hiệu quả được review; không suy ra hiệu quả từ một chuỗi quan sát.", "Outcome là snapshot do người nhập khai báo; pending không phải zero; không cộng dồn snapshot hoặc dùng kết quả test làm kết quả kinh doanh."}}
	raw, err := json.Marshal(e)
	if err != nil {
		return bad("serialization failed")
	}
	if _, s := m05.DecodeM05Evaluation(raw); s != "VALID" {
		return bad(s)
	}
	if s := m05.ValidateEvaluationRecord(e, *action, selected); s != "VALID" {
		return bad(s)
	}
	return e, nil
}

func loadEvaluations(path string, history []HistoryRecord, actions []m03.HumanActionRecord, outcomes []m03.OutcomeRecord) ([]m05.EvaluationRecord, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("evaluation store must be a regular file, not a symlink")
	}
	f, err := (store.JSONL{}).Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) > 0 && data[len(data)-1] != '\n' {
			return 0, nil, fmt.Errorf("evaluation store has incomplete JSONL framing")
		}
		return bufio.ScanLines(data, atEOF)
	})
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	result := []m05.EvaluationRecord{}
	seen := map[string]bool{}
	for scanner.Scan() {
		raw := scanner.Bytes()
		if len(raw) > store.MaxHistoryRecordBytes {
			return nil, fmt.Errorf("evaluation record too large")
		}
		e, status := m05.DecodeM05Evaluation(raw)
		if status != "VALID" {
			return nil, fmt.Errorf("evaluation store: %s", status)
		}
		want, err := buildEvaluation(evaluationConfig{e.EvaluationID, e.DecisionID, e.EffectRef, e.OutcomeIDs, e.EvaluatedAt}, history, actions, outcomes)
		if err != nil {
			return nil, err
		}
		if !reflect.DeepEqual(e, want) || seen[e.EvaluationID] {
			return nil, fmt.Errorf("evaluation store inconsistent or duplicate")
		}
		seen[e.EvaluationID] = true
		result = append(result, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func runEvaluationStore(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": "evaluation", "status": status, "execution_permitted": false, "auto_apply": false}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) == 0 || (args[0] != "create" && args[0] != "list") || (args[0] == "create" && len(args) != 6) || (args[0] == "list" && len(args) != 5) {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot evaluation create HISTORY ACTIONS OUTCOMES EVALUATIONS CONFIG | bot evaluation list HISTORY ACTIONS OUTCOMES EVALUATIONS"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	if args[0] == "create" {
		release, lockErr := acquireHistoryRuntimeGate(args[1])
		if lockErr != nil {
			return emit("BUSY", nil, lockErr, 1)
		}
		defer release()
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
	existing, err := loadEvaluations(args[4], h, a, o)
	if err != nil && !(args[0] == "create" && os.IsNotExist(err)) {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if args[0] == "list" {
		return emit("VALID", existing, nil, 0)
	}
	raw, err := readCampaignFile(args[5], store.MaxHistoryRecordBytes)
	if err != nil {
		return emit("CONFIG_ERROR", nil, err, 1)
	}
	var c evaluationConfig
	if err := contracts.DecodeStrict(raw, &c); err != nil {
		return emit("CONFIG_ERROR", nil, err, 1)
	}
	e, err := buildEvaluation(c, h, a, o)
	if err != nil {
		return emit("EVALUATION_ERROR", nil, err, 1)
	}
	for _, old := range existing {
		if old.EvaluationID == e.EvaluationID {
			if reflect.DeepEqual(old, e) {
				return emit("EXACT_DUPLICATE", e, nil, 0)
			}
			return emit("CONFLICT", nil, fmt.Errorf("evaluation_id reused with different content"), 1)
		}
	}
	encoded, err := json.Marshal(e)
	if err != nil {
		return emit("INVALID_SCHEMA", nil, err, 1)
	}
	if err := (store.JSONL{}).AppendLine(args[4], encoded); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	return emit("APPENDED", e, nil, 0)
}
