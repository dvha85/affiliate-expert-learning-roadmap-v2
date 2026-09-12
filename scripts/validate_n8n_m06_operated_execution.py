"""Validate captured M06 n8n executions against canonical history without credentials."""
import argparse
import json
import re
from pathlib import Path
from typing import Optional


REQUIRED_NODES = (
    "Build and Append Canonical M06 Adapter",
    "Require Canonical Store ACK",
    "Report Canonical M06 Result",
)


def load_execution(path: Path) -> dict:
    raw = path.read_text(encoding="utf-8")
    for match in reversed(list(re.finditer(r'(?m)^\{\s*"data"\s*:', raw))):
        try:
            value, _ = json.JSONDecoder().raw_decode(raw[match.start() :])
        except json.JSONDecodeError:
            continue
        if isinstance(value, dict):
            return value
    raise AssertionError("no n8n execution JSON found; pass --rawOutput captured from n8n execute")


def first_json(run_data: dict, name: str) -> dict:
    runs = run_data.get(name)
    if not isinstance(runs, list) or not runs or runs[0].get("executionStatus") != "success":
        raise AssertionError(f"{name} did not succeed")
    try:
        value = runs[0]["data"]["main"][0][0]["json"]
    except (KeyError, IndexError, TypeError) as error:
        raise AssertionError(f"{name} has no usable JSON output") from error
    if not isinstance(value, dict):
        raise AssertionError(f"{name} has non-object JSON output")
    return value


def history_records(history: Path) -> list[dict]:
    if not history.is_file():
        raise AssertionError(f"canonical history is missing: {history}")
    records = []
    for line in history.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        value = json.loads(line)
        if not isinstance(value, dict) or not isinstance(value.get("record_id"), str):
            raise AssertionError("canonical history contains an invalid record")
        records.append(value)
    return records


def validate_success(execution: dict, history: Path, *, expected_result: Optional[str] = None) -> str:
    if execution.get("status") != "success" or execution.get("finished") is not True:
        raise AssertionError("n8n execution did not finish successfully")
    result_data = execution.get("data", {}).get("resultData", {})
    run_data = result_data.get("runData", {})
    if not isinstance(run_data, dict):
        raise AssertionError("n8n execution has no run data")
    for name in REQUIRED_NODES:
        first_json(run_data, name)
    handoff = first_json(run_data, "Build and Append Canonical M06 Adapter").get("body", {})
    ack = first_json(run_data, "Require Canonical Store ACK")
    report = first_json(run_data, "Report Canonical M06 Result")
    if not isinstance(handoff, dict):
        raise AssertionError("M06 adapter did not return a response object")
    result = report.get("result")
    if result not in {"APPENDED", "EXACT_DUPLICATE"} or (expected_result is not None and result != expected_result):
        raise AssertionError("M06 final report has an unexpected canonical result")
    record_id = report.get("record_id")
    if not isinstance(record_id, str) or not record_id:
        raise AssertionError("M06 final report is missing a record ID")
    for value in (handoff, ack, report):
        if value.get("record_id") != record_id:
            raise AssertionError("M06 adapter ACK chain does not bind one record ID")
    if handoff.get("status") != result or ack.get("status") != result:
        raise AssertionError("M06 ACK chain does not preserve adapter status")
    if ack.get("canonical_history_handoff") != "ACK" or report.get("canonical_history_handoff") != "ACK":
        raise AssertionError("M06 report did not retain canonical history ACK")
    if ack.get("canonical_history_persisted") is not True or report.get("canonical_history_persisted") is not True:
        raise AssertionError("M06 reported a non-persisted canonical history result")
    if ack.get("execution_permitted") is not False or report.get("execution_permitted") is not False:
        raise AssertionError("M06 output violated read-only execution boundary")
    if record_id not in {record["record_id"] for record in history_records(history)}:
        raise AssertionError("M06 ACK record does not resolve from canonical history")
    return record_id


def validate_rejection(execution: dict, history: Path, *, expected_record_count: int) -> None:
    result_data = execution.get("data", {}).get("resultData", {})
    run_data = result_data.get("runData", {})
    if execution.get("status") != "error" or result_data.get("lastNodeExecuted") != "Build and Append Canonical M06 Adapter":
        raise AssertionError("M06 rejected execution did not stop at the canonical adapter")
    if not isinstance(run_data, dict):
        raise AssertionError("M06 rejected execution has no run data")
    for forbidden in ("Require Canonical Store ACK", "Report Canonical M06 Result"):
        if forbidden in run_data:
            raise AssertionError(f"M06 rejection reached forbidden node: {forbidden}")
    if len(history_records(history)) != expected_record_count:
        raise AssertionError("M06 rejected execution changed canonical history inventory")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("execution_json", type=Path)
    parser.add_argument("--history", type=Path, required=True)
    parser.add_argument("--expect-reject", action="store_true")
    parser.add_argument("--expected-record-count", type=int)
    parser.add_argument("--expected-result", choices=("APPENDED", "EXACT_DUPLICATE"))
    args = parser.parse_args()
    execution = load_execution(args.execution_json)
    if args.expect_reject:
        if args.expected_record_count is None:
            raise AssertionError("rejection validation requires --expected-record-count")
        validate_rejection(execution, args.history, expected_record_count=args.expected_record_count)
        print("N8N M06 OPERATED REJECT PASS: adapter rejected before canonical ACK or report")
        return
    record_id = validate_success(execution, args.history, expected_result=args.expected_result)
    print(f"N8N M06 OPERATED EXECUTION PASS: canonical read-only ACK resolved {record_id}")


if __name__ == "__main__":
    main()
