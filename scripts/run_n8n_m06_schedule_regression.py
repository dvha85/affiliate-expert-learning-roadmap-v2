"""Exercise M06 through n8n's real Schedule Trigger in an isolated runtime.

The existing engine regression uses n8n's ``execute`` command, whose entrypoint
requires an Execute Workflow Trigger.  This runner keeps the reviewed M06
Schedule Trigger, starts a disposable n8n server, and proves both scheduled
admission and fail-closed adapter behaviour.  It uses only the checked-in
synthetic fixture and a loopback canonical-store adapter.
"""
import argparse
import json
import os
import shutil
import socket
import sqlite3
import subprocess
import sys
import tempfile
import time
import uuid
from pathlib import Path
from typing import Optional

from run_n8n_engine_regression import (
    BOT_DIR,
    M06_BLUEPRINT,
    ROOT,
    choose_port,
    command_prefix,
    node_json,
    run,
    start_adapter,
    stop_adapter,
)
from validate_n8n_m06_operated_execution import validate_success


def import_scheduled_workflow(prefix: list[str], env: dict[str, str], runtime: Path, workflow_id: str, adapter_port: int) -> None:
    """Import a one-second Schedule Trigger copy without changing the blueprint."""
    workflow = json.loads(M06_BLUEPRINT.read_text(encoding="utf-8"))
    workflow["id"] = workflow_id
    workflow["name"] = f"M06 schedule fixture {workflow_id} (temporary)"
    for node in workflow["nodes"]:
        if node["name"] == "Schedule Trigger":
            node["parameters"] = {"rule": {"interval": [{"field": "seconds", "secondsInterval": 1}]}}
        if node["name"] == "M06 Adapter Input":
            for assignment in node["parameters"]["assignments"]["assignments"]:
                if assignment["name"] == "adapter_url":
                    assignment["value"] = f"http://127.0.0.1:{adapter_port}"
    source = runtime / f"{workflow_id}.json"
    source.write_text(json.dumps(workflow), encoding="utf-8")
    run(prefix + ["import:workflow", f"--input={source}"], env=env)


def database_path(n8n_home: Path) -> Path:
    # n8n appends .n8n to N8N_USER_FOLDER; wait because CLI import creates it.
    result = n8n_home / ".n8n" / "database.sqlite"
    if not result.is_file():
        raise AssertionError(f"n8n import did not create expected database: {result}")
    return result


def activate_schedule_workflow(database: Path, workflow_id: str) -> None:
    """Give the imported workflow a valid active-version history graph.

    The CLI importer deliberately imports inactive workflows.  n8n requires an
    activeVersionId that points to workflow_history before it will admit a
    Schedule Trigger at startup.  This only mutates the throwaway SQLite file;
    the import/export blueprint remains inactive and portable.
    """
    connection = sqlite3.connect(database)
    try:
        row = connection.execute(
            "SELECT nodes, connections, name, description, nodeGroups FROM workflow_entity WHERE id = ?", (workflow_id,)
        ).fetchone()
        if row is None:
            raise AssertionError("scheduled workflow was not imported")
        nodes, connections, name, description, node_groups = row
        version_id = str(uuid.uuid4())
        # workflow_entity.activeVersionId and workflow_history.workflowId form a
        # circular FK.  Populate their one-time graph atomically in the isolated
        # database, then turn FK validation back on before the server starts.
        connection.execute("PRAGMA foreign_keys = OFF")
        connection.execute("BEGIN")
        connection.execute(
            "INSERT INTO workflow_history (versionId, workflowId, authors, nodes, connections, name, description, nodeGroups) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
            (version_id, workflow_id, "[]", nodes, connections, name, description, node_groups or "[]"),
        )
        connection.execute(
            "UPDATE workflow_entity SET active = 1, activeVersionId = ?, updatedAt = STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW') WHERE id = ?",
            (version_id, workflow_id),
        )
        connection.commit()
        connection.execute("PRAGMA foreign_keys = ON")
        violations = connection.execute("PRAGMA foreign_key_check").fetchall()
        if violations:
            raise AssertionError(f"isolated n8n activation graph has foreign-key violations: {violations}")
    finally:
        connection.close()


