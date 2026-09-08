"""Static wiring check: n8n must delegate M07 decisions to the shared adapter."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
blueprint = json.loads((ROOT / "lab/n8n/M07-readonly-evidence-agent.blueprint.json").read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint.get("nodes", [])}
required = {
    "Manual Trigger",
    "M07 Adapter Input",
    "Adapter Preflight",
    "Preflighted Read-only HTTP GET",
    "Build Tool Result Handoff",
    "Register Tool Result Adapter",
    "Canonical M07 Context Adapter",
    "Read-only Evidence Agent",
    "Validate Grounding Adapter",
    "Persist Agent Proposal Adapter",
}
missing = required - nodes.keys()
if missing:
    raise SystemExit(f"missing M07 adapter nodes: {sorted(missing)}")
for retired in {"Grounding Boundary", "Registry Policy Preflight", "public_http GET only", "Resolve Canonical Evidence Payload"}:
    if retired in nodes:
        raise SystemExit(f"retired local M07 policy node remains: {retired}")

input_values = {
    item["name"]: item["value"]
    for item in nodes["M07 Adapter Input"]["parameters"]["assignments"]["assignments"]
}
for field in ("adapter_url", "record_id", "tool_registry_json", "tool_request_json", "instruction"):
    if field not in input_values:
        raise SystemExit(f"adapter input missing {field}")
for marker in ("untrusted data", "Never request or claim write authority", "proposed_action"):
    if marker not in input_values["instruction"]:
        raise SystemExit(f"M07 instruction safety marker missing: {marker}")

preflight = nodes["Adapter Preflight"]["parameters"]
if "/v1/m07/preflight" not in preflight.get("url", "") or preflight.get("method") != "POST":
    raise SystemExit("M07 preflight must call shared adapter before remote request")
fetch = nodes["Preflighted Read-only HTTP GET"]["parameters"]
if "Adapter Preflight" not in fetch.get("method", "") or "Adapter Preflight" not in fetch.get("url", ""):
    raise SystemExit("remote request must consume adapter preflight artifact")
response = fetch.get("options", {}).get("response", {}).get("response", {})
redirect = fetch.get("options", {}).get("redirect", {}).get("redirect", {})
if response.get("fullResponse") is not True or response.get("neverError") is not True or redirect.get("followRedirects") is not False:
    raise SystemExit("remote request must preserve status and disable redirects")
handoff = nodes["Build Tool Result Handoff"]["parameters"]["jsCode"]
for marker in ("statusCode", "response.body", "tool_result", "ADAPTER_OR_FULL_RESPONSE_REQUIRED"):
    if marker not in handoff:
        raise SystemExit(f"tool result handoff marker missing: {marker}")

for name, endpoint in {
    "Register Tool Result Adapter": "/v1/m07/register-tool-result",
    "Canonical M07 Context Adapter": "/v1/m07/context",
    "Validate Grounding Adapter": "/v1/m07/validate",
    "Persist Agent Proposal Adapter": "/v1/m07/register-proposal",
}.items():
    params = nodes[name]["parameters"]
    if params.get("method") != "POST" or endpoint not in params.get("url", ""):
        raise SystemExit(f"{name} must call {endpoint}")

connections = blueprint.get("connections", {})
expected = [
    ("M07 Adapter Input", "Adapter Preflight"),
    ("Adapter Preflight", "Preflighted Read-only HTTP GET"),
    ("Preflighted Read-only HTTP GET", "Build Tool Result Handoff"),
    ("Build Tool Result Handoff", "Register Tool Result Adapter"),
    ("Register Tool Result Adapter", "Canonical M07 Context Adapter"),
    ("Canonical M07 Context Adapter", "Read-only Evidence Agent"),
    ("Read-only Evidence Agent", "Validate Grounding Adapter"),
    ("Validate Grounding Adapter", "Persist Agent Proposal Adapter"),
]
for source, target in expected:
    destinations = {item["node"] for branch in connections.get(source, {}).get("main", []) for item in branch}
    if target not in destinations:
        raise SystemExit(f"missing M07 adapter flow {source} -> {target}")

print("N8N M07 STATIC WIRING PASS: preflight, full-response handoff, trace registration, validation and proposal persistence use the shared adapter")
