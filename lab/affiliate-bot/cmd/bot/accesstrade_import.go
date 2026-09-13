package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
)

// AccesstradeReportManifest deliberately maps an already-sanitized order key to
// a recorded human action. UTM is useful for analysis, but is not affiliate
// attribution and is therefore never used to create this link.
type AccesstradeReportManifest struct {
	SnapshotID string                     `json:"snapshot_id"`
	ObservedAt string                     `json:"observed_at"`
	SourceRef  string                     `json:"source_ref"`
	Currency   string                     `json:"currency"`
	Mappings   []AccesstradeReportMapping `json:"mappings"`
}

type AccesstradeReportMapping struct {
	OrderID   string `json:"order_id"`
	OutcomeID string `json:"outcome_id"`
	ActionID  string `json:"action_id"`
}

const maxAccesstradeReportBytes = 16 << 20

func decodeAccesstradeManifest(raw []byte) (AccesstradeReportManifest, error) {
	var manifest AccesstradeReportManifest
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return manifest, fmt.Errorf("manifest contains trailing data")
	}
	if strings.TrimSpace(manifest.SnapshotID) == "" || !strings.HasPrefix(manifest.SourceRef, "accesstrade:") || manifest.Currency != "VND" || len(manifest.Mappings) == 0 {
		return manifest, fmt.Errorf("manifest requires snapshot_id, source_ref beginning accesstrade:, currency=VND and mappings")
	}
	if _, err := parseRFC3339(manifest.ObservedAt); err != nil {
		return manifest, fmt.Errorf("manifest observed_at: %w", err)
	}
	seenOrders, seenOutcomes := map[string]bool{}, map[string]bool{}
	for _, mapping := range manifest.Mappings {
		if strings.TrimSpace(mapping.OrderID) == "" || strings.TrimSpace(mapping.OutcomeID) == "" || strings.TrimSpace(mapping.ActionID) == "" {
			return manifest, fmt.Errorf("every mapping requires order_id, outcome_id and action_id")
		}
		if seenOrders[mapping.OrderID] || seenOutcomes[mapping.OutcomeID] {
			return manifest, fmt.Errorf("manifest reuses an order_id or outcome_id")
		}
		seenOrders[mapping.OrderID], seenOutcomes[mapping.OutcomeID] = true, true
	}
	return manifest, nil
}

func parseRFC3339(value string) (string, error) {
	value = strings.TrimSpace(value)
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return "", err
	}
	return value, nil
}

func accesstradeColumn(headers []string, names ...string) int {
	for index, header := range headers {
		for _, name := range names {
			if strings.EqualFold(strings.TrimSpace(strings.TrimPrefix(header, "\ufeff")), name) {
				return index
			}
		}
	}
	return -1
}

// parseAccesstradeCSV is intentionally small but handles RFC-4180's important
// quoted-comma and escaped-quote cases. The report is size-limited before this
// function runs, so retaining the parsed rows is bounded.
func parseAccesstradeCSV(input string) ([][]string, error) {
	rows := [][]string{}
	row := []string{}
	var field strings.Builder
	inQuotes, afterQuote := false, false
	flushField := func() { row = append(row, field.String()); field.Reset(); afterQuote = false }
	flushRow := func() {
		flushField()
		if len(row) != 1 || row[0] != "" {
			rows = append(rows, row)
		}
		row = nil
	}
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if inQuotes {
			if ch == '"' {
				if i+1 < len(input) && input[i+1] == '"' {
					field.WriteByte('"')
					i++
					continue
				}
				inQuotes, afterQuote = false, true
				continue
			}
			field.WriteByte(ch)
			continue
		}
		switch ch {
		case '"':
			if field.Len() != 0 || afterQuote {
				return nil, fmt.Errorf("unexpected quote")
			}
			inQuotes = true
		case ',':
			flushField()
		case '\n':
			flushRow()
		case '\r':
			if i+1 < len(input) && input[i+1] == '\n' {
				continue
			}
			flushRow()
		default:
			if afterQuote {
				return nil, fmt.Errorf("unexpected character after closing quote")
			}
			field.WriteByte(ch)
		}
	}
	if inQuotes {
		return nil, fmt.Errorf("unterminated quoted field")
	}
	if field.Len() != 0 || len(row) != 0 || afterQuote {
		flushRow()
	}
	return rows, nil
}

func parseAccesstradeAmount(raw string) (float64, bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, false, nil // blank is unknown, not zero
	}
	value = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '₫' {
			return -1
		}
		return r
	}, value)
	value = strings.TrimSuffix(strings.TrimSuffix(value, "VND"), "đ")
	if strings.HasPrefix(value, "-") || value == "" || len(value) > 64 {
		return 0, false, fmt.Errorf("invalid monetary value")
	}
	if strings.Contains(value, ".") && strings.Contains(value, ",") {
		value = strings.ReplaceAll(strings.ReplaceAll(value, ".", ""), ",", ".")
	} else if strings.Count(value, ".") > 1 || (strings.Count(value, ".") == 1 && len(value)-strings.LastIndex(value, ".")-1 == 3) {
		value = strings.ReplaceAll(value, ".", "")
	} else if strings.Contains(value, ",") {
		value = strings.ReplaceAll(value, ",", ".")
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 {
		return 0, false, fmt.Errorf("invalid monetary value")
	}
	return number, true, nil
}

