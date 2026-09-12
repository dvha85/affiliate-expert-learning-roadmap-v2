package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

func linkedOutcome(raw []byte, actions []m03.HumanActionRecord) (m03.OutcomeRecord, string) {
	o, status := m03.DecodeM03Outcome(raw)
	if status != "VALID" {
		return o, status
	}
	// Reject metric rounding/underflow instead of persisting a different value.
	var source struct {
		Metrics map[string]json.RawMessage `json:"metrics"`
	}
	if json.Unmarshal(raw, &source) != nil {
		return o, "INVALID_SCHEMA"
	}
	for key, value := range source.Metrics {
		decimal := strings.TrimSpace(string(value))
		if len(decimal) > 128 {
			return o, "METRIC_PRECISION_ERROR"
		}
		if at := strings.IndexAny(decimal, "eE"); at >= 0 {
			exp, err := strconv.Atoi(decimal[at+1:])
			if err != nil || exp < -400 || exp > 400 {
				return o, "METRIC_PRECISION_ERROR"
			}
		}
		original, ok := new(big.Rat).SetString(decimal)
		if !ok {
			return o, "METRIC_PRECISION_ERROR"
		}
		b, err := json.Marshal(o.Metrics[key])
		if err != nil {
			return o, "METRIC_PRECISION_ERROR"
		}
		projected, ok := new(big.Rat).SetString(string(b))
		if !ok || original.Cmp(projected) != 0 {
			return o, "METRIC_PRECISION_ERROR"
		}
	}
	if o.EffectRef.EffectKind != "HUMAN_ACTION" {
		return o, "REJECT_MACHINE_EXECUTION"
	}
	var selected *m03.HumanActionRecord
	for i := range actions {
		if actions[i].ActionID == o.EffectRef.EffectID {
			if selected != nil {
				return o, "AMBIGUOUS_ACTION"
			}
			selected = &actions[i]
		}
	}
	if selected == nil {
		return o, "ORPHAN_ACTION"
	}
	return o, m03.ValidateActionOutcomeLink(*selected, o)
}

func loadOutcomes(path string, actions []m03.HumanActionRecord) (out []m03.OutcomeRecord, err error) {
	f, err := (store.JSONL{}).Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			out = nil
			err = closeErr
		}
	}()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	out = []m03.OutcomeRecord{}
	ids := map[string]bool{}
	for scanner.Scan() {
		if len(scanner.Bytes()) > store.MaxHistoryRecordBytes {
			return nil, fmt.Errorf("outcome record too large")
		}
		o, status := linkedOutcome(scanner.Bytes(), actions)
		if status != "VALID" {
			return nil, fmt.Errorf("outcome store: %s", status)
		}
		if ids[o.OutcomeID] {
			return nil, fmt.Errorf("duplicate outcome_id in store")
		}
		ids[o.OutcomeID] = true
		out = append(out, o)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func sameOutcome(a, b m03.OutcomeRecord) bool { return reflect.DeepEqual(a, b) }

func appendOutcome(path string, encoded []byte) error {
	return (store.JSONL{}).AppendLine(path, encoded)
}

func runOutcomeStore(args []string, stdout, stderr io.Writer) int {
	command := "outcome"
	if len(args) > 0 {
		command += " " + args[0]
	}
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		envelope := map[string]any{"command": command, "status": status, "execution_permitted": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) > 0 && args[0] == "accesstrade-import" {
		return runAccesstradeOutcomeImport(args, stdout, stderr)
	}
	if len(args) > 0 && args[0] == "accesstrade-receipts" {
		return runAccesstradeReceiptList(args, stdout, stderr)
	}
	if len(args) == 0 || (args[0] != "import" && args[0] != "list") || (args[0] == "import" && len(args) != 5) || (args[0] == "list" && len(args) != 4) {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot outcome import HISTORY ACTIONS OUTCOMES INPUT | bot outcome list HISTORY ACTIONS OUTCOMES"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	// An outcome list is a history/action/outcome join. Use the same local gate
	// as imports so it cannot expose a mixed derived-store snapshot.
	release, lockErr := acquireHistoryRuntimeGate(args[1])
	if lockErr != nil {
		return emit("BUSY", nil, lockErr, 1)
	}
	defer release()
	history, err := LoadHistory(args[1])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	actions, err := loadActions(args[2], history)
	if err != nil {
		return emit("ACTION_STORE_ERROR", nil, err, 1)
	}
	outcomes, err := loadOutcomes(args[3], actions)
	if err != nil && !(args[0] == "import" && os.IsNotExist(err)) {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if args[0] == "list" {
		return emit("VALID", outcomes, nil, 0)
	}
	raw, err := os.ReadFile(args[4])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	o, status := linkedOutcome(raw, actions)
	if status != "VALID" {
		return emit(status, nil, fmt.Errorf("outcome rejected: %s", status), 1)
	}
	for _, old := range outcomes {
		if old.OutcomeID == o.OutcomeID {
			if sameOutcome(old, o) {
				return emit("EXACT_DUPLICATE", o, nil, 0)
			}
			return emit("CONFLICT", nil, fmt.Errorf("outcome_id reused with different content"), 1)
		}
	}
	encoded, err := json.Marshal(o)
	if err != nil {
		return emit("INVALID_SCHEMA", nil, err, 1)
	}
	if err := appendOutcome(args[3], encoded); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	return emit("APPENDED", o, nil, 0)
}
