package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m03"
	corem07 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m07"
	corem08 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m08"
	corem10 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m10"
	corem11 "github.com/dvha85/affiliate-expert-learning-roadmap-v2/core/m11"
	"github.com/dvha85/affiliate-expert-learning-roadmap-v2/lab/affiliate-bot/internal/store"
)

const missionStateVersion = "learner-m08-m11/v1"

// missionClock is intentionally process-owned: production commands always use
// the wall clock, while in-package tests may replace this unexported seam to
// exercise exact authority-expiry boundaries. No CLI argument or environment
// variable can choose it, so an operator cannot backdate authority admission.
var missionClock = time.Now

func missionNowUTC() time.Time { return missionClock().UTC() }

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
	corem10.CanaryGrant
	Status         string `json:"status"`
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
	ReservationID   string `json:"reservation_id"`
	GrantID         string `json:"grant_id"`
	IntentID        string `json:"intent_id"`
	IntentHash      string `json:"intent_hash"`
	CostMinor       int64  `json:"cost_minor"`
	CostBoundID     string `json:"cost_bound_id,omitempty"`
	CostBoundHash   string `json:"cost_bound_hash,omitempty"`
	AuthorizationID string `json:"authorization_id,omitempty"`
	ExecutionID     string `json:"execution_id,omitempty"`
	ReservationMode string `json:"reservation_mode,omitempty"`
	ReservedAt      string `json:"reserved_at"`
}

func trustedCostBoundsPath(dir string) string   { return filepath.Join(dir, "trusted-cost-bounds.jsonl") }
func m10ArtifactRegistryPath(dir string) string { return filepath.Join(dir, "m10-artifacts.jsonl") }
func m10OutcomeStorePath(dir string) string     { return filepath.Join(dir, "m10-outcomes.jsonl") }
func m10ExecutionJournalPath(dir string) string {
	return filepath.Join(dir, "m10-execution-journal.json")
}
func m11OutcomeStorePath(dir string) string   { return filepath.Join(dir, "m11-outcomes.jsonl") }
func m11OutcomeJournalPath(dir string) string { return filepath.Join(dir, "m11-outcome-journal.json") }

func missionErrorStatus(err error) string {
	if err != nil && strings.HasPrefix(err.Error(), "durable STOP:") {
		return "STOPPED"
	}
	return "REJECTED"
}
func m11FailedExecutionJournalPath(dir string) string {
	return filepath.Join(dir, "m11-failed-execution-journal.json")
}
func m11UnknownStopJournalPath(dir string) string {
	return filepath.Join(dir, "m11-unknown-stop-journal.json")
}

func m11JournalRecoveryRequired(dir string) error {
	for _, item := range []struct {
		name string
		path string
	}{
		{name: "failed execution", path: m11FailedExecutionJournalPath(dir)},
		{name: "unknown STOP", path: m11UnknownStopJournalPath(dir)},
		{name: "outcome", path: m11OutcomeJournalPath(dir)},
	} {
		info, err := os.Lstat(item.path)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
				return fmt.Errorf("M11 %s journal path is not a regular file", item.name)
			}
			return fmt.Errorf("M11 %s journal requires a locked writer recovery", item.name)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// readM11Journal refuses a symlink or special file before parsing a recovery
// plan. A pending journal is an authority boundary: following a path outside
// the runtime could turn an unrelated file into a replay instruction.
func readM11Journal(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("M11 journal path is not a regular file")
	}
	return os.ReadFile(path)
}

// m11OutcomeAppendFault is a test-only seam for the two-file M11 outcome
// commit. It is never configurable through the CLI.
var m11OutcomeAppendFault func(phase string) error

func m11OutcomeWriteFault(phase string) error {
	if m11OutcomeAppendFault == nil {
		return nil
	}
	return m11OutcomeAppendFault(phase)
}

type m11OutcomeJournal struct {
	Version                string                   `json:"version"`
	PredecessorArtifactID  string                   `json:"predecessor_artifact_id"`
	PredecessorContentHash string                   `json:"predecessor_content_hash"`
	Outcome                m03.OutcomeRecord        `json:"outcome"`
	Ledger                 corem11.ProductionLedger `json:"ledger"`
}

// m10ExecutionJournal makes the immutable execution-record append and mutable
// reservation binding recoverable as one bounded local transition. It records
// no side effect and does not make the M10 fixture executor operational.
type m10ExecutionJournal struct {
	Version         string                  `json:"version"`
	ReservationID   string                  `json:"reservation_id"`
	AuthorizationID string                  `json:"authorization_id"`
	Record          corem10.ExecutionRecord `json:"record"`
}

// m11FailedExecutionJournal makes the known FAILED/NOT_PERFORMED execution
// transition replayable. The immutable execution and its post-execution ledger
// must either both become visible or remain recoverable from this exact plan.
type m11FailedExecutionJournal struct {
	Version                string                            `json:"version"`
	PredecessorArtifactID  string                            `json:"predecessor_artifact_id"`
	PredecessorContentHash string                            `json:"predecessor_content_hash"`
	Execution              corem11.ProductionExecutionRecord `json:"execution"`
	Ledger                 corem11.ProductionLedger          `json:"ledger"`
}

// m11UnknownStopJournal makes the UNKNOWN → stopped-ledger → durable STOP
// transition replayable after an interrupted write. It is not a multi-file
// transaction; the journal only gives the loader an exact, fail-closed replay
// plan for this one transition.
type m11UnknownStopJournal struct {
	Version                string                            `json:"version"`
	PredecessorArtifactID  string                            `json:"predecessor_artifact_id"`
	PredecessorContentHash string                            `json:"predecessor_content_hash"`
	Execution              corem11.ProductionExecutionRecord `json:"execution"`
	StoppedLedger          corem11.ProductionLedger          `json:"stopped_ledger"`
}

