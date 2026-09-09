"""Audit readiness claims against the structured matrix and CI wiring."""
import json
import re
import sys
from pathlib import Path


DEFAULT_ROOT = Path(__file__).resolve().parents[1]
ALLOWED = {"IMPLEMENTED", "IMPLEMENTED_OFFLINE", "PARTIAL", "OPEN"}
EXPECTED = {"BR-13", "BR-14", "BR-15", "BR-16a", "BR-17", "BR-18b", "BR-19"}
CI_REQUIRED = {
    "scripts/smoke_br16a_offline.py": ".github/workflows/curriculum-ci.yml",
    "scripts/smoke_br18b_backup_restore.py": ".github/workflows/curriculum-ci.yml",
    "scripts/run_n8n_engine_regression.py": ".github/workflows/mission-agent-path-ci.yml",
}


def fail(message):
    raise AssertionError(message)


def audit(root):
    matrix_path = root / "docs/plans/READINESS-MATRIX.json"
    plan_path = root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
    matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
    if matrix.get("version") != "readiness-matrix/v1":
        fail("unsupported readiness matrix version")
    if matrix.get("overall") != "NOT_READY_FOR_PRODUCTION":
        fail("readiness matrix must remain NOT_READY_FOR_PRODUCTION")
    criteria = matrix.get("criteria")
    if not isinstance(criteria, list) or not criteria:
        fail("readiness matrix requires non-empty criteria")
    seen, partial = set(), []
    for item in criteria:
        item_id, status = item.get("id"), item.get("status")
        if item_id in seen or status not in ALLOWED:
            fail(f"invalid or duplicate criterion: {item_id}")
        seen.add(item_id)
        for field in ("implementation_refs", "test_refs", "missing_evidence"):
            values = item.get(field)
            if not isinstance(values, list) or (field != "missing_evidence" and not values):
                fail(f"{item_id} lacks structured {field}")
            if field != "missing_evidence":
                for ref in values:
                    if not isinstance(ref, str) or not (root / ref).is_file():
                        fail(f"{item_id} has missing ref: {ref}")
        if status == "IMPLEMENTED" and item["missing_evidence"]:
            fail(f"implemented item still has missing evidence: {item_id}")
        if status in {"PARTIAL", "OPEN", "IMPLEMENTED_OFFLINE"}:
            if not item["missing_evidence"]:
                fail(f"non-final item lacks explicit remaining evidence: {item_id}")
            partial.append(item_id)
    if seen != EXPECTED:
        fail(f"unexpected criterion IDs: {sorted(seen)}")
    for script, workflow in CI_REQUIRED.items():
        if script not in (root / workflow).read_text(encoding="utf-8"):
            fail(f"required regression is not wired to CI: {script}")
    for line_number, line in enumerate(plan_path.read_text(encoding="utf-8").splitlines(), start=1):
        normalized = line.casefold()
        if "ready for production" in normalized and not re.search(r"not[ _-]?ready|chưa|không", normalized):
            fail(f"plan overclaims production readiness at line {line_number}")
    return matrix, partial


def main():
    root = Path(sys.argv[1]).resolve() if len(sys.argv) == 2 else DEFAULT_ROOT
    matrix, partial = audit(root)
    print(f"READINESS AUDIT: {matrix['overall']}")
    print("- structured partial/open criteria: " + ", ".join(partial))
    print("- implementation/test refs exist; required M00-M11 regressions are wired to CI")
    print("- plan contains no unqualified production-readiness claim")


if __name__ == "__main__":
    main()
