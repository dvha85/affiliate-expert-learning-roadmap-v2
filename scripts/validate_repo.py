from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]

REQUIRED = [
    "README.md",
    "CURRICULUM.md",
    "ROADMAP.md",
    "curriculum/README.md",
    "curriculum/M00/M00.1-affiliate-intelligence-objective.md",
    "curriculum/M00/M00.2-evidence-uncertainty.md",
    "curriculum/M00/M00.3-decision-approval-execution.md",
    "missions/README.md",
    "missions/M00-first-real-evidence-packet.md",
    "missions/M01-smallest-deterministic-bot.md",
    "docs/architecture/ARCHITECTURE.md",
    "docs/technology/TECHNOLOGY-PROFILE.md",
    "contracts/observation.schema.json",
    "contracts/decision-packet.schema.json",
    "contracts/action-intent.schema.json",
    "contracts/policy-decision.schema.json",
]

errors = []
for rel in REQUIRED:
    if not (ROOT / rel).exists():
        errors.append(f"missing required file: {rel}")

curriculum = (ROOT / "CURRICULUM.md").read_text(encoding="utf-8")
for marker in [
    "M00 | First Real Evidence Packet",
    "M01 | Smallest Deterministic Bot v0.1",
    "M03 | First Tracked Human Action",
    "M07 | Read-only Evidence Agent",
    "M11 | Production Closed Loop",
    "Decision != Approval != Execution",
    "Tool result != trusted evidence",
]:
    if marker not in curriculum:
        errors.append(f"curriculum marker missing: {marker}")

for path in (ROOT / "contracts").glob("*.json"):
    try:
        json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {path.relative_to(ROOT)}: {exc}")

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

print("VALIDATION PASS: clean v2 authority, contracts and learner path are consistent")
