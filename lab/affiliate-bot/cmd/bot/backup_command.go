package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

type backupManifest struct {
	Version  string            `json:"version"`
	Files    map[string]string `json:"files"`
	Required []string          `json:"required"`
}

// backupCopyFault is a test-only seam for an interrupted snapshot. Manifest
// publication remains after every copy succeeds, so a partial target never
// advertises itself as a restorable backup.
var backupCopyFault func(relativePath string) error

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
func backupRelativePath(name string) (string, error) {
	if name == "" || filepath.IsAbs(name) {
		return "", fmt.Errorf("unsafe backup path")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe backup path")
	}
	canonical := filepath.ToSlash(clean)
	if canonical != name {
		return "", fmt.Errorf("backup path is not normalized")
	}
	return clean, nil
}

func backupFiles(source string) ([]string, error) {
	if info, err := os.Lstat(source); err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("runtime source must be a non-symlink directory")
	}
	out := []string{}
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == source {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("backup does not follow symlink %s", path)
		}
		if entry.IsDir() {
			if entry.Name() == runtimeGateName {
				return filepath.SkipDir
			}
			if entry.Name() == ".mission.lock" {
				return fmt.Errorf("runtime has an active writer lock")
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("backup only accepts regular files")
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == "manifest.json" {
			return nil
		}
		if _, err := backupRelativePath(name); err != nil {
			return err
		}
		out = append(out, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("runtime created no backup artifacts")
	}
	sort.Strings(out)
	return out, nil
}

func requiredBackupFiles(source string) ([]string, error) {
	required := []string{"history.jsonl", "mission-state.json"}
	m00ToM05Files, err := m00ToM05BackupFiles(source)
	if err != nil {
		return nil, err
	}
	required = append(required, m00ToM05Files...)
	m07Files, err := m07BackupFiles(source)
	if err != nil {
		return nil, err
	}
	required = append(required, m07Files...)
	accesstradeReceiptRequired, err := accesstradeBackupReceiptRequired(source)
	if err != nil {
		return nil, err
	}
	if accesstradeReceiptRequired {
		required = append(required, filepath.Base(accesstradeReceiptPath(filepath.Join(source, "outcomes.jsonl"))))
	}
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

func optionalRuntimeFile(dir, name string) (bool, error) {
	info, err := os.Lstat(filepath.Join(dir, name))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return false, fmt.Errorf("runtime artifact %s must be a regular file", name)
	}
	return true, nil
}

// m00ToM05BackupFiles resolves every optional M03–M05 store from the prior
// store. A backup may legitimately contain history alone, but once a later
// store exists its complete upstream lineage must be present and replayable.
func m00ToM05BackupFiles(dir string) ([]string, error) {
	history, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return nil, err
	}
	files := []string{}
	hasActions, err := optionalRuntimeFile(dir, "actions.jsonl")
	if err != nil {
		return nil, err
	}
	hasOutcomes, err := optionalRuntimeFile(dir, "outcomes.jsonl")
	if err != nil {
		return nil, err
	}
	hasEvaluations, err := optionalRuntimeFile(dir, "evaluations.jsonl")
	if err != nil {
		return nil, err
	}
	hasProposals, err := optionalRuntimeFile(dir, "proposals.jsonl")
	if err != nil {
		return nil, err
	}
	hasReviews, err := optionalRuntimeFile(dir, "reviews.jsonl")
	if err != nil {
		return nil, err
	}
	if (hasOutcomes || hasEvaluations || hasProposals || hasReviews) && !hasActions {
		return nil, fmt.Errorf("M03+ store requires actions.jsonl")
	}
	if (hasEvaluations || hasProposals || hasReviews) && !hasOutcomes {
		return nil, fmt.Errorf("M05 store requires outcomes.jsonl")
	}
	if (hasProposals || hasReviews) && !hasEvaluations {
		return nil, fmt.Errorf("improvement store requires evaluations.jsonl")
	}
	if hasReviews && !hasProposals {
		return nil, fmt.Errorf("review store requires proposals.jsonl")
	}
	if !hasActions {
		return files, nil
	}
	actions, err := loadActions(filepath.Join(dir, "actions.jsonl"), history)
	if err != nil {
		return nil, err
	}
	files = append(files, "actions.jsonl")
	if !hasOutcomes {
		return files, nil
	}
	outcomes, err := loadOutcomes(filepath.Join(dir, "outcomes.jsonl"), actions)
	if err != nil {
		return nil, err
	}
	files = append(files, "outcomes.jsonl")
	if !hasEvaluations {
		return files, nil
	}
	evaluations, err := loadEvaluations(filepath.Join(dir, "evaluations.jsonl"), history, actions, outcomes)
	if err != nil {
		return nil, err
	}
	files = append(files, "evaluations.jsonl")
	if !hasProposals {
		return files, nil
	}
	decodeProposal := func(raw []byte) (m05.ImprovementProposal, error) { return linkedProposal(raw, evaluations) }
	proposals, err := loadImprovementRecords(filepath.Join(dir, "proposals.jsonl"), decodeProposal, func(p m05.ImprovementProposal) string { return p.ProposalID })
	if err != nil {
		return nil, err
	}
	files = append(files, "proposals.jsonl")
	if !hasReviews {
		return files, nil
	}
	_, err = loadImprovementRecords(filepath.Join(dir, "reviews.jsonl"), func(raw []byte) (m05.ReviewRecord, error) { return linkedReview(raw, proposals, evaluations) }, func(r m05.ReviewRecord) string { return r.ReviewID })
	if err != nil {
		return nil, err
	}
	return append(files, "reviews.jsonl"), nil
}

