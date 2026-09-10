package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
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

func writeMissionTestJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
}

func missionRuntimeSnapshot(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	result := map[string][]byte{}
	for _, name := range []string{"mission-state.json", "m10-artifacts.jsonl", "trusted-cost-bounds.jsonl"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read runtime snapshot %s: %v", name, err)
		}
		result[name] = raw
	}
	return result
}

func assertMissionRuntimeUnchanged(t *testing.T, before map[string][]byte, dir string) {
	t.Helper()
	after := missionRuntimeSnapshot(t, dir)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejected authority operation changed canonical runtime state")
	}
}

func buildMissionBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "bot")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build mission binary: %v: %s", err, output)
	}
	return binary
}

func missionBinaryCall(t *testing.T, binary string, args ...string) (int, map[string]any) {
	t.Helper()
	command := exec.Command(binary, args...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	response := map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode mission binary response %q: %v (stderr: %s)", stdout.String(), err, stderr.String())
	}
	return code, response
}

// TestMissionAuthorityExpirySubprocessHelper is a test-binary-only clock seam.
// The production Bot neither reads this environment variable nor accepts a
// caller-controlled clock. It lets the combined backup/restore regression run
// the real command dispatch in a fresh process at exact expiry boundaries.
func TestMissionAuthorityExpirySubprocessHelper(t *testing.T) {
	if os.Getenv("GO_WANT_AUTHORITY_EXPIRY_SUBPROCESS") != "1" {
		return
	}
	clockValue, err := time.Parse(time.RFC3339Nano, os.Getenv("GO_AUTHORITY_EXPIRY_CLOCK"))
	if err != nil {
		os.Exit(2)
	}
	separator := -1
	for index, value := range os.Args {
		if value == "--" {
			separator = index
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		os.Exit(2)
	}
	previousClock := missionClock
	missionClock = func() time.Time { return clockValue }
	defer func() { missionClock = previousClock }()
	os.Exit(runMissionCommand(os.Args[separator+1:], os.Stdout, os.Stderr))
}

func missionExpirySubprocessCall(t *testing.T, clock time.Time, args ...string) (int, map[string]any) {
	t.Helper()
	command := exec.Command(os.Args[0], append([]string{"-test.run=^TestMissionAuthorityExpirySubprocessHelper$", "--"}, args...)...)
	command.Env = append(os.Environ(), "GO_WANT_AUTHORITY_EXPIRY_SUBPROCESS=1", "GO_AUTHORITY_EXPIRY_CLOCK="+clock.UTC().Format(time.RFC3339Nano))
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = exit.ExitCode()
	}
	response := map[string]any{}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode expiry subprocess response %q: %v (stderr: %s)", stdout.String(), err, stderr.String())
	}
	return code, response
}

