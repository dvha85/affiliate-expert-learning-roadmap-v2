"""Run offline M06/M07 engine paths through a real n8n CLI engine.

This script deliberately creates a disposable n8n SQLite runtime.  The checked-in
blueprints are portable exports and have no instance-specific workflow ID, so the
copy imported into that disposable runtime receives an ID and a loopback adapter
URL. It never edits the checked-in blueprint, imports a user credential, or calls
an affiliate/network endpoint. The M07 happy path imports only a disposable
OpenAI-compatible credential that is pinned to an in-process loopback model stub.
"""
import argparse
import http.server
import json
import os
import re
import shutil
import socket
import subprocess
import sys
import tempfile
import threading
import time
import urllib.request
from pathlib import Path
from typing import Optional

from validate_n8n_m06_operated_execution import validate_rejection as validate_m06_operated_rejection
from validate_n8n_m06_operated_execution import validate_success as validate_m06_operated_success


ROOT = Path(__file__).resolve().parents[1]
BOT_DIR = ROOT / "lab" / "affiliate-bot"
M06_BLUEPRINT = ROOT / "lab" / "n8n" / "M06-readonly-watcher.blueprint.json"
M07_BLUEPRINT = ROOT / "lab" / "n8n" / "M07-readonly-evidence-agent.blueprint.json"


def m07_model_output_from_prompt(payload: dict) -> dict:
    messages = payload.get("messages")
    if not isinstance(messages, list):
        raise ValueError("model stub received no messages")
    text_parts = []
    for message in messages:
        if not isinstance(message, dict):
            continue
        content = message.get("content")
        if isinstance(content, str):
            text_parts.append(content)
        elif isinstance(content, list):
            text_parts.extend(str(part.get("text", "")) for part in content if isinstance(part, dict))
    prompt = "\n".join(text_parts)
    if "canonical_context=" not in prompt or "\nregistered_tool_evidence=" not in prompt:
        raise ValueError("model stub did not receive canonical and registered-tool context")
    context_raw = prompt.split("canonical_context=", 1)[1].split("\nregistered_tool_evidence=", 1)[0]
    context = json.loads(context_raw)
    evidence = context.get("evidence")
    if not isinstance(evidence, list):
        raise ValueError("model stub received invalid canonical evidence")
    scalar = next((item for item in evidence if isinstance(item, dict) and item.get("field_or_claim") == "price" and isinstance(item.get("evidence_id"), str)), None)
    if scalar is None:
        raise ValueError("model stub could not find canonical price evidence")
    value = scalar.get("value")
    value_json = json.dumps(value, ensure_ascii=False, separators=(",", ":"))
    claim_text = f"price={value_json} [evidence:{scalar['evidence_id']}]"
    return {
        "state": "HUMAN_REVIEW",
        "answer": claim_text,
        "claims": [{"text": claim_text, "field_or_claim": "price", "value": value, "evidence_ids": [scalar["evidence_id"]]}],
        "evidence_ids": [scalar["evidence_id"]],
        "tool_calls": [],
        "authority": "A2-RO",
        "write_permission": False,
        "proposed_action": {"action_type": "DRAFT", "target": "human_review", "parameters": {}},
    }


