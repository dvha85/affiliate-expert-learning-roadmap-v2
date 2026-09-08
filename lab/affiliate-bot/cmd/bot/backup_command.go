package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

type backupManifest struct {
	Version  string            `json:"version"`
	Files    map[string]string `json:"files"`
	Required []string          `json:"required"`
}

func fileDigest(path string) (string, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", e
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:]), nil
}
func regularFile(path string) error {
	i, e := os.Lstat(path)
	if e != nil {
		return e
	}
	if !i.Mode().IsRegular() {
		return fmt.Errorf("%s must be a regular file", path)
	}
	return nil
}
func backupFiles(source string) ([]string, error) {
	out := []string{}
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.Name() == "manifest.json" || entry.IsDir() {
			continue
		}
		name := entry.Name()
		p := filepath.Join(source, name)
		if err := regularFile(p); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("runtime created no backup artifacts")
	}
	sort.Strings(out)
	return out, nil
}

func requiredBackupFiles(source string) ([]string, error) {
	required := []string{"history.jsonl", "mission-state.json"}
	state, err := loadMissionState(source)
	if err != nil {
		return nil, err
	}
	if state.Canary != nil {
		required = append(required, filepath.Base(m10ArtifactRegistryPath(source)))
		entries, err := loadM10ArtifactRegistry(source)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.ArtifactKind != corem10.ArtifactKindExecutionRecord {
				continue
			}
			record, err := corem10.ValidateExecutionRecord(entry.Artifact)
			if err != nil {
				return nil, err
			}
			if record.Status == "FAILED" {
				required = append(required, filepath.Base(m10OutcomeStorePath(source)))
				break
			}
		}
	}
	if _, err := os.Stat(m11ArtifactRegistryPath(source)); err == nil {
		required = append(required, filepath.Base(m11ArtifactRegistryPath(source)))
		entries, err := loadM11ArtifactRegistry(source)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.ArtifactKind != corem11.ArtifactKindExecution {
				continue
			}
			record, status := corem11.DecodeArtifact("execution", entry.Artifact)
			if status != corem11.Valid {
				return nil, fmt.Errorf("invalid M11 execution artifact")
			}
			if record.(*corem11.ProductionExecutionRecord).Status == "FAILED" {
				required = append(required, filepath.Base(m11OutcomeStorePath(source)))
				break
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	sort.Strings(required)
	return required, nil
}

func validateM10BackupGraph(dir string) error {
	state, err := loadMissionState(dir)
	if err != nil || state.Canary == nil {
		return err
	}
	entries, err := loadM10ArtifactRegistry(dir)
	if err != nil {
		return err
	}
	executions := map[string]corem10.ExecutionRecord{}
	for _, entry := range entries {
		if entry.ArtifactKind != corem10.ArtifactKindExecutionRecord {
			continue
		}
		record, err := corem10.ValidateExecutionRecord(entry.Artifact)
		if err != nil || !reservationForExecution(state, record.ExecutionID) {
			return fmt.Errorf("M10 execution record is orphaned from restored reservation")
		}
		executions[record.ExecutionID] = record
	}
	outcomes, err := loadM10FixtureOutcomes(dir, state)
	if err != nil {
		return err
	}
	failedOutcomeIDs := map[string]bool{}
	for _, outcome := range outcomes {
		failedOutcomeIDs[outcome.EffectRef.EffectID] = true
	}
	for executionID, record := range executions {
		if record.Status == "FAILED" && !failedOutcomeIDs[executionID] {
			return fmt.Errorf("failed execution is missing restored fixture outcome")
		}
	}
	return nil
}

func validateM11BackupGraph(dir string) error {
	entries, err := loadM11ArtifactRegistry(dir)
	if err != nil {
		return fmt.Errorf("M11 artifact graph is invalid: %w", err)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return fmt.Errorf("M11 outcome store is invalid: %w", err)
	}
	linked := map[string]bool{}
	ledgers := []corem11.ProductionLedger{}
	resolutions := map[string]corem11.ProductionReconciliationResolution{}
	executions := []corem11.ProductionExecutionRecord{}
	for _, outcome := range outcomes {
		linked[outcome.EffectRef.EffectID] = true
	}
	for _, entry := range entries {
		switch entry.ArtifactKind {
		case corem11.ArtifactKindLedger:
			value, status := corem11.DecodeArtifact("ledger", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 ledger artifact is invalid")
			}
			ledgers = append(ledgers, *value.(*corem11.ProductionLedger))
		case corem11.ArtifactKindReconciliation:
			value, status := corem11.DecodeArtifact("resolution", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 reconciliation artifact is invalid")
			}
			resolution := *value.(*corem11.ProductionReconciliationResolution)
			resolutions[resolution.ExecutionID] = resolution
		case corem11.ArtifactKindExecution:
			value, status := corem11.DecodeArtifact("execution", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 execution artifact is invalid")
			}
			executions = append(executions, *value.(*corem11.ProductionExecutionRecord))
		}
	}
	for _, record := range executions {
		if record.Status == "FAILED" && !linked[record.ExecutionID] {
			return fmt.Errorf("failed M11 execution is missing restored fixture outcome")
		}
		if record.Status != "RECONCILIATION_REQUIRED" || record.SideEffectState != "UNKNOWN" {
			continue
		}
		matched := false
		for _, ledger := range ledgers {
			if ledger.LeaseID != record.ProductionLeaseID || ledger.LeaseVersion != record.ProductionLeaseVersion || ledger.LeaseHash != record.ProductionLeaseHash || ledger.ControlMode != "STOPPED" {
				continue
			}
			if resolution, resolved := resolutions[record.ExecutionID]; resolved {
				if !ledger.ReconciliationRequired && ledger.StopReason == "RECOVERY_REVIEW_REQUIRED" {
					for _, id := range ledger.ReconciliationResolutionIDs {
						matched = matched || id == resolution.ResolutionID
					}
				}
			} else if ledger.ReconciliationRequired && ledger.StopReason == "RECONCILIATION_REQUIRED" {
				matched = true
			}
		}
		if !matched {
			return fmt.Errorf("unknown M11 execution lacks matching stopped reconciliation ledger")
		}
	}
	return nil
}

func sameFileSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func containsAll(files, required []string) bool {
	present := map[string]bool{}
	for _, file := range files {
		present[file] = true
	}
	for _, file := range required {
		if !present[file] {
			return false
		}
	}
	return true
}

func verifyBackup(dir string) (backupManifest, error) {
	var m backupManifest
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m); err != nil {
		return m, err
	}
	if m.Version != "affiliate-bot-backup/v2" || len(m.Files) == 0 {
		return m, fmt.Errorf("invalid backup manifest")
	}
	for name, want := range m.Files {
		if filepath.Base(name) != name {
			return m, fmt.Errorf("unsafe backup path")
		}
		p := filepath.Join(dir, name)
		if err := regularFile(p); err != nil {
			return m, err
		}
		got, e := fileDigest(p)
		if e != nil || got != want {
			return m, fmt.Errorf("backup checksum mismatch for %s", name)
		}
	}
	expectedRequired, err := requiredBackupFiles(dir)
	if err != nil {
		return m, fmt.Errorf("backup required artifact graph is invalid: %w", err)
	}
	sort.Strings(m.Required)
	if !sameFileSet(m.Required, expectedRequired) {
		return m, fmt.Errorf("backup required artifact inventory does not match runtime graph")
	}
	for _, required := range expectedRequired {
		if _, ok := m.Files[required]; !ok {
			return m, fmt.Errorf("backup is missing required %s", required)
		}
	}
	return m, nil
}

