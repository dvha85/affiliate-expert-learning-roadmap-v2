package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const missionStateVersion = "learner-m08-m11/v1"

type LearnerIntent struct {
	IntentID            string         `json:"intent_id"`
	DecisionID          string         `json:"decision_id"`
	EvidenceIDs         []string       `json:"evidence_ids"`
	ActionType          string         `json:"action_type"`
	Target              string         `json:"target"`
	Parameters          map[string]any `json:"parameters"`
	ProposedBy          string         `json:"proposed_by"`
	ProposalRef         string         `json:"proposal_ref,omitempty"`
	CreatedAt           string         `json:"created_at"`
	ExpiresAt           string         `json:"expires_at"`
	CorrelationID       string         `json:"correlation_id"`
	IdempotencyKey      string         `json:"idempotency_key"`
	IntentHash          string         `json:"intent_hash"`
	IntentMode          string         `json:"intent_mode"`
	ExecutionAuthorized bool           `json:"execution_authorized"`
}
type LearnerPolicy struct {
	PolicyVersion        string `json:"policy_version"`
	IntentID             string `json:"intent_id"`
	IntentHash           string `json:"intent_hash"`
	Decision             string `json:"decision"`
	RiskClass            string `json:"risk_class"`
	Reason               string `json:"reason"`
	PolicyReviewRequired bool   `json:"policy_review_required"`
	PolicyMode           string `json:"policy_mode"`
	ExecutionAuthorized  bool   `json:"execution_authorized"`
	PolicyCheckedAt      string `json:"policy_checked_at"`
}
type LearnerApproval struct {
	ApprovalID    string `json:"approval_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	PolicyVersion string `json:"policy_version"`
	Decision      string `json:"decision"`
	ApprovedBy    string `json:"approved_by"`
	ApproverID    string `json:"approver_id"`
	ApprovedAt    string `json:"approved_at"`
	ExpiresAt     string `json:"expires_at"`
	CorrelationID string `json:"correlation_id"`
	OneTime       bool   `json:"one_time"`
}
type LearnerCanary struct {
	GrantID        string `json:"grant_id"`
	IntentID       string `json:"intent_id"`
	IntentHash     string `json:"intent_hash"`
	ApprovalID     string `json:"approval_id"`
	MaxExecutions  int    `json:"max_executions"`
	MaxCostMinor   int64  `json:"max_cost_minor"`
	Currency       string `json:"currency"`
	Status         string `json:"status"`
	GrantedAt      string `json:"granted_at"`
	ExecutionsUsed int    `json:"executions_used"`
	CostUsedMinor  int64  `json:"cost_used_minor"`
}
type LearnerLease struct {
	LeaseID       string `json:"lease_id"`
	CanaryGrantID string `json:"canary_grant_id"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
}
type LearnerMissionState struct {
	Version    string           `json:"version"`
	Intent     *LearnerIntent   `json:"intent,omitempty"`
	Policy     *LearnerPolicy   `json:"policy,omitempty"`
	Approval   *LearnerApproval `json:"approval,omitempty"`
	Canary     *LearnerCanary   `json:"canary,omitempty"`
	Lease      *LearnerLease    `json:"lease,omitempty"`
	Stop       bool             `json:"stop"`
	StopReason string           `json:"stop_reason,omitempty"`
	UpdatedAt  string           `json:"updated_at"`
}

type intentHashPayload struct {
	IntentID            string         `json:"intent_id"`
	DecisionID          string         `json:"decision_id"`
	EvidenceIDs         []string       `json:"evidence_ids"`
	ActionType          string         `json:"action_type"`
	Target              string         `json:"target"`
	Parameters          map[string]any `json:"parameters"`
	ProposedBy          string         `json:"proposed_by"`
	ProposalRef         string         `json:"proposal_ref,omitempty"`
	CreatedAt           string         `json:"created_at"`
	ExpiresAt           string         `json:"expires_at"`
	CorrelationID       string         `json:"correlation_id"`
	IdempotencyKey      string         `json:"idempotency_key"`
	IntentMode          string         `json:"intent_mode"`
	ExecutionAuthorized bool           `json:"execution_authorized"`
}

