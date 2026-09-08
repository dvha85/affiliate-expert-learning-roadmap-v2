package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

const accesstradeReceiptVersion = "accesstrade-import-receipt/v1"

// AccesstradeImportReceipt stores replayable provenance without retaining the
// CSV's order identifiers, customer data, or a private tracking URL.
type AccesstradeImportReceipt struct {
	Version          string   `json:"version"`
	ReceiptID        string   `json:"receipt_id"`
	SnapshotID       string   `json:"snapshot_id"`
	SourceRef        string   `json:"source_ref"`
	ObservedAt       string   `json:"observed_at"`
	Currency         string   `json:"currency"`
	ReportSHA256     string   `json:"report_sha256"`
	ManifestSHA256   string   `json:"manifest_sha256"`
	OutcomeStoreName string   `json:"outcome_store_name"`
	OutcomeIDs       []string `json:"outcome_ids"`
	ActionIDs        []string `json:"action_ids"`
}

func sameReceipt(left, right AccesstradeImportReceipt) bool { return reflect.DeepEqual(left, right) }

func accesstradeReceiptPath(outcomesPath string) string {
	return filepath.Join(filepath.Dir(outcomesPath), "accesstrade-receipts.jsonl")
}

func accesstradeJournalPath(outcomesPath string) string {
	return filepath.Join(filepath.Dir(outcomesPath), "accesstrade-import.pending.json")
}