def start_n8n(prefix: list[str], env: dict[str, str], port: int, runtime: Path) -> tuple[subprocess.Popen[str], Path]:
    log = runtime / "n8n-server.log"
    handle = log.open("w", encoding="utf-8")
    child_env = dict(env)
    child_env.update({"N8N_PORT": str(port), "N8N_LISTEN_ADDRESS": "127.0.0.1", "N8N_HOST": "127.0.0.1", "N8N_SECURE_COOKIE": "false"})
    process = subprocess.Popen(prefix + ["start"], cwd=ROOT, env=child_env, text=True, stdout=handle, stderr=subprocess.STDOUT)
    return process, log


def stop_n8n(process: subprocess.Popen[str], log: Path) -> None:
    if process.poll() is None:
        process.terminate()
    try:
        process.wait(timeout=15)
    except subprocess.TimeoutExpired as error:
        process.kill()
        process.wait(timeout=15)
        raise AssertionError("disposable n8n server did not stop") from error
    if process.returncode not in {0, -15}:
        raise AssertionError(f"disposable n8n server exited {process.returncode}:\n{log.read_text(encoding='utf-8')[-8000:]}")


def wait_for_execution(database: Path, workflow_id: str, status: str, server: subprocess.Popen[str], log: Path, *, timeout: float = 35) -> tuple[int, dict]:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        if server.poll() is not None:
            raise AssertionError(f"n8n server stopped before scheduled execution:\n{log.read_text(encoding='utf-8')[-8000:]}")
        try:
            connection = sqlite3.connect(database, timeout=0.2)
            finished_clause = "AND finished = 1" if status == "success" else ""
            row = connection.execute(
                f"SELECT id FROM execution_entity WHERE workflowId = ? AND mode = 'trigger' AND status = ? {finished_clause} ORDER BY id DESC LIMIT 1",
                (workflow_id, status),
            ).fetchone()
            connection.close()
            if row is not None:
                execution_id = int(row[0])
                connection = sqlite3.connect(database, timeout=0.2)
                data = connection.execute("SELECT data FROM execution_data WHERE executionId = ?", (execution_id,)).fetchone()
                connection.close()
                if data is not None:
                    return execution_id, decode_flatted_execution(data[0])
        except (sqlite3.OperationalError, json.JSONDecodeError):
            pass
        time.sleep(0.2)
    raise AssertionError(f"timed out waiting for scheduled {status} execution:\n{log.read_text(encoding='utf-8')[-8000:]}")


def decode_flatted_execution(raw: str) -> dict:
    """Decode n8n 2.38's flatted SQLite execution payload without Node modules."""
    values = json.loads(raw)
    if not isinstance(values, list) or not values:
        raise AssertionError("n8n execution data is not a flatted array")
    resolving: set[int] = set()

    def resolve(value: object) -> object:
        if not isinstance(value, str):
            if isinstance(value, list):
                return [resolve(item) for item in value]
            if isinstance(value, dict):
                return {key: resolve(item) for key, item in value.items()}
            return value
        try:
            index = int(value)
        except ValueError:
            return value
        if str(index) != value or index < 0 or index >= len(values):
            return value
        if index in resolving:
            # M06 has no cycles, but preserve a harmless marker if n8n adds one.
            return {"$flatted_ref": index}
        resolving.add(index)
        try:
            return resolve(values[index])
        finally:
            resolving.remove(index)

    decoded = resolve(values[0])
    if not isinstance(decoded, dict):
        raise AssertionError("decoded n8n execution data is not an object")
    return decoded


def assert_success(execution_id: int, execution: dict, history: Path) -> None:
    result = execution.get("resultData", execution.get("data", {}).get("resultData", {}))
    wrapped = {"status": "success", "finished": True, "data": {"resultData": result}}
    record_id = validate_success(wrapped, history, expected_result="APPENDED")
    report = node_json(wrapped, "Report Canonical M06 Result")
    if report.get("execution_permitted") is not False or report.get("canonical_history_handoff") != "ACK":
        raise AssertionError("scheduled M06 report crossed its read-only boundary")
    records = [line for line in history.read_text(encoding="utf-8").splitlines() if line.strip()]
    if len(records) != 1:
        raise AssertionError("scheduled M06 fixture did not produce exactly one idempotent canonical record")
    print(f"M06 SCHEDULE SUCCESS: trigger execution={execution_id} canonical_record={record_id}")


