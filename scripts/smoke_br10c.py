"""Acceptance audit: M00 through persisted action/outcomes in a fresh lab."""
import json
import os
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run


def main():
    with tempfile.TemporaryDirectory(prefix="br10c-") as directory:
        work = Path(directory)
        bot = work / "bot"
        run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
        history, actions, outcomes = [work / name for name in ["history", "actions", "outcomes"]]
        source, observations = work / "input.json", work / "observations.json"

        def command(args, expected="VALID", code=0):
            result = run([bot, *args], code=code)
            envelope = json.loads(result.stdout)
            assert envelope["status"] == expected and envelope["execution_permitted"] is False
            if code:
                assert "artifact" not in envelope and result.stderr
            return envelope.get("artifact")

        imported = command(["evidence", "import", ROOT / "examples/m00-import/packet-t2.json"])
        observations.write_text(json.dumps(imported))
        capture = [bot, "history", "capture", history, observations, "audit-d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"]
        assert "APPENDED" in run(capture).stdout
        packet = json.loads(run([bot, "history", "decision", history, "audit-d", ROOT / "lab/affiliate-bot/data/m02-decision-context.json"]).stdout)
        assert packet["decision_id"] == "audit-d" and packet["evidence_ids"] == ["br09-product-a-t2"] and packet["action"] is None
        action = {"action_id": "audit-a", "decision_id": "audit-d", "action_type": "synthetic_manual", "target": "fixture:never-published", "performed_by": "human", "performed_at": "2026-09-04T00:00:00Z", "measurement_window_end": "2026-09-05T00:00:00Z", "compliance_reviewed": True}
        source.write_text(json.dumps(action))
        command(["action", "record", history, actions, source], "APPENDED")
        command(["action", "record", history, actions, source], "EXACT_DUPLICATE")
        expected = []
        for oid, status, when, metrics in [
            ("p", "PENDING", "2026-09-04T01:00:00Z", {}),
            ("z", "NO_OBSERVED_OUTCOME", "2026-09-05T07:00:00+07:00", {"clicks": 0}),
            ("l", "PAID", "2026-09-06T00:00:00Z", {"commission": 8}),
        ]:
            outcome = {"outcome_id": oid, "effect_ref": {"effect_kind": "HUMAN_ACTION", "effect_id": "audit-a"}, "observed_at": when, "status": status, "metrics": metrics, "source_ref": "fixture:br10c"}
            source.write_text(json.dumps(outcome))
            command(["outcome", "import", history, actions, outcomes, source], "APPENDED")
            command(["outcome", "import", history, actions, outcomes, source], "EXACT_DUPLICATE")
            expected.append(outcome)
        before = [p.read_bytes() for p in [history, actions, outcomes]]
        assert command(["action", "list", history, actions]) == [action]
        assert command(["outcome", "list", history, actions, outcomes]) == expected
        assert "audit-d" in run([bot, "history", "list", history]).stdout
        assert "replay=MATCH" in run([bot, "history", "replay", history]).stdout
        print("M00 -> decision audit-d -> action audit-a -> outcomes p/z/l: exact IDs, restart/list/replay PASS")

        # Mutations only in disposable copies, never the canonical lab files.
        bad_history, bad_actions, bad_outcomes = [work / name for name in ["bad-history", "bad-actions", "bad-outcomes"]]
        bad_history.write_text("")
        command(["outcome", "list", bad_history, actions, outcomes], "ACTION_STORE_ERROR", 1)
        bad_actions.write_text("")
        command(["outcome", "list", history, bad_actions, outcomes], "STORE_ERROR", 1)
        bad_outcomes.write_bytes(before[2] + before[2].splitlines(keepends=True)[0])
        command(["outcome", "list", history, actions, bad_outcomes], "STORE_ERROR", 1)
        bad_outcomes.write_text("corrupt\n")
        command(["outcome", "import", history, actions, bad_outcomes, source], "STORE_ERROR", 1)
        assert bad_outcomes.read_text() == "corrupt\n"
        alias = work / "alias"
        os.link(actions, alias)
        command(["outcome", "import", history, actions, alias, source], "PATH_ERROR", 1)
        print("Broken upstream, duplicate/corrupt outcome store and hardlink alias: fail closed PASS")

        # Exact shared 1 MiB payload limit through actual binary/store reader.
        limit = 1 << 20
        for size in [limit - 1, limit, limit + 1]:
            large = dict(expected[0], outcome_id="size", source_ref="x")
            compact = lambda obj: json.dumps(obj, separators=(",", ":"), ensure_ascii=False)
            large["source_ref"] = "x" * (size - len(compact(large).encode()) + 1)
            source.write_text(compact(large))
            path = work / f"size-{size}"
            command(["outcome", "import", history, actions, path, source], "APPENDED" if size <= limit else "STORE_ERROR", 0 if size <= limit else 1)
            if size <= limit:
                assert len(path.read_bytes()) == size + 1
                assert command(["outcome", "list", history, actions, path]) == [large]
            else:
                assert not path.exists()
        assert before == [p.read_bytes() for p in [history, actions, outcomes]]
        print("Outcome payload limit -1/exact/+1, rejected write isolation: PASS")
    print("BR-10c AUDIT PASS: synthetic lab only; no platform/live proof")


if __name__ == "__main__":
    main()
