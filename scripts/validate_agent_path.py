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
    "lab/mission-runtime/cmd/demo/m08_test.go",
    "lab/mission-runtime/cmd/demo/mission_runtime_test.go",
    "lab/n8n/M06-readonly-watcher.blueprint.json",
    "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    "contracts/action-record.schema.json",
    "contracts/outcome-record.schema.json",
    "contracts/advisor-output.schema.json",
    "contracts/evaluation-record.schema.json",
    "contracts/improvement-proposal.schema.json",
    "contracts/review-record.schema.json",
    "contracts/tool-registry.schema.json",
    "contracts/action-intent.schema.json",
    "contracts/policy-decision.schema.json",
]
starter_names = {
    "M03": "tracked-human-action",
    "M04": "grounded-ai-advisor",
    "M05": "reviewed-improvement",
    "M06": "readonly-watcher",
    "M07": "readonly-evidence-agent",
    "M08": "shadow-policy",
}
for mission in range(3, 9):
    mid = f"M{mission:02d}"
    required.extend([
        f"starter-kits/{mid}-{starter_names[mid]}/CHECKPOINTS.md",
        f"starter-kits/{mid}-{starter_names[mid]}/{mid}-OPERATED-EVIDENCE-TEMPLATE.md",
    ])

for rel in required:
    if not (ROOT / rel).exists():
        errors.append(f"missing required file: {rel}")

for mission in range(3, 9):
    mid = f"M{mission:02d}"
    if not (ROOT / "curriculum" / mid).is_dir(): errors.append(f"missing curriculum directory: {mid}")
    if not any((ROOT / "missions").glob(f"{mid}-*.md")): errors.append(f"missing mission contract: {mid}")
    if not any((ROOT / "starter-kits").glob(f"{mid}-*/README.md")): errors.append(f"missing starter README: {mid}")
    if not any((ROOT / "evals").glob(f"{mid}-*/cases.json")): errors.append(f"missing executable eval pack: {mid}")

json_files = [
    "contracts/action-record.schema.json","contracts/outcome-record.schema.json","contracts/advisor-output.schema.json",
    "contracts/evaluation-record.schema.json","contracts/improvement-proposal.schema.json","contracts/review-record.schema.json",
    "contracts/tool-registry.schema.json","contracts/action-intent.schema.json","contracts/policy-decision.schema.json",
    "lab/n8n/M06-readonly-watcher.blueprint.json","lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    "evals/M08-shadow-policy/cases.json",
]
for rel in json_files:
    try: json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc: errors.append(f"invalid JSON {rel}: {exc}")

mission_index = (ROOT / "missions/README.md").read_text(encoding="utf-8")
for mid in ["M03","M04","M05","M06","M07","M08"]:
    line = next((line for line in mission_index.splitlines() if line.startswith(f"| {mid} |")), "")
    if "| ready |" not in line: errors.append(f"{mid} must be ready in mission index")

runtime = "\n".join((ROOT / "lab/mission-runtime/cmd/demo" / name).read_text(encoding="utf-8") for name in ["m03_m05.go","m06_m07.go","m08.go"])
for marker in [
    "DRY_RUN_ONLY","BROKEN_LINK","REJECT_MACHINE_EXECUTION","REJECT_WRITE_REQUEST","ABSTAIN_FUTURE",
    "REJECT_AUTO_APPLY","REJECT_WRITE_METHOD","REJECT_TOOL","REJECT_UNGROUNDED","NormalizeWatchObservation",
    "ValidateEvaluationRecord","ValidateReviewRecord","TAMPERED_INTENT","EXPIRED_INTENT","IDEMPOTENCY_COLLISION",
    "POLICY_UNAVAILABLE","execution_authorized","SHADOW_POLICY_ALLOW","KnownProposalIDs","UNKNOWN_ACTION_POLICY",
]:
    if marker not in runtime: errors.append(f"runtime safety/integration marker missing: {marker}")

# O00 must be a functional artifact chain, not a list of stage labels.
o00_eval = json.loads((ROOT / "evals/O00-safe-walkthrough/cases.json").read_text(encoding="utf-8"))
required_kinds = {"Observation","ActionIntent","PolicyDecision","ExecutionRecord","EvaluationRecord"}
seen_kinds = {c.get("required_kind") for c in o00_eval}
if not required_kinds.issubset(seen_kinds): errors.append("O00 eval must require the functional artifact chain")