// authorityExpiryFixture drives the actual learner Bot M08→M10 admission
// path. The unexported mission clock is a test-only runtime seam; command
// arguments never select it.
func authorityExpiryFixture(t *testing.T, expiring string) (string, string, string, time.Time) {
	t.Helper()
	base := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	far := base.Add(2 * time.Hour).Format(time.RFC3339)
	boundary := base.Add(time.Minute)
	expires := map[string]string{"intent": far, "approval": far, "grant": far, "cost": far}
	expires[expiring] = boundary.Format(time.RFC3339)
	clock := base
	previousClock := missionClock
	missionClock = func() time.Time { return clock }
	t.Cleanup(func() { missionClock = previousClock })

	dir, history, requestPath := missionFixture(t)
	runtimeDir := filepath.Join(dir, "runtime")
	var request learnerIntentRequest
	if err := readJSON(requestPath, &request); err != nil {
		t.Fatal(err)
	}
	request.CreatedAt = base.Add(-time.Minute).Format(time.RFC3339)
	request.ExpiresAt = expires["intent"]
	writeMissionTestJSON(t, requestPath, request)
	intentPath := filepath.Join(dir, "intent.json")
	if code, response := missionCall(t, "m08-intent", history, requestPath, intentPath); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("intent setup failed: code=%d response=%+v", code, response)
	}
	// The backup validator resolves the intent's decision from canonical history,
	// so place the real history artifact in this runtime before binding it.
	historyRaw, err := os.ReadFile(history)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runtimeDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runtimeDir, "history.jsonl"), historyRaw, 0600); err != nil {
		t.Fatal(err)
	}
	policyPath := filepath.Join(dir, "policy.json")
	policyOut := filepath.Join(dir, "policy-out.json")
	writeMissionTestJSON(t, policyPath, learnerPolicyRequest{PolicyVersion: "expiry-policy", Now: base.Format(time.RFC3339), AllowedHosts: []string{"example.com"}, ActionRisk: map[string]string{"DRAFT": "RISK0"}, SeenIdempotency: map[string]string{}})
	if code, response := missionCall(t, "m08-policy", intentPath, policyPath, policyOut); code != 0 || response["status"] != "ALLOW" {
		t.Fatalf("policy setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "bind", runtimeDir, intentPath, policyOut); code != 0 || response["status"] != "BOUND" {
		t.Fatalf("bind setup failed: code=%d response=%+v", code, response)
	}
	var intent LearnerIntent
	var policy LearnerPolicy
	if err := readJSON(intentPath, &intent); err != nil {
		t.Fatal(err)
	}
	if err := readJSON(policyOut, &policy); err != nil {
		t.Fatal(err)
	}
	approvalPath := filepath.Join(dir, "approval.json")
	writeMissionTestJSON(t, approvalPath, LearnerApproval{ApprovalID: "expiry-approval", IntentID: intent.IntentID, IntentHash: intent.IntentHash, PolicyVersion: policy.PolicyVersion, Decision: "APPROVE", ApprovedBy: "human", ApproverID: "expiry-reviewer", ApprovedAt: base.Format(time.RFC3339), ExpiresAt: expires["approval"], CorrelationID: intent.CorrelationID, OneTime: true})
	if code, response := missionCall(t, "m09-approval", runtimeDir, approvalPath); code != 0 || response["status"] != "ACK" {
		t.Fatalf("approval setup failed: code=%d response=%+v", code, response)
	}
	grant := corem10.CanaryGrant{GrantID: "expiry-grant", GrantVersion: "v1", PolicyVersion: policy.PolicyVersion, ApprovalRef: "expiry-approval", ApprovedBy: "human", ApproverID: "expiry-reviewer", ApprovedAt: base.Format(time.RFC3339), ValidFrom: base.Format(time.RFC3339), ExpiresAt: expires["grant"], AllowedRiskClasses: []string{"RISK0"}, AllowedActionTypes: []string{"DRAFT"}, AllowedHosts: []string{"example.com"}, ExecutorIDs: []string{"fixture_stub"}, MaxExecutionsTotal: 2, MaxExecutionsPerWindow: 2, WindowSeconds: 60, MaxCostMinorTotal: 10, Currency: "USD", MaxPendingOutcomes: 1, KillSwitchRequired: true, CorrelationID: intent.CorrelationID, HashVersion: "go-json-v1"}
	grant.GrantHash = corem10.ComputeCanaryGrantHash(grant)
	grantPath := filepath.Join(dir, "grant.json")
	writeMissionTestJSON(t, grantPath, grant)
	if code, response := missionCall(t, "m10-canary", runtimeDir, grantPath); code != 0 || response["status"] != "ACK" {
		t.Fatalf("grant setup failed: code=%d response=%+v", code, response)
	}
	bound := corem10.TrustedCostBound{CostBoundID: "expiry-cost", IntentID: intent.IntentID, IntentHash: intent.IntentHash, MaxCostMinor: 1, Currency: "USD", SourceRef: "fixture:expiry", ObservedAt: base.Format(time.RFC3339), ExpiresAt: expires["cost"], CorrelationID: intent.CorrelationID, HashVersion: "go-json-v1"}
	bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
	boundPath := filepath.Join(dir, "cost.json")
	writeMissionTestJSON(t, boundPath, bound)
	if code, response := missionCall(t, "m10-cost-register", runtimeDir, boundPath); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("cost-bound setup failed: code=%d response=%+v", code, response)
	}
	gatePath := filepath.Join(dir, "gate.json")
	if code, response := missionCall(t, "m10-gate", runtimeDir, boundPath, gatePath, base.Format(time.RFC3339)); code != 0 || response["status"] != "ALLOW_CANARY" {
		t.Fatalf("pre-expiry gate failed: code=%d response=%+v", code, response)
	}
	return runtimeDir, boundPath, gatePath, boundary
}