func runBackupCommand(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		o := map[string]any{"command": "backup", "status": status, "execution_permitted": false}
		if artifact != nil {
			o["artifact"] = artifact
		}
		if e := json.NewEncoder(stdout).Encode(o); e != nil {
			return 1
		}
		return code
	}
	if len(args) != 3 || (args[0] != "create" && args[0] != "restore") {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot backup create RUNTIME_DIR BACKUP_DIR | bot backup restore BACKUP_DIR EMPTY_TARGET_DIR"), 2)
	}
	if args[0] == "create" {
		files, e := backupFiles(args[1])
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		if e = os.MkdirAll(args[2], 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		required, e := requiredBackupFiles(args[1])
		if e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		if !containsAll(files, required) {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime is missing required backup artifacts"), 1)
		}
		if e = validateM10BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime M10 graph is invalid: %w", e), 1)
		}
		if e = validateM11BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		m := backupManifest{Version: "affiliate-bot-backup/v2", Files: map[string]string{}, Required: required}
		for _, name := range files {
			b, e := os.ReadFile(filepath.Join(args[1], name))
			if e != nil {
				return emit("INPUT_ERROR", nil, e, 1)
			}
			target := filepath.Join(args[2], name)
			if e = os.WriteFile(target, b, 0600); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
			m.Files[name], _ = fileDigest(target)
		}
		if e = writeJSON(filepath.Join(args[2], "manifest.json"), m); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		return emit("BACKED_UP", m, nil, 0)
	}
	m, e := verifyBackup(args[1])
	if e != nil {
		return emit("VERIFY_FAILED", nil, e, 1)
	}
	if _, e = os.Stat(args[2]); e == nil {
		entries, _ := os.ReadDir(args[2])
		if len(entries) > 0 {
			return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target must be empty"), 1)
		}
	} else if os.IsNotExist(e) {
		if e = os.MkdirAll(args[2], 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	} else {
		return emit("TARGET_ERROR", nil, e, 1)
	}
	for name := range m.Files {
		b, _ := os.ReadFile(filepath.Join(args[1], name))
		if e = os.WriteFile(filepath.Join(args[2], name), b, 0600); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	}
	if _, ok := m.Files["history.jsonl"]; ok {
		var records []HistoryRecord
		if records, e = LoadHistory(filepath.Join(args[2], "history.jsonl")); e != nil {
			return emit("REPLAY_FAILED", nil, e, 1)
		}
		for _, record := range records {
			if replay := Replay(record); replay.State != replayMatch {
				return emit("REPLAY_FAILED", nil, fmt.Errorf("record %s replayed as %s: %s", record.RecordID, replay.State, replay.Reason), 1)
			}
		}
	}
	if _, ok := m.Files["mission-state.json"]; ok {
		if _, e = loadMissionState(args[2]); e != nil {
			return emit("STATE_FAILED", nil, e, 1)
		}
	}
	if e = validateM10BackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM11BackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	return emit("RESTORED", m, nil, 0)
}
