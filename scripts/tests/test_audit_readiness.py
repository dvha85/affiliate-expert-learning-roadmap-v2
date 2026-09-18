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
        for relative in ("scripts/audit_readiness.py", "scripts/smoke_br16a_offline.py", "scripts/mutate_m10_identity_guard.py", "scripts/mutate_m11_identity_guard.py", "scripts/mutate_registry_graph_envelope_integrity.py", "scripts/mutate_backup_source_guard.py", "scripts/mutate_runtime_store_path_guard.py", "scripts/mutate_recovery_journal_path_guard.py", "scripts/mutate_m07_tool_artifact_path_guard.py", "scripts/mutate_m07_portable_input_guard.py", "scripts/mutate_general_portable_input_guard.py", "scripts/mutate_internal_portable_input_guard.py", "scripts/mutate_m07_backup_sidecar_path_guard.py", "scripts/mutate_m08_m07_proposal_path_guard.py", "scripts/mutate_mission_portable_input_guard.py", "scripts/mutate_mission_stop_immutability_guard.py", "scripts/mutate_m07_strict_output_decoder.py", "scripts/mutate_m07_registry_strict_decoder.py", "lab/affiliate-bot/cmd/bot/advisor_fixture.go", "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go", "lab/affiliate-bot/cmd/bot/advisor_writer_parent_swap_test.go", "lab/affiliate-bot/cmd/bot/artifact_publish_test.go", "lab/affiliate-bot/cmd/bot/history_schema.go", "lab/affiliate-bot/cmd/bot/watcher_fetch.go", "lab/affiliate-bot/internal/store/history.go", "lab/affiliate-bot/internal/store/history_test.go", "lab/affiliate-bot/cmd/bot/action_store_test.go", "lab/affiliate-bot/cmd/bot/outcome_store_test.go", "lab/affiliate-bot/cmd/bot/learner_schema_alias_test.go", "lab/affiliate-bot/cmd/bot/m11_registry.go", "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go", "lab/affiliate-bot/cmd/bot/advisor_budget.go", "lab/affiliate-bot/cmd/bot/advisor_results.go", "lab/affiliate-bot/cmd/bot/advisor_report.go", "lab/affiliate-bot/cmd/bot/advisor_canary.go", "lab/affiliate-bot/cmd/bot/backup_command.go", "lab/affiliate-bot/cmd/bot/advisor_br10_campaign.go", "lab/affiliate-bot/cmd/bot/advisor_canary_test.go", "lab/affiliate-bot/cmd/bot/accesstrade_import.go", "lab/affiliate-bot/cmd/bot/accesstrade_receipt.go", "lab/affiliate-bot/cmd/bot/accesstrade_import_test.go", "lab/affiliate-bot/cmd/bot/stable_append_posix.go", "lab/affiliate-bot/cmd/bot/stable_append_other.go", "lab/affiliate-bot/cmd/bot/registry_parent_swap_test.go", "lab/n8n/COMPATIBILITY.md", "README.md", "curriculum/README.md", "docs/plans/READINESS-MATRIX.json", "docs/plans/READINESS-EVIDENCE-GRAPH.json", "docs/plans/BEGINNER-READINESS-PLAN.md", "docs/plans/PRE-MERGE-REMEDIATION-737E85A.md", "docs/plans/REVIEW-REMEDIATION-PLAN.md", "docs/architecture/EVIDENCE-M06-ACCESSTRADE-SHOPEE-OPERATED-20260915.md", "docs/architecture/EVIDENCE-BR18B-LOCAL-RECOVERY-DRILL-20260915.md", "docs/architecture/EVIDENCE-BR16B-ASSISTED-FRESH-WORKSPACE-20260915.md", "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-376-20260916.md", "docs/architecture/EVIDENCE-M11-RESTORE-AUTH-EXECUTION-LINEAGE-20260916.md", "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-387-20260916.md", "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-408-20260917.md", "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-410-20260917.md", "docs/architecture/EVIDENCE-REGISTRY-GRAPH-ENVELOPE-MUTATION-20260917.md", "docs/architecture/EVIDENCE-M11-RECOVERY-ADMISSION-FIELD-INTEGRITY-20260917.md", "docs/architecture/EVIDENCE-RP04-SHARED-CANONICAL-CONTEXT-20260918.md", "docs/architecture/EVIDENCE-PR425-POST-MERGE-20260918.md", ".github/workflows/curriculum-ci.yml", ".github/workflows/mission-agent-path-ci.yml"):
            source, target = ROOT / relative, self.root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
        mutation_terminal_chain = self.root / "scripts/mutate_backup_m11_terminal_chain.py"
        mutation_terminal_chain.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "scripts/mutate_backup_m11_terminal_chain.py", mutation_terminal_chain)
        mutation_failed_outcome = self.root / "scripts/mutate_backup_m11_failed_outcome.py"
        mutation_failed_outcome.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "scripts/mutate_backup_m11_failed_outcome.py", mutation_failed_outcome)
        evidence_windows_runtime = self.root / "docs/architecture/EVIDENCE-RP01-WINDOWS-RUNTIME-CI-20260918.md"
        evidence_windows_runtime.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP01-WINDOWS-RUNTIME-CI-20260918.md", evidence_windows_runtime)
        evidence_windows_ancestor_race = self.root / "docs/architecture/EVIDENCE-RP01-WINDOWS-ANCESTOR-RACE-20260918.md"
        evidence_windows_ancestor_race.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP01-WINDOWS-ANCESTOR-RACE-20260918.md", evidence_windows_ancestor_race)
        evidence_post_416 = self.root / "docs/architecture/EVIDENCE-PR416-POST-MERGE-20260918.md"
        evidence_post_416.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR416-POST-MERGE-20260918.md", evidence_post_416)
        evidence_rp02_m08 = self.root / "docs/architecture/EVIDENCE-RP02-SHARED-M08-POLICY-CONTEXT-20260918.md"
        evidence_rp02_m08.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP02-SHARED-M08-POLICY-CONTEXT-20260918.md", evidence_rp02_m08)
        evidence_pr417 = self.root / "docs/architecture/EVIDENCE-PR417-POST-MERGE-20260918.md"
        evidence_pr417.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR417-POST-MERGE-20260918.md", evidence_pr417)
        evidence_pr427 = self.root / "docs/architecture/EVIDENCE-PR427-POST-MERGE-20260918.md"
        evidence_pr427.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR427-POST-MERGE-20260918.md", evidence_pr427)
        evidence_pr429 = self.root / "docs/architecture/EVIDENCE-PR429-POST-MERGE-20260918.md"
        evidence_pr429.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR429-POST-MERGE-20260918.md", evidence_pr429)
        evidence_pr431 = self.root / "docs/architecture/EVIDENCE-PR431-POST-MERGE-20260918.md"
        evidence_pr431.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR431-POST-MERGE-20260918.md", evidence_pr431)
        evidence_rp07a = self.root / "docs/architecture/EVIDENCE-RP07A-M11-ADMISSION-RESTORE-20260918.md"
        evidence_rp07a.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP07A-M11-ADMISSION-RESTORE-20260918.md", evidence_rp07a)
        evidence_rp07b = self.root / "docs/architecture/EVIDENCE-RP07B-M11-RESTORE-CHAIN-20260918.md"
        evidence_rp07b.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP07B-M11-RESTORE-CHAIN-20260918.md", evidence_rp07b)
        evidence_rp08 = self.root / "docs/architecture/EVIDENCE-RP08-M11-TERMINAL-MUTATION-20260918.md"
        evidence_rp08.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP08-M11-TERMINAL-MUTATION-20260918.md", evidence_rp08)
        evidence_rp03_grant = self.root / "docs/architecture/EVIDENCE-RP03-M10-CANONICAL-GRANT-REGISTRY-20260918.md"
        evidence_rp03_grant.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP03-M10-CANONICAL-GRANT-REGISTRY-20260918.md", evidence_rp03_grant)
        evidence_rp03_expiry_stop = self.root / "docs/architecture/EVIDENCE-RP03-EXPIRY-STOP-RESERVE-20260918.md"
        evidence_rp03_expiry_stop.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP03-EXPIRY-STOP-RESERVE-20260918.md", evidence_rp03_expiry_stop)
        evidence_rp03_budget = self.root / "docs/architecture/EVIDENCE-RP03-BUDGET-MONOTONICITY-20260918.md"
        evidence_rp03_budget.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP03-BUDGET-MONOTONICITY-20260918.md", evidence_rp03_budget)
        evidence_outcome_process_kill = self.root / "docs/architecture/EVIDENCE-M11-OUTCOME-PROCESS-KILL-20260916.md"
        evidence_outcome_process_kill.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-OUTCOME-PROCESS-KILL-20260916.md", evidence_outcome_process_kill)
        evidence_after_ledger = self.root / "docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md"
        evidence_after_ledger.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md", evidence_after_ledger)
        evidence_m10_canonical_append = self.root / "docs/architecture/EVIDENCE-M10-CANONICAL-APPEND-PROCESS-KILL-20260916.md"
        evidence_m10_canonical_append.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M10-CANONICAL-APPEND-PROCESS-KILL-20260916.md", evidence_m10_canonical_append)
        evidence_m10_execution_canonical_append = self.root / "docs/architecture/EVIDENCE-M10-EXECUTION-CANONICAL-APPEND-PROCESS-KILL-20260916.md"
        evidence_m10_execution_canonical_append.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M10-EXECUTION-CANONICAL-APPEND-PROCESS-KILL-20260916.md", evidence_m10_execution_canonical_append)
        evidence_377 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-377-20260916.md"
        evidence_377.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-377-20260916.md", evidence_377)
        evidence_source_grant = self.root / "docs/architecture/EVIDENCE-M11-SOURCE-CANARY-GRANT-LINEAGE-20260916.md"
        evidence_source_grant.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-SOURCE-CANARY-GRANT-LINEAGE-20260916.md", evidence_source_grant)
        evidence_379 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-379-20260916.md"
        evidence_379.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-379-20260916.md", evidence_379)
        evidence_recovery_concurrency = self.root / "docs/architecture/EVIDENCE-M11-RECOVERY-ADMISSION-CONCURRENCY-20260916.md"
        evidence_recovery_concurrency.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-RECOVERY-ADMISSION-CONCURRENCY-20260916.md", evidence_recovery_concurrency)
        evidence_381 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-381-20260916.md"
        evidence_381.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-381-20260916.md", evidence_381)
        evidence_expiry = self.root / "docs/architecture/EVIDENCE-M11-EXPIRY-AUTHORITY-NOMUTATION-20260916.md"
        evidence_expiry.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-EXPIRY-AUTHORITY-NOMUTATION-20260916.md", evidence_expiry)
        evidence_exact_lineage = self.root / "docs/architecture/EVIDENCE-M11-AUTHORIZATION-GATE-EXACT-LINEAGE-20260916.md"
        evidence_exact_lineage.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-AUTHORIZATION-GATE-EXACT-LINEAGE-20260916.md", evidence_exact_lineage)
        evidence_post_392 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-392-20260916.md"
        evidence_post_392.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-392-20260916.md", evidence_post_392)
        evidence_post_403 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-403-20260917.md"
        evidence_post_403.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-403-20260917.md", evidence_post_403)
        evidence_post_406 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-406-20260917.md"
        evidence_post_406.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-406-20260917.md", evidence_post_406)
        evidence_registry_parent = self.root / "docs/architecture/EVIDENCE-RP01-REGISTRY-PARENT-OPENAT-20260917.md"
        evidence_registry_parent.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-RP01-REGISTRY-PARENT-OPENAT-20260917.md", evidence_registry_parent)
        evidence_correlation = self.root / "docs/architecture/EVIDENCE-M11-CORRELATION-LINEAGE-20260917.md"
        evidence_correlation.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-CORRELATION-LINEAGE-20260917.md", evidence_correlation)
        evidence_reverse_ledger = self.root / "docs/architecture/EVIDENCE-M11-REVERSE-LEDGER-GRAPH-20260917.md"
        evidence_reverse_ledger.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-REVERSE-LEDGER-GRAPH-20260917.md", evidence_reverse_ledger)
        lineage_mutation = self.root / "scripts/mutate_m11_authorization_lineage.py"
        lineage_mutation.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "scripts/mutate_m11_authorization_lineage.py", lineage_mutation)
        lineage_mutation_evidence = self.root / "docs/architecture/EVIDENCE-M11-AUTHORIZATION-LINEAGE-MUTATION-20260916.md"
        lineage_mutation_evidence.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-M11-AUTHORIZATION-LINEAGE-MUTATION-20260916.md", lineage_mutation_evidence)
        evidence_383 = self.root / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-383-20260916.md"
        evidence_383.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-383-20260916.md", evidence_383)
        for relative in ("lab/affiliate-bot/cmd/bot/runtime_gate.go", "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"):
            source, target = ROOT / relative, self.root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
        for relative in ("lab/affiliate-bot/cmd/bot/m11_manual_stop_process_kill_test.go", "docs/architecture/EVIDENCE-M11-MANUAL-STOP-PROCESS-KILL-20260917.md"):
            source, target = ROOT / relative, self.root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source, target)
        evidence_post_merge = self.root / "docs/architecture/EVIDENCE-PR412-POST-MERGE-20260917.md"
        evidence_post_merge.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR412-POST-MERGE-20260917.md", evidence_post_merge)
        evidence_current_baseline = self.root / "docs/architecture/EVIDENCE-PR413-POST-MERGE-20260918.md"
        evidence_current_baseline.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(ROOT / "docs/architecture/EVIDENCE-PR413-POST-MERGE-20260918.md", evidence_current_baseline)
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

    def test_missing_learner_schema_identity_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("type LearnerIntent = corem08.Intent", "type LearnerIntent struct"), encoding="utf-8")
        self.assertIn("learner schema identity regression is missing", self.run_audit(False))

    def test_missing_registry_append_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/stable_append_posix.go"
        source.write_text(source.read_text(encoding="utf-8").replace("unix.Openat", "os.OpenFile"), encoding="utf-8")
        self.assertIn("registry append parent-guard is missing", self.run_audit(False))

    def test_missing_advisor_campaign_writer_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_budget.go"
        source.write_text(source.read_text(encoding="utf-8").replace("openStableRegularFileForAppendPath", "os.OpenFile"), encoding="utf-8")
        self.assertIn("advisor campaign-writer parent guard is missing", self.run_audit(False))

    def test_missing_advisor_campaign_writer_parent_swap_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_writer_parent_swap_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestAdvisorResultRejectsParentSwapBeforeOpenat", "RemovedAdvisorResultParentSwapRegression", 1), encoding="utf-8")
        self.assertIn("advisor campaign-writer parent-swap regression is missing", self.run_audit(False))

    def test_missing_registry_graph_envelope_integrity_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        original = source.read_text(encoding="utf-8")
        graph_prefix, graph_body = original.split("func ValidateArtifactGraph", 1)
        graph_body = graph_body.replace("entry.ContentHash != expected.ContentHash", "registry graph envelope guard removed", 1)
        source.write_text(graph_prefix + "func ValidateArtifactGraph" + graph_body, encoding="utf-8")
        self.assertIn("registry graph envelope-integrity guard is missing", self.run_audit(False))

    def test_missing_post_merge_pr412_evidence_is_rejected(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["recent_updates"] = [entry for entry in matrix["recent_updates"] if entry.get("id") != "RP-09-post-412-merge-ci-20260917"]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("post-merge PR #412 acceptance evidence", self.run_audit(False))

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
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("learner-bot-tests-shard-1:", "removed-learner-bot-tests-shard-1:", 1), encoding="utf-8")
        self.assertIn("deterministic CI shard/cache is missing", self.run_audit(False))

    def test_deterministic_smoke_group_removal_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("deterministic-smokes-m06-m07:", "removed-deterministic-smokes-m06-m07:", 1), encoding="utf-8")
        self.assertIn("deterministic CI shard/cache is missing", self.run_audit(False))

    def test_missing_accesstrade_pending_import_recovery_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/accesstrade_import.go"
        source.write_text(source.read_text(encoding="utf-8").replace('accesstradeJournalCleanupFailure("after_remove_before_parent_sync")', "journalCleanupAcknowledgementRemoved", 1), encoding="utf-8")
        self.assertIn("ACCESSTRADE pending-import recovery regression is missing", self.run_audit(False))

    def test_missing_campaign_result_acknowledgement_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_results.go"
        source.write_text(source.read_text(encoding="utf-8").replace("campaignResultVisible(path, candidate)", "campaignResultVisibilityRemoved", 1), encoding="utf-8")
        self.assertIn("campaign-result acknowledgement regression is missing", self.run_audit(False))

    def test_missing_committed_journal_cleanup_acknowledgement_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func removeCommittedRecoveryJournal", "func removedCommittedRecoveryJournal", 1), encoding="utf-8")
        self.assertIn("committed journal-cleanup acknowledgement regression is missing", self.run_audit(False))

    def test_missing_m11_post_remove_journal_cleanup_acknowledgement_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace('m11JournalCleanupWriteFault("unknown_after_remove_before_parent_sync")', 'm11JournalCleanupWriteFault("removed_after_remove")', 1), encoding="utf-8")
        self.assertIn("M11 post-remove journal-cleanup acknowledgement regression is missing", self.run_audit(False))

    def test_missing_m10_canary_cost_journal_path_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func readM10CostBoundJournal", "func removedM10CostBoundJournal", 1), encoding="utf-8")
        self.assertIn("M10 canary/cost journal path-guard regression is missing", self.run_audit(False))

    def test_missing_m11_recovery_admission_approval_guard_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("approval, approvalOK := approvals[x.NewApprovalID]", "approval guard removed", 1), encoding="utf-8")
        self.assertIn("M11 recovery-admission approval regression is missing", self.run_audit(False))

    def test_missing_m11_authorization_gate_exact_lineage_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("gate.HealthSnapshotID != x.ProductionHealthSnapshotID", "health lineage guard removed", 1), encoding="utf-8")
        self.assertIn("M11 authorization gate exact-lineage regression is missing", self.run_audit(False))

    def test_missing_m11_correlation_lineage_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("bound.CorrelationID != lease.CorrelationID", "correlation lineage guard removed", 1), encoding="utf-8")
        self.assertIn("M11 correlation-lineage regression is missing", self.run_audit(False))

    def test_missing_m11_reverse_ledger_graph_guard_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("production ledger reconciliation link is orphaned or mismatched", "reverse ledger graph guard removed", 1), encoding="utf-8")
        self.assertIn("M11 reverse-ledger graph regression is missing", self.run_audit(False))

    def test_missing_m11_evaluation_evidence_guard_is_rejected(self):
        source = self.root / "core/m11/artifact_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("len(x.EvidenceIDs) != 1 || x.EvidenceIDs[0] != x.OutcomeID", "evaluation evidence guard removed", 1), encoding="utf-8")
        self.assertIn("M11 reverse-ledger graph regression is missing", self.run_audit(False))

    def test_missing_m11_manual_stop_process_kill_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_manual_stop_process_kill_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestM11ManualStopProcessKillRequiresLockedRecovery", "TestM11ManualStopProcessKillCoverageRemoved"), encoding="utf-8")
        self.assertIn("M11 direct STOP process-kill regression is missing", self.run_audit(False))

    def test_missing_m11_authorization_lineage_mutation_is_rejected(self):
        source = self.root / "scripts/mutate_m11_authorization_lineage.py"
        source.write_text(source.read_text(encoding="utf-8").replace("LINEAGE_GUARDS", "REMOVED_GUARD_SET"), encoding="utf-8")
        self.assertIn("M11 authorization lineage mutation proof is missing", self.run_audit(False))

    def test_missing_m11_fixture_outcome_execution_cardinality_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        prefix, marker, suffix = source.read_text(encoding="utf-8").partition("func loadM11FixtureOutcomes")
        source.write_text(prefix + marker + suffix.replace("|| seenExecution[outcome.EffectRef.EffectID] {", "", 1), encoding="utf-8")
        self.assertIn("M11 fixture-outcome execution cardinality regression is missing", self.run_audit(False))

    def test_missing_m11_recovery_handoff_strict_decoder_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry.go"
        source.write_text(source.read_text(encoding="utf-8").replace("contracts.Decode(raw)", "json.Unmarshal(raw, &value)", 1), encoding="utf-8")
        self.assertIn("M11 recovery-handoff strict decoder regression is missing", self.run_audit(False))

    def test_missing_m06_history_handoff_strict_decoder_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/watcher.go"
        source.write_text(source.read_text(encoding="utf-8").replace("record, err := decodeHistoryHandoffRecord(body)", "var record HistoryRecord\n\t\terr := json.Unmarshal(body, &record)", 1), encoding="utf-8")
        self.assertIn("M06 history-handoff strict decoder regression is missing", self.run_audit(False))

    def test_missing_m06_adapter_visible_append_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/watcher.go"
        prefix, marker, suffix = source.read_text(encoding="utf-8").partition("func m06AdapterHandler(historyPath string) http.HandlerFunc {")
        source.write_text(prefix + marker + suffix.replace("if err != nil && !isPublishedAppendUncertainty(err)", "if err != nil", 1), encoding="utf-8")
        self.assertIn("M06 adapter visible-append regression is missing", self.run_audit(False))

    def test_missing_m06_fixture_import_visible_append_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/watcher.go"
        prefix, marker, suffix = source.read_text(encoding="utf-8").partition("func runWatcher(args []string, stdout, stderr io.Writer) int {")
        fixture, delimiter, rest = suffix.partition("// m06AdapterRequest carries")
        source.write_text(prefix + marker + fixture.replace("if err != nil && !isPublishedAppendUncertainty(err)", "if err != nil", 1) + delimiter + rest, encoding="utf-8")
        self.assertIn("M06 adapter visible-append regression is missing", self.run_audit(False))

    def test_missing_m06_pinned_fetch_visible_append_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/watcher_fetch.go"
        source.write_text(source.read_text(encoding="utf-8").replace("if err != nil && !isPublishedAppendUncertainty(err)", "if err != nil", 1), encoding="utf-8")
        self.assertIn("M06 adapter visible-append regression is missing", self.run_audit(False))

    def test_missing_m07_visible_artifact_recovery_disclosure_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/watcher.go"
        source.write_text(source.read_text(encoding="utf-8").replace('response["artifact_id"] = artifactID', 'response["artifact_id"] = "removed"', 1), encoding="utf-8")
        self.assertIn("immutable artifact post-publish regression is missing", self.run_audit(False))

    def test_missing_advisor_fixture_output_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
        source.write_text(source.read_text(encoding="utf-8").replace("fixture output parent must be an existing non-symlink directory", "fixture output parent guard removed", 1), encoding="utf-8")
        self.assertIn("advisor fixture output-parent regression is missing", self.run_audit(False))

    def test_missing_backup_restore_missing_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func ensureOutputParentBeforeCreate", "func removedOutputParentPreflight", 1), encoding="utf-8")
        self.assertIn("backup/restore missing-parent regression is missing", self.run_audit(False))

    def test_missing_backup_target_absence_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("backup target appeared while staging", "backup target appearance ignored", 1), encoding="utf-8")
        self.assertIn("backup target-absence guard regression is missing", self.run_audit(False))

    def test_missing_backup_process_exit_lock_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
        source.write_text(source.read_text(encoding="utf-8").replace("syscall.Flock", "removedFlock"), encoding="utf-8")
        self.assertIn("backup process-exit lock regression is missing", self.run_audit(False))

    def test_missing_backup_process_exit_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds", "RemovedBackupProcessExitRegression"), encoding="utf-8")
        self.assertIn("backup process-exit lock regression is missing", self.run_audit(False))

    def test_missing_backup_process_kill_regression_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestBackupProcessKillBeforePublishLeavesNoTargetAndRetrySucceeds", "RemovedBackupProcessKillRegression"), encoding="utf-8")
        self.assertIn("backup process-exit lock regression is missing", self.run_audit(False))

    def test_missing_managed_lock_path_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
        source.write_text(source.read_text(encoding="utf-8").replace("unix.O_NOFOLLOW", "removedNoFollow"), encoding="utf-8")
        self.assertIn("managed lock path-guard regression is missing", self.run_audit(False))

    def test_missing_managed_lock_parent_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
        source.write_text(source.read_text(encoding="utf-8").replace("validateManagedLockParent", "removedManagedLockParent"), encoding="utf-8")
        self.assertIn("managed lock path-guard regression is missing", self.run_audit(False))

    def test_missing_managed_lock_openat_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
        source.write_text(source.read_text(encoding="utf-8").replace("unix.Openat", "removedOpenat"), encoding="utf-8")
        self.assertIn("managed lock path-guard regression is missing", self.run_audit(False))

    def test_missing_m11_process_kill_journal_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestM11ProcessKillAfterJournalBeforeLedgerAppendLeavesJournalForFreshRecovery", "RemovedM11ProcessKillRegression"), encoding="utf-8")
        self.assertIn("M11 process-kill journal regression is missing", self.run_audit(False))

    def test_missing_m11_outcome_process_kill_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend", "func RemovedM11OutcomeProcessKillRegression", 1), encoding="utf-8")
        self.assertIn("M11 outcome process-kill regression is missing", self.run_audit(False))

    def test_missing_m11_outcome_post_append_process_kill_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func TestM11OutcomeProcessKillAfterOutcomeAppend", "func RemovedM11OutcomePostAppendProcessKillRegression", 1), encoding="utf-8")
        self.assertIn("M11 outcome process-kill regression is missing", self.run_audit(False))

    def test_missing_m11_process_kill_after_ledger_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace('phase == "after_sync"', 'phase == "removed_after_sync"', 1), encoding="utf-8")
        self.assertIn("M11 post-ledger process-kill regression is missing", self.run_audit(False))

    def test_missing_m10_process_kill_after_canonical_append_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay", "RemovedM10CanonicalAppendProcessKillRegression"), encoding="utf-8")
        self.assertIn("M10 post-canonical-append process-kill regression is missing", self.run_audit(False))

    def test_missing_m10_execution_process_kill_after_canonical_append_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay", "RemovedM10ExecutionCanonicalAppendProcessKillRegression"), encoding="utf-8")
        self.assertIn("M10 execution post-canonical-append process-kill regression is missing", self.run_audit(False))

    def test_missing_backup_orphan_staging_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("func cleanupStaleStaging", "func removedStaleStaging", 1), encoding="utf-8")
        self.assertIn("backup orphan-staging cleanup regression is missing", self.run_audit(False))

    def test_missing_backup_post_publish_process_kill_guard_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
        source.write_text(source.read_text(encoding="utf-8").replace("TestBackupProcessKillAfterPublishLeavesVisibleCompleteTarget", "RemovedBackupPostPublishProcessKillRegression", 1), encoding="utf-8")
        self.assertIn("backup post-publish process-kill regression is missing", self.run_audit(False))

    def test_missing_immutable_artifact_parent_recheck_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/mission_command.go"
        source.write_text(source.read_text(encoding="utf-8").replace("if err := requireArtifactOutputDirectory(dir); err != nil {\n\t\treturn \"\", err\n\t}\n\tif err := os.Link(temporary, path);", "if err := os.Link(temporary, path);", 1), encoding="utf-8")
        self.assertIn("immutable artifact parent-recheck regression is missing", self.run_audit(False))

    def test_missing_advisor_fixture_parent_recheck_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
        content = source.read_text(encoding="utf-8")
        marker = 'if err := requireAdvisorFixtureOutputParent(parent); err != nil {\n\t\treturn emit("PATH_ERROR", "", err, 1)\n\t}'
        first, separator, remainder = content.partition(marker)
        self.assertTrue(separator)
        source.write_text(first + separator + remainder.replace(marker, "", 1), encoding="utf-8")
        self.assertIn("advisor fixture parent-recheck regression is missing", self.run_audit(False))

    def test_missing_advisor_fixture_failed_staging_cleanup_is_rejected(self):
        source = self.root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
        source.write_text(source.read_text(encoding="utf-8").replace("var advisorFixtureBuild = buildBR10AdvisorFixture", "var removedAdvisorFixtureBuild = buildBR10AdvisorFixture", 1), encoding="utf-8")
        self.assertIn("advisor fixture failed-staging cleanup regression is missing", self.run_audit(False))

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

    def test_missing_registry_graph_envelope_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_registry_graph_envelope_integrity.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("registry graph envelope mutation proof is missing or not wired to CI", self.run_audit(False))

    def test_missing_m11_recovery_admission_field_integrity_is_rejected(self):
        source = self.root / "core/m11/artifact.go"
        source.write_text(source.read_text(encoding="utf-8").replace("anyBlank(recoveryFields)", "false", 1), encoding="utf-8")
        self.assertIn("M11 recovery-admission field-integrity regression is missing", self.run_audit(False))

    def test_missing_backup_source_mutation_proof_is_rejected(self):
        workflow = self.root / ".github/workflows/curriculum-ci.yml"
        workflow.write_text(workflow.read_text(encoding="utf-8").replace("python scripts/mutate_backup_source_guard.py", "python scripts/removed.py"), encoding="utf-8")
        self.assertIn("required regression is not wired", self.run_audit(False))

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

    def test_selected_source_operated_record_is_required(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["recent_updates"] = [entry for entry in matrix["recent_updates"] if entry.get("id") != "RP-10-m06-selected-source-local-operated-run"]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("selected-source operated run record lacks", self.run_audit(False))

    def test_local_recovery_drill_record_is_required(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["recent_updates"] = [entry for entry in matrix["recent_updates"] if entry.get("id") != "RP-10-br18b-local-recovery-drill"]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("local recovery drill record lacks", self.run_audit(False))

    def test_assisted_fresh_workspace_record_is_required(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        matrix["recent_updates"] = [entry for entry in matrix["recent_updates"] if entry.get("id") != "RP-10-br16b-assisted-fresh-workspace"]
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("assisted fresh-workspace record lacks", self.run_audit(False))

    def test_selected_source_forged_commission_scope_is_required(self):
        matrix_path = self.root / "docs/plans/READINESS-MATRIX.json"
        matrix = json.loads(matrix_path.read_text(encoding="utf-8"))
        br15 = next(item for item in matrix["criteria"] if item["id"] == "BR-15")
        br15["missing_evidence"][0] = "tests: selected campaign metadata reaches M07"
        matrix_path.write_text(json.dumps(matrix), encoding="utf-8")
        self.assertIn("does not disclose selected-source forged-commission", self.run_audit(False))

    def test_m07_raw_json_transport_guard_is_required(self):
        runner = self.root / "scripts/run_n8n_engine_regression.py"
        runner.write_text(
            runner.read_text(encoding="utf-8").replace(
                "tool_result_text:$('M07 Adapter Input').item.json.tool_request_json",
                "tool_result:JSON.parse($('M07 Adapter Input').item.json.tool_request_json)",
                1,
            ),
            encoding="utf-8",
        )
        self.assertIn("M07 raw JSON transport regression is missing", self.run_audit(False))

    def test_beginner_plan_selected_source_boundary_is_required(self):
        plan = self.root / "docs/plans/BEGINNER-READINESS-PLAN.md"
        plan.write_text(plan.read_text(encoding="utf-8").replace("sanitized/read-only", "unscoped", 1), encoding="utf-8")
        self.assertIn("beginner plan does not disclose the selected-source offline boundary", self.run_audit(False))


if __name__ == "__main__":
    unittest.main()
