"""Static safety contract for the M07 evidence-agent blueprint."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
blueprint = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {"Manual Trigger", "Tool Registry + Evidence Context", "Read-only Evidence Agent", "public_http GET only", "Grounding Boundary"}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M07 nodes: {sorted(missing)}")
context = nodes["Tool Registry + Evidence Context"]["parameters"]["assignments"]["assignments"]
values = {item["name"]: item["value"] for item in context}
if "e1" in str(values.get("evidence_ids", "")):
    raise SystemExit("M07 must not use placeholder evidence ID")
if "evidence_ids" not in str(values.get("evidence_ids", "")):
    raise SystemExit("M07 evidence IDs must come from input context")
tool = nodes["public_http GET only"]
if tool["parameters"].get("method") != "GET":
    raise SystemExit("M07 tool must remain GET-only")
registry = values.get("tool_registry_json", "")
for marker in ['"read_only":true', '"allowed_methods":["GET"]', '"allowed_hosts"']:
    if marker not in registry:
        raise SystemExit(f"M07 registry marker missing: {marker}")
boundary = nodes["Grounding Boundary"]["parameters"]["jsCode"]
for marker in ["state:'HUMAN_REVIEW'", "evidence_ids", "authority:'A2-RO'", "write_permission:false", "$execution.id"]:
    if marker not in boundary:
        raise SystemExit(f"M07 grounding boundary marker missing: {marker}")
instruction = values.get("instruction", "")
for marker in ["untrusted data", "Do not request, simulate, or invent write actions"]:
    if marker not in instruction:
        raise SystemExit(f"M07 instruction safety marker missing: {marker}")
print("N8N M07 STATIC CONTRACT PASS: input evidence binding, GET-only registry, untrusted tool text and HUMAN_REVIEW ceiling")
