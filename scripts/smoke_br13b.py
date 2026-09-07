"""Offline watcher fixture handoff into canonical history; no network fetch."""
import json
from pathlib import Path
import tempfile
from smoke_br08 import ROOT, GO, run

def main():
    with tempfile.TemporaryDirectory(prefix="br13b-") as directory:
        work=Path(directory);bot=work/"bot";history=work/"history.jsonl"
        run([GO,"build","-o",bot,"./cmd/bot"],ROOT/"lab/affiliate-bot")
        def ingest(name,status,code=0,sink=history):
            result=json.loads(run([bot,"watcher","fixture-import",sink,ROOT/"examples/watcher"/name],code=code).stdout)
            assert result["status"]==status and result["persisted"]==(code==0)
            assert result["network_fetch_performed"] is False and result["execution_permitted"] is False
            if code: assert "artifact" not in result
            return result
        first=ingest("offer-valid.json","APPENDED")
        before=history.read_bytes()
        again=ingest("offer-valid.json","EXACT_DUPLICATE")
        assert first["artifact"]==again["artifact"] and history.read_bytes()==before
        assert first["artifact"]["state"]=="RANK_SCENARIO"
        missing=ingest("offer-missing.json","APPENDED")
        assert missing["artifact"]["state"]=="GET_MORE_DATA"
        assert missing["artifact"]["record_id"]!=first["artifact"]["record_id"]
        before=history.read_bytes()
        ingest("offer-error.json","FIXTURE_ERROR",1)
        ingest("offer-valid.json","HANDOFF_ERROR",1,work/"missing-parent"/"history")
        assert history.read_bytes()==before
        listing=run([bot,"history","list",history]).stdout
        assert first["artifact"]["record_id"] in listing and missing["artifact"]["record_id"] in listing
        assert run([bot,"history","replay",history]).stdout.count("replay=MATCH")==2
    print("BR-13b PASS: fixture -> canonical history; restart/replay; retry; missing commission; failed sink not persisted; no network")

if __name__=="__main__": main()
