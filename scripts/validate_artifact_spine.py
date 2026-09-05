from pathlib import Path
import json
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []


def load_json(rel):
    try:
        return json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc:
        errors.append(f"invalid JSON {rel}: {exc}")
        return {}


observation = load_json("contracts/observation.schema.json")
observation_required = set(observation.get("required", []))
for field in [
    "observation_id", "subject_id", "observed_at", "access_method",
    "evidence_kind", "claim_kind", "state", "limitation",
]:
    if field not in observation_required:
        errors.append(f"Observation canonical identity/provenance field missing: {field}")
source_options = observation.get("anyOf", [])
if not any("source_url" in option.get("required", []) for option in source_options):
    errors.append("Observation must permit/require source_url provenance")
if not any("source_ref" in option.get("required", []) for option in source_options):
    errors.append("Observation must permit/require source_ref provenance for local fixtures")

decision = load_json("contracts/decision-packet.schema.json")
decision_required = set(decision.get("required", []))
for field in ["decision_id", "evidence_ids", "state", "reason", "action"]:
    if field not in decision_required:
        errors.append(f"DecisionPacket lineage field missing: {field}")
if decision.get("properties", {}).get("action", {}).get("type") != "null":
    errors.append("M00 DecisionPacket.action must remain null; ActionIntent is a separate artifact")
evidence_ids = decision.get("properties", {}).get("evidence_ids", {})
if evidence_ids.get("minItems") != 1 or evidence_ids.get("uniqueItems") is not True:
    errors.append("DecisionPacket.evidence_ids must be non-empty and unique")

history = load_json("contracts/history-record.schema.json")
items = history.get("properties", {}).get("observations", {}).get("items", {})
refs = [part.get("$ref") for part in items.get("allOf", []) if isinstance(part, dict)]
if "observation.schema.json" not in refs:
    errors.append("HistoryRecord observations must reuse canonical observation.schema.json")
item_required = set(items.get("required", []))
for field in ["observation_id", "subject_id", "product_id", "observed_at", "evidence_kind"]:
    if field not in item_required:
        errors.append(f"HistoryRecord canonical/domain observation field missing: {field}")
recorded = history.get("properties", {}).get("recorded_result", {})
recorded_required = set(recorded.get("required", []))
for field in ["decision_id", "evidence_ids", "formula_version", "state"]:
    if field not in recorded_required:
        errors.append(f"HistoryRecord recorded_result lineage field missing: {field}")

template = (ROOT / "templates/M00-EVIDENCE-PACKET.md").read_text(encoding="utf-8")
for marker in ["observation_id", "subject_id", "evidence_ids"]:
    if marker not in template:
        errors.append(f"M00 evidence template missing canonical marker: {marker}")

# Inspect the actual packet field, not the checklist or prose mentioning null.
# Accept plain/inline-code keys and values, but require exactly one declaration.
packet_sections = re.findall(
    r"^## Human DecisionPacket[^\n]*\n(.*?)(?=^## |\Z)",
    template, re.MULTILINE | re.DOTALL,
)
action_values = []
if len(packet_sections) == 1:
    action_values = re.findall(
        r"^[ \t]*[-*+] +(?:`action`|action)[ \t]*:[ \t]*(.*)$",
        packet_sections[0], re.MULTILINE,
    )
if len(action_values) != 1 or action_values[0].strip() not in ("null", "`null`"):
    errors.append("M00 Human DecisionPacket must declare exactly one action field with value null")

m02_fixture = load_json("lab/affiliate-bot/data/m02-sample-observations.json")
if not isinstance(m02_fixture, list) or not m02_fixture:
    errors.append("M02 canonical fixture must contain observations")
else:
    seen = set()
    for index, item in enumerate(m02_fixture):
        oid = item.get("observation_id")
        if not oid or oid in seen:
            errors.append(f"M02 fixture observation_id missing/duplicate at index {index}")
        seen.add(oid)
        if item.get("subject_id") != item.get("product_id"):
            errors.append(f"M02 fixture subject_id/product_id mismatch for {oid}")
        for field in ["observed_at", "access_method", "claim_kind", "state", "limitation"]:
            if not item.get(field):
                errors.append(f"M02 fixture missing {field} for {oid}")
        if not item.get("source_url") and not item.get("source_ref"):
            errors.append(f"M02 fixture missing source provenance for {oid}")

runtime = (ROOT / "lab/affiliate-bot/cmd/bot/history.go").read_text(encoding="utf-8")
for marker in [
    "validateCanonicalHistoryObservation",
    "recorded decision_id must equal record_id",
    "recorded evidence_ids must exactly match canonical observation_ids",
    "canonicalEvidenceIDs",
]:
    if marker not in runtime:
        errors.append(f"M02 runtime lineage guard missing: {marker}")

tests = (ROOT / "lab/affiliate-bot/cmd/bot/history_test.go").read_text(encoding="utf-8")
for marker in ["TestHistoryDecisionEvidenceLinkage", "TestHistoryRejectsNonCanonicalObservation"]:
    if marker not in tests:
        errors.append(f"M02 runtime test missing: {marker}")

if errors:
    print("ARTIFACT SPINE VALIDATION FAILED")
    for error in errors:
        print(f"- {error}")
    sys.exit(1)

print("ARTIFACT SPINE VALIDATION PASS: Observation -> DecisionPacket -> HistoryRecord identity and provenance are canonical")
