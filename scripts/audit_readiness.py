"""Audit readiness claims against the structured matrix and CI wiring."""
import json
import re
import sys
from pathlib import Path


DEFAULT_ROOT = Path(__file__).resolve().parents[1]
ALLOWED = {"IMPLEMENTED", "IMPLEMENTED_OFFLINE", "PARTIAL", "OPEN"}
EXPECTED = {"BR-13", "BR-14", "BR-15", "BR-16a", "BR-17", "BR-18b", "BR-19"}
CLAIM_KINDS = {"implementation", "test", "operated", "external"}
CLAIM_STATUSES = {"IMPLEMENTED_OFFLINE", "VERIFIED_OFFLINE", "PARTIAL", "MISSING"}
CI_REQUIRED = {
    "scripts/smoke_br16a_offline.py": ".github/workflows/curriculum-ci.yml",
    "scripts/smoke_br18b_backup_restore.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m10_identity_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m11_identity_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_backup_source_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_runtime_store_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_recovery_journal_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m07_tool_artifact_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m07_backup_sidecar_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/run_n8n_engine_regression.py": ".github/workflows/mission-agent-path-ci.yml",
    "scripts/run_n8n_m06_schedule_regression.py": ".github/workflows/mission-agent-path-ci.yml",
}
PUBLIC_READINESS_DOCS = ("README.md", "curriculum/README.md")
PACKAGE_IDS = {f"RP-{number:02d}" for number in range(1, 11)}
PACKAGE_STATUSES = {"PARTIAL", "OPEN"}
REVIEW_IDS = {f"R{number:02d}" for number in range(1, 17)}
REVIEW_STATUSES = {"PARTIAL", "OPEN"}
BASELINE_RE = re.compile(r"^[0-9a-f]{40}$")
PLAN_METADATA_RE = re.compile(r"^<!-- readiness-(as-of|main-baseline): ([^>]+) -->$", re.MULTILINE)
PLAN_PACKAGE_RE = re.compile(r"^\| (RP-\d+) \|.*\| (PARTIAL|OPEN) [—-]", re.MULTILINE)


def fail(message):
    raise AssertionError(message)


def audit_snapshot_metadata(matrix, graph, plan_text):
    """Require the three readiness artifacts to describe one reviewed main base."""
    plan_metadata = dict(PLAN_METADATA_RE.findall(plan_text))
    if set(plan_metadata) != {"as-of", "main-baseline"}:
        fail("plan lacks exact readiness snapshot metadata")
    as_of = matrix.get("as_of")
    baseline = matrix.get("main_baseline")
    if not isinstance(as_of, str) or not re.fullmatch(r"\d{4}-\d{2}-\d{2}", as_of):
        fail("matrix has invalid readiness as_of")
    if not isinstance(baseline, str) or not BASELINE_RE.fullmatch(baseline):
        fail("matrix has invalid main baseline")
    if graph.get("as_of") != as_of or graph.get("main_baseline") != baseline:
        fail("matrix/evidence graph readiness snapshot mismatch")
    if plan_metadata["as-of"] != as_of or plan_metadata["main-baseline"] != baseline:
        fail("matrix/plan readiness snapshot mismatch")


def audit_package_statuses(matrix, plan_text):
    packages = matrix.get("packages")
    if not isinstance(packages, list):
        fail("matrix lacks structured remediation packages")
    expected = {}
    for package in packages:
        if not isinstance(package, dict):
            fail("invalid remediation package")
        package_id, status = package.get("id"), package.get("status")
        if package_id in expected or package_id not in PACKAGE_IDS or status not in PACKAGE_STATUSES:
            fail(f"invalid or duplicate remediation package: {package_id}")
        expected[package_id] = status
    if set(expected) != PACKAGE_IDS:
        fail(f"unexpected remediation package IDs: {sorted(expected)}")
    actual = dict(PLAN_PACKAGE_RE.findall(plan_text))
    if actual != expected:
        fail("plan package status does not match readiness matrix")


