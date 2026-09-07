package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateLearnerPolicyRequiresReviewForRiskTwo(t *testing.T) {
	i := LearnerIntent{IntentID: "i", DecisionID: "d", EvidenceIDs: []string{"e"}, ActionType: "PUBLISH", Target: "https://example.com/publish", ProposedBy: "human", CreatedAt: "2099-01-01T00:00:00Z", ExpiresAt: "2099-01-01T02:00:00Z", CorrelationID: "c", IdempotencyKey: "k", IntentMode: "PROPOSAL_ONLY"}
	i.IntentHash = learnerIntentHash(i)
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.json")
	if err := os.WriteFile(path, []byte(`{"policy_version":"v1","now":"2099-01-01T01:00:00Z","allowed_hosts":["example.com"],"action_risk":{"PUBLISH":"RISK2"},"seen_idempotency":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := evaluateLearnerPolicy(i, path)
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
