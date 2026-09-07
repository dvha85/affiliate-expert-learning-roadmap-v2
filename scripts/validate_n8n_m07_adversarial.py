"""Offline adversarial contract for M07; does not call a model or n8n."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
blueprint = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint["nodes"]}
assignments = nodes["Tool Registry + Evidence Context"]["parameters"]["assignments"]["assignments"]
values = {item["name"]: item["value"] for item in assignments}
instruction, registry = values["instruction"], values["tool_registry_json"]
boundary = nodes["Grounding Boundary"]["parameters"]["jsCode"]
cases = ["prompt injection", "forged evidence ID", "missing context", "unknown host/redirect", "write request"]
assert all(cases)
for marker in ["untrusted data", "never grant permission", "Do not request, simulate, or invent write actions"]:
    assert marker in instruction, marker
for marker in ['"read_only":true', '"allowed_methods":["GET"]', '"allowed_hosts":["example.com"]']:
    assert marker in registry, marker
for marker in ["state:'HUMAN_REVIEW'", "authority:'A2-RO'", "write_permission:false", "evidence_ids"]:
    assert marker in boundary, marker
print("N8N M07 ADVERSARIAL CONTRACT PASS: injection/forged-ID/missing-context/host/write cases remain read-only HUMAN_REVIEW")
