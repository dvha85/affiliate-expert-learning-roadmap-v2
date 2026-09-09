package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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

func TestAccesstradeImporterLifecycle(t *testing.T) {
	history, actions, outcomes, dir := setupAccesstradeImport(t)
	report := filepath.Join(dir, "report.csv")
	manifest := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(report, []byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng\nredacted-1,Tạm duyệt,250.000,17.500\nredacted-2,Từ chối,150.000,0\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, []byte(`{"snapshot_id":"snapshot-1","observed_at":"2026-09-08T12:00:00+07:00","source_ref":"accesstrade:sanitized:snapshot-1","currency":"VND","mappings":[{"order_id":"redacted-1","outcome_id":"outcome-1","action_id":"manual-action-001"},{"order_id":"redacted-2","outcome_id":"outcome-2","action_id":"manual-action-001"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
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

func TestAccesstradeImporterDoesNotMapUTM(t *testing.T) {
	manifest := AccesstradeReportManifest{SnapshotID: "snapshot", ObservedAt: "2026-09-08T00:00:00Z", SourceRef: "fixture", Currency: "VND", Mappings: []AccesstradeReportMapping{{OrderID: "order-1", OutcomeID: "outcome-1", ActionID: "action-1"}}}
	if _, err := decodeAccesstradeOutcomes([]byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng,UTM Source\norder-1,Tạm duyệt,1,0,action-1\n"), manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := decodeAccesstradeOutcomes([]byte("Mã đơn,Trạng thái,Giá trị đơn hàng,Hoa hồng,UTM Source\norder-2,Tạm duyệt,1,0,action-1\n"), manifest); err == nil {
		t.Fatal("UTM source was accepted as an implicit action mapping")
	}
}

func TestAccesstradeManifestRequiresBoundedSourceRef(t *testing.T) {
	_, err := decodeAccesstradeManifest([]byte(`{"snapshot_id":"snapshot","observed_at":"2026-09-08T00:00:00Z","source_ref":"fixture:unbound","currency":"VND","mappings":[{"order_id":"order-1","outcome_id":"outcome-1","action_id":"action-1"}]}`))
	if err == nil {
		t.Fatal("accepted an unbound source_ref for the ACCESSTRADE importer")
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
