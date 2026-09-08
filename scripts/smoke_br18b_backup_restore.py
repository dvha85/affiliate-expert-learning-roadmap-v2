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


def write_production_lease(path):
    payload = {
        "lease_id": "br18-production-lease", "lease_version": "v1", "policy_version": "br18-v1",
        "approval_ref": "br18-production-approval", "reviewed_by": "human", "reviewer_id": "pilot-human",
        "reviewed_at": "2026-09-07T01:05:00Z", "promotion_review_ref": "fixture:br18-promotion-review",
        "source_canary_grant_id": "br18-g", "source_canary_grant_version": "v1",
        "source_canary_grant_hash": "sha256:" + "a" * 64, "valid_from": "2026-09-07T01:05:00Z",
        "expires_at": "2099-09-03T02:50:00Z", "allowed_risk_classes": ["RISK0"],
        "allowed_action_types": ["DRAFT"], "allowed_hosts": ["example.com"], "executor_ids": ["fixture_stub"],
        "max_executions_total": 1, "max_executions_per_window": 1, "window_seconds": 60,
        "max_cost_minor_total": 4, "currency": "USD", "max_pending_outcomes": 1,
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


def write_production_health(path, lease_path):
    lease = json.loads(lease_path.read_text(encoding="utf-8"))
    payload = {
        "snapshot_id": "br18-production-health", "lease_id": lease["lease_id"], "lease_version": lease["lease_version"],
        "lease_hash": lease["lease_hash"], "observed_at": "2026-09-08T00:00:00Z", "source_refs": ["fixture:br18-health"],
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
        "resolved_at": "2026-09-08T00:00:03Z", "effect_state": "NOT_PERFORMED",
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
        failed_execution = root / "fixture-failed-execution.json"
        failed = invoke(bot, "mission", "m10-record-failed", runtime, authorization, failed_execution, "2026-09-08T00:01:00Z", "fixture-dispatch-failed-before-executor", env=env)
        assert failed["status"] == "APPENDED" and failed["artifact"]["status"] == "FAILED"
        machine_outcome = root / "machine-outcome.json"
        machine_outcome.write_text(json.dumps({"outcome_id":"br18-machine-o","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":failed["artifact"]["execution_id"]},"observed_at":"2026-09-08T00:02:00Z","status":"CANCELLED","metrics":{},"source_ref":"fixture:m10-outcome/br18-failed"}), encoding="utf-8")
        assert invoke(bot, "mission", "m10-outcome", runtime, machine_outcome, env=env)["status"] == "APPENDED"
        production_lease = root / "production-lease.json"; write_production_lease(production_lease)
        recovery_runtime = root / "reconciliation-runtime"; recovery_backup = root / "reconciliation-backup"; recovery_restored = root / "reconciliation-restored"
        shutil.copytree(runtime, recovery_runtime)
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_LEASE", production_lease, env=env)["status"] == "APPENDED"
        production_approval = root / "production-approval.json"; write_production_lease_approval(production_approval, production_lease)
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_LEASE_APPROVAL", production_approval, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", runtime, "br18-production-lease", "2026-09-08T00:00:00Z", env=env)["status"] == "APPENDED"
        production_health = root / "production-health.json"; write_production_health(production_health, production_lease)
        assert invoke(bot, "mission", "m11-register", runtime, "PRODUCTION_HEALTH_SNAPSHOT", production_health, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", runtime, "TRUSTED_COST_BOUND", cost, env=env)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-gate", runtime, "br18-production-lease", "missing-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", expected=1, env=env)["status"] == "REJECTED"
        gate = invoke(bot, "mission", "m11-gate", runtime, "br18-production-lease", "br18-production-health", "br18-cost", "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:00Z", env=env)
        assert gate["status"] == "ALLOW_PRODUCTION" and gate["artifact"]["execution_authorized"] is False
        authorization = invoke(bot, "mission", "m11-authorize", runtime, "br18-production-lease", gate["artifact"]["gate_id"], "fixture_stub", "2026-09-08T00:00:00Z", env=env)
        assert authorization["status"] == "APPENDED" and authorization["artifact"]["execution_authorized"] is True
        reservation_ledger_id = "br18-production-lease/2026-09-08T00:00:01Z"
        reservation = invoke(bot, "mission", "m11-reserve-authorization", runtime, authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)
        assert reservation["status"] == "APPENDED" and reservation["artifact"]["pending_outcomes"] == 1
        assert invoke(bot, "mission", "m11-reserve-authorization", runtime, authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)["status"] == "EXACT_DUPLICATE"
        production_failed = invoke(bot, "mission", "m11-record-failed", runtime, authorization["artifact"]["authorization_id"], reservation_ledger_id, "2026-09-08T00:00:02Z", "fixture-production-dispatch-failed", env=env)
        assert production_failed["status"] == "APPENDED" and production_failed["artifact"]["execution"]["status"] == "FAILED" and production_failed["artifact"]["execution"]["side_effect_state"] == "NOT_PERFORMED"
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
        invoke(bot, "mission", "m11-stop", runtime, "backup-drill", env=env)
        backup_result = invoke(bot, "backup", "create", runtime, backup, env=env)
        assert backup_result["status"] == "BACKED_UP"
        assert backup_result["artifact"]["version"] == "affiliate-bot-backup/v2"
        assert {"m10-artifacts.jsonl", "m10-outcomes.jsonl", "m11-artifacts.jsonl", "m11-outcomes.jsonl"}.issubset(backup_result["artifact"]["required"])
        invalid_source = root / "invalid-source"; shutil.copytree(runtime, invalid_source)
        (invalid_source / "m10-outcomes.jsonl").unlink()
        assert invoke(bot, "backup", "create", invalid_source, root / "invalid-source-backup", expected=1, env=env)["status"] == "INPUT_ERROR"
        (backup / "history.jsonl").write_text("tampered\n", encoding="utf-8")
        assert invoke(bot, "backup", "restore", backup, restored, expected=1, env=env)["status"] == "VERIFY_FAILED"
        # Recreate the backup from the unchanged runtime, then restore into a
        # fresh directory and validate checksum, inventory, and semantic graph.
        invoke(bot, "backup", "create", runtime, backup, env=env)
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
        invalid_manifest["files"]["m10-outcomes.jsonl"] = hashlib.sha256(invalid_outcome_bytes).hexdigest()
        (invalid_graph_backup / "manifest.json").write_text(json.dumps(invalid_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", invalid_graph_backup, root / "invalid-graph-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", backup, restored, env=env)["status"] == "RESTORED"
        assert "replay=MATCH" in run([bot, "history", "replay", restored / "history.jsonl"], env=env).stdout
        status = invoke(bot, "mission", "status", restored, env=env)
        restored_state = status["artifact"]
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
        recovery_reservation = invoke(bot, "mission", "m11-reserve-authorization", recovery_runtime, recovery_authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:00Z", "2026-09-08T00:00:01Z", env=env)
        assert recovery_reservation["status"] == "APPENDED"
        recovery_unknown = invoke(bot, "mission", "m11-record-unknown", recovery_runtime, recovery_authorization["artifact"]["authorization_id"], "br18-production-lease/2026-09-08T00:00:01Z", "2026-09-08T00:00:02Z", "fixture provider timeout after dispatch", env=env)
        assert recovery_unknown["status"] == "APPENDED" and recovery_unknown["artifact"]["execution"]["side_effect_state"] == "UNKNOWN"
        assert invoke(bot, "mission", "m11-activate", recovery_runtime, "br18-production-lease", "2026-09-08T00:00:03Z", expected=1, env=env)["status"] == "REJECTED"
        resolution = root / "production-resolution.json"; write_production_resolution(resolution, production_lease, recovery_unknown["artifact"]["execution"]["execution_id"])
        assert invoke(bot, "mission", "m11-register", recovery_runtime, "PRODUCTION_RECONCILIATION", resolution, env=env)["status"] == "APPENDED"
        stopped_ledger_id = recovery_unknown["artifact"]["stopped_ledger"]["lease_id"] + "/" + recovery_unknown["artifact"]["stopped_ledger"]["updated_at"]
        recovery_resolution = invoke(bot, "mission", "m11-reconcile", recovery_runtime, "br18-production-resolution", stopped_ledger_id, env=env)
        assert recovery_resolution["status"] == "APPENDED" and recovery_resolution["artifact"]["stopped_ledger"]["control_mode"] == "STOPPED" and recovery_resolution["artifact"]["stopped_ledger"]["reconciliation_required"] is False
        assert invoke(bot, "mission", "m11-reconcile", recovery_runtime, "br18-production-resolution", stopped_ledger_id, env=env)["status"] == "EXACT_DUPLICATE"
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
        invalid_reconciliation_manifest["files"]["m11-artifacts.jsonl"] = hashlib.sha256(changed_bytes).hexdigest()
        (invalid_reconciliation_backup / "manifest.json").write_text(json.dumps(invalid_reconciliation_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", invalid_reconciliation_backup, root / "invalid-reconciliation-restored", expected=1, env=env)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", recovery_backup, recovery_restored, env=env)["status"] == "RESTORED"
        assert invoke(bot, "mission", "status", recovery_restored, env=env)["artifact"]["stop"] is True
        assert invoke(bot, "mission", "m11-resolve", recovery_restored, "PRODUCTION_RECONCILIATION", "br18-production-resolution", env=env)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-activate", recovery_restored, "br18-production-lease", "2026-09-08T00:00:04Z", expected=1, env=env)["status"] == "REJECTED"
        invalid_backup = root / "invalid-backup"; invalid_restored = root / "invalid-restored"; shutil.copytree(backup, invalid_backup)
        invalid_state = json.loads((invalid_backup / "mission-state.json").read_text(encoding="utf-8")); invalid_state["canary"]["executions_used"] = -1
        invalid_state_bytes = json.dumps(invalid_state).encode()
        (invalid_backup / "mission-state.json").write_bytes(invalid_state_bytes)
        invalid_manifest = json.loads((invalid_backup / "manifest.json").read_text(encoding="utf-8"))
        invalid_manifest["files"]["mission-state.json"] = hashlib.sha256(invalid_state_bytes).hexdigest()
        (invalid_backup / "manifest.json").write_text(json.dumps(invalid_manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", invalid_backup, invalid_restored, expected=1, env=env)["status"] == "VERIFY_FAILED"
    print("BR-18b PASS: runtime-created M10 graph, governed M11 failed-fixture outcome, and UNKNOWN-to-human-reconciliation chain use a v2 manifest; checksum, required inventory, orphaned outcome, restart, and durable STOP are verified")


if __name__ == "__main__":
    main()