// The registry lives beside mission-state.json and is append-only. It owns the
// canonical compact JSON used for later resolution; user-supplied output files
// are only portable views of those registered artifacts.
func loadM10ArtifactRegistry(dir string) ([]corem10.ArtifactEntry, error) {
	raw, err := os.ReadFile(m10ArtifactRegistryPath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entries := []corem10.ArtifactEntry{}
	seen := map[string]string{}
	for _, line := range bytes.Split(raw, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		entry, err := corem10.ValidateArtifactEntry(line)
		if err != nil {
			return nil, fmt.Errorf("invalid M10 artifact registry entry: %w", err)
		}
		key := entry.ArtifactKind + "\x00" + entry.ArtifactID
		if prior, exists := seen[key]; exists {
			if prior == entry.ContentHash {
				return nil, fmt.Errorf("duplicate M10 artifact registry entry")
			}
			return nil, fmt.Errorf("M10 artifact ID reused with different content")
		}
		seen[key] = entry.ContentHash
		entries = append(entries, entry)
	}
	if err := validateM10ArtifactGraph(entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func validateM10ArtifactGraph(entries []corem10.ArtifactEntry) error {
	grants := map[string]corem10.CanaryGrant{}
	bounds := map[string]corem10.TrustedCostBound{}
	gates := map[string]corem10.CanaryGateDecision{}
	authorizations := map[string]corem10.ExecutionAuthorization{}
	records := []corem10.ExecutionRecord{}
	for _, entry := range entries {
		switch entry.ArtifactKind {
		case corem10.ArtifactKindCanaryGrant:
			grant, status := corem10.DecodeCanaryGrant(entry.Artifact)
			if status != "VALID" {
				return fmt.Errorf("invalid registered canary grant")
			}
			grants[grant.GrantID] = grant
		case corem10.ArtifactKindTrustedCostBound:
			bound, status := corem10.DecodeTrustedCostBound(entry.Artifact)
			if status != "VALID" {
				return fmt.Errorf("invalid registered trusted cost bound")
			}
			bounds[bound.CostBoundID] = bound
		case corem10.ArtifactKindCanaryGate:
			gate, err := corem10.ValidateCanaryGateDecision(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered canary gate: %w", err)
			}
			gates[gate.GateID] = gate
		case corem10.ArtifactKindExecutionAuthorization:
			authorization, err := corem10.ValidateExecutionAuthorization(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered execution authorization: %w", err)
			}
			authorizations[authorization.AuthorizationID] = authorization
		case corem10.ArtifactKindExecutionRecord:
			record, err := corem10.ValidateExecutionRecord(entry.Artifact)
			if err != nil {
				return fmt.Errorf("invalid registered execution record: %w", err)
			}
			records = append(records, record)
		default:
			return fmt.Errorf("unsupported registered M10 artifact kind")
		}
	}
	for _, gate := range gates {
		grant, grantOK := grants[gate.GrantID]
		bound, boundOK := bounds[gate.CostBoundID]
		if !grantOK || !boundOK || grant.GrantVersion != gate.GrantVersion || grant.GrantHash != gate.GrantHash || bound.CostBoundHash != gate.CostBoundHash || bound.MaxCostMinor != gate.CostBoundMinor || bound.IntentID != gate.IntentID || bound.IntentHash != gate.IntentHash || grant.PolicyVersion != gate.PolicyVersion {
			return fmt.Errorf("canary gate has an orphaned or mismatched registry link")
		}
	}
	for _, authorization := range authorizations {
		grant, grantOK := grants[authorization.CanaryGrantID]
		gate, gateOK := gates[authorization.CanaryGateID]
		bound, boundOK := bounds[authorization.CanaryCostBoundID]
		if !grantOK || !gateOK || !boundOK || grant.GrantVersion != authorization.CanaryGrantVersion || grant.GrantHash != authorization.CanaryGrantHash || gate.GrantID != authorization.CanaryGrantID || gate.GrantVersion != authorization.CanaryGrantVersion || gate.GrantHash != authorization.CanaryGrantHash || gate.IntentID != authorization.IntentID || gate.IntentHash != authorization.IntentHash || gate.PolicyVersion != authorization.PolicyVersion || gate.CostBoundID != authorization.CanaryCostBoundID || gate.CostBoundHash != authorization.CanaryCostBoundHash || gate.CostBoundMinor != authorization.CanaryCostBoundMinor || bound.CostBoundHash != authorization.CanaryCostBoundHash || bound.MaxCostMinor != authorization.CanaryCostBoundMinor {
			return fmt.Errorf("execution authorization has an orphaned or mismatched registry link")
		}
	}
	for _, record := range records {
		authorization, exists := authorizations[record.AuthorizationID]
		attemptedAt, attemptedErr := time.Parse(time.RFC3339, record.AttemptedAt)
		authorizedAt, authorizedErr := time.Parse(time.RFC3339, authorization.AuthorizedAt)
		expiresAt, expiresErr := time.Parse(time.RFC3339, authorization.ExpiresAt)
		if !exists || attemptedErr != nil || authorizedErr != nil || expiresErr != nil || attemptedAt.Before(authorizedAt) || !attemptedAt.Before(expiresAt) || authorization.IntentID != record.IntentID || authorization.IntentHash != record.IntentHash || authorization.ExecutorID != record.ExecutorID || authorization.IdempotencyKey != record.IdempotencyKey || authorization.CorrelationID != record.CorrelationID || authorization.CanaryGrantID != record.CanaryGrantID || authorization.CanaryGrantVersion != record.CanaryGrantVersion || authorization.CanaryGrantHash != record.CanaryGrantHash || authorization.CanaryGateID != record.CanaryGateID || authorization.CanaryCostBoundID != record.CanaryCostBoundID || authorization.CanaryCostBoundHash != record.CanaryCostBoundHash || authorization.CanaryCostBoundMinor != record.CanaryCostBoundMinor {
			return fmt.Errorf("execution record has an orphaned or mismatched registry link")
		}
	}
	return nil
}

func registerM10Artifact(dir, kind string, raw []byte) (corem10.ArtifactEntry, string, error) {
	entry, err := corem10.NewArtifactEntry(kind, raw)
	if err != nil {
		return entry, "", err
	}
	entries, err := loadM10ArtifactRegistry(dir)
	if err != nil {
		return entry, "", err
	}
	for _, registered := range entries {
		if registered.ArtifactKind != entry.ArtifactKind || registered.ArtifactID != entry.ArtifactID {
			continue
		}
		if registered.ContentHash == entry.ContentHash {
			return entry, appendDuplicate, nil
		}
		return entry, "", fmt.Errorf("M10 artifact ID reused with different content")
	}
	if err := validateM10ArtifactGraph(append(entries, entry)); err != nil {
		return entry, "", err
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return entry, "", err
	}
	f, err := os.OpenFile(m10ArtifactRegistryPath(dir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return entry, "", err
	}
	_, err = f.Write(append(line, '\n'))
	if syncErr := f.Sync(); err == nil {
		err = syncErr
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return entry, "", err
	}
	return entry, appendAdded, nil
}

func resolveM10Artifact(dir, kind string, raw []byte) bool {
	expected, err := corem10.NewArtifactEntry(kind, raw)
	if err != nil {
		return false
	}
	entries, err := loadM10ArtifactRegistry(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.ArtifactKind == expected.ArtifactKind && entry.ArtifactID == expected.ArtifactID && entry.ContentHash == expected.ContentHash {
			return true
		}
	}
	return false
}

func resolveM10ArtifactByID(dir, kind, artifactID, contentHash string) (corem10.ArtifactEntry, error) {
	entries, err := loadM10ArtifactRegistry(dir)
	if err != nil {
		return corem10.ArtifactEntry{}, err
	}
	for _, entry := range entries {
		if entry.ArtifactKind != kind || entry.ArtifactID != artifactID {
			continue
		}
		if contentHash != "" && entry.ContentHash != contentHash {
			return corem10.ArtifactEntry{}, fmt.Errorf("M10 artifact content hash does not match")
		}
		return entry, nil
	}
	return corem10.ArtifactEntry{}, fmt.Errorf("M10 artifact is not registered")
}

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
	raw, err := json.Marshal(bound)
	if err != nil || !resolveM10Artifact(dir, corem10.ArtifactKindTrustedCostBound, raw) {
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
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	return decoder.Decode(value)
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
	// A crash can occur after M11 has durably written a stopped ledger but
	// before mission-state/STOP is committed. Never let that inconsistent
	// directory resume on the stale mutable state.
	if !s.Stop {
		entries, loadErr := loadM11ArtifactRegistry(dir)
		if loadErr != nil {
			return fmt.Errorf("M11 registry integrity check failed: %w", loadErr)
		}
		for _, entry := range entries {
			if entry.ArtifactKind != corem11.ArtifactKindLedger {
				continue
			}
			value, status := corem11.DecodeArtifact("ledger", entry.Artifact)
			if status != corem11.Valid {
				return fmt.Errorf("M11 ledger integrity check failed")
			}
			if value.(*corem11.ProductionLedger).ControlMode == "STOPPED" {
				return fmt.Errorf("stopped M11 ledger requires durable mission STOP")
			}
		}
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
		grantRaw, marshalErr := json.Marshal(s.Canary.CanaryGrant)
		_, grantStatus := corem10.DecodeCanaryGrant(grantRaw)
		validFrom, validFromErr := time.Parse(time.RFC3339, s.Canary.ValidFrom)
		if marshalErr != nil || grantStatus != "VALID" || validFromErr != nil || s.Intent == nil || s.Approval == nil || s.Policy == nil || s.Canary.GrantID == "" || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed > s.Canary.MaxExecutionsTotal || s.Canary.CostUsedMinor > s.Canary.MaxCostMinorTotal || corem10.ValidCanaryGrantFor(s.Canary.CanaryGrant, s.Intent.IntentID, s.Intent.IntentHash, s.Policy.PolicyVersion, s.Approval.ApprovalID, s.Approval.ApproverID, s.Intent.CorrelationID, s.Policy.RiskClass, s.Intent.ActionType, s.Intent.Target, validFrom) != "VALID" {
			return fmt.Errorf("mission canary integrity/binding/budget check failed")
		}
	}
	seenReservations := map[string]bool{}
	seenAuthorizations := map[string]bool{}
	reservedCostMinor := int64(0)
	for _, r := range s.Reservations {
		if r.ReservationID == "" || seenReservations[r.ReservationID] || s.Canary == nil || s.Intent == nil || r.GrantID != s.Canary.GrantID || r.IntentID != s.Intent.IntentID || r.IntentHash != s.Intent.IntentHash || r.CostMinor < 0 {
			return fmt.Errorf("mission reservation integrity/binding check failed")
		}
		if r.AuthorizationID != "" {
			if r.ReservationMode != "GOVERNED_AUTHORIZATION" || r.CostBoundID == "" || r.CostBoundHash == "" || seenAuthorizations[r.AuthorizationID] {
				return fmt.Errorf("mission governed reservation integrity/binding check failed")
			}
			seenAuthorizations[r.AuthorizationID] = true
		} else if r.ExecutionID != "" || (r.ReservationMode != "" && r.ReservationMode != "LEGACY_COMPAT") {
			return fmt.Errorf("mission legacy reservation cannot bind execution")
		}
		if _, err := time.Parse(time.RFC3339, r.ReservedAt); err != nil {
			return fmt.Errorf("mission reservation timestamp is invalid")
		}
		if r.CostMinor > 0 && reservedCostMinor > (int64(^uint64(0)>>1))-r.CostMinor {
			return fmt.Errorf("mission reservation cost ledger overflows")
		}
		reservedCostMinor += r.CostMinor
		seenReservations[r.ReservationID] = true
	}
	if s.Canary != nil && (s.Canary.ExecutionsUsed != len(s.Reservations) || s.Canary.CostUsedMinor != reservedCostMinor) {
		return fmt.Errorf("mission canary usage does not match reservations")
	}
	return nil
}
func saveMissionState(dir string, s LearnerMissionState) error {
	s.Version = missionStateVersion
	s.UpdatedAt = missionNowUTC().Format(time.RFC3339Nano)
	return writeJSONAtomic(missionStatePath(dir), s)
}

// missionStateWriteFault is a test-only fault seam. It is never set from CLI
// input or environment and lets tests prove that the atomic rename boundary
// preserves the prior state on a failed write.
var missionStateWriteFault func(phase string) error

func missionWriteFault(phase string) error {
	if missionStateWriteFault == nil {
		return nil
	}
	return missionStateWriteFault(phase)
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
	if err := missionWriteFault("after_temp_sync"); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := missionWriteFault("before_rename"); err != nil {
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

// missionCanaryActive is deliberately separate from loadMissionState. A
// restored historical grant may be expired yet still be a valid record to
// replay and audit; only an operation that would create new authority or a
// reservation must require it to be active now.
func missionCanaryActive(s LearnerMissionState, now time.Time) error {
	if err := missionAuthorityActive(s, now); err != nil {
		return err
	}
	if s.Canary == nil {
		return fmt.Errorf("canary grant is required")
	}
	if status := corem10.ValidCanaryGrantFor(s.Canary.CanaryGrant, s.Intent.IntentID, s.Intent.IntentHash, s.Policy.PolicyVersion, s.Approval.ApprovalID, s.Approval.ApproverID, s.Intent.CorrelationID, s.Policy.RiskClass, s.Intent.ActionType, s.Intent.Target, now); status != "VALID" {
		return fmt.Errorf("canary grant: %s", status)
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

func resolveM07Proposal(record HistoryRecord, proposalPath string) (corem07.RegisteredAgentProposal, corem07.AgentOutput, error) {
	raw, err := os.ReadFile(proposalPath)
	if err != nil {
		return corem07.RegisteredAgentProposal{}, corem07.AgentOutput{}, err
	}
	ctx, err := m07EvidenceContext(record)
	if err != nil {
		return corem07.RegisteredAgentProposal{}, corem07.AgentOutput{}, err
	}
	return corem07.ValidateRegisteredAgentProposal(raw, ctx.Evidence, nil, record.RecordID)
}

func sameParameters(raw json.RawMessage, parameters map[string]any) bool {
	value, err := contracts.Decode(raw)
	if err != nil {
		return false
	}
	canonicalProposal, err := json.Marshal(value)
	if err != nil {
		return false
	}
	canonicalIntent, err := json.Marshal(parameters)
	return err == nil && bytes.Equal(canonicalProposal, canonicalIntent)
}

func buildLearnerIntent(historyPath, requestPath, proposalPath string) (LearnerIntent, error) {
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
	// M08 accepts the same canonical field IDs that M07 can ground, rather
	// than inventing a parallel aggregate-only ID namespace. The resolver has
	// already required exactly one replay-MATCH HistoryRecord.
	context, err := m07EvidenceContext(record)
	if err != nil {
		return LearnerIntent{}, fmt.Errorf("decision evidence resolution: %w", err)
	}
	allowed := map[string]bool{}
	for _, id := range context.EvidenceIDs {
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
		if proposalPath == "" {
			return LearnerIntent{}, fmt.Errorf("agent proposal_ref requires a persisted M07 proposal")
		}
		proposal, output, err := resolveM07Proposal(record, proposalPath)
		if err != nil {
			return LearnerIntent{}, fmt.Errorf("agent proposal resolution: %w", err)
		}
		if req.ProposalRef != proposal.ProposalID || output.ProposedAction == nil || req.ActionType != output.ProposedAction.ActionType || req.Target != output.ProposedAction.Target || !sameParameters(output.ProposedAction.Parameters, req.Parameters) {
			return LearnerIntent{}, fmt.Errorf("agent intent must exactly bind the persisted M07 proposed_action")
		}
		proposalEvidence := map[string]bool{}
		for _, id := range output.EvidenceIDs {
			proposalEvidence[id] = true
		}
		for _, id := range req.EvidenceIDs {
			if !proposalEvidence[id] {
				return LearnerIntent{}, fmt.Errorf("agent intent evidence_id %s is absent from proposal", id)
			}
		}
	} else if proposalPath != "" {
		return LearnerIntent{}, fmt.Errorf("human intent must not supply an agent proposal")
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

func evaluateLearnerPolicy(i LearnerIntent, path string, knownProposalIDs []string) (LearnerPolicy, error) {
	var req learnerPolicyRequest
	raw, err := os.ReadFile(path)
	if err != nil {
		return LearnerPolicy{}, err
	}
	if err := contracts.DecodeStrict(raw, &req); err != nil {
		return LearnerPolicy{}, err
	}
	ctx := corem08.PolicyContext{PolicyVersion: req.PolicyVersion, Now: req.Now, KnownDecisionIDs: []string{i.DecisionID}, KnownEvidenceIDs: append([]string(nil), i.EvidenceIDs...), KnownProposalIDs: knownProposalIDs, AllowedHosts: req.AllowedHosts, ActionRisk: req.ActionRisk, SeenIdempotency: req.SeenIdempotency}
	return LearnerPolicy(corem08.EvaluatePolicy(corem08.Intent(i), ctx)), nil
}

func evaluateLearnerCanaryGate(s LearnerMissionState, bound corem10.TrustedCostBound, evaluatedAt string) (corem10.CanaryGateDecision, error) {
	if s.Intent == nil || s.Policy == nil || s.Approval == nil || s.Canary == nil {
		return corem10.CanaryGateDecision{}, fmt.Errorf("active intent, policy, approval and canary grant required")
	}
	gate := corem10.EvaluateCanaryGate(corem10.CanaryGateInput{
		Grant: s.Canary.CanaryGrant, CostBound: bound, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash,
		PolicyVersion: s.Policy.PolicyVersion, PolicyDecision: s.Policy.Decision, RiskClass: s.Policy.RiskClass,
		ApprovalID: s.Approval.ApprovalID, ApproverID: s.Approval.ApproverID, CorrelationID: s.Intent.CorrelationID,
		ActionType: s.Intent.ActionType, Target: s.Intent.Target, Now: evaluatedAt,
		Ledger: corem10.CanaryLedgerSnapshot{ExecutionsTotal: s.Canary.ExecutionsUsed, ExecutionsInWindow: s.Canary.ExecutionsUsed, CostMinorTotal: s.Canary.CostUsedMinor},
	})
	raw, err := json.Marshal(gate)
	if err != nil {
		return corem10.CanaryGateDecision{}, err
	}
	if _, err := corem10.ValidateCanaryGateDecision(raw); err != nil {
		return corem10.CanaryGateDecision{}, err
	}
	return gate, nil
}

func authorizationBindsMissionState(authorization corem10.ExecutionAuthorization, s LearnerMissionState) bool {
	if s.Intent == nil || s.Policy == nil || s.Canary == nil {
		return false
	}
	return authorization.IntentID == s.Intent.IntentID && authorization.IntentHash == s.Intent.IntentHash &&
		authorization.PolicyVersion == s.Policy.PolicyVersion && authorization.CanaryGrantID == s.Canary.GrantID &&
		authorization.CanaryGrantVersion == s.Canary.GrantVersion && authorization.CanaryGrantHash == s.Canary.GrantHash &&
		authorization.CorrelationID == s.Intent.CorrelationID && authorization.IdempotencyKey == s.Intent.IdempotencyKey
}

func resolveAuthorizationCostBound(dir string, authorization corem10.ExecutionAuthorization) (corem10.TrustedCostBound, error) {
	entry, err := resolveM10ArtifactByID(dir, corem10.ArtifactKindTrustedCostBound, authorization.CanaryCostBoundID, "")
	if err != nil {
		return corem10.TrustedCostBound{}, err
	}
	bound, status := corem10.DecodeTrustedCostBound(entry.Artifact)
	if status != "VALID" || bound.CostBoundHash != authorization.CanaryCostBoundHash || bound.MaxCostMinor != authorization.CanaryCostBoundMinor {
		return corem10.TrustedCostBound{}, fmt.Errorf("authorization cost bound does not resolve exactly")
	}
	return bound, nil
}

func reservationForExecution(s LearnerMissionState, executionID string) bool {
	for _, reservation := range s.Reservations {
		if reservation.ReservationMode == "GOVERNED_AUTHORIZATION" && reservation.ExecutionID == executionID {
			return true
		}
	}
	return false
}

func reservationIndexForAuthorization(s LearnerMissionState, authorizationID string) int {
	for index, reservation := range s.Reservations {
		if reservation.AuthorizationID == authorizationID && reservation.ReservationMode == "GOVERNED_AUTHORIZATION" {
			return index
		}
	}
	return -1
}

func validateExecutionReservation(s LearnerMissionState, authorization corem10.ExecutionAuthorization, record corem10.ExecutionRecord) (int, error) {
	index := reservationIndexForAuthorization(s, authorization.AuthorizationID)
	if index == -1 {
		return -1, fmt.Errorf("governed authorization requires a bound reservation before execution record")
	}
	if priorExecutionID := s.Reservations[index].ExecutionID; priorExecutionID != "" && priorExecutionID != record.ExecutionID {
		return -1, fmt.Errorf("reservation already binds a different execution record")
	}
	reservedAt, reservedAtErr := time.Parse(time.RFC3339, s.Reservations[index].ReservedAt)
	attemptedAt, attemptedAtErr := time.Parse(time.RFC3339, record.AttemptedAt)
	if reservedAtErr != nil || attemptedAtErr != nil || attemptedAt.Before(reservedAt) {
		return -1, fmt.Errorf("execution record predates its budget reservation")
	}
	return index, nil
}

func removeM10ExecutionJournal(dir string) error {
	if err := os.Remove(m10ExecutionJournalPath(dir)); err != nil && !os.IsNotExist(err) {
		return err
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = parent.Sync()
	closeErr := parent.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func recoverM10ExecutionJournal(dir string) error {
	raw, err := os.ReadFile(m10ExecutionJournalPath(dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var journal m10ExecutionJournal
	if err := contracts.DecodeStrict(raw, &journal); err != nil || journal.Version != "m10-execution-journal/v1" {
		return fmt.Errorf("M10 execution journal is invalid")
	}
	recordRaw, err := json.Marshal(journal.Record)
	if err != nil {
		return err
	}
	if _, err := corem10.ValidateExecutionRecord(recordRaw); err != nil {
		return fmt.Errorf("M10 execution journal record is invalid: %w", err)
	}
	s, err := loadMissionState(dir)
	if err != nil {
		return err
	}
	reservationIndex := -1
	for index, reservation := range s.Reservations {
		if reservation.ReservationID == journal.ReservationID {
			reservationIndex = index
			break
		}
	}
	if reservationIndex == -1 || journal.AuthorizationID == "" || s.Reservations[reservationIndex].AuthorizationID != journal.AuthorizationID || journal.Record.AuthorizationID != journal.AuthorizationID {
		return fmt.Errorf("M10 execution journal reservation does not resolve")
	}
	authorizationEntry, err := resolveM10ArtifactByID(dir, corem10.ArtifactKindExecutionAuthorization, journal.AuthorizationID, "")
	if err != nil {
		return fmt.Errorf("M10 execution journal authorization does not resolve: %w", err)
	}
	authorization, err := corem10.ValidateExecutionAuthorization(authorizationEntry.Artifact)
	if err != nil || !authorizationBindsMissionState(authorization, s) {
		return fmt.Errorf("M10 execution journal authorization is invalid or mismatched")
	}
	if _, err := validateExecutionReservation(s, authorization, journal.Record); err != nil {
		return fmt.Errorf("M10 execution journal reservation is invalid: %w", err)
	}
	if _, _, err := registerM10Artifact(dir, corem10.ArtifactKindExecutionRecord, recordRaw); err != nil {
		return fmt.Errorf("M10 execution journal registry recovery failed: %w", err)
	}
	if bound := s.Reservations[reservationIndex].ExecutionID; bound == "" {
		s.Reservations[reservationIndex].ExecutionID = journal.Record.ExecutionID
		if err := saveMissionState(dir, s); err != nil {
			return err
		}
	} else if bound != journal.Record.ExecutionID {
		return fmt.Errorf("M10 execution journal conflicts with bound reservation execution")
	}
	return removeM10ExecutionJournal(dir)
}

func commitM10ExecutionRecord(dir string, s LearnerMissionState, reservationIndex int, authorization corem10.ExecutionAuthorization, record corem10.ExecutionRecord) error {
	if reservationIndex < 0 || reservationIndex >= len(s.Reservations) || s.Reservations[reservationIndex].AuthorizationID != authorization.AuthorizationID || record.AuthorizationID != authorization.AuthorizationID {
		return fmt.Errorf("M10 execution record reservation does not match authorization")
	}
	journal := m10ExecutionJournal{Version: "m10-execution-journal/v1", ReservationID: s.Reservations[reservationIndex].ReservationID, AuthorizationID: authorization.AuthorizationID, Record: record}
	if err := writeJSONAtomic(m10ExecutionJournalPath(dir), journal); err != nil {
		return err
	}
	return recoverM10ExecutionJournal(dir)
}

// validateM10FixtureOutcome deliberately permits only a terminal no-side-effect
// record. It is a local fixture measurement, not evidence of business impact.
func validateM10FixtureOutcome(dir string, s LearnerMissionState, raw []byte) (m03.OutcomeRecord, string) {
	outcome, status := m03.DecodeM03Outcome(raw)
	if status != "VALID" {
		return outcome, status
	}
	if outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" {
		return outcome, "REQUIRE_MACHINE_EXECUTION"
	}
	entry, err := resolveM10ArtifactByID(dir, corem10.ArtifactKindExecutionRecord, outcome.EffectRef.EffectID, "")
	if err != nil {
		return outcome, "ORPHAN_EXECUTION"
	}
	record, err := corem10.ValidateExecutionRecord(entry.Artifact)
	if err != nil || !reservationForExecution(s, record.ExecutionID) {
		return outcome, "ORPHAN_EXECUTION"
	}
	attemptedAt, attemptedErr := time.Parse(time.RFC3339, record.AttemptedAt)
	observedAt, observedErr := time.Parse(time.RFC3339, outcome.ObservedAt)
	if attemptedErr != nil || observedErr != nil || observedAt.Before(attemptedAt) {
		return outcome, "OUTCOME_BEFORE_EXECUTION"
	}
	if record.SideEffectState != "NOT_PERFORMED" || outcome.Status != "CANCELLED" || len(outcome.Metrics) != 0 || !strings.HasPrefix(outcome.SourceRef, "fixture:m10-outcome/") {
		return outcome, "INVALID_NO_SIDE_EFFECT_OUTCOME"
	}
	return outcome, "VALID"
}

func loadM10FixtureOutcomes(dir string, s LearnerMissionState) ([]m03.OutcomeRecord, error) {
	f, err := (store.JSONL{}).Open(m10OutcomeStorePath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	outcomes := []m03.OutcomeRecord{}
	seen := map[string]bool{}
	seenExecution := map[string]bool{}
	for scanner.Scan() {
		outcome, status := validateM10FixtureOutcome(dir, s, scanner.Bytes())
		if status != "VALID" || seen[outcome.OutcomeID] || seenExecution[outcome.EffectRef.EffectID] {
			return nil, fmt.Errorf("invalid M10 fixture outcome store")
		}
		seen[outcome.OutcomeID] = true
		seenExecution[outcome.EffectRef.EffectID] = true
		outcomes = append(outcomes, outcome)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return outcomes, nil
}

// validateM11FixtureOutcome accepts only an observation that proves a local
// fixture did not perform a side effect. It is deliberately not a production
// business-outcome ingestion path.
func validateM11FixtureOutcome(dir string, raw []byte) (m03.OutcomeRecord, corem11.ProductionExecutionRecord, string) {
	outcome, status := m03.DecodeM03Outcome(raw)
	if status != "VALID" {
		return outcome, corem11.ProductionExecutionRecord{}, status
	}
	if outcome.EffectRef.EffectKind != "MACHINE_EXECUTION" {
		return outcome, corem11.ProductionExecutionRecord{}, "REQUIRE_MACHINE_EXECUTION"
	}
	value, err := m11ArtifactValue(dir, corem11.ArtifactKindExecution, outcome.EffectRef.EffectID)
	if err != nil {
		return outcome, corem11.ProductionExecutionRecord{}, "ORPHAN_EXECUTION"
	}
	record := *value.(*corem11.ProductionExecutionRecord)
	attemptedAt, attemptedErr := time.Parse(time.RFC3339, record.AttemptedAt)
	observedAt, observedErr := time.Parse(time.RFC3339, outcome.ObservedAt)
	if attemptedErr != nil || observedErr != nil || observedAt.Before(attemptedAt) {
		return outcome, record, "OUTCOME_BEFORE_EXECUTION"
	}
	if record.Status != "FAILED" || record.SideEffectState != "NOT_PERFORMED" || outcome.Status != "CANCELLED" || len(outcome.Metrics) != 0 || !strings.HasPrefix(outcome.SourceRef, "fixture:m11-outcome/") {
		return outcome, record, "INVALID_NO_SIDE_EFFECT_OUTCOME"
	}
	return outcome, record, "VALID"
}

func loadM11FixtureOutcomes(dir string) ([]m03.OutcomeRecord, error) {
	f, err := (store.JSONL{}).Open(m11OutcomeStorePath(dir))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), store.MaxHistoryRecordBytes+2)
	outcomes := []m03.OutcomeRecord{}
	seen := map[string]bool{}
	for scanner.Scan() {
		outcome, _, status := validateM11FixtureOutcome(dir, scanner.Bytes())
		if status != "VALID" || seen[outcome.OutcomeID] {
			return nil, fmt.Errorf("invalid M11 fixture outcome store")
		}
		seen[outcome.OutcomeID] = true
		outcomes = append(outcomes, outcome)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return outcomes, nil
}

func appendM11FixtureOutcome(dir string, outcome m03.OutcomeRecord) (string, error) {
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return "", err
	}
	for _, prior := range outcomes {
		if prior.OutcomeID == outcome.OutcomeID {
			if reflect.DeepEqual(prior, outcome) {
				return appendDuplicate, nil
			}
			return "", fmt.Errorf("outcome_id reused with different content")
		}
		if prior.EffectRef.EffectID == outcome.EffectRef.EffectID {
			return "", fmt.Errorf("M11 execution already has an outcome")
		}
	}
	encoded, err := json.Marshal(outcome)
	if err != nil {
		return "", err
	}
	if err := m11OutcomeWriteFault("before_append"); err != nil {
		return "", err
	}
	if err := (store.JSONL{}).AppendLine(m11OutcomeStorePath(dir), encoded); err != nil {
		return "", err
	}
	if err := m11OutcomeWriteFault("after_append"); err != nil {
		return "", err
	}
	return appendAdded, nil
}

func nextM11FixtureOutcomeLedger(ledger corem11.ProductionLedger, outcome m03.OutcomeRecord, record corem11.ProductionExecutionRecord) (corem11.ProductionLedger, error) {
	if ledger.LeaseID != record.ProductionLeaseID || ledger.LeaseVersion != record.ProductionLeaseVersion || ledger.LeaseHash != record.ProductionLeaseHash || ledger.ReconciliationRequired {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 ledger cannot accept the outcome")
	}
	observed, observedErr := time.Parse(time.RFC3339, outcome.ObservedAt)
	updated, updatedErr := time.Parse(time.RFC3339, ledger.UpdatedAt)
	if observedErr != nil || updatedErr != nil || !observed.After(updated) {
		return corem11.ProductionLedger{}, fmt.Errorf("outcome time must advance the ledger")
	}
	pendingIndex := -1
	for index, id := range ledger.PendingExecutionIDs {
		if id == record.ExecutionID {
			pendingIndex = index
		}
	}
	if pendingIndex < 0 || ledger.PendingOutcomes < 1 {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 outcome has no pending execution")
	}
	for _, link := range ledger.OutcomeLinks {
		if link.OutcomeID == outcome.OutcomeID || link.ExecutionID == record.ExecutionID {
			return corem11.ProductionLedger{}, fmt.Errorf("M11 execution already has an outcome")
		}
	}
	next := ledger
	next.PendingOutcomes--
	// Keep the in-memory transition in the same canonical form as its JSON
	// artifact. In particular, an empty list must be [] rather than nil, or a
	// journal replay would correctly reject its own decoded snapshot.
	next.PendingExecutionIDs = make([]string, 0, len(ledger.PendingExecutionIDs)-1)
	next.PendingExecutionIDs = append(next.PendingExecutionIDs, ledger.PendingExecutionIDs[:pendingIndex]...)
	next.PendingExecutionIDs = append(next.PendingExecutionIDs, ledger.PendingExecutionIDs[pendingIndex+1:]...)
	next.OutcomeLinks = append(append([]corem11.ProductionOutcomeLink(nil), ledger.OutcomeLinks...), corem11.ProductionOutcomeLink{OutcomeID: outcome.OutcomeID, ExecutionID: record.ExecutionID, ObservedAt: outcome.ObservedAt})
	next.ConsecutiveFailures = 0
	next.LastOutcomeAt = outcome.ObservedAt
	next.UpdatedAt = outcome.ObservedAt
	return next, nil
}

// recoverM11OutcomeJournal completes a durable two-file outcome transition.
// The journal is written before either file is changed. Each replay step is
// idempotent, so a crash after the ledger append or after the outcome append
// cannot leave a caller with a false ACK or a permanently orphaned transition.
func recoverM11OutcomeJournal(dir string) error {
	raw, err := readM11Journal(m11OutcomeJournalPath(dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var journal m11OutcomeJournal
	if err := contracts.DecodeStrict(raw, &journal); err != nil || journal.Version != "m11-outcome-journal/v1" {
		return fmt.Errorf("M11 outcome journal is invalid")
	}
	ledgerRaw, err := json.Marshal(journal.Ledger)
	if err != nil {
		return err
	}
	if _, status := corem11.DecodeArtifact("ledger", ledgerRaw); status != corem11.Valid {
		return fmt.Errorf("M11 outcome journal ledger is invalid")
	}
	outcomeRaw, err := json.Marshal(journal.Outcome)
	if err != nil {
		return err
	}
	_, record, status := validateM11FixtureOutcome(dir, outcomeRaw)
	if status != "VALID" {
		return fmt.Errorf("M11 outcome journal outcome is invalid")
	}
	predecessorEntry, err := resolveM11Artifact(dir, corem11.ArtifactKindLedger, journal.PredecessorArtifactID, journal.PredecessorContentHash)
	if err != nil {
		return fmt.Errorf("M11 outcome journal predecessor does not resolve: %w", err)
	}
	predecessorValue, predecessorStatus := corem11.DecodeArtifact("ledger", predecessorEntry.Artifact)
	if predecessorStatus != corem11.Valid {
		return fmt.Errorf("M11 outcome journal predecessor is invalid")
	}
	expectedLedger, err := nextM11FixtureOutcomeLedger(*predecessorValue.(*corem11.ProductionLedger), journal.Outcome, record)
	if err != nil || !reflect.DeepEqual(expectedLedger, journal.Ledger) {
		return fmt.Errorf("M11 outcome journal transition does not match its predecessor")
	}
	expectedEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		return err
	}
	currentEntry, _, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil {
		return err
	}
	if currentEntry.ArtifactID == predecessorEntry.ArtifactID && currentEntry.ContentHash == predecessorEntry.ContentHash {
		if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, ledgerRaw); err != nil {
			return fmt.Errorf("M11 outcome journal ledger recovery failed: %w", err)
		}
	} else if currentEntry.ArtifactID != expectedEntry.ArtifactID || currentEntry.ContentHash != expectedEntry.ContentHash {
		return fmt.Errorf("M11 outcome journal is no longer the active ledger transition")
	}
	if _, err := appendM11FixtureOutcome(dir, journal.Outcome); err != nil {
		return fmt.Errorf("M11 outcome journal outcome recovery failed: %w", err)
	}
	if err := os.Remove(m11OutcomeJournalPath(dir)); err != nil && !os.IsNotExist(err) {
		return err
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = parent.Sync()
	closeErr := parent.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func nextM11FailedExecutionLedger(ledger corem11.ProductionLedger, record corem11.ProductionExecutionRecord) (corem11.ProductionLedger, error) {
	if record.Status != "FAILED" || record.SideEffectState != "NOT_PERFORMED" || ledger.LeaseID != record.ProductionLeaseID || ledger.LeaseVersion != record.ProductionLeaseVersion || ledger.LeaseHash != record.ProductionLeaseHash || ledger.ControlMode != "NORMAL" || ledger.ReconciliationRequired {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 failed execution journal cannot transition its predecessor")
	}
	attempted, attemptedErr := time.Parse(time.RFC3339, record.AttemptedAt)
	updated, updatedErr := time.Parse(time.RFC3339, ledger.UpdatedAt)
	if attemptedErr != nil || updatedErr != nil || !attempted.After(updated) {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 failed execution time must advance the ledger")
	}
	pending := false
	for _, id := range ledger.PendingExecutionIDs {
		pending = pending || id == record.ExecutionID
	}
	if !pending || ledger.PendingOutcomes < 1 {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 failed execution has no pending reservation")
	}
	next := ledger
	next.ConsecutiveFailures++
	next.LastExecutionAt = record.AttemptedAt
	next.UpdatedAt = record.AttemptedAt
	return next, nil
}

// recoverM11FailedExecutionJournal completes the bounded execution → ledger
// transition after a write error. It never manufactures an execution: both
// payloads are validated against the exact predecessor before either append is
// retried, and a competing head leaves the journal in place for investigation.
func recoverM11FailedExecutionJournal(dir string) error {
	raw, err := readM11Journal(m11FailedExecutionJournalPath(dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var journal m11FailedExecutionJournal
	if err := contracts.DecodeStrict(raw, &journal); err != nil || journal.Version != "m11-failed-execution-journal/v1" {
		return fmt.Errorf("M11 failed execution journal is invalid")
	}
	executionRaw, err := json.Marshal(journal.Execution)
	if err != nil {
		return err
	}
	if _, status := corem11.DecodeArtifact("execution", executionRaw); status != corem11.Valid {
		return fmt.Errorf("M11 failed execution journal execution is invalid")
	}
	ledgerRaw, err := json.Marshal(journal.Ledger)
	if err != nil {
		return err
	}
	if _, status := corem11.DecodeArtifact("ledger", ledgerRaw); status != corem11.Valid {
		return fmt.Errorf("M11 failed execution journal ledger is invalid")
	}
	predecessorEntry, err := resolveM11Artifact(dir, corem11.ArtifactKindLedger, journal.PredecessorArtifactID, journal.PredecessorContentHash)
	if err != nil {
		return fmt.Errorf("M11 failed execution journal predecessor does not resolve: %w", err)
	}
	predecessorValue, status := corem11.DecodeArtifact("ledger", predecessorEntry.Artifact)
	if status != corem11.Valid {
		return fmt.Errorf("M11 failed execution journal predecessor is invalid")
	}
	expectedLedger, err := nextM11FailedExecutionLedger(*predecessorValue.(*corem11.ProductionLedger), journal.Execution)
	if err != nil || !reflect.DeepEqual(expectedLedger, journal.Ledger) {
		return fmt.Errorf("M11 failed execution journal transition does not match its predecessor")
	}
	expectedLedgerEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		return err
	}
	expectedExecutionEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindExecution, executionRaw)
	if err != nil {
		return err
	}
	currentEntry, _, err := m11LedgerHead(dir, journal.Ledger.LeaseID)
	if err != nil {
		return err
	}
	if currentEntry.ArtifactID == predecessorEntry.ArtifactID && currentEntry.ContentHash == predecessorEntry.ContentHash {
		if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindExecution, executionRaw); err != nil {
			return fmt.Errorf("M11 failed execution journal execution recovery failed: %w", err)
		}
		if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, ledgerRaw); err != nil {
			return fmt.Errorf("M11 failed execution journal ledger recovery failed: %w", err)
		}
	} else if currentEntry.ArtifactID != expectedLedgerEntry.ArtifactID || currentEntry.ContentHash != expectedLedgerEntry.ContentHash {
		return fmt.Errorf("M11 failed execution journal is no longer the active ledger transition")
	} else {
		executionEntry, resolveErr := resolveM11Artifact(dir, corem11.ArtifactKindExecution, journal.Execution.ExecutionID, expectedExecutionEntry.ContentHash)
		if resolveErr != nil || executionEntry.ContentHash != expectedExecutionEntry.ContentHash {
			if resolveErr != nil {
				return fmt.Errorf("M11 failed execution journal execution is missing after ledger: %w", resolveErr)
			}
			return fmt.Errorf("M11 failed execution journal execution differs after ledger")
		}
	}
	if err := os.Remove(m11FailedExecutionJournalPath(dir)); err != nil && !os.IsNotExist(err) {
		return err
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = parent.Sync()
	closeErr := parent.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func nextM11UnknownStopLedger(ledger corem11.ProductionLedger, record corem11.ProductionExecutionRecord) (corem11.ProductionLedger, error) {
	if record.Status != "RECONCILIATION_REQUIRED" || record.SideEffectState != "UNKNOWN" || ledger.LeaseID != record.ProductionLeaseID || ledger.LeaseVersion != record.ProductionLeaseVersion || ledger.LeaseHash != record.ProductionLeaseHash || ledger.ControlMode != "NORMAL" || ledger.ReconciliationRequired || !mustM11Time(record.AttemptedAt).After(mustM11Time(ledger.UpdatedAt)) {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 unknown STOP journal cannot transition its predecessor")
	}
	pending := false
	for _, id := range ledger.PendingExecutionIDs {
		pending = pending || id == record.ExecutionID
	}
	if !pending || ledger.PendingOutcomes < 1 {
		return corem11.ProductionLedger{}, fmt.Errorf("M11 unknown STOP journal has no governed reservation")
	}
	next := ledger
	next.ControlMode = "STOPPED"
	next.StopReason = "RECONCILIATION_REQUIRED"
	next.ReconciliationRequired = true
	next.LastExecutionAt = record.AttemptedAt
	next.UpdatedAt = record.AttemptedAt
	return next, nil
}

// recoverM11UnknownStopJournal completes the one bounded UNKNOWN transition.
// The stopped ledger is appended before mission-state/STOP, and any later
// retry is exact. A malformed, stale, or competing journal remains on disk and
// blocks mutation rather than guessing which side effect occurred.
func recoverM11UnknownStopJournal(dir string) error {
	raw, err := readM11Journal(m11UnknownStopJournalPath(dir))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var journal m11UnknownStopJournal
	if err := contracts.DecodeStrict(raw, &journal); err != nil || journal.Version != "m11-unknown-stop-journal/v1" {
		return fmt.Errorf("M11 unknown STOP journal is invalid")
	}
	executionRaw, err := json.Marshal(journal.Execution)
	if err != nil {
		return err
	}
	if _, status := corem11.DecodeArtifact("execution", executionRaw); status != corem11.Valid {
		return fmt.Errorf("M11 unknown STOP journal execution is invalid")
	}
	ledgerRaw, err := json.Marshal(journal.StoppedLedger)
	if err != nil {
		return err
	}
	if _, status := corem11.DecodeArtifact("ledger", ledgerRaw); status != corem11.Valid {
		return fmt.Errorf("M11 unknown STOP journal ledger is invalid")
	}
	predecessorEntry, err := resolveM11Artifact(dir, corem11.ArtifactKindLedger, journal.PredecessorArtifactID, journal.PredecessorContentHash)
	if err != nil {
		return fmt.Errorf("M11 unknown STOP journal predecessor does not resolve: %w", err)
	}
	predecessorValue, predecessorStatus := corem11.DecodeArtifact("ledger", predecessorEntry.Artifact)
	if predecessorStatus != corem11.Valid {
		return fmt.Errorf("M11 unknown STOP journal predecessor is invalid")
	}
	expectedLedger, err := nextM11UnknownStopLedger(*predecessorValue.(*corem11.ProductionLedger), journal.Execution)
	if err != nil || !reflect.DeepEqual(expectedLedger, journal.StoppedLedger) {
		return fmt.Errorf("M11 unknown STOP journal transition does not match its predecessor")
	}
	expectedEntry, err := corem11.NewArtifactEntry(corem11.ArtifactKindLedger, ledgerRaw)
	if err != nil {
		return err
	}
	currentEntry, _, err := m11LedgerHead(dir, journal.StoppedLedger.LeaseID)
	if err != nil {
		return err
	}
	if currentEntry.ArtifactID == predecessorEntry.ArtifactID && currentEntry.ContentHash == predecessorEntry.ContentHash {
		if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindExecution, executionRaw); err != nil {
			return fmt.Errorf("M11 unknown STOP journal execution recovery failed: %w", err)
		}
		if _, _, err := registerM11Artifact(dir, corem11.ArtifactKindLedger, ledgerRaw); err != nil {
			return fmt.Errorf("M11 unknown STOP journal ledger recovery failed: %w", err)
		}
	} else if currentEntry.ArtifactID != expectedEntry.ArtifactID || currentEntry.ContentHash != expectedEntry.ContentHash {
		return fmt.Errorf("M11 unknown STOP journal is no longer the active ledger transition")
	} else if _, err := m11ArtifactValue(dir, corem11.ArtifactKindExecution, journal.Execution.ExecutionID); err != nil {
		return fmt.Errorf("M11 unknown STOP journal execution is missing after stopped ledger: %w", err)
	}
	if err := recoverM11DurableStop(dir); err != nil {
		return fmt.Errorf("M11 unknown STOP journal durable STOP recovery failed: %w", err)
	}
	if err := os.Remove(m11UnknownStopJournalPath(dir)); err != nil && !os.IsNotExist(err) {
		return err
	}
	parent, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = parent.Sync()
	closeErr := parent.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func recordM11FixtureOutcome(dir, ledgerID string, raw []byte) (m03.OutcomeRecord, corem11.ProductionLedger, string, error) {
	if err := recoverM11OutcomeJournal(dir); err != nil {
		return m03.OutcomeRecord{}, corem11.ProductionLedger{}, "", err
	}
	state, err := loadMissionState(dir)
	if err != nil {
		return m03.OutcomeRecord{}, corem11.ProductionLedger{}, "", err
	}
	if state.Stop {
		return m03.OutcomeRecord{}, corem11.ProductionLedger{}, "", fmt.Errorf("durable STOP: %s", state.StopReason)
	}
	outcome, record, validation := validateM11FixtureOutcome(dir, raw)
	if validation != "VALID" {
		return outcome, corem11.ProductionLedger{}, "", fmt.Errorf("M11 outcome rejected: %s", validation)
	}
	outcomes, err := loadM11FixtureOutcomes(dir)
	if err != nil {
		return outcome, corem11.ProductionLedger{}, "", err
	}
	for _, prior := range outcomes {
		if prior.OutcomeID != outcome.OutcomeID {
			if prior.EffectRef.EffectID == record.ExecutionID {
				return outcome, corem11.ProductionLedger{}, "", fmt.Errorf("M11 execution already has an outcome")
			}
			continue
		}
		if reflect.DeepEqual(prior, outcome) {
			return outcome, corem11.ProductionLedger{}, appendDuplicate, nil
		}
		return outcome, corem11.ProductionLedger{}, "", fmt.Errorf("outcome_id reused with different content")
	}
	ledgerEntry, ledger, err := m11LedgerHead(dir, record.ProductionLeaseID)
	if err != nil {
		return outcome, corem11.ProductionLedger{}, "", err
	}
	if ledgerEntry.ArtifactID != ledgerID {
		return outcome, corem11.ProductionLedger{}, "", fmt.Errorf("production ledger is not the current head")
	}
	next, err := nextM11FixtureOutcomeLedger(ledger, outcome, record)
	if err != nil {
		return outcome, corem11.ProductionLedger{}, "", err
	}
	journal := m11OutcomeJournal{Version: "m11-outcome-journal/v1", PredecessorArtifactID: ledgerEntry.ArtifactID, PredecessorContentHash: ledgerEntry.ContentHash, Outcome: outcome, Ledger: next}
	if err := writeJSONAtomic(m11OutcomeJournalPath(dir), journal); err != nil {
		return outcome, next, "", err
	}
	if err := recoverM11OutcomeJournal(dir); err != nil {
		return outcome, next, "", err
	}
	return outcome, next, appendAdded, nil
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
		return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT | m08-policy INTENT POLICY OUT | bind STATE_DIR INTENT POLICY | m09-approval STATE_DIR APPROVAL | m10-canary STATE_DIR GRANT | m10-cost-register STATE_DIR COST_BOUND | m10-gate STATE_DIR COST_BOUND OUT EVALUATED_AT | m10-authorize STATE_DIR COST_BOUND GATE OUT AUTHORIZED_AT EXECUTOR_ID | m10-reserve-authorization STATE_DIR AUTHORIZATION RESERVATION_ID | m10-record-failed STATE_DIR AUTHORIZATION OUT ATTEMPTED_AT FIXTURE_REASON | m10-cancel STATE_DIR AUTHORIZATION OUT ATTEMPTED_AT REASON | m10-outcome STATE_DIR OUTCOME_INPUT | m10-resolve STATE_DIR KIND ARTIFACT_ID [CONTENT_HASH] | m10-reserve STATE_DIR COST_MINOR|COST_BOUND [RESERVATION_ID] | m11-register STATE_DIR KIND ARTIFACT_INPUT | m11-resolve STATE_DIR KIND ARTIFACT_ID [CONTENT_HASH] | m11-activate STATE_DIR LEASE_ID ACTIVATED_AT | m11-ledger-init STATE_DIR LEASE_ID INITIALIZED_AT | m11-gate STATE_DIR LEASE_ID HEALTH_ID COST_BOUND_ID LEDGER_ID EVALUATED_AT | m11-authorize STATE_DIR LEASE_ID GATE_ID EXECUTOR_ID AUTHORIZED_AT | m11-reserve-authorization STATE_DIR AUTHORIZATION_ID LEDGER_ID RESERVED_AT | m11-record-failed STATE_DIR AUTHORIZATION_ID RESERVATION_LEDGER_ID ATTEMPTED_AT FIXTURE_REASON | m11-record-unknown STATE_DIR AUTHORIZATION_ID RESERVATION_LEDGER_ID ATTEMPTED_AT REASON | m11-reconcile STATE_DIR RESOLUTION_ID STOPPED_LEDGER_ID | m11-recovery-export STATE_DIR RESOLUTION_ID STOPPED_LEDGER_ID OUT | m11-recovery-admit NEW_STATE_DIR OLD_STATE_DIR HANDOFF_INPUT ADMISSION_INPUT | m11-outcome STATE_DIR OUTCOME_INPUT LEDGER_ID | m11-evaluate STATE_DIR OUTCOME_ID EVALUATION_ID EVALUATED_AT | m11-close-cycle STATE_DIR CYCLE_ID EVALUATION_ID CLOSED_AT | m11-stop STATE_DIR REASON | status STATE_DIR"), 2)
	}
	// Directory creation and an exclusive lock make the mutable mission state
	// single-writer across processes. A stale lock fails closed and requires an
	// explicit recovery procedure rather than silently risking double reserve.
	mutatesState := map[string]bool{"bind": true, "m09-approval": true, "approval": true, "m10-canary": true, "canary": true, "m10-cost-register": true, "m10-gate": true, "m10-authorize": true, "m10-reserve-authorization": true, "m10-record-failed": true, "m10-cancel": true, "m10-outcome": true, "m10-reserve": true, "reserve": true, "m11-register": true, "m11-activate": true, "m11-ledger-init": true, "m11-gate": true, "m11-authorize": true, "m11-reserve-authorization": true, "m11-record-failed": true, "m11-record-unknown": true, "m11-reconcile": true, "m11-recovery-admit": true, "m11-outcome": true, "m11-evaluate": true, "m11-close-cycle": true, "m11-stop": true, "stop": true, "init": true}[args[0]]
	if mutatesState && len(args) >= 2 {
		if err := os.MkdirAll(args[1], 0700); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		releaseGate, err := acquireRuntimeGate(args[1])
		if err != nil {
			return emit("BUSY", nil, err, 1)
		}
		defer releaseGate()
		lockPath := filepath.Join(args[1], ".mission.lock")
		if err := os.Mkdir(lockPath, 0700); err != nil {
			if os.IsExist(err) {
				return emit("BUSY", nil, fmt.Errorf("mission state is locked; explicit recovery required after an interrupted writer"), 1)
			}
			return emit("STORE_ERROR", nil, err, 1)
		}
		defer os.Remove(lockPath)
		if err := recoverM10ExecutionJournal(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		if err := recoverM11FailedExecutionJournal(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		if err := recoverM11UnknownStopJournal(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		if err := recoverM11OutcomeJournal(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		if err := recoverM11DurableStop(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
	}
	switch args[0] {
	case "m08-intent", "intent":
		if len(args) != 4 && len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-intent HISTORY REQUEST OUT | bot mission m08-intent HISTORY REQUEST M07_PROPOSAL OUT"), 2)
		}
		if err := distinctPaths(args[1:]...); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		proposalPath, outputPath := "", args[3]
		if len(args) == 5 {
			proposalPath, outputPath = args[3], args[4]
		}
		i, err := buildLearnerIntent(args[1], args[2], proposalPath)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		status, err := writeNewJSON(outputPath, i)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		return emit(status, i, nil, 0)
	case "m08-policy", "policy":
		if len(args) != 4 && len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m08-policy INTENT POLICY OUT | bot mission m08-policy HISTORY INTENT POLICY M07_PROPOSAL OUT"), 2)
		}
		if err := distinctPaths(args[1:]...); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		intentPath, policyPath, outputPath := args[1], args[2], args[3]
		if len(args) == 6 {
			intentPath, policyPath, outputPath = args[2], args[3], args[5]
		}
		raw, err := os.ReadFile(intentPath)
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		decoded, state := corem08.DecodeIntent(raw)
		if state != "VALID" {
			return emit("INPUT_ERROR", nil, fmt.Errorf("invalid canonical M08 intent"), 1)
		}
		i := LearnerIntent(decoded)
		knownProposalIDs := []string{}
		if i.ProposedBy == "agent" {
			if len(args) != 6 {
				return emit("INPUT_ERROR", nil, fmt.Errorf("agent intent policy requires HISTORY and persisted M07 proposal"), 1)
			}
			record, err := resolveCanonicalRecord(args[1], i.DecisionID)
			if err != nil {
				return emit("INPUT_ERROR", nil, err, 1)
			}
			proposal, output, err := resolveM07Proposal(record, args[4])
			if err != nil || proposal.ProposalID != i.ProposalRef || output.ProposedAction == nil || i.ActionType != output.ProposedAction.ActionType || i.Target != output.ProposedAction.Target || !sameParameters(output.ProposedAction.Parameters, i.Parameters) {
				return emit("INPUT_ERROR", nil, fmt.Errorf("agent intent does not resolve to its persisted M07 proposal"), 1)
			}
			knownProposalIDs = []string{proposal.ProposalID}
		}
		p, err := evaluateLearnerPolicy(i, policyPath, knownProposalIDs)
		if err != nil {
			return emit("DENY", nil, err, 1)
		}
		status, err := writeNewJSON(outputPath, p)
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
		var p LearnerPolicy
		intentRaw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		decodedIntent, state := corem08.DecodeIntent(intentRaw)
		if state != "VALID" {
			return emit("INPUT_ERROR", nil, fmt.Errorf("invalid canonical M08 intent"), 1)
		}
		i := LearnerIntent(decodedIntent)
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
		// An intent ID is an immutable authority identity. Rebinding the same ID
		// to a different hash (including a changed expiry) would silently discard
		// its approval/grant history and reopen review under an ambiguous identity.
		if s.Intent != nil && s.Intent.IntentID == i.IntentID && s.Intent.IntentHash != i.IntentHash {
			return emit("REJECTED", nil, fmt.Errorf("cannot rebind an existing intent_id with different content"), 1)
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
		now := missionNowUTC()
		if approvedErr != nil || expiresErr != nil || !expiresAt.After(approvedAt) || !expiresAt.After(now) || approvedAt.After(now) {
			return emit("REJECTED", nil, fmt.Errorf("approval timestamps are invalid or expired"), 1)
		}
		if intentExpiry, intentErr := time.Parse(time.RFC3339, s.Intent.ExpiresAt); intentErr != nil || !intentExpiry.After(now) {
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
		if err := missionAuthorityActive(s, missionNowUTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		grant, status := corem10.DecodeCanaryGrant(raw)
		if status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("canary grant: %s", status), 1)
		}
		if status := corem10.ValidCanaryGrantFor(grant, s.Intent.IntentID, s.Intent.IntentHash, s.Policy.PolicyVersion, s.Approval.ApprovalID, s.Approval.ApproverID, s.Intent.CorrelationID, s.Policy.RiskClass, s.Intent.ActionType, s.Intent.Target, missionNowUTC()); status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("canary grant: %s", status), 1)
		}
		c := LearnerCanary{CanaryGrant: grant, Status: "ACTIVE"}
		if s.Canary != nil {
			if s.Canary.GrantID != c.GrantID || s.Canary.GrantHash != c.GrantHash {
				return emit("REJECTED", nil, fmt.Errorf("cannot replace an existing canary binding"), 1)
			}
			c.ExecutionsUsed = s.Canary.ExecutionsUsed
			c.CostUsedMinor = s.Canary.CostUsedMinor
		}
		grantRaw, err := json.Marshal(c.CanaryGrant)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, _, err := registerM10Artifact(args[1], corem10.ArtifactKindCanaryGrant, grantRaw); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
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
		if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, missionNowUTC()); status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
		}
		bounds, err := loadTrustedCostBounds(args[1])
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		boundAlreadyRegistered := false
		for _, old := range bounds {
			if old.CostBoundID == bound.CostBoundID {
				if old.CostBoundHash == bound.CostBoundHash {
					boundAlreadyRegistered = true
					break
				}
				return emit("CONFLICT", nil, fmt.Errorf("cost_bound_id reused with different content"), 1)
			}
		}
		if _, _, err := registerM10Artifact(args[1], corem10.ArtifactKindTrustedCostBound, raw); err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		if boundAlreadyRegistered {
			return emit("EXACT_DUPLICATE", bound, nil, 0)
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
	case "m10-gate":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-gate STATE_DIR COST_BOUND OUT EVALUATED_AT"), 2)
		}
		if err := distinctPaths(args[1], args[2], args[3]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop || s.Intent == nil || s.Policy == nil || s.Approval == nil || s.Canary == nil {
			return emit("REJECTED", nil, fmt.Errorf("active intent, policy, approval and canary grant required"), 1)
		}
		if err := missionCanaryActive(s, missionNowUTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		bound, status := corem10.DecodeTrustedCostBound(raw)
		if status != "VALID" || !resolveTrustedCostBound(args[1], bound) {
			return emit("REJECTED", nil, fmt.Errorf("cost bound is not a registered canonical artifact"), 1)
		}
		if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, missionNowUTC()); status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
		}
		gate, err := evaluateLearnerCanaryGate(s, bound, args[4])
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		gateRaw, err := json.Marshal(gate)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, _, err := registerM10Artifact(args[1], corem10.ArtifactKindCanaryGate, gateRaw); err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		status, err = writeNewJSON(args[3], gate)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		if status == appendDuplicate {
			return emit(status, gate, nil, 0)
		}
		return emit(gate.Decision, gate, nil, 0)
	case "m10-authorize":
		if len(args) != 7 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-authorize STATE_DIR COST_BOUND GATE OUT AUTHORIZED_AT EXECUTOR_ID"), 2)
		}
		if err := distinctPaths(args[1], args[2], args[3], args[4]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if s.Stop || s.Intent == nil || s.Policy == nil || s.Approval == nil || s.Canary == nil {
			return emit("REJECTED", nil, fmt.Errorf("active intent, policy, approval and canary grant required"), 1)
		}
		if err := missionCanaryActive(s, missionNowUTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		boundRaw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		bound, status := corem10.DecodeTrustedCostBound(boundRaw)
		if status != "VALID" || !resolveTrustedCostBound(args[1], bound) {
			return emit("REJECTED", nil, fmt.Errorf("cost bound is not a registered canonical artifact"), 1)
		}
		if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, missionNowUTC()); status != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("cost bound: %s", status), 1)
		}
		gateRaw, err := os.ReadFile(args[3])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		storedGate, err := corem10.ValidateCanaryGateDecision(gateRaw)
		if err != nil {
			return emit("REJECTED", nil, fmt.Errorf("invalid persisted canary gate"), 1)
		}
		if !resolveM10Artifact(args[1], corem10.ArtifactKindCanaryGate, gateRaw) {
			return emit("REJECTED", nil, fmt.Errorf("canary gate is not a registered canonical artifact"), 1)
		}
		if storedGate.EvaluatedAt != args[5] {
			return emit("REJECTED", nil, fmt.Errorf("authorization time must exactly match persisted gate evaluation"), 1)
		}
		currentGate, err := evaluateLearnerCanaryGate(s, bound, args[5])
		if err != nil || storedGate != currentGate {
			return emit("REJECTED", nil, fmt.Errorf("persisted canary gate is stale or does not match current state"), 1)
		}
		authorization, err := corem10.AuthorizeCanary(corem10.CanaryAuthorizationInput{Gate: storedGate, Grant: s.Canary.CanaryGrant, CostBound: bound, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, PolicyVersion: s.Policy.PolicyVersion, IdempotencyKey: s.Intent.IdempotencyKey, CorrelationID: s.Intent.CorrelationID, IntentExpiresAt: s.Intent.ExpiresAt, ExecutorID: args[6], AuthorizedAt: args[5]})
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		authorizationRaw, err := json.Marshal(authorization)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, err := corem10.ValidateExecutionAuthorization(authorizationRaw); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		if _, _, err := registerM10Artifact(args[1], corem10.ArtifactKindExecutionAuthorization, authorizationRaw); err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		status, err = writeNewJSON(args[4], authorization)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		if status == appendDuplicate {
			return emit(status, authorization, nil, 0)
		}
		return emit("AUTHORIZED", authorization, nil, 0)
	case "m10-reserve-authorization":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-reserve-authorization STATE_DIR AUTHORIZATION RESERVATION_ID"), 2)
		}
		if err := distinctPaths(args[1], args[2]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
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
		if err := missionCanaryActive(s, missionNowUTC()); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		authorizationRaw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		authorization, err := corem10.ValidateExecutionAuthorization(authorizationRaw)
		if err != nil || !resolveM10Artifact(args[1], corem10.ArtifactKindExecutionAuthorization, authorizationRaw) || !authorizationBindsMissionState(authorization, s) {
			return emit("REJECTED", nil, fmt.Errorf("execution authorization is invalid, unregistered or mismatched"), 1)
		}
		expiresAt, err := time.Parse(time.RFC3339, authorization.ExpiresAt)
		if err != nil || !expiresAt.After(missionNowUTC()) {
			return emit("REJECTED", nil, fmt.Errorf("execution authorization has expired"), 1)
		}
		bound, err := resolveAuthorizationCostBound(args[1], authorization)
		if err != nil || corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, missionNowUTC()) != "VALID" {
			return emit("REJECTED", nil, fmt.Errorf("authorization cost bound is invalid or expired"), 1)
		}
		reservationID := args[3]
		if reservationID == "" {
			return emit("REJECTED", nil, fmt.Errorf("reservation_id is required"), 1)
		}
		for _, prior := range s.Reservations {
			if prior.ReservationID == reservationID {
				if prior.AuthorizationID == authorization.AuthorizationID && prior.GrantID == authorization.CanaryGrantID && prior.IntentID == authorization.IntentID && prior.IntentHash == authorization.IntentHash && prior.CostMinor == authorization.CanaryCostBoundMinor && prior.CostBoundID == authorization.CanaryCostBoundID && prior.CostBoundHash == authorization.CanaryCostBoundHash && prior.ReservationMode == "GOVERNED_AUTHORIZATION" {
					return emit("EXACT_DUPLICATE", prior, nil, 0)
				}
				return emit("REJECTED", nil, fmt.Errorf("reservation_id reused with different authorization binding"), 1)
			}
			if prior.AuthorizationID == authorization.AuthorizationID {
				return emit("REJECTED", nil, fmt.Errorf("execution authorization already has a reservation"), 1)
			}
		}
		if s.Canary.Status != "ACTIVE" || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed >= s.Canary.MaxExecutionsTotal || authorization.CanaryCostBoundMinor > s.Canary.MaxCostMinorTotal-s.Canary.CostUsedMinor {
			return emit("BUDGET_DENIED", s.Canary, fmt.Errorf("canary budget exhausted"), 1)
		}
		reservation := LearnerReservation{ReservationID: reservationID, GrantID: authorization.CanaryGrantID, IntentID: authorization.IntentID, IntentHash: authorization.IntentHash, CostMinor: authorization.CanaryCostBoundMinor, CostBoundID: authorization.CanaryCostBoundID, CostBoundHash: authorization.CanaryCostBoundHash, AuthorizationID: authorization.AuthorizationID, ReservationMode: "GOVERNED_AUTHORIZATION", ReservedAt: missionNowUTC().Format(time.RFC3339Nano)}
		s.Canary.ExecutionsUsed++
		s.Canary.CostUsedMinor += authorization.CanaryCostBoundMinor
		s.Reservations = append(s.Reservations, reservation)
		if err := saveMissionState(args[1], s); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("RESERVED", reservation, nil, 0)
	case "m10-record-failed":
		if len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-record-failed STATE_DIR AUTHORIZATION OUT ATTEMPTED_AT FIXTURE_REASON"), 2)
		}
		if err := distinctPaths(args[1], args[2], args[3]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		authorizationRaw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		authorization, err := corem10.ValidateExecutionAuthorization(authorizationRaw)
		if err != nil || !resolveM10Artifact(args[1], corem10.ArtifactKindExecutionAuthorization, authorizationRaw) || !authorizationBindsMissionState(authorization, s) {
			return emit("REJECTED", nil, fmt.Errorf("execution authorization is invalid, unregistered or mismatched"), 1)
		}
		record, err := corem10.FailCanaryExecutionFixture(corem10.FailedExecutionInput{Authorization: authorization, AttemptedAt: args[4], Reason: args[5]})
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		reservationIndex, err := validateExecutionReservation(s, authorization, record)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		recordRaw, err := json.Marshal(record)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, err := corem10.ValidateExecutionRecord(recordRaw); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		if err := commitM10ExecutionRecord(args[1], s, reservationIndex, authorization, record); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		// The registry record and mission reservation are canonical runtime state;
		// the requested output is only a portable view. Bind the reservation
		// before attempting that view so an output collision cannot leave the
		// canonical execution orphaned from its budget reservation.
		status, err := writeNewJSON(args[3], record)
		if err != nil {
			return emit("CONFLICT", record, err, 1)
		}
		return emit(status, record, nil, 0)
	case "m10-cancel":
		if len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-cancel STATE_DIR AUTHORIZATION OUT ATTEMPTED_AT REASON"), 2)
		}
		if err := distinctPaths(args[1], args[2], args[3]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		authorizationRaw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		authorization, err := corem10.ValidateExecutionAuthorization(authorizationRaw)
		if err != nil {
			return emit("REJECTED", nil, fmt.Errorf("invalid execution authorization: %w", err), 1)
		}
		if !resolveM10Artifact(args[1], corem10.ArtifactKindExecutionAuthorization, authorizationRaw) {
			return emit("REJECTED", nil, fmt.Errorf("execution authorization is not a registered canonical artifact"), 1)
		}
		if !authorizationBindsMissionState(authorization, s) {
			return emit("REJECTED", nil, fmt.Errorf("execution authorization does not bind to current mission state"), 1)
		}
		record, err := corem10.CancelCanaryExecution(corem10.CancelledExecutionInput{Authorization: authorization, AttemptedAt: args[4], Reason: args[5]})
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		reservationIndex, err := validateExecutionReservation(s, authorization, record)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		recordRaw, err := json.Marshal(record)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if _, err := corem10.ValidateExecutionRecord(recordRaw); err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		if err := commitM10ExecutionRecord(args[1], s, reservationIndex, authorization, record); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		status, err := writeNewJSON(args[3], record)
		if err != nil {
			return emit("CONFLICT", record, err, 1)
		}
		return emit(status, record, nil, 0)
	case "m10-outcome":
		if len(args) != 3 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-outcome STATE_DIR OUTCOME_INPUT"), 2)
		}
		if err := distinctPaths(args[1], args[2]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		s, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		outcome, outcomeStatus := validateM10FixtureOutcome(args[1], s, raw)
		if outcomeStatus != "VALID" {
			return emit(outcomeStatus, nil, fmt.Errorf("M10 outcome rejected: %s", outcomeStatus), 1)
		}
		outcomes, err := loadM10FixtureOutcomes(args[1], s)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		for _, prior := range outcomes {
			if prior.OutcomeID != outcome.OutcomeID {
				if prior.EffectRef.EffectID == outcome.EffectRef.EffectID {
					return emit("CONFLICT", nil, fmt.Errorf("execution already has a fixture outcome"), 1)
				}
				continue
			}
			if reflect.DeepEqual(prior, outcome) {
				return emit("EXACT_DUPLICATE", outcome, nil, 0)
			}
			return emit("CONFLICT", nil, fmt.Errorf("outcome_id reused with different content"), 1)
		}
		encoded, err := json.Marshal(outcome)
		if err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		if err := (store.JSONL{}).AppendLine(m10OutcomeStorePath(args[1]), encoded); err != nil {
			return emit("STORE_ERROR", nil, err, 1)
		}
		return emit("APPENDED", outcome, nil, 0)
	case "m10-resolve":
		if len(args) != 4 && len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m10-resolve STATE_DIR KIND ARTIFACT_ID [CONTENT_HASH]"), 2)
		}
		// Do not resolve an artifact against a registry while a local writer can
		// still change its journal or append the corresponding lifecycle state.
		releaseGate, err := acquireRuntimeGate(args[1])
		if err != nil {
			if os.IsNotExist(err) {
				return emit("STATE_ERROR", nil, err, 1)
			}
			return emit("BUSY", nil, err, 1)
		}
		defer releaseGate()
		if _, err := os.Stat(m10ExecutionJournalPath(args[1])); err == nil {
			return emit("RECOVERY_REQUIRED", nil, fmt.Errorf("M10 execution journal requires a locked writer recovery"), 1)
		} else if !os.IsNotExist(err) {
			return emit("STATE_ERROR", nil, err, 1)
		}
		contentHash := ""
		if len(args) == 5 {
			contentHash = args[4]
		}
		entry, err := resolveM10ArtifactByID(args[1], args[2], args[3], contentHash)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		return emit("RESOLVED", entry.Artifact, nil, 0)
	case "m11-register":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-register STATE_DIR KIND ARTIFACT_INPUT"), 2)
		}
		if err := distinctPaths(args[1], args[3]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		managedKinds := map[string]bool{
			corem11.ArtifactKindActivation: true, corem11.ArtifactKindLedger: true,
			corem11.ArtifactKindGate: true, corem11.ArtifactKindAuthorization: true,
			corem11.ArtifactKindExecution: true, corem11.ArtifactKindEvaluation: true,
			corem11.ArtifactKindCycle: true, corem11.ArtifactKindRecoveryAdmission: true,
		}
		if managedKinds[args[2]] {
			return emit("REJECTED", nil, fmt.Errorf("M11 lifecycle artifact must be created by its dedicated command"), 1)
		}
		// A stopped runtime accepts only the human reconciliation artifact needed
		// to establish its read-only recovery handoff. No later lease/approval or
		// lifecycle record can be appended to blur the old STOP boundary.
		if state, err := loadMissionState(args[1]); err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		} else if state.Stop && args[2] != corem11.ArtifactKindReconciliation {
			return emit("STOPPED", nil, fmt.Errorf("durable STOP: %s", state.StopReason), 1)
		}
		raw, err := os.ReadFile(args[3])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		entry, status, err := registerM11Artifact(args[1], args[2], raw)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		return emit(status, entry, nil, 0)
	case "m11-resolve":
		if len(args) != 4 && len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-resolve STATE_DIR KIND ARTIFACT_ID [CONTENT_HASH]"), 2)
		}
		releaseGate, err := acquireRuntimeGate(args[1])
		if err != nil {
			if os.IsNotExist(err) {
				return emit("STATE_ERROR", nil, err, 1)
			}
			return emit("BUSY", nil, err, 1)
		}
		defer releaseGate()
		if err := m11JournalRecoveryRequired(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		hash := ""
		if len(args) == 5 {
			hash = args[4]
		}
		entry, err := resolveM11Artifact(args[1], args[2], args[3], hash)
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		return emit("RESOLVED", entry.Artifact, nil, 0)
	case "m11-activate":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-activate STATE_DIR LEASE_ID ACTIVATED_AT"), 2)
		}
		record, status, err := activateM11Lease(args[1], args[2], args[3])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, record, nil, 0)
	case "m11-ledger-init":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-ledger-init STATE_DIR LEASE_ID INITIALIZED_AT"), 2)
		}
		ledger, status, err := initializeM11Ledger(args[1], args[2], args[3])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, ledger, nil, 0)
	case "m11-gate":
		if len(args) != 7 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-gate STATE_DIR LEASE_ID HEALTH_ID COST_BOUND_ID LEDGER_ID EVALUATED_AT"), 2)
		}
		gate, _, err := evaluateM11Gate(args[1], args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(gate.Decision, gate, nil, 0)
	case "m11-authorize":
		if len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-authorize STATE_DIR LEASE_ID GATE_ID EXECUTOR_ID AUTHORIZED_AT"), 2)
		}
		authorization, status, err := authorizeM11Production(args[1], args[2], args[3], args[4], args[5])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, authorization, nil, 0)
	case "m11-reserve-authorization":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-reserve-authorization STATE_DIR AUTHORIZATION_ID LEDGER_ID RESERVED_AT"), 2)
		}
		ledger, status, err := reserveM11Authorization(args[1], args[2], args[3], args[4])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, ledger, nil, 0)
	case "m11-record-failed":
		if len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-record-failed STATE_DIR AUTHORIZATION_ID RESERVATION_LEDGER_ID ATTEMPTED_AT FIXTURE_REASON"), 2)
		}
		record, executionLedger, status, err := recordFailedM11Execution(args[1], args[2], args[3], args[4], args[5])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, map[string]any{"execution": record, "execution_ledger": executionLedger}, nil, 0)
	case "m11-record-unknown":
		if len(args) != 6 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-record-unknown STATE_DIR AUTHORIZATION_ID RESERVATION_LEDGER_ID ATTEMPTED_AT REASON"), 2)
		}
		record, stoppedLedger, status, err := recordUnknownM11Execution(args[1], args[2], args[3], args[4], args[5])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, map[string]any{"execution": record, "stopped_ledger": stoppedLedger}, nil, 0)
	case "m11-reconcile":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-reconcile STATE_DIR RESOLUTION_ID STOPPED_LEDGER_ID"), 2)
		}
		resolution, ledger, status, err := reconcileM11Execution(args[1], args[2], args[3])
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		return emit(status, map[string]any{"resolution": resolution, "stopped_ledger": ledger}, nil, 0)
	case "m11-recovery-export":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-recovery-export STATE_DIR RESOLUTION_ID STOPPED_LEDGER_ID OUT"), 2)
		}
		if err := distinctPaths(args[1], args[4]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		// A recovery handoff must describe one coherent stopped-runtime
		// snapshot. Treat its read as an exclusive local operation so a writer
		// cannot append a journal or lifecycle artifact between the recovery
		// preflight and the handoff resolution.
		releaseGate, err := acquireRuntimeGate(args[1])
		if err != nil {
			return emit("BUSY", nil, err, 1)
		}
		defer releaseGate()
		if err := m11JournalRecoveryRequired(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		handoff, err := m11RecoveryHandoff(args[1], args[2], args[3])
		if err != nil {
			return emit("REJECTED", nil, err, 1)
		}
		status, err := writeNewJSON(args[4], handoff)
		if err != nil {
			return emit("CONFLICT", nil, err, 1)
		}
		return emit(status, handoff, nil, 0)
	case "m11-recovery-admit":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-recovery-admit NEW_STATE_DIR OLD_STATE_DIR HANDOFF_INPUT ADMISSION_INPUT"), 2)
		}
		if err := distinctPaths(args[1], args[2], args[3], args[4]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		// Do not even read the supplied handoff/admission input while the old
		// lifecycle is mid-recovery. The new runtime must receive no admission
		// derived from a partial old-runtime transition.
		if err := m11JournalRecoveryRequired(args[2]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
		}
		state, err := loadMissionState(args[1])
		if err != nil {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if state.Stop {
			return emit("STOPPED", nil, fmt.Errorf("durable STOP: %s", state.StopReason), 1)
		}
		raw, err := os.ReadFile(args[4])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		entry, status, err := admitM11Recovery(args[1], args[2], args[3], raw)
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, entry, nil, 0)
	case "m11-outcome":
		if len(args) != 4 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-outcome STATE_DIR OUTCOME_INPUT LEDGER_ID"), 2)
		}
		if err := distinctPaths(args[1], args[2]); err != nil {
			return emit("PATH_ERROR", nil, err, 1)
		}
		raw, err := os.ReadFile(args[2])
		if err != nil {
			return emit("INPUT_ERROR", nil, err, 1)
		}
		outcome, ledger, status, err := recordM11FixtureOutcome(args[1], args[3], raw)
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, map[string]any{"outcome": outcome, "post_ledger": ledger}, nil, 0)
	case "m11-evaluate":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-evaluate STATE_DIR OUTCOME_ID EVALUATION_ID EVALUATED_AT"), 2)
		}
		evaluation, status, err := evaluateM11FixtureOutcome(args[1], args[2], args[3], args[4])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, evaluation, nil, 0)
	case "m11-close-cycle":
		if len(args) != 5 {
			return emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-close-cycle STATE_DIR CYCLE_ID EVALUATION_ID CLOSED_AT"), 2)
		}
		cycle, status, err := closeM11FixtureCycle(args[1], args[2], args[3], args[4])
		if err != nil {
			return emit(missionErrorStatus(err), nil, err, 1)
		}
		return emit(status, cycle, nil, 0)
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
		if err := missionCanaryActive(s, missionNowUTC()); err != nil {
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
			if status := corem10.ValidFor(bound, s.Intent.IntentID, s.Intent.IntentHash, s.Intent.CorrelationID, s.Canary.Currency, missionNowUTC()); status != "VALID" {
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
			if prior.AuthorizationID == "" && prior.GrantID == s.Canary.GrantID && prior.IntentID == s.Intent.IntentID && prior.IntentHash == s.Intent.IntentHash && prior.CostMinor == cost {
				return emit("EXACT_DUPLICATE", s.Canary, nil, 0)
			}
			return emit("REJECTED", nil, fmt.Errorf("reservation_id reused with different binding or cost"), 1)
		}
		if s.Canary.Status != "ACTIVE" || s.Canary.ExecutionsUsed < 0 || s.Canary.CostUsedMinor < 0 || s.Canary.ExecutionsUsed >= s.Canary.MaxExecutionsTotal || cost > s.Canary.MaxCostMinorTotal-s.Canary.CostUsedMinor {
			return emit("BUDGET_DENIED", s.Canary, fmt.Errorf("canary budget exhausted"), 1)
		}
		s.Canary.ExecutionsUsed++
		s.Canary.CostUsedMinor += cost
		s.Reservations = append(s.Reservations, LearnerReservation{ReservationID: reservationID, GrantID: s.Canary.GrantID, IntentID: s.Intent.IntentID, IntentHash: s.Intent.IntentHash, CostMinor: cost, CostBoundID: boundID, CostBoundHash: boundHash, ReservationMode: "LEGACY_COMPAT", ReservedAt: missionNowUTC().Format(time.RFC3339Nano)})
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
		releaseGate, err := acquireRuntimeGate(args[1])
		if err != nil {
			if os.IsNotExist(err) {
				return emit("STATE_ERROR", nil, err, 1)
			}
			return emit("BUSY", nil, err, 1)
		}
		defer releaseGate()
		if _, err := os.Stat(m10ExecutionJournalPath(args[1])); err == nil {
			return emit("RECOVERY_REQUIRED", nil, fmt.Errorf("M10 execution journal requires a locked writer recovery"), 1)
		} else if !os.IsNotExist(err) {
			return emit("STATE_ERROR", nil, err, 1)
		}
		if err := m11JournalRecoveryRequired(args[1]); err != nil {
			return emit("RECOVERY_REQUIRED", nil, err, 1)
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
