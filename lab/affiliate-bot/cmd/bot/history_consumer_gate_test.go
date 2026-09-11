package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryConsumersFailClosedWhileWriterIsActive(t *testing.T) {
	dir := t.TempDir()
	history := filepath.Join(dir, "history.jsonl")
	actions := filepath.Join(dir, "actions.jsonl")
	outcomes := filepath.Join(dir, "outcomes.jsonl")
	config := filepath.Join(dir, "advisor.json")
	if err := os.WriteFile(config, []byte(`{"decision_id":"d","question":"q","as_of":"2026-09-07T00:00:00Z","max_age_hours":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	release, err := acquireHistoryRuntimeGate(history)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	for _, command := range []string{"list", "replay"} {
		if err := runHistory([]string{"history", command, history}); err == nil || !strings.Contains(err.Error(), "busy") {
			t.Fatalf("history %s did not fail closed: %v", command, err)
		}
	}
	if err := exportHistoryDecision(&bytes.Buffer{}, []string{history, "d", config}); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("history decision did not fail closed: %v", err)
	}
	var advisorOut, advisorErr bytes.Buffer
	if code := runAdvisor([]string{"mock", history, actions, outcomes, config}, &advisorOut, &advisorErr); code == 0 {
		t.Fatalf("advisor succeeded during history write: %s", advisorOut.String())
	}
	var advisorEnvelope map[string]any
	if err := json.Unmarshal(advisorOut.Bytes(), &advisorEnvelope); err != nil || advisorEnvelope["status"] != "BUSY" {
		t.Fatalf("advisor did not fail closed: response=%s err=%v stderr=%s", advisorOut.String(), err, advisorErr.String())
	}
	var receiptOut, receiptErr bytes.Buffer
	if code := runAccesstradeReceiptList([]string{"accesstrade-receipts", history, actions, outcomes}, &receiptOut, &receiptErr); code == 0 {
		t.Fatalf("receipt list succeeded during history write: %s", receiptOut.String())
	}
	var receiptEnvelope map[string]any
	if err := json.Unmarshal(receiptOut.Bytes(), &receiptEnvelope); err != nil || receiptEnvelope["status"] != "BUSY" {
		t.Fatalf("receipt list did not fail closed: response=%s err=%v stderr=%s", receiptOut.String(), err, receiptErr.String())
	}
}
