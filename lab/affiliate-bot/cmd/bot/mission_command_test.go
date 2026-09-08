package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
)

func missionCall(t *testing.T, args ...string) (int, map[string]any) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runMissionCommand(args, &stdout, &stderr)
	var envelope map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode mission response %q: %v (stderr: %s)", stdout.String(), err, stderr.String())
	}
	return code, envelope
}

func missionFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	if _, err := buildBR10AdvisorFixture(dir); err != nil {
		t.Fatal(err)
	}
	history := filepath.Join(dir, "history.jsonl")
	records, err := LoadHistory(history)
	if err != nil {
		t.Fatal(err)
	}
	request := filepath.Join(dir, "intent-request.json")
	value := learnerIntentRequest{
		IntentID: "path-safe-intent", DecisionID: records[0].RecordID,
		EvidenceIDs: records[0].RecordedResult.EvidenceIDs,
		ActionType:  "DRAFT", Target: "https://example.com/draft", Parameters: map[string]any{},
		ProposedBy: "human", CreatedAt: "2026-09-07T00:00:00Z", ExpiresAt: "2099-09-07T03:00:00Z",
		CorrelationID: "path-safe-correlation", IdempotencyKey: "path-safe-key",
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return dir, history, request
}

func TestMissionM08IntentRejectsInputOutputAliasesWithoutMutation(t *testing.T) {
	dir, history, request := missionFixture(t)
	historyBefore, err := os.ReadFile(history)
	if err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, history); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("history-as-output was not rejected: code=%d response=%+v", code, response)
	}
	if historyAfter, err := os.ReadFile(history); err != nil || !bytes.Equal(historyBefore, historyAfter) {
		t.Fatalf("history changed after rejected command: %v", err)
	}
	requestBefore, err := os.ReadFile(request)
	if err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, request); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("request-as-output was not rejected: code=%d response=%+v", code, response)
	}
	if requestAfter, err := os.ReadFile(request); err != nil || !bytes.Equal(requestBefore, requestAfter) {
		t.Fatalf("request changed after rejected command: %v", err)
	}

	alias := filepath.Join(dir, "history-alias.json")
	if err := os.Link(history, alias); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, alias); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("hardlink output was not rejected: code=%d response=%+v", code, response)
	}
	if historyAfter, err := os.ReadFile(history); err != nil || !bytes.Equal(historyBefore, historyAfter) {
		t.Fatalf("history changed through hardlink alias: %v", err)
	}

	symlink := filepath.Join(dir, "history-symlink.json")
	if err := os.Symlink(history, symlink); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, symlink); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("symlink output was not rejected: code=%d response=%+v", code, response)
	}
	if historyAfter, err := os.ReadFile(history); err != nil || !bytes.Equal(historyBefore, historyAfter) {
		t.Fatalf("history changed through symlink alias: %v", err)
	}
}

