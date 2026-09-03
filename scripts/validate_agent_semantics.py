from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []


def load_json(rel):
    try:
        return json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {rel}: {exc}")
        return {}


manifest = load_json("missions/manifest.json")
missions = {item.get("id"): item for item in manifest.get("missions", []) if isinstance(item, dict)}

semantic_files = [
    "contracts/action-record.schema.json", "contracts/effect-ref.schema.json", "contracts/outcome-record.schema.json",
    "contracts/advisor-output.schema.json", "contracts/evaluation-record.schema.json",
    "contracts/improvement-proposal.schema.json", "contracts/review-record.schema.json",
    "contracts/tool-registry.schema.json", "contracts/action-intent.schema.json", "contracts/policy-decision.schema.json",
    "contracts/approval-record.schema.json", "contracts/execution-authorization.schema.json", "contracts/execution-record.schema.json",
    "contracts/canary-grant.schema.json", "contracts/trusted-cost-bound.schema.json",
    "contracts/canary-ledger.schema.json", "contracts/canary-gate-decision.schema.json",
    "lab/n8n/M06-readonly-watcher.blueprint.json", "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
]
for rel in semantic_files:
    if not (ROOT / rel).exists():
        errors.append(f"missing semantic asset: {rel}")
    elif rel.endswith(".json"):
        load_json(rel)

for mid in ["M03", "M05", "M08", "M09", "M10"]:
    rel = missions.get(mid, {}).get("eval")
    if not rel:
        errors.append(f"{mid} eval path must come from missions/manifest.json")
    elif not (ROOT / rel).exists():
        errors.append(f"{mid} manifest eval missing: {rel}")

runtime = "\n".join(
    (ROOT / "lab/mission-runtime/cmd/demo" / name).read_text(encoding="utf-8")
    for name in ["m03_m05.go", "m06_m07.go", "m08.go", "m08_compat.go", "m09.go", "m10.go", "trusted_cost_bound.go"]
)
for marker in [
    "DRY_RUN_ONLY", "BROKEN_LINK", "REJECT_MACHINE_EXECUTION", "REJECT_WRITE_REQUEST", "ABSTAIN_FUTURE", "REJECT_AUTO_APPLY",
    "EffectRef", "HUMAN_ACTION", "MACHINE_EXECUTION", "REJECT_WRITE_METHOD", "REJECT_TOOL", "REJECT_UNGROUNDED",
    "NormalizeWatchObservation", "ValidateEvaluationRecord", "ValidateReviewRecord", "TAMPERED_INTENT", "EXPIRED_INTENT",
    "IDEMPOTENCY_COLLISION", "POLICY_UNAVAILABLE", "execution_authorized", "SHADOW_POLICY_ALLOW", "PROPOSAL_ONLY",
    "NON_AUTHORIZING", "INTENT_AUTHORITY_FORBIDDEN", "PolicyReviewRequired", "KnownProposalIDs", "UNKNOWN_ACTION_POLICY",
    "WAIT_APPROVAL", "DENY_INVALID_APPROVER", "DENY_APPROVAL_MISMATCH", "DENY_APPROVAL_BEFORE_POLICY",
    "DENY_POLICY_REVALIDATION", "DENY_KILL_SWITCH", "WAIT_ALREADY_EXECUTED", "WAIT_APPROVAL_CONSUMED",
    "WAIT_RECONCILIATION", "PersistM09State", "LoadM09State", "ExecuteLocalSandbox", "AllowedExecutorIDs", "APPROVED_LIVE",
    "CanaryGrant", "TrustedCostBound", "CanaryLedger", "CanaryGateDecision", "GOVERNED_CANARY",
    "RISK2_PER_ACTION_APPROVAL_REQUIRED", "OUTCOME_BACKPRESSURE", "GRANT_REVOKED", "TAMPERED_GRANT",
    "ExecuteCanaryLocalSandbox", "RecordCanaryOutcome",
]:
    if marker not in runtime:
        errors.append(f"runtime semantic marker missing: {marker}")

# M03/M05 stage-neutral effect semantics.
effect = load_json("contracts/effect-ref.schema.json")
if effect.get("properties", {}).get("effect_kind", {}).get("enum") != ["HUMAN_ACTION", "MACHINE_EXECUTION"]:
    errors.append("EffectRef must distinguish HUMAN_ACTION and MACHINE_EXECUTION")
for rel in ["contracts/outcome-record.schema.json", "contracts/evaluation-record.schema.json"]:
    schema = load_json(rel)
    if schema.get("properties", {}).get("effect_ref", {}).get("$ref") != "effect-ref.schema.json":
        errors.append(f"{rel} must reference canonical EffectRef")

# M06 watcher: exact canonical snapshot determines change state; fingerprint is diagnostic only.
m06 = load_json(missions.get("M06", {}).get("orchestration_blueprint", "lab/n8n/M06-readonly-watcher.blueprint.json"))
nodes = {node.get("name"): node for node in m06.get("nodes", [])}
for name in ["Schedule Trigger", "Allowed Source", "Read-only HTTP GET", "Normalize + Change Detect", "Canonical History Handoff", "Report NEW UNCHANGED CHANGED"]:
    if name not in nodes:
        errors.append(f"M06 blueprint missing node: {name}")
