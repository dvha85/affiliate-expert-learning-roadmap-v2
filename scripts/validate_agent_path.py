from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []

required = [
    "curriculum/O00/O00.1-safe-system-walkthrough.md",
    "missions/O00-safe-synthetic-walkthrough.md",
    "lab/mission-runtime/go.mod",
    "lab/mission-runtime/cmd/demo/main.go",
    "lab/mission-runtime/cmd/demo/m03_m05.go",
    "lab/mission-runtime/cmd/demo/m06_m07.go",
    "lab/mission-runtime/cmd/demo/m08.go",
    "lab/mission-runtime/cmd/demo/m08_compat.go",
    "lab/mission-runtime/cmd/demo/m08_test.go",
    "lab/mission-runtime/cmd/demo/m09.go",
    "lab/mission-runtime/cmd/demo/m09_test.go",
    "lab/mission-runtime/cmd/demo/m10.go",
    "lab/mission-runtime/cmd/demo/m10_test.go",
    "lab/mission-runtime/cmd/demo/trusted_cost_bound.go",
    "lab/mission-runtime/cmd/demo/mission_runtime_test.go",
    "lab/n8n/M06-readonly-watcher.blueprint.json",
    "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    "contracts/action-record.schema.json",
    "contracts/effect-ref.schema.json",
    "contracts/outcome-record.schema.json",
    "contracts/advisor-output.schema.json",
    "contracts/evaluation-record.schema.json",
    "contracts/improvement-proposal.schema.json",
    "contracts/review-record.schema.json",
    "contracts/tool-registry.schema.json",
    "contracts/action-intent.schema.json",
    "contracts/policy-decision.schema.json",
    "contracts/approval-record.schema.json",
    "contracts/execution-authorization.schema.json",
    "contracts/execution-record.schema.json",
    "contracts/canary-grant.schema.json",
    "contracts/trusted-cost-bound.schema.json",
    "contracts/canary-ledger.schema.json",
    "contracts/canary-gate-decision.schema.json",
]
starter_names = {
    "M03": "tracked-human-action", "M04": "grounded-ai-advisor", "M05": "reviewed-improvement",
    "M06": "readonly-watcher", "M07": "readonly-evidence-agent", "M08": "shadow-policy",
    "M09": "approval-execution", "M10": "governed-canary",
}
for mission in range(3, 11):
    mid = f"M{mission:02d}"
    required.extend([
        f"starter-kits/{mid}-{starter_names[mid]}/CHECKPOINTS.md",
        f"starter-kits/{mid}-{starter_names[mid]}/{mid}-OPERATED-EVIDENCE-TEMPLATE.md",
    ])
for rel in required:
    if not (ROOT / rel).exists():
        errors.append(f"missing required file: {rel}")
for mission in range(3, 11):
    mid = f"M{mission:02d}"
    if not (ROOT / "curriculum" / mid).is_dir():
        errors.append(f"missing curriculum directory: {mid}")
    if not any((ROOT / "missions").glob(f"{mid}-*.md")):
        errors.append(f"missing mission contract: {mid}")
    if not any((ROOT / "starter-kits").glob(f"{mid}-*/README.md")):
        errors.append(f"missing starter README: {mid}")
    if not any((ROOT / "evals").glob(f"{mid}-*/cases.json")):
        errors.append(f"missing executable eval pack: {mid}")

json_files = [
    "contracts/action-record.schema.json", "contracts/effect-ref.schema.json", "contracts/outcome-record.schema.json",
    "contracts/advisor-output.schema.json", "contracts/evaluation-record.schema.json",
    "contracts/improvement-proposal.schema.json", "contracts/review-record.schema.json",
    "contracts/tool-registry.schema.json", "contracts/action-intent.schema.json", "contracts/policy-decision.schema.json",
    "contracts/approval-record.schema.json", "contracts/execution-authorization.schema.json", "contracts/execution-record.schema.json",
    "contracts/canary-grant.schema.json", "contracts/trusted-cost-bound.schema.json",
    "contracts/canary-ledger.schema.json", "contracts/canary-gate-decision.schema.json",
    "lab/n8n/M06-readonly-watcher.blueprint.json", "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    "evals/M03-tracked-human-action/cases.json", "evals/M05-reviewed-improvement/cases.json",
    "evals/M08-shadow-policy/cases.json", "evals/M09-approval-execution/cases.json", "evals/M10-governed-canary/cases.json",
]
for rel in json_files:
    try:
        json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {rel}: {exc}")

