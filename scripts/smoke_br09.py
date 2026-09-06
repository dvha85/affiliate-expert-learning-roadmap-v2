"""M00 packet -> M01/M02 -> decision; synthetic, temporary, separate processes."""
import json
import os
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run


def main():
    with tempfile.TemporaryDirectory(prefix="br09-smoke-") as directory:
        work = Path(directory)
        bot = work / ("bot.exe" if os.name == "nt" else "bot")
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        history = work / "history.jsonl"
        inputs = []
        for n, expected in [(1, "GET_MORE_DATA"), (2, "RANK_SCENARIO")]:
            source = ROOT / f"examples/m00-import/packet-t{n}.json"
            original = source.read_bytes()
            converted = json.loads(run([bot, "evidence", "import", source]).stdout)
            assert converted["status"] == "VALID" and converted["execution_permitted"] is False
            assert source.read_bytes() == original
            observations = work / f"input-t{n}.json"
            observations.write_text(json.dumps(converted["artifact"]))
            inputs.append(converted["artifact"])
            assert expected in run([bot, observations]).stdout
            capture = [bot, "history", "capture", history, observations, f"br09-t{n}", f"2026-09-0{n}T01:00:00Z", f"2026-09-0{n}T02:00:00Z"]
            assert "APPENDED" in run(capture).stdout
            before = history.read_bytes()
            assert "EXACT_DUPLICATE" in run(capture).stdout
            assert history.read_bytes() == before
            packet = json.loads(run([bot, "history", "decision", history, f"br09-t{n}", ROOT / "lab/affiliate-bot/data/m02-decision-context.json"]).stdout)
            assert packet["decision_id"] == f"br09-t{n}" and packet["state"] == expected and packet["action"] is None
            assert packet["evidence_ids"] == [f"br09-product-a-t{n}"]
            print(f"t{n}: import VALID; M01/M02 {expected}; duplicate EXACT_DUPLICATE; decision IDs resolved")
        before = history.read_bytes()
        assert run([bot, "history", "replay", history]).stdout.count("replay=MATCH") == 2
        listing = run([bot, "history", "list", history]).stdout
        assert "br09-t1" in listing and "br09-t2" in listing
        # Changed source content under an old field ID, with fresh aggregate and
        # record IDs: must not evade conflict detection.
        source = json.loads((ROOT / "examples/m00-import/packet-t2.json").read_text())
        source["products"][0]["observation_id"] = "fresh-aggregate"
        source["products"][0]["fields"][0]["value"] = 101
        changed = work / "changed.json"
        changed.write_text(json.dumps(source))
        converted = json.loads(run([bot, "evidence", "import", changed]).stdout)
        changed.write_text(json.dumps(converted["artifact"]))
        capture = [bot, "history", "capture", history, changed, "fresh-record", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"]
        assert "CONFLICT" in run(capture, code=1).stderr
        # Projection tamper must fail before persistence.
        converted["artifact"][0]["commission_rate"] = 0.9
        changed.write_text(json.dumps(converted["artifact"]))
        assert "projection/provenance mismatch" in run(capture, code=1).stderr
        assert history.read_bytes() == before
        records = [json.loads(line) for line in history.read_text().splitlines()]
        assert records[0]["observations"][0]["commission_rate"] is None
        provenance = json.loads(records[0]["observations"][0]["transformation_or_method"])
        assert provenance["product"]["fields"][0]["state"] == "pending"
        assert all(record["recorded_result"]["decision_id"] == record["record_id"] for record in records)
        print("restart/list/replay: 2 MATCH; source-ID conflict and projection tamper rejected; history unchanged")
    print("BR-09 SMOKE PASS (synthetic; no live proof)")


if __name__ == "__main__":
    main()
