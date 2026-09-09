"""Run offline M06/M07 rejection paths through a real n8n CLI engine.

This script deliberately creates a disposable n8n SQLite runtime.  The checked-in
blueprints are portable exports and have no instance-specific workflow ID, so the
copy imported into that disposable runtime receives an ID and a loopback adapter
URL.  It never edits the checked-in blueprint, imports a user credential, or calls
an affiliate/network endpoint.
"""
import argparse
import json
import os
import re
import shutil
import socket
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import Optional


ROOT = Path(__file__).resolve().parents[1]
BOT_DIR = ROOT / "lab" / "affiliate-bot"
M06_BLUEPRINT = ROOT / "lab" / "n8n" / "M06-readonly-watcher.blueprint.json"
M07_BLUEPRINT = ROOT / "lab" / "n8n" / "M07-readonly-evidence-agent.blueprint.json"


def command_prefix(args: argparse.Namespace) -> list[str]:
    if args.n8n_node:
        if not args.n8n_cli:
            raise AssertionError("--n8n-node requires --n8n-cli")
        return [args.n8n_node, args.n8n_cli]
    if args.n8n_cli:
        return [args.n8n_cli]
    return ["n8n"]


def choose_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
        listener.bind(("127.0.0.1", 0))
        return int(listener.getsockname()[1])


def run(command: list[str], *, env: dict[str, str], cwd: Path = ROOT, expected: int = 0) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True)
    if result.returncode != expected:
        raise AssertionError(
            "command failed:\n"
            + " ".join(command)
            + f"\nexit={result.returncode}\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}"
        )
    return result


def wait_for_adapter(port: int, process: subprocess.Popen[str]) -> None:
    for _ in range(100):
        if process.poll() is not None:
            stdout, stderr = process.communicate()
            raise AssertionError(f"canonical adapter stopped early:\n{stdout}\n{stderr}")
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return
        except OSError:
            time.sleep(0.05)
    raise AssertionError("canonical adapter did not start on loopback")


