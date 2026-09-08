package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m05"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
)

func backupCall(t *testing.T, args ...string) (int, map[string]any) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runBackupCommand(args, &stdout, &stderr)
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode backup response %q: %v (stderr: %s)", stdout.String(), err, stderr.String())
	}
	if code != 0 {
		t.Logf("backup stderr: %s", stderr.String())
	}
	return code, envelope
}

func m07AdapterCall(t *testing.T, history, endpoint string, request m07AdapterRequest) map[string]any {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	m07AdapterHandler(history).ServeHTTP(recorder, httptest.NewRequest("POST", endpoint, bytes.NewReader(body)))
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode M07 adapter response %q: %v", recorder.Body.String(), err)
	}
	if recorder.Code != 200 {
		t.Fatalf("M07 adapter rejected %s: code=%d response=%+v", endpoint, recorder.Code, response)
	}
	return response
}

func TestRuntimeGateRejectsAnotherProcess(t *testing.T) {
	if os.Getenv("GO_WANT_RUNTIME_GATE_HELPER") == "1" {
		_, err := acquireRuntimeGate(os.Args[len(os.Args)-1])
		if err != nil {
			os.Exit(0)
		}
		os.Exit(1)
	}
	dir := t.TempDir()
	release, err := acquireRuntimeGate(dir)
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=TestRuntimeGateRejectsAnotherProcess", "--", dir)
	child.Env = append(os.Environ(), "GO_WANT_RUNTIME_GATE_HELPER=1")
	if err := child.Run(); err != nil {
		t.Fatalf("second process acquired the runtime gate: %v", err)
	}
	release()
	child = exec.Command(os.Args[0], "-test.run=TestRuntimeGateRejectsAnotherProcess", "--", dir)
	child.Env = append(os.Environ(), "GO_WANT_RUNTIME_GATE_HELPER=1")
	if err := child.Run(); err == nil {
		t.Fatal("child kept the runtime gate after it acquired it")
	}
}