// m07BackupFiles lists durable adapter artifacts only when history is stored
// inside this runtime root. A history sidecar outside the root is not silently
// skipped: it cannot be restored as part of this runtime snapshot.
func m07BackupFiles(source string) ([]string, error) {
	history := filepath.Join(source, "history.jsonl")
	sidecar := filepath.Clean(history) + ".m07"
	info, err := os.Lstat(sidecar)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("M07 artifact sidecar must be a non-symlink directory")
	}
	files := []string{}
	err = filepath.WalkDir(sidecar, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sidecar {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || entry.IsDir() && entry.Name() == ".mission.lock" {
			return fmt.Errorf("unsafe M07 artifact path")
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("M07 artifact must be a regular file")
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if _, err := backupRelativePath(name); err != nil {
			return err
		}
		files = append(files, name)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func validM07ArtifactFile(kind, name, id string) bool {
	if kind != "tool-results" && kind != "proposals" || len(id) != 64 || name != id+".json" {
		return false
	}
	for _, r := range id {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	return true
}

func validateM07BackupGraph(dir string) error {
	sidecar := filepath.Join(dir, "history.jsonl.m07")
	if _, err := os.Lstat(sidecar); os.IsNotExist(err) {
		state, stateErr := loadMissionState(dir)
		if stateErr != nil {
			return stateErr
		}
		if state.Intent != nil && state.Intent.ProposedBy == "agent" {
			return fmt.Errorf("agent intent is missing its persisted M07 proposal")
		}
		return nil
	} else if err != nil {
		return err
	}
	records, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return err
	}
	byID := map[string]HistoryRecord{}
	for _, record := range records {
		byID[record.RecordID] = record
	}
	proposalPaths := []string{}
	toolEvidence := map[string][]corem07.Evidence{}
	err = filepath.WalkDir(sidecar, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == sidecar {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("unsafe M07 artifact")
		}
		rel, err := filepath.Rel(sidecar, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if entry.IsDir() {
			if len(parts) != 1 || parts[0] != "tool-results" && parts[0] != "proposals" {
				return fmt.Errorf("invalid M07 artifact layout")
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsafe M07 artifact")
		}
		if len(parts) != 2 || !validM07ArtifactFile(parts[0], parts[1], strings.TrimSuffix(parts[1], ".json")) {
			return fmt.Errorf("invalid M07 artifact layout")
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if parts[0] == "tool-results" {
			var stored corem07.RegisteredToolResult
			if err := json.Unmarshal(raw, &stored); err != nil || !strings.HasPrefix(stored.TraceID, "sha256:") || strings.TrimPrefix(stored.TraceID, "sha256:") != strings.TrimSuffix(parts[1], ".json") {
				return fmt.Errorf("M07 tool result filename does not bind its trace")
			}
			registered, err := corem07.ValidateStoredRegisteredToolResult(raw, stored.RecordID)
			if err != nil || byID[registered.RecordID].RecordID == "" {
				return fmt.Errorf("M07 tool result is not valid for canonical history")
			}
			toolEvidence[registered.RecordID] = append(toolEvidence[registered.RecordID], registered.Evidence())
			return nil
		}
		proposalPaths = append(proposalPaths, path)
		return nil
	})
	if err != nil {
		return err
	}
	proposals := map[string]corem07.AgentOutput{}
	for _, path := range proposalPaths {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var stored corem07.RegisteredAgentProposal
		if err := json.Unmarshal(raw, &stored); err != nil || !strings.HasPrefix(stored.ProposalID, "sha256:") || strings.TrimPrefix(stored.ProposalID, "sha256:") != strings.TrimSuffix(filepath.Base(path), ".json") {
			return fmt.Errorf("M07 proposal filename does not bind its proposal")
		}
		record, ok := byID[stored.RecordID]
		if !ok {
			return fmt.Errorf("M07 proposal references missing canonical history")
		}
		ctx, err := m07EvidenceContext(record)
		if err != nil {
			return err
		}
		ctx.Evidence = append(ctx.Evidence, toolEvidence[record.RecordID]...)
		proposal, output, err := corem07.ValidateRegisteredAgentProposal(raw, ctx.Evidence, nil, record.RecordID)
		if err != nil {
			return fmt.Errorf("M07 proposal is not grounded after restore: %w", err)
		}
		proposals[proposal.ProposalID] = output
	}
	state, err := loadMissionState(dir)
	if err != nil {
		return err
	}
	if state.Intent == nil || state.Intent.ProposedBy != "agent" {
		return nil
	}
	output, ok := proposals[state.Intent.ProposalRef]
	if !ok || output.ProposedAction == nil || output.ProposedAction.ActionType != state.Intent.ActionType || output.ProposedAction.Target != state.Intent.Target || !sameParameters(output.ProposedAction.Parameters, state.Intent.Parameters) {
		return fmt.Errorf("agent intent does not resolve to restored M07 proposal")
	}
	for _, id := range state.Intent.EvidenceIDs {
		found := false
		for _, proposalID := range output.EvidenceIDs {
			found = found || id == proposalID
		}
		if !found {
			return fmt.Errorf("agent intent has evidence absent from restored M07 proposal")
		}
	}
	return nil
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
		clean, err := backupRelativePath(name)
		if err != nil {
			return m, err
		}
		p := filepath.Join(dir, clean)
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
	if err := validateM07BackupGraph(dir); err != nil {
		return m, fmt.Errorf("backup M07 graph is invalid: %w", err)
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
		source, sourceErr := filepath.Abs(args[1])
		target, targetErr := filepath.Abs(args[2])
		if sourceErr != nil || targetErr != nil || source == target || strings.HasPrefix(target, source+string(filepath.Separator)) {
			return emit("INPUT_ERROR", nil, fmt.Errorf("backup target must not be the runtime or a child of it"), 1)
		}
		if info, statErr := os.Lstat(args[2]); statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return emit("TARGET_ERROR", nil, fmt.Errorf("backup target must be a non-symlink directory"), 1)
			}
			entries, readErr := os.ReadDir(args[2])
			if readErr != nil {
				return emit("TARGET_ERROR", nil, readErr, 1)
			}
			if len(entries) != 0 {
				return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("backup target must be empty"), 1)
			}
		} else if !os.IsNotExist(statErr) {
			return emit("TARGET_ERROR", nil, statErr, 1)
		}
		release, lockErr := acquireRuntimeGate(args[1])
		if lockErr != nil {
			return emit("BUSY", nil, lockErr, 1)
		}
		defer release()
		if e := validateAccesstradeBackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime ACCESSTRADE receipt graph is invalid: %w", e), 1)
		}
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
		if e = validateM07BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, fmt.Errorf("runtime M07 graph is invalid: %w", e), 1)
		}
		if e = validateM11BackupGraph(args[1]); e != nil {
			return emit("INPUT_ERROR", nil, e, 1)
		}
		m := backupManifest{Version: "affiliate-bot-backup/v2", Files: map[string]string{}, Required: required}
		for _, name := range files {
			if backupCopyFault != nil {
				if e = backupCopyFault(name); e != nil {
					return emit("STORE_ERROR", nil, e, 1)
				}
			}
			b, e := os.ReadFile(filepath.Join(args[1], name))
			if e != nil {
				return emit("INPUT_ERROR", nil, e, 1)
			}
			clean, e := backupRelativePath(name)
			if e != nil {
				return emit("INPUT_ERROR", nil, e, 1)
			}
			target := filepath.Join(args[2], clean)
			if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
				return emit("STORE_ERROR", nil, e, 1)
			}
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
	if info, statErr := os.Lstat(args[2]); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return emit("TARGET_ERROR", nil, fmt.Errorf("restore target must be a non-symlink directory"), 1)
		}
		entries, _ := os.ReadDir(args[2])
		if len(entries) > 0 {
			return emit("TARGET_NOT_EMPTY", nil, fmt.Errorf("restore target must be empty"), 1)
		}
	} else if os.IsNotExist(statErr) {
		if e = os.MkdirAll(args[2], 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
	} else {
		return emit("TARGET_ERROR", nil, statErr, 1)
	}
	for name := range m.Files {
		clean, pathErr := backupRelativePath(name)
		if pathErr != nil {
			return emit("VERIFY_FAILED", nil, pathErr, 1)
		}
		b, _ := os.ReadFile(filepath.Join(args[1], clean))
		target := filepath.Join(args[2], clean)
		if e = os.MkdirAll(filepath.Dir(target), 0700); e != nil {
			return emit("STORE_ERROR", nil, e, 1)
		}
		if e = os.WriteFile(target, b, 0600); e != nil {
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
	if _, e = m00ToM05BackupFiles(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM10BackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM07BackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateM11BackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, e, 1)
	}
	if e = validateAccesstradeBackupGraph(args[2]); e != nil {
		return emit("GRAPH_FAILED", nil, fmt.Errorf("restored ACCESSTRADE receipt graph is invalid: %w", e), 1)
	}
	return emit("RESTORED", m, nil, 0)
}
