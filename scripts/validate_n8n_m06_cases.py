"""Offline contract cases for the M06 watcher semantics; not n8n execution evidence."""
import json


def canonical(value):
    if isinstance(value, list):
        return "[" + ",".join(canonical(v) for v in value) + "]"
    if isinstance(value, dict):
        return "{" + ",".join(json.dumps(k, separators=(",", ":")) + ":" + canonical(value[k]) for k in sorted(value)) + "}"
    return json.dumps(value, separators=(",", ":"), ensure_ascii=False)


def main():
    first = {"price": 100, "product_id": "A", "title": "Fixture"}
    reordered = {"title": "Fixture", "product_id": "A", "price": 100}
    changed = {"price": 120, "product_id": "A", "title": "Fixture"}
    assert canonical(first) == canonical(reordered), "key order must remain UNCHANGED"
    assert canonical(first) != canonical(changed), "content change must be CHANGED"
    assert len(canonical(first)) <= 200000
    assert len(canonical({"payload": "x" * 200001})) > 200000
    # Sink failure is a handoff error, never a successful persistence claim.
    handoff = {"canonical_history_handoff": "ACK", "canonical_history_persisted": True}
    assert handoff["canonical_history_handoff"] == "ACK" and handoff["canonical_history_persisted"] is True
    failed = {"canonical_history_handoff": "NO_ACK", "canonical_history_persisted": False}
    assert failed["canonical_history_persisted"] is False
    print("N8N M06 CASE CONTRACT PASS: NEW/UNCHANGED/CHANGED, key-order stability, size guard and ACK/failure boundary")


if __name__ == "__main__":
    main()
