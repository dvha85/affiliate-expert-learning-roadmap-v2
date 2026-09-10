"""One shared-artifact BR-16a chain from M00 through M11 learner entrypoints.

Pass --workspace to retain the generated fixture artifacts for the beginner
walkthrough. The supplied directory must be empty; the smoke never removes a
caller-owned workspace.
"""
import argparse
import contextlib
import json
import hashlib
import os
import shutil
import socket
import subprocess
import tempfile
import time
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BOT_DIR = ROOT / "lab/affiliate-bot"
REGISTRY = ROOT / "lab/mission-runtime/testdata/m07-registry.json"


def workspace_context(workspace):
    """Return an owned temporary workspace or a validated retained workspace."""
    if workspace is None:
        return tempfile.TemporaryDirectory(prefix="br16a-shared-")
    root = Path(workspace).expanduser().resolve()
    if root.exists():
        if not root.is_dir() or any(root.iterdir()):
            raise ValueError("--workspace must name an empty directory")
    else:
        root.mkdir(parents=True)
    return contextlib.nullcontext(str(root))


def run(command, cwd=ROOT, expected=0, env=None):
    result = subprocess.run([str(x) for x in command], cwd=cwd, text=True, capture_output=True, env=env)
    if result.returncode != expected:
        raise AssertionError((command, result.stdout, result.stderr))
    return result


def invoke(bot, *args, expected=0, env=None):
    result = run([bot, *args], expected=expected, env=env)
    return json.loads(result.stdout)


def start_m07_adapter(bot, history, env):
    listener = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    listener.bind(("127.0.0.1", 0))
    address = f"127.0.0.1:{listener.getsockname()[1]}"
    listener.close()
    process = subprocess.Popen([str(bot), "watcher", "serve", str(history), address], cwd=ROOT, text=True, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, env=env)
    base_url = f"http://{address}"
    for _ in range(50):
        try:
            with urllib.request.urlopen(base_url + "/healthz", timeout=0.2) as response:
                if response.status == 200:
                    return process, base_url
        except OSError:
            time.sleep(0.05)
    process.terminate()
    _, stderr = process.communicate(timeout=2)
    raise AssertionError(("M07 adapter did not start", stderr))