def start_adapter(bot: Path, history: Path, port: int, env: dict[str, str]) -> subprocess.Popen[str]:
    process = subprocess.Popen(
        [str(bot), "watcher", "serve", str(history), f"127.0.0.1:{port}"],
        cwd=ROOT,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    wait_for_adapter(port, process)
    return process


def stop_adapter(process: subprocess.Popen[str]) -> None:
    if process.poll() is None:
        process.terminate()
    try:
        process.communicate(timeout=10)
    except subprocess.TimeoutExpired as error:
        process.kill()
        process.communicate(timeout=10)
        raise AssertionError("canonical adapter did not stop") from error


def json_from_n8n(raw: str) -> dict:
    decoder = json.JSONDecoder()
    starts = [match.start() for match in re.finditer(r'(?m)^\{\s*"data"\s*:', raw)]
    for start in reversed(starts):
        try:
            value, _ = decoder.raw_decode(raw[start:])
        except json.JSONDecodeError:
            continue
        if isinstance(value, dict) and "data" in value:
            return value
    raise AssertionError(f"no n8n execution JSON found:\n{raw}")


def node_json(execution: dict, node: str) -> dict:
    runs = execution.get("data", {}).get("resultData", {}).get("runData", {}).get(node)
    if not isinstance(runs, list) or not runs or runs[0].get("executionStatus") != "success":
        raise AssertionError(f"n8n node did not succeed: {node}")
    try:
        return runs[0]["data"]["main"][0][0]["json"]
    except (KeyError, IndexError, TypeError) as error:
        raise AssertionError(f"n8n node has no JSON output: {node}") from error


def import_workflow(prefix: list[str], env: dict[str, str], blueprint: Path, work: Path, workflow_id: str, adapter_port: int, *, m06_fixture: Optional[dict] = None, m07_case: Optional[str] = None, m07_record_id: Optional[str] = None) -> None:
    data = json.loads(blueprint.read_text(encoding="utf-8"))
    data["id"] = workflow_id
    for node in data["nodes"]:
        # n8n's non-interactive `execute` command starts only at this trigger.
        # The production blueprint keeps its reviewed Schedule/Manual trigger;
        # only this disposable import substitutes the CLI entrypoint.
        if node["name"] in {"Schedule Trigger", "Manual Trigger"}:
            node["type"] = "n8n-nodes-base.executeWorkflowTrigger"
            node["typeVersion"] = 1
            node["parameters"] = {}
        if node["name"] == "M06 Adapter Input" or node["name"] == "M07 Adapter Input":
            for assignment in node["parameters"]["assignments"]["assignments"]:
                if assignment["name"] == "adapter_url":
                    assignment["value"] = f"http://127.0.0.1:{adapter_port}"
                if m06_fixture is not None and assignment["name"] == "fixture_json":
                    assignment["value"] = json.dumps(m06_fixture, separators=(",", ":"), ensure_ascii=False)
                if m07_record_id is not None and assignment["name"] == "record_id":
                    assignment["value"] = m07_record_id
                if m07_case == "post" and assignment["name"] == "tool_request_json":
                    assignment["value"] = '{"tool_name":"public_http","method":"POST","target":"https://example.com/br13/offer"}'
                if m07_case in {"get", "redirect-registry"} and assignment["name"] == "tool_request_json":
                    assignment["value"] = '{"tool_name":"public_http","method":"GET","target":"https://example.com/br13/offer"}'
                if m07_case == "redirect-registry" and assignment["name"] == "tool_registry_json":
                    assignment["value"] = '[{"name":"public_http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":10000,"follow_redirects":true}]'
    imported = work / f"{workflow_id}.json"
    imported.write_text(json.dumps(data), encoding="utf-8")
    run(prefix + ["import:workflow", f"--input={imported}"], env=env)


def execute(prefix: list[str], env: dict[str, str], workflow_id: str, *, expected: int = 0) -> dict:
    result = run(prefix + ["execute", f"--id={workflow_id}", "--rawOutput"], env=env, expected=expected)
    execution = json_from_n8n(result.stdout)
    if "status" not in execution:
        raise AssertionError(f"n8n raw output has no execution status:\n{result.stdout}")
    return execution


def require_m06_success(execution: dict, expected_status: str) -> str:
    if execution.get("status") != "success" or execution.get("finished") is not True:
        result = execution.get("data", {}).get("resultData", {})
        raise AssertionError(
            "M06 n8n execution did not finish successfully: "
            f"status={execution.get('status')!r} finished={execution.get('finished')!r} "
            f"last_node={result.get('lastNodeExecuted')!r} error={result.get('error')!r}"
        )
    report = node_json(execution, "Report Canonical M06 Result")
    if report.get("result") != expected_status or report.get("canonical_history_handoff") != "ACK":
        raise AssertionError(f"M06 did not return {expected_status} canonical ACK")
    if report.get("canonical_history_persisted") is not True or not report.get("record_id"):
        raise AssertionError("M06 reported persistence without a canonical record ID")
    return str(report["record_id"])


def require_m06_sink_failure(execution: dict) -> None:
    result = execution.get("data", {}).get("resultData", {})
    if execution.get("status") != "error" or result.get("lastNodeExecuted") != "Build and Append Canonical M06 Adapter":
        raise AssertionError("M06 sink failure did not stop at the adapter request")
    if "Report Canonical M06 Result" in result.get("runData", {}):
        raise AssertionError("M06 sink failure reached a persistence report")


def require_m06_rejection(execution: dict) -> None:
    result = execution.get("data", {}).get("resultData", {})
    if execution.get("status") != "error" or result.get("lastNodeExecuted") != "Build and Append Canonical M06 Adapter":
        raise AssertionError("M06 fixture rejection did not stop at the canonical adapter")
    if "Require Canonical Store ACK" in result.get("runData", {}) or "Report Canonical M06 Result" in result.get("runData", {}):
        raise AssertionError("M06 fixture rejection reached an ACK or persistence report")


def m06_fixture(*, correlation_id: str = "event-1", body: str = '{"product_id":"a","product_name":"Fixture A","currency":"USD","price":100,"commission_rate":0.08}', url: str = "https://example.com/br13/offer") -> dict:
    return {
        "version": "br13-offer-fixture/v1",
        "method": "GET",
        "url": url,
        "observed_at": "2026-09-03T00:00:00Z",
        "correlation_id": correlation_id,
        "status_code": 200,
        "body": body,
    }


def require_m07_rejection(execution: dict, proposal_store: Path, marker: Optional[str] = None) -> None:
    result = execution.get("data", {}).get("resultData", {})
    if execution.get("status") != "error" or result.get("lastNodeExecuted") != "Fetch and Register Tool Adapter":
        raise AssertionError("M07 POST request was not rejected at the policy adapter")
    if marker is not None and marker not in json.dumps(result, separators=(",", ":")):
        raise AssertionError(f"M07 fetch rejection did not contain expected marker {marker}")
    for forbidden in ("Read-only Evidence Agent", "Persist Agent Proposal Adapter", "Report Persisted M07 Proposal"):
        if forbidden in result.get("runData", {}):
            raise AssertionError(f"policy-rejected M07 request reached {forbidden}")
    if proposal_store.exists() and any(proposal_store.glob("*.json")):
        raise AssertionError("policy-rejected M07 request persisted a proposal")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--n8n-cli", help="n8n executable or CLI JavaScript entrypoint")
    parser.add_argument("--n8n-node", help="Node executable when --n8n-cli is a JavaScript entrypoint")
    parser.add_argument("--keep-runtime", action="store_true", help="preserve the disposable runtime after a successful run too")
    args = parser.parse_args()
    prefix = command_prefix(args)
    if shutil.which(prefix[0]) is None and not Path(prefix[0]).is_file():
        raise SystemExit(f"n8n command is unavailable: {prefix[0]}")
    if len(prefix) == 2 and not Path(prefix[1]).is_file():
        raise SystemExit(f"n8n CLI entrypoint is unavailable: {prefix[1]}")

    runtime = Path(tempfile.mkdtemp(prefix="affiliate-n8n-engine-"))
    keep = args.keep_runtime
    try:
        n8n_home = runtime / "n8n"
        env = dict(os.environ)
        env.update(
            {
                "N8N_USER_FOLDER": str(n8n_home),
                "N8N_ENCRYPTION_KEY": "n8n-ci-isolated-fixture-key-not-a-secret",
                "N8N_DIAGNOSTICS_ENABLED": "false",
                "N8N_PERSONALIZATION_ENABLED": "false",
                "N8N_ENFORCE_SETTINGS_FILE_PERMISSIONS": "false",
                "N8N_RUNNERS_BROKER_PORT": str(choose_port()),
                "GOWORK": "off",
                "GOCACHE": str(runtime / "go-cache"),
            }
        )
        if args.n8n_node:
            env["PATH"] = str(Path(args.n8n_node).resolve().parent) + os.pathsep + env.get("PATH", "")
        bot = runtime / "bot"
        run(["go", "build", "-o", str(bot), "./cmd/bot"], env=env, cwd=BOT_DIR)
        history = runtime / "history.jsonl"
        port = choose_port()
        adapter = start_adapter(bot, history, port, env)
        try:
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06", port)
            first = execute(prefix, env, "rp08-m06")
            record_id = require_m06_success(first, "APPENDED")
            second = execute(prefix, env, "rp08-m06")
            if require_m06_success(second, "EXACT_DUPLICATE") != record_id:
                raise AssertionError("M06 duplicate did not resolve the original canonical record")
            reordered = m06_fixture(body='{"commission_rate":0.08,"price":100,"currency":"USD","product_name":"Fixture A","product_id":"a"}')
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-reordered", port, m06_fixture=reordered)
            if require_m06_success(execute(prefix, env, "rp08-m06-reordered"), "EXACT_DUPLICATE") != record_id:
                raise AssertionError("M06 JSON key reordering did not resolve the original canonical record")
            history_before_rejections = history.read_bytes()
            changed_same_event = m06_fixture(body='{"product_id":"a","product_name":"Fixture A","currency":"USD","price":120,"commission_rate":0.08}')
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-conflict", port, m06_fixture=changed_same_event)
            require_m06_rejection(execute(prefix, env, "rp08-m06-conflict", expected=1))
            if history.read_bytes() != history_before_rejections:
                raise AssertionError("M06 conflicting content changed canonical history")
            bad_source = m06_fixture(url="https://other.invalid/br13/offer")
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-source-reject", port, m06_fixture=bad_source)
            require_m06_rejection(execute(prefix, env, "rp08-m06-source-reject", expected=1))
            if history.read_bytes() != history_before_rejections:
                raise AssertionError("M06 rejected source changed canonical history")
            changed_new_event = m06_fixture(correlation_id="event-2", body=changed_same_event["body"])
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-changed", port, m06_fixture=changed_new_event)
            changed_record_id = require_m06_success(execute(prefix, env, "rp08-m06-changed"), "APPENDED")
            if changed_record_id == record_id:
                raise AssertionError("M06 changed event reused the original canonical record")
            replay = run([str(bot), "history", "replay", str(history)], env=env)
            if "replay=MATCH" not in replay.stdout:
                raise AssertionError("canonical history did not replay after n8n M06 execution")
            stop_adapter(adapter)
            adapter = None
            require_m06_sink_failure(execute(prefix, env, "rp08-m06", expected=1))
            adapter = start_adapter(bot, history, port, env)
            run([str(bot), "history", "replay", str(history)], env=env)
            proposal_store = history.with_name(history.name + ".m07") / "proposals"
            import_workflow(prefix, env, M07_BLUEPRINT, runtime, "rp08-m07-post", port, m07_case="post", m07_record_id=record_id)
            require_m07_rejection(execute(prefix, env, "rp08-m07-post", expected=1), proposal_store, "TOOL_TRANSPORT_REJECTED")
            import_workflow(prefix, env, M07_BLUEPRINT, runtime, "rp08-m07-redirect", port, m07_case="redirect-registry", m07_record_id=record_id)
            require_m07_rejection(execute(prefix, env, "rp08-m07-redirect", expected=1), proposal_store, "INVALID_REQUEST")
            stop_adapter(adapter)
            adapter = None
            import_workflow(prefix, env, M07_BLUEPRINT, runtime, "rp08-m07-sink", port, m07_case="get", m07_record_id=record_id)
            require_m07_rejection(execute(prefix, env, "rp08-m07-sink", expected=1), proposal_store)
        finally:
            if adapter is not None:
                stop_adapter(adapter)
    except BaseException:
        keep = keep or True
        raise
    finally:
        if keep:
            print(f"N8N engine regression runtime retained at {runtime}", file=sys.stderr)
        else:
            shutil.rmtree(runtime)
    print("N8N ENGINE REGRESSION PASS: M06 persisted/replayed via n8n with key-order retry, changed-event append, conflict/source rejection and sink failure; M07 POST, redirect-registry, and unavailable-adapter paths failed closed")


if __name__ == "__main__":
    main()
