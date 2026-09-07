"""Static safety contract for the M06 n8n blueprint; not engine execution evidence."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
path = ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json"
blueprint = json.loads(path.read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {
    "Schedule Trigger",
    "Allowed Source",
    "Read-only HTTP GET",
    "Normalize + Change Detect",
    "Canonical History Handoff",
    "Report NEW UNCHANGED CHANGED",
}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M06 nodes: {sorted(missing)}")
http = nodes["Read-only HTTP GET"]
if http["parameters"].get("method") != "GET":
    raise SystemExit("M06 HTTP node must remain GET-only")
source_url = nodes["Allowed Source"]["parameters"]["assignments"]["assignments"][1]["value"]
if not source_url.startswith("https://"):
    raise SystemExit("M06 source placeholder must be HTTPS")
normalize = nodes["Normalize + Change Detect"]["parameters"]["jsCode"]
for marker in [
    "MAX_CACHE_CHARS",
    "SOURCE_TOO_LARGE_FOR_SAFE_CHANGE_DETECTION",
    "stableCanonical",
    "change_state=previous===undefined?'NEW'",
    "'UNCHANGED'",
    "'CHANGED'",
    "watcher_cache_only:true",
]:
    if marker not in normalize:
        raise SystemExit(f"M06 normalize safety marker missing: {marker}")
handoff = nodes["Canonical History Handoff"]
code = handoff["parameters"]["jsCode"]
for marker in ["canonical_history_handoff:'REQUIRED'", "canonical_history_owner:'DETERMINISTIC_CORE'", "n8n_static_data_is_canonical_history:false"]:
    if marker not in code:
        raise SystemExit(f"M06 handoff boundary marker missing: {marker}")
print("N8N M06 STATIC CONTRACT PASS: nodes, GET-only, exact change detection, cache limit and canonical handoff boundary")
