"""Static wiring check: M06 delegates construction and persistence to the shared adapter."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
path = ROOT / "lab/n8n/M06-readonly-watcher.blueprint.json"
blueprint = json.loads(path.read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {
    "Schedule Trigger",
    "M06 Adapter Input",
    "Build and Append Canonical M06 Adapter",
    "Require Canonical Store ACK",
    "Report Canonical M06 Result",
}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M06 adapter nodes: {sorted(missing)}")
for retired in {"Allowed Source + Canonical Store", "Read-only HTTP GET", "Parse Provenance + Change Detect", "Build Canonical History Record", "Canonical History Adapter"}:
    if retired in nodes:
        raise SystemExit(f"retired local M06 builder node remains: {retired}")

assignments = {
    item["name"]: item["value"]
    for item in nodes["M06 Adapter Input"]["parameters"]["assignments"]["assignments"]
}
for name in {"adapter_url", "fixture_json"}:
    if name not in assignments:
        raise SystemExit(f"M06 adapter input missing {name}")
if "br13-offer-fixture/v1" not in assignments["fixture_json"]:
    raise SystemExit("M06 adapter input must declare the synthetic fixture profile")

handoff = nodes["Build and Append Canonical M06 Adapter"]["parameters"]
if handoff.get("method") != "POST" or "/v1/m06/fixture-import" not in handoff.get("url", ""):
    raise SystemExit("M06 must use the shared fixture-import adapter")
if "fixture_json" not in handoff.get("jsonBody", "") or "JSON.parse" not in handoff.get("jsonBody", ""):
    raise SystemExit("M06 adapter must receive the fixture as JSON data")

ack = nodes["Require Canonical Store ACK"]["parameters"]["jsCode"]
for marker in {"CANONICAL_HISTORY_NOT_ACKNOWLEDGED", "canonical_history_ack!==true", "canonical_history_persisted!==true", "canonical_history_handoff:'ACK'"}:
    if marker not in ack:
        raise SystemExit(f"M06 ACK boundary marker missing: {marker}")

connections = blueprint.get("connections", {})
for source, target in [
    ("Schedule Trigger", "M06 Adapter Input"),
    ("M06 Adapter Input", "Build and Append Canonical M06 Adapter"),
    ("Build and Append Canonical M06 Adapter", "Require Canonical Store ACK"),
    ("Require Canonical Store ACK", "Report Canonical M06 Result"),
]:
    destinations = {item["node"] for branch in connections.get(source, {}).get("main", []) for item in branch}
    if target not in destinations:
        raise SystemExit(f"missing M06 adapter flow {source} -> {target}")

print("N8N M06 STATIC CONTRACT PASS: shared adapter builds, persists and resolves canonical history before ACK")
