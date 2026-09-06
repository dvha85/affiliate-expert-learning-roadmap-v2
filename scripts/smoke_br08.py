"""BR-08 integration smoke; all generated artifacts stay in a temporary workspace."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
ENV = dict(os.environ, GOWORK="off")
GO = os.environ.get("GO_BIN", "go")


def run(args, cwd=ROOT, code=0):
    result = subprocess.run([str(x) for x in args], cwd=cwd, env=ENV, capture_output=True, text=True)
    if result.returncode != code:
        raise RuntimeError(f"{args}: exit {result.returncode}, expected {code}: {result.stderr}")
    return result


def main():
    with tempfile.TemporaryDirectory(prefix="br08-smoke-") as directory:
        work = Path(directory)
        suffix = ".exe" if os.name == "nt" else ""
        bot, harness = work / ("bot" + suffix), work / ("harness" + suffix)
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        run([GO, "build", "-o", harness, "./cmd/demo"], ROOT / "lab/mission-runtime")
        action = work / "action.json"
        outcome = work / "outcome.json"
        action.write_bytes((ROOT / "lab/mission-runtime/testdata/m03-action.json").read_bytes())
        original = json.loads((ROOT / "lab/mission-runtime/testdata/m03-outcome.json").read_text())
        for name, changes, expected in [
            ("valid", {}, "VALID"),
            ("window", {"observed_at": "2026-09-03T11:00:00+07:00"}, "MEASUREMENT_WINDOW_OPEN"),
            ("link", {"effect_ref": {"effect_kind": "HUMAN_ACTION", "effect_id": "other"}}, "BROKEN_LINK"),
            ("schema", {"unexpected": True}, "INVALID_SCHEMA"),
        ]:
            outcome.write_text(json.dumps(dict(original, **changes)))
            before = (action.read_bytes(), outcome.read_bytes(), sorted(p.name for p in work.iterdir()))
            code = 0 if expected == "VALID" else 1
            learner = run([bot, "action", "validate", action, outcome], code=code)
            oracle = run([harness, "m03-check", action, outcome], code=code)
            envelope = json.loads(learner.stdout)
            assert envelope["status"] == expected and envelope["execution_permitted"] is False
            if code == 0:
                assert json.loads(oracle.stdout)["result"] == "VALID"
                assert envelope["artifact"]["action"]["action_id"] == "syn-a"
                assert not learner.stderr and not oracle.stderr
            else:
                assert "artifact" not in envelope and not oracle.stdout and expected in oracle.stderr
            assert before == (action.read_bytes(), outcome.read_bytes(), sorted(p.name for p in work.iterdir()))
            print(f"M03 {name}: {expected} PASS")
        assert json.loads(run([bot, "action"], code=2).stdout)["status"] == "USAGE_ERROR"
        history = work / "history.jsonl"
        capture = [bot, "history", "capture", history, ROOT / "lab/affiliate-bot/data/sample-observations.json", "br08-history", "2026-09-02T00:00:00Z", "2026-09-02T00:00:00Z"]
        assert "APPENDED" in run(capture).stdout
        before = history.read_bytes()
        assert "EXACT_DUPLICATE" in run(capture).stdout
        assert "replay=MATCH" in run([bot, "history", "replay", history]).stdout
        assert "br08-history" in run([bot, "history", "list", history]).stdout
        assert history.read_bytes() == before
        print("M02 capture/duplicate/list/replay: PASS; no history rewrite")
    print("BR-08 INTEGRATION SMOKE PASS (synthetic/local only)")


if __name__ == "__main__":
    main()