mission_index = (ROOT / "missions/README.md").read_text(encoding="utf-8")
for mid in ["M03", "M04", "M05", "M06", "M07", "M08", "M09", "M10"]:
    line = next((line for line in mission_index.splitlines() if line.startswith(f"| {mid} |")), "")
    if "| ready |" not in line:
        errors.append(f"{mid} must be ready in mission index")

runtime = "\n".join(
    (ROOT / "lab/mission-runtime/cmd/demo" / name).read_text(encoding="utf-8")
    for name in ["m03_m05.go", "m06_m07.go", "m08.go", "m08_compat.go", "m09.go", "m10.go", "trusted_cost_bound.go"]
)
for marker in [
    "DRY_RUN_ONLY", "BROKEN_LINK", "REJECT_MACHINE_EXECUTION", "REJECT_WRITE_REQUEST", "ABSTAIN_FUTURE", "REJECT_AUTO_APPLY",
    "EffectRef", "HUMAN_ACTION", "MACHINE_EXECUTION",
    "REJECT_WRITE_METHOD", "REJECT_TOOL", "REJECT_UNGROUNDED", "NormalizeWatchObservation", "ValidateEvaluationRecord", "ValidateReviewRecord",
    "TAMPERED_INTENT", "EXPIRED_INTENT", "IDEMPOTENCY_COLLISION", "POLICY_UNAVAILABLE", "execution_authorized", "SHADOW_POLICY_ALLOW",
    "PROPOSAL_ONLY", "NON_AUTHORIZING", "INTENT_AUTHORITY_FORBIDDEN", "PolicyReviewRequired",
    "KnownProposalIDs", "UNKNOWN_ACTION_POLICY", "WAIT_APPROVAL", "DENY_INVALID_APPROVER", "DENY_APPROVAL_MISMATCH", "DENY_APPROVAL_BEFORE_POLICY", "DENY_POLICY_REVALIDATION", "DENY_KILL_SWITCH",
    "WAIT_ALREADY_EXECUTED", "WAIT_APPROVAL_CONSUMED", "WAIT_RECONCILIATION", "PersistM09State", "LoadM09State", "ExecuteLocalSandbox", "AllowedExecutorIDs", "APPROVED_LIVE",
    "CanaryGrant", "TrustedCostBound", "CanaryLedger", "CanaryGateDecision", "GOVERNED_CANARY", "RISK2_PER_ACTION_APPROVAL_REQUIRED", "OUTCOME_BACKPRESSURE", "GRANT_REVOKED", "TAMPERED_GRANT", "ExecuteCanaryLocalSandbox", "RecordCanaryOutcome",
]:
    if marker not in runtime:
        errors.append(f"runtime safety/integration marker missing: {marker}")

# O00 functional artifact chain.
o00_eval = json.loads((ROOT / "evals/O00-safe-walkthrough/cases.json").read_text(encoding="utf-8"))
required_kinds = {"Observation", "ActionIntent", "PolicyDecision", "ExecutionRecord", "EvaluationRecord"}
if not required_kinds.issubset({c.get("required_kind") for c in o00_eval}):
    errors.append("O00 eval must require the functional artifact chain")

# M03/M05 stage-neutral effect semantics.
effect_schema = (ROOT / "contracts/effect-ref.schema.json").read_text(encoding="utf-8")
outcome_schema = (ROOT / "contracts/outcome-record.schema.json").read_text(encoding="utf-8")
evaluation_schema = (ROOT / "contracts/evaluation-record.schema.json").read_text(encoding="utf-8")
for marker in ["HUMAN_ACTION", "MACHINE_EXECUTION", "effect_id"]:
    if marker not in effect_schema:
        errors.append(f"EffectRef contract missing: {marker}")
for rel, text in [("OutcomeRecord", outcome_schema), ("EvaluationRecord", evaluation_schema)]:
    if '"effect_ref"' not in text or "effect-ref.schema.json" not in text:
        errors.append(f"{rel} must use canonical EffectRef")

