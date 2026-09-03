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
    "missions/README.md", "missions/M00-first-real-evidence-packet.md",
    "missions/M01-smallest-deterministic-bot.md",
    "starter-kits/M01-deterministic-bot/CHECKPOINTS.md",
    "starter-kits/M01-deterministic-bot/M01-OPERATED-EVIDENCE-TEMPLATE.md",
    "evals/M01-deterministic-bot/cases.json",
    "docs/architecture/ARCHITECTURE.md", "docs/technology/TECHNOLOGY-PROFILE.md",
    "contracts/observation.schema.json", "contracts/decision-packet.schema.json",
    "contracts/action-intent.schema.json", "contracts/policy-decision.schema.json",
]

errors = []
for rel in REQUIRED:
    if not (ROOT / rel).exists():
        errors.append(f"missing required file: {rel}")

curriculum = (ROOT / "CURRICULUM.md").read_text(encoding="utf-8")
for marker in [
    "M00 | First Real Evidence Packet", "M01 | Smallest Deterministic Bot v0.1",
    "M03 | First Tracked Human Action", "M07 | Read-only Evidence Agent",
    "M11 | Production Closed Loop", "Decision != Approval != Execution",
    "Tool result != trusted evidence",
]:
    if marker not in curriculum:
        errors.append(f"curriculum marker missing: {marker}")

for path in (ROOT / "contracts").glob("*.json"):
    try:
        json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {path.relative_to(ROOT)}: {exc}")

try:
    eval_cases = json.loads((ROOT / "evals/M01-deterministic-bot/cases.json").read_text(encoding="utf-8"))
    if len(eval_cases) < 8:
        errors.append("M01 eval pack must contain at least 8 cases")
    ids = [case.get("case_id") for case in eval_cases]
    if len(ids) != len(set(ids)):
        errors.append("M01 eval case_id values must be unique")
except Exception as exc:
    errors.append(f"invalid M01 eval pack: {exc}")

m01 = (ROOT / "missions/M01-smallest-deterministic-bot.md").read_text(encoding="utf-8")
for marker in ["status: ready", "starter-kits/M01-deterministic-bot/", "evals/M01-deterministic-bot/", "lab/affiliate-bot/"]:
    if marker not in m01:
        errors.append(f"M01 ready contract missing marker: {marker}")

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

print("VALIDATION PASS: clean v2 authority, M00/M01 delivery assets and contracts are consistent")
