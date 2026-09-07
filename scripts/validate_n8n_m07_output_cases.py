"""Offline parser contract for M07 Agent output; no model/network calls."""
import json


def parse(raw, allowed):
    if not isinstance(raw, dict) or raw.get("state") != "HUMAN_REVIEW":
        return "ABSTAIN"
    ids = raw.get("evidence_ids")
    if not isinstance(ids, list) or not ids or any(i not in allowed for i in ids):
        return "ABSTAIN"
    if raw.get("write_permission") is not False or raw.get("authority") != "A2-RO":
        return "ABSTAIN"
    return "ACCEPT"


def main():
    allowed = {"decision-1", "observation-1"}
    cases = {
        "normal": ({"state": "HUMAN_REVIEW", "evidence_ids": ["observation-1"], "authority": "A2-RO", "write_permission": False}, "ACCEPT"),
        "forged-id": ({"state": "HUMAN_REVIEW", "evidence_ids": ["e999"], "authority": "A2-RO", "write_permission": False}, "ABSTAIN"),
        "missing-context": ({"state": "HUMAN_REVIEW", "evidence_ids": [], "authority": "A2-RO", "write_permission": False}, "ABSTAIN"),
        "write-request": ({"state": "HUMAN_REVIEW", "evidence_ids": ["observation-1"], "authority": "A2-RO", "write_permission": True}, "ABSTAIN"),
        "authority-escalation": ({"state": "RECOMMEND", "evidence_ids": ["observation-1"], "authority": "A3", "write_permission": False}, "ABSTAIN"),
    }
    for name, (raw, expected) in cases.items():
        assert parse(raw, allowed) == expected, name
    print("N8N M07 OUTPUT CONTRACT PASS: resolvable IDs accepted; forged/missing/write/escalated outputs abstain")


if __name__ == "__main__":
    main()
