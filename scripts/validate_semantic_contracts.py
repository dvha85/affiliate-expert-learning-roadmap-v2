from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []


def load(rel):
    try:
        return json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {rel}: {exc}")
        return {}


for rel in [
    "contracts/effect-ref.schema.json",
    "contracts/outcome-record.schema.json",
    "contracts/evaluation-record.schema.json",
    "contracts/action-intent.schema.json",
    "contracts/policy-decision.schema.json",
    "contracts/trusted-cost-bound.schema.json",
    "contracts/production-activation-record.schema.json",
]:
    if not (ROOT / rel).exists():
        errors.append(f"missing semantic contract: {rel}")

if (ROOT / "contracts/canary-cost-bound.schema.json").exists():
    errors.append("mission-specific canary-cost-bound schema must not coexist with canonical TrustedCostBound")

effect = load("contracts/effect-ref.schema.json")
if effect.get("properties", {}).get("effect_kind", {}).get("enum") != ["HUMAN_ACTION", "MACHINE_EXECUTION"]:
    errors.append("EffectRef must distinguish HUMAN_ACTION and MACHINE_EXECUTION")
for field in ["effect_kind", "effect_id"]:
    if field not in effect.get("required", []):
        errors.append(f"EffectRef required field missing: {field}")

for rel in ["contracts/outcome-record.schema.json", "contracts/evaluation-record.schema.json"]:
    schema = load(rel)
    props = schema.get("properties", {})
    if props.get("effect_ref", {}).get("$ref") != "effect-ref.schema.json":
        errors.append(f"{rel} must reference canonical EffectRef")
    if "effect_ref" not in schema.get("required", []):
        errors.append(f"{rel} must require effect_ref")
    if "action_id" in props:
        errors.append(f"{rel} must not overload canonical action_id for machine execution")

intent = load("contracts/action-intent.schema.json")
iprops = intent.get("properties", {})
if iprops.get("intent_mode", {}).get("const") != "PROPOSAL_ONLY":
    errors.append("ActionIntent.intent_mode must be PROPOSAL_ONLY")
if iprops.get("execution_authorized", {}).get("const") is not False:
    errors.append("ActionIntent must never authorize execution")
for obsolete in ["shadow_only", "dry_run"]:
    if obsolete in iprops:
        errors.append(f"obsolete ActionIntent canonical field must be removed: {obsolete}")

policy = load("contracts/policy-decision.schema.json")
pprops = policy.get("properties", {})
if pprops.get("policy_mode", {}).get("const") != "NON_AUTHORIZING":
    errors.append("PolicyDecision.policy_mode must be NON_AUTHORIZING")
if pprops.get("execution_authorized", {}).get("const") is not False:
    errors.append("PolicyDecision must never authorize execution")
if "policy_review_required" not in pprops:
    errors.append("PolicyDecision must expose policy_review_required")
for obsolete in ["approval_required", "shadow_only"]:
    if obsolete in pprops:
        errors.append(f"obsolete PolicyDecision canonical field must be removed: {obsolete}")

cost = load("contracts/trusted-cost-bound.schema.json")
for field in ["cost_bound_id", "intent_id", "intent_hash", "max_cost_minor", "currency", "source_ref", "observed_at", "expires_at", "cost_bound_hash"]:
    if field not in cost.get("required", []):
        errors.append(f"TrustedCostBound required field missing: {field}")

activation = load("contracts/production-activation-record.schema.json")
for field in ["lease_id", "lease_version", "lease_hash", "activated_at"]:
    if field not in activation.get("required", []):
        errors.append(f"ProductionActivationRecord required field missing: {field}")

m02 = (ROOT / "missions/M02-trustworthy-history-replay.md").read_text(encoding="utf-8")
for marker in ["minimum_evidence: E1", "readiness_target: E3", "readiness_target: E3\n!= E3 evidence achieved"]:
    if marker not in m02:
        errors.append(f"M02 evidence/readiness semantic marker missing: {marker}")
m09 = (ROOT / "missions/M09-durable-approval-controlled-executor.md").read_text(encoding="utf-8")
for marker in ["minimum_evidence: E4", "readiness_target: E5", "policy_review_required", "ApprovalRecord"]:
    if marker not in m09:
        errors.append(f"M09 evidence/approval semantic marker missing: {marker}")

runtime = "\n".join(
    (ROOT / "lab/mission-runtime/cmd/demo" / name).read_text(encoding="utf-8")
    for name in ["m03_m05.go", "m08.go", "m08_compat.go", "trusted_cost_bound.go", "production_activation.go"]
)
for marker in [
    "EffectRef", "HUMAN_ACTION", "MACHINE_EXECUTION",
    "PROPOSAL_ONLY", "NON_AUTHORIZING", "INTENT_AUTHORITY_FORBIDDEN",
    "TrustedCostBound", "ProductionActivationRecord",
]:
    if marker not in runtime:
        errors.append(f"semantic runtime marker missing: {marker}")

if errors:
    print("SEMANTIC CONTRACT VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("SEMANTIC CONTRACT VALIDATION PASS: effect, intent, policy, cost and readiness semantics are unambiguous")
