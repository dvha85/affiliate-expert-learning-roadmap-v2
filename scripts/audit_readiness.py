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
    "scripts/mutate_m07_portable_input_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_general_portable_input_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_internal_portable_input_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m07_backup_sidecar_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m08_m07_proposal_path_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_mission_portable_input_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_mission_stop_immutability_guard.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m07_strict_output_decoder.py": ".github/workflows/curriculum-ci.yml",
    "scripts/mutate_m07_registry_strict_decoder.py": ".github/workflows/curriculum-ci.yml",
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


def audit_runtime_acceptance(root, matrix):
    """Prevent a summary row from claiming a process-barrier proof it cannot point to."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    barrier = updates.get("RP-03-concurrent-reservation-barrier")
    if not isinstance(barrier, dict):
        fail("matrix lacks the R10 process-barrier acceptance record")
    if "lab/affiliate-bot/cmd/bot/mission_command_test.go" not in barrier.get("test_refs", []):
        fail("R10 process-barrier record lacks its real Bot regression")
    scope = barrier.get("scope")
    if not isinstance(scope, str) or "24-process" not in scope or "cap=1" not in scope:
        fail("R10 process-barrier record lacks a bounded cap=1 disclosure")
    source = root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
    if not source.is_file() or "TestMissionM10ReservationCapOneAcrossTwentyFourBotProcesses" not in source.read_text(encoding="utf-8"):
        fail("R10 process-barrier regression is missing from the learner Bot test path")
    commit_fault = updates.get("RP-03-m10-reservation-commit-fault")
    if not isinstance(commit_fault, dict):
        fail("matrix lacks the R10 reservation commit-fault acceptance record")
    if "lab/affiliate-bot/cmd/bot/mission_command_test.go" not in commit_fault.get("test_refs", []):
        fail("R10 reservation commit-fault record lacks its real Bot regression")
    fault_scope = commit_fault.get("scope")
    if not isinstance(fault_scope, str) or "STORE_ERROR" not in fault_scope or "cap=1" not in fault_scope or "after temporary-file sync" not in fault_scope:
        fail("R10 reservation commit-fault record lacks a bounded cap disclosure")
    source_text = source.read_text(encoding="utf-8")
    # Keep implementation and command-path assertions separate: a test name by
    # itself must not satisfy an acceptance rule when the uncertain-publish
    # payload boundary was removed from the Bot.
    mission_source = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    mission_source_text = mission_source.read_text(encoding="utf-8") if mission_source.is_file() else ""
    if "TestMissionM10ReservationCommitFaultDoesNotConsumeCap" not in source_text or '"after_temp_sync", "before_rename"' not in source_text:
        fail("R10 reservation commit-fault regression is missing from the learner Bot test path")
    publish_fault = updates.get("RP-03-mission-state-publish-recovery-status")
    if not isinstance(publish_fault, dict):
        fail("matrix lacks the mission-state post-rename acceptance record")
    if "lab/affiliate-bot/cmd/bot/mission_command_test.go" not in publish_fault.get("test_refs", []):
        fail("mission-state post-rename record lacks its real Bot regression")
    publish_scope = publish_fault.get("scope")
    if not isinstance(publish_scope, str) or "PUBLISHED_RECOVERY_REQUIRED" not in publish_scope or "EXACT_DUPLICATE" not in publish_scope:
        fail("mission-state post-rename record lacks recovery status disclosure")
    if "TestMissionM10ReservationPostRenameSyncFaultRequiresRecoveryInsteadOfRetry" not in source_text or '"after_rename_before_parent_sync"' not in source_text:
        fail("mission-state post-rename regression is missing from the learner Bot test path")
    stop_source = root / "lab/affiliate-bot/cmd/bot/mission_state_fault_test.go"
    stop_text = stop_source.read_text(encoding="utf-8") if stop_source.is_file() else ""
    if "TestMissionStopPostRenameSyncFaultKeepsDurableStop" not in stop_text or '"PUBLISHED_RECOVERY_REQUIRED"' not in stop_text or '"state", "mission-state.json", 1, false' not in stop_text or '"marker", "STOP", 2, true' not in stop_text:
        fail("durable STOP post-rename regression is missing from the learner Bot test path")
    if "TestMissionStopCannotOverwriteDurableReason" not in stop_text or '"replacement-stop"' not in stop_text or "!bytes.Equal(stateBefore, stateAfter)" not in stop_text or "!bytes.Equal(markerBefore, markerAfter)" not in stop_text:
        fail("durable STOP immutability regression is missing from the learner Bot test path")
    artifact_publish = updates.get("RP-01-atomic-artifact-publish")
    if not isinstance(artifact_publish, dict):
        fail("matrix lacks immutable artifact post-publish acceptance record")
    expected_refs = {
        "lab/affiliate-bot/cmd/bot/artifact_publish_test.go",
        "lab/affiliate-bot/cmd/bot/mission_command_test.go",
        "lab/affiliate-bot/cmd/bot/m07_test.go",
        "lab/affiliate-bot/cmd/bot/watcher_test.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
    }
    if not expected_refs.issubset(set(artifact_publish.get("test_refs", []))):
        fail("immutable artifact post-publish record lacks real publisher/CLI/adapter regressions")
    artifact_scope = artifact_publish.get("scope")
    if not isinstance(artifact_scope, str) or "PUBLISHED_RECOVERY_REQUIRED" not in artifact_scope or "after Link" not in artifact_scope:
        fail("immutable artifact post-publish record lacks visible-artifact uncertainty disclosure")
    artifact_source = root / "lab/affiliate-bot/cmd/bot/artifact_publish_test.go"
    watcher_source = root / "lab/affiliate-bot/cmd/bot/watcher_test.go"
    artifact_text = artifact_source.read_text(encoding="utf-8") if artifact_source.is_file() else ""
    watcher_text = watcher_source.read_text(encoding="utf-8") if watcher_source.is_file() else ""
    m07_source = root / "lab/affiliate-bot/cmd/bot/m07_test.go"
    m07_text = m07_source.read_text(encoding="utf-8") if m07_source.is_file() else ""
    m11_fault_source = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    m11_fault_text = m11_fault_source.read_text(encoding="utf-8") if m11_fault_source.is_file() else ""
    if "TestWriteNewJSONReportsVisibleArtifactWhenParentSyncIsUnconfirmed" not in artifact_text or "TestMissionM08IntentReportsUnconfirmedVisibleArtifact" not in source_text or "TestMissionM08PolicyReportsUnconfirmedVisibleArtifact" not in source_text or "TestM07CLIDisclosesUnconfirmedVisibleArtifact" not in m07_text or "TestM07HTTPAdapterReportsUnconfirmedVisibleToolArtifact" not in watcher_text or "recovery handoff uncertainty was not disclosed" not in m11_fault_text or '"m11-recovery-export"' not in m11_fault_text:
        fail("immutable artifact post-publish regression is missing from a real publisher, CLI, or adapter path")
    canonical_output = updates.get("RP-03-m10-canonical-output-disclosure")
    if not isinstance(canonical_output, dict):
        fail("matrix lacks M10 canonical-output disclosure acceptance record")
    if "lab/affiliate-bot/cmd/bot/mission_command.go" not in canonical_output.get("implementation_refs", []) or "lab/affiliate-bot/cmd/bot/mission_command_test.go" not in canonical_output.get("test_refs", []):
        fail("M10 canonical-output disclosure record lacks implementation/test refs")
    canonical_output_scope = canonical_output.get("scope")
    if not isinstance(canonical_output_scope, str) or "CANONICAL_ARTIFACT_REGISTERED_OUTPUT_UNAVAILABLE" not in canonical_output_scope or "PUBLISHED_RECOVERY_REQUIRED" not in canonical_output_scope or "m10-resolve" not in canonical_output_scope:
        fail("M10 canonical-output disclosure record lacks bounded recovery disclosure")
    if "TestMissionM10DisclosesCanonicalArtifactWhenPortableOutputConflicts" not in source_text or "CANONICAL_ARTIFACT_REGISTERED_OUTPUT_UNAVAILABLE" not in source_text or '"m10-resolve"' not in source_text:
        fail("M10 canonical-output disclosure regression is missing from learner Bot path")
    gate_identity = updates.get("RP-03-m10-gate-evaluation-identity")
    if not isinstance(gate_identity, dict):
        fail("matrix lacks M10 gate evaluation-identity acceptance record")
    required_gate_identity_refs = {
        "core/m10/canary_gate.go",
        "core/m10/artifact_registry.go",
        "core/m10/cost_bound_test.go",
    }
    if not required_gate_identity_refs.issubset(set(gate_identity.get("implementation_refs", [])) | set(gate_identity.get("test_refs", []))):
        fail("M10 gate evaluation-identity record lacks implementation/test refs")
    gate_identity_scope = gate_identity.get("scope")
    if not isinstance(gate_identity_scope, str) or "evaluation time" not in gate_identity_scope or "distinct immutable IDs" not in gate_identity_scope:
        fail("M10 gate evaluation-identity record lacks bounded identity disclosure")
    gate_source = root / "core/m10/canary_gate.go"
    gate_registry_source = root / "core/m10/artifact_registry.go"
    gate_test = root / "core/m10/cost_bound_test.go"
    gate_source_text = gate_source.read_text(encoding="utf-8") if gate_source.is_file() else ""
    gate_registry_text = gate_registry_source.read_text(encoding="utf-8") if gate_registry_source.is_file() else ""
    gate_test_text = gate_test.read_text(encoding="utf-8") if gate_test.is_file() else ""
    if "in.Now, in.Ledger" not in gate_source_text or "Now: gate.EvaluatedAt" not in gate_registry_text or "TestCanaryGateIdentityIncludesEvaluationTime" not in gate_test_text or "TestMissionM10RegistersDistinctGatesForDistinctEvaluationTimes" not in source_text:
        fail("M10 gate evaluation-identity regression is missing from canonical path")
    m11_authorization_identity = updates.get("RP-07-m11-authorization-time-identity")
    if not isinstance(m11_authorization_identity, dict):
        fail("matrix lacks M11 authorization time-identity acceptance record")
    required_m11_authorization_refs = {
        "core/m11/artifact.go",
        "core/m11/artifact_registry.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "core/m11/artifact_test.go",
        "scripts/smoke_br18b_backup_restore.py",
        "scripts/mutate_m11_identity_guard.py",
    }
    if not required_m11_authorization_refs.issubset(set(m11_authorization_identity.get("implementation_refs", [])) | set(m11_authorization_identity.get("test_refs", []))):
        fail("M11 authorization time-identity record lacks implementation/test refs")
    m11_authorization_scope = m11_authorization_identity.get("scope")
    if not isinstance(m11_authorization_scope, str) or "authorized_at" not in m11_authorization_scope or "distinct IDs" not in m11_authorization_scope:
        fail("M11 authorization time-identity record lacks bounded identity disclosure")
    m11_artifact_source = root / "core/m11/artifact.go"
    m11_registry_source = root / "core/m11/artifact_registry.go"
    m11_artifact_test = root / "core/m11/artifact_test.go"
    m11_smoke = root / "scripts/smoke_br18b_backup_restore.py"
    m11_artifact_text = m11_artifact_source.read_text(encoding="utf-8") if m11_artifact_source.is_file() else ""
    m11_registry_text = m11_registry_source.read_text(encoding="utf-8") if m11_registry_source.is_file() else ""
    m11_test_text = m11_artifact_test.read_text(encoding="utf-8") if m11_artifact_test.is_file() else ""
    m11_smoke_text = m11_smoke.read_text(encoding="utf-8") if m11_smoke.is_file() else ""
    if "gateID, executorID, authorizedAt string" not in m11_artifact_text or "gate.GateID, x.ExecutorID, x.AuthorizedAt" not in m11_registry_text or "distinct authorization times reused an immutable ID" not in m11_test_text or "same_gate_later_auth" not in m11_smoke_text:
        fail("M11 authorization time-identity regression is missing from canonical path")
    registry_publish = updates.get("RP-03-07-registry-publish-uncertainty")
    if not isinstance(registry_publish, dict):
        fail("matrix lacks M10/M11 registry publish-uncertainty acceptance record")
    required_registry_publish_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "lab/affiliate-bot/cmd/bot/mission_command_test.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
    }
    if not required_registry_publish_refs.issubset(set(registry_publish.get("implementation_refs", [])) | set(registry_publish.get("test_refs", []))):
        fail("M10/M11 registry publish-uncertainty record lacks implementation/test refs")
    registry_publish_scope = registry_publish.get("scope")
    if not isinstance(registry_publish_scope, str) or "PUBLISHED_RECOVERY_REQUIRED" not in registry_publish_scope or "exact-retries" not in registry_publish_scope:
        fail("M10/M11 registry publish-uncertainty record lacks bounded recovery disclosure")
    m11_fault_source = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    m11_fault_text = m11_fault_source.read_text(encoding="utf-8") if m11_fault_source.is_file() else ""
    if "artifactRegistryPublishFailure" not in source_text or "TestMissionM10GateDisclosesRegistryPublishUncertainty" not in source_text or "TestMissionM10AuthorizationDisclosesRegistryPublishUncertainty" not in source_text or "TestMissionM11LifecycleDisclosesRegistryPublishUncertainty" not in source_text or "TestMissionM11GateAuthorizationAndReservationDiscloseRegistryPublishUncertainty" not in source_text or "TestMissionM11EvaluationDisclosesRegistryPublishUncertainty" not in source_text or "TestMissionM11CycleDisclosesRegistryPublishUncertainty" not in source_text or '"m11-activate"' not in source_text or '"m11-ledger-init"' not in source_text or '"m11-gate"' not in source_text or '"m11-authorize"' not in source_text or '"m11-reserve-authorization"' not in source_text or '"m11-evaluate"' not in source_text or '"m11-close-cycle"' not in source_text or "uncertain M11 registry append was not retained" not in m11_fault_text or "TestMissionM11ReconcileDisclosesRegistryPublishUncertainty" not in m11_fault_text or '"m11-reconcile"' not in m11_fault_text or "recovery admission uncertainty was not disclosed" not in m11_fault_text or '"m11-recovery-admit"' not in m11_fault_text:
        fail("M10/M11 registry publish-uncertainty regression is missing from learner paths")
    visible_journal_ids = {"RP-03-m10-cost-bound-two-store-journal", "RP-03-m10-canary-registry-state-journal"}
    visible_journals = {item_id: updates.get(item_id) for item_id in visible_journal_ids}
    if any(not isinstance(item, dict) for item in visible_journals.values()):
        fail("matrix lacks M10 visible-journal recovery acceptance records")
    if any("PUBLISHED_RECOVERY_REQUIRED" not in item.get("scope", "") for item in visible_journals.values()):
        fail("M10 visible-journal records lack publish-uncertainty disclosure")
    if "TestMissionM10VisibleJournalPublishUncertaintyDefersLockedReplay" not in source_text or "TestMissionM10ExecutionJournalVisiblePublishUncertaintyDefersLockedReplay" not in source_text or '"after_publish_before_parent_sync"' not in source_text or '"after_rename_before_parent_sync"' not in source_text or '"m10-canary"' not in source_text or '"m10-cost-register"' not in source_text or '"m10-record-failed"' not in source_text:
        fail("M10 visible-journal recovery regression is missing from learner path")
    m11_outcome_journal = updates.get("RP-07-m11-outcome-visible-journal")
    if not isinstance(m11_outcome_journal, dict) or "PUBLISHED_RECOVERY_REQUIRED" not in m11_outcome_journal.get("scope", ""):
        fail("matrix lacks M11 outcome visible-journal recovery acceptance")
    if "TestMissionM11OutcomeJournalVisiblePublishUncertaintyDefersLockedReplay" not in source_text or '"m11-outcome"' not in source_text or "artifactIfAtomicPublishUncertain" not in mission_source_text:
        fail("M11 outcome visible-journal recovery regression is missing from learner path")
    m11_execution_journals = updates.get("RP-07-m11-execution-visible-journals")
    if not isinstance(m11_execution_journals, dict) or "PUBLISHED_RECOVERY_REQUIRED" not in m11_execution_journals.get("scope", ""):
        fail("matrix lacks M11 execution visible-journal recovery acceptance")
    if "TestMissionM11ExecutionJournalsVisiblePublishUncertaintyDefersLockedReplay" not in source_text or '"m11-record-failed"' not in source_text or '"m11-record-unknown"' not in source_text or "artifactIfAtomicPublishUncertain" not in mission_source_text:
        fail("M11 execution visible-journal recovery regression is missing from learner path")
    framing = updates.get("RP-03-canonical-jsonl-framing")
    if not isinstance(framing, dict):
        fail("matrix lacks the canonical JSONL framing acceptance record")
    required_framing_refs = {
        "lab/affiliate-bot/internal/store/history.go",
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "lab/affiliate-bot/cmd/bot/accesstrade_receipt.go",
        "lab/affiliate-bot/internal/store/history_test.go",
        "lab/affiliate-bot/cmd/bot/action_store_test.go",
        "lab/affiliate-bot/cmd/bot/outcome_store_test.go",
        "lab/affiliate-bot/cmd/bot/mission_command_test.go",
        "lab/affiliate-bot/cmd/bot/accesstrade_import_test.go",
    }
    if not required_framing_refs.issubset(set(framing.get("implementation_refs", [])) | set(framing.get("test_refs", []))):
        fail("canonical JSONL framing record lacks implementation/test refs")
    framing_scope = framing.get("scope")
    if not isinstance(framing_scope, str) or "without LF" not in framing_scope or "M10 artifact/cost-bound" not in framing_scope or "M11 artifact" not in framing_scope or "without changing" not in framing_scope:
        fail("canonical JSONL framing record lacks bounded scope disclosure")
    store_source = root / "lab/affiliate-bot/internal/store/history.go"
    store_test = root / "lab/affiliate-bot/internal/store/history_test.go"
    framing_source = store_source.read_text(encoding="utf-8") if store_source.is_file() else ""
    framing_test = store_test.read_text(encoding="utf-8") if store_test.is_file() else ""
    if "RequireCompleteJSONLFraming" not in framing_source or "incomplete final line framing" not in framing_source or "TestJSONLOpenRejectsIncompleteFinalLine" not in framing_test:
        fail("canonical JSONL shared framing regression is missing")
    mission_source = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    mission_text = mission_source.read_text(encoding="utf-8") if mission_source.is_file() else ""
    if mission_text.count("RequireCompleteJSONLFraming(raw)") < 4 or "TestFixtureOutcomeLoadersRejectIncompleteJSONLFraming" not in source_text:
        fail("M10/M11 canonical registry/outcome framing regression is missing")
    m11_source = root / "lab/affiliate-bot/cmd/bot/m11_registry.go"
    m11_text = m11_source.read_text(encoding="utf-8") if m11_source.is_file() else ""
    receipt_source = root / "lab/affiliate-bot/cmd/bot/accesstrade_receipt.go"
    receipt_text = receipt_source.read_text(encoding="utf-8") if receipt_source.is_file() else ""
    receipt_test = root / "lab/affiliate-bot/cmd/bot/accesstrade_import_test.go"
    if "RequireCompleteJSONLFraming(raw)" not in m11_text or "RequireCompleteJSONLFraming(raw)" not in receipt_text or not receipt_test.is_file() or "TestAccesstradeBackupRequirementRejectsIncompleteOutcomeJSONL" not in receipt_test.read_text(encoding="utf-8"):
        fail("M11/ACCESSTRADE canonical JSONL framing regression is missing")
    action_test = root / "lab/affiliate-bot/cmd/bot/action_store_test.go"
    outcome_test = root / "lab/affiliate-bot/cmd/bot/outcome_store_test.go"
    if not action_test.is_file() or "unterminated action store was changed" not in action_test.read_text(encoding="utf-8"):
        fail("M03 incomplete JSONL no-mutation regression is missing")
    if not outcome_test.is_file() or "unterminated outcome store was changed" not in outcome_test.read_text(encoding="utf-8"):
        fail("M04 incomplete JSONL no-mutation regression is missing")


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


def audit_selected_source_disclosure(root, criteria_by_id, plan_text):
    """Keep selected-source engine coverage scoped when its real runner exists."""
    runner = root / "scripts/run_n8n_engine_regression.py"
    if not runner.is_file():
        return
    runner_text = runner.read_text(encoding="utf-8")
    if "M06_SELECTED_SOURCE_BLUEPRINT" in runner_text:
        marker = "M06 selected-source engine CI"
        if marker not in plan_text:
            fail("selected-source M06 engine runner lacks a scoped plan marker")
        for criterion_id in ("BR-13", "BR-14"):
            scope = " ".join(criteria_by_id[criterion_id]["missing_evidence"])
            if "selected-source" not in scope or "n8n" not in scope.casefold():
                fail(f"{criterion_id} does not disclose selected-source n8n coverage")
    if "model-forged-commission" in runner_text:
        marker = "M07 selected-source engine grounding CI"
        if marker not in plan_text:
            fail("selected-source M07 grounding runner lacks a scoped plan marker")
        scope = " ".join(criteria_by_id["BR-15"]["missing_evidence"])
        if "selected campaign sanitized fixture" not in scope or "commission_rate:0.9" not in scope:
            fail("BR-15 does not disclose selected-source forged-commission grounding coverage")
    beginner_plan = root / "docs/plans/BEGINNER-READINESS-PLAN.md"
    if beginner_plan.is_file() and "AccesstradeShopeeSmartlinkURL" in (root / "core/m06/accesstrade_shopee.go").read_text(encoding="utf-8"):
        beginner_text = beginner_plan.read_text(encoding="utf-8")
        required = ("ACCESSTRADE Shopee Smartlink", "sanitized/read-only", "operated selected-source")
        if any(marker not in beginner_text for marker in required):
            fail("beginner plan does not disclose the selected-source offline boundary")
        if "Provider diversity và selected source còn mở" in beginner_text:
            fail("beginner plan incorrectly treats the selected-source contract as absent")


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
    audit_selected_source_disclosure(root, criteria_by_id, plan_text)
    audit_review_findings(matrix, criteria_by_id, plan_text)
    audit_runtime_acceptance(root, matrix)
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