func TestMissionM08ArtifactOutputHasCreateRetryConflictSemantics(t *testing.T) {
	dir, history, request := missionFixture(t)
	intent := filepath.Join(dir, "intent.json")
	if code, response := missionCall(t, "m08-intent", history, request, intent); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("intent create failed: code=%d response=%+v", code, response)
	}
	first, err := os.ReadFile(intent)
	if err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, intent); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("intent retry was not exact duplicate: code=%d response=%+v", code, response)
	}
	if retry, err := os.ReadFile(intent); err != nil || !bytes.Equal(first, retry) {
		t.Fatalf("intent changed on retry: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "conflict.json"), []byte("keep-this-output"), 0600); err != nil {
		t.Fatal(err)
	}
	conflict := filepath.Join(dir, "conflict.json")
	if code, response := missionCall(t, "m08-intent", history, request, conflict); code == 0 || response["status"] != "CONFLICT" {
		t.Fatalf("existing output was not rejected: code=%d response=%+v", code, response)
	}
	if got, err := os.ReadFile(conflict); err != nil || string(got) != "keep-this-output" {
		t.Fatalf("existing output was overwritten: %q, %v", got, err)
	}

	policyConfig := filepath.Join(dir, "policy-input.json")
	if err := os.WriteFile(policyConfig, []byte(`{"policy_version":"path-safe-v1","now":"2026-09-07T01:00:00Z","allowed_hosts":["example.com"],"action_risk":{"DRAFT":"RISK0"},"seen_idempotency":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	policy := filepath.Join(dir, "policy-output.json")
	if code, response := missionCall(t, "m08-policy", intent, policyConfig, policy); code != 0 || response["status"] != "ALLOW" {
		t.Fatalf("policy create failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m08-policy", intent, policyConfig, policy); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("policy retry was not exact duplicate: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m08-policy", intent, policyConfig, intent); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("policy input-as-output was not rejected: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m08-policy", intent, policyConfig, policyConfig); code == 0 || response["status"] != "PATH_ERROR" {
		t.Fatalf("policy configuration-as-output was not rejected: code=%d response=%+v", code, response)
	}
}

func TestLearnerM08PreservesNumbersAndFailsClosedForInvalidProposals(t *testing.T) {
	dir, history, request := missionFixture(t)
	records, err := LoadHistory(history)
	if err != nil {
		t.Fatal(err)
	}
	valid := bytes.Replace([]byte(`{"intent_id":"number-intent","decision_id":"br11-decision","evidence_ids":["EVIDENCE_ID"],"action_type":"DRAFT","target":"https://example.com/draft","parameters":{"id":9007199254740993},"proposed_by":"human","created_at":"2026-09-07T00:00:00Z","expires_at":"2099-09-07T03:00:00Z","correlation_id":"number-correlation","idempotency_key":"number-key"}`), []byte("EVIDENCE_ID"), []byte(records[0].RecordedResult.EvidenceIDs[0]), 1)
	if err := os.WriteFile(request, valid, 0600); err != nil {
		t.Fatal(err)
	}
	intent := filepath.Join(dir, "number-intent.json")
	if code, response := missionCall(t, "m08-intent", history, request, intent); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("large-number intent rejected: code=%d response=%+v", code, response)
	}
	stored, err := os.ReadFile(intent)
	if err != nil || !bytes.Contains(stored, []byte(`9007199254740993`)) {
		t.Fatalf("large number was not preserved: %s, %v", stored, err)
	}
	if err := os.WriteFile(request, bytes.Replace(valid, []byte(`"parameters":{"id":9007199254740993}`), []byte(`"parameters":null`), 1), 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, filepath.Join(dir, "null.json")); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("null parameters accepted: code=%d response=%+v", code, response)
	}
	if err := os.WriteFile(request, bytes.Replace(valid, []byte(`"proposed_by":"human"`), []byte(`"proposed_by":"agent"`), 1), 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m08-intent", history, request, filepath.Join(dir, "agent.json")); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("unresolved agent proposal accepted: code=%d response=%+v", code, response)
	}
}

func TestMissionStateLockRejectsConcurrentMutationWithoutWriting(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	statePath := missionStatePath(dir)
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".mission.lock"), 0700); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-stop", dir, "concurrent-stop"); code == 0 || response["status"] != "BUSY" {
		t.Fatalf("locked mutation was accepted: code=%d response=%+v", code, response)
	}
	after, err := os.ReadFile(statePath)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatalf("locked mutation changed state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "STOP")); !os.IsNotExist(err) {
		t.Fatalf("locked mutation created STOP marker: %v", err)
	}
}

func TestMissionAuthorityRejectsExpiredApprovalBeforeReserve(t *testing.T) {
	now := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	s := LearnerMissionState{
		Intent:   &LearnerIntent{IntentID: "i", IntentHash: "h", ExpiresAt: "2026-09-08T01:00:00Z"},
		Policy:   &LearnerPolicy{PolicyVersion: "v1", PolicyCheckedAt: "2026-09-08T00:00:00Z"},
		Approval: &LearnerApproval{IntentID: "i", IntentHash: "h", PolicyVersion: "v1", CorrelationID: "", Decision: "APPROVE", OneTime: true, ExpiresAt: "2026-09-07T23:59:59Z"},
	}
	if err := missionAuthorityActive(s, now); err == nil {
		t.Fatal("expired approval was accepted")
	}
}

func TestTrustedCostBoundRegistryResolvesOnlyCanonicalEntry(t *testing.T) {
	dir := t.TempDir()
	bound := corem10.TrustedCostBound{CostBoundID: "cost-1", IntentID: "intent-1", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", MaxCostMinor: 100, Currency: "USD", SourceRef: "fixture:registry", ObservedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", CorrelationID: "corr-1", HashVersion: "go-json-v1"}
	bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
	raw, err := json.Marshal(bound)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trustedCostBoundsPath(dir), append(raw, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := registerM10Artifact(dir, corem10.ArtifactKindTrustedCostBound, raw); err != nil {
		t.Fatal(err)
	}
	if !resolveTrustedCostBound(dir, bound) {
		t.Fatal("registered canonical bound did not resolve")
	}
	bound.MaxCostMinor = 1
	if resolveTrustedCostBound(dir, bound) {
		t.Fatal("tampered bound resolved from registry")
	}
}

func TestM10ArtifactRegistryRejectsOrphanAuthorization(t *testing.T) {
	dir := t.TempDir()
	authorization := corem10.ExecutionAuthorization{
		AuthorizationID: "canary-auth-orphan", IntentID: "intent-1", IntentHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", PolicyVersion: "p1",
		CanaryGrantID: "grant-1", CanaryGrantVersion: "v1", CanaryGrantHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", CanaryGateID: "gate-1",
		CanaryCostBoundID: "cost-1", CanaryCostBoundHash: "sha256:0000000000000000000000000000000000000000000000000000000000000000", CanaryCostBoundMinor: 100,
		ExecutorID: "local_sandbox", AuthorizedAt: "2026-09-08T00:00:00Z", ExpiresAt: "2026-09-08T01:00:00Z", IdempotencyKey: "key", CorrelationID: "corr", ExecutionMode: "GOVERNED_CANARY", ExecutionAuthorized: true,
	}
	raw, err := json.Marshal(authorization)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := corem10.NewArtifactEntry(corem10.ArtifactKindExecutionAuthorization, raw)
	if err != nil {
		t.Fatal(err)
	}
	line, _ := json.Marshal(entry)
	if err := os.WriteFile(m10ArtifactRegistryPath(dir), append(line, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadM10ArtifactRegistry(dir); err == nil {
		t.Fatal("orphan authorization was accepted into M10 graph")
	}
}

func TestMissionM11RegistryUsesCanonicalCoreDecoder(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	lease := corem11.ProductionLease{LeaseID: "m11-lease", LeaseVersion: "v1", PolicyVersion: "policy-v1", ApprovalRef: "m11-approval", ReviewedBy: "human", ReviewerID: "reviewer", ReviewedAt: "2026-09-08T00:00:00Z", PromotionReviewRef: "review", SourceCanaryGrantID: "canary", SourceCanaryGrantVersion: "v1", SourceCanaryGrantHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ValidFrom: "2026-09-08T00:00:00Z", ExpiresAt: "2099-09-08T00:00:00Z", AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 1, MaxExecutionsPerWindow: 1, WindowSeconds: 60, MaxCostMinorTotal: 1, Currency: "USD", MaxPendingOutcomes: 1, MaxConsecutiveFailures: 1, MaxOutcomeAgeSeconds: 60, MaxHealthSnapshotAgeSeconds: 60, KillSwitchRequired: true, CorrelationID: "m11-correlation", HashVersion: "go-json-v1"}
	lease.LeaseHash = corem11.ComputeProductionLeaseHash(lease)
	raw, err := json.Marshal(lease)
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(dir, "lease.json")
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-register", dir, corem11.ArtifactKindLease, input); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("M11 lease registration failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-resolve", dir, corem11.ArtifactKindLease, lease.LeaseID); code != 0 || response["status"] != "RESOLVED" {
		t.Fatalf("M11 lease did not resolve: code=%d response=%+v", code, response)
	}
	lease.MaxCostMinorTotal = 2
	tampered, err := json.Marshal(lease)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, tampered, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-register", dir, corem11.ArtifactKindLease, input); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("tampered M11 lease was registered: code=%d response=%+v", code, response)
	}
}

func TestEvaluateLearnerPolicyRequiresReviewForRiskTwo(t *testing.T) {
	i := LearnerIntent{IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "PUBLISH", Target: "https://example.com/publish", Parameters: map[string]any{}, ProposedBy: "human", CreatedAt: "2099-01-01T00:00:00Z", ExpiresAt: "2099-01-01T02:00:00Z", CorrelationID: "c", IdempotencyKey: "k", IntentMode: "PROPOSAL_ONLY"}
	i.IntentHash = learnerIntentHash(i)
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(path, []byte(`{"policy_version":"v1","now":"2099-01-01T01:00:00Z","allowed_hosts":["example.com"],"action_risk":{"PUBLISH":"RISK2"},"seen_idempotency":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := evaluateLearnerPolicy(i, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.Decision != "HUMAN_REVIEW" || !p.PolicyReviewRequired || p.ExecutionAuthorized {
		t.Fatalf("unexpected risk-two policy: %+v", p)
	}
}

func TestLoadMissionStateRejectsUnreflectedStopMarker(t *testing.T) {
	dir := t.TempDir()
	state, _ := json.Marshal(LearnerMissionState{Version: missionStateVersion})
	if err := os.WriteFile(filepath.Join(dir, "mission-state.json"), state, 0600); err != nil {
		t.Fatal(err)
	}
	marker, _ := json.Marshal(map[string]any{"active": true, "reason": "test"})
	if err := os.WriteFile(filepath.Join(dir, "STOP"), marker, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadMissionState(dir); err == nil {
		t.Fatal("active STOP marker without stopped state was accepted")
	}
}

func TestMissionInitDoesNotOverwriteExistingState(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mission-state.json"), []byte(`{"version":"learner-m08-m11/v1","stop":true,"stop_reason":"keep"}`), 0600); err != nil {
		t.Fatal(err)
	}
	code := runMissionCommand([]string{"init", dir}, os.Stdout, os.Stderr)
	if code == 0 {
		t.Fatal("mission init overwrote an existing state")
	}
}
