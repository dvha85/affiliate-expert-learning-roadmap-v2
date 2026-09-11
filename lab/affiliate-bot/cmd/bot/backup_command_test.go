package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
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

func TestRestoreTargetGateRejectsConcurrentManagedPublisher(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "restored")
	release, err := acquireRestoreTargetGate(target)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := acquireRestoreTargetGate(target); err == nil {
		t.Fatal("second restore acquired the same target gate")
	}
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

func TestBackupRestoreRejectsExpiredM10AuthorityWithoutMutation(t *testing.T) {
	binary := buildMissionBinary(t)
	for _, expiring := range []string{"intent", "approval", "grant", "cost"} {
		t.Run(expiring, func(t *testing.T) {
			runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, expiring)
			root := filepath.Dir(runtimeDir)
			backupDir := filepath.Join(root, "backup")
			restoredDir := filepath.Join(root, "restored")
			if code, response := backupCall(t, "create", runtimeDir, backupDir); code != 0 || response["status"] != "BACKED_UP" {
				t.Fatalf("backup failed: code=%d response=%+v", code, response)
			}
			if code, response := backupCall(t, "restore", backupDir, restoredDir); code != 0 || response["status"] != "RESTORED" {
				t.Fatalf("restore failed: code=%d response=%+v", code, response)
			}
			before := missionRuntimeSnapshot(t, restoredDir)
			authorizationPath := filepath.Join(root, "restored-expired-authorization.json")
			if code, response := missionBinaryCall(t, binary, "mission", "m10-authorize", restoredDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code == 0 || response["status"] != "REJECTED" {
				t.Fatalf("restored expired %s authority was accepted: code=%d response=%+v", expiring, code, response)
			}
			if _, err := os.Stat(authorizationPath); !os.IsNotExist(err) {
				t.Fatalf("restored expired %s authority created portable output: %v", expiring, err)
			}
			assertMissionRuntimeUnchanged(t, before, restoredDir)
		})
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
	for _, name := range []string{"action-input.json", "outcome-input.json", "evaluation-config.json", "proposal.json", "review.json"} {
		if err := os.Remove(filepath.Join(runtime, name)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
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
	manifest.Files["outcomes.jsonl"], err = backupFileMetadata(outcomesPath, "outcomes.jsonl")
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
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
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

func TestBackupRestoreRejectsStateCanaryMissingRegistryGrant(t *testing.T) {
	dir := t.TempDir()
	if _, err := buildBR10AdvisorFixture(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(dir, name)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("mission init failed: code=%d response=%+v", code, response)
	}
	intent := LearnerIntent{IntentID: "registry-link-intent", DecisionID: "br11-decision", EvidenceIDs: []string{"br11-observation"}, ActionType: "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2026-09-01T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: "registry-link-correlation", IdempotencyKey: "registry-link-key", IntentMode: "PROPOSAL_ONLY"}
	intent.IntentHash = learnerIntentHash(intent)
	approval := LearnerApproval{ApprovalID: "registry-link-approval", IntentID: intent.IntentID, IntentHash: intent.IntentHash, PolicyVersion: "registry-link-policy", Decision: "APPROVE", ApprovedBy: "human", ApproverID: "registry-link-reviewer", ApprovedAt: "2026-09-01T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: intent.CorrelationID, OneTime: true}
	grant := corem10.CanaryGrant{GrantID: "registry-link-grant", GrantVersion: "v1", PolicyVersion: approval.PolicyVersion, ApprovalRef: approval.ApprovalID, ApprovedBy: "human", ApproverID: approval.ApproverID, ApprovedAt: approval.ApprovedAt, ValidFrom: approval.ApprovedAt, ExpiresAt: approval.ExpiresAt, AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: intent.CorrelationID, HashVersion: "go-json-v1"}
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
	backup := filepath.Join(filepath.Dir(dir), "registry-link-backup")
	if code, response := backupCall(t, "create", dir, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}
	registryPath := m10ArtifactRegistryPath(dir)
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(registryPath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", dir, filepath.Join(filepath.Dir(dir), "registry-link-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a state canary missing its registry grant: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(registryPath, registryRaw, 0600); err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(filepath.Dir(dir), "registry-link-broken")
	copyFlatBackup(t, backup, broken)
	registryPath = filepath.Join(broken, "m10-artifacts.jsonl")
	if err := os.WriteFile(registryPath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-artifacts.jsonl"], err = backupFileMetadata(registryPath, "m10-artifacts.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(filepath.Dir(dir), "registry-link-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted a state canary missing its registry grant: code=%d response=%+v", code, response)
	}
}

func removeM10ArtifactEntry(t *testing.T, path, kind, artifactID string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	kept := make([][]byte, 0)
	removed := false
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, err := corem10.ValidateArtifactEntry(line)
		if err != nil {
			t.Fatal(err)
		}
		if entry.ArtifactKind == kind && entry.ArtifactID == artifactID {
			removed = true
			continue
		}
		kept = append(kept, line)
	}
	if !removed {
		t.Fatalf("missing M10 artifact to remove: %s/%s", kind, artifactID)
	}
	if err := os.WriteFile(path, append(bytes.Join(kept, []byte{'\n'}), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func replaceM10ArtifactEntry(t *testing.T, path, kind, artifactID string, artifact []byte) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	replacement, err := corem10.NewArtifactEntry(kind, artifact)
	if err != nil {
		t.Fatal(err)
	}
	lines := make([][]byte, 0)
	replaced := false
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, err := corem10.ValidateArtifactEntry(line)
		if err != nil {
			t.Fatal(err)
		}
		if entry.ArtifactKind == kind && entry.ArtifactID == artifactID {
			line, err = json.Marshal(replacement)
			if err != nil {
				t.Fatal(err)
			}
			replaced = true
		}
		lines = append(lines, line)
	}
	if !replaced {
		t.Fatalf("missing M10 artifact to replace: %s/%s", kind, artifactID)
	}
	if err := os.WriteFile(path, append(bytes.Join(lines, []byte{'\n'}), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBackupRestoreRejectsReservationMissingRegistryAuthorization(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "orphaned-authorization-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	var authorization corem10.ExecutionAuthorization
	if err := readJSON(authorizationPath, &authorization); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(root, "authorization-link-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	registryPath := m10ArtifactRegistryPath(runtime)
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionAuthorization, authorization.AuthorizationID)
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "authorization-link-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a reservation missing its registry authorization: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(registryPath, registryRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "authorization-link-broken")
	copyFlatBackup(t, backup, broken)
	registryPath = filepath.Join(broken, "m10-artifacts.jsonl")
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionAuthorization, authorization.AuthorizationID)
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-artifacts.jsonl"], err = backupFileMetadata(registryPath, "m10-artifacts.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "authorization-link-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted a reservation missing its registry authorization: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsReservationMissingRegistryExecution(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "execution-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "orphaned-execution-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "execution-link-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	statePath := missionStatePath(runtime)
	stateRaw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	state, err := loadMissionState(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ExecutionID = "missing-registry-execution"
	if err := saveMissionState(runtime, state); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "execution-link-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a reservation missing its registry execution: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(statePath, stateRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "execution-link-broken")
	copyFlatBackup(t, backup, broken)
	state, err = loadMissionState(broken)
	if err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ExecutionID = "missing-registry-execution"
	if err := saveMissionState(broken, state); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["mission-state.json"], err = backupFileMetadata(missionStatePath(broken), "mission-state.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "execution-link-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted a reservation missing its registry execution: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsCanaryUsageMismatchedReservations(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "usage-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "usage-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "usage-ledger-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	statePath := missionStatePath(runtime)
	stateRaw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state LearnerMissionState
	if err := readJSON(statePath, &state); err != nil {
		t.Fatal(err)
	}
	state.Canary.ExecutionsUsed = 0
	state.Canary.CostUsedMinor = 0
	if err := saveMissionState(runtime, state); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "usage-ledger-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted canary usage below its reservations: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(statePath, stateRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "usage-ledger-broken")
	copyFlatBackup(t, backup, broken)
	if err := readJSON(missionStatePath(broken), &state); err != nil {
		t.Fatal(err)
	}
	state.Canary.ExecutionsUsed = 0
	state.Canary.CostUsedMinor = 0
	if err := saveMissionState(broken, state); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["mission-state.json"], err = backupFileMetadata(missionStatePath(broken), "mission-state.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "usage-ledger-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("restore accepted canary usage below its reservations: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsReservationOutsideAuthorizationLifetime(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "lifetime-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "lifetime-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "lifetime-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	statePath := missionStatePath(runtime)
	stateRaw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var state LearnerMissionState
	if err := readJSON(statePath, &state); err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ReservedAt = "2026-09-08T00:01:00Z"
	if err := saveMissionState(runtime, state); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "lifetime-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a reservation after authorization expiry: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(statePath, stateRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "lifetime-broken")
	copyFlatBackup(t, backup, broken)
	if err := readJSON(missionStatePath(broken), &state); err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ReservedAt = "2026-09-08T00:01:00Z"
	if err := saveMissionState(broken, state); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["mission-state.json"], err = backupFileMetadata(missionStatePath(broken), "mission-state.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "lifetime-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted a reservation after authorization expiry: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsReservationMissingRegistryCostBound(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	if code, response := missionCall(t, "m10-reserve", runtime, boundPath, "legacy-cost-bound-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	var bound corem10.TrustedCostBound
	if err := readJSON(boundPath, &bound); err != nil {
		t.Fatal(err)
	}
	var gate corem10.CanaryGateDecision
	if err := readJSON(gatePath, &gate); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(root, "cost-bound-link-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	registryPath := m10ArtifactRegistryPath(runtime)
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindCanaryGate, gate.GateID)
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindTrustedCostBound, bound.CostBoundID)
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "cost-bound-link-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted a reservation missing its registry cost bound: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(registryPath, registryRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "cost-bound-link-broken")
	copyFlatBackup(t, backup, broken)
	registryPath = filepath.Join(broken, "m10-artifacts.jsonl")
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindCanaryGate, gate.GateID)
	removeM10ArtifactEntry(t, registryPath, corem10.ArtifactKindTrustedCostBound, bound.CostBoundID)
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-artifacts.jsonl"], err = backupFileMetadata(registryPath, "m10-artifacts.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "cost-bound-link-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted a reservation missing its registry cost bound: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsDuplicateM10FixtureOutcomeForExecution(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "outcome-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "outcome-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	recordPath := filepath.Join(root, "failed-record.json")
	if code, response := missionCall(t, "m10-record-failed", runtime, authorizationPath, recordPath, "2026-09-08T00:00:01Z", "fixture failure"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("failed record setup failed: code=%d response=%+v", code, response)
	}
	var record corem10.ExecutionRecord
	if err := readJSON(recordPath, &record); err != nil {
		t.Fatal(err)
	}
	first := m03.OutcomeRecord{OutcomeID: "m10-fixture-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: record.ExecutionID}, ObservedAt: "2026-09-08T00:00:02Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m10-outcome/duplicate-link"}
	firstPath := filepath.Join(root, "first-outcome.json")
	if err := writeJSON(firstPath, first); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m10-outcome", runtime, firstPath); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("first outcome setup failed: code=%d response=%+v", code, response)
	}
	second := first
	second.OutcomeID = "m10-fixture-outcome-duplicate"
	secondPath := filepath.Join(root, "second-outcome.json")
	if err := writeJSON(secondPath, second); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m10-outcome", runtime, secondPath); code == 0 || response["status"] != "CONFLICT" {
		t.Fatalf("runtime accepted a second outcome for one execution: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "duplicate-outcome-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	outcomeStorePath := m10OutcomeStorePath(runtime)
	storeRaw, err := os.ReadFile(outcomeStorePath)
	if err != nil {
		t.Fatal(err)
	}
	secondRaw, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outcomeStorePath, append(storeRaw, append(secondRaw, '\n')...), 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "duplicate-outcome-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted duplicate M10 fixture outcomes: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(outcomeStorePath, storeRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "duplicate-outcome-broken")
	copyFlatBackup(t, backup, broken)
	outcomeStorePath = m10OutcomeStorePath(broken)
	if err := os.WriteFile(outcomeStorePath, append(storeRaw, append(secondRaw, '\n')...), 0600); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-outcomes.jsonl"], err = backupFileMetadata(outcomeStorePath, "m10-outcomes.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "duplicate-outcome-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted duplicate M10 fixture outcomes: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsExecutionOutsideAuthorizationLifetime(t *testing.T) {
	runtime, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "execution-lifetime-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "execution-lifetime-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	recordPath := filepath.Join(root, "execution-lifetime-record.json")
	if code, response := missionCall(t, "m10-record-failed", runtime, authorizationPath, recordPath, "2026-09-08T00:00:01Z", "fixture failure"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("record setup failed: code=%d response=%+v", code, response)
	}
	var record corem10.ExecutionRecord
	if err := readJSON(recordPath, &record); err != nil {
		t.Fatal(err)
	}
	outcome := m03.OutcomeRecord{OutcomeID: "execution-lifetime-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: record.ExecutionID}, ObservedAt: "2026-09-08T00:00:02Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m10-outcome/execution-lifetime"}
	outcomePath := filepath.Join(root, "execution-lifetime-outcome.json")
	if err := writeJSON(outcomePath, outcome); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m10-outcome", runtime, outcomePath); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("outcome setup failed: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "execution-lifetime-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	altered := record
	altered.AttemptedAt = "2026-09-08T02:00:00Z"
	alteredRaw, err := json.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	registryPath := m10ArtifactRegistryPath(runtime)
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	replaceM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionRecord, record.ExecutionID, alteredRaw)
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "execution-lifetime-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted an execution at authorization expiry: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(registryPath, registryRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "execution-lifetime-broken")
	copyFlatBackup(t, backup, broken)
	registryPath = m10ArtifactRegistryPath(broken)
	replaceM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionRecord, record.ExecutionID, alteredRaw)
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-artifacts.jsonl"], err = backupFileMetadata(registryPath, "m10-artifacts.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "execution-lifetime-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("restore accepted an execution at authorization expiry: code=%d response=%+v", code, response)
	}
}

func TestBackupRestoreRejectsExecutionBeforeReservation(t *testing.T) {
	runtime, boundPath, gatePath, expiryTime := authorityExpiryFixture(t, "cost")
	reservationTime := expiryTime.Add(-30 * time.Second)
	root := filepath.Dir(runtime)
	authorizationPath := filepath.Join(root, "execution-order-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtime, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	missionClock = func() time.Time { return reservationTime }
	if code, response := missionCall(t, "m10-reserve-authorization", runtime, authorizationPath, "execution-order-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	recordPath := filepath.Join(root, "execution-order-record.json")
	if code, response := missionCall(t, "m10-record-failed", runtime, authorizationPath, recordPath, reservationTime.Format(time.RFC3339), "fixture failure"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("record setup failed: code=%d response=%+v", code, response)
	}
	var record corem10.ExecutionRecord
	if err := readJSON(recordPath, &record); err != nil {
		t.Fatal(err)
	}
	outcome := m03.OutcomeRecord{OutcomeID: "execution-order-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: record.ExecutionID}, ObservedAt: "2026-09-08T00:01:01Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m10-outcome/execution-order"}
	outcomePath := filepath.Join(root, "execution-order-outcome.json")
	if err := writeJSON(outcomePath, outcome); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m10-outcome", runtime, outcomePath); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("outcome setup failed: code=%d response=%+v", code, response)
	}
	backup := filepath.Join(root, "execution-order-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed: code=%d response=%+v", code, response)
	}

	var authorization corem10.ExecutionAuthorization
	if err := readJSON(authorizationPath, &authorization); err != nil {
		t.Fatal(err)
	}
	altered, err := corem10.FailCanaryExecutionFixture(corem10.FailedExecutionInput{Authorization: authorization, AttemptedAt: "2026-09-08T00:00:15Z", Reason: record.Error})
	if err != nil {
		t.Fatal(err)
	}
	alteredRaw, err := json.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	registryPath := m10ArtifactRegistryPath(runtime)
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatal(err)
	}
	statePath := missionStatePath(runtime)
	stateRaw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	outcomeStorePath := m10OutcomeStorePath(runtime)
	outcomeStoreRaw, err := os.ReadFile(outcomeStorePath)
	if err != nil {
		t.Fatal(err)
	}
	var alteredOutcome m03.OutcomeRecord
	if err := json.Unmarshal(bytes.TrimSpace(outcomeStoreRaw), &alteredOutcome); err != nil {
		t.Fatal(err)
	}
	alteredOutcome.EffectRef.EffectID = altered.ExecutionID
	alteredOutcomeRaw, err := json.Marshal(alteredOutcome)
	if err != nil {
		t.Fatal(err)
	}
	replaceM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionRecord, record.ExecutionID, alteredRaw)
	state, err := loadMissionState(runtime)
	if err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ExecutionID = altered.ExecutionID
	if err := saveMissionState(runtime, state); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outcomeStorePath, append(alteredOutcomeRaw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(root, "execution-order-source-invalid")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted an execution before its reservation: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(registryPath, registryRaw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, stateRaw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outcomeStorePath, outcomeStoreRaw, 0600); err != nil {
		t.Fatal(err)
	}

	broken := filepath.Join(root, "execution-order-broken")
	copyFlatBackup(t, backup, broken)
	registryPath = m10ArtifactRegistryPath(broken)
	replaceM10ArtifactEntry(t, registryPath, corem10.ArtifactKindExecutionRecord, record.ExecutionID, alteredRaw)
	state, err = loadMissionState(broken)
	if err != nil {
		t.Fatal(err)
	}
	state.Reservations[0].ExecutionID = altered.ExecutionID
	if err := saveMissionState(broken, state); err != nil {
		t.Fatal(err)
	}
	outcomeStorePath = m10OutcomeStorePath(broken)
	if err := os.WriteFile(outcomeStorePath, append(alteredOutcomeRaw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	var manifest backupManifest
	if err := readJSON(filepath.Join(broken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-artifacts.jsonl"], err = backupFileMetadata(registryPath, "m10-artifacts.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	manifest.Files["mission-state.json"], err = backupFileMetadata(missionStatePath(broken), "mission-state.json")
	if err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-outcomes.jsonl"], err = backupFileMetadata(outcomeStorePath, "m10-outcomes.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(broken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", broken, filepath.Join(root, "execution-order-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted an execution before its reservation: code=%d response=%+v", code, response)
	}

	// A manifest checksum only proves the attacker updated the copy
	// consistently. The graph gate must still reject an outcome that claims to
	// have been observed before the execution it cites.
	outcomeBroken := filepath.Join(root, "execution-order-outcome-broken")
	copyFlatBackup(t, backup, outcomeBroken)
	outcomeStore := m10OutcomeStorePath(outcomeBroken)
	outcomeRaw, err := os.ReadFile(outcomeStore)
	if err != nil {
		t.Fatal(err)
	}
	var earlyOutcome m03.OutcomeRecord
	if err := json.Unmarshal(bytes.TrimSpace(outcomeRaw), &earlyOutcome); err != nil {
		t.Fatal(err)
	}
	earlyOutcome.ObservedAt = "2026-09-08T00:00:00Z"
	earlyOutcomeRaw, err := json.Marshal(earlyOutcome)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outcomeStore, append(earlyOutcomeRaw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(filepath.Join(outcomeBroken, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Files["m10-outcomes.jsonl"], err = backupFileMetadata(outcomeStore, "m10-outcomes.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(outcomeBroken, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", outcomeBroken, filepath.Join(root, "execution-order-outcome-restored")); code == 0 || response["status"] != "GRAPH_FAILED" {
		t.Fatalf("restore accepted an outcome before its execution: code=%d response=%+v", code, response)
	}
}

func TestBackupRejectsUnknownAndUninventoriedArtifacts(t *testing.T) {
	runtime := t.TempDir()
	if _, err := buildBR10AdvisorFixture(runtime); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"action-input.json", "outcome-input.json"} {
		if err := os.Remove(filepath.Join(runtime, name)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	if code, response := missionCall(t, "init", runtime); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("mission init failed: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(filepath.Join(runtime, "operator-notes.txt"), []byte("not a canonical artifact\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "create", runtime, filepath.Join(filepath.Dir(runtime), "unknown-artifact-backup")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup accepted unsupported runtime file: code=%d response=%+v", code, response)
	}
	if err := os.Remove(filepath.Join(runtime, "operator-notes.txt")); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(filepath.Dir(runtime), "typed-backup")
	if code, response := backupCall(t, "create", runtime, backup); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("typed backup failed: code=%d response=%+v", code, response)
	}
	knownExtra := filepath.Join(filepath.Dir(runtime), "known-extra-backup")
	copyFlatBackup(t, backup, knownExtra)
	if err := os.WriteFile(filepath.Join(knownExtra, "reviews.jsonl"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", knownExtra, filepath.Join(filepath.Dir(runtime), "known-extra-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("restore accepted uninventoried known artifact: code=%d response=%+v", code, response)
	}
	wrongKind := filepath.Join(filepath.Dir(runtime), "wrong-kind-backup")
	copyFlatBackup(t, backup, wrongKind)
	var manifest backupManifest
	if err := readJSON(filepath.Join(wrongKind, "manifest.json"), &manifest); err != nil {
		t.Fatal(err)
	}
	metadata := manifest.Files["history.jsonl"]
	metadata.Kind = "M03_ACTION_STORE"
	manifest.Files["history.jsonl"] = metadata
	if err := writeJSON(filepath.Join(wrongKind, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", wrongKind, filepath.Join(filepath.Dir(runtime), "wrong-kind-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("restore accepted a wrong artifact kind: code=%d response=%+v", code, response)
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
	if _, err := os.Stat(interrupted); !os.IsNotExist(err) {
		t.Fatalf("interrupted backup occupied its target instead of cleaning staging: %v", err)
	}
	if code, response := backupCall(t, "restore", interrupted, filepath.Join(root, "interrupted-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("interrupted snapshot was restorable: code=%d response=%+v", code, response)
	}
	if code, response := backupCall(t, "create", runtime, interrupted); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("retry after interrupted backup did not publish a complete snapshot: code=%d response=%+v", code, response)
	}

	missionBytes, err := os.ReadFile(filepath.Join(runtime, "mission-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	proposalBytes, err := os.ReadFile(proposalPath)
	if err != nil {
		t.Fatal(err)
	}
	multiFileConflict := filepath.Join(root, "multi-file-conflict")
	changed := false
	backupCopyFault = func(name string) error {
		if name != "mission-state.json" || changed {
			return nil
		}
		changed = true
		if err := os.WriteFile(filepath.Join(runtime, "mission-state.json"), append(missionBytes, '\n'), 0600); err != nil {
			return err
		}
		return os.WriteFile(proposalPath, append(proposalBytes, '\n'), 0600)
	}
	if code, response := backupCall(t, "create", runtime, multiFileConflict); code == 0 || response["status"] != "SNAPSHOT_CONFLICT" {
		t.Fatalf("backup accepted a mixed multi-file snapshot: code=%d response=%+v", code, response)
	}
	backupCopyFault = nil
	if err := os.WriteFile(filepath.Join(runtime, "mission-state.json"), missionBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proposalPath, proposalBytes, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(multiFileConflict, "manifest.json")); !os.IsNotExist(err) {
		t.Fatalf("multi-file conflict published a manifest: %v", err)
	}
	if code, response := backupCall(t, "restore", multiFileConflict, filepath.Join(root, "multi-file-conflict-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("mixed multi-file snapshot was restorable: code=%d response=%+v", code, response)
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
	manifest := backupManifest{Version: backupManifestVersion, Files: map[string]backupFile{}, Required: []string{"history.jsonl", "mission-state.json", "history.jsonl.m07/proposals/" + proposal.ProposalID[len("sha256:"):] + ".json", "history.jsonl.m07/tool-results/" + tool.TraceID[len("sha256:"):] + ".json"}}
	manifest.Profile, err = backupProfileFor(manifest.Required, files)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		metadata, err := backupFileMetadata(filepath.Join(badBackup, filepath.FromSlash(name)), name)
		if err != nil {
			t.Fatal(err)
		}
		manifest.Files[name] = metadata
	}
	if err := writeJSON(filepath.Join(badBackup, "manifest.json"), manifest); err != nil {
		t.Fatal(err)
	}
	if code, response := backupCall(t, "restore", badBackup, filepath.Join(root, "bad-restored")); code == 0 || response["status"] != "VERIFY_FAILED" {
		t.Fatalf("checksum-valid forged M07 proposal was restored: code=%d response=%+v", code, response)
	}
}
