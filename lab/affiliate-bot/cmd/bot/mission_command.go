package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
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
type LearnerReservation struct {
	ReservationID string `json:"reservation_id"`
	GrantID       string `json:"grant_id"`
	IntentID      string `json:"intent_id"`
	IntentHash    string `json:"intent_hash"`
	CostMinor     int64  `json:"cost_minor"`
	CostBoundID   string `json:"cost_bound_id,omitempty"`
	CostBoundHash string `json:"cost_bound_hash,omitempty"`
	ReservedAt    string `json:"reserved_at"`
}

func trustedCostBoundsPath(dir string) string { return filepath.Join(dir, "trusted-cost-bounds.jsonl") }

func loadTrustedCostBounds(dir string) ([]corem10.TrustedCostBound, error) {
	raw, err := os.ReadFile(trustedCostBoundsPath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var bounds []corem10.TrustedCostBound
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		bound, status := corem10.DecodeTrustedCostBound(line)
		if status != "VALID" {
			return nil, fmt.Errorf("invalid trusted cost-bound registry entry: %s", status)
		}
		bounds = append(bounds, bound)
	}
	return bounds, nil
}

func resolveTrustedCostBound(dir string, bound corem10.TrustedCostBound) bool {
	if bound.CostBoundHash == "" || bound.CostBoundHash != corem10.ComputeTrustedCostBoundHash(bound) {
		return false
	}
	bounds, err := loadTrustedCostBounds(dir)
	if err != nil {
		return false
	}
	for _, registered := range bounds {
		if registered.CostBoundID == bound.CostBoundID && registered.CostBoundHash == bound.CostBoundHash {
			return true
		}
	}
	return false
}

type LearnerMissionState struct {
	Version      string               `json:"version"`
	Intent       *LearnerIntent       `json:"intent,omitempty"`
	Policy       *LearnerPolicy       `json:"policy,omitempty"`
	Approval     *LearnerApproval     `json:"approval,omitempty"`
	Canary       *LearnerCanary       `json:"canary,omitempty"`
	Reservations []LearnerReservation `json:"reservations,omitempty"`
	Lease        *LearnerLease        `json:"lease,omitempty"`
	Stop         bool                 `json:"stop"`
	StopReason   string               `json:"stop_reason,omitempty"`
	UpdatedAt    string               `json:"updated_at"`
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
	return corem08.ComputeIntentHash(corem08.Intent(i))
}

