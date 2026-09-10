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
        for relative in ("scripts/audit_readiness.py", "docs/plans/READINESS-MATRIX.json", "docs/plans/READINESS-EVIDENCE-GRAPH.json", "docs/plans/PRE-MERGE-REMEDIATION-737E85A.md", "docs/plans/REVIEW-REMEDIATION-PLAN.md", ".github/workflows/curriculum-ci.yml", ".github/workflows/mission-agent-path-ci.yml"):
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

    def test_unqualified_production_claim_is_rejected(self):
        plan = self.root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
        plan.write_text(plan.read_text(encoding="utf-8") + "\nRepository is ready for production.\n", encoding="utf-8")
        self.assertIn("overclaims production readiness", self.run_audit(False))

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


if __name__ == "__main__":
    unittest.main()
