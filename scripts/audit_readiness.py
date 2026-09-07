"""Validate the structured readiness matrix and report its actual open evidence."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MATRIX = ROOT / "docs/plans/READINESS-MATRIX.json"
ALLOWED = {"IMPLEMENTED", "IMPLEMENTED_OFFLINE", "PARTIAL", "OPEN"}


def main():
    matrix = json.loads(MATRIX.read_text(encoding="utf-8"))
    assert matrix.get("version") == "readiness-matrix/v1"
    criteria = matrix.get("criteria")
    assert isinstance(criteria, list) and criteria
    ids = set()
    open_items = []
    expected = {"BR-13", "BR-14", "BR-15", "BR-16a", "BR-17", "BR-18b", "BR-19"}
    for item in criteria:
        assert item["id"] not in ids
        ids.add(item["id"])
        assert item["status"] in ALLOWED
        for field in ("implementation_refs", "test_refs", "missing_evidence"):
            assert isinstance(item.get(field), list)
            for ref in item[field]:
                if field != "missing_evidence":
                    assert (ROOT / ref).exists(), (item["id"], ref)
        if item["status"] in {"PARTIAL", "OPEN", "IMPLEMENTED_OFFLINE"} and item["missing_evidence"]:
            open_items.append(item)
        if item["status"] == "IMPLEMENTED" and item["missing_evidence"]:
            raise AssertionError(f"implemented item still has missing evidence: {item['id']}")
    assert ids == expected, (ids, expected)
    assert matrix["overall"] == "NOT_READY_FOR_PRODUCTION"
    assert open_items
    print(f"READINESS AUDIT: {matrix['overall']}")
    for item in criteria:
        gap = "; ".join(item["missing_evidence"]) if item["missing_evidence"] else "none"
        print(f"- {item['id']}: {item['status']} | missing: {gap}")
    print("Offline implementation/tests are separated from operated, live and beginner-pilot evidence.")


if __name__ == "__main__":
    main()