if nodes.get("Read-only HTTP GET", {}).get("parameters", {}).get("method") != "GET":
    errors.append("M06 n8n HTTP node must be GET-only")
normalize = nodes.get("Normalize + Change Detect", {}).get("parameters", {}).get("jsCode", "")
for marker in ["stableCanonical", "previous.canonical===canonical", "MAX_CACHE_CHARS", "watcher_cache", "content_fingerprint", "observation_id"]:
    if marker not in normalize:
        errors.append(f"M06 exact change-detection marker missing: {marker}")
for obsolete in ["let h=5381", "previous===fingerprint"]:
    if obsolete in normalize:
        errors.append(f"M06 must not decide change state from collision-prone fingerprint: {obsolete}")

# M07 read-only Agent boundary.
m07 = load_json(missions.get("M07", {}).get("orchestration_blueprint", "lab/n8n/M07-readonly-evidence-agent.blueprint.json"))
types = {node.get("type") for node in m07.get("nodes", [])}
for node_type in ["@n8n/n8n-nodes-langchain.agent", "@n8n/n8n-nodes-langchain.lmChatOpenAi", "n8n-nodes-base.httpRequestTool"]:
    if node_type not in types:
        errors.append(f"M07 blueprint missing Agent component: {node_type}")
tool = next((node for node in m07.get("nodes", []) if node.get("type") == "n8n-nodes-base.httpRequestTool"), {})
if tool.get("parameters", {}).get("method") != "GET":
    errors.append("M07 connected HTTP tool must be GET-only")

# M08 proposal-only policy semantics.
intent = load_json("contracts/action-intent.schema.json")
policy = load_json("contracts/policy-decision.schema.json")
if intent.get("properties", {}).get("intent_mode", {}).get("const") != "PROPOSAL_ONLY":
    errors.append("ActionIntent must remain PROPOSAL_ONLY")
if intent.get("properties", {}).get("execution_authorized", {}).get("const") is not False:
    errors.append("ActionIntent must never authorize execution")
if policy.get("properties", {}).get("policy_mode", {}).get("const") != "NON_AUTHORIZING":
    errors.append("PolicyDecision must remain NON_AUTHORIZING")
if policy.get("properties", {}).get("execution_authorized", {}).get("const") is not False:
    errors.append("PolicyDecision must never authorize execution")

m08_cases = load_json(missions.get("M08", {}).get("eval", ""))
for reason in ["SHADOW_POLICY_ALLOW", "TAMPERED_INTENT", "INTENT_AUTHORITY_FORBIDDEN", "EXPIRED_INTENT", "MISSING_DECISION_LINK", "MISSING_EVIDENCE_LINK", "MISSING_PROPOSAL_LINK", "DUPLICATE_INTENT", "IDEMPOTENCY_COLLISION", "POLICY_UNAVAILABLE", "UNKNOWN_ACTION_POLICY"]:
    if not any(case.get("expected_reason") == reason for case in m08_cases):
        errors.append(f"M08 eval missing semantic case: {reason}")

m09_cases = load_json(missions.get("M09", {}).get("eval", ""))
for expected in ["AUTHORIZED", "WAIT_APPROVAL", "DENY_REJECTED", "DENY_INVALID_APPROVER", "DENY_APPROVAL_MISMATCH", "DENY_APPROVAL_BEFORE_POLICY", "DENY_EXPIRED_APPROVAL", "DENY_KILL_SWITCH", "DENY_EXECUTOR", "DENY_POLICY_STATE", "WAIT_ALREADY_EXECUTED", "DENY_TAMPERED_INTENT", "DENY_EXPIRED_INTENT"]:
    if not any(case.get("expected") == expected for case in m09_cases):
        errors.append(f"M09 eval missing semantic case: {expected}")

m10_cases = load_json(missions.get("M10", {}).get("eval", ""))
for decision, reason in [
    ("ALLOW_CANARY", "CANARY_ELIGIBLE"), ("REQUIRE_APPROVAL", "RISK2_PER_ACTION_APPROVAL_REQUIRED"),
    ("REQUIRE_APPROVAL", "RISK_NOT_DELEGATED"), ("REQUIRE_APPROVAL", "SCOPE_NOT_DELEGATED"),
    ("DENY", "GRANT_REVOKED"), ("DENY", "KILL_SWITCH_ACTIVE"), ("DENY", "TAMPERED_GRANT"),
    ("REQUIRE_APPROVAL", "CANARY_TOTAL_BUDGET_EXHAUSTED"), ("WAIT", "RATE_LIMIT_REACHED"),
    ("REQUIRE_APPROVAL", "CANARY_COST_BUDGET_EXHAUSTED"), ("WAIT", "OUTCOME_BACKPRESSURE"),
    ("WAIT", "RECONCILIATION_REQUIRED"),
]:
    if not any(case.get("expected_decision") == decision and case.get("expected_reason") == reason for case in m10_cases):
        errors.append(f"M10 eval missing semantic case: {decision}/{reason}")

if errors:
    print("AGENT SEMANTIC VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)
print("AGENT SEMANTIC VALIDATION PASS: M03-M10 behavior is authority-bounded without duplicating structural source-of-truth")
