package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

func setupAccesstradeImport(t *testing.T) (string, string, string, string) {
	t.Helper()
	dir := t.TempDir()
	history, actions, outcomes, input := filepath.Join(dir, "history"), filepath.Join(dir, "actions"), filepath.Join(dir, "outcomes"), filepath.Join(dir, "input")
	observations, err := loadHistoryObservations("../../data/m02-sample-observations.json")
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewHistoryRecord("decision-1", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z", observations)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(history, record); err != nil {
		t.Fatal(err)
	}
	action := m03.HumanActionRecord{ActionID: "manual-action-001", DecisionID: "decision-1", ActionType: "manual_fixture", Target: "fixture:no-publication", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}
	raw, _ := json.Marshal(action)
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if runActionStore([]string{"record", history, actions, input}, &stdout, &stderr) != 0 {
		t.Fatal(stderr.String())
	}
	return history, actions, outcomes, dir
}

func writeAccesstradeImportInputs(t *testing.T, dir string) (string, string) {
	t.Helper()
	report := filepath.Join(dir, "report.csv")
	manifest := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Tạm duyệt,250.000,17.500\nredacted-2,Từ chối,150.000,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(`{"snapshot_id":"snapshot-1","observed_at":"2026-09-08T12:00:00+07:00","source_ref":"accesstrade:sanitized:snapshot-1","currency":"VND","mappings":[{"order_id":"redacted-1","outcome_id":"outcome-1","action_id":"manual-action-001"},{"order_id":"redacted-2","outcome_id":"outcome-2","action_id":"manual-action-001"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	return report, manifest
}

func TestAccesstradeImporterLifecycle(t *testing.T) {
	history, actions, outcomes, dir := setupAccesstradeImport(t)
	report, manifest := writeAccesstradeImportInputs(t, dir)
	invoke := func(want string, code int) map[string]any {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if got := runAccesstradeOutcomeImport([]string{"accesstrade-import", history, actions, outcomes, report, manifest}, &stdout, &stderr); got != code {
			t.Fatalf("code=%d stderr=%s", got, stderr.String())
		}
		var result map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil || result["status"] != want || result["execution_permitted"] != false {
			t.Fatalf("result=%s err=%v", stdout.String(), err)
		}
		return result
	}
	invoke("APPENDED", 0)
	invoke("EXACT_DUPLICATE", 0)
	var receiptOut, receiptErr bytes.Buffer
	if got := runOutcomeStore([]string{"accesstrade-receipts", history, actions, outcomes}, &receiptOut, &receiptErr); got != 0 {
		t.Fatalf("receipt list code=%d stderr=%s", got, receiptErr.String())
	}
	var receipts struct {
		Status   string                     `json:"status"`
		Artifact []AccesstradeImportReceipt `json:"artifact"`
	}
	if err := json.Unmarshal(receiptOut.Bytes(), &receipts); err != nil || receipts.Status != "VALID" || len(receipts.Artifact) != 1 || receipts.Artifact[0].ReceiptID != "accesstrade:snapshot-1" || !validSHA256Reference(receipts.Artifact[0].ReportSHA256) {
		t.Fatalf("receipt list=%s err=%v", receiptOut.String(), err)
	}
	receiptPath := accesstradeReceiptPath(outcomes)
	originalReceipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	tamperedReceipt := receipts.Artifact[0]
	tamperedReceipt.ActionIDs = append(tamperedReceipt.ActionIDs, "unlinked-action")
	tamperedRaw, _ := json.Marshal(tamperedReceipt)
	if err := os.WriteFile(receiptPath, append(tamperedRaw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if got := runOutcomeStore([]string{"accesstrade-receipts", history, actions, outcomes}, &receiptOut, &receiptErr); got != 1 {
		t.Fatalf("tampered receipt code=%d", got)
	}
	if err := os.WriteFile(receiptPath, originalReceipt, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadOutcomes(outcomes, []m03.HumanActionRecord{{ActionID: "manual-action-001", DecisionID: "decision-1", ActionType: "manual_fixture", Target: "fixture:no-publication", PerformedBy: "human", PerformedAt: "2026-09-04T00:00:00Z", MeasurementWindowEnd: "2026-09-05T00:00:00Z", ComplianceReviewed: true}})
	if err != nil || len(loaded) != 2 {
		t.Fatal(err)
	}
	if loaded[0].Status != "PENDING" || loaded[0].Metrics["order_value_vnd"] != 250000 || loaded[0].Metrics["reported_commission_vnd"] != 17500 || loaded[1].Status != "CANCELLED" || loaded[1].Metrics["reported_commission_vnd"] != 0 {
		t.Fatal(loaded)
	}
	before, err := os.ReadFile(outcomes)
	if err != nil {
		t.Fatal(err)
	}
	// Canonically equal amounts still do not make a changed source CSV an exact
	// snapshot retry: its receipt hash must bind the bytes that were reviewed.
	if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Tạm duyệt,250000,17500\nredacted-2,Từ chối,150000,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	invoke("RECEIPT_CONFLICT", 1)
	if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Đã thanh toán,250.000,17.500\nredacted-2,Từ chối,150.000,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	invoke("REPORT_ERROR", 1)
	after, _ := os.ReadFile(outcomes)
	if !bytes.Equal(before, after) {
		t.Fatal("rejected report changed outcome store")
	}
	if err := os.WriteFile(accesstradeJournalPath(outcomes), []byte(`{"interrupted":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	invoke("RECEIPT_STORE_ERROR", 1)
	if err := os.Remove(accesstradeJournalPath(outcomes)); err != nil {
		t.Fatal(err)
	}
}

func TestAccesstradeImporterRecoveryReplaysOnlyExactPendingSnapshot(t *testing.T) {
	invoke := func(t *testing.T, args []string, want string, code int) map[string]any {
		t.Helper()
		var stdout, stderr bytes.Buffer
		if got := runOutcomeStore(args, &stdout, &stderr); got != code {
			t.Fatalf("code=%d want=%d stdout=%s stderr=%s", got, code, stdout.String(), stderr.String())
		}
		var response map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &response); err != nil || response["status"] != want || response["execution_permitted"] != false {
			t.Fatalf("response=%s err=%v", stdout.String(), err)
		}
		return response
	}
	setupPending := func(t *testing.T) (string, string, string, string, string, []m03.OutcomeRecord, AccesstradeImportReceipt) {
		t.Helper()
		history, actionsPath, outcomes, dir := setupAccesstradeImport(t)
		report, manifest := writeAccesstradeImportInputs(t, dir)
		historyRecords, err := LoadHistory(history)
		if err != nil {
			t.Fatal(err)
		}
		actions, err := loadActions(actionsPath, historyRecords)
		if err != nil {
			t.Fatal(err)
		}
		candidates, receipt, err := loadAccesstradeImportInputs([]string{"accesstrade-recover", history, actionsPath, outcomes, report, manifest}, actions)
		if err != nil {
			t.Fatal(err)
		}
		if err := writeJSONAtomic(accesstradeJournalPath(outcomes), receipt); err != nil {
			t.Fatal(err)
		}
		return history, actionsPath, outcomes, report, manifest, candidates, receipt
	}

	t.Run("empty snapshot is completed and journal is removed", func(t *testing.T) {
		history, actions, outcomes, report, manifest, _, _ := setupPending(t)
		args := []string{"accesstrade-recover", history, actions, outcomes, report, manifest}
		response := invoke(t, args, "RECOVERED", 0)
		if response["artifact"] == nil {
			t.Fatal("recovery did not return the exact recovered artifact")
		}
		if _, err := os.Lstat(accesstradeJournalPath(outcomes)); !os.IsNotExist(err) {
			t.Fatal("pending journal remains after complete recovery", err)
		}
		invoke(t, []string{"accesstrade-import", history, actions, outcomes, report, manifest}, "EXACT_DUPLICATE", 0)
	})

	t.Run("visible outcomes without receipt are completed without rewriting outcomes", func(t *testing.T) {
		history, actions, outcomes, report, manifest, candidates, _ := setupPending(t)
		encoded, err := encodeAccesstradeOutcomes(candidates)
		if err != nil {
			t.Fatal(err)
		}
		if err := appendOutcomesAtomically(outcomes, encoded); err != nil {
			t.Fatal(err)
		}
		before, err := os.ReadFile(outcomes)
		if err != nil {
			t.Fatal(err)
		}
		invoke(t, []string{"accesstrade-recover", history, actions, outcomes, report, manifest}, "RECOVERED", 0)
		after, err := os.ReadFile(outcomes)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("recovery rewrote visible exact outcomes: err=%v", err)
		}
	})

	t.Run("changed report and partial outcomes fail closed", func(t *testing.T) {
		history, actions, outcomes, report, manifest, candidates, _ := setupPending(t)
		journalBefore, err := os.ReadFile(accesstradeJournalPath(outcomes))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Tạm duyệt,250000,17500\nredacted-2,Từ chối,150000,0\n"), 0600); err != nil {
			t.Fatal(err)
		}
		invoke(t, []string{"accesstrade-recover", history, actions, outcomes, report, manifest}, "REPLAY_MISMATCH", 1)
		if current, err := os.ReadFile(accesstradeJournalPath(outcomes)); err != nil || !bytes.Equal(current, journalBefore) {
			t.Fatalf("mismatch changed pending journal: err=%v", err)
		}
		if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Tạm duyệt,250.000,17.500\nredacted-2,Từ chối,150.000,0\n"), 0600); err != nil {
			t.Fatal(err)
		}
		encoded, err := encodeAccesstradeOutcomes(candidates[:1])
		if err != nil {
			t.Fatal(err)
		}
		if err := appendOutcomesAtomically(outcomes, encoded); err != nil {
			t.Fatal(err)
		}
		before, err := os.ReadFile(outcomes)
		if err != nil {
			t.Fatal(err)
		}
		invoke(t, []string{"accesstrade-recover", history, actions, outcomes, report, manifest}, "PARTIAL_DUPLICATE", 1)
		if after, err := os.ReadFile(outcomes); err != nil || !bytes.Equal(before, after) {
			t.Fatalf("partial recovery changed outcomes: err=%v", err)
		}
		if _, err := os.Lstat(accesstradeJournalPath(outcomes)); err != nil {
			t.Fatal("partial recovery removed pending journal", err)
		}
	})
}