# M06/M07 concrete workflow guards.
m06 = json.loads((ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json").read_text(encoding="utf-8"))
m06_nodes = {n.get("name"): n for n in m06.get("nodes", [])}
for name in ["Schedule Trigger", "Allowed Source", "Read-only HTTP GET", "Normalize + Change Detect", "Append Canonical History", "Report NEW UNCHANGED CHANGED"]:
    if name not in m06_nodes:
        errors.append(f"M06 blueprint missing node: {name}")
if m06_nodes.get("Read-only HTTP GET", {}).get("parameters", {}).get("method") != "GET":
    errors.append("M06 n8n HTTP node must be GET-only")
normalize_code = m06_nodes.get("Normalize + Change Detect", {}).get("parameters", {}).get("jsCode", "")
for marker in ["$getWorkflowStaticData", "observation_id", "subject_id", "source_url", "observed_at", "evidence_kind", "claim_kind", "limitation", "change_state"]:
    if marker not in normalize_code:
        errors.append(f"M06 normalize/change-detect missing: {marker}")
m07 = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
types = {n.get("type") for n in m07.get("nodes", [])}
for node_type in ["@n8n/n8n-nodes-langchain.agent", "@n8n/n8n-nodes-langchain.lmChatOpenAi", "n8n-nodes-base.httpRequestTool"]:
    if node_type not in types:
        errors.append(f"M07 blueprint missing real Agent component: {node_type}")
tool = next((n for n in m07.get("nodes", []) if n.get("type") == "n8n-nodes-base.httpRequestTool"), {})
if tool.get("parameters", {}).get("method") != "GET":
    errors.append("M07 connected HTTP tool must be GET-only")

# M08 proposal-only policy semantics.
action_schema = (ROOT / "contracts/action-intent.schema.json").read_text(encoding="utf-8")
policy_schema = (ROOT / "contracts/policy-decision.schema.json").read_text(encoding="utf-8")
for marker in ["intent_hash", "idempotency_key", "correlation_id", "intent_mode", "PROPOSAL_ONLY", "execution_authorized", "evidence_ids"]:
    if marker not in action_schema:
        errors.append(f"M08 ActionIntent contract missing: {marker}")
for obsolete in ["shadow_only", "dry_run"]:
    if obsolete in action_schema:
        errors.append(f"M08 ActionIntent obsolete contract field remains: {obsolete}")
for marker in ["intent_hash", "policy_version", "policy_review_required", "policy_mode", "NON_AUTHORIZING", "execution_authorized", "HUMAN_REVIEW"]:
    if marker not in policy_schema:
        errors.append(f"M08 PolicyDecision contract missing: {marker}")
for obsolete in ["approval_required", "shadow_only"]:
    if obsolete in policy_schema:
        errors.append(f"M08 PolicyDecision obsolete contract field remains: {obsolete}")
m08_cases = json.loads((ROOT / "evals/M08-shadow-policy/cases.json").read_text(encoding="utf-8"))
for reason in [
    "SHADOW_POLICY_ALLOW", "TAMPERED_INTENT", "INTENT_AUTHORITY_FORBIDDEN", "EXPIRED_INTENT",
    "MISSING_DECISION_LINK", "MISSING_EVIDENCE_LINK", "MISSING_PROPOSAL_LINK", "DUPLICATE_INTENT",
    "IDEMPOTENCY_COLLISION", "POLICY_UNAVAILABLE", "UNKNOWN_ACTION_POLICY",
]:
    if not any(c.get("expected_reason") == reason for c in m08_cases):
        errors.append(f"M08 eval missing failure/boundary case: {reason}")

# M09 approval/execution semantics.
approval = (ROOT / "contracts/approval-record.schema.json").read_text(encoding="utf-8")
auth = (ROOT / "contracts/execution-authorization.schema.json").read_text(encoding="utf-8")
execution = (ROOT / "contracts/execution-record.schema.json").read_text(encoding="utf-8")
for marker in ["approved_by", "human", "approver_id", "intent_hash", "policy_version", "one_time", "expires_at"]:
    if marker not in approval:
        errors.append(f"M09 ApprovalRecord missing: {marker}")
for marker in ["approval_id", "executor_id", "execution_authorized", "idempotency_key", "intent_hash", "execution_mode", "APPROVED_LIVE"]:
    if marker not in auth:
        errors.append(f"M09 ExecutionAuthorization missing: {marker}")
for marker in ["authorization_id", "approval_id", "idempotency_key", "side_effect_state", "RECONCILIATION_REQUIRED", "UNKNOWN", "status", "intent_hash"]:
    if marker not in execution:
        errors.append(f"M09 ExecutionRecord missing: {marker}")
m09_cases = json.loads((ROOT / "evals/M09-approval-execution/cases.json").read_text(encoding="utf-8"))
for expected in [
    "AUTHORIZED", "WAIT_APPROVAL", "DENY_REJECTED", "DENY_INVALID_APPROVER", "DENY_APPROVAL_MISMATCH",
    "DENY_APPROVAL_BEFORE_POLICY", "DENY_EXPIRED_APPROVAL", "DENY_KILL_SWITCH", "DENY_EXECUTOR",
    "DENY_POLICY_STATE", "WAIT_ALREADY_EXECUTED", "DENY_TAMPERED_INTENT", "DENY_EXPIRED_INTENT",
]:
    if not any(c.get("expected") == expected for c in m09_cases):
        errors.append(f"M09 eval missing gate case: {expected}")

# M10 governed canary semantics.
grant_schema = (ROOT / "contracts/canary-grant.schema.json").read_text(encoding="utf-8")
cost_schema = (ROOT / "contracts/trusted-cost-bound.schema.json").read_text(encoding="utf-8")
ledger_schema = (ROOT / "contracts/canary-ledger.schema.json").read_text(encoding="utf-8")
gate_schema = (ROOT / "contracts/canary-gate-decision.schema.json").read_text(encoding="utf-8")
for marker in ["approval_ref", "approved_by", "human", "grant_hash", "hash_version", "allowed_risk_classes", "max_executions_total", "max_executions_per_window", "max_cost_minor_total", "max_pending_outcomes", "kill_switch_required"]:
    if marker not in grant_schema:
        errors.append(f"M10 CanaryGrant missing: {marker}")
if "RISK2" in json.loads((ROOT / "contracts/canary-grant.schema.json").read_text(encoding="utf-8"))["properties"]["allowed_risk_classes"]["items"]["enum"]:
    errors.append("M10 CanaryGrant must never delegate RISK2")
for marker in ["cost_bound_id", "intent_hash", "max_cost_minor", "source_ref", "cost_bound_hash"]:
    if marker not in cost_schema:
        errors.append(f"TrustedCostBound missing: {marker}")
for marker in ["executions_total", "executions_in_window", "cost_minor_total", "pending_outcomes", "successful_idempotency_keys", "reconciliation_required"]:
    if marker not in ledger_schema:
        errors.append(f"M10 CanaryLedger missing: {marker}")
for marker in ["ALLOW_CANARY", "REQUIRE_APPROVAL", "WAIT", "DENY", "per_action_approval_required", "execution_authorized"]:
    if marker not in gate_schema:
        errors.append(f"M10 CanaryGateDecision missing: {marker}")
for marker in ["GOVERNED_CANARY", "canary_grant_id", "canary_grant_version", "canary_gate_id"]:
    if marker not in auth:
        errors.append(f"M10 ExecutionAuthorization missing: {marker}")
for marker in ["canary_grant_id", "canary_grant_version", "canary_gate_id"]:
    if marker not in execution:
        errors.append(f"M10 ExecutionRecord missing: {marker}")
m10_cases = json.loads((ROOT / "evals/M10-governed-canary/cases.json").read_text(encoding="utf-8"))
for decision, reason in [
    ("ALLOW_CANARY", "CANARY_ELIGIBLE"),
    ("REQUIRE_APPROVAL", "RISK2_PER_ACTION_APPROVAL_REQUIRED"),
    ("REQUIRE_APPROVAL", "RISK_NOT_DELEGATED"),
    ("REQUIRE_APPROVAL", "SCOPE_NOT_DELEGATED"),
    ("DENY", "GRANT_REVOKED"),
    ("DENY", "KILL_SWITCH_ACTIVE"),
    ("DENY", "TAMPERED_GRANT"),
    ("REQUIRE_APPROVAL", "CANARY_TOTAL_BUDGET_EXHAUSTED"),
    ("WAIT", "RATE_LIMIT_REACHED"),
    ("REQUIRE_APPROVAL", "CANARY_COST_BUDGET_EXHAUSTED"),
    ("WAIT", "OUTCOME_BACKPRESSURE"),
    ("WAIT", "RECONCILIATION_REQUIRED"),
]:
    if not any(c.get("expected_decision") == decision and c.get("expected_reason") == reason for c in m10_cases):
        errors.append(f"M10 eval missing gate case: {decision}/{reason}")

if errors:
    print("AGENT PATH VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)
print("AGENT PATH VALIDATION PASS: O00 and M03-M10 are learner-operable, linked and authority-bounded")