func copyFlatBackup(t *testing.T, source, target string) {
	t.Helper()
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		body, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, entry.Name()), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBackupRestoreReplaysM00ToM05Graph(t *testing.T) {
	proposalArgs, reviewArgs := improvementFixture(t)
	improvementRun(t, "proposal", proposalArgs, "APPENDED")
	improvementRun(t, "review", reviewArgs, "APPENDED")
	runtime := filepath.Dir(proposalArgs[5])
	if code, response := missionCall(t, "init", runtime); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("mission init failed: code=%d response=%+v", code, response)
	}
	backup, restored := filepath.Join(filepath.Dir(runtime), "m00-m05-backup"), filepath.Join(filepath.Dir(runtime), "m00-m05-restored")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("M00-M05 backup failed: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("M00-M05 restore failed: code=%d response=%+v", code, response)
	}
	if _, err := loadImprovementRecords(filepath.Join(restored, "reviews.jsonl"), func(raw []byte) (m05.ReviewRecord, error) {
		history, loadErr := LoadHistory(filepath.Join(restored, "history.jsonl"))
		if loadErr != nil {
			return m05.ReviewRecord{}, loadErr
		}
		actions, loadErr := loadActions(filepath.Join(restored, "actions.jsonl"), history)
		if loadErr != nil {
			return m05.ReviewRecord{}, loadErr
		}
		outcomes, loadErr := loadOutcomes(filepath.Join(restored, "outcomes.jsonl"), actions)
		if loadErr != nil {
			return m05.ReviewRecord{}, loadErr
		}
		evaluations, loadErr := loadEvaluations(filepath.Join(restored, "evaluations.jsonl"), history, actions, outcomes)
		if loadErr != nil {
			return m05.ReviewRecord{}, loadErr
		}
		proposals, loadErr := loadImprovementRecords(filepath.Join(restored, "proposals.jsonl"), func(value []byte) (m05.ImprovementProposal, error) { return linkedProposal(value, evaluations) }, func(p m05.ImprovementProposal) string { return p.ProposalID })
		if loadErr != nil {
			return m05.ReviewRecord{}, loadErr
		}
		return linkedReview(raw, proposals, evaluations)
	}, func(r m05.ReviewRecord) string { return r.ReviewID }); err != nil {
		t.Fatalf("restored M05 review did not resolve from canonical graph: %v", err)
	}

	broken := filepath.Join(filepath.Dir(runtime), "m00-m05-broken")
	copyFlatBackup(t, backup, broken)
	outcomesPath := filepath.Join(broken, "outcomes.jsonl")
	brokenOutcomes, err := os.ReadFile(outcomesPath)
	if err != nil {
		t.Fatal(err)
	}
	brokenOutcomes = bytes.Replace(brokenOutcomes, []byte("br11-action"), []byte("missing-action"), 1)
	if err := os.WriteFile(outcomesPath, brokenOutcomes, 0600); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["outcomes.jsonl"], err = fileDigest(outcomesPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(filepath.Dir(runtime), "m00-m05-broken-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("checksum-valid orphan M03 outcome was restored: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreReplaysExpiredAuthorityButBlocksNewReservation(t *testing.T) {
	dir := t.TempDir()
	if _, err := buildBR10AdvisorFixture(dir); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("mission init failed: code=%d response=%+v", code, response)
	}
	intent := LearnerIntent{IntentID: "expired-intent", DecisionID: "br11-decision", EvidenceIDs: []string{"br11-observation"}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-01T00:00:00Z", ExpiresAt: "2026-09-08T00:00:00Z", CorrelationID: "expired-correlation", IdempotencyKey: "expired-key", IntentMode: "PROPOSAL_ONLY"}
	intent.IntentHash = learnerIntentHash(intent)
	approval := LearnerApproval{ApprovalID: "expired-approval", IntentID: intent.IntentID, IntentHash: intent.IntentHash, PolicyVersion: "expired-policy", Decision: "APPROVE", ApprovedBy: "human", ApproverID: "expired-reviewer", ApprovedAt: "2026-09-01T00:00:00Z", ExpiresAt: "2026-09-08T00:00:00Z", CorrelationID: intent.CorrelationID, OneTime: true}
	grant := corem10.CanaryGrant{GrantID: "expired-grant", GrantVersion: "v1", PolicyVersion: "expired-policy", ApprovalRef: approval.ApprovalID, ApprovedBy: "human", ApproverID: approval.ApproverID, ApprovedAt: approval.ApprovedAt, ValidFrom: approval.ApprovedAt, ExpiresAt: approval.ExpiresAt, AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: intent.CorrelationID, HashVersion: "go-json-v1"}
	grant.GrantHash = corem10.ComputeCanaryGrantHash(grant)
	state, err := loadMissionState(dir)
	if err != nil {
		t.Fatal(err)
	}
	state.Intent = &intent
	state.Policy = &LearnerPolicy{PolicyVersion: approval.PolicyVersion, IntentID: intent.IntentID, IntentHash: intent.IntentHash, Decision: "ALLOW", RiskClass: "RISK0", PolicyCheckedAt: approval.ApprovedAt}
	state.Approval = &approval
	state.Canary = &LearnerCanary{CanaryGrant: grant, Status: "ACTIVE"}
	if err := saveMissionState(dir, state); err != nil {
		t.Fatal(err)
	}
	grantRaw, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM10Artifact(dir, corem10.ArtifactKindCanaryGrant, grantRaw); err != nil {
		t.Fatal(err)
	}
	backup, restored := filepath.Join(filepath.Dir(dir), "expired-backup"), filepath.Join(filepath.Dir(dir), "expired-restored")
	if code, response := backupCall(t, "create", dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("expired historical authority was not backed up: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("expired historical authority was not restored: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve", restored, "1", "must-remain-blocked"); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("expired restored authority accepted a new reservation: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreCarriesAndValidatesM07Sidecar(t *testing.T) {
	root := t.TempDir()
	runtime := filepath.Join(root, "runtime")
	history := filepath.Join(runtime, "history.jsonl")
	if err := os.MkdirAll(runtime, 0700); err != nil {
		t.Fatal(err)
	}
	record, err := NewHistoryRecord("m07-backup", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("m07-o", "p", "P", 100, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(history, record); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "init", runtime); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("mission init failed: code=%d response=%+v", code, response)
	}

	registry := []corem07.ToolSpec{{Name: "public_http", ReadOnly: true, AllowedMethods: []string{"GET"}, AllowedHosts: []string{"example.com"}, TimeoutMS: 1000}}
	toolRaw := []byte(`{"record_id":"m07-backup","tool_call":{"tool_name":"public_http","method":"GET","target":"https://example.com/a"},"status_code":200,"received_at":"2026-09-01T00:02:00Z","redirected":false,"body":{"price":100}}`)
	tool, err := corem07.RegisterToolResult(toolRaw, registry)
	if err != nil {
		t.Fatal(err)
	}
	if response := m07AdapterCall(t, history, "/v1/m07/register-tool-result", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ToolResult: toolRaw}); response["status"] != "ACK" {
		t.Fatalf("adapter did not persist M07 tool result: %+v", response)
	}
	toolPath, err := m07ArtifactPath(history, "tool-results", tool.TraceID)
	if err != nil {
		t.Fatal(err)
	}
	evidence := tool.Evidence()
	claim := corem07.Claim{FieldOrClaim: evidence.FieldOrClaim, Value: json.RawMessage(`{"price":100}`), EvidenceIDs: []string{evidence.EvidenceID}}
	claim.Text = corem07.RenderGroundedAnswer([]corem07.Claim{claim})
	model := corem07.AgentOutput{State: "HUMAN_REVIEW", Answer: claim.Text, EvidenceIDs: []string{evidence.EvidenceID}, Claims: []corem07.Claim{claim}, ToolCalls: []corem07.ToolRequest{}, Authority: "A2-RO", WritePermission: false, ProposedAction: &corem07.ProposedAction{ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: json.RawMessage(`{"id":1}`)}}
	modelRaw, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := corem07.RegisterAgentProposal(modelRaw, []corem07.Evidence{evidence}, registry, record.RecordID)
	if err != nil {
		t.Fatal(err)
	}
	proposalPath, err := m07ArtifactPath(history, "proposals", proposal.ProposalID)
	if err != nil {
		t.Fatal(err)
	}
	if response := m07AdapterCall(t, history, "/v1/m07/register-proposal", m07AdapterRequest{RecordID: record.RecordID, Registry: registry, ModelOutput: modelRaw, ToolResultID: tool.TraceID}); response["status"] != "ACK" {
		t.Fatalf("adapter did not persist M07 proposal: %+v", response)
	}
	intent := LearnerIntent{IntentID: "m07-backup-intent", DecisionID: record.RecordID, EvidenceIDs: []string{evidence.EvidenceID}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{"id": 1}, ProposedBy: "agent", ProposalRef: proposal.ProposalID, CreatedAt: "2026-09-01T00:02:00Z", ExpiresAt: "2099-09-01T00:02:00Z", CorrelationID: "m07-backup-c", IdempotencyKey: "m07-backup-k", IntentMode: "PROPOSAL_ONLY"}
	intent.IntentHash = learnerIntentHash(intent)
	state, err := loadMissionState(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Intent = &intent
	state.Policy = &LearnerPolicy{PolicyVersion: "m07-backup-policy", IntentID: intent.IntentID, IntentHash: intent.IntentHash}
	if err := saveMissionState(runtime, state); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(runtime, "must-not-be-created")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup target inside runtime was accepted: code=%d response=%+v", code, response)
	}
	beforeBusyWrite, err := os.ReadFile(history)
	if err != nil {
		t.Fatal(err)
	}
	releaseGate, err := acquireRuntimeGate(runtime)
	if err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "busy-backup")); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("backup entered a runtime with an active writer: code=%d response=%+v", code, response)
	}
	blockedRecord, err := NewHistoryRecord("m07-blocked", "2026-09-01T01:00:00Z", "2026-09-01T00:01:00Z", []Observation{historyObservation("m07-blocked-o", "p2", "P2", 101, .1, "2026-09-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AppendHistory(history, blockedRecord); err == nil {
		t.Fatal("history writer entered a runtime while snapshot gate was held")
	}
	releaseGate()
	if afterBusyWrite, err := os.ReadFile(history); err != nil || !bytes.Equal(beforeBusyWrite, afterBusyWrite) {
		t.Fatalf("blocked history writer changed bytes: %v", err)
	}

	backup, restored := filepath.Join(root, "backup"), filepath.Join(root, "restored")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup did not accept M07 sidecar: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "restore", backup, restored); code != 0 || response["status"] != "RESTORED" {
		t.Fatalf("restore did not validate M07 sidecar: code=%d response=%+v", code, response)
	}
	restoredTool, err := os.ReadFile(filepath.Join(restored, "history.jsonl.m07", "tool-results", tool.TraceID[len("sha256:"):]+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := corem07.ValidateStoredRegisteredToolResult(restoredTool, record.RecordID); err != nil {
		t.Fatalf("restored tool trace did not replay: %v", err)
	}
	if _, err := os.Stat(filepath.Join(restored, "history.jsonl.m07", "proposals", proposal.ProposalID[len("sha256:"):]+".json")); err != nil {
		t.Fatalf("restored proposal is missing: %v", err)
	}
	interrupted := filepath.Join(root, "interrupted-backup")
	backupCopyFault = func(name string) error {
		if name == "mission-state.json" {
			return os.ErrClosed
		}
		return nil
	}
	defer func() { backupCopyFault = nil }()
	if code, response := backupCall(t, "create", runtime, interrupted); code == 0 || response["status"] != "STORE_ERROR" {
		t.Fatalf("interrupted backup was accepted: code=%d response=%+v", code, response)
	}
	backupCopyFault = nil
	if _, err := os.Stat(filepath.Join(interrupted, "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("interrupted backup published a manifest: %v", err)
	}
	if code, response := backupCall(t, "restore", interrupted, filepath.Join(root, "interrupted-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("interrupted snapshot was restorable: code=%d response=%+v", code, response)
	}

	badBackup := filepath.Join(root, "bad-backup")
	if err := os.MkdirAll(filepath.Join(badBackup, "history.jsonl.m07", "proposals"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"history.jsonl", "mission-state.json"} {
		body, err := os.ReadFile(filepath.Join(backup, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(badBackup, name), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, item := range []struct{ source, target string }{{toolPath, filepath.Join(badBackup, "history.jsonl.m07", "tool-results", tool.TraceID[len("sha256:"):]+".json")}, {proposalPath, filepath.Join(badBackup, "history.jsonl.m07", "proposals", proposal.ProposalID[len("sha256:"):]+".json")}} {
		body, err := os.ReadFile(item.source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(item.target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(item.target, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(badBackup, "history.jsonl.m07", "proposals", proposal.ProposalID[len("sha256:"):]+".json"), []byte(`{"version":"m07-agent-proposal/v1"}`), 0600); err != nil {
		t.Fatal(err)
	}
	files, err := backupFiles(badBackup)
	if err != nil {
		t.Fatal(err)
	}
	manifest := backupManifest{Version: "affiliate-bot-backup/v2", Files: map[string]string{}, Required: []string{"history.jsonl", "mission-state.json", "history.jsonl.m07/proposals/" + proposal.ProposalID[len("sha256:"):] + ".json", "history.jsonl.m07/tool-results/" + tool.TraceID[len("sha256:"):] + ".json"}}
	for _, name := range files {
		digest, err := fileDigest(filepath.Join(badBackup, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		manifest.Files[name] = digest
	}
	if err := writeJSON(filepath.Join(badBackup, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", badBackup, filepath.Join(root, "bad-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("checksum-valid forged M07 proposal was restored: code=%d response=%+v", code, response)
	}
}