func TestMissionM10AuthorityExpiryRejectsWithoutMutation(t *testing.T) {
	for _, expiring := range []string{"intent", "approval", "grant", "cost"} {
		t.Run(expiring, func(t *testing.T) {
			runtimeDir, boundPath, gatePath, boundary := authorityExpiryFixture(t, expiring)
			clock := boundary
			previousClock := missionClock
			missionClock = func() time.Time { return clock }
			t.Cleanup(func() { missionClock = previousClock })
			before := missionRuntimeSnapshot(t, runtimeDir)
			for _, now := range []time.Time{boundary, boundary.Add(time.Nanosecond)} {
				clock = now
				authorizationPath := filepath.Join(filepath.Dir(runtimeDir), "authorization-"+now.Format("150405.000000000")+".json")
				if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code == 0 || response["status"] != "REJECTED" {
					t.Fatalf("expired %s authority was accepted at %s: code=%d response=%+v", expiring, now, code, response)
				}
				if _, err := os.Stat(authorizationPath); !os.IsNotExist(err) {
					t.Fatalf("expired %s authority created portable output: %v", expiring, err)
				}
				assertMissionRuntimeUnchanged(t, before, runtimeDir)
			}
		})
	}
}

func TestMissionM10AuthorityExpiryRejectsInFreshProcessWithoutMutation(t *testing.T) {
	binary := buildMissionBinary(t)
	for _, expiring := range []string{"intent", "approval", "grant", "cost"} {
		t.Run(expiring, func(t *testing.T) {
			runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, expiring)
			before := missionRuntimeSnapshot(t, runtimeDir)
			authorizationPath := filepath.Join(filepath.Dir(runtimeDir), "fresh-process-authorization.json")
			// The child has the normal production wall clock. Its authority can
			// only be expired; no test clock reaches the command-line process.
			if code, response := missionBinaryCall(t, binary, "mission", "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code == 0 || response["status"] != "REJECTED" {
				t.Fatalf("fresh process accepted expired %s authority: code=%d response=%+v", expiring, code, response)
			}
			if _, err := os.Stat(authorizationPath); !os.IsNotExist(err) {
				t.Fatalf("fresh process created output for expired %s authority: %v", expiring, err)
			}
			assertMissionRuntimeUnchanged(t, before, runtimeDir)
		})
	}
}

func TestMissionM10AuthorityExpiryBoundariesAfterBackupRestoreInFreshProcess(t *testing.T) {
	for _, expiring := range []string{"intent", "approval", "grant", "cost"} {
		for _, boundaryCase := range []struct {
			name   string
			delta  time.Duration
			status string
			code   int
		}{
			{name: "before", delta: -time.Nanosecond, status: "AUTHORIZED", code: 0},
			{name: "at", delta: 0, status: "REJECTED", code: 1},
			{name: "after", delta: time.Nanosecond, status: "REJECTED", code: 1},
		} {
			t.Run(expiring+"/"+boundaryCase.name, func(t *testing.T) {
				runtimeDir, boundPath, gatePath, boundary := authorityExpiryFixture(t, expiring)
				root := filepath.Dir(runtimeDir)
				backupDir := filepath.Join(root, "boundary-backup")
				restoredDir := filepath.Join(root, "boundary-restored")
				if code, response := backupCall(t, "create", runtimeDir, backupDir); code != 0 || response["status"] != "BACKED_UP" {
					t.Fatalf("backup failed: code=%d response=%+v", code, response)
				}
				if code, response := backupCall(t, "restore", backupDir, restoredDir); code != 0 || response["status"] != "RESTORED" {
					t.Fatalf("restore failed: code=%d response=%+v", code, response)
				}
				outputPath := filepath.Join(root, "boundary-"+boundaryCase.name+".json")
				before := missionRuntimeSnapshot(t, restoredDir)
				code, response := missionExpirySubprocessCall(t, boundary.Add(boundaryCase.delta), "m10-authorize", restoredDir, boundPath, gatePath, outputPath, "2026-09-08T00:00:00Z", "fixture_stub")
				if code != boundaryCase.code || response["status"] != boundaryCase.status {
					t.Fatalf("%s %s boundary mismatch: code=%d response=%+v", expiring, boundaryCase.name, code, response)
				}
				if boundaryCase.code != 0 {
					if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
						t.Fatalf("expired %s %s boundary created portable output: %v", expiring, boundaryCase.name, err)
					}
					assertMissionRuntimeUnchanged(t, before, restoredDir)
				}
			})
		}
	}
}