func learnerIntentHash(i LearnerIntent) string {
	ids := append([]string(nil), i.EvidenceIDs...)
	// Evidence IDs are already canonical in HistoryRecord; preserve order to
	// make tampering visible instead of silently normalizing a submitted intent.
	b, _ := json.Marshal(intentHashPayload{i.IntentID, i.DecisionID, ids, i.ActionType, i.Target, i.Parameters, i.ProposedBy, i.ProposalRef, i.CreatedAt, i.ExpiresAt, i.CorrelationID, i.IdempotencyKey, i.IntentMode, i.ExecutionAuthorized})
	s := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(s[:])
}

func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
func readJSON(path string, value any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, value)
}
func missionStatePath(dir string) string { return filepath.Join(dir, "mission-state.json") }
func loadMissionState(dir string) (LearnerMissionState, error) {
	var s LearnerMissionState
	if err := readJSON(missionStatePath(dir), &s); err != nil {
		return s, err
	}
	if s.Version != missionStateVersion {
		return s, fmt.Errorf("unsupported mission state")
	}
	if err := validateMissionState(dir, s); err != nil {
		return s, err
	}
	return s, nil
}

func stopMarkerActive(dir string) (bool, error) {
	var marker struct {
		Active bool `json:"active"`
	}
	if err := readJSON(filepath.Join(dir, "STOP"), &marker); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return marker.Active, nil
}

func validateMissionState(dir string, s LearnerMissionState) error {
	marker, err := stopMarkerActive(dir)
	if err != nil {
		return err
	}
	if marker && !s.Stop {
		return fmt.Errorf("STOP marker is active but mission state is not stopped")
	}
	if s.Intent != nil {
		if s.Intent.IntentHash == "" || s.Intent.IntentHash != learnerIntentHash(*s.Intent) || s.Intent.ExecutionAuthorized || s.Intent.IntentMode != "PROPOSAL_ONLY" {
			return fmt.Errorf("mission intent integrity/authority check failed")
		}
	}
	if s.Policy != nil && (s.Intent == nil || s.Policy.IntentID != s.Intent.IntentID || s.Policy.IntentHash != s.Intent.IntentHash || s.Policy.ExecutionAuthorized) {
		return fmt.Errorf("mission policy binding check failed")
	}
	if s.Approval != nil {
		if s.Intent == nil || s.Policy == nil || s.Approval.ApprovalID == "" || s.Approval.ApprovedBy != "human" || s.Approval.ApproverID == "" || s.Approval.IntentID != s.Intent.IntentID || s.Approval.IntentHash != s.Intent.IntentHash || s.Approval.PolicyVersion != s.Policy.PolicyVersion || s.Approval.CorrelationID != s.Intent.CorrelationID || !s.Approval.OneTime {
			return fmt.Errorf("mission approval integrity/binding check failed")
		}
	}
	if s.Canary != nil {
		if s.Intent == nil || s.Approval == nil || s.Canary.GrantID == "" || s.Canary.IntentID != s.Intent.IntentID || s.Canary.IntentHash != s.Intent.IntentHash || s.Canary.ApprovalID != s.Approval.ApprovalID || s.Canary.MaxExecutions <= 0 || s.Canary.MaxCostMinor < 0 || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed > s.Canary.MaxExecutions || s.Canary.CostUsedMinor > s.Canary.MaxCostMinor {
			return fmt.Errorf("mission canary integrity/binding/budget check failed")
		}
	}
	return nil
}
func saveMissionState(dir string, s LearnerMissionState) error {
	s.Version = missionStateVersion
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return writeJSON(missionStatePath(dir), s)
}

type learnerIntentRequest struct {
	IntentID       string         `json:"intent_id"`
	DecisionID     string         `json:"decision_id"`
	EvidenceIDs    []string       `json:"evidence_ids"`
	ActionType     string         `json:"action_type"`
	Target         string         `json:"target"`
	Parameters     map[string]any `json:"parameters"`
	ProposedBy     string         `json:"proposed_by"`
	ProposalRef    string         `json:"proposal_ref,omitempty"`
	CreatedAt      string         `json:"created_at"`
	ExpiresAt      string         `json:"expires_at"`
	CorrelationID  string         `json:"correlation_id"`
	IdempotencyKey string         `json:"idempotency_key"`
}

