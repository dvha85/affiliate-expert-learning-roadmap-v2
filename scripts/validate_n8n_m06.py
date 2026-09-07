"""Static safety contract for the M06 n8n blueprint; not engine execution evidence."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
path = ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json"
blueprint = json.loads(path.read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {
    "Schedule Trigger",
    "Allowed Source + Canonical Store",
    "Read-only HTTP GET",
    "Parse Provenance + Change Detect",
    "Build Canonical History Record",
    "Canonical History Adapter",
    "Require Canonical Store ACK",
    "Report NEW UNCHANGED CHANGED",
}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M06 nodes: {sorted(missing)}")
http = nodes["Read-only HTTP GET"]
if http["parameters"].get("method") != "GET":
    raise SystemExit("M06 HTTP node must remain GET-only")
source_url = nodes["Allowed Source + Canonical Store"]["parameters"]["assignments"]["assignments"][1]["value"]
if not source_url.startswith("https://"):
    raise SystemExit("M06 source placeholder must be HTTPS")
normalize = nodes["Parse Provenance + Change Detect"]["parameters"]["jsCode"]
for marker in [
    "SOURCE_TOO_LARGE_FOR_SAFE_CHANGE_DETECTION",
    "content_hash",
    "change_state=previous===undefined?'NEW'",
    "'UNCHANGED'",
    "'CHANGED'",
    "source_authority_or_role",
    "claim_kind",
]:
    if marker not in normalize:
        raise SystemExit(f"M06 normalize safety marker missing: {marker}")
handoff = nodes["Require Canonical Store ACK"]
code = handoff["parameters"]["jsCode"]
for marker in ["CANONICAL_HISTORY_NOT_ACKNOWLEDGED", "canonical_history_handoff:'ACK'", "canonical_history_persisted:true"]:
    if marker not in code:
        raise SystemExit(f"M06 handoff boundary marker missing: {marker}")
record = nodes["Build Canonical History Record"]["parameters"]["jsCode"]
for marker in ["input_hash", "recorded_result", "decision_id", "evidence_ids", "formula_version"]:
    if marker not in record:
        raise SystemExit(f"M06 canonical record marker missing: {marker}")
print("N8N M06 STATIC CONTRACT PASS: nodes, GET-only, provenance parser, canonical adapter and ACK gate")
