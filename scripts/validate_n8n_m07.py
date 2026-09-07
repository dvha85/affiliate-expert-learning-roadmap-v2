"""Static safety contract for the M07 evidence-agent blueprint."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
blueprint = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {"Manual Trigger", "Tool Registry + Evidence Context", "Canonical History GET", "Resolve Canonical Evidence Payload", "Read-only Evidence Agent", "public_http GET only", "Grounding Boundary"}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M07 nodes: {sorted(missing)}")
context = nodes["Tool Registry + Evidence Context"]["parameters"]["assignments"]["assignments"]
values = {item["name"]: item["value"] for item in context}
if "e1" in str(values.get("evidence_ids", "")):
    raise SystemExit("M07 must not use placeholder evidence ID")
if "record_id" not in str(values) or "canonical_store_url" not in str(values):
    raise SystemExit("M07 must resolve a record from the canonical store")
tool = nodes["public_http GET only"]
if tool["parameters"].get("method") != "GET":
    raise SystemExit("M07 tool must remain GET-only")
registry = values.get("tool_registry_json", "")
for marker in ['"read_only":true', '"allowed_methods":["GET"]', '"allowed_hosts"']:
    if marker not in registry:
        raise SystemExit(f"M07 registry marker missing: {marker}")
boundary = nodes["Grounding Boundary"]["parameters"]["jsCode"]
for marker in ["AGENT_OUTPUT_NOT_JSON", "FORGED_EVIDENCE_ID", "CLAIM_VALUE_NOT_BOUND", "field_or_claim", "authority:'A2-RO'", "write_permission:false", "$execution.id"]:
    if marker not in boundary:
        raise SystemExit(f"M07 grounding boundary marker missing: {marker}")
instruction = values.get("instruction", "")
for marker in ["untrusted data", "Never request, simulate, or invent write actions"]:
    if marker not in instruction:
        raise SystemExit(f"M07 instruction safety marker missing: {marker}")
if "Resolve Canonical Evidence Payload" not in nodes or "Registry Policy Preflight" not in nodes:
    raise SystemExit("M07 must preflight registry before Agent/tool request")
print("N8N M07 STATIC CONTRACT PASS: canonical payload, preflight registry, GET-only tool, grounded output and A2-RO ceiling")