func writeJSON(path string, value any) error {
	b, err := marshalJSON(value)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func marshalJSON(value any) ([]byte, error) {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// writeNewJSON creates an immutable command artifact. A retry is successful
// only when it finds the byte-for-byte artifact it would have created; a
// different existing file, symlink, or special file is never overwritten.
func writeNewJSON(path string, value any) (string, error) {
	b, err := marshalJSON(value)
	if err != nil {
		return "", err
	}
	if info, err := os.Lstat(path); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return "", fmt.Errorf("artifact output must be a new regular file")
		}
		existing, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if bytes.Equal(existing, b) {
			return appendDuplicate, nil
		}
		return "", fmt.Errorf("artifact output already exists with different content")
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return appendAdded, nil
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
	seenReservations := map[string]bool{}
	for _, r := range s.Reservations {
		if r.ReservationID == "" || seenReservations[r.ReservationID] || s.Canary == nil || r.GrantID != s.Canary.GrantID || r.IntentID != s.Canary.IntentID || r.IntentHash != s.Canary.IntentHash || r.CostMinor < 0 {
			return fmt.Errorf("mission reservation integrity/binding check failed")
		}
		if _, err := time.Parse(time.RFC3339, r.ReservedAt); err != nil {
			return fmt.Errorf("mission reservation timestamp is invalid")
		}
		seenReservations[r.ReservationID] = true
	}
	return nil
}
func saveMissionState(dir string, s LearnerMissionState) error {
	s.Version = missionStateVersion
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return writeJSONAtomic(missionStatePath(dir), s)
}

// Mission state is mutable, unlike M08 artifacts. Commit it by atomic rename
// while the mission directory lock is held; never truncate the prior state.
func writeJSONAtomic(path string, value any) error {
	b, err := marshalJSON(value)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".mission-state-")
	if err != nil {
		return err
	}
	temporary := f.Name()
	defer os.Remove(temporary)
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = d.Sync()
	closeErr := d.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func missionAuthorityActive(s LearnerMissionState, now time.Time) error {
	if s.Intent == nil || s.Policy == nil || s.Approval == nil {
		return fmt.Errorf("intent, policy and approval are required")
	}
	intentExpiry, intentErr := time.Parse(time.RFC3339, s.Intent.ExpiresAt)
	approvalExpiry, approvalErr := time.Parse(time.RFC3339, s.Approval.ExpiresAt)
	policyChecked, policyErr := time.Parse(time.RFC3339, s.Policy.PolicyCheckedAt)
	if intentErr != nil || approvalErr != nil || policyErr != nil || !intentExpiry.After(now) || !approvalExpiry.After(now) || policyChecked.After(now) {
		return fmt.Errorf("intent, approval or policy time binding is invalid or expired")
	}
	if s.Approval.Decision != "APPROVE" || !s.Approval.OneTime || s.Approval.IntentID != s.Intent.IntentID || s.Approval.IntentHash != s.Intent.IntentHash || s.Approval.PolicyVersion != s.Policy.PolicyVersion || s.Approval.CorrelationID != s.Intent.CorrelationID {
		return fmt.Errorf("approval does not bind to current intent/policy")
	}
	return nil
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
	raw, err := os.ReadFile(requestPath)
	if err != nil {
		return LearnerIntent{}, err
	}
	if err := contracts.DecodeStrict(raw, &req); err != nil {
		return LearnerIntent{}, err
	}
	decoded, err := contracts.Decode(raw)
	if err != nil {
		return LearnerIntent{}, err
	}
	object, ok := decoded.(map[string]any)
	if !ok {
		return LearnerIntent{}, fmt.Errorf("intent request must be an object")
	}
	parameters, ok := object["parameters"].(map[string]any)
	if !ok {
		return LearnerIntent{}, fmt.Errorf("parameters must be a JSON object")
	}
	req.Parameters = parameters
	record, err := resolveCanonicalRecord(historyPath, req.DecisionID)
	if err != nil {
		return LearnerIntent{}, fmt.Errorf("decision_id resolution: %w", err)
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
	if req.ProposedBy == "agent" {
		return LearnerIntent{}, fmt.Errorf("agent proposal_ref must resolve in the canonical proposal store")
	}
	created, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		return LearnerIntent{}, err
	}
	expires, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil || !expires.After(created) {
		return LearnerIntent{}, fmt.Errorf("expires_at must be after created_at")
	}
	sealed := corem08.SealIntent(corem08.Intent{IntentID: req.IntentID, DecisionID: req.DecisionID, EvidenceIDs: req.EvidenceIDs, ActionType: req.ActionType, Target: req.Target, Parameters: req.Parameters, ProposedBy: req.ProposedBy, ProposalRef: req.ProposalRef, CreatedAt: req.CreatedAt, ExpiresAt: req.ExpiresAt, CorrelationID: req.CorrelationID, IdempotencyKey: req.IdempotencyKey})
	return LearnerIntent(sealed), nil
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
	raw, err := os.ReadFile(path)
	if err != nil {
		return LearnerPolicy{}, err
	}
	if err := contracts.DecodeStrict(raw, &req); err != nil {
		return LearnerPolicy{}, err
	}
	ctx := corem08.PolicyContext{PolicyVersion: req.PolicyVersion, Now: req.Now, KnownDecisionIDs: []string{i.DecisionID}, KnownEvidenceIDs: append([]string(nil), i.EvidenceIDs...), AllowedHosts: req.AllowedHosts, ActionRisk: req.ActionRisk, SeenIdempotency: req.SeenIdempotency}
	return LearnerPolicy(corem08.EvaluatePolicy(corem08.Intent(i), ctx)), nil
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
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT | m08-policy INTENT POLICY OUT | bind STATE_DIR INTENT POLICY | m09-approval STATE_DIR APPROVAL | m10-canary STATE_DIR GRANT | m10-cost-register STATE_DIR COST_BOUND | m10-reserve STATE_DIR COST_MINOR|COST_BOUND [RESERVATION_ID] | m11-stop STATE_DIR REASON | status STATE_DIR"), 2)
	}
	// Directory creation and an exclusive lock make the mutable mission state
	// single-writer across processes. A stale lock fails closed and requires an
	// explicit recovery procedure rather than silently risking double reserve.
	mutatesState := map[string]bool{"bind": true, "m09-approval": true, "approval": true, "m10-canary": true, "canary": true, "m10-cost-register": true, "m10-reserve": true, "reserve": true, "m11-stop": true, "stop": true, "init": true}[args[0]]
	if mutatesState && len(args) >= 2 {
		if err := os.MkdirAll(args[1], 0700); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		lockPath := filepath.Join(args[1], ".mission.lock")
		if err := os.Mkdir(lockPath, 0700); err != nil {
			if os.IsExist(err) {
				return emit("BUSY", nil, fmt.Errorf("mission state is locked; explicit recovery required after an interrupted writer"), 1)
			}
			return emit("STORE_ERROR", nil, err, 1)
		}
		defer os.Remove(lockPath)
	}
	switch args[0] {
	case "m08-intent", "intent":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT"), 2)
		}
		if err := distinctPaths(args[1:]...); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		i, err := buildLearnerIntent(args[1], args[2])
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		status, err := writeNewJSON(args[3], i)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		return emit(status, i, nil, 0)
	case "m08-policy", "policy":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-policy INTENT POLICY OUT"), 2)
		}
		if err := distinctPaths(args[1:]...); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		raw, err := os.ReadFile(args[1])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		decoded, state := corem08.DecodeIntent(raw)
		if state != "VALID" {
			return emit("INPUT_ERROR", nil, fmt.Errorf("invalid canonical M08 intent"), 1)
		}
		i := LearnerIntent(decoded)
		p, err := evaluateLearnerPolicy(i, args[2])
		if err != nil {
			return emit("DENY", nil, err, 1)
		}
		status, err := writeNewJSON(args[3], p)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		if status == appendDuplicate {
			return emit(status, p, nil, 0)
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
			s.Reservations = nil
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
		if intentExpiry, intentErr := time.Parse(time.RFC3339, s.Intent.ExpiresAt); intentErr != nil || !intentExpiry.After(time.Now().UTC()) {
			return emit("REJECTED", nil, fmt.Errorf("intent is expired"), 1)
		}
		if s.Approval != nil {
			if *s.Approval == a {
				return emit("EXACT_DUPLICATE", a, nil, 0)
			}
			return emit("REJECTED", nil, fmt.Errorf("cannot replace an existing approval"), 1)
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
		if err := missionAuthorityActive(s, time.Now().UTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
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
			if s.Canary.MaxExecutions != c.MaxExecutions || s.Canary.MaxCostMinor != c.MaxCostMinor || s.Canary.Currency != c.Currency {
				return emit("REJECTED", nil, fmt.Errorf("canary grant limits and currency are immutable"), 1)
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
	case "m10-cost-register":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-cost-register STATE_DIR COST_BOUND"), 2)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop || s.Intent == nil || s.Canary == nil {
			return emit("REJECTED", nil, fmt.Errorf("active intent and canary grant required"), 1)
		}
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		bound, status := corem10.DecodeTrustedCostBound(raw)
		if status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
		}
		if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, time.Now().UTC()); status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
		}
		bounds, err := loadTrustedCostBounds(args[1])
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		for _, old := range bounds {
			if old.CostBoundID == bound.CostBoundID {
				if old.CostBoundHash == bound.CostBoundHash {
					return emit("EXACT_DUPLICATE", bound, nil, 0)
				}
				return emit("CONFLICT", nil, fmt.Errorf("cost_bound_id reused with different content"), 1)
			}
		}
		f, err := os.OpenFile(trustedCostBoundsPath(args[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		_, err = f.Write(append(bytes.TrimSpace(raw), '\n'))
		if syncErr := f.Sync(); err == nil {
			err = syncErr
		}
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("APPENDED", bound, nil, 0)
	case "m10-reserve", "reserve":
		if len(args) != 3 && len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-reserve STATE_DIR COST_MINOR [RESERVATION_ID]"), 2)
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
		if err := missionAuthorityActive(s, time.Now().UTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		cost, parseErr := strconv.ParseInt(args[2], 10, 64)
		boundID, boundHash := "", ""
		if parseErr != nil {
			raw, err := os.ReadFile(args[2])
			if err != nil {
				return emit("INPUT_ERROR", nil, err, 1)
			}
			bound, status := corem10.DecodeTrustedCostBound(raw)
			if status != "VALID" {
				return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
			}
			if !resolveTrustedCostBound(args[1], bound) {
				return emit("REJECTED", nil, fmt.Errorf("cost bound is not registered"), 1)
			}
			if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, time.Now().UTC()); status != "VALID" {
				return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
			}
			cost, boundID, boundHash = bound.MaxCostMinor, bound.CostBoundID, bound.CostBoundHash
		} else if cost < 0 {
			return emit("REJECTED", nil, fmt.Errorf("cost must be a non-negative integer"), 1)
		}
		reservationID := "legacy:" + args[2]
		if len(args) == 4 {
			reservationID = args[3]
		}
		if reservationID == "" {
			return emit("REJECTED", nil, fmt.Errorf("reservation_id is required"), 1)
		}
		for _, prior := range s.Reservations {
			if prior.ReservationID != reservationID {
				continue
			}
			if prior.GrantID == s.Canary.GrantID && prior.IntentID == s.Intent.IntentID && prior.IntentHash == s.Intent.IntentHash && prior.CostMinor == cost {
				return emit("EXACT_DUPLICATE", s.Canary, nil, 0)
			}
			return emit("REJECTED", nil, fmt.Errorf("reservation_id reused with different binding or cost"), 1)
		}
		if s.Canary.Status != "ACTIVE" || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed >= s.Canary.MaxExecutions || cost > s.Canary.MaxCostMinor-s.Canary.CostUsedMinor {
			return emit("BUDGET_DENIED", s.Canary, fmt.Errorf("canary budget exhausted"), 1)
		}
		s.Canary.ExecutionsUsed++
		s.Canary.CostUsedMinor += cost
		s.Reservations = append(s.Reservations, LearnerReservation{ReservationID: reservationID, GrantID: s.Canary.GrantID, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, CostMinor: cost, CostBoundID: boundID, CostBoundHash: boundHash, ReservedAt: time.Now().UTC().Format(time.RFC3339Nano)})
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
		if err = writeJSONAtomic(filepath.Join(args[1], "STOP"), map[string]any{"active": true, "reason": args[2]}); err != nil {
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