func TestAccesstradeImporterDisclosesVisibleJournalCleanupUncertainty(t *testing.T) {
	history, actions, outcomes, dir := setupAccesstradeImport(t)
	report, manifest := writeAccesstradeImportInputs(t, dir)
	args := []string{"accesstrade-import", history, actions, outcomes, report, manifest}
	accesstradeJournalCleanupFault = func(phase string) error {
		if phase == "after_remove_before_parent_sync" {
			return errors.New("simulated ACCESSTRADE journal parent-sync acknowledgement loss")
		}
		return nil
	}
	t.Cleanup(func() { accesstradeJournalCleanupFault = nil })
	var stdout, stderr bytes.Buffer
	if code := runOutcomeStore(args, &stdout, &stderr); code != 1 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var response map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil || response["status"] != "PUBLISHED_RECOVERY_REQUIRED" || response["artifact"] == nil || response["execution_permitted"] != false {
		t.Fatalf("response=%s stderr=%s err=%v", stdout.String(), stderr.String(), err)
	}
	if _, err := os.Lstat(accesstradeJournalPath(outcomes)); !os.IsNotExist(err) {
		t.Fatalf("journal cleanup was not visible: %v", err)
	}
	var receiptOut, receiptErr bytes.Buffer
	if code := runOutcomeStore([]string{"accesstrade-receipts", history, actions, outcomes}, &receiptOut, &receiptErr); code != 0 {
		t.Fatalf("visible stores do not resolve: code=%d stderr=%s", code, receiptErr.String())
	}
	accesstradeJournalCleanupFault = nil
	stdout.Reset()
	stderr.Reset()
	if code := runOutcomeStore(args, &stdout, &stderr); code != 0 {
		t.Fatalf("exact retry code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	response = map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil || response["status"] != "EXACT_DUPLICATE" || response["artifact"] == nil {
		t.Fatalf("exact retry=%s err=%v", stdout.String(), err)
	}
}

func TestAccesstradeImporterDoesNotMapUTM(t *testing.T) {
	manifest := AccesstradeReportManifest{SnapshotID: "snapshot", ObservedAt: "2026-09-08T00:00:00Z", SourceRef: "fixture", Currency: "VND", Mappings: []AccesstradeReportMapping{{OrderID: "order-1", OutcomeID: "outcome-1", ActionID: "action-1"}}}
	if _, err := decodeAccesstradeOutcomes([]byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng,UTM Source\norder-1,Tạm duyệt,1,0,action-1\n"), manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeAccesstradeOutcomes([]byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng,UTM Source\norder-2,Tạm duyệt,1,0,action-1\n"), manifest); err == nil {
		t.Fatal("UTM source was accepted as an implicit action mapping")
	}
}

