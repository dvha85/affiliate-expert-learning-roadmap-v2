"""Offline M00->M05 continuity plus an isolated, content-verified rollback lab."""
import difflib
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
from smoke_br08 import ROOT, GO, run, ENV


def chain(work):
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

    observations = command(["evidence", "import", ROOT / "examples/m00-import/packet-t2.json"])
    write(observations)
    assert "APPENDED" in run([bot, "history", "capture", h, source, "d12", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"]).stdout
    packet = json.loads(run([bot,"history","decision",h,"d12",ROOT / "lab/affiliate-bot/data/m02-decision-context.json"]).stdout)
    assert packet["decision_id"] == "d12" and packet["evidence_ids"] == ["br09-product-a-t2"] and packet["action"] is None
    write({"action_id":"a12","decision_id":"d12","action_type":"synthetic","target":"fixture:none","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":True})
    command(["action","record",h,a,source],"APPENDED")
    write({"outcome_id":"o12","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a12"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:br12d"})
    command(["outcome","import",h,a,o,source],"APPENDED")
    write({"decision_id":"d12","question":"Thiếu bằng chứng gì?","as_of":"2026-09-06T00:00:00Z","max_age_hours":8760})
    advisor = command(["advisor","mock",h,a,o,source],"SUPPORTED")
    assert advisor["advisor_output"]["state"] == "HUMAN_REVIEW"
    assert set(advisor["advisor_output"]["evidence_ids"]) == {"d12","br09-product-a-t2","br09-commission_rate-t2","br09-price-t2","a12","o12"}, advisor["advisor_output"]["evidence_ids"]
    assert advisor["advisor_output"]["write_tool_requested"] is False
    write({"evaluation_id":"e12","decision_id":"d12","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a12"},"outcome_ids":["o12"],"evaluated_at":"2026-09-06T00:00:00Z"})
    evaluation = command(["evaluation","create",h,a,o,e,source],"APPENDED")
    assert evaluation["result"] == "INCONCLUSIVE" and evaluation["outcome_ids"] == ["o12"] and evaluation["evidence_ids"] == ["o12"]
    write({"proposal_id":"p12","evaluation_ids":["e12"],"current_version":"command-label/v1","proposed_version":"command-label/v2","change_summary":"Thêm import/list vào nhãn diagnostic","expected_benefit":"Dễ phân biệt thao tác trong log; chưa đo lợi ích","risks":["Consumer so khớp command cũ cần cập nhật"],"rollback":"Phục hồi improvement_store.go v1 trong bản sao tạm; không sửa store","auto_apply":False})
    proposal = command(["proposal","import",h,a,o,e,p,source],"APPENDED")
    # Schema fixture ONLY: not an actual signed/operated human approval. Actual
    # implementation review lives in the GitHub PR, not this synthetic record.
    write({"review_id":"r12","proposal_id":"p12","reviewed_by":"human","reviewed_at":"2026-09-07T00:00:00Z","decision":"REQUEST_CHANGES","reason":"SYNTHETIC fixture: yêu cầu test FAIL/PASS và rollback; không là review người thật"})
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
    print("M00->M05 PASS: exact IDs; fresh-process list/replay; INCONCLUSIVE; synthetic review; no execution")


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
    raw = previous.encode()
    blob = hashlib.sha1(b"blob " + str(len(raw)).encode() + b"\0" + raw).hexdigest()
    assert blob == "a3363b51a073a83a4b6006834eaec2d99930069d", "rollback no longer matches actual baseline file"
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
    print(json.dumps({"regression_exit_codes":phases,"rollback_blob":blob,"business_effectiveness":"NOT_MEASURED","human_review_fixture_only":True,"execution_permitted":False}))


def main():
    with tempfile.TemporaryDirectory(prefix="br12d-chain-") as directory:
        chain(Path(directory))
    with tempfile.TemporaryDirectory(prefix="br12d-rollback-") as directory:
        rollback(Path(directory))
    print("BR-12d PASS: synthetic continuity and isolated FAIL/PASS/rollback; no business or human-pilot proof")


if __name__ == "__main__":
    main()
