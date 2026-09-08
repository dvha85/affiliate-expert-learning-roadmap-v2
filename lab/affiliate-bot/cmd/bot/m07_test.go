package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
)

func writeM07File(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestM07RegistersToolResultBeforeItCanBeCited(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	record, err := NewHistoryRecord("r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("o1", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(historyPath, record); err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(dir, "registry.json")
	resultPath := filepath.Join(dir, "tool-result.json")
	registeredPath := filepath.Join(dir, "registered.json")
	writeM07File(t, registryPath, []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}})
	writeM07File(t, resultPath, corem07.ToolResult{RecordID: record.RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-01T00:02:00Z", Body: json.RawMessage(`{"price":100}`)})

	var out, errOut bytes.Buffer
	if code := runM07([]string{"register-tool-result", historyPath, record.RecordID, registryPath, resultPath, registeredPath}, &out, &errOut); code != 0 {
		t.Fatalf("register failed (%d): %s", code, errOut.String())
	}
	registeredRaw, err := os.ReadFile(registeredPath)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := loadM07Registry(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	registered, err := corem07.ValidateRegisteredToolResult(registeredRaw, registry, record.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	evidence := registered.Evidence()
	claim := corem07.Claim{FieldOrClaim: evidence.FieldOrClaim, Value: json.RawMessage(`{"price":100}`), EvidenceIDs: []string{evidence.EvidenceID}}
	claim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{claim})
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Claims: []corem07.Claim{claim}, EvidenceIDs: []string{evidence.EvidenceID}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false}
	model.Answer = corem07.RenderGroundedAnswer(model.Claims)
	modelPath := filepath.Join(dir, "model.json")
	writeM07File(t, modelPath, model)
	out.Reset()
	errOut.Reset()
	if code := runM07([]string{"validate", historyPath, record.RecordID, modelPath, registryPath, registeredPath}, &out, &errOut); code != 0 {
		t.Fatalf("registered evidence was not citable (%d): %s", code, errOut.String())
	}
	proposalPath := filepath.Join(dir, "proposal.json")
	out.Reset()
	errOut.Reset()
	if code := runM07([]string{"register-proposal", historyPath, record.RecordID, modelPath, registryPath, proposalPath, registeredPath}, &out, &errOut); code != 0 {
		t.Fatalf("validated agent output was not persisted (%d): %s", code, errOut.String())
	}
	proposalRaw, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := corem07.ValidateRegisteredAgentProposal(proposalRaw, []corem07.Evidence{registered.Evidence()}, registry, record.RecordID); err != nil {
		t.Fatal(err)
	}

	model.Answer = "guaranteed profit"
	writeM07File(t, modelPath, model)
	out.Reset()
	errOut.Reset()
	if code := runM07([]string{"validate", historyPath, record.RecordID, modelPath, registryPath, registeredPath}, &out, &errOut); code == 0 {
		t.Fatal("forged prose was accepted")
	}

	var tamperedRegistered corem07.RegisteredToolResult
	if err := json.Unmarshal(registeredRaw, &tamperedRegistered); err != nil {
		t.Fatal(err)
	}
	tamperedRegistered.TraceID = "sha256:forged"
	tamperedTrace, err := json.Marshal(tamperedRegistered)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := corem07.ValidateRegisteredToolResult(tamperedTrace, registry, record.RecordID); err == nil {
		t.Fatal("core accepted a forged trace")
	}
	if err := os.WriteFile(registeredPath, tamperedTrace, 0o600); err != nil {
		t.Fatal(err)
	}
	model.Answer = corem07.RenderGroundedAnswer(model.Claims)
	writeM07File(t, modelPath, model)
	out.Reset()
	errOut.Reset()
	if code := runM07([]string{"validate", historyPath, record.RecordID, modelPath, registryPath, registeredPath}, &out, &errOut); code == 0 {
		t.Fatal("forged registered trace was accepted")
	}
}
