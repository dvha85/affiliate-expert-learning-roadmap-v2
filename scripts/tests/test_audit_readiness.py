"""Negative fixtures run the real readiness auditor in isolated copies."""
import json
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]


class ReadinessAuditTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        for relative in ("scripts/audit_readiness.py", "scripts/mutate_m10_identity_guard.py", "scripts/mutate_m11_identity_guard.py", "scripts/mutate_backup_source_guard.py", "scripts/mutate_runtime_store_path_guard.py", "scripts/mutate_recovery_journal_path_guard.py", "scripts/mutate_m07_tool_artifact_path_guard.py", "scripts/mutate_m07_backup_sidecar_path_guard.py", "scripts/mutate_m08_m07_proposal_path_guard.py", "scripts/mutate_m07_strict_output_decoder.py", "README.md", "curriculum/README.md", "docs/plans/READINESS-MATRIX.json", "docs/plans/READINESS-EVIDENCE-GRAPH.json", "docs/plans/PRE-MERGE-REMEDIATION-737E85A.md", "docs/plans/REVIEW-REMEDIATION-PLAN.md", ".github/workflows/curriculum-ci.yml", ".github/workflows/mission-agent-path-ci.yml"):
            source, target = ROOT / relative, self.root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
        matrix = json.loads((ROOT / "docs/plans/READINESS-MATRIX.json").read_text(encoding="utf-8"))
        for item in matrix["criteria"]:
            for field in ("implementation_refs", "test_refs"):
                for relative in item[field]:
                    source, target = ROOT / relative, self.root / relative
                    target.parent.mkdir(parents=True, exist_ok=True)
                    shutil.copy2(source, target)

    def run_audit(self, valid):
        result = subprocess.run([sys.executable, str(self.root / "scripts/audit_readiness.py"), str(self.root)], text=True, capture_output=True)
        self.assertEqual(result.returncode, 0 if valid else 1, result.stdout + result.stderr)
        return result.stdout + result.stderr

    def test_canonical_matrix(self):
        self.assertIn("NOT_READY_FOR_PRODUCTION", self.run_audit(True))

    def test_snapshot_mismatch_is_rejected(self):
        graph_path = self.root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
        graph = json.loads(graph_path.read_text(encoding="utf-8"))
        graph["main_baseline"] = "0" * 40
        graph_path.write_text(json.dumps(graph), encoding="utf-8")
        self.assertIn("snapshot mismatch", self.run_audit(False))

    def test_plan_package_status_mismatch_is_rejected(self):
        plan_path = self.root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
        plan_path.write_text(plan_path.read_text(encoding="utf-8").replace("| RP-07 | M11 lifecycle + mở rộng restore (07a), rồi full chain/walkthrough (07b) | 07a sau RP-02…RP-06; 07b sau gate lifecycle/restore của 07a | L; chia 07a/07b | PARTIAL", "| RP-07 | M11 lifecycle + mở rộng restore (07a), rồi full chain/walkthrough (07b) | 07a sau RP-02…RP-06; 07b sau gate lifecycle/restore của 07a | L; chia 07a/07b | TODO"), encoding="utf-8")
        self.assertIn("package status does not match", self.run_audit(False))

    def test_missing_reference_is_rejected(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["criteria"][0]["test_refs"] = ["scripts/missing.py"]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("missing ref", self.run_audit(False))

    def test_final_status_with_gap_is_rejected(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["criteria"][0]["status"] = "IMPLEMENTED"
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("still has missing evidence", self.run_audit(False))

    def test_missing_ci_regression_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/smoke_br16a_offline.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_identity_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m10_identity_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

    def test_missing_backup_source_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_backup_source_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_runtime_store_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_runtime_store_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_recovery_journal_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_recovery_journal_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m07_tool_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_tool_artifact_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m07_backup_sidecar_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_backup_sidecar_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m08_m07_proposal_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m08_m07_proposal_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m07_strict_output_decoder_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_strict_output_decoder.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

    def test_unqualified_production_claim_is_rejected(self):
        plan = self.root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
        plan.write_text(plan.read_text(encoding="utf-8") + "\nRepository is ready for production.\n", encoding="utf-8")
        self.assertIn("overclaims production readiness", self.run_audit(False))

    def test_public_readiness_boundary_is_required(self):
        curriculum = self.root / "curriculum/README.md"
        curriculum.write_text(curriculum.read_text(encoding="utf-8").replace("NOT_READY_FOR_PRODUCTION", "CURRENT_STATUS"), encoding="utf-8")
        self.assertIn("public readiness document lacks current boundary", self.run_audit(False))

    def test_public_learner_operable_claim_is_rejected(self):
        curriculum = self.root / "curriculum/README.md"
        curriculum.write_text(curriculum.read_text(encoding="utf-8") + "\nThe curriculum is learner-operable.\n", encoding="utf-8")
        self.assertIn("overclaims learner-operable readiness", self.run_audit(False))

    def test_graph_claim_with_unwired_ci_command_is_rejected(self):
        graph_path = self.root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
        graph = json.loads(graph_path.read_text(encoding="utf-8"))
        graph["criteria"][0]["claims"][1]["ci"][0]["command"] = "python scripts/not-wired.py"
        graph_path.write_text(json.dumps(graph), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_graph_claim_with_unresolved_plan_marker_is_rejected(self):
        graph_path = self.root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
        graph = json.loads(graph_path.read_text(encoding="utf-8"))
        graph["criteria"][0]["claims"][0]["plan_refs"][0]["marker"] = "missing-plan-marker"
        graph_path.write_text(json.dumps(graph), encoding="utf-8")
        self.assertIn("unresolved plan ref", self.run_audit(False))

    def test_graph_without_external_partition_is_rejected(self):
        graph_path = self.root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
        graph = json.loads(graph_path.read_text(encoding="utf-8"))
        graph["criteria"][0]["claims"] = graph["criteria"][0]["claims"][:2]
        graph_path.write_text(json.dumps(graph), encoding="utf-8")
        self.assertIn("evidence partition", self.run_audit(False))

    def test_verified_operated_claim_without_fixture_boundaries_is_rejected(self):
        graph_path = self.root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
        graph = json.loads(graph_path.read_text(encoding="utf-8"))
        claim = next(item for item in graph["criteria"][0]["claims"] if item["kind"] == "operated")
        claim["scope"] = "Operated run completed."
        graph_path.write_text(json.dumps(graph), encoding="utf-8")
        self.assertIn("operated scope must retain", self.run_audit(False))

    def test_missing_review_finding_mapping_is_rejected(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["review_findings"] = matrix["review_findings"][:-1]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("review finding mapping is incomplete", self.run_audit(False))


if __name__ == "__main__":
    unittest.main()
