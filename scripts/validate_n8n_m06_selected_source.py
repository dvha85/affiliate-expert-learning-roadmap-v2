"""Static contract for the selected ACCESSTRADE campaign metadata handoff."""
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
path = ROOT / "lab/n8n/M06-accesstrade-shopee-readonly.blueprint.json"
blueprint = json.loads(path.read_text(encoding="utf-8"))
nodes = {node["name"]: node for node in blueprint["nodes"]}
required = {
    "Manual Sanitized Capture Trigger",
    "Selected Source Capture Input",
    "Build and Append Selected Campaign Metadata",
    "Require Selected Source Canonical ACK",
    "Report Selected Source Read-Only Result",
}
if required - nodes.keys():
    raise SystemExit(f"missing selected-source nodes: {sorted(required - nodes.keys())}")
if blueprint.get("active") is not False:
    raise SystemExit("selected-source workflow must not auto-run")
serialized = json.dumps(blueprint, ensure_ascii=False)
for forbidden in ("campaign/5087153089503673507", "pub2.accesstrade.vn", "source_page_sha256"):
    if forbidden.casefold() in serialized.casefold():
        raise SystemExit(f"selected-source blueprint contains forbidden source/account data: {forbidden}")
handoff = nodes["Build and Append Selected Campaign Metadata"]["parameters"]
if handoff.get("method") != "POST" or "/v1/m06/accesstrade-shopee-campaign" not in handoff.get("url", ""):
    raise SystemExit("selected-source handoff bypasses the fixed M06 adapter")
if "JSON.parse" not in handoff.get("jsonBody", "") or "capture" not in handoff.get("jsonBody", ""):
    raise SystemExit("selected-source handoff must pass only a sanitized capture object")
ack = nodes["Require Selected Source Canonical ACK"]["parameters"]["jsCode"]
for marker in ("SELECTED_SOURCE_CANONICAL_HISTORY_NOT_ACKNOWLEDGED", "canonical_history_ack!==true", "canonical_history_persisted!==true", "execution_permitted!==false"):
    if marker not in ack:
        raise SystemExit(f"selected-source ACK boundary lacks {marker}")
if "observed_campaign_metadata_not_business_outcome" not in serialized:
    raise SystemExit("selected-source report loses business-outcome boundary")
print("N8N M06 SELECTED SOURCE STATIC CONTRACT PASS: sanitized capture reaches fixed read-only canonical adapter only")
