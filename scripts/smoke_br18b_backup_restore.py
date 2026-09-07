"""Backup/restore smoke using artifacts created and replayed by the learner Bot."""
import json
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
        grant = root / "grant.json"; grant.write_text(json.dumps({"grant_id":"br18-g","max_executions":2,"max_cost_minor":10,"currency":"USD"}), encoding="utf-8")
        assert invoke(bot, "mission", "m10-canary", runtime, grant, env=env)["status"] == "ACK"
        assert invoke(bot, "mission", "m10-reserve", runtime, "4", env=env)["status"] == "RESERVED"
        invoke(bot, "mission", "m11-stop", runtime, "backup-drill", env=env)
        assert invoke(bot, "backup", "create", runtime, backup, env=env)["status"] == "BACKED_UP"
        (backup / "history.jsonl").write_text("tampered\n", encoding="utf-8")
        assert invoke(bot, "backup", "restore", backup, restored, expected=1, env=env)["status"] == "VERIFY_FAILED"
        # Recreate the backup from the unchanged runtime, then restore into a
        # fresh directory and let a new process validate the artifacts.
        invoke(bot, "backup", "create", runtime, backup, env=env)
        assert invoke(bot, "backup", "restore", backup, restored, env=env)["status"] == "RESTORED"
        assert "replay=MATCH" in run([bot, "history", "replay", restored / "history.jsonl"], env=env).stdout
        status = invoke(bot, "mission", "status", restored, env=env)
        restored_state = status["artifact"]
        assert restored_state["stop"] is True and restored_state["stop_reason"] == "backup-drill"
        assert restored_state["approval"]["approval_id"] == "br18-ap"
        assert restored_state["canary"]["executions_used"] == 1 and restored_state["canary"]["cost_used_minor"] == 4
        assert (restored / "actions.jsonl").exists()
        assert invoke(bot, "mission", "m10-reserve", restored, "1", expected=1, env=env)["status"] == "STOPPED"
        invalid_state = json.loads((runtime / "mission-state.json").read_text(encoding="utf-8")); invalid_state["canary"]["executions_used"] = -1
        (runtime / "mission-state.json").write_text(json.dumps(invalid_state), encoding="utf-8")
        invalid_backup = root / "invalid-backup"; invalid_restored = root / "invalid-restored"
        invoke(bot, "backup", "create", runtime, invalid_backup, env=env)
        assert invoke(bot, "backup", "restore", invalid_backup, invalid_restored, expected=1, env=env)["status"] == "STATE_FAILED"
    print("BR-18b PASS: runtime-created history/state/STOP backed up with manifest, tamper rejected, fresh-process replay and durable STOP verified")


if __name__ == "__main__":
    main()
