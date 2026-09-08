"""Run the learner Bot's ACCESSTRADE CSV adapter through its real CLI path."""
import json
from pathlib import Path
import tempfile

from smoke_br08 import GO, ROOT, run


def main():
    with tempfile.TemporaryDirectory(prefix="br10d-accesstrade-") as directory:
        work = Path(directory)
        bot = work / "bot"
        runtime = work / "runtime"
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        runtime.mkdir()
        history, actions, outcomes = [runtime / item for item in ["history.jsonl", "actions.jsonl", "outcomes.jsonl"]]
        source = work / "action.json"
        report = ROOT / "examples" / "accesstrade-report" / "sanitized-orders.csv"
        manifest = ROOT / "examples" / "accesstrade-report" / "manifest.json"
        assert json.loads(run([bot, "mission", "init", runtime]).stdout)["status"] == "INITIALIZED"
        run([bot, "history", "capture", history, ROOT / "lab/affiliate-bot/data/m02-sample-observations.json", "decision-1", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"])
        source.write_text(json.dumps({"action_id":"manual-action-001","decision_id":"decision-1","action_type":"manual_fixture","target":"fixture:no-publication","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True}))
        run([bot, "action", "record", history, actions, source])
        command = [bot, "outcome", "accesstrade-import", history, actions, outcomes, report, manifest]
        result = json.loads(run(command).stdout)
        assert result["status"] == "APPENDED" and result["execution_permitted"] is False
        assert json.loads(run(command).stdout)["status"] == "EXACT_DUPLICATE"
        receipt_list = json.loads(run([bot, "outcome", "accesstrade-receipts", history, actions, outcomes]).stdout)
        assert receipt_list["status"] == "VALID"
        assert receipt_list["artifact"][0]["snapshot_id"] == "at-sanitized-2026-09-08"
        assert receipt_list["artifact"][0]["report_sha256"].startswith("sha256:")
        listed = json.loads(run([bot, "outcome", "list", history, actions, outcomes]).stdout)["artifact"]
        assert [(entry["status"], entry["metrics"]) for entry in listed] == [
            ("PENDING", {"order_value_vnd": 250000, "reported_commission_vnd": 17500}),
            ("CANCELLED", {"order_value_vnd": 150000, "reported_commission_vnd": 0}),
        ]
        backup, restored = work / "backup", work / "restored"
        backup_result = json.loads(run([bot, "backup", "create", runtime, backup]).stdout)
        assert backup_result["status"] == "BACKED_UP"
        assert "accesstrade-receipts.jsonl" in backup_result["artifact"]["required"]
        assert json.loads(run([bot, "backup", "restore", backup, restored]).stdout)["status"] == "RESTORED"
        restored_receipts = json.loads(run([bot, "outcome", "accesstrade-receipts", restored / "history.jsonl", restored / "actions.jsonl", restored / "outcomes.jsonl"]).stdout)
        assert restored_receipts["status"] == "VALID"
        assert restored_receipts["artifact"] == receipt_list["artifact"]
    print("BR-10d PASS: ACCESSTRADE sanitized CSV mapping, receipt provenance, and backup/restore through learner CLI; no UTM attribution or execution")


if __name__ == "__main__":
    main()
