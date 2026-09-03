"""Validate the complete learner-path structure from missions/manifest.json.

Mission-specific validators deliberately own semantic assertions. This script
owns only the single canonical O00 → M11 spine, structural assets and runtime
roles.
"""

from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
MANIFEST_PATH = ROOT / "missions/manifest.json"
EXPECTED_SPINE = ["O00"] + [f"M{number:02d}" for number in range(12)]
errors = []


def fail(message):
    errors.append(message)


try:
    manifest = json.loads(MANIFEST_PATH.read_text(encoding="utf-8"))
except Exception as exc:
    print(f"MISSION STRUCTURE VALIDATION FAILED\n- invalid missions/manifest.json: {exc}")
    sys.exit(1)

if manifest.get("schema_version") != 2:
    fail("manifest schema_version must be 2")
if manifest.get("spine") != EXPECTED_SPINE:
    fail(f"manifest spine must be exactly {' → '.join(EXPECTED_SPINE)}")

missions = manifest.get("missions")
if not isinstance(missions, list):
    fail("manifest missions must be an array")
    missions = []
ids = [mission.get("id") for mission in missions if isinstance(mission, dict)]
if ids != EXPECTED_SPINE:
    fail("manifest mission IDs must occur exactly once in canonical spine order")

index = (ROOT / "missions/README.md").read_text(encoding="utf-8")
curriculum_index = (ROOT / "CURRICULUM.md").read_text(encoding="utf-8")
for mission in missions:
    if not isinstance(mission, dict):
        fail("manifest mission entry must be an object")
        continue
    mid = mission.get("id", "<missing>")
    for key in ("contract", "curriculum", "status"):
        if not mission.get(key):
            fail(f"{mid} manifest missing {key}")
    for key in (
        "contract", "curriculum", "starter", "eval", "template",
        "learner_runtime", "conformance_runtime", "orchestration_blueprint",
    ):
        rel = mission.get(key)
        if rel and not (ROOT / rel).exists():
            fail(f"{mid} {key} asset missing: {rel}")
    curriculum_dir = ROOT / mission.get("curriculum", "")
    if curriculum_dir.is_dir() and not any(curriculum_dir.glob("*.md")):
        fail(f"{mid} curriculum directory has no lesson")
    if mid not in index or mid not in curriculum_index:
        fail(f"{mid} missing from learner-facing index")
    expected_status = mission.get("status")
    line = next((line for line in index.splitlines() if line.startswith(f"| {mid} |")), "")
    if expected_status and f"| {expected_status} |" not in line:
        fail(f"{mid} mission index status must be {expected_status}")
    if mid != "O00":
        contract = ROOT / mission.get("contract", "")
        if contract.exists() and "status: ready" not in contract.read_text(encoding="utf-8"):
            fail(f"{mid} ready Mission contract must declare status: ready")

by_id = {mission.get("id"): mission for mission in missions if isinstance(mission, dict)}
if "learner_runtime" in by_id.get("O00", {}):
    fail("O00 is orientation only and must not declare learner_runtime")
if not by_id.get("O00", {}).get("conformance_runtime"):
    fail("O00 must declare conformance_runtime")
for mid in ("M01", "M02"):
    if not by_id.get(mid, {}).get("learner_runtime"):
        fail(f"{mid} must declare learner_runtime")
    if by_id.get(mid, {}).get("conformance_runtime"):
        fail(f"{mid} must not route learner execution through mission-runtime")
for number in range(3, 12):
    mid = f"M{number:02d}"
    if not by_id.get(mid, {}).get("learner_runtime"):
        fail(f"{mid} must declare learner_runtime continuity anchor")
    if not by_id.get(mid, {}).get("conformance_runtime"):
        fail(f"{mid} must declare conformance_runtime separately")
for mid in ("M06", "M07"):
    if not by_id.get(mid, {}).get("orchestration_blueprint"):
        fail(f"{mid} must declare orchestration_blueprint")

m01 = (ROOT / by_id.get("M01", {}).get("contract", "")).read_text(encoding="utf-8")
if "minimum_evidence: E0" not in m01 or "reality_context: E1" not in m01:
    fail("M01 must separate E0 capability evidence from E1 reality context")

if errors:
    print("MISSION STRUCTURE VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("MISSION STRUCTURE VALIDATION PASS: manifest owns one O00 → M11 structural spine and explicit runtime roles")