func buildLearnerIntent(historyPath, requestPath string) (LearnerIntent, error) {
	var req learnerIntentRequest
	if err := readJSON(requestPath, &req); err != nil {
		return LearnerIntent{}, err
	}
	history, err := LoadHistory(historyPath)
	if err != nil {
		return LearnerIntent{}, err
	}
	var record *HistoryRecord
	for i := range history {
		if history[i].RecordID == req.DecisionID {
			copy := history[i]
			record = &copy
		}
	}
	if record == nil {
		return LearnerIntent{}, fmt.Errorf("decision_id does not resolve in canonical history")
	}
	allowed := map[string]bool{}
	for _, id := range record.RecordedResult.EvidenceIDs {
		allowed[id] = true
	}
	if len(req.EvidenceIDs) == 0 {
		req.EvidenceIDs = append([]string(nil), record.RecordedResult.EvidenceIDs...)
	}
	for _, id := range req.EvidenceIDs {
		if !allowed[id] {
			return LearnerIntent{}, fmt.Errorf("evidence_id %s is not linked to decision", id)
		}
	}
	if req.IntentID == "" || req.ActionType == "" || req.Target == "" || (req.ProposedBy != "human" && req.ProposedBy != "agent") || req.CorrelationID == "" || req.IdempotencyKey == "" || req.CreatedAt == "" || req.ExpiresAt == "" {
		return LearnerIntent{}, fmt.Errorf("intent request missing required field")
	}
	created, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		return LearnerIntent{}, err
	}
	expires, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil || !expires.After(created) {
		return LearnerIntent{}, fmt.Errorf("expires_at must be after created_at")
	}
	i := LearnerIntent{IntentID: req.IntentID, DecisionID: req.DecisionID, EvidenceIDs: req.EvidenceIDs, ActionType: strings.ToUpper(req.ActionType), Target: req.Target, Parameters: req.Parameters, ProposedBy: req.ProposedBy, ProposalRef: req.ProposalRef, CreatedAt: req.CreatedAt, ExpiresAt: req.ExpiresAt, CorrelationID: req.CorrelationID, IdempotencyKey: req.IdempotencyKey, IntentMode: "PROPOSAL_ONLY", ExecutionAuthorized: false}
	i.IntentHash = learnerIntentHash(i)
	return i, nil
}

type learnerPolicyRequest struct {
	PolicyVersion   string            `json:"policy_version"`
	Now             string            `json:"now"`
	AllowedHosts    []string          `json:"allowed_hosts"`
	ActionRisk      map[string]string `json:"action_risk"`
	SeenIdempotency map[string]string `json:"seen_idempotency"`
}

func evaluateLearnerPolicy(i LearnerIntent, path string) (LearnerPolicy, error) {
	var req learnerPolicyRequest
	if err := readJSON(path, &req); err != nil {
		return LearnerPolicy{}, err
	}
	if req.PolicyVersion == "" {
		return LearnerPolicy{}, fmt.Errorf("policy_version required")
	}
	if i.IntentHash != learnerIntentHash(i) {
		return LearnerPolicy{}, fmt.Errorf("intent hash mismatch")
	}
	now, err := time.Parse(time.RFC3339, req.Now)
	if err != nil {
		return LearnerPolicy{}, err
	}
	expires, _ := time.Parse(time.RFC3339, i.ExpiresAt)
	if expires.IsZero() {
		return LearnerPolicy{}, fmt.Errorf("intent expires_at is invalid")
	}
	decision := "DENY"
	reason := "UNKNOWN_ACTION_POLICY"
	risk := req.ActionRisk[i.ActionType]
	if risk != "RISK0" && risk != "RISK1" && risk != "RISK2" {
		return LearnerPolicy{}, fmt.Errorf("unknown action risk %q", risk)
	}
	if risk == "RISK0" {
		decision = "ALLOW"
		reason = "SHADOW_POLICY_ALLOW"
	} else {
		decision = "HUMAN_REVIEW"
		reason = risk + "_REQUIRES_REVIEW"
	}
	target, parseErr := url.Parse(i.Target)
	hostAllowed := parseErr == nil && target.Scheme == "https" && target.Hostname() != "" && target.Port() == "" && target.User == nil && target.RawQuery == "" && target.Fragment == ""
	if hostAllowed {
		hostAllowed = false
		for _, host := range req.AllowedHosts {
			if strings.EqualFold(strings.TrimSpace(host), target.Hostname()) {
				hostAllowed = true
				break
			}
		}
	}
	if !hostAllowed {
		decision = "DENY"
		reason = "TARGET_HOST_DENIED"
	}
	if _, seen := req.SeenIdempotency[i.IdempotencyKey]; seen {
		decision = "DENY"
		reason = "IDEMPOTENCY_REPLAY"
	}
	if !expires.After(now) {
		decision = "DENY"
		reason = "INTENT_EXPIRED"
	}
	p := LearnerPolicy{PolicyVersion: req.PolicyVersion, IntentID: i.IntentID, IntentHash: i.IntentHash, Decision: decision, RiskClass: risk, Reason: reason, PolicyReviewRequired: decision == "HUMAN_REVIEW", PolicyMode: "NON_AUTHORIZING", ExecutionAuthorized: false, PolicyCheckedAt: req.Now}
	return p, nil
}