func TestMissionM10RecordRejectsAuthorizationExpiryWithoutMutation(t *testing.T) {
	runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtimeDir)
	authorizationPath := filepath.Join(root, "record-expiry-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "record-expiry-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	before := missionRuntimeSnapshot(t, runtimeDir)
	recordPath := filepath.Join(root, "expired-record.json")
	if code, response := missionCall(t, "m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T02:00:00Z", "attempt after authorization expiry"); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("record at authorization expiry was accepted: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(recordPath); !os.IsNotExist(err) {
		t.Fatalf("expired authorization record created output: %v", err)
	}
	assertMissionRuntimeUnchanged(t, before, runtimeDir)
}

func TestMissionM10RecordRetriesAfterRegistryStateCommitFault(t *testing.T) {
	runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtimeDir)
	authorizationPath := filepath.Join(root, "record-fault-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "record-fault-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	beforeState := missionRuntimeSnapshot(t, runtimeDir)
	recordPath := filepath.Join(root, "record-fault.json")
	atomicWrites := 0
	missionStateWriteFault = func(phase string) error {
		if phase == "before_rename" {
			atomicWrites++ // journal, then reservation binding
			if atomicWrites == 2 {
				return errors.New("injected state commit fault after M10 registry append")
			}
		}
		return nil
	}
	if code, response := missionCall(t, "m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T00:00:00Z", "fixture state commit fault"); code == 0 || response["status"] != "STORE_ERROR" {
		t.Fatalf("state commit fault was not surfaced: code=%d response=%+v", code, response)
	}
	missionStateWriteFault = nil
	if _, err := os.Stat(recordPath); !os.IsNotExist(err) {
		t.Fatalf("state commit fault created portable output: %v", err)
	}
	if _, err := os.Stat(m10ExecutionJournalPath(runtimeDir)); err != nil {
		t.Fatalf("M10 journal was not retained after state commit fault: %v", err)
	}
	if code, response := missionCall(t, "status", runtimeDir); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("status did not fail closed on M10 execution journal: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-resolve", runtimeDir, corem10.ArtifactKindExecutionRecord, "unresolved-while-journal-pending"); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("M10 resolver exposed a partial execution journal: code=%d response=%+v", code, response)
	}
	afterFaultState := missionRuntimeSnapshot(t, runtimeDir)
	if !bytes.Equal(beforeState["mission-state.json"], afterFaultState["mission-state.json"]) {
		t.Fatal("registry/state fault changed mutable mission state")
	}
	if bytes.Equal(beforeState["m10-artifacts.jsonl"], afterFaultState["m10-artifacts.jsonl"]) {
		t.Fatal("fixture did not reach the registry-before-state failure seam")
	}
	if code, response := backupCall(t, "create", runtimeDir, filepath.Join(root, "record-fault-backup-recovery")); code == 0 || response["status"] != "INPUT_ERROR" {
		t.Fatalf("backup did not recover the M10 journal before fixture-graph validation: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m10ExecutionJournalPath(runtimeDir)); !os.IsNotExist(err) {
		t.Fatalf("backup did not complete the valid M10 journal recovery: %v", err)
	}
	recoveredState, err := loadMissionState(runtimeDir)
	if err != nil || len(recoveredState.Reservations) != 1 || recoveredState.Reservations[0].ExecutionID == "" {
		t.Fatalf("backup journal recovery did not bind the execution: state=%+v err=%v", recoveredState, err)
	}
	if code, response := missionCall(t, "m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T00:00:00Z", "fixture state commit fault"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("locked writer did not recover registry/state link before exact retry: code=%d response=%+v", code, response)
	} else {
		if _, err := os.Stat(m10ExecutionJournalPath(runtimeDir)); !os.IsNotExist(err) {
			t.Fatalf("M10 journal remains after locked recovery: %v", err)
		}
		artifact, ok := response["artifact"].(map[string]any)
		if !ok {
			t.Fatalf("record response has no artifact: %+v", response)
		}
		executionID, ok := artifact["execution_id"].(string)
		if !ok {
			t.Fatalf("record response has no execution ID: %+v", response)
		}
		outcomePath := filepath.Join(root, "record-fault-outcome.json")
		writeMissionTestJSON(t, outcomePath, m03.OutcomeRecord{OutcomeID: "record-fault-outcome", EffectRef: m03.EffectRef{EffectKind: "MACHINE_EXECUTION", EffectID: executionID}, ObservedAt: "2026-09-08T00:00:00Z", Status: "CANCELLED", Metrics: map[string]float64{}, SourceRef: "fixture:m10-outcome/record-fault"})
		if code, response := missionCall(t, "m10-outcome", runtimeDir, outcomePath); code != 0 || response["status"] != "APPENDED" {
			t.Fatalf("outcome setup after retry failed: code=%d response=%+v", code, response)
		}
	}
	if code, response := backupCall(t, "create", runtimeDir, filepath.Join(root, "record-fault-backup-after-retry")); code != 0 || response["status"] != "BACKED_UP" {
		t.Fatalf("backup failed after exact retry repaired state: code=%d response=%+v", code, response)
	}
}

