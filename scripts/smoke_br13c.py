"""Pinned HTTPS fixture fetch -> parser -> canonical history smoke."""
import json
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run

def main():
    with tempfile.TemporaryDirectory(prefix="br13c-") as directory:
        work=Path(directory); bot=work/"bot"; history=work/"history.jsonl"
        run([GO,"build","-o",bot,"./cmd/bot"],ROOT/"lab/affiliate-bot")
        def invoke(expected,code=0):
            result=json.loads(run([bot,"watcher","fetch-fixture",history],code=code).stdout)
            assert result["status"]==expected, result
            assert result["execution_permitted"] is False
            assert result["network_fetch_attempted"] is (code==0)
            assert result["persisted"] is (code==0)
            if code: assert "artifact" not in result
            return result
        first=invoke("APPENDED")
        second=invoke("EXACT_DUPLICATE")
        assert first["artifact"]["response_sha256"]=="f70e18b84cf036e4973e8545f61f7ba0a111b8d69ecd25858ea60ceac877817b"
        assert first["artifact"]["source_url"].startswith("https://raw.githubusercontent.com/")
        assert first["artifact"]["evidence_kind"]=="synthetic"
        assert first["artifact"]["record_id"]==second["artifact"]["record_id"]
        assert len(history.read_bytes())>0
        assert "replay=MATCH" in run([bot,"history","replay",history]).stdout
    print("BR-13c PASS: pinned HTTPS fixture fetch; digest/source provenance; retry/replay; synthetic only")

if __name__=="__main__": main()
