package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

// Resolve the recorded decision, not a caller-supplied packet or an approval.
func resolveActionDecision(records []HistoryRecord, id string) error {
	count := 0
	for _, r := range records {
		if r.RecordedResult.DecisionID == id {
			count++
			if Replay(r).State != replayMatch {
				return fmt.Errorf("decision replay is not MATCH")
			}
		}
	}
	if count != 1 {
		return fmt.Errorf("decision_id must resolve exactly once (found %d)", count)
	}
	return nil
}

func loadActions(path string, records []HistoryRecord) (actions []m03.HumanActionRecord, err error) {
	f, err := (store.JSONL{}).Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); err == nil && closeErr != nil {
			actions = nil
			err = closeErr
		}
	}()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	actions = []m03.HumanActionRecord{}
	seen := map[string]bool{}
	for scanner.Scan() {
		if len(scanner.Bytes()) > store.MaxHistoryRecordBytes {
			return nil, fmt.Errorf("action record too large")
		}
		a, status := m03.DecodeM03Action(scanner.Bytes())
		if status != "VALID" || m03.ValidateHumanActionRecord(a) != "VALID" {
			return nil, fmt.Errorf("invalid action store record")
		}
		if seen[a.ActionID] {
			return nil, fmt.Errorf("duplicate action_id in store")
		}
		seen[a.ActionID] = true
		if err := resolveActionDecision(records, a.DecisionID); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return actions, nil
}

// Compare lexical paths and existing inode aliases, including symlinks/hardlinks.
// It is shared by append stores and immutable command artifacts so no command
// can use a source path as its output destination.
func distinctPaths(paths ...string) error {
	for i, p := range paths {
		for _, q := range paths[:i] {
			a, err := filepath.Abs(p)
			if err != nil {
				return err
			}
			b, err := filepath.Abs(q)
			if err != nil {
				return err
			}
			if a == b {
				return fmt.Errorf("input and output paths must be distinct")
			}
			pi, pe := os.Stat(p)
			qi, qe := os.Stat(q)
			if pe == nil && qe == nil && os.SameFile(pi, qi) {
				return fmt.Errorf("aliased input/output paths")
			}
		}
	}
	return nil
}

func distinctActionPaths(paths ...string) error { return distinctPaths(paths...) }

// Record/list orchestration remains next to M02 application ownership. The
// JSONL adapter only writes validated bytes; no executor or network is involved.
func runActionStore(args []string, stdout, stderr io.Writer) int {
	command := "action"
	if len(args) > 0 {
		command += " " + args[0]
	}
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		out := map[string]any{"command": command, "status": status, "execution_permitted": false}
		if artifact != nil {
			out["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) < 1 || (args[0] != "record" && args[0] != "list") || (args[0] == "record" && len(args) != 4) || (args[0] == "list" && len(args) != 3) {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot action record HISTORY.jsonl ACTIONS.jsonl ACTION.json | bot action list HISTORY.jsonl ACTIONS.jsonl"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	// List responses join canonical history with derived records, so readers
	// must observe the same local boundary as record writers and watcher appends.
	release, err := acquireHistoryRuntimeGate(args[1])
	if err != nil {
		return emit("BUSY", nil, err, 1)
	}
	defer release()
	records, err := LoadHistory(args[1])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	actions, err := loadActions(args[2], records)
	if err != nil && !(args[0] == "record" && os.IsNotExist(err)) {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if args[0] == "list" {
		return emit("VALID", actions, nil, 0)
	}
	raw, err := os.ReadFile(args[3])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	a, status := m03.DecodeM03Action(raw)
	if status == "VALID" {
		status = m03.ValidateHumanActionRecord(a)
	}
	if status != "VALID" {
		return emit(status, nil, fmt.Errorf("action rejected: %s", status), 1)
	}
	if err := resolveActionDecision(records, a.DecisionID); err != nil {
		return emit("DECISION_ERROR", nil, err, 1)
	}
	for _, old := range actions {
		if old.ActionID == a.ActionID {
			if old == a {
				return emit("EXACT_DUPLICATE", a, nil, 0)
			}
			return emit("CONFLICT", nil, fmt.Errorf("action_id reused with different content"), 1)
		}
	}
	encoded, err := json.Marshal(a)
	if err != nil {
		return emit("INVALID_SCHEMA", nil, err, 1)
	}
	if err := (store.JSONL{}).AppendLine(args[2], encoded); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	return emit("APPENDED", a, nil, 0)
}
