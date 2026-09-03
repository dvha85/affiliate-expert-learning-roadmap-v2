package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ApprovalRecord struct {
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

type ExecutorProfile struct {
	ExecutorID         string   `json:"executor_id"`
	AllowedActionTypes []string `json:"allowed_action_types"`
	AllowedHosts       []string `json:"allowed_hosts"`
}

type ExecutionAuthorization struct {
	AuthorizationID     string `json:"authorization_id"`
	IntentID            string `json:"intent_id"`
	IntentHash          string `json:"intent_hash"`
	PolicyVersion       string `json:"policy_version"`
	ApprovalID          string `json:"approval_id"`
	ExecutorID          string `json:"executor_id"`
	AuthorizedAt        string `json:"authorized_at"`
	ExpiresAt           string `json:"expires_at"`
	IdempotencyKey      string `json:"idempotency_key"`
	CorrelationID       string `json:"correlation_id"`
	ExecutionMode       string `json:"execution_mode"`
	ExecutionAuthorized bool   `json:"execution_authorized"`
}

type ExecutionRecord struct {
	ExecutionID     string `json:"execution_id"`
	AuthorizationID string `json:"authorization_id"`
	ApprovalID      string `json:"approval_id"`
	IntentID        string `json:"intent_id"`
	IntentHash      string `json:"intent_hash"`
	ExecutorID      string `json:"executor_id"`
	IdempotencyKey  string `json:"idempotency_key"`
	AttemptedAt     string `json:"attempted_at"`
	Status          string `json:"status"`
	SideEffectState string `json:"side_effect_state"`
	ExternalRef     string `json:"external_ref,omitempty"`
	Error           string `json:"error,omitempty"`
	CorrelationID   string `json:"correlation_id"`
}

type M09State struct {
	Intent               ShadowActionIntent       `json:"intent"`
	Policy               ShadowPolicyDecision     `json:"policy"`
	Approval             *ApprovalRecord          `json:"approval,omitempty"`
	Authorization        *ExecutionAuthorization  `json:"authorization,omitempty"`
	Execution            *ExecutionRecord         `json:"execution,omitempty"`
	ConsumedApprovalIDs  map[string]bool          `json:"consumed_approval_ids,omitempty"`
	SucceededIdempotency map[string]bool          `json:"succeeded_idempotency,omitempty"`
}

type M09Context struct {
	Now                string
	KillSwitch         bool
	Executor           ExecutorProfile
	AllowedExecutorIDs []string
	PolicyContext      ShadowPolicyContext
}

func containsFold(xs []string, v string) bool {
	for _, x := range xs {
		if strings.EqualFold(strings.TrimSpace(x), strings.TrimSpace(v)) {
			return true
		}
	}
	return false
}

func executorAllows(p ExecutorProfile, i ShadowActionIntent) bool {
	if p.ExecutorID == "" || !containsFold(p.AllowedActionTypes, i.ActionType) {
		return false
	}
	u, err := url.Parse(i.Target)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return false
	}
	return containsFold(p.AllowedHosts, u.Hostname())
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func policyEquivalentForExecution(stored, fresh ShadowPolicyDecision) bool {
	return stored.PolicyVersion == fresh.PolicyVersion &&
		stored.IntentID == fresh.IntentID &&
		stored.IntentHash == fresh.IntentHash &&
		stored.Decision == fresh.Decision &&
		stored.RiskClass == fresh.RiskClass &&
		stored.Reason == fresh.Reason &&
		stored.ApprovalRequired == fresh.ApprovalRequired &&
		stored.ShadowOnly && fresh.ShadowOnly &&
		!stored.ExecutionAuthorized && !fresh.ExecutionAuthorized
}

func executionForKey(state M09State, key string) *ExecutionRecord {
	if state.Execution == nil || state.Execution.IdempotencyKey != key {
		return nil
	}
	return state.Execution
}

func AuthorizeM09(state M09State, ctx M09Context) (ExecutionAuthorization, string) {
	i, p := state.Intent, state.Policy
	if i.IntentHash == "" || i.IntentHash != ComputeShadowIntentHash(i) {
		return ExecutionAuthorization{}, "DENY_TAMPERED_INTENT"
	}
	// M09 không mutate ActionIntent của M08 thành live intent. Intent vẫn là proposal
	// shadow/dry-run; live authority chỉ xuất hiện ở ExecutionAuthorization riêng.
	if !i.ShadowOnly || !i.DryRun {
		return ExecutionAuthorization{}, "DENY_INTENT_MODE"
	}

	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	intentExpires, errIntentExpires := time.Parse(time.RFC3339, i.ExpiresAt)
	if errNow != nil || errIntentExpires != nil || !intentExpires.After(now) {
		return ExecutionAuthorization{}, "DENY_EXPIRED_INTENT"
	}

	policyChecked, errPolicyChecked := time.Parse(time.RFC3339, p.PolicyCheckedAt)
	if p.IntentID != i.IntentID || p.IntentHash != i.IntentHash || p.PolicyVersion == "" ||
		errPolicyChecked != nil || policyChecked.After(now) || !p.ShadowOnly || p.ExecutionAuthorized ||
		(p.Decision != "ALLOW" && p.Decision != "HUMAN_REVIEW") {
		return ExecutionAuthorization{}, "DENY_POLICY_STATE"
	}

	freshCtx := ctx.PolicyContext
	freshCtx.Now = ctx.Now
	freshPolicy := EvaluateShadowPolicy(i, freshCtx)
	if freshPolicy.Reason == "POLICY_UNAVAILABLE" || !policyEquivalentForExecution(p, freshPolicy) {
		return ExecutionAuthorization{}, "DENY_POLICY_REVALIDATION"
	}
	if freshPolicy.Decision != "ALLOW" && freshPolicy.Decision != "HUMAN_REVIEW" {
		return ExecutionAuthorization{}, "DENY_POLICY_STATE"
	}

	if state.Approval == nil {
		return ExecutionAuthorization{}, "WAIT_APPROVAL"
	}
	a := *state.Approval
	if a.ApprovedBy != "human" || strings.TrimSpace(a.ApproverID) == "" || !a.OneTime {
		return ExecutionAuthorization{}, "DENY_INVALID_APPROVER"
	}
	if a.Decision == "REJECT" {
		return ExecutionAuthorization{}, "DENY_REJECTED"
	}
	if a.Decision != "APPROVE" {
		return ExecutionAuthorization{}, "DENY_INVALID_APPROVER"
	}
	if a.IntentID != i.IntentID || a.IntentHash != i.IntentHash || a.PolicyVersion != p.PolicyVersion || a.CorrelationID != i.CorrelationID {
		return ExecutionAuthorization{}, "DENY_APPROVAL_MISMATCH"
	}

	approvedAt, errApprovedAt := time.Parse(time.RFC3339, a.ApprovedAt)
	approvalExpires, errApprovalExpires := time.Parse(time.RFC3339, a.ExpiresAt)
	if errApprovedAt != nil || errApprovalExpires != nil || approvedAt.After(now) ||
		!approvalExpires.After(now) || !approvalExpires.After(approvedAt) {
		return ExecutionAuthorization{}, "DENY_EXPIRED_APPROVAL"
	}
	if approvedAt.Before(policyChecked) {
		return ExecutionAuthorization{}, "DENY_APPROVAL_BEFORE_POLICY"
	}

	if existing := executionForKey(state, i.IdempotencyKey); existing != nil {
		switch existing.Status {
		case "SUCCEEDED":
			return ExecutionAuthorization{}, "WAIT_ALREADY_EXECUTED"
		case "RECONCILIATION_REQUIRED":
			return ExecutionAuthorization{}, "WAIT_RECONCILIATION"
		}
	}
	if state.SucceededIdempotency != nil && state.SucceededIdempotency[i.IdempotencyKey] {
		return ExecutionAuthorization{}, "WAIT_ALREADY_EXECUTED"
	}
	if state.ConsumedApprovalIDs != nil && state.ConsumedApprovalIDs[a.ApprovalID] {
		return ExecutionAuthorization{}, "WAIT_APPROVAL_CONSUMED"
	}

	if ctx.KillSwitch {
		return ExecutionAuthorization{}, "DENY_KILL_SWITCH"
	}
	if !containsFold(ctx.AllowedExecutorIDs, ctx.Executor.ExecutorID) || !executorAllows(ctx.Executor, i) {
		return ExecutionAuthorization{}, "DENY_EXECUTOR"
	}

	expires := minTime(intentExpires, approvalExpires)
	return ExecutionAuthorization{
		AuthorizationID:     "auth-" + a.ApprovalID,
		IntentID:            i.IntentID,
		IntentHash:          i.IntentHash,
		PolicyVersion:       p.PolicyVersion,
		ApprovalID:          a.ApprovalID,
		ExecutorID:          ctx.Executor.ExecutorID,
		AuthorizedAt:        ctx.Now,
		ExpiresAt:           expires.Format(time.RFC3339),
		IdempotencyKey:      i.IdempotencyKey,
		CorrelationID:       i.CorrelationID,
		ExecutionMode:       "APPROVED_LIVE",
		ExecutionAuthorized: true,
	}, "AUTHORIZED"
}

func PersistM09State(path string, s M09State) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadM09State(path string) (M09State, error) {
	var s M09State
	b, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(b, &s)
	return s, err
}

func authorizationBindingMatches(got, expected ExecutionAuthorization) bool {
	return got.AuthorizationID == expected.AuthorizationID &&
		got.IntentID == expected.IntentID &&
		got.IntentHash == expected.IntentHash &&
		got.PolicyVersion == expected.PolicyVersion &&
		got.ApprovalID == expected.ApprovalID &&
		got.ExecutorID == expected.ExecutorID &&
		got.ExpiresAt == expected.ExpiresAt &&
		got.IdempotencyKey == expected.IdempotencyKey &&
		got.CorrelationID == expected.CorrelationID &&
		got.ExecutionMode == "APPROVED_LIVE" &&
		got.ExecutionAuthorized
}

func ensureM09StateMaps(state *M09State) {
	if state.ConsumedApprovalIDs == nil {
		state.ConsumedApprovalIDs = map[string]bool{}
	}
	if state.SucceededIdempotency == nil {
		state.SucceededIdempotency = map[string]bool{}
	}
}

func sandboxIdempotencyPath(dir, key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(dir, "idem-"+hex.EncodeToString(sum[:])+".json")
}

func executionFailureRecord(state M09State, auth ExecutionAuthorization, now, status, effectState, externalRef, message string) ExecutionRecord {
	return ExecutionRecord{
		ExecutionID:     "exec-" + auth.AuthorizationID,
		AuthorizationID: auth.AuthorizationID,
		ApprovalID:      auth.ApprovalID,
		IntentID:        auth.IntentID,
		IntentHash:      auth.IntentHash,
		ExecutorID:      auth.ExecutorID,
		IdempotencyKey:  auth.IdempotencyKey,
		AttemptedAt:     now,
		Status:          status,
		SideEffectState: effectState,
		ExternalRef:     externalRef,
		Error:           message,
		CorrelationID:   state.Intent.CorrelationID,
	}
}

func ExecuteLocalSandbox(state *M09State, auth ExecutionAuthorization, ctx M09Context, dir string) (ExecutionRecord, string) {
	if state == nil {
		return ExecutionRecord{}, "DENY_AUTHORIZATION"
	}
	if ctx.KillSwitch {
		return ExecutionRecord{}, "DENY_KILL_SWITCH"
	}
	if state.Authorization == nil || *state.Authorization != auth {
		return ExecutionRecord{}, "DENY_AUTHORIZATION"
	}

	// Executor không tin artifact authorization một cách mù quáng: revalidate lại
	// intent + deterministic policy + human approval + executor + idempotency ở thời điểm side effect.
	expected, status := AuthorizeM09(*state, ctx)
	if status != "AUTHORIZED" {
		return ExecutionRecord{}, status
	}
	if !authorizationBindingMatches(auth, expected) {
		return ExecutionRecord{}, "DENY_AUTHORIZATION"
	}

	now, errNow := time.Parse(time.RFC3339, ctx.Now)
	authorizedAt, errAuthorizedAt := time.Parse(time.RFC3339, auth.AuthorizedAt)
	expiresAt, errExpiresAt := time.Parse(time.RFC3339, auth.ExpiresAt)
	if errNow != nil || errAuthorizedAt != nil || errExpiresAt != nil || authorizedAt.After(now) || !expiresAt.After(now) {
		return ExecutionRecord{}, "DENY_EXPIRED_AUTHORIZATION"
	}
	if state.Approval == nil {
		return ExecutionRecord{}, "DENY_AUTHORIZATION"
	}
	approvedAt, errApprovedAt := time.Parse(time.RFC3339, state.Approval.ApprovedAt)
	if errApprovedAt != nil || authorizedAt.Before(approvedAt) {
		return ExecutionRecord{}, "DENY_AUTHORIZATION"
	}
	if !containsFold(ctx.AllowedExecutorIDs, ctx.Executor.ExecutorID) || !executorAllows(ctx.Executor, state.Intent) || ctx.Executor.ExecutorID != "local_sandbox" {
		return ExecutionRecord{}, "DENY_EXECUTOR"
	}

	ensureM09StateMaps(state)
	if err := os.MkdirAll(dir, 0755); err != nil {
		rec := executionFailureRecord(*state, auth, ctx.Now, "FAILED", "NOT_PERFORMED", "", err.Error())
		state.Execution = &rec
		return rec, "EXECUTION_FAILED"
	}

	path := sandboxIdempotencyPath(dir, auth.IdempotencyKey)
	payload := map[string]any{
		"intent_id":       state.Intent.IntentID,
		"intent_hash":     state.Intent.IntentHash,
		"action_type":     state.Intent.ActionType,
		"target":          state.Intent.Target,
		"parameters":      state.Intent.Parameters,
		"approval_id":     auth.ApprovalID,
		"approved_by":     "human",
		"execution_mode":  auth.ExecutionMode,
		"idempotency_key": auth.IdempotencyKey,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		rec := executionFailureRecord(*state, auth, ctx.Now, "FAILED", "NOT_PERFORMED", "", err.Error())
		state.Execution = &rec
		return rec, "EXECUTION_FAILED"
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		if os.IsExist(err) {
			rec := executionFailureRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", path, "idempotency marker already exists without durable success state")
			state.Execution = &rec
			state.ConsumedApprovalIDs[auth.ApprovalID] = true
			return rec, "WAIT_RECONCILIATION"
		}
		rec := executionFailureRecord(*state, auth, ctx.Now, "FAILED", "NOT_PERFORMED", "", err.Error())
		state.Execution = &rec
		return rec, "EXECUTION_FAILED"
	}

	uncertain := func(message string) (ExecutionRecord, string) {
		rec := executionFailureRecord(*state, auth, ctx.Now, "RECONCILIATION_REQUIRED", "UNKNOWN", path, message)
		state.Execution = &rec
		state.ConsumedApprovalIDs[auth.ApprovalID] = true
		return rec, "WAIT_RECONCILIATION"
	}

	if _, err = f.Write(b); err != nil {
		_ = f.Close()
		return uncertain(err.Error())
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return uncertain(err.Error())
	}
	if err = f.Close(); err != nil {
		return uncertain(err.Error())
	}

	rec := ExecutionRecord{
		ExecutionID:     "exec-" + auth.AuthorizationID,
		AuthorizationID: auth.AuthorizationID,
		ApprovalID:      auth.ApprovalID,
		IntentID:        auth.IntentID,
		IntentHash:      auth.IntentHash,
		ExecutorID:      auth.ExecutorID,
		IdempotencyKey:  auth.IdempotencyKey,
		AttemptedAt:     ctx.Now,
		Status:          "SUCCEEDED",
		SideEffectState: "PERFORMED",
		ExternalRef:     path,
		CorrelationID:   auth.CorrelationID,
	}
	state.Execution = &rec
	state.ConsumedApprovalIDs[auth.ApprovalID] = true
	state.SucceededIdempotency[auth.IdempotencyKey] = true
	return rec, "EXECUTED"
}

func demoM09() (map[string]any, error) {
	i := SealShadowActionIntent(ShadowActionIntent{
		IntentID: "i9", DecisionID: "d9", EvidenceIDs: []string{"e9"},
		ActionType: "UPDATE_DRAFT", Target: "https://example.com/draft",
		Parameters: map[string]any{"title": "approved sandbox write"}, ProposedBy: "human",
		CreatedAt: "2026-09-03T07:00:00Z", ExpiresAt: "2026-09-03T09:00:00Z",
		CorrelationID: "c9", IdempotencyKey: "k9",
	})
	policyCtx := ShadowPolicyContext{
		PolicyVersion: "m09-v1", Now: "2026-09-03T07:10:00Z",
		KnownDecisionIDs: []string{"d9"}, KnownEvidenceIDs: []string{"e9"}, AllowedHosts: []string{"example.com"},
		ActionRisk: map[string]string{"UPDATE_DRAFT": "RISK1"}, SeenIdempotency: map[string]string{},
	}
	p := EvaluateShadowPolicy(i, policyCtx)
	a := ApprovalRecord{
		ApprovalID: "ap9", IntentID: i.IntentID, IntentHash: i.IntentHash, PolicyVersion: p.PolicyVersion,
		Decision: "APPROVE", ApprovedBy: "human", ApproverID: "learner",
		ApprovedAt: "2026-09-03T07:20:00Z", ExpiresAt: "2026-09-03T08:30:00Z",
		CorrelationID: i.CorrelationID, OneTime: true,
	}
	s := M09State{Intent: i, Policy: p, Approval: &a}
	ctx := M09Context{
		Now: "2026-09-03T08:00:00Z",
		Executor: ExecutorProfile{ExecutorID: "local_sandbox", AllowedActionTypes: []string{"UPDATE_DRAFT"}, AllowedHosts: []string{"example.com"}},
		AllowedExecutorIDs: []string{"local_sandbox"}, PolicyContext: policyCtx,
	}
	auth, status := AuthorizeM09(s, ctx)
	if status != "AUTHORIZED" {
		return nil, errors.New(status)
	}
	s.Authorization = &auth
	dir, err := os.MkdirTemp("", "m09-sandbox-")
	if err != nil {
		return nil, err
	}
	rec, execStatus := ExecuteLocalSandbox(&s, auth, ctx, dir)
	if execStatus != "EXECUTED" {
		return nil, errors.New(execStatus)
	}
	return map[string]any{"authorization": auth, "execution": rec, "authority": "HUMAN_APPROVAL_REQUIRED"}, nil
}
