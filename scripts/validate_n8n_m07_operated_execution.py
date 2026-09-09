"""Validate a captured, successful n8n M07 execution without reading credentials."""
import argparse
import json
import re
from pathlib import Path


REQUIRED_NODES = (
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
)


def load_execution(path: Path) -> dict:
    raw = path.read_text(encoding="utf-8")
    starts = list(re.finditer(r"(?m)^\{\s*\"data\"\s*:", raw))
    if not starts:
        raise AssertionError("no n8n execution JSON found; pass --rawOutput captured from n8n execute")
    for match in reversed(starts):
        try:
            return json.JSONDecoder().raw_decode(raw[match.start() :])[0]
        except json.JSONDecodeError:
            continue
    raise AssertionError("n8n execution JSON is malformed")


def first_json(run_data: dict, name: str) -> dict:
    runs = run_data.get(name)
    if not isinstance(runs, list) or not runs:
        raise AssertionError(f"missing execution data for {name}")
    run = runs[0]
    if run.get("executionStatus") != "success":
        raise AssertionError(f"{name} did not succeed")
    try:
        return run["data"]["main"][0][0]["json"]
    except (KeyError, IndexError, TypeError) as error:
        raise AssertionError(f"{name} has no usable JSON output") from error


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("execution_json", type=Path)
    parser.add_argument("--proposal-store", type=Path, required=True)
    args = parser.parse_args()

    execution = load_execution(args.execution_json)
    if execution.get("status") != "success" or execution.get("finished") is not True:
        raise AssertionError("n8n execution did not finish successfully")
    run_data = execution.get("data", {}).get("resultData", {}).get("runData", {})
    for name in REQUIRED_NODES:
        first_json(run_data, name)

    grounding = first_json(run_data, "Validate Grounding Adapter").get("body", {})
    artifact = grounding.get("artifact", {})
    if grounding.get("status") != "VALID" or grounding.get("execution_permitted") is not False:
        raise AssertionError("grounding response is not a read-only VALID result")
    if artifact.get("state") != "HUMAN_REVIEW" or artifact.get("authority") != "A2-RO" or artifact.get("write_permission") is not False:
        raise AssertionError("grounded artifact violates the read-only human-review contract")

    report = first_json(run_data, "Report Persisted M07 Proposal")
    proposal_id = report.get("proposal_id", "")
    record_id = report.get("record_id", "")
    if report.get("status") != "ACK" or report.get("execution_permitted") is not False:
        raise AssertionError("proposal report is not a read-only ACK")
    if not isinstance(proposal_id, str) or not proposal_id.startswith("sha256:") or not record_id:
        raise AssertionError("proposal report is missing canonical IDs")
    proposal_path = args.proposal_store / f"{proposal_id.removeprefix('sha256:')}.json"
    if not proposal_path.is_file():
        raise AssertionError(f"persisted proposal is missing: {proposal_path}")
    proposal = json.loads(proposal_path.read_text(encoding="utf-8"))
    if proposal.get("proposal_id") != proposal_id or proposal.get("record_id") != record_id:
        raise AssertionError("persisted proposal does not match n8n ACK")
    print(f"N8N M07 OPERATED EXECUTION PASS: {proposal_id} persisted read-only for {record_id}")


if __name__ == "__main__":
    main()
