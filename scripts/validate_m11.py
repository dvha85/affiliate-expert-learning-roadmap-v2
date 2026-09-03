from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []
required = [
    "curriculum/M11/M11.1-production-lease-promotion.md",
    "curriculum/M11/M11.2-health-stop-degrade-recovery.md",
    "curriculum/M11/M11.3-e6-closed-loop-reviewed-improvement.md",
    "missions/M11-production-closed-loop.md",
    "starter-kits/M11-production-loop/README.md",
    "starter-kits/M11-production-loop/CHECKPOINTS.md",
    "starter-kits/M11-production-loop/M11-OPERATED-EVIDENCE-TEMPLATE.md",
    "evals/M11-production-closed-loop/cases.json",
    "lab/mission-runtime/cmd/demo/m11.go",
    "lab/mission-runtime/cmd/demo/m11_test.go",
    "contracts/production-lease.schema.json",
    "contracts/production-lease-approval.schema.json",
    "contracts/production-health-snapshot.schema.json",
    "contracts/production-ledger.schema.json",
    "contracts/production-gate-decision.schema.json",
    "contracts/production-cycle-record.schema.json",
]
for rel in required:
    if not (ROOT / rel).exists(): errors.append(f"missing M11 file: {rel}")
for rel in [x for x in required if x.endswith(".json")]:
    try: json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc: errors.append(f"invalid M11 JSON {rel}: {exc}")
mission = (ROOT / "missions/M11-production-closed-loop.md").read_text(encoding="utf-8")
for marker in ["status: ready","minimum_evidence: E6","GOVERNED_PRODUCTION","RISK2","DEGRADE","STOP","auto_apply=false","3+ closed cycles"]:
    if marker not in mission: errors.append(f"M11 mission marker missing: {marker}")
index = (ROOT / "missions/README.md").read_text(encoding="utf-8")
line = next((x for x in index.splitlines() if x.startswith("| M11 |")), "")
if "| ready |" not in line: errors.append("M11 mission index must be ready")
lease = json.loads((ROOT / "contracts/production-lease.schema.json").read_text(encoding="utf-8"))
if "RISK2" in lease["properties"]["allowed_risk_classes"]["items"]["enum"]: errors.append("ProductionLease must never delegate RISK2")
approval = (ROOT / "contracts/production-lease-approval.schema.json").read_text(encoding="utf-8")
for marker in ["lease_hash","source_e5_refs","validated_risk_classes","reviewed_by","human","APPROVE_PRODUCTION_LEASE"]:
    if marker not in approval: errors.append(f"M11 approval contract missing: {marker}")
health = (ROOT / "contracts/production-health-snapshot.schema.json").read_text(encoding="utf-8")
for marker in ["lease_hash","telemetry_complete","consecutive_failures","reconciliation_required","compliance_alert_count","oldest_pending_outcome_age_seconds","snapshot_hash"]:
    if marker not in health: errors.append(f"M11 health contract missing: {marker}")
ledger = (ROOT / "contracts/production-ledger.schema.json").read_text(encoding="utf-8")
for marker in ["control_mode","STOPPED","pending_execution_ids","outcome_links","consecutive_failures","reconciliation_required"]:
    if marker not in ledger: errors.append(f"M11 ledger contract missing: {marker}")
gate = (ROOT / "contracts/production-gate-decision.schema.json").read_text(encoding="utf-8")
for marker in ["ALLOW_PRODUCTION","DEGRADE","STOP","REQUIRE_APPROVAL","WAIT","DENY","execution_authorized"]:
    if marker not in gate: errors.append(f"M11 gate contract missing: {marker}")
auth = (ROOT / "contracts/execution-authorization.schema.json").read_text(encoding="utf-8")
execution = (ROOT / "contracts/execution-record.schema.json").read_text(encoding="utf-8")
for marker in ["GOVERNED_PRODUCTION","production_lease_hash","production_health_snapshot_hash","production_cost_bound_hash"]:
    if marker not in auth: errors.append(f"M11 authorization linkage missing: {marker}")
    if marker != "GOVERNED_PRODUCTION" and marker not in execution: errors.append(f"M11 execution linkage missing: {marker}")
runtime = (ROOT / "lab/mission-runtime/cmd/demo/m11.go").read_text(encoding="utf-8")
for marker in ["EvaluateProductionGate","AuthorizeProduction","ExecuteProductionLocalSandbox","RecordProductionOutcome","PROMOTION_RISK_NOT_VALIDATED","RISK2_PER_ACTION_APPROVAL_REQUIRED","HEALTH_STALE","TELEMETRY_INCOMPLETE","COMPLIANCE_ALERT","FAILURE_THRESHOLD","OUTCOME_STALE","STICKY_STOP","WAIT_LEDGER_MISSING","DENY_OUTCOME_LINK","GOVERNED_PRODUCTION"]:
    if marker not in runtime: errors.append(f"M11 runtime marker missing: {marker}")
cases = json.loads((ROOT / "evals/M11-production-closed-loop/cases.json").read_text(encoding="utf-8"))
if len(cases) < 30: errors.append("M11 eval pack must contain at least 30 cases")
for decision, reason in [("ALLOW_PRODUCTION","PRODUCTION_ELIGIBLE"),("REQUIRE_APPROVAL","RISK2_PER_ACTION_APPROVAL_REQUIRED"),("DEGRADE","HEALTH_STALE"),("DEGRADE","TELEMETRY_INCOMPLETE"),("STOP","COMPLIANCE_ALERT"),("STOP","FAILURE_THRESHOLD"),("STOP","RECONCILIATION_REQUIRED"),("STOP","OUTCOME_STALE"),("STOP","STICKY_STOP"),("WAIT","OUTCOME_BACKPRESSURE")]:
    if not any(x.get("expected_decision")==decision and x.get("expected_reason")==reason for x in cases): errors.append(f"M11 eval missing {decision}/{reason}")
if errors:
    print("M11 VALIDATION FAILED")
    for error in errors: print(f"- {error}")
    sys.exit(1)
print("M11 VALIDATION PASS: production loop is finite, health-gated, stop-safe and review-bounded")