func sha256Reference(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func sortedUnique(values []string) []string {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	out := make([]string, 0, len(set))
	for value := range set {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func newAccesstradeReceipt(manifest AccesstradeReportManifest, report, rawManifest []byte, outcomesPath string, outcomes []m03.OutcomeRecord) AccesstradeImportReceipt {
	outcomeIDs, actionIDs := make([]string, 0, len(outcomes)), make([]string, 0, len(outcomes))
	for _, outcome := range outcomes {
		outcomeIDs = append(outcomeIDs, outcome.OutcomeID)
		actionIDs = append(actionIDs, outcome.EffectRef.EffectID)
	}
	return AccesstradeImportReceipt{
		Version:          accesstradeReceiptVersion,
		ReceiptID:        "accesstrade:" + manifest.SnapshotID,
		SnapshotID:       manifest.SnapshotID,
		SourceRef:        manifest.SourceRef,
		ObservedAt:       manifest.ObservedAt,
		Currency:         manifest.Currency,
		ReportSHA256:     sha256Reference(report),
		ManifestSHA256:   sha256Reference(rawManifest),
		OutcomeStoreName: filepath.Base(outcomesPath),
		OutcomeIDs:       sortedUnique(outcomeIDs),
		ActionIDs:        sortedUnique(actionIDs),
	}
}

func validSHA256Reference(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func validateAccesstradeReceipt(receipt AccesstradeImportReceipt) error {
	if receipt.Version != accesstradeReceiptVersion || strings.TrimSpace(receipt.SnapshotID) == "" || receipt.ReceiptID != "accesstrade:"+receipt.SnapshotID || !strings.HasPrefix(receipt.SourceRef, "accesstrade:") || receipt.Currency != "VND" {
		return fmt.Errorf("invalid ACCESSTRADE receipt identity")
	}
	if _, err := parseRFC3339(receipt.ObservedAt); err != nil {
		return fmt.Errorf("invalid ACCESSTRADE receipt observed_at: %w", err)
	}
	if !validSHA256Reference(receipt.ReportSHA256) || !validSHA256Reference(receipt.ManifestSHA256) || filepath.Base(receipt.OutcomeStoreName) != receipt.OutcomeStoreName || receipt.OutcomeStoreName == "." {
		return fmt.Errorf("invalid ACCESSTRADE receipt provenance")
	}
	if len(receipt.OutcomeIDs) == 0 || len(receipt.ActionIDs) == 0 || !sameStringSet(receipt.OutcomeIDs, sortedUnique(receipt.OutcomeIDs)) || !sameStringSet(receipt.ActionIDs, sortedUnique(receipt.ActionIDs)) {
		return fmt.Errorf("invalid ACCESSTRADE receipt links")
	}
	return nil
}

func sameStringSet(left, right []string) bool {
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

func loadAccesstradeReceipts(path string) ([]AccesstradeImportReceipt, error) {
	f, err := (store.JSONL{}).Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	receipts := []AccesstradeImportReceipt{}
	seen := map[string]bool{}
	for scanner.Scan() {
		var receipt AccesstradeImportReceipt
		decoder := json.NewDecoder(strings.NewReader(string(scanner.Bytes())))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&receipt); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
			return nil, fmt.Errorf("invalid ACCESSTRADE receipt")
		}
		if err := validateAccesstradeReceipt(receipt); err != nil || seen[receipt.ReceiptID] {
			return nil, fmt.Errorf("invalid or duplicate ACCESSTRADE receipt")
		}
		seen[receipt.ReceiptID] = true
		receipts = append(receipts, receipt)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return receipts, nil
}

func validateAccesstradeReceiptGraph(receipts []AccesstradeImportReceipt, outcomes []m03.OutcomeRecord, outcomesPath string) error {
	byOutcome := map[string]m03.OutcomeRecord{}
	for _, outcome := range outcomes {
		byOutcome[outcome.OutcomeID] = outcome
	}
	claimed := map[string]bool{}
	for _, receipt := range receipts {
		if receipt.OutcomeStoreName != filepath.Base(outcomesPath) {
			return fmt.Errorf("ACCESSTRADE receipt is bound to another outcome store")
		}
		linkedActionIDs := make([]string, 0, len(receipt.OutcomeIDs))
		for _, outcomeID := range receipt.OutcomeIDs {
			outcome, ok := byOutcome[outcomeID]
			if !ok || claimed[outcomeID] || outcome.SourceRef != receipt.SourceRef || outcome.EffectRef.EffectKind != "HUMAN_ACTION" {
				return fmt.Errorf("ACCESSTRADE receipt has an orphaned outcome")
			}
			claimed[outcomeID] = true
			linkedActionIDs = append(linkedActionIDs, outcome.EffectRef.EffectID)
		}
		if !sameStringSet(receipt.ActionIDs, sortedUnique(linkedActionIDs)) {
			return fmt.Errorf("ACCESSTRADE receipt has an orphaned action")
		}
	}
	for _, outcome := range outcomes {
		if strings.HasPrefix(outcome.SourceRef, "accesstrade:") && !claimed[outcome.OutcomeID] {
			return fmt.Errorf("ACCESSTRADE outcome is missing a receipt")
		}
	}
	return nil
}

func receiptJournalPresent(outcomesPath string) error {
	if _, err := os.Lstat(accesstradeJournalPath(outcomesPath)); err == nil {
		return fmt.Errorf("ACCESSTRADE import recovery required")
	} else if !os.IsNotExist(err) {
		return err
	}
	return nil
}

func validateAccesstradeReceiptStore(outcomesPath string, outcomes []m03.OutcomeRecord) ([]AccesstradeImportReceipt, error) {
	if err := receiptJournalPresent(outcomesPath); err != nil {
		return nil, err
	}
	receipts, err := loadAccesstradeReceipts(accesstradeReceiptPath(outcomesPath))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := validateAccesstradeReceiptGraph(receipts, outcomes, outcomesPath); err != nil {
		return nil, err
	}
	return receipts, nil
}

// accesstradeBackupReceiptRequired derives the receipt requirement from the
// canonical outcome graph, rather than trusting a caller-supplied inventory.
// A receipt file on its own is also material: it must not be silently omitted
// from a backup even if its neighbouring outcome store has been damaged.
func accesstradeBackupReceiptRequired(dir string) (bool, error) {
	outcomesPath := filepath.Join(dir, "outcomes.jsonl")
	if err := receiptJournalPresent(outcomesPath); err != nil {
		return false, err
	}
	_, receiptErr := os.Lstat(accesstradeReceiptPath(outcomesPath))
	hasReceipt := receiptErr == nil
	if receiptErr != nil && !os.IsNotExist(receiptErr) {
		return false, receiptErr
	}
	if _, err := os.Lstat(outcomesPath); err != nil {
		if os.IsNotExist(err) {
			if hasReceipt {
				return true, fmt.Errorf("ACCESSTRADE receipt exists without outcomes.jsonl")
			}
			return false, nil
		}
		return false, err
	}
	if err := regularFile(outcomesPath); err != nil {
		return false, err
	}
	raw, err := os.ReadFile(outcomesPath)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		outcome, status := m03.DecodeM03Outcome([]byte(line))
		if status != "VALID" {
			return false, fmt.Errorf("invalid outcome while deriving ACCESSTRADE receipt requirement")
		}
		if strings.HasPrefix(outcome.SourceRef, "accesstrade:") {
			return true, nil
		}
	}
	return hasReceipt, nil
}

// validateAccesstradeBackupGraph is intentionally separate from the generic
// outcome CLI: the backup boundary must reject a package that has an
// ACCESSTRADE outcome but no matching provenance receipt, even if its manifest
// was manually edited to hide that missing file.
func validateAccesstradeBackupGraph(dir string) error {
	required, err := accesstradeBackupReceiptRequired(dir)
	if err != nil || !required {
		return err
	}
	history, err := LoadHistory(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return err
	}
	actions, err := loadActions(filepath.Join(dir, "actions.jsonl"), history)
	if err != nil {
		return err
	}
	outcomes, err := loadOutcomes(filepath.Join(dir, "outcomes.jsonl"), actions)
	if err != nil {
		return err
	}
	_, err = validateAccesstradeReceiptStore(filepath.Join(dir, "outcomes.jsonl"), outcomes)
	return err
}

func runAccesstradeReceiptList(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		envelope := map[string]any{"command": "outcome accesstrade-receipts", "status": status, "execution_permitted": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			return 1
		}
		return code
	}
	if len(args) != 4 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot outcome accesstrade-receipts HISTORY ACTIONS OUTCOMES"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
	history, err := LoadHistory(args[1])
	if err != nil {
		return emit("HISTORY_ERROR", nil, err, 1)
	}
	actions, err := loadActions(args[2], history)
	if err != nil {
		return emit("ACTION_STORE_ERROR", nil, err, 1)
	}
	outcomes, err := loadOutcomes(args[3], actions)
	if err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	receipts, err := validateAccesstradeReceiptStore(args[3], outcomes)
	if err != nil {
		return emit("RECEIPT_STORE_ERROR", nil, err, 1)
	}
	return emit("VALID", receipts, nil, 0)
}
