from pathlib import Path
import json
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []

required_files = [
    "docs/architecture/LEARNER-BOT-CONTINUITY.md",
    "starter-kits/CONTINUITY-CHECKPOINT.md",
    "lab/affiliate-bot/README.md",
    "lab/mission-runtime/README.md",
    "curriculum/M06/M06.3-n8n-readonly-workflow.md",
    "lab/n8n/M06-readonly-watcher.blueprint.json",
]
for rel in required_files:
    if not (ROOT / rel).exists():
        errors.append(f"missing continuity asset: {rel}")

continuity = (ROOT / "docs/architecture/LEARNER-BOT-CONTINUITY.md").read_text(encoding="utf-8")
for marker in [
    "M01 learner Bot baseline",
    "lab/mission-runtime/",
    "Conformance oracle/harness",
    "n8n static data != canonical history",
    "previous_mission_artifact_refs",
]:
    if marker not in continuity:
        errors.append(f"continuity architecture marker missing: {marker}")

curriculum = (ROOT / "curriculum/README.md").read_text(encoding="utf-8")
for marker in ["Continuity Gate", "starter-kits/CONTINUITY-CHECKPOINT.md", "lab/mission-runtime", "!= Bot thứ hai"]:
    if marker not in curriculum:
        errors.append(f"curriculum continuity marker missing: {marker}")

bot_readme = (ROOT / "lab/affiliate-bot/README.md").read_text(encoding="utf-8")
for marker in ["continuity anchor", "M03–M11", "conformance oracle/harness", "Integration Evidence"]:
    if marker not in bot_readme:
        errors.append(f"affiliate-bot continuity marker missing: {marker}")

runtime_readme = (ROOT / "lab/mission-runtime/README.md").read_text(encoding="utf-8")
for marker in ["O00 và M03–M11", "go run ./cmd/demo M11", "không phải Affiliate Bot thứ hai"]:
    if marker not in runtime_readme:
        errors.append(f"mission-runtime role marker missing: {marker}")

try:
    blueprint = json.loads((ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json").read_text(encoding="utf-8"))
    nodes = {node.get("name"): node for node in blueprint.get("nodes", [])}
    normalize = nodes.get("Parse Provenance + Change Detect", {}).get("parameters", {}).get("jsCode", "")
    handoff = nodes.get("Canonical History Adapter", {})
    handoff_code = json.dumps(handoff.get("parameters", {}))
    ack_code = nodes.get("Require Canonical Store ACK", {}).get("parameters", {}).get("jsCode", "")
    if "watcher_fingerprint" not in normalize:
        errors.append("M06 n8n must name static state watcher_fingerprint")
    if "observation_id" not in normalize:
        errors.append("M06 n8n must emit canonical observation_id")
    if "content_hash" not in normalize or "previous===content_hash" not in normalize:
        errors.append("M06 change detection must compare content hashes")
    if "200000" not in normalize:
        errors.append("M06 watcher cache must fail closed on oversized snapshots")
    if "store.history" in normalize or "store.history" in handoff_code:
        errors.append("M06 n8n static data must never be canonical history")
    for marker in ["canonical_store_url"]:
        if marker not in handoff_code:
            errors.append(f"M06 canonical-history handoff marker missing: {marker}")
    for marker in ["canonical_history_handoff:'ACK'", "canonical_history_persisted:true", "CANONICAL_HISTORY_NOT_ACKNOWLEDGED"]:
        if marker not in ack_code:
            errors.append(f"M06 ACK gate marker missing: {marker}")
    if "ACK" not in ack_code:
        errors.append("M06 adapter must expose an ACK gate")
except Exception as exc:
    errors.append(f"invalid M06 n8n blueprint: {exc}")

m06_lesson = (ROOT / "curriculum/M06/M06.3-n8n-readonly-workflow.md").read_text(encoding="utf-8")
for marker in ["Watcher cache (bộ nhớ đệm) != canonical history", "canonical_history_handoff=ACK", "Canonical History Adapter", "Continuity Gate"]:
    if marker not in m06_lesson:
        errors.append(f"M06 lesson boundary marker missing: {marker}")

m06_runtime = (ROOT / "core/m06/m06.go").read_text(encoding="utf-8")
if not re.search(r'ObservationID\s+string\s+`json:"observation_id"`', m06_runtime):
    errors.append("M06 conformance runtime must emit canonical observation_id")

if errors:
    print("CONTINUITY VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("CONTINUITY VALIDATION PASS: learner Bot stays continuous and n8n cache does not own canonical history")