def audit_review_findings(matrix, criteria_by_id, plan_text):
    """Keep every reviewed risk explicitly linked to its owning package/BRs."""
    findings = matrix.get("review_findings")
    if not isinstance(findings, list):
        fail("matrix lacks structured review findings")
    seen = set()
    for finding in findings:
        if not isinstance(finding, dict):
            fail("invalid review finding")
        finding_id = finding.get("id")
        package = finding.get("package")
        criteria = finding.get("criteria")
        status = finding.get("status")
        scope = finding.get("scope")
        if finding_id in seen or finding_id not in REVIEW_IDS:
            fail(f"invalid or duplicate review finding: {finding_id}")
        if package not in PACKAGE_IDS or status not in REVIEW_STATUSES:
            fail(f"review finding has invalid package/status: {finding_id}")
        if not isinstance(criteria, list) or not criteria or any(item not in criteria_by_id for item in criteria):
            fail(f"review finding has invalid criteria links: {finding_id}")
        if not isinstance(scope, str) or not scope.strip():
            fail(f"review finding lacks scope: {finding_id}")
        if not re.search(rf"^\| {re.escape(finding_id)} / P[12] \|", plan_text, re.MULTILINE):
            fail(f"review finding is missing from plan: {finding_id}")
        seen.add(finding_id)
    if seen != REVIEW_IDS:
        fail(f"review finding mapping is incomplete: {sorted(seen)}")


def audit_evidence_graph(root, criteria_by_id):
    graph_path = root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
    graph = json.loads(graph_path.read_text(encoding="utf-8"))
    if graph.get("version") != "readiness-evidence-graph/v1":
        fail("unsupported readiness evidence graph version")
    entries = graph.get("criteria")
    if not isinstance(entries, list) or not entries:
        fail("readiness evidence graph requires non-empty criteria")
    graph_ids, claim_ids = set(), set()
    for entry in entries:
        criterion_id = entry.get("criterion_id")
        if criterion_id in graph_ids or criterion_id not in criteria_by_id:
            fail(f"invalid or duplicate evidence graph criterion: {criterion_id}")
        graph_ids.add(criterion_id)
        claims = entry.get("claims")
        if not isinstance(claims, list) or not claims:
            fail(f"{criterion_id} has no evidence graph claims")
        kinds = set()
        for claim in claims:
            claim_id, kind, status = claim.get("id"), claim.get("kind"), claim.get("status")
            if not isinstance(claim_id, str) or not claim_id or claim_id in claim_ids:
                fail(f"invalid or duplicate evidence claim: {claim_id}")
            claim_ids.add(claim_id)
            if kind not in CLAIM_KINDS or status not in CLAIM_STATUSES:
                fail(f"{criterion_id} has invalid evidence claim kind/status: {claim_id}")
            kinds.add(kind)
            if not isinstance(claim.get("scope"), str) or not claim["scope"].strip():
                fail(f"{claim_id} lacks evidence scope")
            if kind == "operated" and status == "VERIFIED_OFFLINE":
                scope = claim["scope"].casefold()
                for boundary in ("synthetic", "read-only", "local"):
                    if boundary not in scope:
                        fail(f"{claim_id} operated scope must retain {boundary} boundary")
            matrix_field = claim.get("matrix_field")
            if kind in {"implementation", "test"}:
                expected_field = f"{kind}_refs"
                if matrix_field != expected_field:
                    fail(f"{claim_id} must link {expected_field}")
                if not criteria_by_id[criterion_id].get(matrix_field):
                    fail(f"{claim_id} links an empty matrix field")
            elif matrix_field is not None:
                fail(f"{claim_id} must not link a matrix ref field")
            plan_refs = claim.get("plan_refs")
            if not isinstance(plan_refs, list) or not plan_refs:
                fail(f"{claim_id} lacks plan refs")
            for plan_ref in plan_refs:
                if not isinstance(plan_ref, dict):
                    fail(f"{claim_id} has invalid plan ref")
                path, marker = plan_ref.get("path"), plan_ref.get("marker")
                target = root / path if isinstance(path, str) else None
                if target is None or not target.is_file() or not isinstance(marker, str) or marker not in target.read_text(encoding="utf-8"):
                    fail(f"{claim_id} has unresolved plan ref")
            ci = claim.get("ci", [])
            if kind == "test" and status == "VERIFIED_OFFLINE" and (not isinstance(ci, list) or not ci):
                fail(f"{claim_id} lacks CI evidence")
            if not isinstance(ci, list):
                fail(f"{claim_id} has invalid CI evidence")
            for ci_ref in ci:
                if not isinstance(ci_ref, dict):
                    fail(f"{claim_id} has invalid CI ref")
                workflow, command = ci_ref.get("workflow"), ci_ref.get("command")
                workflow_path = root / workflow if isinstance(workflow, str) else None
                if workflow_path is None or not workflow_path.is_file() or not isinstance(command, str) or command not in workflow_path.read_text(encoding="utf-8"):
                    fail(f"{claim_id} has unresolved CI evidence")
        if not {"implementation", "test"}.issubset(kinds) or not kinds.intersection({"operated", "external"}):
            fail(f"{criterion_id} lacks a complete implementation/test/external evidence partition")
        if criteria_by_id[criterion_id]["status"] == "IMPLEMENTED" and any(claim["status"] in {"PARTIAL", "MISSING"} for claim in claims):
            fail(f"implemented criterion has incomplete evidence graph claims: {criterion_id}")
    if graph_ids != set(criteria_by_id):
        fail(f"evidence graph/matrix criterion mismatch: {sorted(graph_ids)}")
    return len(claim_ids)


