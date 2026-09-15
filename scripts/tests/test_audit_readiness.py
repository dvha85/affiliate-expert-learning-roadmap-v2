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
        for relative in ("scripts/audit_readiness.py", "scripts/smoke_br16a_offline.py", "scripts/mutate_m10_identity_guard.py", "scripts/mutate_m11_identity_guard.py", "scripts/mutate_backup_source_guard.py", "scripts/mutate_runtime_store_path_guard.py", "scripts/mutate_recovery_journal_path_guard.py", "scripts/mutate_m07_tool_artifact_path_guard.py", "scripts/mutate_m07_portable_input_guard.py", "scripts/mutate_general_portable_input_guard.py", "scripts/mutate_internal_portable_input_guard.py", "scripts/mutate_m07_backup_sidecar_path_guard.py", "scripts/mutate_m08_m07_proposal_path_guard.py", "scripts/mutate_mission_portable_input_guard.py", "scripts/mutate_mission_stop_immutability_guard.py", "scripts/mutate_m07_strict_output_decoder.py", "scripts/mutate_m07_registry_strict_decoder.py", "lab/affiliate-bot/cmd/bot/advisor_fixture.go", "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go", "lab/affiliate-bot/cmd/bot/artifact_publish_test.go", "lab/affiliate-bot/internal/store/history.go", "lab/affiliate-bot/internal/store/history_test.go", "lab/affiliate-bot/cmd/bot/action_store_test.go", "lab/affiliate-bot/cmd/bot/outcome_store_test.go", "lab/affiliate-bot/cmd/bot/m11_registry.go", "lab/affiliate-bot/cmd/bot/accesstrade_receipt.go", "lab/affiliate-bot/cmd/bot/accesstrade_import_test.go", "lab/n8n/COMPATIBILITY.md", "README.md", "curriculum/README.md", "docs/plans/READINESS-MATRIX.json", "docs/plans/READINESS-EVIDENCE-GRAPH.json", "docs/plans/BEGINNER-READINESS-PLAN.md", "docs/plans/PRE-MERGE-REMEDIATION-737E85A.md", "docs/plans/REVIEW-REMEDIATION-PLAN.md", ".github/workflows/curriculum-ci.yml", ".github/workflows/mission-agent-path-ci.yml"):
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

    def test_n8n_engine_node_runtime_drift_is_rejected(self):
        workflow = self.root / ".github/workflows/mission-agent-path-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace('node-version: "24"', 'node-version: "22"', 1), encoding="utf-8")
        self.assertIn("n8n engine CI no longer pins", self.run_audit(False))

    def test_n8n_schedule_engine_regression_removal_is_rejected(self):
        workflow = self.root / ".github/workflows/mission-agent-path-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("run_n8n_m06_schedule_regression.py", "removed_n8n_schedule_regression.py", 1), encoding="utf-8")
        self.assertIn("n8n engine CI no longer pins", self.run_audit(False))

    def test_n8n_engine_cache_gate_removal_is_rejected(self):
        workflow = self.root / ".github/workflows/mission-agent-path-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("actions/cache@v4", "removed_n8n_cache", 1), encoding="utf-8")
        self.assertIn("n8n engine CI cache/gate is missing", self.run_audit(False))

    def test_deterministic_shard_removal_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("learner-bot-race:", "removed-learner-bot-race:", 1), encoding="utf-8")
        self.assertIn("deterministic CI shard/cache is missing", self.run_audit(False))

    def test_missing_m11_recovery_admission_approval_guard_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("approval, approvalOK := approvals[x.NewApprovalID]", "approval guard removed", 1), encoding="utf-8")
        self.assertIn("M11 recovery-admission approval regression is missing", self.run_audit(False))

    def test_missing_advisor_fixture_output_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
        source.write_text(source.read_text(encoding="utf-8").replace("fixture output parent must be an existing non-symlink directory", "fixture output parent guard removed", 1), encoding="utf-8")
        self.assertIn("advisor fixture output-parent regression is missing", self.run_audit(False))

    def test_missing_backup_restore_missing_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func ensureOutputParentBeforeCreate", "func removedOutputParentPreflight", 1), encoding="utf-8")
        self.assertIn("backup/restore missing-parent regression is missing", self.run_audit(False))

    def test_missing_immutable_artifact_missing_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func artifactOutputDirectory(path string) (string, error) {\n\tdir := filepath.Dir(path)\n\tif err := ensureOutputParentBeforeCreate(dir);", "func artifactOutputDirectory(path string) (string, error) {\n\tdir := filepath.Dir(path)\n\tif err := removedOutputParentPreflight(dir);", 1), encoding="utf-8")
        self.assertIn("immutable artifact missing-parent regression is missing", self.run_audit(False))

    def test_missing_runtime_missing_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func ensureRuntimeDirectory(dir string) error {\n\tif err := ensureOutputParentBeforeCreate(dir);", "func ensureRuntimeDirectory(dir string) error {\n\tif err := removedOutputParentPreflight(dir);", 1), encoding="utf-8")
        self.assertIn("runtime missing-parent regression is missing", self.run_audit(False))

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

    def test_missing_m07_portable_input_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_portable_input_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_general_portable_input_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_general_portable_input_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

    def test_missing_internal_portable_input_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_internal_portable_input_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

    def test_missing_m07_backup_sidecar_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_backup_sidecar_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m08_m07_proposal_path_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m08_m07_proposal_path_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_mission_portable_input_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_mission_portable_input_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_durable_stop_immutability_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_mission_stop_immutability_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_m07_strict_output_decoder_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_strict_output_decoder.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

    def test_missing_m07_strict_registry_decoder_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_m07_registry_strict_decoder.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("unresolved CI evidence", self.run_audit(False))

    def test_missing_r10_process_barrier_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ReservationCapOneAcrossTwentyFourBotProcesses", "MissingProcessBarrierRegression", 1), encoding="utf-8")
        self.assertIn("R10 process-barrier regression is missing", self.run_audit(False))

    def test_missing_m11_process_barrier_smoke_regression_is_rejected(self):
        source = self.root / "scripts/smoke_br16a_offline.py"
        source.write_text(source.read_text(encoding="utf-8").replace('"m11-reserve-barrier"', '"missing-m11-reserve-barrier"', 1), encoding="utf-8")
        self.assertIn("R10 M11 process-barrier regression is missing", self.run_audit(False))

    def test_missing_m11_outcome_process_barrier_smoke_regression_is_rejected(self):
        source = self.root / "scripts/smoke_br16a_offline.py"
        source.write_text(source.read_text(encoding="utf-8").replace('"m11-outcome-barrier"', '"missing-m11-outcome-barrier"', 1), encoding="utf-8")
        self.assertIn("R10 M11 process-barrier regression is missing", self.run_audit(False))

    def test_missing_r10_reservation_commit_fault_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ReservationCommitFaultDoesNotConsumeCap", "MissingCommitFaultRegression", 1), encoding="utf-8")
        self.assertIn("R10 reservation commit-fault regression is missing", self.run_audit(False))

    def test_missing_mission_state_post_rename_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ReservationPostRenameSyncFaultRequiresRecoveryInsteadOfRetry", "MissingPostRenameRegression", 1), encoding="utf-8")
        self.assertIn("mission-state post-rename regression is missing", self.run_audit(False))

    def test_missing_durable_stop_post_rename_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_state_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionStopJournalRecoversEveryPostRenameBoundary", "MissingStopPostRenameRegression", 1), encoding="utf-8")
        self.assertIn("durable STOP post-rename regression is missing", self.run_audit(False))

    def test_missing_durable_stop_journal_recovery_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("recoverM11ManualStopJournal", "missingM11ManualStopJournal"), encoding="utf-8")
        self.assertIn("durable STOP journal recovery is missing", self.run_audit(False))

    def test_missing_durable_stop_journal_backup_recovery_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("recoverM11ManualStopJournal(args[1])", "missingM11ManualStopJournal(args[1])", 1), encoding="utf-8")
        self.assertIn("durable STOP journal recovery is missing from the backup path", self.run_audit(False))

    def test_missing_durable_stop_immutability_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_state_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionStopCannotOverwriteDurableReason", "MissingStopImmutabilityRegression", 1), encoding="utf-8")
        self.assertIn("durable STOP immutability regression is missing", self.run_audit(False))

    def test_missing_jsonl_framing_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/internal/store/history_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestJSONLOpenRejectsIncompleteFinalLine", "MissingJSONLFramingRegression", 1), encoding="utf-8")
        self.assertIn("canonical JSONL shared framing regression is missing", self.run_audit(False))

    def test_missing_m10_canonical_output_disclosure_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10DisclosesCanonicalArtifactWhenPortableOutputConflicts", "MissingCanonicalOutputDisclosureRegression", 1), encoding="utf-8")
        self.assertIn("M10 canonical-output disclosure regression is missing", self.run_audit(False))

    def test_missing_m10_gate_evaluation_identity_regression_is_rejected(self):
        source = self.root / "core/m10/cost_bound_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestCanaryGateIdentityIncludesEvaluationTime", "MissingGateEvaluationIdentityRegression", 1), encoding="utf-8")
        self.assertIn("M10 gate evaluation-identity regression is missing", self.run_audit(False))

    def test_missing_m11_authorization_time_identity_regression_is_rejected(self):
        source = self.root / "core/m11/artifact_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("distinct authorization times reused an immutable ID", "missing authorization time identity regression", 1), encoding="utf-8")
        self.assertIn("M11 authorization time-identity regression is missing", self.run_audit(False))

    def test_missing_registry_publish_uncertainty_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10GateDisclosesRegistryPublishUncertainty", "MissingRegistryPublishUncertaintyRegression", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_lifecycle_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11LifecycleDisclosesRegistryPublishUncertainty", "MissingM11LifecycleRegistryPublishUncertainty", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_evaluation_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11EvaluationDisclosesRegistryPublishUncertainty", "MissingM11EvaluationRegistryPublishUncertainty", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_cycle_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11CycleDisclosesRegistryPublishUncertainty", "MissingM11CycleRegistryPublishUncertainty", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_gate_authorization_reservation_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11GateAuthorizationAndReservationDiscloseRegistryPublishUncertainty", "MissingM11GateAuthorizationReservationRegistryPublishUncertainty", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_reconciliation_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11ReconcileDisclosesRegistryPublishUncertainty", "MissingM11ReconcileRegistryPublishUncertainty", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m11_recovery_admission_registry_publish_uncertainty_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("recovery admission uncertainty was not disclosed", "missing recovery admission uncertainty regression", 1), encoding="utf-8")
        self.assertIn("M10/M11 registry publish-uncertainty regression is missing", self.run_audit(False))

    def test_missing_m08_policy_visible_artifact_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM08PolicyReportsUnconfirmedVisibleArtifact", "MissingM08PolicyVisibleArtifactRegression", 1), encoding="utf-8")
        self.assertIn("immutable artifact post-publish regression is missing", self.run_audit(False))

    def test_missing_m11_recovery_export_visible_artifact_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("recovery handoff uncertainty was not disclosed", "missing recovery handoff uncertainty regression", 1), encoding="utf-8")
        self.assertIn("immutable artifact post-publish regression is missing", self.run_audit(False))

    def test_missing_m10_visible_journal_publish_uncertainty_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10VisibleJournalPublishUncertaintyDefersLockedReplay", "MissingM10VisibleJournalPublishUncertaintyRegression", 1), encoding="utf-8")
        self.assertIn("M10 visible-journal recovery regression is missing", self.run_audit(False))

    def test_missing_m10_execution_journal_publish_uncertainty_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ExecutionJournalVisiblePublishUncertaintyDefersLockedReplay", "MissingM10ExecutionJournalPublishUncertaintyRegression", 1), encoding="utf-8")
        self.assertIn("M10 visible-journal recovery regression is missing", self.run_audit(False))

    def test_missing_m11_outcome_journal_publish_uncertainty_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11OutcomeJournalVisiblePublishUncertaintyDefersLockedReplay", "MissingM11OutcomeJournalPublishUncertaintyRegression", 1), encoding="utf-8")
        self.assertIn("M11 outcome visible-journal recovery regression is missing", self.run_audit(False))

    def test_missing_m11_visible_outcome_append_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11OutcomeDisclosesVisibleAppendAcknowledgementUncertainty", "MissingM11VisibleOutcomeAppendRegression", 1), encoding="utf-8")
        self.assertIn("M11 visible outcome-append acknowledgement regression is missing", self.run_audit(False))

    def test_missing_m11_published_outcome_journal_cleanup_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11OutcomeDisclosesPublishedJournalCleanupUncertainty", "MissingM11PublishedOutcomeJournalCleanupRegression", 1), encoding="utf-8")
        self.assertIn("M11 published outcome journal cleanup regression is missing", self.run_audit(False))

    def test_missing_m11_execution_journal_publish_uncertainty_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11ExecutionJournalsVisiblePublishUncertaintyDefersLockedReplay", "MissingM11ExecutionJournalPublishUncertaintyRegression", 1), encoding="utf-8")
        self.assertIn("M11 execution visible-journal recovery regression is missing", self.run_audit(False))

    def test_missing_m11_execution_journal_cleanup_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM11ExecutionJournalsDisclosePublishedCleanupUncertainty", "MissingM11ExecutionJournalCleanupRegression", 1), encoding="utf-8")
        self.assertIn("M11 execution visible-journal recovery regression is missing", self.run_audit(False))

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

    def test_selected_source_grounding_disclosure_is_required(self):
        plan_path = self.root / "docs/plans/REVIEW-REMEDIATION-PLAN.md"
        plan_path.write_text(plan_path.read_text(encoding="utf-8").replace("M07 selected-source engine grounding CI", "removed selected-source grounding marker", 1), encoding="utf-8")
        self.assertIn("selected-source M07 grounding runner lacks a scoped plan marker", self.run_audit(False))

    def test_selected_source_forged_commission_scope_is_required(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        br15 = next(item for item in matrix["criteria"] if item["id"] == "BR-15")
        br15["missing_evidence"][0] = "tests: selected campaign metadata reaches M07"
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("does not disclose selected-source forged-commission", self.run_audit(False))

    def test_beginner_plan_selected_source_boundary_is_required(self):
        plan = self.root / "docs/plans/BEGINNER-READINESS-PLAN.md"
        plan.write_text(plan.read_text(encoding="utf-8").replace("sanitized/read-only", "unscoped", 1), encoding="utf-8")
        self.assertIn("beginner plan does not disclose the selected-source offline boundary", self.run_audit(False))


if __name__ == "__main__":
    unittest.main()
