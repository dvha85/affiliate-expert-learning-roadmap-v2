"""Offline M00->M05 continuity plus an isolated, content-verified rollback lab."""
import difflib
import json
import os
import re
import shutil
from pathlib import Path
import subprocess
import tempfile
from smoke_br08 import ROOT, GO, run, ENV
from validate_learner_walkthroughs import load_rehearsal_inputs


def chain(work):
    samples = load_rehearsal_inputs(ROOT)
    bot = work / "bot"
    run([GO, "build", "-o", bot, "./cmd/bot"], ROOT / "lab/affiliate-bot")
    h, a, o, e, p, r = [work / name for name in ["history", "actions", "outcomes", "evaluations", "proposals", "reviews"]]
    source = work / "input.json"

    def command(args, want="VALID", code=0):
        result = json.loads(run([bot, *args], code=code).stdout)
        assert result["status"] == want and result["execution_permitted"] is False, result
        if args[0] in ("proposal", "review"):
            assert result["command"] == " ".join(args[:2]) and result["auto_apply"] is False
        if code:
            assert "artifact" not in result
        return result.get("artifact")

    def write(value):
        source.write_text(json.dumps(value), encoding="utf-8")

    # Walk one artifact line from both M00 snapshots through M01/M02. Reuse this
    # exact history for the M03-M05 CLI below so the integration smoke proves
    # continuity instead of joining separate demo workspaces after the fact.
    for n, expected in ((1, "GET_MORE_DATA"), (2, "RANK_SCENARIO")):
        observations = command(["evidence", "import", ROOT / f"examples/m00-import/packet-t{n}.json"])
        write(observations)
        assert expected in run([bot, source]).stdout
        record_id = f"d1{n}"
        capture = [bot, "history", "capture", h, source, record_id, f"2026-09-0{n}T01:00:00Z", f"2026-09-0{n}T02:00:00Z"]
        assert "APPENDED" in run(capture).stdout
        history_before_retry = h.read_bytes()
        assert "EXACT_DUPLICATE" in run(capture).stdout
        assert h.read_bytes() == history_before_retry
        packet = json.loads(run([bot, "history", "decision", h, record_id, ROOT / "lab/affiliate-bot/data/m02-decision-context.json"]).stdout)
        assert packet["decision_id"] == record_id and packet["state"] == expected and packet["action"] is None
        assert packet["evidence_ids"] == [f"br09-product-a-t{n}"]
    assert "d11" in run([bot, "history", "list", h]).stdout
    assert "d12" in run([bot, "history", "list", h]).stdout
    assert run([bot, "history", "replay", h]).stdout.count("replay=MATCH") == 2
    write(samples["action.json"])
    command(["action","record",h,a,source],"APPENDED")
    assert command(["action", "list", h, a]) == [{"action_id":"a12","decision_id":"d12","action_type":"synthetic","target":"fixture:none","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True}]
    write(samples["outcome.json"])
    command(["outcome","import",h,a,o,source],"APPENDED")
    assert command(["outcome", "list", h, a, o]) == [{"outcome_id":"o12","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a12"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:br12d"}]
    write(samples["advisor-config.json"])
    advisor = command(["advisor","mock",h,a,o,source],"SUPPORTED")
    assert advisor["advisor_output"]["state"] == "HUMAN_REVIEW"
    assert set(advisor["advisor_output"]["evidence_ids"]) == {"d12","br09-product-a-t2","br09-commission_rate-t2","br09-price-t2","a12","o12"}, advisor["advisor_output"]["evidence_ids"]
    assert advisor["advisor_output"]["write_tool_requested"] is False
    write(samples["evaluation-config.json"])
    evaluation = command(["evaluation","create",h,a,o,e,source],"APPENDED")
    assert evaluation["result"] == "INCONCLUSIVE" and evaluation["outcome_ids"] == ["o12"] and evaluation["evidence_ids"] == ["o12"]
    write(samples["proposal.json"])
    proposal = command(["proposal","import",h,a,o,e,p,source],"APPENDED")
    # Schema fixture ONLY: not an actual signed/operated human approval. Actual
    # implementation review lives in the GitHub PR, not this synthetic record.
    write(samples["review.json"])
    review = command(["review","import",h,a,o,e,p,r,source],"APPENDED")
    stores = [h,a,o,e,p,r]
    before = [path.read_bytes() for path in stores]
    assert command(["evaluation","list",h,a,o,e]) == [evaluation]
    assert command(["proposal","list",h,a,o,e,p]) == [proposal]
    assert command(["review","list",h,a,o,e,p,r]) == [review]
    assert "replay=MATCH" in run([bot,"history","replay",h]).stdout
    command(["review","import",h,a,o,e,p,r,source],"EXACT_DUPLICATE")
    bad = work / "empty-evaluations"; bad.write_text("")
    command(["review","list",h,a,o,bad,p,r],"PROPOSAL_STORE_ERROR",1)
    assert before == [path.read_bytes() for path in stores]
    print("M00->M05 PASS: t1 GET_MORE_DATA + t2 RANK_SCENARIO share one history; M03-M05 resolve d12/a12/o12; fresh-process list/replay; INCONCLUSIVE; synthetic review; no execution")


def rollback(work):
    relative = "lab/affiliate-bot/cmd/bot/improvement_store.go"
    test = "lab/affiliate-bot/cmd/bot/improvement_command_test.go"
    paths = subprocess.check_output(["git","ls-files","-z","core","contracts","lab/affiliate-bot"],cwd=ROOT).decode().split("\0")
    # Only tracked source/schema/module files, plus this PR's known regression.
    for name in sorted(set(paths + [test])):
        if not name or not name.endswith((".go", ".json", "go.mod", "go.sum")):
            continue
        source = ROOT / name
        assert not source.is_symlink()
        target = work / name; target.parent.mkdir(parents=True,exist_ok=True)
        target.write_bytes(source.read_bytes())
    target = work / relative
    current = target.read_text()
    block = '\tcommand := kind\n\tif len(args) > 0 && (args[0] == "import" || args[0] == "list") {\n\t\tcommand += " " + args[0]\n\t}\n'
    assert current.count(block) == 1 and current.count('"command": command,') == 1
    previous = current.replace(block, "", 1).replace('"command": command,', '"command": kind,', 1)
    # The rollback lab is defined by the targeted behavior above, not a
    # repository-wide blob SHA. Unrelated safe edits to this source file must
    # not make a valid before/after/rollback regression fail in CI.
    assert previous != current, "rollback fixture did not produce a distinct baseline"
    print("".join(difflib.unified_diff(previous.splitlines(True),current.splitlines(True),fromfile="command-label/v1",tofile="command-label/v2")),end="")
    env = dict(ENV); env.pop("DEEPSEEK_API_KEY",None)
    def regression(source, expected):
        target.write_text(source)
        result = subprocess.run([GO,"test","./cmd/bot","-run","^TestImprovementCommandIdentifiesOperation$","-count=1"],cwd=work / "lab/affiliate-bot",env=env,capture_output=True,text=True,timeout=120)
        assert result.returncode == expected, result.stdout + result.stderr
        if expected:
            assert "operation label: got proposal, want proposal import" in result.stdout
        return result.returncode
    phases = {"before":regression(previous,1),"after":regression(current,0),"rollback":regression(previous,1),"restore":regression(current,0)}
    assert target.read_text() == current
    print(json.dumps({"regression_exit_codes":phases,"business_effectiveness":"NOT_MEASURED","human_review_fixture_only":True,"execution_permitted":False}))


def documented_walkthrough(directory, shell_kind):
    """Execute the published command blocks, with independent output assertions."""
    text = (ROOT / "curriculum/PRACTICE-M00-M05.md").read_text(encoding="utf-8")
    pattern = re.compile(r"<!-- run:" + shell_kind + r":([\w]+) -->\s*([`~]{3})\w+\n(.*?)\n\2", re.S)
    blocks = {m[1]: m[3] for m in pattern.finditer(text)}
    assert set(blocks) == {"setup", "import", "m01", "m02", "m03", "m04", "m05"}
    samples = load_rehearsal_inputs(ROOT)
    env = dict(ENV, TMP=str(directory), TEMP=str(directory), TMPDIR=str(directory))
    if shell_kind == "ps":
        setup = blocks["setup"]
        for name, sample in samples.items():
            value = json.dumps(sample, ensure_ascii=True).replace("'", "''")
            setup += f"\n[IO.File]::WriteAllText((Join-Path $work '{name}'), '{value}', [Text.UTF8Encoding]::new($false))"
        commands = [setup] + [blocks[phase] for phase in ("import", "m01", "m02", "m03", "m04", "m05")]
        # Instrument native exit codes: PowerShell does not throw on them.
        script = "\n".join(commands)
        script = re.sub(r"(?m)^(& \$bot[^\n]*)$", r'\1\nif ($LASTEXITCODE -ne 0) { throw "CLI failed" }', script)
        script += '\nWrite-Output ("WALKTHROUGH_DIR=" + $work)\n'
        # These are the same inline commands the learner pastes in a terminal;
        # do not change the host's script execution policy.
        prefix = [shutil.which("powershell") or "pwsh", "-NoProfile", "-Command", script]
    else:
        setup = blocks["setup"]
        for name, sample in samples.items():
            # JSON fixture strings are passed in single quoted arguments.
            value = json.dumps(sample, ensure_ascii=True).replace("'", "'\"'\"'")
            setup += f"\nprintf '%s' '{value}' > \"$work/{name}\""
        script = "\n".join([setup] + [blocks[phase] for phase in ("import", "m01", "m02", "m03", "m04", "m05")])
        script += '\nprintf "WALKTHROUGH_DIR=%s\\n" "$work"\n'
        path = directory / "walkthrough.sh"
        path.write_text(script, encoding="utf-8")
        prefix = ["bash", str(path)]
    result = subprocess.run(prefix, cwd=ROOT, env=env, text=True, capture_output=True, timeout=180, encoding="utf-8", errors="replace")
    assert result.returncode == 0, result.stdout + result.stderr
    match = re.search(r"(?m)^WALKTHROUGH_DIR=(.+)$", result.stdout)
    assert match, result.stdout
    work = Path(match[1].strip()).resolve()
    assert work.is_relative_to(directory.resolve())
    records = [json.loads(line) for line in (work / "history.jsonl").read_text(encoding="utf-8").splitlines()]
    assert [r["record_id"] for r in records] == ["d11", "d12"]
    assert [r["recorded_result"]["state"] for r in records] == ["GET_MORE_DATA", "RANK_SCENARIO"]
    for store, field, value in (("actions", "decision_id", "d12"), ("outcomes", "outcome_id", "o12"),
                                ("evaluations", "result", "INCONCLUSIVE"), ("proposals", "auto_apply", False),
                                ("reviews", "decision", "REQUEST_CHANGES")):
        items = [json.loads(line) for line in (work / f"{store}.jsonl").read_text(encoding="utf-8").splitlines()]
        assert len(items) == 1 and items[0][field] == value, store
    assert result.stdout.count("replay=MATCH") == 4
    assert "EXACT_DUPLICATE" in result.stdout and "HUMAN_REVIEW" in result.stdout
    print(f"DOCUMENTED WALKTHROUGH {shell_kind} PASS: exact published commands and six JSON inputs; one history; M00-M05 fresh processes")


def main():
    with tempfile.TemporaryDirectory(prefix="ca-doc-commands-") as directory:
        documented_walkthrough(Path(directory), "ps" if os.name == "nt" else "sh")
    with tempfile.TemporaryDirectory(prefix="br12d-chain-") as directory:
        chain(Path(directory))
    with tempfile.TemporaryDirectory(prefix="br12d-rollback-") as directory:
        rollback(Path(directory))
    print("BR-12d PASS: synthetic continuity and isolated FAIL/PASS/rollback; no business or human-pilot proof")


if __name__ == "__main__":
    main()
