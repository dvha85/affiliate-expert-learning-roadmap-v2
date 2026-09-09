"""Static wiring check: n8n must delegate M07 decisions to the shared adapter."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
blueprint = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {
    "Manual Trigger",
    "M07 Adapter Input",
    "Fetch and Register Tool Adapter",
    "Require Registered Tool ACK",
    "Canonical M07 Context Adapter",
    "Require Canonical M07 Context",
    "Read-only Evidence Agent",
    "Validate Grounding Adapter",
    "Require Grounded Proposal",
    "Persist Agent Proposal Adapter",
    "Require Persisted Agent Proposal ACK",
    "Report Persisted M07 Proposal",
}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M07 adapter nodes: {sorted(missing)}")
for retired in {"Grounding Boundary", "Registry Policy Preflight", "public_http GET only", "Resolve Canonical Evidence Payload", "Adapter Preflight", "Preflighted Read-only HTTP GET", "Build Tool Result Handoff", "Register Tool Result Adapter"}:
    if retired in nodes:
        raise SystemExit(f"retired local M07 policy node remains: {retired}")

input_values = {
    item["name"]: item["value"]
    for item in nodes["M07 Adapter Input"]["parameters"]["assignments"]["assignments"]
}
for field in ("adapter_url", "record_id", "tool_registry_json", "tool_request_json", "instruction"):
    if field not in input_values:
        raise SystemExit(f"adapter input missing {field}")
if input_values["adapter_url"] != "http://127.0.0.1:8787":
    raise SystemExit("M07 adapter_url must be a fixed loopback boundary, not event input")
if "$json.tool_registry" in input_values["tool_registry_json"] or '"allowed_hosts":["example.com"]' not in input_values["tool_registry_json"]:
    raise SystemExit("M07 tool registry must be a fixed reviewed policy, not event input")
for marker in ("untrusted data", "Never request or claim write authority", "proposed_action"):
    if marker not in input_values["instruction"]:
        raise SystemExit(f"M07 instruction safety marker missing: {marker}")

fetch = nodes["Fetch and Register Tool Adapter"]["parameters"]
if fetch.get("method") != "POST" or "/v1/m07/fetch-and-register" not in fetch.get("url", ""):
    raise SystemExit("M07 fetch must be owned by the shared adapter")
if "tool_request" not in fetch.get("jsonBody", "") or "tool_registry_json" not in fetch.get("jsonBody", ""):
    raise SystemExit("adapter fetch must receive the request and registry")

for name, markers in {
    "Require Registered Tool ACK": {"REGISTERED_TOOL_TRACE_NOT_ACKNOWLEDGED", "response.status!=='ACK'", "response.artifact_id", "response.evidence.evidence_id", "response.evidence_raw_json"},
    "Require Canonical M07 Context": {"CANONICAL_M07_CONTEXT_NOT_ACKNOWLEDGED", "response.status!=='VALID'", "response.artifact.record_id!==expected", "Array.isArray(response.artifact.evidence)", "response.artifact_raw_json"},
    "Require Grounded Proposal": {"GROUNDING_NOT_A_PERSISTABLE_PROPOSAL", "response.artifact.state!=='HUMAN_REVIEW'", "response.artifact.proposed_action", "response.execution_permitted!==false"},
    "Require Persisted Agent Proposal ACK": {"AGENT_PROPOSAL_NOT_ACKNOWLEDGED", "response.status!=='ACK'", "response.artifact_id", "execution_permitted:false"},
}.items():
    code = nodes[name]["parameters"].get("jsCode", "")
    for marker in markers:
        if marker not in code:
            raise SystemExit(f"{name} is missing enforced ACK marker: {marker}")

for name, endpoint in {
    "Canonical M07 Context Adapter": "/v1/m07/context",
    "Validate Grounding Adapter": "/v1/m07/validate",
    "Persist Agent Proposal Adapter": "/v1/m07/register-proposal",
}.items():
    params = nodes[name]["parameters"]
    if params.get("method") != "POST" or endpoint not in params.get("url", ""):
        raise SystemExit(f"{name} must call {endpoint}")

connections = blueprint.get("connections", {})
expected = [
    ("M07 Adapter Input", "Fetch and Register Tool Adapter"),
    ("Fetch and Register Tool Adapter", "Require Registered Tool ACK"),
    ("Require Registered Tool ACK", "Canonical M07 Context Adapter"),
    ("Canonical M07 Context Adapter", "Require Canonical M07 Context"),
    ("Require Canonical M07 Context", "Read-only Evidence Agent"),
    ("Read-only Evidence Agent", "Validate Grounding Adapter"),
    ("Validate Grounding Adapter", "Require Grounded Proposal"),
    ("Require Grounded Proposal", "Persist Agent Proposal Adapter"),
    ("Persist Agent Proposal Adapter", "Require Persisted Agent Proposal ACK"),
    ("Require Persisted Agent Proposal ACK", "Report Persisted M07 Proposal"),
]
for source, target in expected:
    destinations = {item["node"] for branch in connections.get(source, {}).get("main", []) for item in branch}
    if target not in destinations:
        raise SystemExit(f"missing M07 adapter flow {source} -> {target}")

agent_text = nodes["Read-only Evidence Agent"]["parameters"].get("text", "")
if "artifact_raw_json" not in agent_text or "evidence_raw_json" not in agent_text or "JSON.stringify" in agent_text:
    raise SystemExit("M07 agent must receive adapter-preserved JSON text, not reserialized numeric values")
for name in ("Validate Grounding Adapter", "Persist Agent Proposal Adapter"):
    body = nodes[name]["parameters"].get("jsonBody", "")
    if "model_output_text" not in body or "model_output:JSON.parse" in body:
        raise SystemExit(f"{name} must send model JSON as exact text to the adapter")

print("N8N M07 STATIC WIRING PASS: fixed loopback adapter, registered trace/context/grounding/proposal ACKs, and persistence handoff are mandatory")