def call_m07_adapter(base_url, endpoint, payload):
    request = urllib.request.Request(base_url + endpoint, data=json.dumps(payload).encode(), method="POST", headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(request, timeout=2) as response:
        if response.status != 200:
            raise AssertionError((endpoint, response.status))
        return json.loads(response.read())


def stop_m07_adapter(process):
    process.terminate()
    try:
        process.wait(timeout=2)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait(timeout=2)


def write_cost_bound(path, intent, amount, expires_at, bound_id):
    payload = {
        "cost_bound_id": bound_id, "intent_id": intent["intent_id"], "intent_hash": intent["intent_hash"],
        "max_cost_minor": amount, "currency": "USD", "source_ref": "fixture:br16-cost-registry",
        "observed_at": "2026-09-07T01:06:00Z", "expires_at": expires_at,
        "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    digest = hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    payload["cost_bound_hash"] = "sha256:" + digest
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_canary_grant(path, intent, policy, approval, max_executions, max_cost, executor_ids=None):
    if executor_ids is None:
        executor_ids = ["local_sandbox"]
    payload = {
        "grant_id": "br16-g", "grant_version": "v1", "policy_version": policy["policy_version"],
        "approval_ref": approval["approval_id"], "approved_by": "human", "approver_id": approval["approver_id"],
        "approved_at": approval["approved_at"], "valid_from": approval["approved_at"], "expires_at": approval["expires_at"],
        "allowed_risk_classes": [policy["risk_class"]], "allowed_action_types": [intent["action_type"]], "allowed_hosts": ["example.com"],
        "executor_ids": executor_ids, "max_executions_total": max_executions, "max_executions_per_window": max_executions,
        "window_seconds": 3600, "max_cost_minor_total": max_cost, "currency": "USD", "max_pending_outcomes": 1,
        "kill_switch_required": True, "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    payload["grant_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def canonical_m10_snapshot(state):
    """Bytes that a rejected M08-M10 authority operation must not mutate."""
    return {
        name: (state / name).read_bytes()
        for name in ("mission-state.json", "m10-artifacts.jsonl", "trusted-cost-bounds.jsonl")
    }


def write_production_lease(path, grant, policy, intent):
    payload = {
        "lease_id": "br16-production-lease", "lease_version": "v1", "policy_version": policy["policy_version"],
        "approval_ref": "br16-production-approval", "reviewed_by": "human", "reviewer_id": "pilot-human",
        "reviewed_at": "2026-09-07T01:05:00Z", "promotion_review_ref": "fixture:br16-promotion-review",
        "source_canary_grant_id": grant["grant_id"], "source_canary_grant_version": grant["grant_version"],
        "source_canary_grant_hash": grant["grant_hash"], "valid_from": "2026-09-07T01:05:00Z",
        "expires_at": "2099-09-03T02:50:00Z", "allowed_risk_classes": [policy["risk_class"]],
        "allowed_action_types": [intent["action_type"]], "allowed_hosts": ["example.com"], "executor_ids": ["fixture_stub"],
        "max_executions_total": 1, "max_executions_per_window": 1, "window_seconds": 60,
        "max_cost_minor_total": 100, "currency": "USD", "max_pending_outcomes": 1,
        "max_consecutive_failures": 1, "max_outcome_age_seconds": 60, "max_health_snapshot_age_seconds": 60,
        "kill_switch_required": True, "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    payload["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_approval(path, lease):
    payload = {
        "approval_id": lease["approval_ref"], "lease_id": lease["lease_id"], "lease_version": lease["lease_version"],
        "lease_hash": lease["lease_hash"], "promotion_review_ref": lease["promotion_review_ref"],
        "source_canary_grant_id": lease["source_canary_grant_id"], "source_canary_grant_version": lease["source_canary_grant_version"],
        "source_canary_grant_hash": lease["source_canary_grant_hash"], "source_e5_refs": ["fixture:br16-e5"],
        "validated_risk_classes": ["RISK0"], "reviewed_by": "human", "reviewer_id": lease["reviewer_id"],
        "reviewed_at": lease["reviewed_at"], "decision": "APPROVE_PRODUCTION_LEASE",
    }
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_health(path, lease):
    payload = {
        "snapshot_id": "br16-production-health", "lease_id": lease["lease_id"], "lease_version": lease["lease_version"],
        "lease_hash": lease["lease_hash"], "observed_at": "2026-09-08T00:00:00Z", "source_refs": ["fixture:br16-health"],
        "dependency_state": "HEALTHY", "telemetry_complete": True, "consecutive_failures": 0,
        "reconciliation_required": False, "compliance_alert_count": 0, "oldest_pending_outcome_age_seconds": 0,
        "hash_version": "go-json-v1",
    }
    payload["snapshot_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_recovery_lease(path, grant, policy, intent):
    payload = {
        "lease_id": "br16-recovery-lease", "lease_version": "v1", "policy_version": policy["policy_version"],
        "approval_ref": "br16-recovery-approval", "reviewed_by": "human", "reviewer_id": "pilot-human",
        "reviewed_at": "2026-09-07T01:05:00Z", "promotion_review_ref": "fixture:br16-recovery-review",
        "source_canary_grant_id": grant["grant_id"], "source_canary_grant_version": grant["grant_version"],
        "source_canary_grant_hash": grant["grant_hash"], "valid_from": "2026-09-07T01:05:00Z",
        "expires_at": "2099-09-03T02:50:00Z", "allowed_risk_classes": [policy["risk_class"]],
        "allowed_action_types": [intent["action_type"]], "allowed_hosts": ["example.com"], "executor_ids": ["fixture_stub"],
        "max_executions_total": 1, "max_executions_per_window": 1, "window_seconds": 60,
        "max_cost_minor_total": 100, "currency": "USD", "max_pending_outcomes": 1,
        "max_consecutive_failures": 1, "max_outcome_age_seconds": 60, "max_health_snapshot_age_seconds": 60,
        "kill_switch_required": True, "correlation_id": intent["correlation_id"], "hash_version": "go-json-v1",
    }
    payload["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(payload, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
    path.write_text(json.dumps(payload), encoding="utf-8")


def write_production_resolution(path, lease, execution_id):
    payload = {
        "resolution_id": "br16-recovery-resolution", "lease_id": lease["lease_id"],
        "lease_version": lease["lease_version"], "lease_hash": lease["lease_hash"],
        "execution_id": execution_id, "resolved_by": "human", "resolver_id": "pilot-human",
        "resolved_at": "2026-09-08T00:00:05Z", "effect_state": "NOT_PERFORMED",
        "reason": "fixture provider audit confirmed no side effect",
    }
    path.write_text(json.dumps(payload), encoding="utf-8")


def grounded_claim(field, value, evidence_id):
    value = json.loads(json.dumps(value, separators=(',', ':'), ensure_ascii=False, sort_keys=True))
    rendered = f"{field}={json.dumps(value, separators=(',', ':'), ensure_ascii=False, sort_keys=True)} [evidence:{evidence_id}]"
    return rendered, {"text": rendered, "field_or_claim": field, "value": value, "evidence_ids": [evidence_id]}


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workspace", help="empty directory retained after a successful smoke")
    args = parser.parse_args(argv)
    go = shutil.which(os.environ.get("GO_BIN", "go")) or os.environ.get("GO_BIN", "go")
    try:
        workspace = workspace_context(args.workspace)
    except ValueError as error:
        parser.error(str(error))
    with workspace as directory:
        work = Path(directory); bot = work / "bot"; env = dict(os.environ, GOWORK="off", GOCACHE=str(work / "go-cache"))
        run([go, "build", "-o", bot, "./cmd/bot"], BOT_DIR, env=env)
        state = work / "runtime"; state.mkdir()
        history, observations, model = state / "history.jsonl", work / "observations.json", work / "model.json"
        tool_result, registered_tool = work / "tool-result.json", work / "registered-tool-result.json"
        action, actions = work / "action.json", state / "actions.jsonl"
        outcome, outcomes = work / "outcome.json", state / "outcomes.jsonl"
        evaluations, proposals, reviews = [state / name for name in ("evaluations.jsonl", "proposals.jsonl", "reviews.jsonl")]
        # M00 output is consumed by M02, not regenerated by the next smoke.
        imported = invoke(bot, "evidence", "import", ROOT / "examples/m00-import/packet-t2.json", env=env)["artifact"]
        observations.write_text(json.dumps(imported), encoding="utf-8")
        assert "APPENDED" in run([bot, "history", "capture", history, observations, "br16-d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"], env=env).stdout
        context = invoke(bot, "m07", "context", history, "br16-d")["artifact"]
        evidence_id = context["evidence_ids"][0]
        evidence = next(item for item in context["evidence"] if item["evidence_id"] == evidence_id)
        # Synthetic adapter fixture: registration must happen before the model
        # can cite the returned body, even in this offline shared workspace.
        tool_result.write_text(json.dumps({"record_id":"br16-d","tool_call":{"tool_name":"public_http","method":"GET","target":"https://example.com/a"},"status_code":200,"received_at":"2026-09-03T00:01:00Z","redirected":False,"body":{"fixture_price":100}}), encoding="utf-8")
        registry = json.loads(REGISTRY.read_text(encoding="utf-8"))
        adapter, adapter_url = start_m07_adapter(bot, history, env)
        registration = call_m07_adapter(adapter_url, "/v1/m07/register-tool-result", {"record_id": "br16-d", "registry": registry, "tool_result": json.loads(tool_result.read_text(encoding="utf-8"))})
        assert registration["status"] == "ACK"
        registered_tool.write_text(json.dumps(registration["artifact"]), encoding="utf-8")
        tool_evidence = registration["evidence"]
        answer, claim = grounded_claim(tool_evidence["field_or_claim"], tool_evidence["value"], tool_evidence["evidence_id"])
        model.write_text(json.dumps({"state":"HUMAN_REVIEW","answer":answer,"claims":[claim],"evidence_ids":[tool_evidence["evidence_id"]],"tool_calls":[],"authority":"A2-RO","write_permission":False}), encoding="utf-8")
        assert invoke(bot, "m07", "validate", history, "br16-d", model, REGISTRY, registered_tool)["status"] == "VALID"
        action.write_text(json.dumps({"action_id":"br16-a","decision_id":"br16-d","action_type":"synthetic_manual","target":"fixture:br16","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True}), encoding="utf-8")
        assert invoke(bot, "action", "record", history, actions, action)["status"] == "APPENDED"
        outcome.write_text(json.dumps({"outcome_id":"br16-o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"br16-a"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:br16"}), encoding="utf-8")
        assert invoke(bot, "outcome", "import", history, actions, outcomes, outcome)["status"] == "APPENDED"
        config = work / "evaluation.json"; config.write_text(json.dumps({"evaluation_id":"br16-e","decision_id":"br16-d","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"br16-a"},"outcome_ids":["br16-o"],"evaluated_at":"2026-09-06T00:00:00Z"}), encoding="utf-8")
        assert invoke(bot, "evaluation", "create", history, actions, outcomes, evaluations, config)["status"] == "APPENDED"

        # The subsequent M08 intent must come from the persisted M07 proposal,
        # not a caller-declared human intent or a replacement target.
        proposal_model, agent_proposal = work / "agent-model.json", work / "agent-proposal.json"
        proposal_answer, proposal_claim = grounded_claim(evidence["field_or_claim"], evidence["value"], evidence_id)
        proposal_model.write_text(json.dumps({"state":"HUMAN_REVIEW","answer":proposal_answer,"claims":[proposal_claim],"evidence_ids":[evidence_id],"tool_calls":[],"authority":"A2-RO","write_permission":False,"proposed_action":{"action_type":"DRAFT","target":"https://example.com/draft","parameters":{"id":9007199254740993}}}), encoding="utf-8")
        proposal = call_m07_adapter(adapter_url, "/v1/m07/register-proposal", {"record_id": "br16-d", "registry": registry, "model_output": json.loads(proposal_model.read_text(encoding="utf-8"))})
        stop_m07_adapter(adapter)
        assert proposal["status"] == "ACK"
        agent_proposal.write_text(json.dumps(proposal["artifact"]), encoding="utf-8")
        proposal_id = proposal["artifact"]["proposal_id"]
        request = work / "intent-request.json"; intent = work / "intent.json"; policy_config = work / "policy.json"; policy = work / "policy-output.json"
        request.write_text(json.dumps({"intent_id":"br16-i","decision_id":"br16-d","evidence_ids":[evidence_id],"action_type":"DRAFT","target":"https://example.com/draft","parameters":{"id":9007199254740993},"proposed_by":"agent","proposal_ref":proposal_id,"created_at":"2026-09-07T00:00:00Z","expires_at":"2099-09-03T03:00:00Z","correlation_id":"br16-c","idempotency_key":"br16-k"}), encoding="utf-8")
        tampered_request = work / "intent-request-tampered.json"
        tampered_request.write_text(request.read_text(encoding="utf-8").replace("https://example.com/draft", "https://example.com/changed"), encoding="utf-8")
        assert invoke(bot, "mission", "m08-intent", history, tampered_request, agent_proposal, work / "tampered-intent.json", expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m08-intent", history, request, agent_proposal, intent)["status"] == "APPENDED"
        policy_config.write_text(json.dumps({"policy_version":"br16-v1","now":"2026-09-07T01:00:00Z","allowed_hosts":["example.com"],"action_risk":{"DRAFT":"RISK0"},"seen_idempotency":{}}), encoding="utf-8")
        assert invoke(bot, "mission", "m08-policy", history, intent, policy_config, agent_proposal, policy)["status"] == "ALLOW"
        assert invoke(bot, "mission", "bind", state, intent, policy)["status"] == "BOUND"
        i = json.loads(intent.read_text()); p = json.loads(policy.read_text())
        approval = work / "approval.json"; approval.write_text(json.dumps({"approval_id":"br16-ap","intent_id":i["intent_id"],"intent_hash":i["intent_hash"],"policy_version":p["policy_version"],"decision":"APPROVE","approved_by":"human","approver_id":"pilot-human","approved_at":"2026-09-07T01:05:00Z","expires_at":"2099-09-03T02:50:00Z","correlation_id":i["correlation_id"],"one_time":True}), encoding="utf-8")
        assert invoke(bot, "mission", "m09-approval", state, approval)["status"] == "ACK"
        grant = work / "grant.json"; write_canary_grant(grant, i, p, json.loads(approval.read_text()), 3, 300, ["fixture_stub", "local_sandbox"])
        assert invoke(bot, "mission", "m10-canary", state, grant)["status"] == "ACK"
        cost = work / "cost-bound.json"; write_cost_bound(cost, i, 100, "2099-09-03T02:45:00Z", "br16-cost")
        assert invoke(bot, "mission", "m10-reserve", state, cost, "unregistered", expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m10-cost-register", state, cost)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m10-cost-register", state, cost)["status"] == "EXACT_DUPLICATE"

        # The shared M07→M08→M10 runtime must reject a same-ID rebind before
        # it can discard the existing approval/grant lineage. Each rejected
        # reimport below is checked against the canonical state bytes, not only
        # against its response status.
        rebind_request = work / "intent-request-rebind.json"
        rebind_request_data = json.loads(request.read_text(encoding="utf-8"))
        rebind_request_data["expires_at"] = "2099-09-03T04:00:00Z"
        rebind_request.write_text(json.dumps(rebind_request_data), encoding="utf-8")
        rebind_intent, rebind_policy = work / "intent-rebind.json", work / "policy-rebind.json"
        assert invoke(bot, "mission", "m08-intent", history, rebind_request, agent_proposal, rebind_intent)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m08-policy", history, rebind_intent, policy_config, agent_proposal, rebind_policy)["status"] == "ALLOW"
        before_rebind = canonical_m10_snapshot(state)
        assert invoke(bot, "mission", "bind", state, rebind_intent, rebind_policy, expected=1)["status"] == "REJECTED"
        assert canonical_m10_snapshot(state) == before_rebind

        altered_grant = work / "grant-rebind-cap.json"
        altered_grant_data = json.loads(grant.read_text(encoding="utf-8"))
        altered_grant_data["max_cost_minor_total"] += 1
        altered_grant_data.pop("grant_hash")
        altered_grant_data["grant_hash"] = "sha256:" + hashlib.sha256(json.dumps(altered_grant_data, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        altered_grant.write_text(json.dumps(altered_grant_data), encoding="utf-8")
        assert invoke(bot, "mission", "m10-canary", state, altered_grant, expected=1)["status"] == "REJECTED"
        assert canonical_m10_snapshot(state) == before_rebind

        altered_cost = work / "cost-bound-rebind-currency.json"
        altered_cost_data = json.loads(cost.read_text(encoding="utf-8"))
        altered_cost_data["currency"] = "VND"
        altered_cost_data.pop("cost_bound_hash")
        altered_cost_data["cost_bound_hash"] = "sha256:" + hashlib.sha256(json.dumps(altered_cost_data, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        altered_cost.write_text(json.dumps(altered_cost_data), encoding="utf-8")
        assert invoke(bot, "mission", "m10-cost-register", state, altered_cost, expected=1)["status"] == "REJECTED"
        assert canonical_m10_snapshot(state) == before_rebind

        gate = work / "canary-gate.json"
        gate_time = "2026-09-08T00:00:00Z"
        gate_response = invoke(bot, "mission", "m10-gate", state, cost, gate, gate_time)
        assert gate_response["status"] == "ALLOW_CANARY" and gate_response["artifact"]["execution_authorized"] is False
        assert invoke(bot, "mission", "m10-gate", state, cost, gate, gate_time)["status"] == "EXACT_DUPLICATE"
        forged_gate = work / "forged-canary-gate.json"
        forged_gate_data = json.loads(gate.read_text(encoding="utf-8")); forged_gate_data["gate_id"] = "gate-forged-but-schema-valid"
        forged_gate.write_text(json.dumps(forged_gate_data), encoding="utf-8")
        assert invoke(bot, "mission", "m10-authorize", state, cost, forged_gate, work / "forged-authorization.json", gate_time, "local_sandbox", expected=1)["status"] == "REJECTED"
        authorization = work / "canary-authorization.json"
        authorization_response = invoke(bot, "mission", "m10-authorize", state, cost, gate, authorization, gate_time, "local_sandbox")
        assert authorization_response["status"] == "AUTHORIZED" and authorization_response["artifact"]["execution_authorized"] is True
        assert invoke(bot, "mission", "m10-authorize", state, cost, gate, authorization, gate_time, "local_sandbox")["status"] == "EXACT_DUPLICATE"
        failed_authorization = work / "fixture-failed-authorization.json"
        failed_authorization_response = invoke(bot, "mission", "m10-authorize", state, cost, gate, failed_authorization, gate_time, "fixture_stub")
        assert failed_authorization_response["status"] == "AUTHORIZED" and failed_authorization_response["artifact"]["executor_id"] == "fixture_stub"
        cancelled = work / "cancelled-execution.json"
        forged_authorization = work / "forged-authorization.json"
        forged_authorization_data = json.loads(authorization.read_text(encoding="utf-8")); forged_authorization_data["authorization_id"] = "canary-auth-forged-but-schema-valid"
        forged_authorization.write_text(json.dumps(forged_authorization_data), encoding="utf-8")
        assert invoke(bot, "mission", "m10-cancel", state, forged_authorization, work / "forged-execution.json", "2026-09-08T00:01:00Z", "must-not-resolve-forged-authorization", expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m10-cancel", state, authorization, work / "unreserved-execution.json", "2026-09-08T00:01:00Z", "must-reserve-before-execution-record", expected=1)["status"] == "REJECTED"
        reservation = invoke(bot, "mission", "m10-reserve-authorization", state, authorization, "br16-governed-r1")
        assert reservation["status"] == "RESERVED" and reservation["artifact"]["authorization_id"] == authorization_response["artifact"]["authorization_id"] and reservation["artifact"]["reservation_mode"] == "GOVERNED_AUTHORIZATION"
        cancellation_attempted_at = reservation["artifact"]["reserved_at"]
        assert invoke(bot, "mission", "m10-reserve-authorization", state, authorization, "br16-governed-r1")["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m10-reserve-authorization", state, authorization, "br16-governed-r2", expected=1)["status"] == "REJECTED"
        failed_reservation = invoke(bot, "mission", "m10-reserve-authorization", state, failed_authorization, "br16-governed-r2")
        assert failed_reservation["status"] == "RESERVED" and failed_reservation["artifact"]["authorization_id"] == failed_authorization_response["artifact"]["authorization_id"]
        failed_attempted_at = failed_reservation["artifact"]["reserved_at"]
        cancellation = invoke(bot, "mission", "m10-cancel", state, authorization, cancelled, cancellation_attempted_at, "learner-cancelled-before-executor")
        assert cancellation["status"] == "APPENDED" and cancellation["artifact"]["status"] == "CANCELLED" and cancellation["artifact"]["side_effect_state"] == "NOT_PERFORMED"
        assert invoke(bot, "mission", "m10-cancel", state, authorization, cancelled, cancellation_attempted_at, "learner-cancelled-before-executor")["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m10-cancel", state, authorization, work / "different-execution.json", cancellation_attempted_at, "different-execution-is-rejected", expected=1)["status"] == "REJECTED"
        failed_execution = work / "fixture-failed-execution.json"
        failed_execution_conflict = work / "fixture-failed-execution-conflict.json"
        failed_execution_conflict.write_text("keep-this-portable-output", encoding="utf-8")
        failed_conflict = invoke(bot, "mission", "m10-record-failed", state, failed_authorization, failed_execution_conflict, failed_attempted_at, "fixture-dispatch-failed-before-executor", expected=1)
        assert failed_conflict["status"] == "CONFLICT"
        state_after_conflict = invoke(bot, "mission", "status", state)
        bound_reservation = next(item for item in state_after_conflict["artifact"]["reservations"] if item.get("authorization_id") == failed_authorization_response["artifact"]["authorization_id"])
        assert bound_reservation["execution_id"] == failed_conflict["artifact"]["execution_id"]
        failed = invoke(bot, "mission", "m10-record-failed", state, failed_authorization, failed_execution, failed_attempted_at, "fixture-dispatch-failed-before-executor")
        assert failed["status"] == "APPENDED" and failed["artifact"]["status"] == "FAILED" and failed["artifact"]["side_effect_state"] == "NOT_PERFORMED"
        assert invoke(bot, "mission", "m10-record-failed", state, failed_authorization, failed_execution, failed_attempted_at, "fixture-dispatch-failed-before-executor")["status"] == "EXACT_DUPLICATE"
        machine_outcome = work / "machine-outcome.json"
        machine_outcome.write_text(json.dumps({"outcome_id":"br16-machine-o","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":failed["artifact"]["execution_id"]},"observed_at":failed_attempted_at,"status":"CANCELLED","metrics":{},"source_ref":"fixture:m10-outcome/br16-failed"}), encoding="utf-8")
        assert invoke(bot, "mission", "m10-outcome", state, machine_outcome)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m10-outcome", state, machine_outcome)["status"] == "EXACT_DUPLICATE"
        forged_outcome = work / "forged-machine-outcome.json"
        forged_outcome.write_text(machine_outcome.read_text(encoding="utf-8").replace(failed["artifact"]["execution_id"], "canary-exec-orphan"), encoding="utf-8")
        assert invoke(bot, "mission", "m10-outcome", state, forged_outcome, expected=1)["status"] == "ORPHAN_EXECUTION"

        # M11 continues on exactly the same state, intent/policy/approval,
        # canary grant and trusted cost-bound—not a separately constructed lab.
        production_lease_path = work / "production-lease.json"
        grant_value = json.loads(grant.read_text(encoding="utf-8"))
        write_production_lease(production_lease_path, grant_value, p, i)
        production_lease = json.loads(production_lease_path.read_text(encoding="utf-8"))
        production_approval_path = work / "production-approval.json"; write_production_approval(production_approval_path, production_lease)
        production_health_path = work / "production-health.json"; write_production_health(production_health_path, production_lease)
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_LEASE", production_lease_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_LEASE_APPROVAL", production_approval_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", state, production_lease["lease_id"], gate_time)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", state, production_lease["lease_id"], gate_time)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_HEALTH_SNAPSHOT", production_health_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", state, "TRUSTED_COST_BOUND", cost)["status"] == "APPENDED"
        production_gate = invoke(bot, "mission", "m11-gate", state, production_lease["lease_id"], "br16-production-health", "br16-cost", production_lease["lease_id"] + "/" + gate_time, gate_time)
        assert production_gate["status"] == "ALLOW_PRODUCTION" and production_gate["artifact"]["execution_authorized"] is False
        production_authorization = invoke(bot, "mission", "m11-authorize", state, production_lease["lease_id"], production_gate["artifact"]["gate_id"], "fixture_stub", gate_time)
        assert production_authorization["status"] == "APPENDED" and production_authorization["artifact"]["intent_id"] == i["intent_id"]
        production_reservation = invoke(bot, "mission", "m11-reserve-authorization", state, production_authorization["artifact"]["authorization_id"], production_lease["lease_id"] + "/" + gate_time, "2026-09-08T00:00:01Z")
        assert production_reservation["status"] == "APPENDED"
        production_failed = invoke(bot, "mission", "m11-record-failed", state, production_authorization["artifact"]["authorization_id"], production_lease["lease_id"] + "/2026-09-08T00:00:01Z", "2026-09-08T00:00:02Z", "fixture-dispatch-failed-before-executor")
        assert production_failed["status"] == "APPENDED" and production_failed["artifact"]["execution"]["side_effect_state"] == "NOT_PERFORMED"
        production_outcome = work / "production-outcome.json"
        production_outcome.write_text(json.dumps({"outcome_id":"br16-production-o","effect_ref":{"effect_kind":"MACHINE_EXECUTION","effect_id":production_failed["artifact"]["execution"]["execution_id"]},"observed_at":"2026-09-08T00:00:03Z","status":"CANCELLED","metrics":{},"source_ref":"fixture:m11-outcome/br16-failed"}), encoding="utf-8")
        execution_ledger = production_failed["artifact"]["execution_ledger"]
        execution_ledger_id = execution_ledger["lease_id"] + "/" + execution_ledger["updated_at"]
        production_outcome_result = invoke(bot, "mission", "m11-outcome", state, production_outcome, execution_ledger_id)
        assert production_outcome_result["status"] == "APPENDED"
        production_evaluation = invoke(bot, "mission", "m11-evaluate", state, "br16-production-o", "br16-production-e", "2026-09-08T00:00:04Z")
        assert production_evaluation["status"] == "APPENDED" and production_evaluation["artifact"]["result"] == "FIXTURE_NO_SIDE_EFFECT"
        assert invoke(bot, "mission", "m11-evaluate", state, "br16-production-o", "br16-production-e", "2026-09-08T00:00:04Z")["status"] == "EXACT_DUPLICATE"
        production_cycle = invoke(bot, "mission", "m11-close-cycle", state, "br16-production-cycle", "br16-production-e", "2026-09-08T00:00:05Z")
        assert production_cycle["status"] == "APPENDED" and production_cycle["artifact"]["status"] == "CLOSED"
        assert invoke(bot, "mission", "m11-close-cycle", state, "br16-production-cycle", "br16-production-e", "2026-09-08T00:00:05Z")["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m11-close-cycle", state, "br16-invalid-cycle", "missing-evaluation", "2026-09-08T00:00:05Z", expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-resolve", state, "PRODUCTION_EXECUTION_RECORD", production_failed["artifact"]["execution"]["execution_id"])["status"] == "RESOLVED"

        resolved = invoke(bot, "mission", "m10-resolve", state, "EXECUTION_RECORD", cancellation["artifact"]["execution_id"])
        assert resolved["status"] == "RESOLVED" and resolved["artifact"] == cancellation["artifact"]
        assert invoke(bot, "mission", "m10-resolve", state, "EXECUTION_RECORD", "canary-exec-not-registered", expected=1)["status"] == "REJECTED"
        current_state = invoke(bot, "mission", "status", state)["artifact"]
        governed_reservation = next(item for item in current_state["reservations"] if item["reservation_id"] == "br16-governed-r1")
        assert governed_reservation["execution_id"] == cancellation["artifact"]["execution_id"]
        artifact_kinds = {entry["artifact_kind"] for entry in map(json.loads, (state / "m10-artifacts.jsonl").read_text(encoding="utf-8").splitlines()) if entry}
        assert {"CANARY_GRANT", "TRUSTED_COST_BOUND", "CANARY_GATE", "EXECUTION_AUTHORIZATION", "EXECUTION_RECORD"}.issubset(artifact_kinds)
        tampered = work / "cost-bound-tampered.json"; tampered.write_text(cost.read_text().replace('"max_cost_minor": 100', '"max_cost_minor": 1'), encoding="utf-8")
        assert invoke(bot, "mission", "m10-gate", state, tampered, work / "tampered-gate.json", gate_time, expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m10-reserve", state, tampered, "tampered", expected=1)["status"] == "REJECTED"
        expired = work / "cost-bound-expired.json"; write_cost_bound(expired, i, 100, "2026-09-07T01:07:00Z", "br16-expired")
        assert invoke(bot, "mission", "m10-cost-register", state, expired, expected=1)["status"] == "REJECTED"
        # Two independent legacy-compatible processes race for the one
        # remaining execution/cost budget after the governed attempt above.
        # The directory lock may make one return BUSY; retrying it must then see
        # the committed budget and cannot create a second reservation.
        attempts = [
            subprocess.Popen([str(bot), "mission", "m10-reserve", str(state), str(cost), reservation], cwd=ROOT, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
            for reservation in ("br16-r1", "br16-r2")
        ]
        responses = {}
        for reservation, process in zip(("br16-r1", "br16-r2"), attempts):
            stdout, stderr = process.communicate()
            if process.returncode not in (0, 1):
                raise AssertionError((stdout, stderr))
            responses[reservation] = json.loads(stdout)
        assert sum(response["status"] == "RESERVED" for response in responses.values()) == 1
        assert all(response["status"] in {"RESERVED", "BUSY", "BUDGET_DENIED"} for response in responses.values())
        winner = next(reservation for reservation, response in responses.items() if response["status"] == "RESERVED")
        loser = "br16-r2" if winner == "br16-r1" else "br16-r1"
        assert invoke(bot, "mission", "m10-canary", state, grant)["artifact"]["executions_used"] == 3
        assert invoke(bot, "mission", "m10-reserve", state, cost, winner)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m10-canary", state, grant)["artifact"]["executions_used"] == 3
        assert invoke(bot, "mission", "m10-reserve", state, cost, loser, expected=1)["status"] == "BUDGET_DENIED"
        assert invoke(bot, "mission", "m10-authorize", state, cost, gate, work / "stale-authorization.json", gate_time, "local_sandbox", expected=1)["status"] == "REJECTED"

        # The UNKNOWN path keeps the same workspace, M07 proposal, M08 intent,
        # policy, canary grant and cost bound. It uses a separately reviewed
        # recovery lease because the first production authorization is one-time.
        recovery_lease_path = work / "recovery-lease.json"
        write_recovery_lease(recovery_lease_path, grant_value, p, i)
        recovery_lease = json.loads(recovery_lease_path.read_text(encoding="utf-8"))
        recovery_approval_path = work / "recovery-approval.json"; write_production_approval(recovery_approval_path, recovery_lease)
        recovery_health_path = work / "recovery-health.json"
        recovery_health = json.loads(production_health_path.read_text(encoding="utf-8"))
        recovery_health.update({"snapshot_id": "br16-recovery-health", "lease_id": recovery_lease["lease_id"], "lease_version": recovery_lease["lease_version"], "lease_hash": recovery_lease["lease_hash"]})
        recovery_health.pop("snapshot_hash")
        recovery_health["snapshot_hash"] = "sha256:" + hashlib.sha256(json.dumps(recovery_health, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        recovery_health_path.write_text(json.dumps(recovery_health), encoding="utf-8")
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_LEASE", recovery_lease_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_LEASE_APPROVAL", recovery_approval_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", state, recovery_lease["lease_id"], gate_time)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", state, recovery_lease["lease_id"], gate_time)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_HEALTH_SNAPSHOT", recovery_health_path)["status"] == "APPENDED"
        recovery_ledger_id = recovery_lease["lease_id"] + "/" + gate_time
        recovery_gate = invoke(bot, "mission", "m11-gate", state, recovery_lease["lease_id"], "br16-recovery-health", "br16-cost", recovery_ledger_id, gate_time)
        assert recovery_gate["status"] == "ALLOW_PRODUCTION"
        recovery_authorization = invoke(bot, "mission", "m11-authorize", state, recovery_lease["lease_id"], recovery_gate["artifact"]["gate_id"], "fixture_stub", gate_time)
        assert recovery_authorization["status"] == "APPENDED"
        recovery_reservation = invoke(bot, "mission", "m11-reserve-authorization", state, recovery_authorization["artifact"]["authorization_id"], recovery_ledger_id, "2026-09-08T00:00:01Z")
        assert recovery_reservation["status"] == "APPENDED"
        recovery_reservation_id = recovery_reservation["artifact"]["lease_id"] + "/" + recovery_reservation["artifact"]["updated_at"]
        recovery_unknown = invoke(bot, "mission", "m11-record-unknown", state, recovery_authorization["artifact"]["authorization_id"], recovery_reservation_id, "2026-09-08T00:00:03Z", "fixture provider timeout after dispatch")
        assert recovery_unknown["status"] == "APPENDED" and recovery_unknown["artifact"]["execution"]["side_effect_state"] == "UNKNOWN"
        unknown_status = invoke(bot, "mission", "status", state)["artifact"]
        assert unknown_status["stop"] is True and unknown_status["stop_reason"] == "RECONCILIATION_REQUIRED"
        registry_before_stopped_gate = (state / "m11-artifacts.jsonl").read_bytes()
        assert invoke(bot, "mission", "m11-gate", state, "missing-lease", "missing-health", "missing-cost", "missing-ledger", "2026-09-08T00:00:04Z", expected=1)["status"] == "STOPPED"
        assert (state / "m11-artifacts.jsonl").read_bytes() == registry_before_stopped_gate
        assert invoke(bot, "mission", "m11-activate", state, recovery_lease["lease_id"], "2026-09-08T00:00:04Z", expected=1)["status"] == "STOPPED"
        recovery_resolution_path = work / "recovery-resolution.json"
        write_production_resolution(recovery_resolution_path, recovery_lease, recovery_unknown["artifact"]["execution"]["execution_id"])
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_RECONCILIATION", recovery_resolution_path)["status"] == "APPENDED"
        stopped_ledger_id = recovery_unknown["artifact"]["stopped_ledger"]["lease_id"] + "/" + recovery_unknown["artifact"]["stopped_ledger"]["updated_at"]
        recovery_resolution = invoke(bot, "mission", "m11-reconcile", state, "br16-recovery-resolution", stopped_ledger_id)
        assert recovery_resolution["status"] == "APPENDED" and recovery_resolution["artifact"]["stopped_ledger"]["control_mode"] == "STOPPED"
        assert invoke(bot, "mission", "m11-reconcile", state, "br16-recovery-resolution", stopped_ledger_id)["status"] == "EXACT_DUPLICATE"
        recovery_handoff = work / "recovery-handoff.json"
        reviewed_ledger_id = recovery_resolution["artifact"]["stopped_ledger"]["lease_id"] + "/" + recovery_resolution["artifact"]["stopped_ledger"]["updated_at"]
        handoff = invoke(bot, "mission", "m11-recovery-export", state, "br16-recovery-resolution", reviewed_ledger_id, recovery_handoff)
        assert handoff["status"] == "APPENDED" and handoff["artifact"]["requires_new_runtime"] is True and handoff["artifact"]["execution_permitted"] is False

        # Recovery never reuses the stopped directory or its lease. A separate
        # runtime must first establish its own state, separately reviewed lease,
        # activation, and empty ledger before it can persist a non-authorizing
        # admission that binds the old reviewed handoff.
        admission_runtime = work / "recovery-admission-runtime"
        assert invoke(bot, "mission", "init", admission_runtime)["status"] == "INITIALIZED"
        shutil.copy2(state / "history.jsonl", admission_runtime / "history.jsonl")
        shutil.copytree(state / "history.jsonl.m07", admission_runtime / "history.jsonl.m07")
        assert invoke(bot, "mission", "bind", admission_runtime, intent, policy)["status"] == "BOUND"
        admission_lease = dict(recovery_lease)
        admission_lease.update({"lease_id": "br16-admission-lease", "lease_version": "v1", "approval_ref": "br16-admission-approval", "reviewed_at": "2026-09-08T00:00:06Z", "valid_from": "2026-09-08T00:00:06Z", "promotion_review_ref": "fixture:br16-admission-review"})
        admission_lease.pop("lease_hash")
        admission_lease["lease_hash"] = "sha256:" + hashlib.sha256(json.dumps(admission_lease, separators=(",", ":"), ensure_ascii=False).encode()).hexdigest()
        admission_lease_path = work / "admission-lease.json"; admission_lease_path.write_text(json.dumps(admission_lease), encoding="utf-8")
        admission_approval_path = work / "admission-approval.json"; write_production_approval(admission_approval_path, admission_lease)
        assert invoke(bot, "mission", "m11-register", state, "PRODUCTION_LEASE", admission_lease_path, expected=1)["status"] == "STOPPED"
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_LEASE", admission_lease_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_LEASE_APPROVAL", admission_approval_path)["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-activate", admission_runtime, admission_lease["lease_id"], "2026-09-08T00:00:07Z")["status"] == "APPENDED"
        assert invoke(bot, "mission", "m11-ledger-init", admission_runtime, admission_lease["lease_id"], "2026-09-08T00:00:07Z")["status"] == "APPENDED"
        admission_input = work / "recovery-admission.json"
        admission_input.write_text(json.dumps({"recovery_admission_id": "br16-recovery-admission", "prior_runtime_dir": str(state.resolve()), "prior_lease_id": handoff["artifact"]["prior_lease_id"], "prior_lease_version": handoff["artifact"]["prior_lease_version"], "prior_lease_hash": handoff["artifact"]["prior_lease_hash"], "prior_approval_id": handoff["artifact"]["prior_approval_id"], "resolution_id": handoff["artifact"]["resolution_id"], "new_runtime_id": "br16-recovery-runtime-v2", "new_runtime_dir": str(admission_runtime.resolve()), "new_lease_id": admission_lease["lease_id"], "new_lease_version": admission_lease["lease_version"], "new_lease_hash": admission_lease["lease_hash"], "new_approval_id": admission_lease["approval_ref"], "reviewed_by": "human", "reviewer_id": "pilot-human", "reviewed_at": "2026-09-08T00:00:07Z", "execution_permitted": False}), encoding="utf-8")
        assert invoke(bot, "mission", "m11-register", admission_runtime, "PRODUCTION_RECOVERY_ADMISSION", admission_input, expected=1)["status"] == "REJECTED"
        admission = invoke(bot, "mission", "m11-recovery-admit", admission_runtime, state, recovery_handoff, admission_input)
        assert admission["status"] == "APPENDED" and admission["execution_permitted"] is False and admission["artifact"]["artifact_kind"] == "PRODUCTION_RECOVERY_ADMISSION"
        assert invoke(bot, "mission", "m11-recovery-admit", admission_runtime, state, recovery_handoff, admission_input)["status"] == "EXACT_DUPLICATE"
        assert invoke(bot, "mission", "m11-recovery-admit", state, state, recovery_handoff, admission_input, expected=1)["status"] == "PATH_ERROR"
        reused_lease = dict(json.loads(admission_input.read_text(encoding="utf-8")))
        reused_lease.update({"recovery_admission_id": "br16-reused-lease", "new_lease_id": reused_lease["prior_lease_id"], "new_lease_version": reused_lease["prior_lease_version"], "new_lease_hash": reused_lease["prior_lease_hash"], "new_approval_id": reused_lease["prior_approval_id"]})
        reused_lease_input = work / "reused-recovery-admission.json"; reused_lease_input.write_text(json.dumps(reused_lease), encoding="utf-8")
        assert invoke(bot, "mission", "m11-recovery-admit", admission_runtime, state, recovery_handoff, reused_lease_input, expected=1)["status"] == "REJECTED"
        assert invoke(bot, "mission", "m11-resolve", admission_runtime, "PRODUCTION_RECOVERY_ADMISSION", "br16-recovery-admission")["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-authorize", admission_runtime, admission_lease["lease_id"], "missing-gate", "fixture_stub", "2026-09-08T00:00:08Z", expected=1)["status"] == "REJECTED"
        admission_backup, admission_restored = work / "admission-backup", work / "admission-restored"
        assert invoke(bot, "backup", "create", admission_runtime, admission_backup)["status"] == "BACKED_UP"
        assert invoke(bot, "backup", "restore", admission_backup, admission_restored)["status"] == "RESTORED"
        assert invoke(bot, "mission", "m11-resolve", admission_restored, "PRODUCTION_RECOVERY_ADMISSION", "br16-recovery-admission")["status"] == "RESOLVED"
        assert invoke(bot, "mission", "status", state)["artifact"]["stop"] is True

        # Snapshot the actual shared runtime, not an M08-M11-only directory.
        # Restore must resolve the persisted M07 proposal, human outcome, M10
        # record and M11 recovery graph before the old STOP can be trusted.
        backup, restored = work / "backup", work / "restored"
        assert invoke(bot, "backup", "create", state, backup)["status"] == "BACKED_UP"
        # A checksum-valid edit to the fixture outcome must still fail the
        # canonical evaluation/cycle graph check; manifest integrity alone is
        # deliberately insufficient evidence for a safe restore.
        broken_backup = work / "backup-broken-evaluation-link"
        shutil.copytree(backup, broken_backup)
        outcomes_path = broken_backup / "m11-outcomes.jsonl"
        broken_outcome = json.loads(outcomes_path.read_text(encoding="utf-8"))
        broken_outcome["outcome_id"] = "br16-production-o-missing"
        broken_bytes = (json.dumps(broken_outcome) + "\n").encode()
        outcomes_path.write_bytes(broken_bytes)
        manifest_path = broken_backup / "manifest.json"
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
        manifest["files"]["m11-outcomes.jsonl"]["size_bytes"] = len(broken_bytes)
        manifest["files"]["m11-outcomes.jsonl"]["sha256"] = hashlib.sha256(broken_bytes).hexdigest()
        manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
        assert invoke(bot, "backup", "restore", broken_backup, work / "broken-restored", expected=1)["status"] == "GRAPH_FAILED"
        assert invoke(bot, "backup", "restore", backup, restored)["status"] == "RESTORED"
        restored_history = restored / "history.jsonl"
        assert "replay=MATCH" in run([bot, "history", "replay", restored_history]).stdout
        restored_context = invoke(bot, "m07", "context", restored_history, "br16-d")["artifact"]
        assert proposal_id and evidence_id in restored_context["evidence_ids"]
        restored_status = invoke(bot, "mission", "status", restored)["artifact"]
        assert restored_status["stop"] is True and restored_status["stop_reason"] == "RECONCILIATION_REQUIRED"
        assert invoke(bot, "mission", "m10-resolve", restored, "EXECUTION_RECORD", cancellation["artifact"]["execution_id"])["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_OUTCOME_EVALUATION", "br16-production-e")["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_CYCLE", "br16-production-cycle")["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_EXECUTION_RECORD", recovery_unknown["artifact"]["execution"]["execution_id"])["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_RECONCILIATION", "br16-recovery-resolution")["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-resolve", restored, "PRODUCTION_LEDGER", reviewed_ledger_id)["status"] == "RESOLVED"
        assert invoke(bot, "mission", "m11-activate", restored, recovery_lease["lease_id"], "2026-09-08T00:00:04Z", expected=1)["status"] == "STOPPED"
        restored_handoff = invoke(bot, "mission", "m11-recovery-export", restored, "br16-recovery-resolution", reviewed_ledger_id, work / "restored-recovery-handoff.json")
        assert restored_handoff["status"] == "APPENDED" and restored_handoff["artifact"]["execution_permitted"] is False
        assert invoke(bot, "mission", "m11-stop", state, "br16a-restart-drill")["status"] == "STOPPED"
        # New process, same workspace: replay and durable stop must survive.
        assert "replay=MATCH" in run([bot, "history", "replay", history]).stdout
        status = invoke(bot, "mission", "status", state)
        assert status["artifact"]["stop"] is True
        assert invoke(bot, "mission", "m10-canary", state, grant, expected=1)["status"] == "STOPPED"
        assert invoke(bot, "mission", "init", state, expected=1)["status"] == "ALREADY_INITIALIZED"
        report = {
            "version": "br16a-walkthrough-result/v1",
            "scope": "offline fixture chain only; no provider call, live executor, business outcome, or production authority",
            "checks": {
                "m00_to_m11_shared_lineage": "PASS",
                "history_replay_after_restore": "MATCH",
                "m10_rebind_no_mutation": "PASS",
                "m11_unknown_requires_stop": "PASS",
                "restart_stop_blocks_canary": "PASS",
                "recovery_admission_execution_permitted": False,
            },
            "paths": {
                "bot": str(bot),
                "runtime": str(state),
                "restored_runtime": str(restored),
                "backup": str(backup),
                "m08_intent": str(intent),
                "m08_policy": str(policy),
                "m09_approval": str(approval),
                "m10_grant": str(grant),
                "m10_cost_bound": str(cost),
                "m10_authorization": str(authorization),
                "m11_production_lease": str(production_lease_path),
                "m11_outcome": str(production_outcome),
                "m11_evaluation": "br16-production-e",
                "m11_cycle": "br16-production-cycle",
            },
        }
        (work / "walkthrough-result.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    suffix = f" workspace={Path(directory).resolve()}" if args.workspace else ""
    print("BR-16a PASS: one shared runtime, M00→M11 artifacts, UNKNOWN reconciliation, backup/restore, restart/replay and durable STOP" + suffix)


if __name__ == "__main__":
    main()
