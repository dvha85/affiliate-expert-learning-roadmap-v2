package m08

// PolicyConformanceCase is a deterministic M08 policy scenario shared by the
// canonical core tests and both runtime adapters. LearnerEquivalent is false
// when the learner's intentionally smaller policy input cannot express the
// scenario's trusted decision/evidence registry state.
type PolicyConformanceCase struct {
	Name             string
	Intent           Intent
	Context          PolicyContext
	Decision         string
	RiskClass        string
	Reason           string
	PolicyReview     bool
	ExecutionAuth    bool
	LearnerEquivalent bool
}

// PolicyConformanceCases returns fresh fixtures for the shared M08 decision
// table. Each caller may safely mutate its returned values while preserving
// the same expected decision/reason contract for the other adapters.
func PolicyConformanceCases() []PolicyConformanceCase {
	baseIntent := func(proposedBy, proposalRef string) Intent {
		return SealIntent(Intent{
			IntentID:       "conf-intent",
			DecisionID:     "conf-decision",
			EvidenceIDs:    []string{"conf-evidence"},
			ActionType:     "DRAFT",
			Target:         "https://example.com/draft",
			Parameters:     map[string]any{"title": "draft-only"},
			ProposedBy:     proposedBy,
			ProposalRef:    proposalRef,
			CreatedAt:      "2026-09-03T01:00:00Z",
			ExpiresAt:      "2026-09-03T03:00:00Z",
			CorrelationID:  "conf-correlation",
			IdempotencyKey: "conf-idempotency",
		})
	}
	baseContext := func() PolicyContext {
		return PolicyContext{
			PolicyVersion:    "m08-conformance-v1",
			Now:              "2026-09-03T02:00:00Z",
			KnownDecisionIDs: []string{"conf-decision"},
			KnownEvidenceIDs: []string{"conf-evidence"},
			KnownProposalIDs: []string{},
			AllowedHosts:     []string{"example.com"},
			ActionRisk:       map[string]string{"DRAFT": "RISK0"},
			SeenIdempotency:  map[string]string{},
		}
	}
	caseFor := func(name string, intent Intent, context PolicyContext, decision, risk, reason string, review, learnerEquivalent bool) PolicyConformanceCase {
		return PolicyConformanceCase{Name: name, Intent: intent, Context: context, Decision: decision, RiskClass: risk, Reason: reason, PolicyReview: review, ExecutionAuth: false, LearnerEquivalent: learnerEquivalent}
	}

	allow := baseIntent("human", "")
	allowContext := baseContext()
	risk1Context := baseContext()
	risk1Context.ActionRisk["DRAFT"] = "RISK1"
	risk2Context := baseContext()
	risk2Context.ActionRisk["DRAFT"] = "RISK2"
	future := baseIntent("human", "")
	future.CreatedAt = "2026-09-03T02:30:00Z"
	future.IntentHash = ComputeIntentHash(future)
	expiredContext := baseContext()
	expiredContext.Now = "2026-09-03T03:00:00Z"
	missingDecisionContext := baseContext()
	missingDecisionContext.KnownDecisionIDs = []string{}
	missingEvidenceContext := baseContext()
	missingEvidenceContext.KnownEvidenceIDs = []string{}
	agentMissingProposal := baseIntent("agent", "conf-proposal")
	targetDenied := baseIntent("human", "")
	targetDenied.Target = "https://outside.example/draft"
	targetDenied.IntentHash = ComputeIntentHash(targetDenied)
	duplicateContext := baseContext()
	duplicateContext.SeenIdempotency[allow.IdempotencyKey] = allow.IntentHash
	collisionContext := baseContext()
	collisionContext.SeenIdempotency[allow.IdempotencyKey] = "sha256:other"
	tampered := baseIntent("human", "")
	tampered.Target = "https://example.com/changed"

	return []PolicyConformanceCase{
		caseFor("allow-risk0", allow, allowContext, "ALLOW", "RISK0", "SHADOW_POLICY_ALLOW", false, true),
		caseFor("human-review-risk1", baseIntent("human", ""), risk1Context, "HUMAN_REVIEW", "RISK1", "RISK1_REQUIRES_REVIEW", true, true),
		caseFor("human-review-risk2", baseIntent("human", ""), risk2Context, "HUMAN_REVIEW", "RISK2", "RISK2_REQUIRES_REVIEW", true, true),
		caseFor("future-intent", future, baseContext(), "WAIT", "RISK0", "INTENT_NOT_YET_VALID", false, true),
		caseFor("expired-intent", allow, expiredContext, "DENY", "RISK0", "EXPIRED_INTENT", false, true),
		caseFor("missing-decision-link", allow, missingDecisionContext, "GET_MORE_DATA", "RISK0", "MISSING_DECISION_LINK", false, false),
		caseFor("missing-evidence-link", allow, missingEvidenceContext, "GET_MORE_DATA", "RISK0", "MISSING_EVIDENCE_LINK", false, false),
		caseFor("missing-agent-proposal-link", agentMissingProposal, baseContext(), "GET_MORE_DATA", "RISK0", "MISSING_PROPOSAL_LINK", false, true),
		caseFor("target-not-allowed", targetDenied, baseContext(), "DENY", "RISK0", "TARGET_NOT_ALLOWED", false, true),
		caseFor("duplicate-intent", allow, duplicateContext, "WAIT", "RISK0", "DUPLICATE_INTENT", false, true),
		caseFor("idempotency-collision", allow, collisionContext, "DENY", "RISK0", "IDEMPOTENCY_COLLISION", false, true),
		caseFor("tampered-intent", tampered, baseContext(), "DENY", "RISK2", "TAMPERED_INTENT", false, true),
	}
}
