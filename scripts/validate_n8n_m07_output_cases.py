"""Exercise the output contract through the learner Bot CLI, not a duplicate parser."""
import json
import os
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
BOT_DIR = ROOT / "lab/affiliate-bot"
REGISTRY = ROOT / "lab/mission-runtime/testdata/m07-registry.json"


def main():
    go = shutil.which(os.environ.get("GO_BIN", "go")) or os.environ.get("GO_BIN", "go")
    with tempfile.TemporaryDirectory(prefix="m07-output-") as directory:
        work = Path(directory); bot = work / "bot"; history = work / "history.jsonl"
        env = dict(os.environ, GOWORK="off", GOCACHE=str(work / "go-cache"))
        subprocess.run([go, "build", "-o", str(bot), "./cmd/bot"], cwd=BOT_DIR, check=True, env=env)
        subprocess.run([str(bot), "history", "capture", str(history), str(BOT_DIR / "data/m02-sample-observations.json"), "m07-d", "2026-09-03T00:00:00Z", "2026-09-03T00:00:00Z"], check=True, capture_output=True, text=True, env=env)
        context = json.loads(subprocess.run([str(bot), "m07", "context", history, "m07-d"], check=True, capture_output=True, text=True, env=env).stdout)["artifact"]
        evidence_id = context["evidence_ids"][0]
        evidence = next(item for item in context["evidence"] if item["evidence_id"] == evidence_id)
        cases = {
            "normal": ({"state": "HUMAN_REVIEW", "answer": "limited", "claims": [{"text": "supported", "field_or_claim": evidence["field_or_claim"], "value": evidence.get("value"), "evidence_ids": [evidence_id]}], "evidence_ids": [evidence_id], "tool_calls": [], "authority": "A2-RO", "write_permission": False}, 0, "VALID"),
            "forged-id": ({"state": "HUMAN_REVIEW", "answer": "forged", "claims": [{"text": "forged", "evidence_ids": ["e999"]}], "evidence_ids": ["e999"], "tool_calls": [], "authority": "A2-RO", "write_permission": False}, 1, "ABSTAIN"),
            "missing-claims": ({"state": "HUMAN_REVIEW", "answer": "unsupported", "claims": [], "evidence_ids": [], "tool_calls": [], "authority": "A2-RO", "write_permission": False}, 1, "ABSTAIN"),
            "write-request": ({"state": "HUMAN_REVIEW", "answer": "write", "claims": [{"text": "supported", "field_or_claim": evidence["field_or_claim"], "value": evidence.get("value"), "evidence_ids": [evidence_id]}], "evidence_ids": [evidence_id], "tool_calls": [{"tool_name": "public_http", "method": "POST", "target": "https://example.com/a"}], "authority": "A2-RO", "write_permission": False}, 1, "ABSTAIN"),
            "authority-escalation": ({"state": "PROPOSE", "answer": "escalated", "claims": [{"text": "supported", "field_or_claim": evidence["field_or_claim"], "value": evidence.get("value"), "evidence_ids": [evidence_id]}], "evidence_ids": [evidence_id], "tool_calls": [], "authority": "A3", "write_permission": False}, 1, "ABSTAIN"),
        }
        model = work / "model.json"
        for name, (value, expected_code, expected_status) in cases.items():
            model.write_text(json.dumps(value), encoding="utf-8")
            result = subprocess.run([str(bot), "m07", "validate", history, "m07-d", model, REGISTRY], capture_output=True, text=True, env=env)
            assert result.returncode == expected_code, (name, result.stdout, result.stderr)
            assert json.loads(result.stdout)["status"] == expected_status, name
    print("N8N M07 OUTPUT EXECUTION PASS: real implementation accepts only grounded, read-only output")


if __name__ == "__main__":
    main()