class m07ModelStubHandler(http.server.BaseHTTPRequestHandler):
    def log_message(self, format: str, *args: object) -> None:
        return

    def _write_json(self, status: int, payload: dict) -> None:
        raw = json.dumps(payload, separators=(",", ":")).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(raw)))
        self.end_headers()
        self.wfile.write(raw)

    def do_GET(self) -> None:
        if self.path == "/v1/models":
            self._write_json(200, {"object": "list", "data": [{"id": "m07-ci-stub", "object": "model"}]})
            return
        self._write_json(404, {"error": {"message": "unsupported model-stub path"}})

    def do_POST(self) -> None:
        if self.path != "/v1/chat/completions":
            self._write_json(404, {"error": {"message": "unsupported model-stub path"}})
            return
        try:
            content_length = int(self.headers.get("Content-Length", "0"))
            if content_length <= 0 or content_length > 1 << 20:
                raise ValueError("invalid model-stub body length")
            request = json.loads(self.rfile.read(content_length))
            output = m07_model_output_from_prompt(request)
        except (ValueError, json.JSONDecodeError) as error:
            self.server.stub_error = str(error)  # type: ignore[attr-defined]
            self._write_json(400, {"error": {"message": "invalid model-stub request"}})
            return
        self.server.request_count += 1  # type: ignore[attr-defined]
        self._write_json(200, {"id": "chatcmpl-m07-ci", "object": "chat.completion", "created": 0, "model": "m07-ci-stub", "choices": [{"index": 0, "message": {"role": "assistant", "content": json.dumps(output, separators=(",", ":"))}, "finish_reason": "stop"}], "usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}})


def start_m07_model_stub(port: int) -> tuple[http.server.ThreadingHTTPServer, threading.Thread]:
    server = http.server.ThreadingHTTPServer(("127.0.0.1", port), m07ModelStubHandler)
    server.request_count = 0  # type: ignore[attr-defined]
    server.stub_error = ""  # type: ignore[attr-defined]
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    return server, thread


def stop_m07_model_stub(server: http.server.ThreadingHTTPServer, thread: threading.Thread) -> None:
    server.shutdown()
    server.server_close()
    thread.join(timeout=10)
    if thread.is_alive():
        raise AssertionError("M07 model stub did not stop")


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


def import_workflow(prefix: list[str], env: dict[str, str], blueprint: Path, work: Path, workflow_id: str, adapter_port: int, *, m06_fixture: Optional[dict] = None, m07_case: Optional[str] = None, m07_record_id: Optional[str] = None, m07_model_stub_port: Optional[int] = None) -> None:
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
                if m07_case == "model-success" and assignment["name"] == "tool_request_json":
                    if m07_record_id is None:
                        raise AssertionError("M07 model-success requires a canonical record")
                    assignment["value"] = json.dumps({"record_id": m07_record_id, "tool_call": {"tool_name": "public_http", "method": "GET", "target": "https://example.com/br13/offer"}, "status_code": 200, "received_at": "2026-09-09T00:00:00Z", "redirected": False, "body": {"notice": "synthetic CI fixture; untrusted data"}}, separators=(",", ":"))
                if m07_case == "redirect-registry" and assignment["name"] == "tool_registry_json":
                    assignment["value"] = '[{"name":"public_http","read_only":true,"allowed_methods":["GET"],"allowed_hosts":["example.com"],"timeout_ms":10000,"follow_redirects":true}]'
        if m07_case == "model-success" and node["name"] == "Fetch and Register Tool Adapter":
            node["parameters"]["url"] = "={{ $('M07 Adapter Input').item.json.adapter_url + '/v1/m07/register-tool-result' }}"
            node["parameters"]["jsonBody"] = "={{ {record_id:$('M07 Adapter Input').item.json.record_id,registry:JSON.parse($('M07 Adapter Input').item.json.tool_registry_json),tool_result:JSON.parse($('M07 Adapter Input').item.json.tool_request_json)} }}"
        if m07_case == "model-success" and node["name"] == "OpenAI Chat Model - configure credential locally":
            if m07_model_stub_port is None:
                raise AssertionError("M07 model-success requires a loopback model stub")
            node["credentials"] = {"openAiApi": {"id": "rp08-m07-model-stub", "name": "M07 CI Loopback Model Stub"}}
            node["parameters"] = {"model": {"__rl": True, "value": "m07-ci-stub", "mode": "list", "cachedResultName": "m07-ci-stub"}, "options": {"maxRetries": 0}}
    imported = work / f"{workflow_id}.json"
    imported.write_text(json.dumps(data), encoding="utf-8")
    run(prefix + ["import:workflow", f"--input={imported}"], env=env)


def import_m07_model_stub_credential(prefix: list[str], env: dict[str, str], work: Path, port: int) -> None:
    credential = [{"id": "rp08-m07-model-stub", "name": "M07 CI Loopback Model Stub", "type": "openAiApi", "data": {"apiKey": "m07-ci-loopback-not-a-secret", "url": f"http://127.0.0.1:{port}/v1"}}]
    source = work / "rp08-m07-model-stub-credential.json"
    source.write_text(json.dumps(credential), encoding="utf-8")
    run(prefix + ["import:credentials", f"--input={source}"], env=env)


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
    if report.get("canonical_history_persisted") is not True or report.get("execution_permitted") is not False or not report.get("record_id"):
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


def require_m07_model_success(execution: dict, expected_record_id: str) -> tuple[str, str]:
    if execution.get("status") != "success" or execution.get("finished") is not True:
        result = execution.get("data", {}).get("resultData", {})
        raise AssertionError(f"M07 model-success did not finish: last_node={result.get('lastNodeExecuted')!r} error={result.get('error')!r}")
    for node in ("Fetch and Register Tool Adapter", "Canonical M07 Context Adapter", "Read-only Evidence Agent", "Validate Grounding Adapter", "Persist Agent Proposal Adapter", "Report Persisted M07 Proposal"):
        node_json(execution, node)
    registered_tool = node_json(execution, "Fetch and Register Tool Adapter").get("body", {})
    tool_result_id = registered_tool.get("artifact_id") if isinstance(registered_tool, dict) else None
    if not isinstance(tool_result_id, str) or not tool_result_id.startswith("sha256:"):
        raise AssertionError("M07 model-success did not register a canonical tool trace")
    report = node_json(execution, "Report Persisted M07 Proposal")
    if report.get("status") != "ACK" or report.get("record_id") != expected_record_id or report.get("execution_permitted") is not False:
        raise AssertionError("M07 model-success did not return a read-only persisted proposal ACK")
    proposal_id = report.get("proposal_id")
    if not isinstance(proposal_id, str) or not proposal_id.startswith("sha256:"):
        raise AssertionError("M07 model-success did not return a canonical proposal ID")
    return proposal_id, tool_result_id


def post_json(url: str, payload: dict) -> tuple[int, dict]:
    raw = json.dumps(payload, separators=(",", ":")).encode("utf-8")
    request = urllib.request.Request(url, data=raw, headers={"Content-Type": "application/json"}, method="POST")
    with urllib.request.urlopen(request, timeout=10) as response:
        return response.status, json.loads(response.read())


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
            validate_m06_operated_success(first, history, expected_result="APPENDED")
            second = execute(prefix, env, "rp08-m06")
            if require_m06_success(second, "EXACT_DUPLICATE") != record_id:
                raise AssertionError("M06 duplicate did not resolve the original canonical record")
            validate_m06_operated_success(second, history, expected_result="EXACT_DUPLICATE")
            reordered = m06_fixture(body='{"commission_rate":0.08,"price":100,"currency":"USD","product_name":"Fixture A","product_id":"a"}')
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-reordered", port, m06_fixture=reordered)
            if require_m06_success(execute(prefix, env, "rp08-m06-reordered"), "EXACT_DUPLICATE") != record_id:
                raise AssertionError("M06 JSON key reordering did not resolve the original canonical record")
            history_before_rejections = history.read_bytes()
            changed_same_event = m06_fixture(body='{"product_id":"a","product_name":"Fixture A","currency":"USD","price":120,"commission_rate":0.08}')
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-conflict", port, m06_fixture=changed_same_event)
            conflict = execute(prefix, env, "rp08-m06-conflict", expected=1)
            require_m06_rejection(conflict)
            validate_m06_operated_rejection(conflict, history, expected_record_count=1)
            if history.read_bytes() != history_before_rejections:
                raise AssertionError("M06 conflicting content changed canonical history")
            bad_source = m06_fixture(url="https://other.invalid/br13/offer")
            import_workflow(prefix, env, M06_BLUEPRINT, runtime, "rp08-m06-source-reject", port, m06_fixture=bad_source)
            source_rejection = execute(prefix, env, "rp08-m06-source-reject", expected=1)
            require_m06_rejection(source_rejection)
            validate_m06_operated_rejection(source_rejection, history, expected_record_count=1)
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
            adapter = start_adapter(bot, history, port, env)
            model_stub_port = choose_port()
            model_stub, model_stub_thread = start_m07_model_stub(model_stub_port)
            try:
                import_m07_model_stub_credential(prefix, env, runtime, model_stub_port)
                import_workflow(prefix, env, M07_BLUEPRINT, runtime, "rp08-m07-model-success", port, m07_case="model-success", m07_record_id=record_id, m07_model_stub_port=model_stub_port)
                proposal_id, tool_result_id = require_m07_model_success(execute(prefix, env, "rp08-m07-model-success"), record_id)
                if model_stub.request_count < 1 or model_stub.stub_error:  # type: ignore[attr-defined]
                    raise AssertionError("M07 Agent did not receive the canonical prompt from the loopback model stub")
                proposal_path = history.with_name(history.name + ".m07") / "proposals" / (proposal_id.removeprefix("sha256:") + ".json")
                if not proposal_path.is_file():
                    raise AssertionError("M07 model-success reported a proposal that was not persisted")
                stop_adapter(adapter)
                adapter = None
                adapter = start_adapter(bot, history, port, env)
                replay = run([str(bot), "history", "replay", str(history)], env=env)
                if "replay=MATCH" not in replay.stdout or not proposal_path.is_file():
                    raise AssertionError("M07 persisted proposal or canonical history did not survive adapter restart")
                proposal = json.loads(proposal_path.read_text(encoding="utf-8"))
                status, validated = post_json(f"http://127.0.0.1:{port}/v1/m07/validate", {"record_id": record_id, "registry": [{"name": "public_http", "read_only": True, "allowed_methods": ["GET"], "allowed_hosts": ["example.com"], "timeout_ms": 10000, "follow_redirects": False}], "model_output_text": json.dumps(proposal["raw_output"], separators=(",", ":")), "tool_result_id": tool_result_id})
                if status != 200 or validated.get("status") != "VALID" or validated.get("execution_permitted") is not False:
                    raise AssertionError("M07 persisted proposal did not revalidate after adapter restart")
            finally:
                stop_m07_model_stub(model_stub, model_stub_thread)
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
    print("N8N ENGINE REGRESSION PASS: M06 persisted/replayed via n8n with key-order retry, changed-event append, conflict/source rejection and sink failure; M07 policy rejections failed closed and the real Agent completed grounding/proposal persistence against a loopback OpenAI-compatible model stub")


if __name__ == "__main__":
    main()