def audit_public_readiness_boundary(root, overall):
    """Keep the two public entrypoints aligned with the scoped readiness state."""
    if overall != "NOT_READY_FOR_PRODUCTION":
        return
    for relative in PUBLIC_READINESS_DOCS:
        path = root / relative
        if not path.is_file():
            fail(f"public readiness document is missing: {relative}")
        content = path.read_text(encoding="utf-8")
        if "NOT_READY_FOR_PRODUCTION" not in content:
            fail(f"public readiness document lacks current boundary: {relative}")
    curriculum = (root / "curriculum/README.md").read_text(encoding="utf-8").casefold()
    if "learner-operable" in curriculum:
        fail("curriculum overclaims learner-operable readiness while production remains not ready")


def audit(root):
    matrix_path = root / "docs/plans/READINESS-MATRIX.json"
    plan_path = root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
    matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
    graph = json.loads((root / "docs/plans/READINESS-EVIDENCE-GRAPH.json").read_text(encoding="utf-8"))
    plan_text = plan_path.read_text(encoding="utf-8")
    if matrix.get("version") != "readiness-matrix/v1":
        fail("unsupported readiness matrix version")
    if matrix.get("overall") != "NOT_READY_FOR_PRODUCTION":
        fail("readiness matrix must remain NOT_READY_FOR_PRODUCTION")
    audit_snapshot_metadata(matrix, graph, plan_text)
    audit_package_statuses(matrix, plan_text)
    criteria = matrix.get("criteria")
    if not isinstance(criteria, list) or not criteria:
        fail("readiness matrix requires non-empty criteria")
    seen, partial, criteria_by_id = set(), [], {}
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
        criteria_by_id[item_id] = item
    if seen != EXPECTED:
        fail(f"unexpected criterion IDs: {sorted(seen)}")
    audit_review_findings(matrix, criteria_by_id, plan_text)
    claim_count = audit_evidence_graph(root, criteria_by_id)
    audit_public_readiness_boundary(root, matrix["overall"])
    for script, workflow in CI_REQUIRED.items():
        if not (root / script).is_file():
            fail(f"required regression script is missing: {script}")
        if script not in (root / workflow).read_text(encoding="utf-8"):
            fail(f"required regression is not wired to CI: {script}")
    for line_number, line in enumerate(plan_text.splitlines(), start=1):
        normalized = line.casefold()
        if "ready for production" in normalized and not re.search(r"not[ _-]?ready|chưa|không", normalized):
            fail(f"plan overclaims production readiness at line {line_number}")
    return matrix, partial, claim_count


def main():
    root = Path(sys.argv[1]).resolve() if len(sys.argv) == 2 else DEFAULT_ROOT
    matrix, partial, claim_count = audit(root)
    print(f"READINESS AUDIT: {matrix['overall']}")
    print("- structured partial/open criteria: " + ", ".join(partial))
    print("- implementation/test refs exist; required M00-M11 regressions are wired to CI")
    print(f"- plan/matrix/evidence graph share main baseline {matrix['main_baseline'][:12]}")
    print(f"- evidence graph resolves {claim_count} scoped claims to matrix refs, plan markers and declared CI commands")
    print("- plan and public entrypoints retain scoped readiness boundaries")


if __name__ == "__main__":
    main()