func accesstradeOutcomeStatus(status string) (string, error) {
	switch strings.TrimSpace(status) {
	case "Chờ xử lý", "Tạm duyệt":
		return "PENDING", nil
	case "Được duyệt":
		return "VALID", nil
	case "Từ chối":
		return "CANCELLED", nil
	default:
		return "", fmt.Errorf("unmapped ACCESSTRADE status %q", status)
	}
}

func decodeAccesstradeOutcomes(report []byte, manifest AccesstradeReportManifest) ([]m03.OutcomeRecord, error) {
	rows, err := parseAccesstradeCSV(string(report))
	if err != nil || len(rows) == 0 {
		return nil, fmt.Errorf("read CSV: %w", err)
	}
	headers := rows[0]
	orderColumn := accesstradeColumn(headers, "Mã đơn")
	statusColumn := accesstradeColumn(headers, "Trạng thái")
	valueColumn := accesstradeColumn(headers, "Giá trị đơn hàng")
	commissionColumn := accesstradeColumn(headers, "Hoa hồng", "Hoa hồng")
	if orderColumn < 0 || statusColumn < 0 || valueColumn < 0 || commissionColumn < 0 {
		return nil, fmt.Errorf("CSV must include Mã đơn, Trạng thái, Giá trị đơn hàng and Hoa hồng")
	}
	mappings := map[string]AccesstradeReportMapping{}
	for _, mapping := range manifest.Mappings {
		mappings[mapping.OrderID] = mapping
	}
	seenOrders := map[string]bool{}
	outcomes := make([]m03.OutcomeRecord, 0, len(manifest.Mappings))
	for line, row := range rows[1:] {
		line += 2
		if len(row) != len(headers) {
			return nil, fmt.Errorf("CSV row %d has a different column count", line)
		}
		orderID := strings.TrimSpace(row[orderColumn])
		mapping, ok := mappings[orderID]
		if !ok || seenOrders[orderID] {
			return nil, fmt.Errorf("CSV row %d has an unmapped or duplicate sanitized order_id", line)
		}
		seenOrders[orderID] = true
		status, err := accesstradeOutcomeStatus(row[statusColumn])
		if err != nil {
			return nil, fmt.Errorf("CSV row %d: %w", line, err)
		}
		metrics := map[string]float64{}
		if amount, present, err := parseAccesstradeAmount(row[valueColumn]); err != nil {
			return nil, fmt.Errorf("CSV row %d order value: %w", line, err)
		} else if present {
			metrics["order_value_vnd"] = amount
		}
		if amount, present, err := parseAccesstradeAmount(row[commissionColumn]); err != nil {
			return nil, fmt.Errorf("CSV row %d commission: %w", line, err)
		} else if present {
			// A report's commission amount is not necessarily paid. The metric
			// name preserves that distinction for PENDING and VALID rows.
			metrics["reported_commission_vnd"] = amount
		}
		outcomes = append(outcomes, m03.OutcomeRecord{
			OutcomeID:  mapping.OutcomeID,
			EffectRef:  m03.EffectRef{EffectKind: "HUMAN_ACTION", EffectID: mapping.ActionID},
			ObservedAt: manifest.ObservedAt,
			Status:     status,
			Metrics:    metrics,
			SourceRef:  manifest.SourceRef,
		})
	}
	if len(outcomes) != len(manifest.Mappings) {
		return nil, fmt.Errorf("CSV does not contain every mapped sanitized order_id")
	}
	return outcomes, nil
}

