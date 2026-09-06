"""Synthetic decision -> persisted human action, separate CLI processes."""
import json
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run


def main():
    with tempfile.TemporaryDirectory(prefix="br10a-") as directory:
        work = Path(directory)
        bot = work / "bot"
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        history, actions, source = work / "history.jsonl", work / "actions.jsonl", work / "action.json"
        run([bot, "history", "capture", history, ROOT / "lab/affiliate-bot/data/m02-sample-observations.json", "br10a-decision", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"])
        action = {"action_id": "br10a-action", "decision_id": "br10a-decision", "action_type": "synthetic_manual_post", "target": "fixture:br10a-no-publication", "performed_by": "human", "performed_at": "2026-09-04T00:00:00Z", "measurement_window_end": "2026-09-05T00:00:00Z", "compliance_reviewed": True}
        source.write_text(json.dumps(action))
        before = history.read_bytes(), source.read_bytes()
        args = [bot, "action", "record", history, actions, source]
        assert json.loads(run(args).stdout)["status"] == "APPENDED"
        saved = actions.read_bytes()
        assert json.loads(run(args).stdout)["status"] == "EXACT_DUPLICATE"
        listed = json.loads(run([bot, "action", "list", history, actions]).stdout)
        assert listed["artifact"] == [action] and listed["execution_permitted"] is False
        assert before == (history.read_bytes(), source.read_bytes())
        action["decision_id"] = "missing"
        source.write_text(json.dumps(action))
        assert json.loads(run(args, code=1).stdout)["status"] == "DECISION_ERROR"
        assert saved == actions.read_bytes()
        assert "replay=MATCH" in run([bot, "history", "replay", history]).stdout
    print("BR-10a PASS: synthetic record/duplicate/restart/list/orphan; no execution")


if __name__ == "__main__":
    main()
