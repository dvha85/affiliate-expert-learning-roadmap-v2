package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
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

func TestLoadM07RegistryRejectsNonCanonicalPolicyJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.json")
	valid := `[{"name":"public_http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":1000,"follow_redirects":false}]`
	cases := []struct {
		name string
		raw  string
	}{
		{"duplicate read-only policy", `[{"name":"public_http","read_only":false,"read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":1000,"follow_redirects":false}]`},
		{"case-variant policy field", `[{"name":"public_http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":1000,"follow_redirects":false,"Allowed_Hosts":["attacker.example"]}]`},
		{"unknown policy field", `[{"name":"public_http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":1000,"follow_redirects":false,"unreviewed_redirect_host":"attacker.example"}]`},
		{"trailing policy JSON", valid + ` []`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(test.raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadM07Registry(path); err == nil {
				t.Fatalf("non-canonical registry policy was accepted: %s", test.raw)
			}
		})
	}
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM07Registry(path); err != nil {
		t.Fatalf("valid registry policy was rejected: %v", err)
	}
}

func TestM07ContextFailsClosedWhileHistoryWriterIsActive(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	record, err := NewHistoryRecord("m07-history-gate-r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("m07-history-gate-o1", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(historyPath, record); err != nil {
		t.Fatal(err)
	}
	release, err := acquireHistoryRuntimeGate(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	var out, errOut bytes.Buffer
	if code := runM07([]string{"context", historyPath, record.RecordID}, &out, &errOut); code == 0 {
		t.Fatalf("M07 context succeeded during history writer activity: %s", out.String())
	}
	var envelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope["status"] != "BUSY" {
		t.Fatalf("M07 context did not fail closed: response=%s err=%v stderr=%s", out.String(), err, errOut.String())
	}
}

func TestM07AdapterContextFailsClosedWhileHistoryWriterIsActive(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	record, err := NewHistoryRecord("m07-adapter-history-gate-r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("m07-adapter-history-gate-o1", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(historyPath, record); err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(m07AdapterRequest{RecordID: record.RecordID, Registry: []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}}})
	if err != nil {
		t.Fatal(err)
	}
	release, err := acquireHistoryRuntimeGate(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	request := httptest.NewRequest(http.MethodPost, "/v1/m07/context", bytes.NewReader(body))
	response := httptest.NewRecorder()
	m07AdapterHandler(historyPath).ServeHTTP(response, request)
	if response.Code != http.StatusConflict {
		t.Fatalf("adapter status=%d body=%s, want %d", response.Code, response.Body.String(), http.StatusConflict)
	}
	var envelope map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil || envelope["status"] != "BUSY" {
		t.Fatalf("adapter did not fail closed: response=%s err=%v", response.Body.String(), err)
	}
}

func TestM07RejectsOutputInputAliasesBeforeParsing(t *testing.T) {
	dir := t.TempDir()
	historyPath := filepath.Join(dir, "history.jsonl")
	record, err := NewHistoryRecord("m07-path-r1", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("m07-path-o1", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(historyPath, record); err != nil {
		t.Fatal(err)
	}
	registryPath := filepath.Join(dir, "registry.json")
	resultPath := filepath.Join(dir, "tool-result.json")
	writeM07File(t, registryPath, []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000, FollowRedirects: false}})
	writeM07File(t, resultPath, corem07.ToolResult{RecordID: record.RecordID, ToolCall: corem07.ToolRequest{ToolName: "public_http", Method: "GET", Target: "https://example.com/a"}, StatusCode: 200, ReceivedAt: "2026-09-01T00:02:00Z", Body: json.RawMessage(`{"price":100}`)})
	inputs := map[string][]byte{}
	for _, path := range []string{historyPath, registryPath, resultPath} {
		inputs[path], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	outputs := []string{historyPath}
	hardlink := filepath.Join(dir, "result-hardlink.json")
	if err := os.Link(resultPath, hardlink); err != nil {
		t.Fatal(err)
	}
	outputs = append(outputs, hardlink)
	if runtime.GOOS != "windows" {
		symlink := filepath.Join(dir, "history-symlink.json")
		if err := os.Symlink(historyPath, symlink); err != nil {
			t.Fatal(err)
		}
		outputs = append(outputs, symlink)
	}
	for _, output := range outputs {
		t.Run(filepath.Base(output), func(t *testing.T) {
			var out, errOut bytes.Buffer
			if code := runM07([]string{"register-tool-result", historyPath, record.RecordID, registryPath, resultPath, output}, &out, &errOut); code == 0 {
				t.Fatalf("input alias was accepted: output=%s response=%s", output, out.String())
			}
			var envelope map[string]any
			if err := json.Unmarshal(out.Bytes(), &envelope); err != nil || envelope["status"] != "PATH_CONFLICT" {
				t.Fatalf("alias did not fail at path preflight: output=%s envelope=%s err=%v stderr=%s", output, out.String(), err, errOut.String())
			}
			for path, before := range inputs {
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatalf("input changed after rejected output alias %s: path=%s err=%v", output, path, err)
				}
			}
		})
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
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Claims: []corem07.Claim{claim}, EvidenceIDs: []string{evidence.EvidenceID}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{"id":9007199254740993}`)}}
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
	// M08 must resolve the persisted proposal and refuse a caller-supplied
	// target/parameter change rather than treating proposal_ref as a label.
	historyCtx, err := m07EvidenceContext(record)
	if err != nil {
		t.Fatal(err)
	}
	historyEvidenceID := record.RecordedResult.EvidenceIDs[0]
	historyEvidence := corem07.Evidence{}
	for _, item := range historyCtx.Evidence {
		if item.EvidenceID == historyEvidenceID {
			historyEvidence = item
		}
	}
	agentClaim := corem07.Claim{FieldOrClaim: historyEvidence.FieldOrClaim, Value: json.RawMessage(`{"observation_id":"o1","subject_id":"p","source_ref":"fixture:history-test","observed_at":"2026-09-01T00:00:00Z","access_method":"local_fixture","evidence_kind":"synthetic","use_context":"test","claim_kind":"assumption","state":"observed","limitation":"Synthetic history test fixture; not market truth.","product_id":"p","product_name":"P","price":100,"commission_rate":0.1,"currency":"USD"}`), EvidenceIDs: []string{historyEvidenceID}}
	agentClaim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{agentClaim})
	agentModel := corem07.AgentOutput{State: "HUMAN_REVIEW", Claims: []corem07.Claim{agentClaim}, EvidenceIDs: []string{historyEvidenceID}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{"id":9007199254740993}`)}}
	agentModel.Answer = corem07.RenderGroundedAnswer(agentModel.Claims)
	writeM07File(t, modelPath, agentModel)
	agentProposalPath := filepath.Join(dir, "agent-proposal.json")
	out.Reset()
	errOut.Reset()
	if code := runM07([]string{"register-proposal", historyPath, record.RecordID, modelPath, registryPath, agentProposalPath}, &out, &errOut); code != 0 {
		t.Fatalf("history-grounded proposal registration failed (%d): %s", code, errOut.String())
	}
	agentProposalRaw, err := os.ReadFile(agentProposalPath)
	if err != nil {
		t.Fatal(err)
	}
	agentProposal, _, err := corem07.ValidateRegisteredAgentProposal(agentProposalRaw, historyCtx.Evidence, registry, record.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	requestPath, intentPath, policyPath, policyOut := filepath.Join(dir, "agent-request.json"), filepath.Join(dir, "agent-intent.json"), filepath.Join(dir, "agent-policy.json"), filepath.Join(dir, "agent-policy-out.json")
	writeM07File(t, requestPath, map[string]any{"intent_id": "agent-i", "decision_id": record.RecordID, "evidence_ids": []string{historyEvidenceID}, "action_type": "DRAFT", "target": "https://example.com/draft", "parameters": map[string]any{"id": json.Number("9007199254740993")}, "proposed_by": "agent", "proposal_ref": agentProposal.ProposalID, "created_at": "2026-09-01T00:02:00Z", "expires_at": "2026-09-01T00:05:00Z", "correlation_id": "agent-c", "idempotency_key": "agent-k"})
	if code := runMissionCommand([]string{"m08-intent", historyPath, requestPath, agentProposalPath, intentPath}, &out, &errOut); code != 0 {
		t.Fatalf("agent intent did not resolve proposal (%d): %s", code, errOut.String())
	}
	writeM07File(t, policyPath, map[string]any{"policy_version": "v1", "now": "2026-09-01T00:03:00Z", "allowed_hosts": []string{"example.com"}, "action_risk": map[string]string{"DRAFT": "RISK0"}, "seen_idempotency": map[string]string{}})
	out.Reset()
	errOut.Reset()
	if code := runMissionCommand([]string{"m08-policy", historyPath, intentPath, policyPath, agentProposalPath, policyOut}, &out, &errOut); code != 0 {
		t.Fatalf("agent policy did not resolve proposal (%d): %s", code, errOut.String())
	}
	writeM07File(t, requestPath, map[string]any{"intent_id": "tampered-i", "decision_id": record.RecordID, "evidence_ids": []string{historyEvidenceID}, "action_type": "DRAFT", "target": "https://example.com/changed", "parameters": map[string]any{"id": json.Number("9007199254740993")}, "proposed_by": "agent", "proposal_ref": agentProposal.ProposalID, "created_at": "2026-09-01T00:02:00Z", "expires_at": "2026-09-01T00:05:00Z", "correlation_id": "tampered-c", "idempotency_key": "tampered-k"})
	out.Reset()
	errOut.Reset()
	if code := runMissionCommand([]string{"m08-intent", historyPath, requestPath, agentProposalPath, filepath.Join(dir, "tampered-intent.json")}, &out, &errOut); code == 0 {
		t.Fatal("agent target changed outside persisted proposal")
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