// appendJSONLLinesAtomically commits a complete validated snapshot or preserves
// prior JSONL bytes. It intentionally does not create a missing parent
// directory, matching the existing append-store boundary.
func appendJSONLLinesAtomically(path string, records [][]byte) error {
	directory := filepath.Dir(path)
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("outcome parent must be a non-symlink directory")
	}
	previous, opened, err := readStableRegularFile(path)
	if os.IsNotExist(err) {
		previous, opened = nil, nil
	} else if err != nil {
		return err
	}
	if len(previous) > 0 && previous[len(previous)-1] != '\n' {
		return fmt.Errorf("outcome store is missing newline framing")
	}
	temporary, err := os.CreateTemp(directory, ".outcome-snapshot-")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(previous); err != nil {
		_ = temporary.Close()
		return err
	}
	for _, record := range records {
		if _, err := temporary.Write(append(append([]byte{}, record...), '\n')); err != nil {
			_ = temporary.Close()
			return err
		}
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if opened == nil {
		// Do not let a name that appeared after the absent check be overwritten.
		// Link publishes only a new path, so a symlink planted at that name cannot
		// be replaced or used as an implicit source of canonical JSONL bytes.
		if err := os.Link(temporaryPath, path); err != nil {
			return err
		}
	} else {
		if err := verifyStableRegularFileName(path, opened); err != nil {
			return err
		}
		if err := os.Rename(temporaryPath, path); err != nil {
			return err
		}
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	err = directoryHandle.Sync()
	closeErr := directoryHandle.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func appendOutcomesAtomically(path string, records [][]byte) error {
	return appendJSONLLinesAtomically(path, records)
}

func runAccesstradeOutcomeImport(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		envelope := map[string]any{"command": "outcome accesstrade-import", "status": status, "execution_permitted": false}
		if artifact != nil {
			envelope["artifact"] = artifact
		}
		if err := json.NewEncoder(stdout).Encode(envelope); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return code
	}
	if len(args) != 6 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot outcome accesstrade-import HISTORY ACTIONS OUTCOMES REPORT.csv MANIFEST.json"), 2)
	}
	if err := distinctActionPaths(args[1:]...); err != nil {
		return emit("PATH_ERROR", nil, err, 1)
	}
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
	existing, err := loadOutcomes(args[3], actions)
	if err != nil && !os.IsNotExist(err) {
		return emit("STORE_ERROR", nil, err, 1)
	}
	receipts, err := validateAccesstradeReceiptStore(args[3], existing)
	if err != nil {
		return emit("RECEIPT_STORE_ERROR", nil, err, 1)
	}
	reportInfo, err := os.Stat(args[4])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	if reportInfo.Size() > maxAccesstradeReportBytes {
		return emit("REPORT_TOO_LARGE", nil, fmt.Errorf("report exceeds %d bytes", maxAccesstradeReportBytes), 1)
	}
	report, err := os.ReadFile(args[4])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	rawManifest, err := os.ReadFile(args[5])
	if err != nil {
		return emit("IO_ERROR", nil, err, 1)
	}
	manifest, err := decodeAccesstradeManifest(rawManifest)
	if err != nil {
		return emit("MANIFEST_ERROR", nil, err, 1)
	}
	candidates, err := decodeAccesstradeOutcomes(report, manifest)
	if err != nil {
		return emit("REPORT_ERROR", nil, err, 1)
	}
	for _, candidate := range candidates {
		raw, err := json.Marshal(candidate)
		if err != nil {
			return emit("INVALID_SCHEMA", nil, err, 1)
		}
		if _, status := linkedOutcome(raw, actions); status != "VALID" {
			return emit(status, nil, fmt.Errorf("outcome rejected: %s", status), 1)
		}
	}
	receipt := newAccesstradeReceipt(manifest, report, rawManifest, args[3], candidates)
	if err := validateAccesstradeReceipt(receipt); err != nil {
		return emit("INVALID_RECEIPT", nil, err, 1)
	}
	byID := map[string]m03.OutcomeRecord{}
	for _, outcome := range existing {
		byID[outcome.OutcomeID] = outcome
	}
	duplicates := 0
	for _, candidate := range candidates {
		if old, exists := byID[candidate.OutcomeID]; exists {
			if !sameOutcome(old, candidate) {
				return emit("CONFLICT", nil, fmt.Errorf("outcome_id reused with different content"), 1)
			}
			duplicates++
		}
	}
	if duplicates == len(candidates) {
		for _, existingReceipt := range receipts {
			if existingReceipt.ReceiptID == receipt.ReceiptID {
				if sameReceipt(existingReceipt, receipt) {
					return emit("EXACT_DUPLICATE", map[string]any{"outcomes": candidates, "receipt": receipt}, nil, 0)
				}
				return emit("RECEIPT_CONFLICT", nil, fmt.Errorf("snapshot_id reused with different source bytes or metadata"), 1)
			}
		}
		return emit("RECEIPT_MISSING", nil, fmt.Errorf("exact outcomes do not have their exact ACCESSTRADE receipt"), 1)
	}
	if duplicates > 0 {
		return emit("PARTIAL_DUPLICATE", nil, fmt.Errorf("all rows in a snapshot must be new or exact duplicates"), 1)
	}
	for _, existingReceipt := range receipts {
		if existingReceipt.ReceiptID == receipt.ReceiptID {
			return emit("RECEIPT_CONFLICT", nil, fmt.Errorf("snapshot_id reused with different content"), 1)
		}
	}
	encodedRecords := make([][]byte, 0, len(candidates))
	for _, candidate := range candidates {
		encodedCandidate, err := json.Marshal(candidate)
		if err != nil {
			return emit("INVALID_SCHEMA", nil, err, 1)
		}
		encodedRecords = append(encodedRecords, encodedCandidate)
	}
	if err := writeJSONAtomic(accesstradeJournalPath(args[3]), receipt); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if err := appendOutcomesAtomically(args[3], encodedRecords); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	encodedReceipt, err := json.Marshal(receipt)
	if err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if err := appendJSONLLinesAtomically(accesstradeReceiptPath(args[3]), [][]byte{encodedReceipt}); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	if err := os.Remove(accesstradeJournalPath(args[3])); err != nil {
		return emit("STORE_ERROR", nil, err, 1)
	}
	return emit("APPENDED", map[string]any{"outcomes": candidates, "receipt": receipt}, nil, 0)
}
