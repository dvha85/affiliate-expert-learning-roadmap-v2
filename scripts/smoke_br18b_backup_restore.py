"""Backup/restore smoke using artifacts created and replayed by the learner Bot."""
import json
import hashlib
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BOT_DIR = ROOT / "lab/affiliate-bot"


def run(command, cwd=ROOT, expected=0, env=None):
    result = subprocess.run([str(x) for x in command], cwd=cwd, text=True, capture_output=True, env=env)
    if result.returncode != expected:
        raise AssertionError((command, result.stdout, result.stderr))
    return result


def invoke(bot, *args, expected=0, env=None):
    return json.loads(run([bot, *args], expected=expected, env=env).stdout)


def invoke_with_stderr(bot, *args, expected=0, env=None):
    result = run([bot, *args], expected=expected, env=env)
    return json.loads(result.stdout), result.stderr


def m11_entries(runtime):
    path = runtime / "m11-artifacts.jsonl"
    return [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines() if line.strip()]


def assert_closed_cycle_snapshot(runtime):
    entries = m11_entries(runtime)
    kinds = {entry["artifact_kind"] for entry in entries}
    required = {
        "PRODUCTION_LEASE",
        "PRODUCTION_LEASE_APPROVAL",
        "PRODUCTION_ACTIVATION",
        "PRODUCTION_HEALTH_SNAPSHOT",
        "TRUSTED_COST_BOUND",
        "PRODUCTION_GATE",
        "PRODUCTION_EXECUTION_AUTHORIZATION",
        "PRODUCTION_EXECUTION_RECORD",
        "PRODUCTION_OUTCOME_EVALUATION",
        "PRODUCTION_CYCLE",
    }
    missing = sorted(required - kinds)
    if missing:
        raise AssertionError(("restored M11 chain is incomplete", missing))
    cycles = [entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_CYCLE"]
    if len(cycles) != 1 or cycles[0]["status"] != "CLOSED":
        raise AssertionError(("restored M11 cycle is not exactly one CLOSED cycle", cycles))
    cycle = cycles[0]
    evaluations = [entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_OUTCOME_EVALUATION"]
    executions = [entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_EXECUTION_RECORD"]
    if len(evaluations) != 1 or len(executions) != 1:
        raise AssertionError(("restored M11 terminal cardinality changed", len(executions), len(evaluations)))
    evaluation, execution = evaluations[0], executions[0]
    if (
        cycle["execution_id"] != execution["execution_id"]
        or cycle["outcome_id"] != evaluation["outcome_id"]
        or cycle["evaluation_id"] != evaluation["evaluation_id"]
        or evaluation["execution_id"] != execution["execution_id"]
        or evaluation["source_profile"] != "OFFLINE_FIXTURE"
    ):
        raise AssertionError(("restored M11 terminal lineage changed", cycle, execution, evaluation))
    post_ledgers = [
        entry["artifact"]
        for entry in entries
        if entry["artifact_kind"] == "PRODUCTION_LEDGER"
        and any(link.get("outcome_id") == cycle["outcome_id"] for link in entry["artifact"].get("outcome_links", []))
    ]
    if len(post_ledgers) != 1:
        raise AssertionError(("restored M11 post-ledger outcome link cardinality changed", post_ledgers))
    post_ledger = post_ledgers[0]
    if post_ledger["control_mode"] != "NORMAL" or post_ledger["pending_outcomes"] != 0:
        raise AssertionError(("restored M11 post-ledger is not a completed NORMAL ledger", post_ledger))


def assert_resolved_stop_snapshot(runtime, resolution_id, execution_id):
    entries = m11_entries(runtime)
    resolutions = [
        entry["artifact"]
        for entry in entries
        if entry["artifact_kind"] == "PRODUCTION_RECONCILIATION"
        and entry["artifact"].get("resolution_id") == resolution_id
    ]
    if len(resolutions) != 1 or resolutions[0]["execution_id"] != execution_id:
        raise AssertionError(("resolved STOP resolution cardinality or execution link changed", resolutions))
    resolved_ledgers = [
        entry["artifact"]
        for entry in entries
        if entry["artifact_kind"] == "PRODUCTION_LEDGER"
        and entry["artifact"].get("reconciliation_resolution_ids") == [resolution_id]
    ]
    if len(resolved_ledgers) != 1:
        raise AssertionError(("resolved STOP ledger cardinality changed", resolved_ledgers))
    ledger = resolved_ledgers[0]
    if ledger["control_mode"] != "STOPPED" or ledger["stop_reason"] != "RECOVERY_REVIEW_REQUIRED" or ledger["reconciliation_required"]:
        raise AssertionError(("resolved STOP ledger is not review-gated", ledger))


def replace_backup_file(backup, name, content):
    path = backup / name
    path.write_bytes(content)
    manifest_path = backup / "manifest.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    manifest["files"][name]["sha256"] = hashlib.sha256(content).hexdigest()
    manifest["files"][name]["size_bytes"] = len(content)
    manifest_path.write_text(json.dumps(manifest), encoding="utf-8")


def rewrite_first_jsonl_record(backup, name, change):
    lines = (backup / name).read_text(encoding="utf-8").splitlines()
    if not lines:
        raise AssertionError(("empty JSONL store", name))
    record = json.loads(lines[0])
    change(record)
    lines[0] = json.dumps(record, separators=(",", ":"), ensure_ascii=False)
    replace_backup_file(backup, name, ("\n".join(lines) + "\n").encode())


def rewrite_m11_registry(backup, change):
    lines = []
    for line in (backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines():
        entry = json.loads(line)
        if change(entry):
            canonical = json.dumps(entry["artifact"], separators=(",", ":"), ensure_ascii=False).encode()
            entry["content_hash"] = "sha256:" + hashlib.sha256(canonical).hexdigest()
        lines.append(json.dumps(entry, separators=(",", ":"), ensure_ascii=False))
    replace_backup_file(backup, "m11-artifacts.jsonl", ("\n".join(lines) + "\n").encode())


def append_m11_registry_entry(backup, kind, artifact):
    canonical = json.dumps(artifact, separators=(",", ":"), ensure_ascii=False).encode()
    entry = {
        "artifact_kind": kind,
        "artifact_id": artifact["authorization_id"] if kind == "PRODUCTION_EXECUTION_AUTHORIZATION" else artifact["execution_id"] if kind == "PRODUCTION_EXECUTION_RECORD" else artifact["lease_id"] + "/" + artifact["updated_at"],
        "content_hash": "sha256:" + hashlib.sha256(canonical).hexdigest(),
        "artifact": artifact,
    }
    path = backup / "m11-artifacts.jsonl"
    lines = path.read_text(encoding="utf-8").splitlines()
    lines.append(json.dumps(entry, separators=(",", ":"), ensure_ascii=False))
    replace_backup_file(backup, "m11-artifacts.jsonl", ("\n".join(lines) + "\n").encode())


def add_m11_failed_execution_without_outcome(backup):
    entries = m11_entries(backup)
    authorization = next(entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_EXECUTION_AUTHORIZATION")
    execution = next(entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_EXECUTION_RECORD")
    ledger = max(
        (entry["artifact"] for entry in entries if entry["artifact_kind"] == "PRODUCTION_LEDGER"),
        key=lambda item: item["updated_at"],
    )
    authorized_at = "2026-09-08T00:00:04Z"
    attempted_at = "2026-09-08T00:00:07Z"
    authorization_id = "prod-auth-" + hashlib.sha256(
        (authorization["production_gate_id"] + "\x00" + authorization["executor_id"] + "\x00" + authorized_at).encode()
    ).digest()[:16].hex()
    new_authorization = json.loads(json.dumps(authorization))
    new_authorization.update({
        "authorization_id": authorization_id,
        "authorized_at": authorized_at,
        "idempotency_key": "br18-mutation-failed-outcome",
    })
    execution_id = "prod-exec-" + authorization_id
    new_execution = json.loads(json.dumps(execution))
    new_execution.update({
        "execution_id": execution_id,
        "authorization_id": authorization_id,
        "idempotency_key": new_authorization["idempotency_key"],
        "attempted_at": attempted_at,
        "status": "FAILED",
        "side_effect_state": "NOT_PERFORMED",
        "error": "fixture-secondary-dispatch-failed-without-outcome",
    })
    new_ledger = json.loads(json.dumps(ledger))
    new_ledger.update({
        "updated_at": attempted_at,
        "last_execution_at": attempted_at,
        "executions_total": ledger["executions_total"] + 1,
        "executions_in_window": ledger["executions_in_window"] + 1,
        "cost_minor_total": ledger["cost_minor_total"] + new_authorization["production_cost_bound_minor"],
        "pending_outcomes": 1,
        "pending_execution_ids": [execution_id],
    })
    append_m11_registry_entry(backup, "PRODUCTION_LEDGER", new_ledger)
    append_m11_registry_entry(backup, "PRODUCTION_EXECUTION_AUTHORIZATION", new_authorization)
    append_m11_registry_entry(backup, "PRODUCTION_EXECUTION_RECORD", new_execution)


def remove_m11_entry(backup, kind, artifact_id):
    lines, removed = [], False
    for line in (backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines():
        entry = json.loads(line)
        if entry["artifact_kind"] == kind and entry["artifact_id"] == artifact_id:
            removed = True
            continue
        lines.append(json.dumps(entry, separators=(",", ":"), ensure_ascii=False))
    if not removed:
        raise AssertionError(("missing M11 artifact to remove", kind, artifact_id))
    replace_backup_file(backup, "m11-artifacts.jsonl", ("\n".join(lines) + "\n").encode())


def insert_duplicate_m11_resolution_before_original(backup):
    lines, inserted = [], False
    for line in (backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines():
        entry = json.loads(line)
        if entry["artifact_kind"] == "PRODUCTION_RECONCILIATION" and not inserted:
            duplicate = json.loads(json.dumps(entry))
            duplicate["artifact"]["resolution_id"] += "-duplicate"
            canonical = json.dumps(duplicate["artifact"], separators=(",", ":"), ensure_ascii=False).encode()
            duplicate["artifact_id"] = duplicate["artifact"]["resolution_id"]
            duplicate["content_hash"] = "sha256:" + hashlib.sha256(canonical).hexdigest()
            lines.append(json.dumps(duplicate, separators=(",", ":"), ensure_ascii=False))
            inserted = True
        lines.append(line)
    if not inserted:
        raise AssertionError("missing M11 reconciliation artifact")
    replace_backup_file(backup, "m11-artifacts.jsonl", ("\n".join(lines) + "\n").encode())


def replace_m11_field(kind, field, value):
    def change(entry):
        if entry["artifact_kind"] != kind:
            return False
        entry["artifact"][field] = value
        return True
    return change


def replace_m11_health_observed_at(value):
    def change(entry):
        if entry["artifact_kind"] != "PRODUCTION_HEALTH_SNAPSHOT":
            return False
        artifact = entry["artifact"]
        artifact["observed_at"] = value
        # Mirror core/m11's hash payload exactly, including deterministic
        # source-ref ordering, so this is a checksum-valid semantic mutation.
        payload = {
            "snapshot_id": artifact["snapshot_id"], "lease_id": artifact["lease_id"],
            "lease_version": artifact["lease_version"], "lease_hash": artifact["lease_hash"],
            "observed_at": artifact["observed_at"], "source_refs": sorted(artifact["source_refs"]),
            "dependency_state": artifact["dependency_state"], "telemetry_complete": artifact["telemetry_complete"],
            "consecutive_failures": artifact["consecutive_failures"], "reconciliation_required": artifact["reconciliation_required"],
            "compliance_alert_count": artifact["compliance_alert_count"],
            "oldest_pending_outcome_age_seconds": artifact["oldest_pending_outcome_age_seconds"],
            "hash_version": artifact["hash_version"],
        }
        artifact["snapshot_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        return True
    return change


def rewrite_m11_source_grant_mismatch(backup):
    entries = [json.loads(line) for line in (backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines()]
    lease = next((entry for entry in entries if entry["artifact_kind"] == "PRODUCTION_LEASE"), None)
    if lease is None:
        raise AssertionError("missing M11 production lease")
    bad_hash = "sha256:" + "0" * 64
    lease["artifact"]["source_canary_grant_hash"] = bad_hash
    lease_payload = dict(lease["artifact"])
    lease_payload.pop("lease_hash", None)
    lease["artifact"]["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(lease_payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    for entry in entries:
        if entry["artifact_kind"] == "PRODUCTION_LEASE_APPROVAL":
            entry["artifact"]["lease_hash"] = lease["artifact"]["lease_hash"]
            entry["artifact"]["source_canary_grant_hash"] = bad_hash
        canonical = json.dumps(entry["artifact"], separators=(",", ":"), ensure_ascii=False).encode()
        entry["content_hash"] = "sha256:" + hashlib.sha256(canonical).hexdigest()
    replace_backup_file(backup, "m11-artifacts.jsonl", ("\n".join(json.dumps(entry, separators=(",", ":"), ensure_ascii=False) for entry in entries) + "\n").encode())


def write_canary_grant(path, intent, policy, approval, max_executions, max_cost):
    payload = {
        "grant_id": "br18-g", "grant_version": "v1", "policy_version": policy["policy_version"],
        "approval_ref": approval["approval_id"], "approved_by": "human", "approver_id": approval["approver_id"],
        "approved_at": approval["approved_at"], "valid_from": approval["approved_at"], "expires_at": approval["expires_at"],
        "allowed_risk_classes": [policy["risk_class"]], "allowed_action_types": [intent["action_type"]], "allowed_hosts": ["example.com"],
        "executor_ids": ["local_sandbox"], "max_executions_total": max_executions, "max_executions_per_window": max_executions,
        "window_seconds": 3600, "max_cost_minor_total": max_cost, "currency": "USD", "max_pending_outcomes": 1,
        "kill_switch_required": True, "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    payload["grant_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_cost_bound(path, intent, amount, expires_at, bound_id):
    payload = {
        "cost_bound_id": bound_id, "intent_id": intent["intent_id"], "intent_hash": intent["intent_hash"],
        "max_cost_minor": amount, "currency": "USD", "source_ref": "fixture:br18-cost-registry",
        "observed_at": "2026-09-07T01:06:00Z", "expires_at": expires_at,
        "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    digest = hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    payload["cost_bound_hash"] = "sha256:" + digest
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_lease(path, source_canary_grant_hash, max_executions=1, max_cost=4, max_pending_outcomes=1):
    payload = {
        "lease_id": "br18-production-lease", "lease_version": "v1", "policy_version": "br18-v1",
        "approval_ref": "br18-production-approval", "reviewed_by": "human", "reviewer_id": "pilot-human",
        "reviewed_at": "2026-09-07T01:05:00Z", "promotion_review_ref": "fixture:br18-promotion-review",
        "source_canary_grant_id": "br18-g", "source_canary_grant_version": "v1",
        "source_canary_grant_hash": source_canary_grant_hash, "valid_from": "2026-09-07T01:05:00Z",
        "expires_at": "2099-09-03T02:50:00Z", "allowed_risk_classes": ["RISK0"],
        "allowed_action_types": ["DRAFT"], "allowed_hosts": ["example.com"], "executor_ids": ["fixture_stub"],
        "max_executions_total": max_executions, "max_executions_per_window": max_executions, "window_seconds": 60,
        "max_cost_minor_total": max_cost, "currency": "USD", "max_pending_outcomes": max_pending_outcomes,
        "max_consecutive_failures": 1, "max_outcome_age_seconds": 60, "max_health_snapshot_age_seconds": 60,
        "kill_switch_required": True, "correlation_id": "br18-c", "hash_version": "go-json-v1",
    }
    payload["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_lease_approval(path, lease_path):
    lease = json.loads(lease_path.read_text(encoding="utf-8"))
    payload = {
        "approval_id": lease["approval_ref"], "lease_id": lease["lease_id"], "lease_version": lease["lease_version"],
        "lease_hash": lease["lease_hash"], "promotion_review_ref": lease["promotion_review_ref"],
        "source_canary_grant_id": lease["source_canary_grant_id"], "source_canary_grant_version": lease["source_canary_grant_version"],
        "source_canary_grant_hash": lease["source_canary_grant_hash"], "source_e5_refs": ["fixture:br18-e5"],
        "validated_risk_classes": ["RISK0"], "reviewed_by": "human", "reviewer_id": lease["reviewer_id"],
        "reviewed_at": lease["reviewed_at"], "decision": "APPROVE_PRODUCTION_LEASE",
    }
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_recovery_admission_lease(path, prior_lease_path):
    payload = json.loads(prior_lease_path.read_text(encoding="utf-8"))
    payload.update({
        "lease_id": "br18-recovery-admission-lease",
        "lease_version": "v1",
        "approval_ref": "br18-recovery-admission-approval",
        "reviewed_at": "2026-09-08T00:00:06Z",
        "valid_from": "2026-09-08T00:00:06Z",
        "promotion_review_ref": "fixture:br18-recovery-admission-review",
    })
    del payload["lease_hash"]
    payload["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_health(path, lease_path, observed_at="2026-09-08T00:00:00Z"):
    lease = json.loads(lease_path.read_text(encoding="utf-8"))
    payload = {
        "snapshot_id": "br18-production-health", "lease_id": lease["lease_id"], "lease_version": lease["lease_version"],
        "lease_hash": lease["lease_hash"], "observed_at": observed_at, "source_refs": ["fixture:br18-health"],
        "dependency_state": "HEALTHY", "telemetry_complete": True, "consecutive_failures": 0,
        "reconciliation_required": False, "compliance_alert_count": 0, "oldest_pending_outcome_age_seconds": 0,
        "hash_version": "go-json-v1",
    }
    payload["snapshot_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_resolution(path, lease_path, execution_id):
    lease = json.loads(lease_path.read_text(encoding="utf-8"))
    payload = {
        "resolution_id": "br18-production-resolution", "lease_id": lease["lease_id"],
        "lease_version": lease["lease_version"], "lease_hash": lease["lease_hash"],
        "execution_id": execution_id, "resolved_by": "human", "resolver_id": "pilot-human",
        "resolved_at": "2026-09-08T00:00:05Z", "effect_state": "NOT_PERFORMED",
        "reason": "fixture provider audit confirmed no side effect",
    }
    path.write_text(json.dumps(payload), encoding="utf-8")


def main():
    go = shutil.which(os.environ.get("GO_BIN", "go")) or os.environ.get("GO_BIN", "go")
    with tempfile.TemporaryDirectory(prefix="br18b-runtime-") as directory:
        root = Path(directory); bot = root / "bot"; runtime = root / "runtime"; backup = root / "backup"; restored = root / "restored"; observations = root / "observations.json"; history = runtime / "history.jsonl"; env = dict(os.environ, GOWORK="off", GOCACHE=str(root / "go-cache"))
        run([go, "build", "-o", bot, "./cmd/bot"], BOT_DIR, env=env)
        runtime.mkdir(); observations.write_text(json.dumps(invoke(bot, "evidence", "import", ROOT / "examples/m00-import/packet-t2.json", env=env)["artifact"]), encoding="utf-8")
        assert "APPENDED" in run([bot, "history", "capture", history, observations, "br18-d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"], env=env).stdout
        action = root / "action.json"; action.write_text(json.dumps({"action_id":"br18-a","decision_id":"br18-d","action_type":"synthetic_manual","target":"fixture:br18","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True}), encoding="utf-8")
        assert invoke(bot, "action", "record", history, runtime / "actions.jsonl", action, env=env)["status"] == "APPENDED"
        invoke(bot, "mission", "init", runtime, env=env)
        request = root / "intent-request.json"; intent = root / "intent.json"; policy_config = root / "policy.json"; policy = root / "policy-output.json"
        request.write_text(json.dumps({"intent_id":"br18-i","decision_id":"br18-d","action_type":"DRAFT","target":"https://example.com/draft","parameters":{},"proposed_by":"human","created_at":"2026-09-07T00:00:00Z","expires_at":"2099-09-03T03:00:00Z","correlation_id":"br18-c","idempotency_key":"br18-k"}), encoding="utf-8")
        assert invoke(bot, "mission", "m08-intent", history, request, intent, env=env)["status"] == "APPENDED"
        policy_config.write_text(json.dumps({"policy_version":"br18-v1","now":"2026-09-07T01:00:00Z","allowed_hosts":["example.com"],"action_risk":{"DRAFT":"RISK0"},"seen_idempotency":{}}), encoding="utf-8")
        assert invoke(bot, "mission", "m08-policy", intent, policy_config, policy, env=env)["status"] == "ALLOW"
        invoke(bot, "mission", "bind", runtime, intent, policy, env=env)
        i = json.loads(intent.read_text()); p = json.loads(policy.read_text()); approval = root / "approval.json"
        approval.write_text(json.dumps({"approval_id":"br18-ap","intent_id":i["intent_id"],"intent_hash":i["intent_hash"],"policy_version":p["policy_version"],"decision":"APPROVE","approved_by":"human","approver_id":"pilot-human","approved_at":"2026-09-07T01:05:00Z","expires_at":"2099-09-03T02:50:00Z","correlation_id":i["correlation_id"],"one_time":True}), encoding="utf-8")
        assert invoke(bot, "mission", "m09-approval", runtime, approval, env=env)["status"] == "ACK"
        grant = root / "grant.json"; write_canary_grant(grant, i, p, json.loads(approval.read_text()), 1, 4)
        assert invoke(bot, "mission", "m10-canary", runtime, grant, env=env)["status"] == "ACK"
        cost = root / "cost-bound.json"; write_cost_bound(cost, i, 4, "2099-09-03T02:45:00Z", "br18-cost")
        assert invoke(bot, "mission", "m10-cost-register", runtime, cost, env=env)["status"] == "APPENDED"
        gate = root / "canary-gate.json"; gate_time = "2026-09-08T00:00:00Z"
        assert invoke(bot, "mission", "m10-gate", runtime, cost, gate, gate_time, env=env)["status"] == "ALLOW_CANARY"
        authorization = root / "canary-authorization.json"
        authorization_result = invoke(bot, "mission", "m10-authorize", runtime, cost, gate, authorization, gate_time, "local_sandbox", env=env)
        assert authorization_result["status"] == "AUTHORIZED"
        reservation = invoke(bot, "mission", "m10-reserve-authorization", runtime, authorization, "br18-governed-r1", env=env)
        assert reservation["status"] == "RESERVED"
        failed_attempted_at = reservation["artifact"]["reserved_at"]
        failed_execution = root / "fixture-failed-execution.json"
        failed = invoke(bot, "mission", "m10-record-failed", runtime, authorization, failed_execution, failed_attempted_at, "fixture-dispatch-failed-before-executor", env=env)
        assert failed["status"] == "APPENDED" and failed["artifact"]["status"] == "FAILED"
        machine_outcome = root / "machine-outcome.json"
        machine_outcome.write_text(json.dumps({"outcome_id":"br18-machine-o","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":failed["artifact"]["execution_id"]},"observed_at":failed_attempted_at,"status":"CANCELLED","metrics":{},"source_ref":"fixture:m10-outcome/br18-failed"}), encoding="utf-8")
        assert invoke(bot, "mission", "m10-outcome", runtime, machine_outcome, env=env)["status"] == "APPENDED"
        grant_hash = json.loads(grant.read_text(encoding="utf-8"))["grant_hash"]
        production_lease = root / "production-lease.json"; write_production_lease(production_lease, grant_hash)
        production_approval = root / "production-approval.json"; write_production_lease_approval(production_approval, production_lease)
        recovery_runtime = root / "reconciliation-runtime"; recovery_backup = root / "reconciliation-backup"; recovery_restored = root / "reconciliation-restored"
        early_gate_runtime = root / "early-gate-runtime"
        shutil.copytree(runtime, recovery_runtime)
        shutil.copytree(runtime, early_gate_runtime)
        assert invoke(bot, "mission", "m11-register", early_gate_runtime, "PRODUCTION_LEASE", production_lease, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", early_gate_runtime, "PRODUCTION_LEASE_APPROVAL", production_approval, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", early_gate_runtime, "br18-production-lease", "2026-09-08T00:00:10Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", early_gate_runtime, "br18-production-lease", "2026-09-08T00:00:10Z", env=env)["status"] == "APPENDED"
        production_health = root / "production-health.json"; write_production_health(production_health, production_lease)
        early_health = root / "early-production-health.json"; write_production_health(early_health, production_lease, "2026-09-08T00:00:10Z")
        assert invoke(bot, "mission", "m11-register", early_gate_runtime, "PRODUCTION_HEALTH_SNAPSHOT", production_health, expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-register", early_gate_runtime, "PRODUCTION_HEALTH_SNAPSHOT", early_health, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", early_gate_runtime, "TRUSTED_COST_BOUND", cost, env=env)["status"] == "APPENDED"
        # The command computes a DENY candidate for an inactive lease, but the
        # canonical registry now independently rejects a gate that predates its
        # registered activation. The artifact must not be persisted through
        # either boundary.
        early_gate = invoke(bot, "mission", "m11-gate", early_gate_runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:10Z", "2026-09-08T00:00:05Z", expected=1, env=env)
        assert early_gate["status"] == "REJECTED"
        # Two valid authorizations can coexist before the first reservation.
        # The second must not consume budget after the first advances the
        # ledger that its gate snapshot authorized.
        stale_gate_runtime = root / "stale-gate-runtime"; stale_lease = root / "stale-gate-lease.json"; stale_approval = root / "stale-gate-approval.json"; stale_health = root / "stale-gate-health.json"
        shutil.copytree(runtime, stale_gate_runtime)
        write_production_lease(stale_lease, grant_hash, max_executions=2, max_cost=8, max_pending_outcomes=2)
        write_production_lease_approval(stale_approval, stale_lease); write_production_health(stale_health, stale_lease)
        for kind, artifact in (("PRODUCTION_LEASE", stale_lease), ("PRODUCTION_LEASE_APPROVAL", stale_approval), ("PRODUCTION_HEALTH_SNAPSHOT", stale_health), ("TRUSTED_COST_BOUND", cost)):
            assert invoke(bot, "mission", "m11-register", stale_gate_runtime, kind, artifact, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", stale_gate_runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", stale_gate_runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        stale_gate_one = invoke(bot, "mission", "m11-gate", stale_gate_runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", env=env)
        stale_gate_two = invoke(bot, "mission", "m11-gate", stale_gate_runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)
        stale_auth_one = invoke(bot, "mission", "m11-authorize", stale_gate_runtime, "br18-production-lease", stale_gate_one["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:00Z", env=env)
        stale_auth_two = invoke(bot, "mission", "m11-authorize", stale_gate_runtime, "br18-production-lease", stale_gate_two["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:01Z", env=env)
        # The same immutable gate/executor may be reviewed at a later time.
        # That second capability must have its own identity; it still cannot
        # reserve after the first gate's ledger has advanced.
        same_gate_later_auth = invoke(bot, "mission", "m11-authorize", stale_gate_runtime, "br18-production-lease", stale_gate_one["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:01Z", env=env)
        assert same_gate_later_auth["status"] == "APPENDED" and same_gate_later_auth["artifact"]["authorization_id"] != stale_auth_one["artifact"]["authorization_id"]
        stale_reservation = invoke(bot, "mission", "m11-reserve-authorization", stale_gate_runtime, stale_auth_one["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:02Z", env=env)
        stale_ledger_id = stale_reservation["artifact"]["lease_id"] + "/" + stale_reservation["artifact"]["updated_at"]
        assert invoke(bot, "mission", "m11-reserve-authorization", stale_gate_runtime, stale_auth_two["artifact"]["authorization_id"], stale_ledger_id, "2026-09-08T00:00:03Z", expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-reserve-authorization", stale_gate_runtime, same_gate_later_auth["artifact"]["authorization_id"], stale_ledger_id, "2026-09-08T00:00:03Z", expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_LEASE", production_lease, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_LEASE_APPROVAL", production_approval, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_HEALTH_SNAPSHOT", production_health, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", runtime, "TRUSTED_COST_BOUND", cost, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-gate", runtime, "br18-production-lease", "missing-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", expected=1, env=env)["status"] == "REJECTED"
        gate = invoke(bot, "mission", "m11-gate", runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", env=env)
        assert gate["status"] == "ALLOW_PRODUCTION" and gate["artifact"]["execution_authorized"] is False
        refreshed_gate = invoke(bot, "mission", "m11-gate", runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)
        assert refreshed_gate["status"] == "ALLOW_PRODUCTION" and refreshed_gate["artifact"]["gate_id"] != gate["artifact"]["gate_id"]
        assert invoke(bot, "mission", "m11-authorize", runtime, "br18-production-lease", gate["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:02:00Z", expected=1, env=env)["status"] == "REJECTED"
        authorization = invoke(bot, "mission", "m11-authorize", runtime, "br18-production-lease", gate["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:00Z", env=env)
        assert authorization["status"] == "APPENDED" and authorization["artifact"]["execution_authorized"] is True
        reservation_ledger_id = "br18-production-lease/2026-09-08T00:00:01Z"
        reservation = invoke(bot, "mission", "m11-reserve-authorization", runtime, authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)
        assert reservation["status"] == "APPENDED" and reservation["artifact"]["pending_outcomes"] == 1
        assert invoke(bot, "mission", "m11-ledger-init", runtime, "br18-production-lease", "2026-09-08T00:00:04Z", expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-reserve-authorization", runtime, authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m11-reserve-authorization", runtime, authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:09Z", expected=1, env=env)["status"] == "REJECTED"
        production_failed = invoke(bot, "mission", "m11-record-failed", runtime, authorization["artifact"]["authorization_id"], reservation_ledger_id, "2026-09-08T00:00:02Z", "fixture-production-dispatch-failed", env=env)
        assert production_failed["status"] == "APPENDED" and production_failed["artifact"]["execution"]["status"] == "FAILED" and production_failed["artifact"]["execution"]["side_effect_state"] == "NOT_PERFORMED"
        # A FAILED transition changes the ledger safety state without changing
        # the four budget counters. The older gate must still be stale.
        assert invoke(bot, "mission", "m11-authorize", runtime, "br18-production-lease", gate["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:02Z", expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-record-failed", runtime, authorization["artifact"]["authorization_id"], reservation_ledger_id, "2026-09-08T00:00:02Z", "fixture-production-dispatch-failed", env=env)["status"] == "EXACT_DUPLICATE"
        production_outcome = root / "production-outcome.json"
        production_outcome.write_text(json.dumps({"outcome_id":"br18-production-o","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":production_failed["artifact"]["execution"]["execution_id"]},"observed_at":"2026-09-08T00:00:03Z","status":"CANCELLED","metrics":{},"source_ref":"fixture:m11-outcome/br18-failed"}), encoding="utf-8")
        execution_ledger_id = production_failed["artifact"]["execution_ledger"]["lease_id"] + "/" + production_failed["artifact"]["execution_ledger"]["updated_at"]
        orphaned_production_outcome = root / "orphaned-production-outcome.json"
        orphaned_production_outcome.write_text(json.dumps({"outcome_id":"br18-production-orphan","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":"missing-m11-execution"},"observed_at":"2026-09-08T00:00:03Z","status":"CANCELLED","metrics":{},"source_ref":"fixture:m11-outcome/orphan"}), encoding="utf-8")
        assert invoke(bot, "mission", "m11-outcome", runtime, orphaned_production_outcome, execution_ledger_id, expected=1, env=env)["status"] == "REJECTED"
        production_outcome_result = invoke(bot, "mission", "m11-outcome", runtime, production_outcome, execution_ledger_id, env=env)
        assert production_outcome_result["status"] == "APPENDED" and production_outcome_result["artifact"]["post_ledger"]["pending_outcomes"] == 0
        assert invoke(bot, "mission", "m11-outcome", runtime, production_outcome, execution_ledger_id, env=env)["status"] == "EXACT_DUPLICATE"
        duplicate_execution_outcome = root / "duplicate-execution-outcome.json"
        duplicate_execution_outcome.write_text(json.dumps({"outcome_id":"br18-production-o-duplicate","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":production_failed["artifact"]["execution"]["execution_id"]},"observed_at":"2026-09-08T00:00:04Z","status":"CANCELLED","metrics":{},"source_ref":"fixture:m11-outcome/duplicate"}), encoding="utf-8")
        assert invoke(bot, "mission", "m11-outcome", runtime, duplicate_execution_outcome, execution_ledger_id, expected=1, env=env)["status"] == "REJECTED"
        # A second approved intent cannot rebind the first runtime after its
        # grant has been consumed. Build a fresh mission envelope, then copy
        # only the immutable M11 registry so this drill can still exercise the
        # shared production-ledger budget boundary without discarding the
        # original intent's authority or reservation history.
        budget_runtime = root / "budget-runtime"; budget_runtime.mkdir()
        shutil.copy2(runtime / "history.jsonl", budget_runtime / "history.jsonl")
        shutil.copy2(runtime / "m11-artifacts.jsonl", budget_runtime / "m11-artifacts.jsonl")
        invoke(bot, "mission", "init", budget_runtime, env=env)
        budget_request = root / "budget-intent-request.json"; budget_intent = root / "budget-intent.json"; budget_policy = root / "budget-policy.json"
        budget_request.write_text(json.dumps({"intent_id":"br18-budget-i","decision_id":"br18-d","action_type":"DRAFT","target":"https://example.com/draft","parameters":{},"proposed_by":"human","created_at":"2026-09-07T00:00:00Z","expires_at":"2099-09-03T03:00:00Z","correlation_id":"br18-c","idempotency_key":"br18-budget-k"}), encoding="utf-8")
        assert invoke(bot, "mission", "m08-intent", budget_runtime / "history.jsonl", budget_request, budget_intent, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m08-policy", budget_intent, policy_config, budget_policy, env=env)["status"] == "ALLOW"
        assert invoke(bot, "mission", "bind", budget_runtime, budget_intent, budget_policy, env=env)["status"] == "BOUND"
        budget_intent_value = json.loads(budget_intent.read_text(encoding="utf-8")); budget_approval = json.loads(approval.read_text(encoding="utf-8"))
        budget_approval.update({"approval_id":"br18-budget-ap","intent_id":budget_intent_value["intent_id"],"intent_hash":budget_intent_value["intent_hash"]})
        budget_approval_path = root / "budget-approval.json"; budget_approval_path.write_text(json.dumps(budget_approval), encoding="utf-8")
        assert invoke(bot, "mission", "m09-approval", budget_runtime, budget_approval_path, env=env)["status"] == "ACK"
        budget_cost = root / "budget-cost.json"; write_cost_bound(budget_cost, budget_intent_value, 4, "2099-09-03T02:45:00Z", "br18-budget-cost")
        assert invoke(bot, "mission", "m11-register", budget_runtime, "TRUSTED_COST_BOUND", budget_cost, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-gate", budget_runtime, "br18-production-lease", "br18-production-health", "br18-budget-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:04Z", expected=1, env=env)["status"] == "REJECTED"
        budget_gate = invoke(bot, "mission", "m11-gate", budget_runtime, "br18-production-lease", "br18-production-health", "br18-budget-cost", production_outcome_result["artifact"]["post_ledger"]["lease_id"] + "/" + production_outcome_result["artifact"]["post_ledger"]["updated_at"], "2026-09-08T00:00:04Z", env=env)
        assert budget_gate["status"] == "DENY" and budget_gate["artifact"]["reason"] == "BUDGET_EXCEEDED"
        evaluation = invoke(bot, "mission", "m11-evaluate", runtime, "br18-production-o", "br18-production-e", "2026-09-08T00:00:04Z", env=env)
        assert evaluation["status"] == "APPENDED" and evaluation["artifact"]["source_profile"] == "OFFLINE_FIXTURE"
        assert invoke(bot, "mission", "m11-evaluate", runtime, "br18-production-o", "br18-production-e", "2026-09-08T00:00:04Z", env=env)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m11-close-cycle", runtime, "br18-production-cycle-too-early", "br18-production-e", "2026-09-08T00:00:03Z", expected=1, env=env)["status"] == "REJECTED"
        cycle = invoke(bot, "mission", "m11-close-cycle", runtime, "br18-production-cycle", "br18-production-e", "2026-09-08T00:00:05Z", env=env)
        assert cycle["status"] == "APPENDED" and cycle["artifact"]["status"] == "CLOSED"
        assert invoke(bot, "mission", "m11-close-cycle", runtime, "br18-production-cycle", "br18-production-e", "2026-09-08T00:00:05Z", env=env)["status"] == "EXACT_DUPLICATE"
        # Complete the real M00-M05 store lineage before taking the snapshot.
        # These are synthetic review artifacts, but they exercise the same
        # learner loaders that backup verification uses for every optional
        # upstream store.
        human_outcome = root / "human-outcome.json"
        human_outcome.write_text(json.dumps({"outcome_id":"br18-human-o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"br18-a"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:br18-human"}), encoding="utf-8")
        assert invoke(bot, "outcome", "import", history, runtime / "actions.jsonl", runtime / "outcomes.jsonl", human_outcome, env=env)["status"] == "APPENDED"
        m05_evaluation = root / "m05-evaluation.json"
        m05_evaluation.write_text(json.dumps({"evaluation_id":"br18-evaluation","decision_id":"br18-d","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"br18-a"},"outcome_ids":["br18-human-o"],"evaluated_at":"2026-09-06T00:00:00Z"}), encoding="utf-8")
        assert invoke(bot, "evaluation", "create", history, runtime / "actions.jsonl", runtime / "outcomes.jsonl", runtime / "evaluations.jsonl", m05_evaluation, env=env)["status"] == "APPENDED"
        m05_proposal = root / "m05-proposal.json"
        m05_proposal.write_text(json.dumps({"proposal_id":"br18-proposal","evaluation_ids":["br18-evaluation"],"current_version":"fixture/v1","proposed_version":"fixture/v2","change_summary":"Synthetic backup graph proposal","expected_benefit":"Exercise restore lineage","risks":["Fixture only"],"rollback":"Discard synthetic proposal","auto_apply":False}), encoding="utf-8")
        assert invoke(bot, "proposal", "import", history, runtime / "actions.jsonl", runtime / "outcomes.jsonl", runtime / "evaluations.jsonl", runtime / "proposals.jsonl", m05_proposal, env=env)["status"] == "APPENDED"
        m05_review = root / "m05-review.json"
        m05_review.write_text(json.dumps({"review_id":"br18-review","proposal_id":"br18-proposal","reviewed_by":"human","reviewed_at":"2026-09-07T00:00:00Z","decision":"REQUEST_CHANGES","reason":"Synthetic backup graph review"}), encoding="utf-8")
        assert invoke(bot, "review", "import", history, runtime / "actions.jsonl", runtime / "outcomes.jsonl", runtime / "evaluations.jsonl", runtime / "proposals.jsonl", runtime / "reviews.jsonl", m05_review, env=env)["status"] == "APPENDED"
        invoke(bot, "mission", "m11-stop", runtime, "backup-drill", env=env)
        backup_result = invoke(bot, "backup", "create", runtime, backup, env=env)
        assert backup_result["status"] == "BACKED_UP"
        assert backup_result["artifact"]["version"] == "affiliate-bot-backup/v3"
        assert {"actions.jsonl", "outcomes.jsonl", "evaluations.jsonl", "proposals.jsonl", "reviews.jsonl", "m10-artifacts.jsonl", "m10-outcomes.jsonl", "m11-artifacts.jsonl", "m11-outcomes.jsonl"}.issubset(backup_result["artifact"]["required"])
        source_grant_mismatch_backup = root / "source-grant-mismatch-backup"; shutil.copytree(backup, source_grant_mismatch_backup)
        rewrite_m11_source_grant_mismatch(source_grant_mismatch_backup)
        source_grant_mismatch_restored = root / "source-grant-mismatch-restored"
        source_grant_mismatch_result = invoke(bot, "backup", "restore", source_grant_mismatch_backup, source_grant_mismatch_restored, expected=1, env=env)
        assert source_grant_mismatch_result["status"] == "GRAPH_FAILED", source_grant_mismatch_result
        assert not source_grant_mismatch_restored.exists()
        invalid_source = root / "invalid-source"; shutil.copytree(runtime, invalid_source)
        (invalid_source / "m10-outcomes.jsonl").unlink()
        assert invoke(bot, "backup", "create", invalid_source, root / "invalid-source-backup", expected=1, env=env)["status"] == "INPUT_ERROR"
        (backup / "history.jsonl").write_text("tampered\n", encoding="utf-8")
        assert invoke(bot, "backup", "restore", backup, restored, expected=1, env=env)["status"] == "VERIFY_FAILED"
        # Recreate the backup from the unchanged runtime, then restore into a
        # fresh empty directory and validate checksum, inventory, and semantic
        # graph. A non-empty backup target is rejected so stale artifacts can
        # never be mistaken for the current snapshot.
        shutil.rmtree(backup)
        invoke(bot, "backup", "create", runtime, backup, env=env)
        # Each mutation keeps a valid checksum but breaks one upstream link.
        # verifyBackup must reject the snapshot before publishing any restore
        # target; a last-write-wins map or skipped optional store would make
        # one of these cases incorrectly look restorable.
        m00_m05_orphans = [
            ("orphan-action-decision", "actions.jsonl", lambda record: record.update({"decision_id": "missing-decision"})),
            ("orphan-outcome-action", "outcomes.jsonl", lambda record: record["effect_ref"].update({"effect_id": "missing-action"})),
            ("orphan-evaluation-outcome", "evaluations.jsonl", lambda record: record.update({"outcome_ids": ["missing-outcome"]})),
            ("orphan-proposal-evaluation", "proposals.jsonl", lambda record: record.update({"evaluation_ids": ["missing-evaluation"]})),
            ("orphan-review-proposal", "reviews.jsonl", lambda record: record.update({"proposal_id": "missing-proposal"})),
        ]
        for label, filename, change in m00_m05_orphans:
            malformed_backup = root / (label + "-backup")
            malformed_restored = root / (label + "-restored")
            shutil.copytree(backup, malformed_backup)
            rewrite_first_jsonl_record(malformed_backup, filename, change)
            malformed_result = invoke(bot, "backup", "restore", malformed_backup, malformed_restored, expected=1, env=env)
            assert malformed_result["status"] == "VERIFY_FAILED", (label, malformed_result)
            assert not malformed_restored.exists(), label
        missing_manifest_backup = root / "missing-manifest-backup"; shutil.copytree(backup, missing_manifest_backup)
        missing_manifest = json.loads((missing_manifest_backup / "manifest.json").read_text(encoding="utf-8"))
        del missing_manifest["files"]["m10-outcomes.jsonl"]
        missing_manifest["required"].remove("m10-outcomes.jsonl")
        (missing_manifest_backup / "manifest.json").write_text(json.dumps(missing_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", missing_manifest_backup, root / "missing-manifest-restored", expected=1, env=env)["status"] == "VERIFY_FAILED"
        invalid_graph_backup = root / "invalid-graph-backup"; shutil.copytree(backup, invalid_graph_backup)
        invalid_outcome = json.loads((invalid_graph_backup / "m10-outcomes.jsonl").read_text(encoding="utf-8"))
        invalid_outcome["effect_ref"]["effect_id"] = "orphaned-execution-after-checksum"
        invalid_outcome_bytes = (json.dumps(invalid_outcome) + "\n").encode()
        (invalid_graph_backup / "m10-outcomes.jsonl").write_bytes(invalid_outcome_bytes)
        invalid_manifest = json.loads((invalid_graph_backup / "manifest.json").read_text(encoding="utf-8"))
        invalid_manifest["files"]["m10-outcomes.jsonl"]["sha256"] = hashlib.sha256(invalid_outcome_bytes).hexdigest()
        invalid_manifest["files"]["m10-outcomes.jsonl"]["size_bytes"] = len(invalid_outcome_bytes)
        (invalid_graph_backup / "manifest.json").write_text(json.dumps(invalid_manifest), encoding="utf-8")
        invalid_graph_restored = root / "invalid-graph-restored"
        assert invoke(bot, "backup", "restore", invalid_graph_backup, invalid_graph_restored, expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert not invalid_graph_restored.exists()
        orphan_evaluation_backup = root / "orphan-evaluation-backup"; shutil.copytree(backup, orphan_evaluation_backup)
        orphan_outcome = json.loads((orphan_evaluation_backup / "m11-outcomes.jsonl").read_text(encoding="utf-8"))
        orphan_outcome["outcome_id"] = "br18-production-o-missing"
        replace_backup_file(orphan_evaluation_backup, "m11-outcomes.jsonl", (json.dumps(orphan_outcome) + "\n").encode())
        assert invoke(bot, "backup", "restore", orphan_evaluation_backup, root / "orphan-evaluation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        orphan_effect_backup = root / "orphan-effect-backup"; shutil.copytree(backup, orphan_effect_backup)
        orphan_effect = json.loads((orphan_effect_backup / "m11-outcomes.jsonl").read_text(encoding="utf-8"))
        orphan_effect["effect_ref"]["effect_id"] = "missing-m11-execution"
        replace_backup_file(orphan_effect_backup, "m11-outcomes.jsonl", (json.dumps(orphan_effect) + "\n").encode())
        assert invoke(bot, "backup", "restore", orphan_effect_backup, root / "orphan-effect-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        orphan_ledger_outcome_backup = root / "orphan-ledger-outcome-backup"; shutil.copytree(backup, orphan_ledger_outcome_backup)
        def orphan_ledger_outcome_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_LEDGER":
                return False
            entry["artifact"]["outcome_links"] = []
            return True
        rewrite_m11_registry(orphan_ledger_outcome_backup, orphan_ledger_outcome_change)
        orphan_ledger_outcome_result = invoke(bot, "backup", "restore", orphan_ledger_outcome_backup, root / "orphan-ledger-outcome-restored", expected=1, env=env)
        assert orphan_ledger_outcome_result["status"] == "GRAPH_FAILED", (
            "orphan-ledger-outcome-restored",
            orphan_ledger_outcome_result,
        )
        expired_m11_execution_backup = root / "expired-m11-execution-backup"; shutil.copytree(backup, expired_m11_execution_backup)
        rewrite_m11_registry(expired_m11_execution_backup, replace_m11_field("PRODUCTION_EXECUTION_RECORD", "attempted_at", "2026-09-08T00:01:00Z"))
        assert invoke(bot, "backup", "restore", expired_m11_execution_backup, root / "expired-m11-execution-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        premature_m11_authorization_backup = root / "premature-m11-authorization-backup"; shutil.copytree(backup, premature_m11_authorization_backup)
        rewrite_m11_registry(premature_m11_authorization_backup, replace_m11_field("PRODUCTION_EXECUTION_AUTHORIZATION", "authorized_at", "2026-09-07T23:59:59Z"))
        assert invoke(bot, "backup", "restore", premature_m11_authorization_backup, root / "premature-m11-authorization-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        drifted_m11_gate_backup = root / "drifted-m11-gate-backup"; shutil.copytree(backup, drifted_m11_gate_backup)
        rewrite_m11_registry(drifted_m11_gate_backup, replace_m11_field("PRODUCTION_GATE", "executions_total_before", 1))
        assert invoke(bot, "backup", "restore", drifted_m11_gate_backup, root / "drifted-m11-gate-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        pre_activation_gate_backup = root / "pre-activation-gate-backup"; shutil.copytree(backup, pre_activation_gate_backup)
        rewrite_m11_registry(pre_activation_gate_backup, replace_m11_field("PRODUCTION_GATE", "evaluated_at", "2026-09-07T23:59:59Z"))
        assert invoke(bot, "backup", "restore", pre_activation_gate_backup, root / "pre-activation-gate-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        pre_activation_health_backup = root / "pre-activation-health-backup"; shutil.copytree(backup, pre_activation_health_backup)
        rewrite_m11_registry(pre_activation_health_backup, replace_m11_health_observed_at("2026-09-07T23:59:59Z"))
        assert invoke(bot, "backup", "restore", pre_activation_health_backup, root / "pre-activation-health-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        activation_outside_lease_backup = root / "activation-outside-lease-backup"; shutil.copytree(backup, activation_outside_lease_backup)
        rewrite_m11_registry(activation_outside_lease_backup, replace_m11_field("PRODUCTION_ACTIVATION", "activated_at", "2026-09-07T01:00:00Z"))
        assert invoke(bot, "backup", "restore", activation_outside_lease_backup, root / "activation-outside-lease-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        health_outside_lease_backup = root / "health-outside-lease-backup"; shutil.copytree(backup, health_outside_lease_backup)
        rewrite_m11_registry(health_outside_lease_backup, replace_m11_health_observed_at("2099-09-03T02:50:00Z"))
        assert invoke(bot, "backup", "restore", health_outside_lease_backup, root / "health-outside-lease-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        gate_outside_lease_backup = root / "gate-outside-lease-backup"; shutil.copytree(backup, gate_outside_lease_backup)
        rewrite_m11_registry(gate_outside_lease_backup, replace_m11_field("PRODUCTION_GATE", "evaluated_at", "2099-09-03T02:50:00Z"))
        assert invoke(bot, "backup", "restore", gate_outside_lease_backup, root / "gate-outside-lease-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        gate_outside_cost_backup = root / "gate-outside-cost-backup"; shutil.copytree(backup, gate_outside_cost_backup)
        rewrite_m11_registry(gate_outside_cost_backup, replace_m11_field("PRODUCTION_GATE", "evaluated_at", "2099-09-03T02:45:00Z"))
        assert invoke(bot, "backup", "restore", gate_outside_cost_backup, root / "gate-outside-cost-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        authorization_outside_lease_backup = root / "authorization-outside-lease-backup"; shutil.copytree(backup, authorization_outside_lease_backup)
        def authorization_outside_lease_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_EXECUTION_AUTHORIZATION":
                return False
            entry["artifact"]["authorized_at"] = "2099-09-03T02:50:00Z"
            entry["artifact"]["expires_at"] = "2099-09-03T02:51:00Z"
            return True
        rewrite_m11_registry(authorization_outside_lease_backup, authorization_outside_lease_change)
        assert invoke(bot, "backup", "restore", authorization_outside_lease_backup, root / "authorization-outside-lease-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        authorization_stale_health_backup = root / "authorization-stale-health-backup"; shutil.copytree(backup, authorization_stale_health_backup)
        def authorization_stale_health_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_EXECUTION_AUTHORIZATION":
                return False
            entry["artifact"]["authorized_at"] = "2026-09-08T00:01:00Z"
            entry["artifact"]["expires_at"] = "2026-09-08T00:02:00Z"
            return True
        rewrite_m11_registry(authorization_stale_health_backup, authorization_stale_health_change)
        assert invoke(bot, "backup", "restore", authorization_stale_health_backup, root / "authorization-stale-health-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        authorization_health_hash_drift_backup = root / "authorization-health-hash-drift-backup"; shutil.copytree(backup, authorization_health_hash_drift_backup)
        rewrite_m11_registry(authorization_health_hash_drift_backup, replace_m11_field("PRODUCTION_EXECUTION_AUTHORIZATION", "production_health_snapshot_hash", "sha256:" + "0" * 64))
        authorization_health_hash_drift_restored = root / "authorization-health-hash-drift-restored"
        assert invoke(bot, "backup", "restore", authorization_health_hash_drift_backup, authorization_health_hash_drift_restored, expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert not authorization_health_hash_drift_restored.exists()
        execution_health_hash_drift_backup = root / "execution-health-hash-drift-backup"; shutil.copytree(backup, execution_health_hash_drift_backup)
        rewrite_m11_registry(execution_health_hash_drift_backup, replace_m11_field("PRODUCTION_EXECUTION_RECORD", "production_health_snapshot_hash", "sha256:" + "0" * 64))
        execution_health_hash_drift_restored = root / "execution-health-hash-drift-restored"
        assert invoke(bot, "backup", "restore", execution_health_hash_drift_backup, execution_health_hash_drift_restored, expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert not execution_health_hash_drift_restored.exists()
        execution_at_authorization_expiry_backup = root / "execution-at-authorization-expiry-backup"; shutil.copytree(backup, execution_at_authorization_expiry_backup)
        rewrite_m11_registry(execution_at_authorization_expiry_backup, replace_m11_field("PRODUCTION_EXECUTION_AUTHORIZATION", "expires_at", "2026-09-08T00:00:02Z"))
        assert invoke(bot, "backup", "restore", execution_at_authorization_expiry_backup, root / "execution-at-authorization-expiry-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        ledger_before_activation_backup = root / "ledger-before-activation-backup"; shutil.copytree(backup, ledger_before_activation_backup)
        def ledger_before_activation_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_LEDGER" or not entry["artifact"].get("outcome_links"):
                return False
            entry["artifact"]["window_started_at"] = "2026-09-07T23:59:59Z"
            return True
        rewrite_m11_registry(ledger_before_activation_backup, ledger_before_activation_change)
        assert invoke(bot, "backup", "restore", ledger_before_activation_backup, root / "ledger-before-activation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        orphan_m11_reservation_backup = root / "orphan-m11-reservation-backup"; shutil.copytree(backup, orphan_m11_reservation_backup)
        def orphan_m11_reservation_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_LEDGER" or production_failed["artifact"]["execution"]["execution_id"] not in entry["artifact"].get("pending_execution_ids", []):
                return False
            entry["artifact"]["pending_execution_ids"] = []
            entry["artifact"]["pending_outcomes"] = 0
            return True
        rewrite_m11_registry(orphan_m11_reservation_backup, orphan_m11_reservation_change)
        assert invoke(bot, "backup", "restore", orphan_m11_reservation_backup, root / "orphan-m11-reservation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        duplicate_m11_artifact_backup = root / "duplicate-m11-artifact-backup"; shutil.copytree(backup, duplicate_m11_artifact_backup)
        m11_artifact_lines = (duplicate_m11_artifact_backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines()
        duplicate_execution_line = next(line for line in m11_artifact_lines if json.loads(line)["artifact_kind"] == "PRODUCTION_EXECUTION_RECORD")
        replace_backup_file(duplicate_m11_artifact_backup, "m11-artifacts.jsonl", ("\n".join(m11_artifact_lines + [duplicate_execution_line]) + "\n").encode())
        duplicate_m11_artifact_restored = root / "duplicate-m11-artifact-restored"
        duplicate_m11_artifact_result = invoke(bot, "backup", "restore", duplicate_m11_artifact_backup, duplicate_m11_artifact_restored, expected=1, env=env)
        assert duplicate_m11_artifact_result["status"] == "VERIFY_FAILED", duplicate_m11_artifact_result
        assert not duplicate_m11_artifact_restored.exists()
        missing_m11_artifacts = [
            ("PRODUCTION_LEASE", "br18-production-lease"),
            ("PRODUCTION_LEASE_APPROVAL", "br18-production-approval"),
            ("PRODUCTION_HEALTH_SNAPSHOT", "br18-production-health"),
            ("TRUSTED_COST_BOUND", "br18-cost"),
            ("PRODUCTION_GATE", gate["artifact"]["gate_id"]),
            ("PRODUCTION_EXECUTION_AUTHORIZATION", authorization["artifact"]["authorization_id"]),
            ("PRODUCTION_EXECUTION_RECORD", production_failed["artifact"]["execution"]["execution_id"]),
            ("PRODUCTION_OUTCOME_EVALUATION", "br18-production-e"),
        ]
        for missing_kind, missing_id in missing_m11_artifacts:
            missing_name = "missing-" + missing_kind.lower().replace("_", "-")
            missing_backup = root / (missing_name + "-backup")
            missing_restored = root / (missing_name + "-restored")
            shutil.copytree(backup, missing_backup)
            remove_m11_entry(missing_backup, missing_kind, missing_id)
            missing_result = invoke(bot, "backup", "restore", missing_backup, missing_restored, expected=1, env=env)
            assert missing_result["status"] in {"VERIFY_FAILED", "GRAPH_FAILED"}, missing_result
            assert not missing_restored.exists()
        orphan_cycle_backup = root / "orphan-cycle-backup"; shutil.copytree(backup, orphan_cycle_backup)
        rewrite_m11_registry(orphan_cycle_backup, replace_m11_field("PRODUCTION_CYCLE", "evaluation_id", "missing-evaluation"))
        assert invoke(bot, "backup", "restore", orphan_cycle_backup, root / "orphan-cycle-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        missing_cycle_backup = root / "missing-cycle-backup"; shutil.copytree(backup, missing_cycle_backup)
        remove_m11_entry(missing_cycle_backup, "PRODUCTION_CYCLE", "br18-production-cycle")
        missing_cycle_restored = root / "missing-cycle-restored"
        missing_cycle_result = invoke(bot, "backup", "restore", missing_cycle_backup, missing_cycle_restored, expected=1, env=env)
        assert missing_cycle_result["status"] == "GRAPH_FAILED", missing_cycle_result
        assert not missing_cycle_restored.exists()
        missing_failed_outcome_backup = root / "missing-failed-outcome-backup"
        shutil.copytree(backup, missing_failed_outcome_backup)
        # Keep the original failed execution/outcome pair valid, then append a
        # second checksum-valid FAILED execution whose outcome is absent. This
        # reaches the semantic restore guard without failing in the outcome
        # loader or an unrelated terminal-lineage check first.
        add_m11_failed_execution_without_outcome(missing_failed_outcome_backup)
        missing_failed_outcome_restored = root / "missing-failed-outcome-restored"
        missing_failed_outcome_result, missing_failed_outcome_error = invoke_with_stderr(
            bot,
            "backup",
            "restore",
            missing_failed_outcome_backup,
            missing_failed_outcome_restored,
            expected=1,
            env=env,
        )
        if (
            missing_failed_outcome_result["status"] != "GRAPH_FAILED"
            or "failed M11 execution is missing restored fixture outcome" not in missing_failed_outcome_error
            or missing_failed_outcome_restored.exists()
        ):
            raise AssertionError(
                (
                    "missing-failed-outcome restore must fail closed",
                    missing_failed_outcome_result,
                    missing_failed_outcome_error,
                )
            )
        pending_cycle_backup = root / "pending-cycle-backup"; shutil.copytree(backup, pending_cycle_backup)
        rewrite_m11_registry(pending_cycle_backup, replace_m11_field("PRODUCTION_CYCLE", "status", "REVIEW_PENDING"))
        assert invoke(bot, "backup", "restore", pending_cycle_backup, root / "pending-cycle-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        forged_evaluation_evidence_backup = root / "forged-evaluation-evidence-backup"; shutil.copytree(backup, forged_evaluation_evidence_backup)
        rewrite_m11_registry(forged_evaluation_evidence_backup, replace_m11_field("PRODUCTION_OUTCOME_EVALUATION", "evidence_ids", ["forged-evidence-id"]))
        assert invoke(bot, "backup", "restore", forged_evaluation_evidence_backup, root / "forged-evaluation-evidence-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        reversed_cycle_time_backup = root / "reversed-cycle-time-backup"; shutil.copytree(backup, reversed_cycle_time_backup)
        rewrite_m11_registry(reversed_cycle_time_backup, replace_m11_field("PRODUCTION_CYCLE", "closed_at", "2026-09-08T00:00:03Z"))
        assert invoke(bot, "backup", "restore", reversed_cycle_time_backup, root / "reversed-cycle-time-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        orphan_history_cycle_backup = root / "orphan-history-cycle-backup"; shutil.copytree(backup, orphan_history_cycle_backup)
        def orphan_history_change(entry):
            if entry["artifact_kind"] != "PRODUCTION_CYCLE":
                return False
            entry["artifact"]["decision_id"] = "missing-canonical-decision"
            entry["artifact"]["observation_ids"] = ["missing-canonical-observation"]
            return True
        rewrite_m11_registry(orphan_history_cycle_backup, orphan_history_change)
        assert invoke(bot, "backup", "restore", orphan_history_cycle_backup, root / "orphan-history-cycle-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", backup, restored, env=env)["status"] == "RESTORED"
        assert "replay=MATCH" in run([bot, "history", "replay", restored / "history.jsonl"], env=env).stdout
        status = invoke(bot, "mission", "status", restored, env=env)
        restored_state = status["artifact"]
        assert_closed_cycle_snapshot(restored)
        assert restored_state["stop"] is True and restored_state["stop_reason"] == "backup-drill"
        assert restored_state["approval"]["approval_id"] == "br18-ap"
        assert restored_state["canary"]["executions_used"] == 1 and restored_state["canary"]["cost_used_minor"] == 4
        assert (restored / "actions.jsonl").exists() and (restored / "m10-artifacts.jsonl").exists() and (restored / "m10-outcomes.jsonl").exists() and (restored / "m11-artifacts.jsonl").exists() and (restored / "m11-outcomes.jsonl").exists()
        assert invoke(bot, "mission", "m10-resolve", restored, "EXECUTION_RECORD", failed["artifact"]["execution_id"], env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_LEASE", "br18-production-lease", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_ACTIVATION", "br18-production-lease/v1", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_LEDGER", "br18-production-lease/2026-09-08T00:00:00Z", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_EXECUTION_AUTHORIZATION", authorization["artifact"]["authorization_id"], env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_EXECUTION_RECORD", production_failed["artifact"]["execution"]["execution_id"], env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_OUTCOME_EVALUATION", "br18-production-e", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_CYCLE", "br18-production-cycle", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_LEDGER", production_outcome_result["artifact"]["post_ledger"]["lease_id"] + "/" + production_outcome_result["artifact"]["post_ledger"]["updated_at"], env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m10-outcome", restored, machine_outcome, env=env)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m10-reserve", restored, "1", expected=1, env=env)["status"] == "STOPPED"

        # A separate runtime proves the UNKNOWN path: its durable STOP survives
        # human reconciliation, backup/restore, and a fresh process. The old
        # lease is never reactivated by the resolution.
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "PRODUCTION_LEASE", production_lease, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "PRODUCTION_LEASE_APPROVAL", production_approval, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", recovery_runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", recovery_runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "PRODUCTION_HEALTH_SNAPSHOT", production_health, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "TRUSTED_COST_BOUND", cost, env=env)["status"] == "APPENDED"
        recovery_gate = invoke(bot, "mission", "m11-gate", recovery_runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", env=env)
        assert recovery_gate["status"] == "ALLOW_PRODUCTION"
        recovery_authorization = invoke(bot, "mission", "m11-authorize", recovery_runtime, "br18-production-lease", recovery_gate["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:00Z", env=env)
        assert recovery_authorization["status"] == "APPENDED"
        reservation_attempts = [
            subprocess.Popen([str(bot), "mission", "m11-reserve-authorization", str(recovery_runtime), recovery_authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", reserved_at], cwd=ROOT, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
            for reserved_at in ("2026-09-08T00:00:01Z", "2026-09-08T00:00:02Z")
        ]
        reservation_responses = []
        for process in reservation_attempts:
            stdout, stderr = process.communicate()
            if process.returncode not in (0, 1):
                raise AssertionError((stdout, stderr))
            reservation_responses.append(json.loads(stdout))
        assert sum(response["status"] == "APPENDED" for response in reservation_responses) == 1
        assert all(response["status"] in {"APPENDED", "BUSY", "REJECTED"} for response in reservation_responses)
        recovery_reservation = next(response for response in reservation_responses if response["status"] == "APPENDED")
        for reserved_at, response in zip(("2026-09-08T00:00:01Z", "2026-09-08T00:00:02Z"), reservation_responses):
            if response["status"] == "BUSY":
                assert invoke(bot, "mission", "m11-reserve-authorization", recovery_runtime, recovery_authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", reserved_at, expected=1, env=env)["status"] == "REJECTED"
        recovery_ledger_id = recovery_reservation["artifact"]["lease_id"] + "/" + recovery_reservation["artifact"]["updated_at"]
        recovery_unknown = invoke(bot, "mission", "m11-record-unknown", recovery_runtime, recovery_authorization["artifact"]["authorization_id"], recovery_ledger_id, "2026-09-08T00:00:03Z", "fixture provider timeout after dispatch", env=env)
        assert recovery_unknown["status"] == "APPENDED" and recovery_unknown["artifact"]["execution"]["side_effect_state"] == "UNKNOWN"
        stopped_state_backup = root / "stopped-state-backup"; stopped_state_partial = root / "stopped-state-partial"
        assert invoke(bot, "backup", "create", recovery_runtime, stopped_state_backup, env=env)["status"] == "BACKED_UP"
        shutil.copytree(stopped_state_backup, stopped_state_partial)
        partial_state = json.loads((stopped_state_partial / "mission-state.json").read_text(encoding="utf-8")); partial_state["stop"] = False; partial_state["stop_reason"] = ""
        partial_state_bytes = json.dumps(partial_state).encode()
        (stopped_state_partial / "mission-state.json").write_bytes(partial_state_bytes)
        (stopped_state_partial / "STOP").unlink()
        partial_state_manifest = json.loads((stopped_state_partial / "manifest.json").read_text(encoding="utf-8"))
        partial_state_manifest["files"]["mission-state.json"]["sha256"] = hashlib.sha256(partial_state_bytes).hexdigest()
        partial_state_manifest["files"]["mission-state.json"]["size_bytes"] = len(partial_state_bytes)
        del partial_state_manifest["files"]["STOP"]
        (stopped_state_partial / "manifest.json").write_text(json.dumps(partial_state_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", stopped_state_partial, root / "stopped-state-partial-restored", expected=1, env=env)["status"] == "VERIFY_FAILED"
        assert invoke(bot, "mission", "m11-activate", recovery_runtime, "br18-production-lease", "2026-09-08T00:00:03Z", expected=1, env=env)["status"] == "STOPPED"
        resolution = root / "production-resolution.json"; write_production_resolution(resolution, production_lease, recovery_unknown["artifact"]["execution"]["execution_id"])
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "PRODUCTION_RECONCILIATION", resolution, env=env)["status"] == "APPENDED"
        stopped_ledger_id = recovery_unknown["artifact"]["stopped_ledger"]["lease_id"] + "/" + recovery_unknown["artifact"]["stopped_ledger"]["updated_at"]
        recovery_resolution = invoke(bot, "mission", "m11-reconcile", recovery_runtime, "br18-production-resolution", stopped_ledger_id, env=env)
        assert recovery_resolution["status"] == "APPENDED" and recovery_resolution["artifact"]["stopped_ledger"]["control_mode"] == "STOPPED" and recovery_resolution["artifact"]["stopped_ledger"]["reconciliation_required"] is False
        assert invoke(bot, "mission", "m11-reconcile", recovery_runtime, "br18-production-resolution", stopped_ledger_id, env=env)["status"] == "EXACT_DUPLICATE"
        recovery_handoff = root / "recovery-handoff.json"
        reviewed_ledger_id = recovery_resolution["artifact"]["stopped_ledger"]["lease_id"] + "/" + recovery_resolution["artifact"]["stopped_ledger"]["updated_at"]
        handoff = invoke(bot, "mission", "m11-recovery-export", recovery_runtime, "br18-production-resolution", reviewed_ledger_id, recovery_handoff, env=env)
        assert handoff["status"] == "APPENDED" and handoff["artifact"]["requires_new_runtime"] is True and handoff["artifact"]["execution_permitted"] is False
        assert invoke(bot, "backup", "create", recovery_runtime, recovery_backup, env=env)["status"] == "BACKED_UP"
        invalid_reconciliation_backup = root / "invalid-reconciliation-backup"; shutil.copytree(recovery_backup, invalid_reconciliation_backup)
        changed_lines = []
        for line in (invalid_reconciliation_backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines():
            entry = json.loads(line)
            if entry["artifact_kind"] == "PRODUCTION_LEDGER" and entry["artifact"].get("stop_reason") == "RECOVERY_REVIEW_REQUIRED":
                entry["artifact"]["reconciliation_resolution_ids"] = []
                canonical = json.dumps(entry["artifact"], separators=(",", ":"), ensure_ascii=False).encode()
                entry["content_hash"] = "sha256:" + hashlib.sha256(canonical).hexdigest()
            changed_lines.append(json.dumps(entry, separators=(",", ":"), ensure_ascii=False))
        changed_bytes = ("\n".join(changed_lines) + "\n").encode()
        (invalid_reconciliation_backup / "m11-artifacts.jsonl").write_bytes(changed_bytes)
        invalid_reconciliation_manifest = json.loads((invalid_reconciliation_backup / "manifest.json").read_text(encoding="utf-8"))
        invalid_reconciliation_manifest["files"]["m11-artifacts.jsonl"]["sha256"] = hashlib.sha256(changed_bytes).hexdigest()
        invalid_reconciliation_manifest["files"]["m11-artifacts.jsonl"]["size_bytes"] = len(changed_bytes)
        (invalid_reconciliation_backup / "manifest.json").write_text(json.dumps(invalid_reconciliation_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", invalid_reconciliation_backup, root / "invalid-reconciliation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        invalid_reconciliation_effect_backup = root / "invalid-reconciliation-effect-backup"; shutil.copytree(recovery_backup, invalid_reconciliation_effect_backup)
        rewrite_m11_registry(invalid_reconciliation_effect_backup, replace_m11_field("PRODUCTION_RECONCILIATION", "effect_state", "PERFORMED"))
        assert invoke(bot, "backup", "restore", invalid_reconciliation_effect_backup, root / "invalid-reconciliation-effect-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        duplicate_reconciliation_backup = root / "duplicate-reconciliation-backup"; shutil.copytree(recovery_backup, duplicate_reconciliation_backup)
        # Inserting the forged resolution before the valid one ensures a
        # last-write-wins map would still select the valid ID and miss it.
        insert_duplicate_m11_resolution_before_original(duplicate_reconciliation_backup)
        assert invoke(bot, "backup", "restore", duplicate_reconciliation_backup, root / "duplicate-reconciliation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        partial_reconciliation_backup = root / "partial-reconciliation-backup"; shutil.copytree(recovery_backup, partial_reconciliation_backup)
        partial_lines = []
        for line in (partial_reconciliation_backup / "m11-artifacts.jsonl").read_text(encoding="utf-8").splitlines():
            entry = json.loads(line)
            if entry["artifact_kind"] == "PRODUCTION_LEDGER" and entry["artifact"].get("stop_reason") == "RECOVERY_REVIEW_REQUIRED":
                continue
            partial_lines.append(line)
        partial_bytes = ("\n".join(partial_lines) + "\n").encode()
        (partial_reconciliation_backup / "m11-artifacts.jsonl").write_bytes(partial_bytes)
        partial_manifest = json.loads((partial_reconciliation_backup / "manifest.json").read_text(encoding="utf-8"))
        partial_manifest["files"]["m11-artifacts.jsonl"]["sha256"] = hashlib.sha256(partial_bytes).hexdigest()
        partial_manifest["files"]["m11-artifacts.jsonl"]["size_bytes"] = len(partial_bytes)
        (partial_reconciliation_backup / "manifest.json").write_text(json.dumps(partial_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", partial_reconciliation_backup, root / "partial-reconciliation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", recovery_backup, recovery_restored, env=env)["status"] == "RESTORED"
        assert invoke(bot, "mission", "status", recovery_restored, env=env)["artifact"]["stop"] is True
        assert invoke(bot, "mission", "m11-resolve", recovery_restored, "PRODUCTION_RECONCILIATION", "br18-production-resolution", env=env)["status"] == "RESOLVED"
        assert_resolved_stop_snapshot(recovery_restored, "br18-production-resolution", recovery_unknown["artifact"]["execution"]["execution_id"])
        assert invoke(bot, "mission", "m11-activate", recovery_restored, "br18-production-lease", "2026-09-08T00:00:04Z", expected=1, env=env)["status"] == "STOPPED"

        # Re-admission starts only from the restored, durably stopped runtime.
        # The new runtime has a separately reviewed lease/approval and fresh
        # ledger; it inherits neither old budget nor old approval state.
        admission_runtime = root / "recovery-admission-runtime"
        assert invoke(bot, "mission", "init", admission_runtime, env=env)["status"] == "INITIALIZED"
        shutil.copy2(recovery_restored / "history.jsonl", admission_runtime / "history.jsonl")
        assert invoke(bot, "mission", "bind", admission_runtime, intent, policy, env=env)["status"] == "BOUND"
        admission_lease = root / "recovery-admission-lease.json"; write_recovery_admission_lease(admission_lease, production_lease)
        admission_approval = root / "recovery-admission-approval.json"; write_production_lease_approval(admission_approval, admission_lease)
        admission_lease_value = json.loads(admission_lease.read_text(encoding="utf-8"))
        assert invoke(bot, "mission", "m11-register", recovery_restored, "PRODUCTION_LEASE", admission_lease, expected=1, env=env)["status"] == "STOPPED"
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_LEASE", admission_lease, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_LEASE_APPROVAL", admission_approval, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", admission_runtime, admission_lease_value["lease_id"], "2026-09-08T00:00:07Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", admission_runtime, admission_lease_value["lease_id"], "2026-09-08T00:00:07Z", env=env)["status"] == "APPENDED"
        restored_handoff = root / "restored-recovery-handoff.json"
        assert invoke(bot, "mission", "m11-recovery-export", recovery_restored, "br18-production-resolution", reviewed_ledger_id, restored_handoff, env=env)["status"] == "APPENDED"
        admission_input = root / "recovery-admission.json"
        admission_input.write_text(json.dumps({"recovery_admission_id": "br18-recovery-admission", "prior_runtime_dir": str(recovery_restored.resolve()), "prior_lease_id": handoff["artifact"]["prior_lease_id"], "prior_lease_version": handoff["artifact"]["prior_lease_version"], "prior_lease_hash": handoff["artifact"]["prior_lease_hash"], "prior_approval_id": handoff["artifact"]["prior_approval_id"], "resolution_id": handoff["artifact"]["resolution_id"], "new_runtime_id": "br18-recovery-runtime-v2", "new_runtime_dir": str(admission_runtime.resolve()), "new_lease_id": admission_lease_value["lease_id"], "new_lease_version": admission_lease_value["lease_version"], "new_lease_hash": admission_lease_value["lease_hash"], "new_approval_id": admission_lease_value["approval_ref"], "reviewed_by": "human", "reviewer_id": "pilot-human", "reviewed_at": "2026-09-08T00:00:07Z", "execution_permitted": False}), encoding="utf-8")
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_RECOVERY_ADMISSION", admission_input, expected=1, env=env)["status"] == "REJECTED"
        admission = invoke(bot, "mission", "m11-recovery-admit", admission_runtime, recovery_restored, restored_handoff, admission_input, env=env)
        assert admission["status"] == "APPENDED" and admission["execution_permitted"] is False
        duplicate_admission = root / "duplicate-recovery-admission.json"
        duplicate_payload = json.loads(admission_input.read_text(encoding="utf-8")); duplicate_payload["recovery_admission_id"] = "br18-recovery-admission-duplicate"
        duplicate_admission.write_text(json.dumps(duplicate_payload), encoding="utf-8")
        assert invoke(bot, "mission", "m11-recovery-admit", admission_runtime, recovery_restored, restored_handoff, duplicate_admission, expected=1, env=env)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-recovery-admit", admission_runtime, recovery_restored, restored_handoff, admission_input, env=env)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m11-recovery-admit", recovery_restored, recovery_restored, restored_handoff, admission_input, expected=1, env=env)["status"] == "PATH_ERROR"
        admission_backup, admission_restored = root / "recovery-admission-backup", root / "recovery-admission-restored"
        assert invoke(bot, "backup", "create", admission_runtime, admission_backup, env=env)["status"] == "BACKED_UP"
        broken_admission_backup = root / "broken-recovery-admission-backup"; shutil.copytree(admission_backup, broken_admission_backup)
        rewrite_m11_registry(broken_admission_backup, replace_m11_field("PRODUCTION_RECOVERY_ADMISSION", "new_lease_hash", "sha256:" + "f" * 64))
        broken_admission_result = invoke(bot, "backup", "restore", broken_admission_backup, root / "broken-recovery-admission-restored", expected=1, env=env)
        assert broken_admission_result["status"] == "VERIFY_FAILED", broken_admission_result
        broken_approval_backup = root / "broken-recovery-approval-backup"; shutil.copytree(admission_backup, broken_approval_backup)
        rewrite_m11_registry(broken_approval_backup, replace_m11_field("PRODUCTION_LEASE_APPROVAL", "lease_hash", "sha256:" + "e" * 64))
        broken_approval_result = invoke(bot, "backup", "restore", broken_approval_backup, root / "broken-recovery-approval-restored", expected=1, env=env)
        assert broken_approval_result["status"] == "VERIFY_FAILED", broken_approval_result
        missing_approval_backup = root / "missing-recovery-approval-backup"; shutil.copytree(admission_backup, missing_approval_backup)
        remove_m11_entry(missing_approval_backup, "PRODUCTION_LEASE_APPROVAL", admission_lease_value["approval_ref"])
        missing_approval_result = invoke(bot, "backup", "restore", missing_approval_backup, root / "missing-recovery-approval-restored", expected=1, env=env)
        # The edited snapshot carries a matching file digest but not a matching
        # manifest-required runtime graph, so restore must stop at its earlier
        # verification boundary. The direct learner-loader regression covers
        # the same approval-less admission at canonical graph validation.
        assert missing_approval_result["status"] == "VERIFY_FAILED", missing_approval_result
        missing_activation_backup = root / "missing-recovery-activation-backup"; shutil.copytree(admission_backup, missing_activation_backup)
        remove_m11_entry(missing_activation_backup, "PRODUCTION_ACTIVATION", admission_lease_value["lease_id"] + "/" + admission_lease_value["lease_version"])
        missing_activation_result = invoke(bot, "backup", "restore", missing_activation_backup, root / "missing-recovery-activation-restored", expected=1, env=env)
        assert missing_activation_result["status"] == "GRAPH_FAILED", missing_activation_result
        missing_ledger_backup = root / "missing-recovery-ledger-backup"; shutil.copytree(admission_backup, missing_ledger_backup)
        remove_m11_entry(missing_ledger_backup, "PRODUCTION_LEDGER", admission_lease_value["lease_id"] + "/2026-09-08T00:00:07Z")
        assert invoke(bot, "backup", "restore", missing_ledger_backup, root / "missing-recovery-ledger-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", admission_backup, admission_restored, env=env)["status"] == "RESTORED"
        assert invoke(bot, "mission", "m11-resolve", admission_restored, "PRODUCTION_RECOVERY_ADMISSION", "br18-recovery-admission", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-authorize", admission_restored, admission_lease_value["lease_id"], "missing-gate", "fixture_stub", "2026-09-08T00:00:08Z", expected=1, env=env)["status"] == "REJECTED"
        invalid_backup = root / "invalid-backup"; invalid_restored = root / "invalid-restored"; shutil.copytree(backup, invalid_backup)
        invalid_state = json.loads((invalid_backup / "mission-state.json").read_text(encoding="utf-8")); invalid_state["canary"]["executions_used"] = -1
        invalid_state_bytes = json.dumps(invalid_state).encode()
        (invalid_backup / "mission-state.json").write_bytes(invalid_state_bytes)
        invalid_manifest = json.loads((invalid_backup / "manifest.json").read_text(encoding="utf-8"))
        invalid_manifest["files"]["mission-state.json"]["sha256"] = hashlib.sha256(invalid_state_bytes).hexdigest()
        invalid_manifest["files"]["mission-state.json"]["size_bytes"] = len(invalid_state_bytes)
        (invalid_backup / "manifest.json").write_text(json.dumps(invalid_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", invalid_backup, invalid_restored, expected=1, env=env)["status"] == "VERIFY_FAILED"
    print("BR-18b PASS: runtime-created M00-M05/M10 graph, M11 fixture evaluation/cycle, full restored CLOSED cycle, and UNKNOWN-to-human-reconciliation chain use a typed v3 manifest; checksum, exact inventory, M00-M05 orphan links, duplicate M11 artifact identity, lease-window activation, activation-bound health, authorization/execution lineage, broken evaluation/cycle links, reversed cycle time, restart, resolved STOP, and durable STOP are verified")


if __name__ == "__main__":
    main()