func TestMissionM10ExecutionJournalTamperFailsClosed(t *testing.T) {
	runtimeDir, boundPath, gatePath, _ := authorityExpiryFixture(t, "cost")
	root := filepath.Dir(runtimeDir)
	authorizationPath := filepath.Join(root, "journal-tamper-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "journal-tamper-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	authorizationRaw, err := os.ReadFile(authorizationPath)
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := corem10.ValidateExecutionAuthorization(authorizationRaw)
	if err != nil {
		t.Fatal(err)
	}
	record, err := corem10.FailCanaryExecutionFixture(corem10.FailedExecutionInput{Authorization: authorization, AttemptedAt: "2026-09-08T00:00:00Z", Reason: "fixture journal tamper"})
	if err != nil {
		t.Fatal(err)
	}
	journal := m10ExecutionJournal{Version: "m10-execution-journal/v1", ReservationID: "missing-reservation", AuthorizationID: authorization.AuthorizationID, Record: record}
	if err := writeJSONAtomic(m10ExecutionJournalPath(runtimeDir), journal); err != nil {
		t.Fatal(err)
	}
	before := missionRuntimeSnapshot(t, runtimeDir)
	outputPath := filepath.Join(root, "journal-tamper-output.json")
	if code, response := missionCall(t, "m10-record-failed", runtimeDir, authorizationPath, outputPath, "2026-09-08T00:00:00Z", "fixture journal tamper"); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("writer accepted a tampered M10 execution journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("tampered journal created portable output: %v", err)
	}
	if code, response := backupCall(t, "create", runtimeDir, filepath.Join(root, "journal-tamper-backup")); code == 0 || response["status"] != "RECOVERY_REQUIRED" {
		t.Fatalf("backup accepted a tampered M10 execution journal: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m10ExecutionJournalPath(runtimeDir)); err != nil {
		t.Fatalf("tampered M10 journal was removed: %v", err)
	}
	after := missionRuntimeSnapshot(t, runtimeDir)
	if !bytes.Equal(before["mission-state.json"], after["mission-state.json"]) || !bytes.Equal(before["m10-artifacts.jsonl"], after["m10-artifacts.jsonl"]) {
		t.Fatal("tampered M10 journal mutated canonical state")
	}
}

func TestMissionM10RecordRejectsBeforeReservationWithoutMutation(t *testing.T) {
	runtimeDir, boundPath, gatePath, expiryTime := authorityExpiryFixture(t, "cost")
	reservationTime := expiryTime.Add(-30 * time.Second)
	root := filepath.Dir(runtimeDir)
	authorizationPath := filepath.Join(root, "record-order-authorization.json")
	if code, response := missionCall(t, "m10-authorize", runtimeDir, boundPath, gatePath, authorizationPath, "2026-09-08T00:00:00Z", "fixture_stub"); code != 0 || response["status"] != "AUTHORIZED" {
		t.Fatalf("authorization setup failed: code=%d response=%+v", code, response)
	}
	missionClock = func() time.Time { return reservationTime }
	if code, response := missionCall(t, "m10-reserve-authorization", runtimeDir, authorizationPath, "record-order-reservation"); code != 0 || response["status"] != "RESERVED" {
		t.Fatalf("reservation setup failed: code=%d response=%+v", code, response)
	}
	before := missionRuntimeSnapshot(t, runtimeDir)
	recordPath := filepath.Join(root, "before-reservation-record.json")
	if code, response := missionCall(t, "m10-record-failed", runtimeDir, authorizationPath, recordPath, "2026-09-08T00:00:15Z", "attempt before reservation"); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("record before reservation was accepted: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(recordPath); !os.IsNotExist(err) {
		t.Fatalf("pre-reservation record created output: %v", err)
	}
	assertMissionRuntimeUnchanged(t, before, runtimeDir)
}

func TestMissionIntentGrantAndCostRebindRejectWithoutMutation(t *testing.T) {
	runtimeDir, boundPath, _, _ := authorityExpiryFixture(t, "cost")
	before := missionRuntimeSnapshot(t, runtimeDir)
	state, err := loadMissionState(runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	// Rebinding the same intent ID with another sealed expiry must fail before
	// it can clear the existing approval/grant/reservation lineage.
	alteredIntent := *state.Intent
	alteredIntent.ExpiresAt = "2026-09-08T03:00:00Z"
	alteredIntent.IntentHash = learnerIntentHash(alteredIntent)
	alteredIntentPath := filepath.Join(filepath.Dir(runtimeDir), "altered-intent.json")
	writeMissionTestJSON(t, alteredIntentPath, alteredIntent)
	policyPath := filepath.Join(filepath.Dir(runtimeDir), "rebind-policy.json")
	writeMissionTestJSON(t, policyPath, LearnerPolicy{PolicyVersion: state.Policy.PolicyVersion, IntentID: alteredIntent.IntentID, IntentHash: alteredIntent.IntentHash, Decision: "ALLOW", RiskClass: "RISK0", PolicyCheckedAt: "2026-09-08T00:00:00Z"})
	if code, response := missionCall(t, "bind", runtimeDir, alteredIntentPath, policyPath); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("altered same-ID intent was rebound: code=%d response=%+v", code, response)
	}
	assertMissionRuntimeUnchanged(t, before, runtimeDir)

	grant := state.Canary.CanaryGrant
	grant.MaxCostMinorTotal++
	grant.GrantHash = corem10.ComputeCanaryGrantHash(grant)
	grantPath := filepath.Join(filepath.Dir(runtimeDir), "altered-grant.json")
	writeMissionTestJSON(t, grantPath, grant)
	if code, response := missionCall(t, "m10-canary", runtimeDir, grantPath); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("altered same-ID grant was rebound: code=%d response=%+v", code, response)
	}
	assertMissionRuntimeUnchanged(t, before, runtimeDir)

	var bound corem10.TrustedCostBound
	if err := readJSON(boundPath, &bound); err != nil {
		t.Fatal(err)
	}
	bound.Currency = "VND"
	bound.CostBoundHash = corem10.ComputeTrustedCostBoundHash(bound)
	alteredBoundPath := filepath.Join(filepath.Dir(runtimeDir), "altered-cost.json")
	writeMissionTestJSON(t, alteredBoundPath, bound)
	if code, response := missionCall(t, "m10-cost-register", runtimeDir, alteredBoundPath); code == 0 || response["status"] != "REJECTED" {
		t.Fatalf("altered same-ID cost bound was rebound: code=%d response=%+v", code, response)
	}
	assertMissionRuntimeUnchanged(t, before, runtimeDir)
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
	approval := corem11.ProductionLeaseApproval{ApprovalID: lease.ApprovalRef, LeaseID: lease.LeaseID, LeaseVersion: lease.LeaseVersion, LeaseHash: lease.LeaseHash, PromotionReviewRef: lease.PromotionReviewRef, SourceCanaryGrantID: lease.SourceCanaryGrantID, SourceCanaryGrantVersion: lease.SourceCanaryGrantVersion, SourceCanaryGrantHash: lease.SourceCanaryGrantHash, SourceE5Refs: []string{"fixture:e5"}, ValidatedRiskClasses: []string{"RISK0"}, ReviewedBy: "human", ReviewerID: lease.ReviewerID, ReviewedAt: lease.ReviewedAt, Decision: "APPROVE_PRODUCTION_LEASE"}
	approvalRaw, err := json.Marshal(approval)
	if err != nil {
		t.Fatal(err)
	}
	approvalInput := filepath.Join(dir, "approval.json")
	if err := os.WriteFile(approvalInput, approvalRaw, 0600); err != nil {
		t.Fatal(err)
	}
	if code, response := missionCall(t, "m11-register", dir, corem11.ArtifactKindLeaseApproval, approvalInput); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("M11 approval registration failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-activate", dir, lease.LeaseID, "2026-09-08T00:00:01Z"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("M11 activation failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-activate", dir, lease.LeaseID, "2026-09-08T00:00:01Z"); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("M11 activation retry failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-ledger-init", dir, lease.LeaseID, "2026-09-08T00:00:01Z"); code != 0 || response["status"] != "APPENDED" {
		t.Fatalf("M11 ledger initialization failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-ledger-init", dir, lease.LeaseID, "2026-09-08T00:00:01Z"); code != 0 || response["status"] != "EXACT_DUPLICATE" {
		t.Fatalf("M11 ledger retry failed: code=%d response=%+v", code, response)
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
	if code, response := missionCall(t, "m11-stop", dir, "runbook-stop-drill"); code != 0 || response["status"] != "STOPPED" {
		t.Fatalf("M11 stop failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-activate", dir, lease.LeaseID, "2026-09-08T00:00:01Z"); code == 0 || response["status"] != "STOPPED" {
		t.Fatalf("M11 activation did not report durable STOP: code=%d response=%+v", code, response)
	}
}

func TestMissionM11DurableStopPreventsLifecycleWrites(t *testing.T) {
	dir := t.TempDir()
	if code, response := missionCall(t, "init", dir); code != 0 || response["status"] != "INITIALIZED" {
		t.Fatalf("init failed: code=%d response=%+v", code, response)
	}
	if code, response := missionCall(t, "m11-stop", dir, "gate-stop-drill"); code != 0 || response["status"] != "STOPPED" {
		t.Fatalf("stop failed: code=%d response=%+v", code, response)
	}
	outcomeInput := filepath.Join(dir, "outcome.json")
	if err := os.WriteFile(outcomeInput, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	oldRuntime := filepath.Join(dir, "old-runtime")
	handoffInput := filepath.Join(dir, "handoff.json")
	admissionInput := filepath.Join(dir, "admission.json")
	for _, args := range [][]string{
		{"m11-activate", "missing-lease", "2026-09-08T00:00:00Z"},
		{"m11-ledger-init", "missing-lease", "2026-09-08T00:00:00Z"},
		{"m11-gate", "missing-lease", "missing-health", "missing-cost", "missing-ledger", "2026-09-08T00:00:00Z"},
		{"m11-authorize", "missing-lease", "missing-gate", "fixture_stub", "2026-09-08T00:00:00Z"},
		{"m11-reserve-authorization", "missing-authorization", "missing-ledger", "2026-09-08T00:00:00Z"},
		{"m11-record-failed", "missing-authorization", "missing-ledger", "2026-09-08T00:00:00Z", "fixture failure"},
		{"m11-outcome", outcomeInput, "missing-ledger"},
		{"m11-evaluate", "missing-outcome", "missing-evaluation", "2026-09-08T00:00:00Z"},
		{"m11-close-cycle", "missing-cycle", "missing-evaluation", "2026-09-08T00:00:00Z"},
	} {
		arguments := append([]string{args[0], dir}, args[1:]...)
		if code, response := missionCall(t, arguments...); code == 0 || response["status"] != "STOPPED" {
			t.Fatalf("stopped %s did not fail closed: code=%d response=%+v", args[0], code, response)
		}
	}
	if code, response := missionCall(t, "m11-recovery-admit", dir, oldRuntime, handoffInput, admissionInput); code == 0 || response["status"] != "STOPPED" {
		t.Fatalf("stopped m11-recovery-admit did not fail closed before reading inputs: code=%d response=%+v", code, response)
	}
	if _, err := os.Stat(m11ArtifactRegistryPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("stopped lifecycle command wrote an M11 registry artifact: %v", err)
	}
	if _, err := os.Stat(m11OutcomeStorePath(dir)); !os.IsNotExist(err) {
		t.Fatalf("stopped lifecycle command wrote an M11 outcome artifact: %v", err)
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
