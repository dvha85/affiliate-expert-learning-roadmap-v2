"""Synthetic offline advisor: grounded context, freshness and no-write boundary."""
import json
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run


def main():
    with tempfile.TemporaryDirectory(prefix="br11a-") as directory:
        work = Path(directory); bot = work / "bot"
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        history, actions, outcomes, action, config = [work / x for x in ["history", "actions", "outcomes", "action.json", "config.json"]]
        run([bot, "history", "capture", history, ROOT / "lab/affiliate-bot/data/m02-sample-observations.json", "d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"])
        action.write_text(json.dumps({"action_id":"a","decision_id":"d","action_type":"synthetic","target":"fixture:none","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True}))
        assert json.loads(run([bot,"action","record",history,actions,action]).stdout)["status"] == "APPENDED"
        outcome=work/"outcome.json"; outcome.write_text(json.dumps({"outcome_id":"o","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:advisor"}))
        assert json.loads(run([bot,"outcome","import",history,actions,outcomes,outcome]).stdout)["status"] == "APPENDED"
        config.write_text(json.dumps({"decision_id":"d","question":"Có nên xem xét offer này?","as_of":"2026-09-06T00:00:00Z","max_age_hours":8760}))
        result=json.loads(run([bot,"advisor","mock",history,actions,outcomes,config]).stdout)
        assert result["status"] == "SUPPORTED" and result["execution_permitted"] is False
        output=result["artifact"]["advisor_output"]; context=result["artifact"]["context"]
        assert output["state"] == "HUMAN_REVIEW" and output["write_tool_requested"] is False
        assert {"d","obs-a-1","obs-b-1","a","o"}.issubset(set(output["evidence_ids"])), output["evidence_ids"]
        assert context["version"] == "br11a-context/v1" and context["config"]["max_age_hours"] == 8760
        before=[p.read_bytes() for p in [history,actions,outcomes]]
        config.write_text(json.dumps({"decision_id":"d","question":"q","as_of":"2026-09-06T00:00:00Z","max_age_hours":1}))
        assert json.loads(run([bot,"advisor","mock",history,actions,outcomes,config],code=1).stdout)["status"] == "ABSTAIN_STALE"
        config.write_text(json.dumps({"decision_id":"d","question":"q","as_of":"2026-09-02T00:00:00Z","max_age_hours":8760}))
        assert json.loads(run([bot,"advisor","mock",history,actions,outcomes,config],code=1).stdout)["status"] == "ABSTAIN_FUTURE"
        config.write_text(json.dumps({"decision_id":"unknown","question":"q","as_of":"2026-09-06T00:00:00Z","max_age_hours":8760}))
        assert json.loads(run([bot,"advisor","mock",history,actions,outcomes,config],code=1).stdout)["status"] == "CONTEXT_ERROR"
        assert before == [p.read_bytes() for p in [history,actions,outcomes]]
    print("BR-11a PASS: mock grounded context; stale/future/orphan boundaries; no writes")


if __name__ == "__main__":
    main()
