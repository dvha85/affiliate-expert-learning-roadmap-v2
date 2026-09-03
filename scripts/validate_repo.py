from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]

REQUIRED = [
    "README.md", "CURRICULUM.md", "ROADMAP.md", "curriculum/README.md",
    "curriculum/M00/M00.1-affiliate-intelligence-objective.md",
    "curriculum/M00/M00.2-evidence-uncertainty.md",
    "curriculum/M00/M00.3-decision-approval-execution.md",
    "curriculum/M01/M01.1-deterministic-contract.md",
    "curriculum/M01/M01.2-missing-zero-invalid.md",
    "curriculum/M01/M01.3-conflict-and-authority.md",
    "curriculum/M01/M01.4-failure-first-operated-proof.md",
    "curriculum/M02/M02.1-identity-time-history.md",
    "curriculum/M02/M02.2-append-only-integrity.md",
    "curriculum/M02/M02.3-versioned-replay-drift.md",
    "curriculum/M02/M02.4-restart-query-operated-proof.md",
    "missions/README.md", "missions/M00-first-real-evidence-packet.md",
    "missions/M01-smallest-deterministic-bot.md",
    "missions/M02-trustworthy-history-replay.md",
    "starter-kits/M01-deterministic-bot/CHECKPOINTS.md",
    "starter-kits/M01-deterministic-bot/M01-OPERATED-EVIDENCE-TEMPLATE.md",
    "starter-kits/M02-history-replay/CHECKPOINTS.md",
    "starter-kits/M02-history-replay/M02-OPERATED-EVIDENCE-TEMPLATE.md",
    "evals/M01-deterministic-bot/cases.json",
    "evals/M02-history-replay/cases.json",
    "docs/architecture/ARCHITECTURE.md", "docs/technology/TECHNOLOGY-PROFILE.md",
    "contracts/observation.schema.json", "contracts/decision-packet.schema.json",
    "contracts/history-record.schema.json", "contracts/action-intent.schema.json",
    "contracts/policy-decision.schema.json",
    "lab/affiliate-bot/cmd/bot/history.go", "lab/affiliate-bot/cmd/bot/history_test.go",
]

errors = []
for rel in REQUIRED:
    if not (ROOT / rel).exists():
        errors.append(f"missing required file: {rel}")

curriculum = (ROOT / "CURRICULUM.md").read_text(encoding="utf-8")
for marker in [
    "M00 | First Real Evidence Packet", "M01 | Smallest Deterministic Bot v0.1",
    "M02 | Trustworthy History + Replay v0.2", "M03 | First Tracked Human Action",
    "M07 | Read-only Evidence Agent", "M11 | Production Closed Loop",
    "Decision != Approval != Execution", "Tool result != trusted evidence",
]:
    if marker not in curriculum:
        errors.append(f"curriculum marker missing: {marker}")

for path in (ROOT / "contracts").glob("*.json"):
    try:
        json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {path.relative_to(ROOT)}: {exc}")


def validate_eval(path_str, minimum, prefix):
    try:
        cases = json.loads((ROOT / path_str).read_text(encoding="utf-8"))
        if len(cases) < minimum:
            errors.append(f"{prefix} eval pack must contain at least {minimum} cases")
        ids = [case.get("case_id") for case in cases]
        if any(not case_id for case_id in ids):
            errors.append(f"{prefix} eval case_id must be non-empty")
        if len(ids) != len(set(ids)):
            errors.append(f"{prefix} eval case_id values must be unique")
    except Exception as exc:
        errors.append(f"invalid {prefix} eval pack: {exc}")


validate_eval("evals/M01-deterministic-bot/cases.json", 8, "M01")
validate_eval("evals/M02-history-replay/cases.json", 8, "M02")

m01 = (ROOT / "missions/M01-smallest-deterministic-bot.md").read_text(encoding="utf-8")
for marker in ["status: ready", "starter-kits/M01-deterministic-bot/", "evals/M01-deterministic-bot/", "lab/affiliate-bot/"]:
    if marker not in m01:
        errors.append(f"M01 ready contract missing marker: {marker}")

m02 = (ROOT / "missions/M02-trustworthy-history-replay.md").read_text(encoding="utf-8")
for marker in [
    "status: ready", "starter-kits/M02-history-replay/", "evals/M02-history-replay/",
    "MATCH | DRIFT | UNREPLAYABLE", "observed_at", "ingested_at", "as_of",
]:
    if marker not in m02:
        errors.append(f"M02 ready contract missing marker: {marker}")

history_schema = json.loads((ROOT / "contracts/history-record.schema.json").read_text(encoding="utf-8"))
history_required = set(history_schema.get("required", []))
for field in ["record_id", "as_of", "ingested_at", "formula_version", "input_hash", "observations", "recorded_result"]:
    if field not in history_required:
        errors.append(f"HistoryRecord required field missing: {field}")

for forbidden in [
    ROOT / "lessons" / "V2-LESSON-MAP.json",
    ROOT / "legacy-v1-M00",
    ROOT / "scripts" / "migrate_curriculum_v1_to_v2.py",
]:
    if forbidden.exists():
        errors.append(f"legacy compatibility asset must not exist: {forbidden.relative_to(ROOT)}")

if errors:
    print("VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("VALIDATION PASS: clean v2 authority and M00/M01/M02 delivery assets are consistent")
