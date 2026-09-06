"""Synthetic decision -> action -> outcome chain in separate processes."""
import json
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run


def main():
    with tempfile.TemporaryDirectory(prefix="br10b-") as directory:
        work = Path(directory)
        bot = work / "bot"
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        history, actions, outcomes, source = [work / name for name in ["history", "actions", "outcomes", "input.json"]]
        run([bot, "history", "capture", history, ROOT / "lab/affiliate-bot/data/m02-sample-observations.json", "d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"])
        action = {"action_id": "a", "decision_id": "d", "action_type": "synthetic_manual", "target": "fixture:no-publication", "performed_by": "human", "performed_at": "2026-09-04T00:00:00Z", "measurement_window_end": "2026-09-05T00:00:00Z", "compliance_reviewed": True}
        source.write_text(json.dumps(action))
        run([bot, "action", "record", history, actions, source])
        upstream = history.read_bytes(), actions.read_bytes()
        base = {"outcome_id": "pending", "effect_ref": {"effect_kind": "HUMAN_ACTION", "effect_id": "a"}, "observed_at": "2026-09-04T01:00:00Z", "status": "PENDING", "metrics": {}, "source_ref": "fixture:synthetic-report"}
        args = [bot, "outcome", "import", history, actions, outcomes, source]
        source.write_text(json.dumps(base))
        assert json.loads(run(args).stdout)["status"] == "APPENDED"
        assert json.loads(run(args).stdout)["status"] == "EXACT_DUPLICATE"
        before = outcomes.read_bytes()
        for changes, expected in [
            ({"effect_ref": {"effect_kind": "HUMAN_ACTION", "effect_id": "absent"}}, "ORPHAN_ACTION"),
            ({"effect_ref": {"effect_kind": "MACHINE_EXECUTION", "effect_id": "a"}}, "REJECT_MACHINE_EXECUTION"),
            ({"observed_at": "2026-09-03T00:00:00Z"}, "OUTCOME_BEFORE_ACTION"),
            ({"status": "NO_OBSERVED_OUTCOME"}, "MEASUREMENT_WINDOW_OPEN"),
            ({"metrics": {"clicks": 0}}, "CONFLICT"),
        ]:
            source.write_text(json.dumps(dict(base, **changes)))
            assert json.loads(run(args, code=1).stdout)["status"] == expected
            assert outcomes.read_bytes() == before
        zero = dict(base, outcome_id="zero", status="NO_OBSERVED_OUTCOME", observed_at="2026-09-05T07:00:00+07:00", metrics={"clicks": 0})
        source.write_text(json.dumps(zero))
        assert json.loads(run(args).stdout)["status"] == "APPENDED"
        late = dict(base, outcome_id="late", status="PAID", observed_at="2026-09-06T00:00:00Z", metrics={"commission": 8})
        source.write_text(json.dumps(late))
        assert json.loads(run(args).stdout)["status"] == "APPENDED"
        result = json.loads(run([bot, "outcome", "list", history, actions, outcomes]).stdout)
        assert result["artifact"] == [base, zero, late] and result["execution_permitted"] is False
        assert upstream == (history.read_bytes(), actions.read_bytes())
    print("BR-10b PASS: linked outcomes; pending/zero/late; duplicate/conflict/orphan/window; restart; synthetic only")


if __name__ == "__main__":
    main()