def assert_adapter_failure(execution: dict, history: Path) -> None:
    result = execution.get("resultData", execution.get("data", {}).get("resultData", {}))
    if result.get("lastNodeExecuted") != "Build and Append Canonical M06 Adapter":
        raise AssertionError("unavailable adapter did not fail at the M06 canonical handoff")
    if "Report Canonical M06 Result" in result.get("runData", {}):
        raise AssertionError("unavailable adapter reached a persistence report")
    if history.exists() and history.read_text(encoding="utf-8").strip():
        raise AssertionError("unavailable adapter changed canonical history")


def run_case(prefix: list[str], args: argparse.Namespace, *, available_adapter: bool) -> None:
    runtime = Path(tempfile.mkdtemp(prefix="affiliate-n8n-m06-schedule-"))
    keep = args.keep_runtime
    adapter: Optional[subprocess.Popen[str]] = None
    server: Optional[subprocess.Popen[str]] = None
    try:
        n8n_home = runtime / "n8n"
        env = dict(os.environ)
        env.update({"N8N_USER_FOLDER": str(n8n_home), "N8N_ENCRYPTION_KEY": "n8n-ci-isolated-fixture-key-not-a-secret", "N8N_DIAGNOSTICS_ENABLED": "false", "N8N_PERSONALIZATION_ENABLED": "false", "N8N_ENFORCE_SETTINGS_FILE_PERMISSIONS": "false", "N8N_RUNNERS_BROKER_PORT": str(choose_port()), "GOWORK": "off", "GOCACHE": str(runtime / "go-cache")})
        if args.n8n_node:
            env["PATH"] = str(Path(args.n8n_node).resolve().parent) + os.pathsep + env.get("PATH", "")
        bot = runtime / "bot"
        run(["go", "build", "-o", str(bot), "./cmd/bot"], env=env, cwd=BOT_DIR)
        history = runtime / "history.jsonl"
        adapter_port = choose_port()
        if available_adapter:
            adapter = start_adapter(bot, history, adapter_port, env)
        workflow_id = "rp08-m06-schedule-success" if available_adapter else "rp08-m06-schedule-adapter-down"
        import_scheduled_workflow(prefix, env, runtime, workflow_id, adapter_port)
        database = database_path(n8n_home)
        activate_schedule_workflow(database, workflow_id)
        server_port = choose_port()
        server, log = start_n8n(prefix, env, server_port, runtime)
        expected_status = "success" if available_adapter else "error"
        execution_id, execution = wait_for_execution(database, workflow_id, expected_status, server, log)
        if available_adapter:
            assert_success(execution_id, execution, history)
            replay = run([str(bot), "history", "replay", str(history)], env=env)
            if "replay=MATCH" not in replay.stdout:
                raise AssertionError("scheduled canonical history did not replay after n8n start")
        else:
            assert_adapter_failure(execution, history)
            print(f"M06 SCHEDULE REJECTION: trigger execution={execution_id} adapter failure stopped before ACK/report")
    except BaseException:
        keep = True
        raise
    finally:
        if server is not None:
            stop_n8n(server, runtime / "n8n-server.log")
        if adapter is not None:
            stop_adapter(adapter)
        if keep:
            print(f"M06 schedule regression runtime retained at {runtime}", file=sys.stderr)
        else:
            shutil.rmtree(runtime)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--n8n-cli", help="n8n executable or CLI JavaScript entrypoint")
    parser.add_argument("--n8n-node", help="Node executable when --n8n-cli is a JavaScript entrypoint")
    parser.add_argument("--keep-runtime", action="store_true", help="preserve disposable runtime")
    args = parser.parse_args()
    prefix = command_prefix(args)
    if shutil.which(prefix[0]) is None and not Path(prefix[0]).is_file():
        raise SystemExit(f"n8n command is unavailable: {prefix[0]}")
    if len(prefix) == 2 and not Path(prefix[1]).is_file():
        raise SystemExit(f"n8n CLI entrypoint is unavailable: {prefix[1]}")
    run_case(prefix, args, available_adapter=True)
    run_case(prefix, args, available_adapter=False)
    print("N8N M06 SCHEDULE REGRESSION PASS: real Schedule Trigger admitted a synthetic read-only canonical handoff and failed closed when its loopback adapter was unavailable")


if __name__ == "__main__":
    main()
