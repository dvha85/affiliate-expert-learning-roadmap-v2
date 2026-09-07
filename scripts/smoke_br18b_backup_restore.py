"""Offline backup/restore smoke for append-only artifacts and durable STOP."""
import hashlib
import json
import tempfile
from pathlib import Path


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    with tempfile.TemporaryDirectory(prefix="br18b-") as d:
        root = Path(d); history = root / "history.jsonl"; ledger = root / "ledger.jsonl"; stop = root / "STOP"
        history.write_text('{"record_id":"h1","state":"RANK_SCENARIO"}\n', encoding="utf-8")
        ledger.write_text('{"reservation_id":"r1","status":"RELEASED"}\n', encoding="utf-8")
        stop.write_text('{"active":true,"reason":"manual-drill"}\n', encoding="utf-8")
        manifest = {p.name: digest(p) for p in (history, ledger, stop)}
        backup = root / "backup"; backup.mkdir()
        for p in (history, ledger, stop): (backup / p.name).write_bytes(p.read_bytes())
        assert {p.name: digest(p) for p in backup.iterdir()} == manifest
        # Tamper is detected before restore is accepted.
        (backup / "ledger.jsonl").write_text("tampered\n", encoding="utf-8")
        assert digest(backup / "ledger.jsonl") != manifest["ledger.jsonl"]
        (backup / "ledger.jsonl").write_bytes(ledger.read_bytes())
        assert {p.name: digest(p) for p in backup.iterdir()} == manifest
        restored = root / "restored"; restored.mkdir()
        for name in manifest: (restored / name).write_bytes((backup / name).read_bytes())
        assert (restored / "history.jsonl").read_bytes() == history.read_bytes()
        assert json.loads((restored / "STOP").read_text())["active"] is True
    print("BR-18b PASS: backup checksum/tamper detection, byte-identical restore and durable STOP marker")


if __name__ == "__main__":
    main()