func runMissionCommand(args []string, stdout, stderr io.Writer) int {
	emit := func(status string, artifact any, err error, code int) int {
		if err != nil {
			fmt.Fprintln(stderr, err)
		}
		o := map[string]any{"command": "mission", "status": status, "execution_permitted": false}
		if artifact != nil {
			o["artifact"] = artifact
		}
		if e := json.NewEncoder(stdout).Encode(o); e != nil {
			return 1
		}
		return code
	}
	if len(args) < 1 {
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT | m08-policy INTENT POLICY OUT | bind STATE_DIR INTENT POLICY | m09-approval STATE_DIR APPROVAL | m10-canary STATE_DIR GRANT | m10-reserve STATE_DIR COST_MINOR | m11-stop STATE_DIR REASON | status STATE_DIR"), 2)
	}
	switch args[0] {
	case "m08-intent", "intent":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT"), 2)
		}
		i, err := buildLearnerIntent(args[1], args[2])
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		if err = writeJSON(args[3], i); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("APPENDED", i, nil, 0)
	case "m08-policy", "policy":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-policy INTENT POLICY OUT"), 2)
		}
		var i LearnerIntent
		if err := readJSON(args[1], &i); err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		p, err := evaluateLearnerPolicy(i, args[2])
		if err != nil {
			return emit("DENY", nil, err, 1)
		}
		if err = writeJSON(args[3], p); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit(p.Decision, p, nil, 0)
	case "bind":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission bind STATE_DIR INTENT POLICY"), 2)
		}
		var i LearnerIntent
		var p LearnerPolicy
		if err := readJSON(args[2], &i); err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		if err := readJSON(args[3], &p); err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		if i.IntentHash != learnerIntentHash(i) || p.IntentID != i.IntentID || p.IntentHash != i.IntentHash {
			return emit("REJECTED", nil, fmt.Errorf("intent/policy link or hash invalid"), 1)
		}
		if err := os.MkdirAll(args[1], 0700); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if os.IsNotExist(err) {
			s = LearnerMissionState{Version: missionStateVersion}
		} else if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop {
			return emit("STOPPED", s, fmt.Errorf("durable STOP: %s", s.StopReason), 1)
		}
		if s.Intent == nil || s.Intent.IntentID != i.IntentID || s.Intent.IntentHash != i.IntentHash {
			s.Approval = nil
			s.Canary = nil
			s.Lease = nil
		}
		s.Intent = &i
		s.Policy = &p
		if err = saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("BOUND", s, nil, 0)
	case "m09-approval", "approval":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m09-approval STATE_DIR APPROVAL"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		var a LearnerApproval
		if err = readJSON(args[2], &a); err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		if s.Intent == nil || s.Policy == nil || (s.Policy.Decision != "ALLOW" && s.Policy.Decision != "HUMAN_REVIEW") || a.ApprovalID == "" || a.ApprovedBy != "human" || a.ApproverID == "" || a.IntentID != s.Intent.IntentID || a.IntentHash != s.Intent.IntentHash || a.PolicyVersion != s.Policy.PolicyVersion || a.CorrelationID != s.Intent.CorrelationID || a.Decision != "APPROVE" || !a.OneTime {
			return emit("REJECTED", nil, fmt.Errorf("approval does not bind to current intent/policy"), 1)
		}
		approvedAt, approvedErr := time.Parse(time.RFC3339, a.ApprovedAt)
		expiresAt, expiresErr := time.Parse(time.RFC3339, a.ExpiresAt)
		if approvedErr != nil || expiresErr != nil || !expiresAt.After(approvedAt) || !expiresAt.After(time.Now().UTC()) || approvedAt.After(time.Now().UTC()) {
			return emit("REJECTED", nil, fmt.Errorf("approval timestamps are invalid or expired"), 1)
		}
		s.Approval = &a
		if err = saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("ACK", a, nil, 0)
	case "m10-canary", "canary":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-canary STATE_DIR GRANT"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop {
			return emit("STOPPED", s, fmt.Errorf("durable STOP: %s", s.StopReason), 1)
		}
		if s.Intent == nil || s.Policy == nil || s.Approval == nil || s.Approval.Decision != "APPROVE" {
			return emit("REJECTED", nil, fmt.Errorf("approved human record required"), 1)
		}
		var c LearnerCanary
		if err = readJSON(args[2], &c); err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		if c.GrantID == "" || c.MaxExecutions <= 0 || c.MaxCostMinor < 0 || c.Currency == "" || c.ExecutionsUsed < 0 || c.CostUsedMinor < 0 {
			return emit("REJECTED", nil, fmt.Errorf("bounded canary fields required"), 1)
		}
		if c.ExecutionsUsed > c.MaxExecutions || c.CostUsedMinor > c.MaxCostMinor {
			return emit("REJECTED", nil, fmt.Errorf("canary usage exceeds grant"), 1)
		}
		if s.Canary != nil {
			if s.Canary.GrantID != c.GrantID || s.Canary.IntentID != s.Intent.IntentID || s.Canary.ApprovalID != s.Approval.ApprovalID {
				return emit("REJECTED", nil, fmt.Errorf("cannot replace an existing canary binding"), 1)
			}
			c.ExecutionsUsed = s.Canary.ExecutionsUsed
			c.CostUsedMinor = s.Canary.CostUsedMinor
			c.GrantedAt = s.Canary.GrantedAt
		}
		c.IntentID = s.Intent.IntentID
		c.IntentHash = s.Intent.IntentHash
		c.ApprovalID = s.Approval.ApprovalID
		c.Status = "ACTIVE"
			if c.GrantedAt == "" {
				c.GrantedAt = time.Now().UTC().Format(time.RFC3339Nano)
			}
		s.Canary = &c
		if err = saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("ACK", c, nil, 0)
	case "m10-reserve", "reserve":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-reserve STATE_DIR COST_MINOR"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop {
			return emit("STOPPED", s, fmt.Errorf("durable STOP: %s", s.StopReason), 1)
		}
		if s.Canary == nil {
			return emit("REJECTED", nil, fmt.Errorf("canary grant required"), 1)
		}
		cost, parseErr := strconv.ParseInt(args[2], 10, 64)
		if parseErr != nil || cost < 0 {
			return emit("REJECTED", nil, fmt.Errorf("cost must be a non-negative integer"), 1)
		}
		if s.Canary.Status != "ACTIVE" || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed >= s.Canary.MaxExecutions || cost > s.Canary.MaxCostMinor-s.Canary.CostUsedMinor {
			return emit("BUDGET_DENIED", s.Canary, fmt.Errorf("canary budget exhausted"), 1)
		}
		s.Canary.ExecutionsUsed++
		s.Canary.CostUsedMinor += cost
		if err := saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("RESERVED", s.Canary, nil, 0)
	case "m11-stop", "stop":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-stop STATE_DIR REASON"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		s.Stop = true
		s.StopReason = args[2]
		if err = saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if err = writeJSON(filepath.Join(args[1], "STOP"), map[string]any{"active": true, "reason": args[2]}); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("STOPPED", s, nil, 0)
	case "status", "m11-status":
		if len(args) != 2 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission status STATE_DIR"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		return emit("VALID", s, nil, 0)
	case "init":
		if len(args) != 2 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission init STATE_DIR"), 2)
		}
		if err := os.MkdirAll(args[1], 0700); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, err := os.Stat(missionStatePath(args[1])); err == nil {
			return emit("ALREADY_INITIALIZED", nil, fmt.Errorf("refusing to overwrite existing mission state"), 1)
		} else if !os.IsNotExist(err) {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if active, err := stopMarkerActive(args[1]); err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		} else if active {
			return emit("STOPPED", nil, fmt.Errorf("active STOP marker requires explicit recovery"), 1)
		}
		s := LearnerMissionState{Version: missionStateVersion}
		if err := saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("INITIALIZED", s, nil, 0)
	default:
		return emit("USAGE_ERROR", nil, fmt.Errorf("unknown mission operation %q", args[0]), 2)
	}
}