# M06 blueprint must implement stateful change detection and canonical Observation mapping.
m06 = json.loads((ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json").read_text(encoding="utf-8"))
m06_nodes = {n.get("name"): n for n in m06.get("nodes", [])}
for name in ["Schedule Trigger","Allowed Source","Read-only HTTP GET","Normalize + Change Detect","Append Canonical History","Report NEW UNCHANGED CHANGED"]:
    if name not in m06_nodes: errors.append(f"M06 blueprint missing node: {name}")
http = m06_nodes.get("Read-only HTTP GET", {}).get("parameters", {})
if http.get("method") != "GET": errors.append("M06 n8n HTTP node must be GET-only")
normalize_code = m06_nodes.get("Normalize + Change Detect", {}).get("parameters", {}).get("jsCode", "")
for marker in ["$getWorkflowStaticData","subject_id","source_url","observed_at","evidence_kind","claim_kind","limitation","change_state"]:
    if marker not in normalize_code: errors.append(f"M06 normalize/change-detect missing: {marker}")

# M07 blueprint must contain a real Agent + model + GET-only tool, not a placeholder.
m07 = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
types = {n.get("type") for n in m07.get("nodes", [])}
for node_type in ["@n8n/n8n-nodes-langchain.agent","@n8n/n8n-nodes-langchain.lmChatOpenAi","n8n-nodes-base.httpRequestTool"]:
    if node_type not in types: errors.append(f"M07 blueprint missing real Agent component: {node_type}")
tool = next((n for n in m07.get("nodes", []) if n.get("type") == "n8n-nodes-base.httpRequestTool"), {})
if tool.get("parameters", {}).get("method") != "GET": errors.append("M07 connected HTTP tool must be GET-only")
context = next((n for n in m07.get("nodes", []) if n.get("name") == "Tool Registry + Evidence Context"), {})
context_text = json.dumps(context, ensure_ascii=False)
for marker in ["read_only","allowed_methods","allowed_hosts","GET"]:
    if marker not in context_text: errors.append(f"M07 Tool Registry metadata missing: {marker}")

# Prompt-injection regression must carry malicious tool text and exercise both reject and safe-read paths.
m07_cases = json.loads((ROOT / "evals/M07-readonly-evidence-agent/cases.json").read_text(encoding="utf-8"))
injection = [c for c in m07_cases if "prompt-injection" in c.get("case_id", "")]
if len(injection) < 2: errors.append("M07 must include prompt-injection reject and safe-data cases")
if any(not c.get("tool_result_text") for c in injection): errors.append("M07 prompt-injection cases must pass malicious tool text to the evaluator")
if not any(c.get("expected") == "REJECT_TOOL" for c in injection): errors.append("M07 prompt injection must test attempted authority escalation")

# M08 must activate exact shadow ActionIntent + deterministic PolicyDecision semantics without execution authority.
action_schema = (ROOT / "contracts/action-intent.schema.json").read_text(encoding="utf-8")
policy_schema = (ROOT / "contracts/policy-decision.schema.json").read_text(encoding="utf-8")
for marker in ["intent_hash","idempotency_key","correlation_id","shadow_only","dry_run","evidence_ids"]:
    if marker not in action_schema: errors.append(f"M08 ActionIntent contract missing: {marker}")
for marker in ["intent_hash","policy_version","shadow_only","execution_authorized","HUMAN_REVIEW"]:
    if marker not in policy_schema: errors.append(f"M08 PolicyDecision contract missing: {marker}")
m08_cases = json.loads((ROOT / "evals/M08-shadow-policy/cases.json").read_text(encoding="utf-8"))
for expected_reason in ["SHADOW_POLICY_ALLOW","TAMPERED_INTENT","EXPIRED_INTENT","MISSING_DECISION_LINK","MISSING_EVIDENCE_LINK","MISSING_PROPOSAL_LINK","DUPLICATE_INTENT","IDEMPOTENCY_COLLISION","POLICY_UNAVAILABLE","UNKNOWN_ACTION_POLICY"]:
    if not any(c.get("expected_reason") == expected_reason for c in m08_cases): errors.append(f"M08 eval missing failure/boundary case: {expected_reason}")
if not any(c.get("case_id") == "M08-E14-agent-proposal-ref-must-resolve" for c in m08_cases): errors.append("M08 eval must reject orphan Agent proposal_ref")

if errors:
    print("AGENT PATH VALIDATION FAILED")
    for error in errors: print(f"- {error}")
    sys.exit(1)
print("AGENT PATH VALIDATION PASS: O00 and M03-M08 are learner-operable, linked and authority-bounded")