func TestAccesstradeBackupRequirementRejectsIncompleteOutcomeJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "outcomes.jsonl")
	original := []byte(`{}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := accesstradeBackupReceiptRequired(path); err == nil {
		t.Fatal("backup receipt requirement accepted an unterminated outcome record")
	}
	if current, err := os.ReadFile(path); err != nil || !bytes.Equal(current, original) {
		t.Fatalf("framing rejection changed outcome store: %q err=%v", current, err)
	}
}

func TestAccesstradeManifestRequiresBoundedSourceRef(t *testing.T) {
	_, err := decodeAccesstradeManifest([]byte(`{"snapshot_id":"snapshot","observed_at":"2026-09-08T00:00:00Z","source_ref":"fixture:unbound","currency":"VND","mappings":[{"order_id":"order-1","outcome_id":"outcome-1","action_id":"action-1"}]}`))
	if err == nil {
		t.Fatal("accepted an unbound source_ref for the ACCESSTRADE importer")
	}
}

func TestAppendJSONLLinesAtomicallyRejectsExternalSymlinkAndPostReadSwap(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions are not portable on Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "outcomes.jsonl")
	external := filepath.Join(t.TempDir(), "external.jsonl")
	if err := os.WriteFile(external, []byte("outside\n"), 0600); err != nil {
		t.Fatal(err)
	}
	externalBefore, err := os.ReadFile(external)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, path); err != nil {
		t.Fatal(err)
	}
	if err := appendJSONLLinesAtomically(path, [][]byte{[]byte(`{"outcome_id":"new"}`)}); err == nil {
		t.Fatal("symlinked JSONL target was accepted")
	}
	externalAfter, err := os.ReadFile(external)
	if err != nil || !bytes.Equal(externalBefore, externalAfter) {
		t.Fatalf("external symlink target changed: before=%q after=%q err=%v", externalBefore, externalAfter, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("canonical\n"), 0600); err != nil {
		t.Fatal(err)
	}
	stableRegularFileContentHook = func(openedPath string) error {
		if openedPath != path {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		return os.Symlink(external, path)
	}
	t.Cleanup(func() { stableRegularFileContentHook = nil })
	if err := appendJSONLLinesAtomically(path, [][]byte{[]byte(`{"outcome_id":"new"}`)}); err == nil {
		t.Fatal("post-read JSONL symlink replacement was accepted")
	}
	externalAfter, err = os.ReadFile(external)
	if err != nil || !bytes.Equal(externalBefore, externalAfter) {
		t.Fatalf("external target changed after post-read swap: before=%q after=%q err=%v", externalBefore, externalAfter, err)
	}
}

func TestParseAccesstradeCSVQuotedFields(t *testing.T) {
	rows, err := parseAccesstradeCSV("Mã đơn,Ghi chú\r\norder-1,\"dấu phẩy, và dấu \"\"nháy\"\"\"\r\n")
	if err != nil || len(rows) != 2 || rows[1][1] != "dấu phẩy, và dấu \"nháy\"" {
		t.Fatalf("rows=%q err=%v", rows, err)
	}
}

func TestAccesstradeReportRejectsUnverifiedShape(t *testing.T) {
	manifest := AccesstradeReportManifest{SnapshotID: "snapshot", ObservedAt: "2026-09-08T00:00:00Z", SourceRef: "fixture", Currency: "VND", Mappings: []AccesstradeReportMapping{{OrderID: "order-1", OutcomeID: "outcome-1", ActionID: "action-1"}}}
	for _, report := range []string{
		"Mã đơn,Trạng thái,Giá trị đơn hàng\norder-1,Tạm duyệt,1\n",
		"Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\norder-1,Hoàn tiền,1,0\n",
		"Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\norder-1,Tạm duyệt,1,-2\n",
	} {
		if _, err := decodeAccesstradeOutcomes([]byte(report), manifest); err == nil {
			t.Fatal("accepted an unverified CSV shape or status")
		}
	}
}
