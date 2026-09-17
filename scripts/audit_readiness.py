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
    "scripts/mutate_m11_authorization_lineage.py": ".github/workflows/curriculum-ci.yml",
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
    backup_parent = updates.get("RP-01-backup-restore-missing-parent-guard")
    if not isinstance(backup_parent, dict):
        fail("matrix lacks backup/restore missing-parent guard acceptance record")
    backup_parent_refs = set(backup_parent.get("test_refs", []))
    if {"lab/affiliate-bot/cmd/bot/backup_command_test.go", "scripts/audit_readiness.py", "scripts/tests/test_audit_readiness.py"} - backup_parent_refs:
        fail("backup/restore missing-parent record lacks CLI and audit regressions")
    backup_parent_scope = backup_parent.get("scope")
    if not isinstance(backup_parent_scope, str) or "MkdirAll" not in backup_parent_scope or "external target remains empty" not in backup_parent_scope:
        fail("backup/restore missing-parent record lacks bounded external-write disclosure")
    backup_parent_source = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    backup_parent_text = backup_parent_source.read_text(encoding="utf-8") if backup_parent_source.is_file() else ""
    backup_parent_test = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    backup_parent_test_text = backup_parent_test.read_text(encoding="utf-8") if backup_parent_test.is_file() else ""
    if "func ensureOutputParentBeforeCreate" not in backup_parent_text or "ensureOutputParentBeforeCreate(parent)" not in backup_parent_text or "TestBackupRestoreRejectMissingParentSymlinkBeforeExternalCreate" not in backup_parent_test_text or "backup created external output before rejection" not in backup_parent_test_text or "restore created external output before rejection" not in backup_parent_test_text:
        fail("backup/restore missing-parent regression is missing from the real CLI path")
    artifact_parent = updates.get("RP-01-immutable-artifact-missing-parent-guard")
    if not isinstance(artifact_parent, dict):
        fail("matrix lacks immutable artifact missing-parent guard acceptance record")
    artifact_parent_refs = set(artifact_parent.get("test_refs", []))
    if {"lab/affiliate-bot/cmd/bot/artifact_publish_test.go", "scripts/audit_readiness.py", "scripts/tests/test_audit_readiness.py"} - artifact_parent_refs:
        fail("immutable artifact missing-parent record lacks publisher and audit regressions")
    artifact_parent_scope = artifact_parent.get("scope")
    if not isinstance(artifact_parent_scope, str) or "MkdirAll" not in artifact_parent_scope or "external target remains empty" not in artifact_parent_scope:
        fail("immutable artifact missing-parent record lacks bounded external-write disclosure")
    mission_source = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    mission_text = mission_source.read_text(encoding="utf-8") if mission_source.is_file() else ""
    artifact_test = root / "lab/affiliate-bot/cmd/bot/artifact_publish_test.go"
    artifact_test_text = artifact_test.read_text(encoding="utf-8") if artifact_test.is_file() else ""
    if "func artifactOutputDirectory(path string) (string, error) {\n\tdir := filepath.Dir(path)\n\tif err := ensureOutputParentBeforeCreate(dir);" not in mission_text or "TestWriteNewJSONRejectsMissingParentSymlinkBeforeExternalCreate" not in artifact_test_text or "artifact publisher created external output before rejection" not in artifact_test_text:
        fail("immutable artifact missing-parent regression is missing from the real publisher path")
    runtime_parent = updates.get("RP-01-runtime-missing-parent-guard")
    if not isinstance(runtime_parent, dict):
        fail("matrix lacks runtime missing-parent guard acceptance record")
    runtime_parent_scope = runtime_parent.get("scope")
    if not isinstance(runtime_parent_scope, str) or "STORE_ERROR" not in runtime_parent_scope or "empty external target" not in runtime_parent_scope:
        fail("runtime missing-parent record lacks bounded external-write disclosure")
    runtime_test = root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
    runtime_test_text = runtime_test.read_text(encoding="utf-8") if runtime_test.is_file() else ""
    if "func ensureRuntimeDirectory(dir string) error {\n\tif err := ensureOutputParentBeforeCreate(dir);" not in mission_text or "TestMissionInitRejectsMissingParentSymlinkBeforeExternalCreate" not in runtime_test_text or "mission init created external runtime before rejection" not in runtime_test_text:
        fail("runtime missing-parent regression is missing from the real mission path")
    fixture_output = updates.get("RP-01-advisor-fixture-output-parent-guard")
    if not isinstance(fixture_output, dict):
        fail("matrix lacks advisor fixture output-parent guard acceptance record")
    fixture_refs = set(fixture_output.get("test_refs", []))
    if {"lab/affiliate-bot/cmd/bot/advisor_fixture_test.go", "scripts/audit_readiness.py", "scripts/tests/test_audit_readiness.py"} - fixture_refs:
        fail("advisor fixture output-parent record lacks CLI and audit regressions")
    fixture_scope = fixture_output.get("scope")
    if not isinstance(fixture_scope, str) or "MkdirTemp" not in fixture_scope or "external target" not in fixture_scope:
        fail("advisor fixture output-parent record lacks bounded external-write disclosure")
    fixture_source = root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
    fixture_text = fixture_source.read_text(encoding="utf-8") if fixture_source.is_file() else ""
    fixture_test = root / "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go"
    fixture_test_text = fixture_test.read_text(encoding="utf-8") if fixture_test.is_file() else ""
    if "os.Lstat(parent)" not in fixture_text or "fixture output parent must be an existing non-symlink directory" not in fixture_text or "TestAdvisorFixtureBundleRejectsSymlinkOutputParent" not in fixture_test_text or "os.ReadDir(outside)" not in fixture_test_text:
        fail("advisor fixture output-parent regression is missing from the real CLI path")
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
    # M11 has a distinct registry/ledger lifecycle, so its real-Bot barrier
    # proof must be independently documented and present.  A passing M10
    # process race cannot stand in for this claim.
    smoke_ref = "scripts/smoke_br16a_offline.py"
    if smoke_ref not in barrier.get("test_refs", []):
        fail("R10 M11 process-barrier record lacks the shared-runtime smoke regression")
    required_m11_scope = (
        "M11",
        "24 distinct canonical authorizations",
        "exactly one appends",
        "pending execution",
        "24-process M11 outcome barrier",
        "one outcome link",
    )
    if not all(token in scope for token in required_m11_scope):
        fail("R10 M11 process-barrier record lacks bounded canonical-authorization disclosure")
    smoke_source = root / smoke_ref
    smoke_text = smoke_source.read_text(encoding="utf-8") if smoke_source.is_file() else ""
    required_m11_smoke = (
        "production_race_authorizations",
        '"m11-reserve-barrier"',
        '"m11-reserve-authorization"',
        'sum(response["status"] == "APPENDED" for response in production_race_responses.values()) == 1',
        "production_race_head",
        'expected=1)["status"] == "REJECTED"',
        "production_outcome_race_responses",
        '"m11-outcome-barrier"',
        'sum(response["status"] == "APPENDED" for response in production_outcome_race_responses.values()) == 1',
        "production_outcome_race_head",
        '"br16-production-outcome-race"',
    )
    if not all(token in smoke_text for token in required_m11_smoke):
        fail("R10 M11 process-barrier regression is missing from the shared real-Bot smoke path")
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
    if "TestMissionStopJournalRecoversEveryPostRenameBoundary" not in stop_text or '"PUBLISHED_RECOVERY_REQUIRED"' not in stop_text or '"journal", "m11-manual-stop-journal.json", 1, false, false' not in stop_text or '"marker", "STOP", 3, true, true' not in stop_text or '"RECOVERY_REQUIRED"' not in stop_text:
        fail("durable STOP post-rename regression is missing from the learner Bot test path")
    if "m11ManualStopJournalPath" not in mission_source_text or "recoverM11ManualStopJournal" not in mission_source_text or '"m11-manual-stop-journal/v1"' not in mission_source_text:
        fail("durable STOP journal recovery is missing from the learner Bot runtime path")
    backup_source = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    backup_text = backup_source.read_text(encoding="utf-8") if backup_source.is_file() else ""
    if "TestBackupCreateRecoversPendingDirectStopJournal" not in stop_text or "recoverM11ManualStopJournal(args[1])" not in backup_text:
        fail("durable STOP journal recovery is missing from the backup path")
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
    m07_runtime_source = root / "lab/affiliate-bot/cmd/bot/m07.go"
    m07_runtime_text = m07_runtime_source.read_text(encoding="utf-8") if m07_runtime_source.is_file() else ""
    watcher_runtime_source = root / "lab/affiliate-bot/cmd/bot/watcher.go"
    watcher_runtime_text = watcher_runtime_source.read_text(encoding="utf-8") if watcher_runtime_source.is_file() else ""
    m11_fault_source = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    m11_fault_text = m11_fault_source.read_text(encoding="utf-8") if m11_fault_source.is_file() else ""
    if "TestWriteNewJSONReportsVisibleArtifactWhenParentSyncIsUnconfirmed" not in artifact_text or "TestMissionM08IntentReportsUnconfirmedVisibleArtifact" not in source_text or "TestMissionM08PolicyReportsUnconfirmedVisibleArtifact" not in source_text or "TestM07CLIDisclosesUnconfirmedVisibleArtifact" not in m07_text or "TestM07CLIProposalDisclosesUnconfirmedVisibleArtifact" not in m07_text or "artifact = proposal" not in m07_runtime_text or "TestM07HTTPAdapterReportsUnconfirmedVisibleToolArtifact" not in watcher_text or "TestM07HTTPAdapterReportsUnconfirmedVisibleProposal" not in watcher_text or "func m07PersistenceFailureResponse" not in watcher_runtime_text or 'response["artifact_id"] = artifactID' not in watcher_runtime_text or "recovery handoff uncertainty was not disclosed" not in m11_fault_text or '"m11-recovery-export"' not in m11_fault_text:
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
    m11_visible_append = updates.get("RP-07-m11-visible-outcome-append")
    if not isinstance(m11_visible_append, dict) or "PUBLISHED_RECOVERY_REQUIRED" not in m11_visible_append.get("scope", "") or "EXACT_DUPLICATE" not in m11_visible_append.get("scope", ""):
        fail("matrix lacks M11 visible outcome-append acknowledgement acceptance")
    m11_outcome_source = root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
    m11_outcome_text = m11_outcome_source.read_text(encoding="utf-8") if m11_outcome_source.is_file() else ""
    if "visibleAppendUncertainError" not in mission_source_text or "m11OutcomeAppendUncertainty" not in mission_source_text or "TestMissionM11OutcomeDisclosesVisibleAppendAcknowledgementUncertainty" not in m11_outcome_text:
        fail("M11 visible outcome-append acknowledgement regression is missing from learner path")
    if 'm11OutcomeWriteFault("before_remove")' not in mission_source_text or "TestMissionM11OutcomeDisclosesPublishedJournalCleanupUncertainty" not in m11_outcome_text or "cleanup fails" not in m11_visible_append.get("scope", ""):
        fail("M11 published outcome journal cleanup regression is missing from learner path")
    m11_execution_journals = updates.get("RP-07-m11-execution-visible-journals")
    if not isinstance(m11_execution_journals, dict) or "PUBLISHED_RECOVERY_REQUIRED" not in m11_execution_journals.get("scope", ""):
        fail("matrix lacks M11 execution visible-journal recovery acceptance")
    if "TestMissionM11ExecutionJournalsVisiblePublishUncertaintyDefersLockedReplay" not in source_text or "TestMissionM11ExecutionJournalsDisclosePublishedCleanupUncertainty" not in source_text or '"m11-record-failed"' not in source_text or '"m11-record-unknown"' not in source_text or "m11JournalCleanupWriteFault" not in mission_source_text or "cleanup fails" not in m11_execution_journals.get("scope", "") or "artifactIfAtomicPublishUncertain" not in mission_source_text:
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


def audit_m07_raw_json_transport(root, matrix, plan_text):
    """Keep n8n's M07 numeric-value boundary on the raw JSON transport path."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-05-m07-n8n-raw-json-transport-20260917")
    if not isinstance(record, dict):
        fail("matrix lacks M07 n8n raw JSON transport acceptance record")
    scope = record.get("scope")
    required_scope = (
        "tool_result_text",
        "model_output_text",
        "9007199254740993",
        "9007199254740992",
        "Node 24",
        "CI",
    )
    if not isinstance(scope, str) or any(marker not in scope for marker in required_scope):
        fail("M07 raw JSON transport record lacks its bounded exact-number/CI scope")
    if "Cập nhật M07 n8n raw JSON transport" not in plan_text:
        fail("M07 raw JSON transport lacks a scoped plan marker")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/watcher.go",
        "lab/affiliate-bot/cmd/bot/watcher_test.go",
        "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
        "scripts/run_n8n_engine_regression.py",
        "scripts/validate_n8n_m07.py",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/mission-agent-path-ci.yml",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M07 raw JSON transport record lacks implementation, regression or CI refs")

    blueprint_path = root / "lab/n8n/M07-readonly-evidence-agent.blueprint.json"
    blueprint = json.loads(blueprint_path.read_text(encoding="utf-8")) if blueprint_path.is_file() else {}
    nodes = {node.get("name"): node for node in blueprint.get("nodes", []) if isinstance(node, dict)}
    raw_node = nodes.get("Require Raw Model JSON Text")
    if not isinstance(raw_node, dict):
        fail("M07 raw JSON transport blueprint node is missing")
    raw_code = raw_node.get("parameters", {}).get("jsCode", "")
    if not all(marker in raw_code for marker in ("MODEL_OUTPUT_MUST_BE_RAW_JSON_TEXT", "typeof $json.output==='string'", "typeof $json.text==='string'")):
        fail("M07 raw model-output string guard is missing from the blueprint")
    for node_name in ("Validate Grounding Adapter", "Persist Agent Proposal Adapter"):
        body = nodes.get(node_name, {}).get("parameters", {}).get("jsonBody", "")
        if "model_output_text" not in body or "model_output:JSON.parse" in body:
            fail(f"{node_name} can reparse model JSON before the adapter")
    agent_text = nodes.get("Read-only Evidence Agent", {}).get("parameters", {}).get("text", "")
    if "artifact_raw_json" not in agent_text or "evidence_raw_json" not in agent_text or "JSON.stringify" in agent_text:
        fail("M07 Agent is not fed adapter-preserved JSON text")
    connections = blueprint.get("connections", {})
    raw_destinations = {
        item.get("node")
        for branch in connections.get("Read-only Evidence Agent", {}).get("main", [])
        for item in branch
        if isinstance(item, dict)
    }
    if "Require Raw Model JSON Text" not in raw_destinations:
        fail("M07 Agent is not connected to the raw model-output guard")

    runner_path = root / "scripts/run_n8n_engine_regression.py"
    runner = runner_path.read_text(encoding="utf-8") if runner_path.is_file() else ""
    required_runner = (
        "M07_MODEL_CASES",
        '"model-exact-number"',
        "tool_result_text:$(\'M07 Adapter Input\').item.json.tool_request_json",
        "expected_large_number=True",
        'b"9007199254740993" not in exact_tool_bytes',
        "exact_proposal_path.read_bytes() != exact_proposal_before_restart",
    )
    if any(marker not in runner for marker in required_runner) or "tool_result:JSON.parse($(\'M07 Adapter Input\').item.json.tool_request_json)" in runner:
        fail("M07 raw JSON transport regression is missing from the n8n runner")

    watcher_path = root / "lab/affiliate-bot/cmd/bot/watcher.go"
    watcher_test_path = root / "lab/affiliate-bot/cmd/bot/watcher_test.go"
    watcher = watcher_path.read_text(encoding="utf-8") if watcher_path.is_file() else ""
    watcher_test = watcher_test_path.read_text(encoding="utf-8") if watcher_test_path.is_file() else ""
    required_watcher = (
        "ToolResultText string",
        "toolResult = json.RawMessage(request.ToolResultText)",
        "TOOL_RESULT_AMBIGUOUS",
    )
    if any(marker not in watcher for marker in required_watcher) or not all(marker in watcher_test for marker in ("ToolResultText:", "9007199254740993", "TOOL_RESULT_AMBIGUOUS")):
        fail("M07 adapter raw tool-result transport regression is missing")


def audit_selected_source_operated_run(root, matrix, plan_text):
    """Require a durable, scoped record when local operated evidence is recorded."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-10-m06-selected-source-local-operated-run")
    required_scope = (
        "local operator-observed",
        "canonical_history_ack=true",
        "execution_permitted=false",
        "no ACCESSTRADE request",
        "not an independently reviewed",
    )
    if not isinstance(record, dict) or any(marker not in record.get("scope", "") for marker in required_scope):
        fail("selected-source operated run record lacks its bounded local evidence scope")
    required_refs = {
        "core/m06/accesstrade_shopee.go",
        "lab/affiliate-bot/cmd/bot/watcher.go",
        "lab/n8n/M06-accesstrade-shopee-readonly.blueprint.json",
        "docs/architecture/EVIDENCE-M06-ACCESSTRADE-SHOPEE-OPERATED-20260915.md",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("selected-source operated run record lacks durable implementation/evidence refs")
    marker = "M06 selected-source operated run"
    evidence_path = root / "docs/architecture/EVIDENCE-M06-ACCESSTRADE-SHOPEE-OPERATED-20260915.md"
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    if marker not in plan_text or "## n8n local run" not in evidence or "GET_MORE_DATA" not in evidence or "replay=MATCH" not in evidence:
        fail("selected-source operated run evidence is missing its durable plan or result record")
    graph_text = (root / "docs/plans/READINESS-EVIDENCE-GRAPH.json").read_text(encoding="utf-8")
    for claim_id in ("BR-13-operated-local-selected-source", "BR-14-operated-local-selected-source"):
        if claim_id not in graph_text:
            fail("selected-source operated run evidence graph claim is missing")


def audit_local_recovery_drill(root, matrix, plan_text):
    """Require a durable, bounded record when the local recovery smoke is operated."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-10-br18b-local-recovery-drill")
    required_scope = (
        "fresh local learner Bot process",
        "typed manifest v3",
        "durable STOP",
        "target-host deployment recovery",
        "RP-10 remains OPEN",
    )
    if not isinstance(record, dict) or any(marker not in record.get("scope", "") for marker in required_scope):
        fail("local recovery drill record lacks its bounded runtime/external scope")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "scripts/smoke_br18b_backup_restore.py",
        "docs/architecture/EVIDENCE-BR18B-LOCAL-RECOVERY-DRILL-20260915.md",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("local recovery drill record lacks durable implementation/evidence refs")
    evidence_path = root / "docs/architecture/EVIDENCE-BR18B-LOCAL-RECOVERY-DRILL-20260915.md"
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    if "BR-18b local recovery drill" not in plan_text or "BR-18b PASS" not in evidence or "## Giới hạn" not in evidence:
        fail("local recovery drill evidence is missing its durable plan or result record")
    graph_text = (root / "docs/plans/READINESS-EVIDENCE-GRAPH.json").read_text(encoding="utf-8")
    if "BR-18b-operated-local-recovery" not in graph_text:
        fail("local recovery drill evidence graph claim is missing")


def audit_assisted_fresh_workspace(root, matrix, plan_text):
    """Require assisted reruns to remain distinct from an independent pilot."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-10-br16b-assisted-fresh-workspace")
    required_scope = (
        "fresh local workspace",
        "assisted automated verification",
        "not a clean-machine self-service pilot PASS",
        "RP-10 remains OPEN",
    )
    if not isinstance(record, dict) or any(marker not in record.get("scope", "") for marker in required_scope):
        fail("assisted fresh-workspace record lacks its pilot boundary")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/main.go",
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "scripts/smoke_quickstart.py",
        "scripts/smoke_br16a_offline.py",
        "docs/architecture/EVIDENCE-BR16B-ASSISTED-FRESH-WORKSPACE-20260915.md",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("assisted fresh-workspace record lacks durable implementation/evidence refs")
    evidence_path = root / "docs/architecture/EVIDENCE-BR16B-ASSISTED-FRESH-WORKSPACE-20260915.md"
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    if "BR-16b — assisted fresh-workspace verification" not in evidence or "QUICKSTART SMOKE PASS" not in evidence or "human pilot `INCOMPLETE`" not in evidence:
        fail("assisted fresh-workspace evidence is missing its result or pilot boundary")
    graph_text = (root / "docs/plans/READINESS-EVIDENCE-GRAPH.json").read_text(encoding="utf-8")
    if "BR-16a-operated-assisted-fresh-workspace" not in graph_text:
        fail("assisted fresh-workspace evidence graph claim is missing")


def audit_n8n_engine_runtime_compatibility(root, matrix, plan_text):
    """Keep the checked-in engine evidence tied to the Node major n8n needs."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-08-n8n-engine-node24-reproduction")
    if not isinstance(record, dict) or "Node 24.21.0" not in record.get("scope", ""):
        fail("matrix lacks scoped n8n Node 24 engine reproduction evidence")
    compatibility = root / "lab/n8n/COMPATIBILITY.md"
    workflow = root / ".github/workflows/mission-agent-path-ci.yml"
    runner = root / "scripts/run_n8n_engine_regression.py"
    compatibility_text = compatibility.read_text(encoding="utf-8") if compatibility.is_file() else ""
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    if "Re-run n8n engine cục bộ (2026-09-14)" not in plan_text or "Node `24.21.0`" not in compatibility_text or "isolated-vm" not in compatibility_text:
        fail("n8n engine compatibility evidence lacks the Node/native boundary")
    schedule_runner = root / "scripts/run_n8n_m06_schedule_regression.py"
    if not runner.is_file() or not schedule_runner.is_file() or 'node-version: "24"' not in workflow_text or 'N8N_VERSION: "2.38.1"' not in workflow_text or "run_n8n_m06_schedule_regression.py" not in workflow_text:
        fail("n8n engine CI no longer pins the compatible Node/runtime pair")
    cache_gate = updates.get("RP-08-n8n-engine-cache-gating")
    if not isinstance(cache_gate, dict) or "full engine coverage remains required" not in cache_gate.get("scope", ""):
        fail("matrix lacks scoped n8n engine cache/gate acceptance")
    if "Select n8n engine coverage" not in workflow_text or "Restore pinned n8n runtime" not in workflow_text or "actions/cache@v4" not in workflow_text or "N8N_VERSION" not in workflow_text or "main_push" not in workflow_text or "n8n_related_change" not in workflow_text or "Report scoped engine skip" not in workflow_text:
        fail("n8n engine CI cache/gate is missing or can silently remove full coverage")


def audit_deterministic_runtime_sharding(root, matrix, plan_text):
    """Keep parallel CI execution from dropping a mandatory regression shard."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-08-deterministic-runtime-sharding")
    if not isinstance(record, dict) or "eight required parallel jobs" not in record.get("scope", ""):
        fail("matrix lacks scoped deterministic CI sharding acceptance")
    if "Cập nhật deterministic CI sharding và Go cache" not in plan_text:
        fail("deterministic CI sharding lacks a scoped plan marker")
    workflow = root / ".github/workflows/curriculum-ci.yml"
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_jobs = (
        "\n  deterministic-runtime:\n",
        "\n  learner-bot-tests-shard-0:\n",
        "\n  learner-bot-tests-shard-1:\n",
        "\n  learner-bot-race:\n",
        "\n  deterministic-quickstart:\n",
        "\n  deterministic-smokes-foundations:\n",
        "\n  deterministic-smokes-m06-m07:\n",
        "\n  deterministic-smokes-backup-mutations:\n",
    )
    required = (
        "Learner Bot race regression",
        "run_learner_bot_test_shard.py --shard-index 0 --shard-count 2",
        "run_learner_bot_test_shard.py --shard-index 1 --shard-count 2",
        "Beginner quickstart isolated-clone smoke",
        "BR-16a shared M00-M11 learner chain",
        "RP-08 M11 canonical identity mutation proof",
        "cache-dependency-path:",
        "lab/affiliate-bot/go.sum",
        "contracts/go.sum",
        "core/go.sum",
        "lab/mission-runtime/go.sum",
    )
    if not all(token in workflow_text for token in required_jobs + required):
        fail("deterministic CI shard/cache is missing a required coverage boundary")


def audit_accesstrade_pending_import_recovery(root, matrix, plan_text):
    """Keep the multi-file ACCESSTRADE sidecar from becoming a manual delete path."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-accesstrade-pending-import-recovery")
    if not isinstance(record, dict) or "only the exact local CSV/manifest snapshot" not in record.get("scope", "") or "PUBLISHED_RECOVERY_REQUIRED" not in record.get("scope", "") or "removal is visible before parent sync" not in record.get("scope", ""):
        fail("matrix lacks scoped ACCESSTRADE pending-import recovery acceptance")
    if "Cập nhật ACCESSTRADE pending import recovery" not in plan_text:
        fail("ACCESSTRADE pending-import recovery lacks a scoped plan marker")
    source = root / "lab/affiliate-bot/cmd/bot/accesstrade_import.go"
    test = root / "lab/affiliate-bot/cmd/bot/accesstrade_import_test.go"
    source_text = source.read_text(encoding="utf-8") if source.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    required = (
        "func runAccesstradeOutcomeRecover",
        "sameReceipt(pending, receipt)",
        "PARTIAL_DUPLICATE",
        "validateAccesstradeReceiptGraph(finalReceipts, finalOutcomes, args[3])",
        "removeAccesstradeJournal(args[3])",
        'emit("PUBLISHED_RECOVERY_REQUIRED", map[string]any{"outcomes": candidates, "receipt": receipt}',
        "accesstradeJournalCleanupFailure(\"after_remove_before_parent_sync\")",
    )
    if not all(token in source_text for token in required) or "TestAccesstradeImporterRecoveryReplaysOnlyExactPendingSnapshot" not in test_text or "TestAccesstradeImporterDisclosesVisibleJournalCleanupUncertainty" not in test_text or "changed report and partial outcomes fail closed" not in test_text:
        fail("ACCESSTRADE pending-import recovery regression is missing from learner path")


def audit_campaign_result_visible_ack(root, matrix, plan_text):
    """Keep an ambiguous paid canary result from looking safely retryable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-03-campaign-result-visible-ack")
    if not isinstance(record, dict) or "PUBLISHED_RECOVERY_REQUIRED" not in record.get("scope", "") or "without a second provider request" not in record.get("scope", ""):
        fail("matrix lacks scoped campaign-result acknowledgement acceptance")
    if "Cập nhật campaign result visible acknowledgement" not in plan_text:
        fail("campaign-result acknowledgement lacks a scoped plan marker")
    results = root / "lab/affiliate-bot/cmd/bot/advisor_results.go"
    canary = root / "lab/affiliate-bot/cmd/bot/advisor_canary.go"
    test = root / "lab/affiliate-bot/cmd/bot/advisor_canary_test.go"
    results_text = results.read_text(encoding="utf-8") if results.is_file() else ""
    canary_text = canary.read_text(encoding="utf-8") if canary.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    required_results = (
        "func campaignResultPublishFailure",
        "campaignResultVisible(path, candidate)",
        "campaignResultWriteFailure(\"after_result_file_sync\")",
        'return n, "PUBLISHED_RECOVERY_REQUIRED", err',
    )
    required_canary = (
        "status == \"PUBLISHED_RECOVERY_REQUIRED\"",
        "resolvedCampaignResult(path, n)",
        "campaignCanaryProvider",
    )
    if not all(token in results_text for token in required_results) or not all(token in canary_text for token in required_canary) or "TestCanaryDisclosesVisibleResultAcknowledgementWithoutProviderRetry" not in test_text or "TestCanaryCLIDisclosesVisibleResultAcknowledgement" not in test_text:
        fail("campaign-result acknowledgement regression is missing from learner path")


def audit_committed_journal_cleanup_ack(root, matrix, plan_text):
    """Keep completed M10/direct-STOP journals from looking safely uncommitted."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-03-07-committed-journal-cleanup-ack")
    if not isinstance(record, dict) or "M10 canary, cost-bound or execution-record" not in record.get("scope", "") or "direct M11 STOP" not in record.get("scope", "") or "PUBLISHED_RECOVERY_REQUIRED" not in record.get("scope", ""):
        fail("matrix lacks scoped committed journal-cleanup acknowledgement acceptance")
    if "Cập nhật committed recovery-journal cleanup acknowledgement" not in plan_text:
        fail("committed journal-cleanup acknowledgement lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    required = (
        "type journalCleanupUncertainError struct",
        "func removeCommittedRecoveryJournal(path string) error",
        'recoveryJournalCleanupWriteFault(path, "after_remove_before_parent_sync")',
        "return removeCommittedRecoveryJournal(m10CanaryJournalPath(dir))",
        "return removeCommittedRecoveryJournal(m10CostBoundJournalPath(dir))",
        "return removeCommittedRecoveryJournal(m10ExecutionJournalPath(dir))",
        "return removeCommittedRecoveryJournal(m11ManualStopJournalPath(dir))",
        "errors.As(err, &cleanupUncertain)",
    )
    if not all(token in source for token in required) or "TestMissionDisclosesCommittedJournalCleanupUncertainty" not in test:
        fail("committed journal-cleanup acknowledgement regression is missing from learner path")


def audit_m11_post_remove_journal_cleanup_ack(root, matrix, plan_text):
    """Keep the M11 post-Remove acknowledgement boundary executable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-post-remove-journal-cleanup-ack")
    if not isinstance(record, dict) or "outcome, FAILED execution or UNKNOWN-to-STOP" not in record.get("scope", "") or "after Remove but before parent-directory acknowledgement" not in record.get("scope", "") or "PUBLISHED_RECOVERY_REQUIRED" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 post-remove journal-cleanup acknowledgement acceptance")
    if "Cập nhật M11 post-remove journal cleanup acknowledgement" not in plan_text:
        fail("M11 post-remove journal-cleanup acknowledgement lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    command_test_path = root / "lab/affiliate-bot/cmd/bot/mission_command_test.go"
    outcome_test_path = root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    command_test = command_test_path.read_text(encoding="utf-8") if command_test_path.is_file() else ""
    outcome_test = outcome_test_path.read_text(encoding="utf-8") if outcome_test_path.is_file() else ""
    required = (
        'm11OutcomeWriteFault("after_remove_before_parent_sync")',
        'm11JournalCleanupWriteFault("failed_after_remove_before_parent_sync")',
        'm11JournalCleanupWriteFault("unknown_after_remove_before_parent_sync")',
    )
    if not all(token in source for token in required) or "TestMissionM11ExecutionJournalsDisclosePostRemoveCleanupUncertainty" not in command_test or "TestMissionM11OutcomeDisclosesPostRemoveCleanupUncertainty" not in outcome_test:
        fail("M11 post-remove journal-cleanup acknowledgement regression is missing from learner path")


def audit_m10_canary_cost_journal_path_guard(root, matrix, plan_text):
    """Keep each M10 authority replay journal on the guarded reader path."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-03-m10-canary-cost-journal-path-guard")
    if not isinstance(record, dict) or "canary and cost-bound recovery journal rejects a FIFO" not in record.get("scope", "") or "same-byte external target replacement after open" not in record.get("scope", ""):
        fail("matrix lacks scoped M10 canary/cost journal path-guard acceptance")
    if "Cập nhật M10 canary/cost journal path guards" not in plan_text:
        fail("M10 canary/cost journal path guard lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/m11_journal_unix_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    required_source = (
        "func readM10CanaryJournal(path string) ([]byte, error)",
        "func readM10CostBoundJournal(path string) ([]byte, error)",
        "func m10CanaryJournalRecoveryRequired(dir string) error",
        "func m10CostBoundJournalRecoveryRequired(dir string) error",
    )
    required_test = (
        "TestM10CanaryAndCostBoundJournalFIFOsFailClosedBeforeRuntimeRead",
        "m10CanaryJournalPath",
        "m10CostBoundJournalPath",
        '"m10-canary":     readM10CanaryJournal',
        '"m10-cost-bound": readM10CostBoundJournal',
    )
    if not all(token in source for token in required_source) or not all(token in test for token in required_test):
        fail("M10 canary/cost journal path-guard regression is missing from learner path")


def audit_m11_recovery_admission_approval_guard(root, matrix, plan_text):
    """Keep recovery handoff admission from accepting a lease with no review."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-recovery-admission-approval-guard")
    if not isinstance(record, dict) or "exact immutable new-lease approval" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 recovery-admission approval guard")
    if "Cập nhật recovery admission approval guard" not in plan_text:
        fail("M11 recovery-admission approval guard lacks a scoped plan marker")
    validator = (root / "core/m11/artifact_registry.go").read_text(encoding="utf-8")
    core_test = (root / "core/m11/artifact_test.go").read_text(encoding="utf-8")
    loader_test = (root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go").read_text(encoding="utf-8")
    if "approval, approvalOK := approvals[x.NewApprovalID]" not in validator or "admissionAt.After(approvalAt)" not in validator or "recovery admission without its immutable lease approval was accepted" not in core_test or "TestM11RegistryLoaderRejectsRecoveryAdmissionWithoutRecordedApproval" not in loader_test:
        fail("M11 recovery-admission approval regression is missing")


def audit_m11_authorization_gate_exact_lineage(root, matrix, plan_text):
    """Keep authorization evidence bound to the gate's exact artifacts."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-authorization-gate-exact-lineage-20260916")
    if not isinstance(record, dict) or "exact gate-evaluated lease ID/version/hash" not in record.get("scope", "") or "health snapshot ID/hash" not in record.get("scope", "") or "cost bound ID/hash/minor" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 authorization gate exact-lineage acceptance")
    if "Cập nhật M11 authorization exact gate lineage" not in plan_text:
        fail("M11 authorization gate exact-lineage acceptance lacks a scoped plan marker")
    validator_path = root / "core/m11/artifact_registry.go"
    test_path = root / "core/m11/artifact_test.go"
    validator = validator_path.read_text(encoding="utf-8") if validator_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    required_validator = (
        "gate.LeaseID != x.ProductionLeaseID",
        "gate.HealthSnapshotID != x.ProductionHealthSnapshotID",
        "gate.CostBoundID != x.ProductionCostBoundID",
        "gate.CostBoundMinor != x.ProductionCostBoundMinor",
    )
    required_test = (
        "func TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks",
        "authorization switched to health/cost artifacts not evaluated by its gate",
        "alternateHealth.SnapshotID = \"health-2\"",
        "alternateCost.CostBoundID = \"cost-2\"",
    )
    if any(marker not in validator for marker in required_validator) or any(marker not in test for marker in required_test):
        fail("M11 authorization gate exact-lineage regression is missing")


def audit_m11_correlation_lineage(root, matrix, plan_text):
    """Keep the M11 correlation root intact across graph and chain readers."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-correlation-lineage-20260917")
    if not isinstance(record, dict) or not all(token in record.get("scope", "") for token in ("lease correlation root", "cost-bound", "execution", "historical-chain")):
        fail("matrix lacks scoped M11 correlation-lineage acceptance")
    if "Cập nhật M11 correlation lineage" not in plan_text:
        fail("M11 correlation lineage lacks a scoped plan marker")
    registry_path = root / "core/m11/artifact_registry.go"
    historical_path = root / "core/m11/historical_chain.go"
    artifact_test_path = root / "core/m11/artifact_test.go"
    registry_test_path = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    chain_test_path = root / "lab/mission-runtime/cmd/demo/m11_chain_test.go"
    evidence_path = root / "docs/architecture/EVIDENCE-M11-CORRELATION-LINEAGE-20260917.md"
    registry = registry_path.read_text(encoding="utf-8") if registry_path.is_file() else ""
    historical = historical_path.read_text(encoding="utf-8") if historical_path.is_file() else ""
    artifact_test = artifact_test_path.read_text(encoding="utf-8") if artifact_test_path.is_file() else ""
    registry_test = registry_test_path.read_text(encoding="utf-8") if registry_test_path.is_file() else ""
    chain_test = chain_test_path.read_text(encoding="utf-8") if chain_test_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_registry = (
        "bound.CorrelationID != lease.CorrelationID",
        "x.CorrelationID != lease.CorrelationID",
        "auth.CorrelationID != x.CorrelationID",
        "auth.CorrelationID != lease.CorrelationID",
    )
    required_historical = (
        "c.Intent.CorrelationID != c.Execution.CorrelationID",
        "c.Authorization.CorrelationID != c.Intent.CorrelationID",
    )
    required_tests = (
        "checksum-valid cost bound from another correlation lineage",
        "TestM11RegistryLoaderRejectsMixedCorrelationLineage",
        '{"execution", "correlation_id", "other", "BROKEN_LINK"}',
    )
    if (
        any(marker not in registry for marker in required_registry)
        or any(marker not in historical for marker in required_historical)
        or any(marker not in artifact_test for marker in required_tests[:1])
        or required_tests[1] not in registry_test
        or required_tests[2] not in chain_test
        or "## Verification" not in evidence
        or "NOT_READY_FOR_PRODUCTION" not in evidence
    ):
        fail("M11 correlation-lineage regression is missing")


def audit_m11_reverse_ledger_graph(root, matrix, plan_text):
    """Keep ledger reconciliation and outcome links resolvable in both directions."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-reverse-ledger-graph-guards-20260917")
    required_scope = ("reverse-resolves", "ReconciliationResolutionIDs", "STOPPED", "OutcomeID", "checksum-valid")
    if not isinstance(record, dict) or any(token not in record.get("scope", "") for token in required_scope):
        fail("matrix lacks scoped M11 reverse-ledger graph acceptance")
    if "Cập nhật M11 reverse-link ledger graph guard" not in plan_text:
        fail("M11 reverse-ledger graph lacks a scoped plan marker")
    registry_path = root / "core/m11/artifact_registry.go"
    artifact_path = root / "core/m11/artifact.go"
    artifact_test_path = root / "core/m11/artifact_test.go"
    loader_test_path = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    evidence_path = root / "docs/architecture/EVIDENCE-M11-REVERSE-LEDGER-GRAPH-20260917.md"
    registry = registry_path.read_text(encoding="utf-8") if registry_path.is_file() else ""
    artifact = artifact_path.read_text(encoding="utf-8") if artifact_path.is_file() else ""
    artifact_test = artifact_test_path.read_text(encoding="utf-8") if artifact_test_path.is_file() else ""
    loader_test = loader_test_path.read_text(encoding="utf-8") if loader_test_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_registry = (
        "resolutionsByID := map[string]ProductionReconciliationResolution{}",
        "production ledger reconciliation link is orphaned or mismatched",
        "production ledger outcome link does not match its evaluation",
        "len(x.EvidenceIDs) != 1 || x.EvidenceIDs[0] != x.OutcomeID",
    )
    required_artifact = (
        "resolutionIDs := map[string]bool{}",
        'if id == "" || resolutionIDs[id] {',
    )
    required_tests = (
        "valid ledger-to-reconciliation reverse link rejected",
        "orphan reconciliation resolution link was accepted",
        "ledger outcome link with a swapped outcome ID was accepted",
        "duplicate reconciliation resolution IDs was accepted",
        "evaluation with evidence unrelated to its fixture outcome was accepted",
        "evaluation with more than one evidence ID was accepted",
    )
    required_loader = (
        "TestM11RegistryLoaderRejectsOrphanReconciliationLedgerLink",
        "schema-valid orphan reconciliation link reached runtime loader",
        "TestM11RegistryLoaderRejectsEvaluationWithMismatchedEvidence",
        "checksum-valid evaluation with mismatched evidence reached runtime loader",
    )
    if (
        any(marker not in registry for marker in required_registry)
        or any(marker not in artifact for marker in required_artifact)
        or any(marker not in artifact_test for marker in required_tests)
        or any(marker not in loader_test for marker in required_loader)
        or "## Verification" not in evidence
        or "NOT_READY_FOR_PRODUCTION" not in evidence
    ):
        fail("M11 reverse-ledger graph regression is missing")
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    required_refs = {
        "core/m11/artifact.go",
        "core/m11/artifact_registry.go",
        "core/m11/artifact_test.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M11-REVERSE-LEDGER-GRAPH-20260917.md",
    }
    if required_refs - refs:
        fail("M11 reverse-ledger graph acceptance lacks implementation, test, audit, evidence or CI refs")
    graph_path = root / "docs/plans/READINESS-EVIDENCE-GRAPH.json"
    graph = json.loads(graph_path.read_text(encoding="utf-8")) if graph_path.is_file() else {}
    br17 = next((item for item in graph.get("criteria", []) if item.get("criterion_id") == "BR-17"), {})
    claims = br17.get("claims", []) if isinstance(br17, dict) else []
    if not any(claim.get("id") == "BR-17-test-m11-reverse-ledger-graph-20260917" for claim in claims if isinstance(claim, dict)):
        fail("M11 reverse-ledger graph evidence claim is missing")
    evaluation_record = updates.get("RP-07-m11-fixture-evaluation-evidence-cardinality-20260917")
    evaluation_scope = evaluation_record.get("scope", "") if isinstance(evaluation_record, dict) else ""
    if not isinstance(evaluation_record, dict) or any(token not in evaluation_scope for token in ("exactly one fixture outcome evidence ID", "checksum-valid", "fail closed")):
        fail("matrix lacks scoped M11 fixture-evaluation evidence cardinality acceptance")
    evaluation_refs = set(evaluation_record.get("implementation_refs", [])) | set(evaluation_record.get("test_refs", []))
    required_evaluation_refs = {
        "core/m11/artifact_registry.go",
        "core/m11/artifact_test.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M11-REVERSE-LEDGER-GRAPH-20260917.md",
    }
    if required_evaluation_refs - evaluation_refs:
        fail("M11 fixture-evaluation evidence acceptance lacks implementation, test, audit, evidence or CI refs")
    if not any(claim.get("id") == "BR-17-test-m11-fixture-evaluation-evidence-20260917" for claim in claims if isinstance(claim, dict)):
        fail("M11 fixture-evaluation evidence claim is missing")


def audit_m11_authorization_lineage_mutation(root, matrix, plan_text):
    """Require CI mutation coverage for the exact authorization lineage guard."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-08-m11-authorization-lineage-mutation-20260916")
    if not isinstance(record, dict) or "disposable-copy mutation" not in record.get("scope", "") or "real core graph regression" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 authorization lineage mutation acceptance")
    if "Cập nhật RP-08 M11 authorization lineage mutation proof" not in plan_text:
        fail("M11 authorization lineage mutation lacks a scoped plan marker")
    script_path = root / "scripts/mutate_m11_authorization_lineage.py"
    script = script_path.read_text(encoding="utf-8") if script_path.is_file() else ""
    workflow = (root / ".github/workflows/curriculum-ci.yml").read_text(encoding="utf-8")
    required_script = (
        "LINEAGE_GUARDS",
        "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks",
        "mutated = source",
        "authorization switched to health/cost artifacts not evaluated by its gate",
        "mutated M11 authorization gate lineage guard unexpectedly passed",
    )
    if any(marker not in script for marker in required_script) or "scripts/mutate_m11_authorization_lineage.py" not in workflow:
        fail("M11 authorization lineage mutation proof is missing or not wired to CI")


def audit_m11_fixture_outcome_execution_cardinality(root, matrix, plan_text):
    """Keep direct M11 JSONL reads aligned with one-outcome-per-execution."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-fixture-outcome-execution-cardinality")
    if not isinstance(record, dict) or "two otherwise valid distinct outcome IDs" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 fixture-outcome execution cardinality")
    if "Cập nhật M11 fixture-outcome execution cardinality" not in plan_text:
        fail("M11 fixture-outcome execution cardinality lacks a scoped plan marker")
    source = (root / "lab/affiliate-bot/cmd/bot/mission_command.go").read_text(encoding="utf-8")
    start = source.find("func loadM11FixtureOutcomes")
    end = source.find("\nfunc appendM11FixtureOutcome", start)
    loader = source[start:end] if start >= 0 and end > start else ""
    test = (root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go").read_text(encoding="utf-8")
    if "|| seenExecution[outcome.EffectRef.EffectID] {" not in loader or "TestM11FixtureOutcomeLoaderRejectsDuplicateExecutionBeforeRuntimeUse" not in test:
        fail("M11 fixture-outcome execution cardinality regression is missing")


def audit_m11_recovery_handoff_strict_decode(root, matrix, plan_text):
    """Keep recovery admission from collapsing duplicate portable-input keys."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-recovery-handoff-strict-decode")
    if not isinstance(record, dict) or "shared strict JSON decoder" not in record.get("scope", ""):
        fail("matrix lacks scoped M11 recovery-handoff strict decoder")
    if "Cập nhật M11 recovery-handoff strict decoder" not in plan_text:
        fail("M11 recovery-handoff strict decoder lacks a scoped plan marker")
    source = (root / "lab/affiliate-bot/cmd/bot/m11_registry.go").read_text(encoding="utf-8")
    test_path = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    if "func decodeM11RecoveryHandoff(raw []byte)" not in source or "contracts.Decode(raw)" not in source or "provided, err := decodeM11RecoveryHandoff(providedRaw)" not in source or "duplicate-key recovery handoff reached admission" not in test:
        fail("M11 recovery-handoff strict decoder regression is missing")


def audit_m06_history_handoff_strict_decode(root, matrix, plan_text):
    """Keep M06 CLI and HTTP history handoff on the same strict raw decoder."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-04-m06-history-handoff-strict-decode")
    if not isinstance(record, dict) or "one strict raw HistoryRecord decoder" not in record.get("scope", ""):
        fail("matrix lacks scoped M06 history-handoff strict decoder")
    if "Cập nhật M06 history-handoff strict decoder" not in plan_text:
        fail("M06 history-handoff strict decoder lacks a scoped plan marker")
    schema = (root / "lab/affiliate-bot/cmd/bot/history_schema.go").read_text(encoding="utf-8")
    watcher = (root / "lab/affiliate-bot/cmd/bot/watcher.go").read_text(encoding="utf-8")
    test_path = root / "lab/affiliate-bot/cmd/bot/watcher_test.go"
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    if "func decodeHistoryHandoffRecord(raw []byte)" not in schema or "contracts.DecodeStrict(raw, &record)" not in schema or "record, err := decodeHistoryHandoffRecord(body)" not in watcher or "record, err := decodeHistoryHandoffRecord(raw)" not in watcher or "TestHistoryHandoffRejectsDuplicateRawKeyBeforePersistence" not in test:
        fail("M06 history-handoff strict decoder regression is missing")


def audit_m06_adapter_visible_append(root, matrix, plan_text):
    """Keep the workflow-facing M06 ACK-loss path distinct from a clean failure."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-04-history-visible-append")
    if not isinstance(record, dict) or "fixture-import CLI" not in record.get("scope", "") or "pinned-fixture fetch CLI" not in record.get("scope", "") or "workflow-facing M06 adapter expose PUBLISHED_RECOVERY_REQUIRED" not in record.get("scope", ""):
        fail("matrix lacks scoped M06 adapter visible-append boundary")
    if "Cập nhật M06 adapter visible append uncertainty" not in plan_text:
        fail("M06 adapter visible-append boundary lacks a scoped plan marker")
    watcher = (root / "lab/affiliate-bot/cmd/bot/watcher.go").read_text(encoding="utf-8")
    _, handler_marker, adapter_and_rest = watcher.partition("func m06AdapterHandler(historyPath string) http.HandlerFunc {")
    adapter, _, _ = adapter_and_rest.partition("func m07ArtifactPath(")
    test_path = root / "lab/affiliate-bot/cmd/bot/watcher_test.go"
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    fixture_prefix, fixture_marker, fixture_and_rest = watcher.partition("func runWatcher(args []string, stdout, stderr io.Writer) int {")
    fixture, _, _ = fixture_and_rest.partition("// m06AdapterRequest carries")
    fetch = (root / "lab/affiliate-bot/cmd/bot/watcher_fetch.go").read_text(encoding="utf-8")
    if not fixture_marker or "if err != nil && !isPublishedAppendUncertainty(err)" not in fixture or "TestM06FixtureImportDisclosesVisibleAppendUncertainty" not in test or "var watcherPinnedFetch = fetchPinnedWatcher" not in fetch or "if err != nil && !isPublishedAppendUncertainty(err)" not in fetch or "TestM06PinnedFetchDisclosesVisibleAppendUncertainty" not in test or not handler_marker or "if err != nil && !isPublishedAppendUncertainty(err)" not in adapter or "\"canonical_history_persisted\": true, \"execution_permitted\": false" not in adapter or "TestM06AdapterDisclosesVisibleAppendUncertainty" not in test:
        fail("M06 adapter visible-append regression is missing")


def audit_backup_target_absence_guard(root, matrix, plan_text):
    """Keep backup publish from deleting an empty caller-owned target."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-06-backup-target-absence-guard")
    if not isinstance(record, dict) or "not already exist" not in record.get("scope", "") or "appears while staging" not in record.get("scope", ""):
        fail("matrix lacks scoped backup target-absence guard")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/backup_command_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("test_refs", [])):
        fail("backup target-absence record lacks CLI and audit regressions")
    if "Cập nhật backup target-absence guard" not in plan_text:
        fail("backup target-absence guard lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    if source.count("backup target must not already exist") < 2 or "backup target appeared while staging" not in source or "os.Remove(args[2])" in source or "TestBackupRejectsPreexistingOrAppearedTargetWithoutMutation" not in test or "backup mutated target that appeared while staging" not in test:
        fail("backup target-absence guard regression is missing from the real CLI path")


def audit_backup_process_exit_lock(root, matrix, plan_text):
    """Keep local backup/restore retries safe after a writer process exits."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-backup-process-exit-before-publish")
    if not isinstance(record, dict):
        fail("matrix lacks backup process-exit lock acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/runtime_gate.go",
        "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go",
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/backup_command_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("backup process-exit record lacks implementation, test or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("process exit", "no incomplete target", "retry", "POSIX")):
        fail("backup process-exit record lacks a bounded local disclosure")
    if "Cập nhật local backup/restore process-exit lock" not in plan_text:
        fail("backup process-exit lock lacks a scoped plan marker")
    gate = root / "lab/affiliate-bot/cmd/bot/runtime_gate.go"
    posix = root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
    backup = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    test = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    gate_text = gate.read_text(encoding="utf-8") if gate.is_file() else ""
    posix_text = posix.read_text(encoding="utf-8") if posix.is_file() else ""
    backup_text = backup.read_text(encoding="utf-8") if backup.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    workflow = root / ".github/workflows/curriculum-ci.yml"
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_gate = ("acquireManagedPathLock(runtimeGatePath(dir))", "errManagedPathLockBusy")
    required_test = (
        "TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds",
        "TestBackupProcessKillBeforePublishLeavesNoTargetAndRetrySucceeds",
        "GO_WANT_BACKUP_PROCESS_TERMINATION",
        "ProcessState.Sys",
        "backup retry after process termination failed",
        "restore retry after process termination failed",
    )
    if not all(token in gate_text for token in required_gate) or "syscall.Flock" not in posix_text or "acquireManagedPathLock(path)" not in backup_text or not all(token in test_text for token in required_test) or "run_learner_bot_test_shard.py" not in workflow_text:
        fail("backup process-exit lock regression is missing from the real CLI/test/CI path")


def audit_managed_lock_path_guard(root, matrix, plan_text):
    """Prevent a managed lock pathname from following an external symlink."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-managed-lock-path-guard")
    if not isinstance(record, dict):
        fail("matrix lacks managed lock path-guard acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go",
        "lab/affiliate-bot/cmd/bot/backup_command_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("managed lock path-guard record lacks implementation, test or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("O_NOFOLLOW", "symlink", "external", "unchanged")):
        fail("managed lock path-guard record lacks a bounded path disclosure")
    if "Cập nhật managed lock path guard" not in plan_text:
        fail("managed lock path guard lacks a scoped plan marker")
    posix = root / "lab/affiliate-bot/cmd/bot/runtime_gate_posix.go"
    test = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    posix_text = posix.read_text(encoding="utf-8") if posix.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    workflow = root / ".github/workflows/curriculum-ci.yml"
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    if "validateManagedLockParent" not in posix_text or "unix.Openat" not in posix_text or "openManagedDirectory" not in posix_text or "unix.O_NOFOLLOW" not in posix_text or "TestManagedPathLockRejectsSymlinkWithoutTouchingExternal" not in test_text or "TestManagedPathLockRejectsSymlinkedParentWithoutTouchingExternal" not in test_text or "TestManagedPathLockRejectsAncestorSwapBeforeOpen" not in test_text or "external lock target changed after symlink rejection" not in test_text or "external parent target changed after symlink rejection" not in test_text or "external ancestor target changed after swap rejection" not in test_text or "run_learner_bot_test_shard.py" not in workflow_text:
        fail("managed lock path-guard regression is missing from the real test/CI path")


def audit_m11_process_kill_journal(root, matrix, plan_text):
    """Keep the real M11 process-boundary recovery proof on the CI path."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-process-kill-journal-recovery-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks M11 process-kill journal recovery acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
        ".github/workflows/curriculum-ci.yml",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("M11 process-kill record lacks implementation, test or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "RECOVERY_REQUIRED", "locked writer", "POSIX")):
        fail("M11 process-kill record lacks a bounded process/journal disclosure")
    if "Cập nhật M11 process-kill journal recovery" not in plan_text:
        fail("M11 process-kill journal recovery lacks a scoped plan marker")
    mission = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    backup = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    test = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    workflow = root / ".github/workflows/curriculum-ci.yml"
    mission_text = mission.read_text(encoding="utf-8") if mission.is_file() else ""
    backup_text = backup.read_text(encoding="utf-8") if backup.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_test = (
        "TestM11ProcessKillAfterJournalBeforeLedgerAppendLeavesJournalForFreshRecovery",
        "TestM11ProcessTerminationChild",
        "GO_WANT_M11_PROCESS_TERMINATION",
        "syscall.SIGKILL",
        "fresh process exposed the interrupted M11 transition",
        "locked writer did not recover",
    )
    if (
        "acquireManagedPathLock(lockPath)" not in mission_text
        or 'entry.Name() == ".mission.lock"' not in backup_text
        or not all(token in test_text for token in required_test)
        or "run_learner_bot_test_shard.py" not in workflow_text
    ):
        fail("M11 process-kill journal regression is missing from the real learner/CI path")


def audit_m11_outcome_process_kill(root, matrix, plan_text):
    """Keep the isolated M11 outcome process-boundary proof discoverable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-outcome-process-kill-recovery-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks M11 outcome process-kill recovery acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M11-OUTCOME-PROCESS-KILL-20260916.md",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M11 outcome process-kill record lacks implementation, test, evidence or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "RECOVERY_REQUIRED", "EXACT_DUPLICATE", "exactly one", "POSIX")):
        fail("M11 outcome process-kill record lacks a bounded outcome disclosure")
    if "Cập nhật M11 outcome process-kill recovery" not in plan_text:
        fail("M11 outcome process-kill recovery lacks a scoped plan marker")
    outcome_test = root / "lab/affiliate-bot/cmd/bot/m11_outcome_journal_test.go"
    helper_test = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    workflow = root / ".github/workflows/curriculum-ci.yml"
    outcome_text = outcome_test.read_text(encoding="utf-8") if outcome_test.is_file() else ""
    helper_text = helper_test.read_text(encoding="utf-8") if helper_test.is_file() else ""
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_test_markers = (
        "func TestM11OutcomeProcessKillAfterJournalBeforeLedgerAppend",
        "func TestM11OutcomeProcessKillAfterOutcomeAppend",
        "GO_WANT_M11_PROCESS_TERMINATION",
        "GO_M11_PROCESS_TERMINATION_MODE",
        "outcome-after-append-kill",
        "syscall.SIGKILL",
        "RECOVERY_REQUIRED",
        "EXACT_DUPLICATE",
        "len(outcomes) != 1",
    )
    if not all(marker in outcome_text for marker in required_test_markers) or "TestM11ProcessTerminationChild" not in helper_text or "run_learner_bot_test_shard.py" not in workflow_text:
        fail("M11 outcome process-kill regression is missing from the real learner/CI path")


def audit_m11_process_kill_after_ledger(root, matrix, plan_text):
    """Keep the post-ledger M11 recovery boundary executable and discoverable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-process-kill-after-ledger-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks M11 post-ledger process-kill acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_registry.go",
        "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M11 post-ledger process-kill record lacks implementation, test, evidence or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "directory-synced", "RECOVERY_REQUIRED", "locked writer", "without duplicate")):
        fail("M11 post-ledger process-kill record lacks a bounded visibility/recovery disclosure")
    if "Cập nhật M11 process-kill sau ledger append" not in plan_text:
        fail("M11 post-ledger process-kill lacks a scoped plan marker")
    mission_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    registry_path = root / "lab/affiliate-bot/cmd/bot/m11_registry.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/m11_registry_fault_test.go"
    workflow_path = root / ".github/workflows/curriculum-ci.yml"
    evidence_path = root / "docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md"
    mission = mission_path.read_text(encoding="utf-8") if mission_path.is_file() else ""
    registry = registry_path.read_text(encoding="utf-8") if registry_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    workflow = workflow_path.read_text(encoding="utf-8") if workflow_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_test = (
        "TestM11ProcessKillAfterLedgerAppendLeavesJournalForFreshRecovery",
        "GO_M11_PROCESS_TERMINATION_MODE=after-ledger-kill",
        'phase == "after_sync"',
        "ArtifactKindLedger",
        "countM11Artifacts",
        "fresh process exposed the interrupted post-ledger transition",
        "post-ledger recovery duplicated execution artifact",
        "syscall.SIGKILL",
    )
    if (
        "acquireManagedPathLock(lockPath)" not in mission
        or "func recordFailedM11Execution" not in registry
        or "func recordUnknownM11Execution" not in registry
        or not all(marker in test for marker in required_test)
        or "run_learner_bot_test_shard.py" not in workflow
        or "## Verification" not in evidence
    ):
        fail("M11 post-ledger process-kill regression is missing from the real learner/CI path")


def audit_m11_manual_stop_process_kill(root, matrix, plan_text):
    """Keep direct M11 STOP recovery tested across a real process boundary."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m11-manual-stop-process-kill-20260917")
    if not isinstance(record, dict):
        fail("matrix lacks M11 direct STOP process-kill acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m11_manual_stop_process_kill_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M11-MANUAL-STOP-PROCESS-KILL-20260917.md",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M11 direct STOP process-kill record lacks implementation, test, evidence or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "RECOVERY_REQUIRED", "STOPPED", "exact", "POSIX")):
        fail("M11 direct STOP process-kill record lacks a bounded recovery disclosure")
    if "Cập nhật M11 direct STOP process-kill recovery" not in plan_text:
        fail("M11 direct STOP process-kill lacks a scoped plan marker")
    test_path = root / "lab/affiliate-bot/cmd/bot/m11_manual_stop_process_kill_test.go"
    workflow_path = root / ".github/workflows/curriculum-ci.yml"
    evidence_path = root / "docs/architecture/EVIDENCE-M11-MANUAL-STOP-PROCESS-KILL-20260917.md"
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    workflow = workflow_path.read_text(encoding="utf-8") if workflow_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_test = (
        "TestM11ManualStopProcessTerminationChild",
        "TestM11ManualStopProcessKillRequiresLockedRecovery",
        "GO_WANT_M11_MANUAL_STOP_PROCESS_TERMINATION",
        "after-journal-rename",
        "after-state-rename",
        "after-marker-rename",
        "syscall.SIGKILL",
        "RECOVERY_REQUIRED",
        "different-stop-reason",
        "bytes.Equal",
    )
    if not all(marker in test for marker in required_test) or "run_learner_bot_test_shard.py" not in workflow or "## Verification" not in evidence:
        fail("M11 direct STOP process-kill regression is missing from the real learner/CI path")


def audit_m10_process_kill_after_canonical_append(root, matrix, plan_text):
    """Keep the later M10 two-store process-boundary proof discoverable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-m10-process-kill-after-canonical-append-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks M10 post-canonical-append process-kill acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M10-CANONICAL-APPEND-PROCESS-KILL-20260916.md",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M10 post-canonical-append process-kill record lacks implementation, test, evidence or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "after canonical append", "RECOVERY_REQUIRED", "locked writer", "exactly one")):
        fail("M10 post-canonical-append record lacks a bounded visibility/recovery disclosure")
    if "Cập nhật M10 process-kill sau canonical append" not in plan_text:
        fail("M10 post-canonical-append process-kill lacks a scoped plan marker")
    mission_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go"
    workflow_path = root / ".github/workflows/curriculum-ci.yml"
    evidence_path = root / "docs/architecture/EVIDENCE-M10-CANONICAL-APPEND-PROCESS-KILL-20260916.md"
    mission = mission_path.read_text(encoding="utf-8") if mission_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    workflow = workflow_path.read_text(encoding="utf-8") if workflow_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_source = (
        "func recoverM10CanaryJournal",
        "func recoverM10CostBoundJournal",
        'm10CanaryFault("after_artifact")',
        'm10CanaryFault("after_state")',
        'm10CostBoundFault("after_index")',
    )
    required_test = (
        "TestMissionM10ProcessKillAfterCanonicalAppendRequiresLockedReplay",
        "GO_M10_PROCESS_TERMINATION_PHASE",
        "canary-after-artifact",
        "canary-after-state",
        "cost-bound-after-artifact",
        "cost-bound-after-index",
        "syscall.SIGKILL",
        "finalArtifactCount",
        "RECOVERY_REQUIRED",
        "EXACT_DUPLICATE",
    )
    if (
        not all(marker in mission for marker in required_source)
        or not all(marker in test for marker in required_test)
        or "run_learner_bot_test_shard.py" not in workflow
        or "## Verification" not in evidence
    ):
        fail("M10 post-canonical-append process-kill regression is missing from the real learner/CI path")


def audit_m10_execution_process_kill_after_canonical_append(root, matrix, plan_text):
    """Keep the later M10 execution two-store process-boundary proof discoverable."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-03-m10-execution-process-kill-after-canonical-append-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks M10 execution post-canonical-append process-kill acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-M10-EXECUTION-CANONICAL-APPEND-PROCESS-KILL-20260916.md",
    }
    refs = set(record.get("implementation_refs", [])) | set(record.get("test_refs", []))
    if required_refs - refs:
        fail("M10 execution post-canonical-append record lacks implementation, test, evidence or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("SIGKILL", "after an M10 governed execution record", "RECOVERY_REQUIRED", "locked retry", "exactly one")):
        fail("M10 execution post-canonical-append record lacks a bounded visibility/recovery disclosure")
    if "Cập nhật M10 execution process-kill sau canonical append" not in plan_text:
        fail("M10 execution post-canonical-append process-kill lacks a scoped plan marker")
    mission_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/m10_process_kill_test.go"
    workflow_path = root / ".github/workflows/curriculum-ci.yml"
    evidence_path = root / "docs/architecture/EVIDENCE-M10-EXECUTION-CANONICAL-APPEND-PROCESS-KILL-20260916.md"
    mission = mission_path.read_text(encoding="utf-8") if mission_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    workflow = workflow_path.read_text(encoding="utf-8") if workflow_path.is_file() else ""
    evidence = evidence_path.read_text(encoding="utf-8") if evidence_path.is_file() else ""
    required_source = (
        "func recoverM10ExecutionJournal",
        'm10ExecutionJournalPublishFault("after_artifact")',
        'm10ExecutionJournalPublishFault("after_state")',
    )
    required_test = (
        "TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay",
        "GO_M10_PROCESS_TERMINATION_KIND=execution",
        "GO_M10_PROCESS_TERMINATION_PHASE=",
        '[]string{"after_artifact", "after_state"}',
        "executionCount",
        "reservation-to-execution",
        "syscall.SIGKILL",
        "RECOVERY_REQUIRED",
        'response["status"] != "APPENDED"',
    )
    if (
        not all(marker in mission for marker in required_source)
        or not all(marker in test for marker in required_test)
        or "run_learner_bot_test_shard.py" not in workflow
        or "## Verification" not in evidence
    ):
        fail("M10 execution post-canonical-append process-kill regression is missing from the real learner/CI path")


def audit_backup_orphan_staging_recovery(root, matrix, plan_text):
    """Keep exact retries from accumulating target-owned staging debris."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-backup-orphan-staging-recovery-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks backup orphan-staging recovery acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/backup_command_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("backup orphan-staging record lacks implementation, test or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("target-owned", "SIGKILL", "POSIX", "other-target")):
        fail("backup orphan-staging record lacks a bounded cleanup disclosure")
    if "Cập nhật backup/restore orphan staging recovery" not in plan_text:
        fail("backup orphan-staging recovery lacks a scoped plan marker")
    source = root / "lab/affiliate-bot/cmd/bot/backup_command.go"
    test = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    workflow = root / ".github/workflows/curriculum-ci.yml"
    source_text = source.read_text(encoding="utf-8") if source.is_file() else ""
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_source = (
        "func backupStagingPrefix(target string) string",
        "func restoreStagingPrefix(target string) string",
        "func cleanupStaleStaging(parent, prefix string) error",
        "cleanupStaleStaging(parent, backupStagingPrefix(args[2]))",
        "cleanupStaleStaging(parent, restoreStagingPrefix(args[2]))",
        "os.ModeSymlink",
        "return syncDirectory(parent)",
    )
    required_test = (
        "TestBackupProcessExitBeforePublishLeavesNoTargetAndRetrySucceeds",
        "TestBackupProcessKillBeforePublishLeavesNoTargetAndRetrySucceeds",
        "TestBackupProcessKillAfterPartialStagingLeavesNoTargetAndRetrySucceeds",
        "TestCleanupStaleStagingOnlyRemovesTargetOwnedDirectories",
        "process termination did not leave exactly one target-owned backup staging tree",
        "backup retry left stale target-owned staging trees",
        "process termination did not leave exactly one target-owned restore staging tree",
        "restore retry left stale target-owned staging trees",
    )
    if not all(token in source_text for token in required_source) or not all(token in test_text for token in required_test) or "GO_BACKUP_PROCESS_TERMINATION_PHASE" not in test_text or "run_learner_bot_test_shard.py" not in workflow_text:
        fail("backup orphan-staging cleanup regression is missing from the real CLI/test/CI path")


def audit_backup_post_publish_process_kill(root, matrix, plan_text):
    """Keep a visible post-rename target safe from an exact retry."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-07-backup-post-publish-process-kill-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks backup post-publish process-kill acceptance record")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/backup_command_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
    }
    if required_refs - set(record.get("implementation_refs", [])) - set(record.get("test_refs", [])):
        fail("backup post-publish process-kill record lacks implementation, test or CI refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or not all(token in scope for token in ("after", "rename", "SIGKILL", "TARGET_NOT_EMPTY", "POSIX")):
        fail("backup post-publish process-kill record lacks a bounded visibility disclosure")
    if "Cập nhật backup/restore post-publish process-kill boundary" not in plan_text:
        fail("backup post-publish process-kill lacks a scoped plan marker")
    test = root / "lab/affiliate-bot/cmd/bot/backup_command_test.go"
    workflow = root / ".github/workflows/curriculum-ci.yml"
    test_text = test.read_text(encoding="utf-8") if test.is_file() else ""
    workflow_text = workflow.read_text(encoding="utf-8") if workflow.is_file() else ""
    required_test = (
        "TestBackupProcessKillAfterPublishLeavesVisibleCompleteTarget",
        "after_rename_before_parent_sync",
        "visible backup after post-rename process termination is invalid",
        "visible restore after post-rename process termination is invalid",
        "visible backup was offered as a retry after post-rename termination",
        "visible restore was offered as a retry after post-rename termination",
    )
    if not all(token in test_text for token in required_test) or "run_learner_bot_test_shard.py" not in workflow_text:
        fail("backup post-publish process-kill regression is missing from the real test/CI path")


def audit_immutable_artifact_parent_recheck(root, matrix, plan_text):
    """Keep hard-link publication from following a swapped parent directory."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-immutable-artifact-parent-recheck")
    if not isinstance(record, dict) or "immediately before hard-link publish" not in record.get("scope", "") or "before_publish seam" not in record.get("scope", ""):
        fail("matrix lacks scoped immutable artifact parent recheck")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/artifact_publish_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("test_refs", [])):
        fail("immutable artifact parent-recheck record lacks publisher and audit regressions")
    if "Cập nhật immutable artifact parent recheck" not in plan_text:
        fail("immutable artifact parent recheck lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/artifact_publish_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    pre_publish = source.partition('if err := artifactWriteFailure("before_publish"); err != nil {')[2].partition("if err := os.Link(temporary, path);")[0]
    if "func requireArtifactOutputDirectory(dir string) error" not in source or "if err := requireArtifactOutputDirectory(dir); err != nil" not in pre_publish or "TestWriteNewJSONRejectsParentSymlinkSwapBeforePublish" not in test or "artifact publisher wrote through swapped parent" not in test:
        fail("immutable artifact parent-recheck regression is missing from the real publisher path")


def audit_advisor_fixture_parent_recheck(root, matrix, plan_text):
    """Keep the disposable advisor bundle inside its selected output parent."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-advisor-fixture-parent-recheck")
    if not isinstance(record, dict) or "immediately before MkdirTemp" not in record.get("scope", ""):
        fail("matrix lacks scoped advisor fixture parent recheck")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("test_refs", [])):
        fail("advisor fixture parent-recheck record lacks CLI and audit regressions")
    if "Cập nhật advisor fixture output-parent guard" not in plan_text:
        fail("advisor fixture parent recheck lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    if "func requireAdvisorFixtureOutputParent(parent string) error" not in source or source.count("requireAdvisorFixtureOutputParent(parent)") < 2 or "TestAdvisorFixtureBundleRejectsParentSymlinkSwapBeforeCreate" not in test or "fixture-run created bundle through swapped parent" not in test:
        fail("advisor fixture parent-recheck regression is missing from the real CLI path")


def audit_advisor_fixture_failed_staging_cleanup(root, matrix, plan_text):
    """Ensure failed disposable advisor bundles do not remain visible."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-advisor-fixture-failed-staging-cleanup")
    if not isinstance(record, dict) or "build/evaluate/write/sync failure" not in record.get("scope", ""):
        fail("matrix lacks scoped advisor fixture failed-staging cleanup")
    required_refs = {
        "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
    }
    if required_refs - set(record.get("test_refs", [])):
        fail("advisor fixture failed-staging record lacks CLI and audit regressions")
    if "Cập nhật advisor fixture failed-staging cleanup" not in plan_text:
        fail("advisor fixture failed-staging cleanup lacks a scoped plan marker")
    source_path = root / "lab/affiliate-bot/cmd/bot/advisor_fixture.go"
    test_path = root / "lab/affiliate-bot/cmd/bot/advisor_fixture_test.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    required_source = (
        "var advisorFixtureBuild = buildBR10AdvisorFixture",
        "c, err := advisorFixtureBuild(dir)",
        "published := false",
        "if !published {",
        "_ = os.RemoveAll(dir)",
        "published = true",
    )
    if any(marker not in source for marker in required_source) or "TestAdvisorFixtureBundleCleansFailedStaging" not in test or "parent still contains failed bundle staging" not in test:
        fail("advisor fixture failed-staging cleanup regression is missing from the real CLI path")


def audit_learner_schema_identity(root, matrix, plan_text):
    """Keep learner M08/M09 state fields on the canonical shared types."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-02-learner-m08-m09-type-identity-20260916")
    if not isinstance(record, dict):
        fail("matrix lacks learner M08/M09 type-identity acceptance record")
    required_impl = {
        "lab/affiliate-bot/cmd/bot/mission_command.go",
        "core/m08/m08.go",
        "core/m09/m09.go",
    }
    if required_impl - set(record.get("implementation_refs", [])):
        fail("learner schema identity record lacks canonical implementation refs")
    required_tests = {
        "lab/affiliate-bot/cmd/bot/learner_schema_alias_test.go",
        "docs/architecture/EVIDENCE-LEARNER-SCHEMA-IDENTITY-20260916.md",
    }
    if required_tests - set(record.get("test_refs", [])):
        fail("learner schema identity record lacks regression/evidence refs")
    scope = record.get("scope")
    if not isinstance(scope, str) or "alias" not in scope or "schema-drift" not in scope:
        fail("learner schema identity record lacks bounded scope disclosure")
    source_path = root / "lab/affiliate-bot/cmd/bot/mission_command.go"
    source = source_path.read_text(encoding="utf-8") if source_path.is_file() else ""
    test_path = root / "lab/affiliate-bot/cmd/bot/learner_schema_alias_test.go"
    test = test_path.read_text(encoding="utf-8") if test_path.is_file() else ""
    required_source = (
        "type LearnerIntent = corem08.Intent",
        "type LearnerPolicy = corem08.PolicyDecision",
        "type LearnerApproval = corem09.ApprovalRecord",
    )
    if any(marker not in source for marker in required_source) or "TestLearnerMissionStateUsesCanonicalM08M09Types" not in test:
        fail("learner schema identity regression is missing from the canonical learner path")


def audit_registry_append_parent_guard(root, matrix, plan_text):
    """Keep M10/M11 registry append parent traversal on the safe path."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-registry-append-parent-openat-20260917")
    if not isinstance(record, dict):
        fail("matrix lacks registry append parent-guard acceptance record")
    required_impl = {
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/stable_append_posix.go",
        "lab/affiliate-bot/cmd/bot/stable_append_other.go",
    }
    required_tests = {
        "lab/affiliate-bot/cmd/bot/registry_parent_swap_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-RP01-REGISTRY-PARENT-OPENAT-20260917.md",
    }
    if required_impl - set(record.get("implementation_refs", [])):
        fail("registry append parent-guard record lacks implementation refs")
    if required_tests - set(record.get("test_refs", [])):
        fail("registry append parent-guard record lacks regression/evidence refs")
    scope = record.get("scope")
    required_scope = ("openat", "O_NOFOLLOW", "M10", "M11", "external", "CI")
    if not isinstance(scope, str) or any(marker not in scope for marker in required_scope):
        fail("registry append parent-guard record lacks bounded scope disclosure")
    if "Cập nhật RP-01 registry append parent openat guard" not in plan_text:
        fail("registry append parent-guard lacks a scoped plan marker")
    for relative in required_impl | required_tests:
        if not (root / relative).is_file():
            fail(f"registry append parent-guard ref is missing: {relative}")
    source = (root / "lab/affiliate-bot/cmd/bot/backup_command.go").read_text(encoding="utf-8")
    posix = (root / "lab/affiliate-bot/cmd/bot/stable_append_posix.go").read_text(encoding="utf-8")
    test = (root / "lab/affiliate-bot/cmd/bot/registry_parent_swap_test.go").read_text(encoding="utf-8")
    required_source = (
        "openStableRegularFileForAppendPath(path, true)",
        "openStableRegularFileForAppendPath(path, false)",
    )
    if any(marker not in source for marker in required_source) or any(marker not in posix for marker in ("unix.Openat", "unix.O_NOFOLLOW", "openedParent", "os.SameFile(openedParent, currentParent)")):
        fail("registry append parent-guard is missing from the real append path")
    if "TestM10RegistryAppendRejectsParentSwapBeforeOpenat" not in test or "TestM11RegistryAppendRejectsParentSwapBeforeOpenat" not in test or "external directory was changed" not in test:
        fail("registry append parent-swap regression is missing from the real M10/M11 paths")


def audit_advisor_campaign_writer_parent_guard(root, matrix, plan_text):
    """Keep the remaining advisor and backup writers on the shared append boundary."""
    updates = {entry.get("id"): entry for entry in matrix.get("recent_updates", []) if isinstance(entry, dict)}
    record = updates.get("RP-01-advisor-campaign-writers-openat-20260917")
    if not isinstance(record, dict):
        fail("matrix lacks advisor campaign-writer parent-guard acceptance record")
    required_impl = {
        "lab/affiliate-bot/cmd/bot/advisor_budget.go",
        "lab/affiliate-bot/cmd/bot/advisor_results.go",
        "lab/affiliate-bot/cmd/bot/advisor_fixture.go",
        "lab/affiliate-bot/cmd/bot/advisor_report.go",
        "lab/affiliate-bot/cmd/bot/advisor_canary.go",
        "lab/affiliate-bot/cmd/bot/backup_command.go",
        "lab/affiliate-bot/cmd/bot/stable_append_posix.go",
        "lab/affiliate-bot/cmd/bot/stable_append_other.go",
    }
    required_tests = {
        "lab/affiliate-bot/cmd/bot/advisor_writer_parent_swap_test.go",
        "scripts/audit_readiness.py",
        "scripts/tests/test_audit_readiness.py",
        ".github/workflows/curriculum-ci.yml",
        "docs/architecture/EVIDENCE-RP01-ADVISOR-CAMPAIGN-WRITERS-OPENAT-20260917.md",
    }
    if required_impl - set(record.get("implementation_refs", [])):
        fail("advisor campaign-writer parent-guard record lacks implementation refs")
    if required_tests - set(record.get("test_refs", [])):
        fail("advisor campaign-writer parent-guard record lacks regression/evidence refs")
    scope = record.get("scope")
    required_scope = ("manifest", "reservation", "result", "fixture", "backup staging", "managed lock", "openat", "O_NOFOLLOW", "external", "CI")
    if not isinstance(scope, str) or any(marker not in scope for marker in required_scope):
        fail("advisor campaign-writer parent-guard record lacks bounded scope disclosure")
    if "Cập nhật RP-01 advisor campaign writers openat guard" not in plan_text:
        fail("advisor campaign-writer parent guard lacks a scoped plan marker")
    for relative in required_impl | required_tests:
        if not (root / relative).is_file():
            fail(f"advisor campaign-writer parent-guard ref is missing: {relative}")
    source_paths = {
        "lab/affiliate-bot/cmd/bot/advisor_budget.go": ("openStableRegularFileForAppendPath", "acquireManagedPathLock"),
        "lab/affiliate-bot/cmd/bot/advisor_results.go": ("openStableRegularFileForAppendPath", "acquireManagedPathLock"),
        "lab/affiliate-bot/cmd/bot/advisor_fixture.go": ("openStableRegularFileForAppendPath",),
        "lab/affiliate-bot/cmd/bot/advisor_report.go": ("acquireManagedPathLock",),
        "lab/affiliate-bot/cmd/bot/advisor_canary.go": ("acquireManagedPathLock",),
        "lab/affiliate-bot/cmd/bot/backup_command.go": ("openStableRegularFileForAppendPath",),
    }
    for relative, markers in source_paths.items():
        source = (root / relative).read_text(encoding="utf-8")
        if any(marker not in source for marker in markers):
            fail(f"advisor campaign-writer parent guard is missing from {relative}")
    posix = (root / "lab/affiliate-bot/cmd/bot/stable_append_posix.go").read_text(encoding="utf-8")
    fallback = (root / "lab/affiliate-bot/cmd/bot/stable_append_other.go").read_text(encoding="utf-8")
    if any(marker not in posix for marker in ("unix.Openat", "unix.O_NOFOLLOW", "openedParent", "os.SameFile(openedParent, currentParent)")) or "os.O_EXCL" not in fallback:
        fail("advisor campaign-writer parent guard is missing from the shared append implementations")
    test = (root / "lab/affiliate-bot/cmd/bot/advisor_writer_parent_swap_test.go").read_text(encoding="utf-8")
    required_tests = (
        "TestAdvisorCampaignInitRejectsParentSwapBeforeManifest",
        "TestAdvisorReservationRejectsParentSwapBeforeOpenat",
        "TestAdvisorResultRejectsParentSwapBeforeOpenat",
        "TestAdvisorFixtureWriterRejectsParentSwapBeforeOpenat",
        "TestBackupStagingWriterRejectsParentSwapBeforeOpenat",
        "external directory was changed",
        "append target was created below the external parent",
    )
    if any(marker not in test for marker in required_tests):
        fail("advisor campaign-writer parent-swap regression is missing from the real writer paths")


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
    audit_m07_raw_json_transport(root, matrix, plan_text)
    audit_n8n_engine_runtime_compatibility(root, matrix, plan_text)
    audit_deterministic_runtime_sharding(root, matrix, plan_text)
    audit_accesstrade_pending_import_recovery(root, matrix, plan_text)
    audit_campaign_result_visible_ack(root, matrix, plan_text)
    audit_committed_journal_cleanup_ack(root, matrix, plan_text)
    audit_m11_post_remove_journal_cleanup_ack(root, matrix, plan_text)
    audit_m10_canary_cost_journal_path_guard(root, matrix, plan_text)
    audit_m11_recovery_admission_approval_guard(root, matrix, plan_text)
    audit_m11_authorization_gate_exact_lineage(root, matrix, plan_text)
    audit_m11_correlation_lineage(root, matrix, plan_text)
    audit_m11_reverse_ledger_graph(root, matrix, plan_text)
    audit_m11_authorization_lineage_mutation(root, matrix, plan_text)
    audit_m11_fixture_outcome_execution_cardinality(root, matrix, plan_text)
    audit_m11_recovery_handoff_strict_decode(root, matrix, plan_text)
    audit_m06_history_handoff_strict_decode(root, matrix, plan_text)
    audit_m06_adapter_visible_append(root, matrix, plan_text)
    audit_backup_target_absence_guard(root, matrix, plan_text)
    audit_backup_process_exit_lock(root, matrix, plan_text)
    audit_managed_lock_path_guard(root, matrix, plan_text)
    audit_m11_process_kill_journal(root, matrix, plan_text)
    audit_m11_outcome_process_kill(root, matrix, plan_text)
    audit_m11_process_kill_after_ledger(root, matrix, plan_text)
    audit_m11_manual_stop_process_kill(root, matrix, plan_text)
    audit_m10_process_kill_after_canonical_append(root, matrix, plan_text)
    audit_m10_execution_process_kill_after_canonical_append(root, matrix, plan_text)
    audit_backup_orphan_staging_recovery(root, matrix, plan_text)
    audit_backup_post_publish_process_kill(root, matrix, plan_text)
    audit_immutable_artifact_parent_recheck(root, matrix, plan_text)
    audit_advisor_fixture_parent_recheck(root, matrix, plan_text)
    audit_advisor_fixture_failed_staging_cleanup(root, matrix, plan_text)
    audit_review_findings(matrix, criteria_by_id, plan_text)
    audit_runtime_acceptance(root, matrix)
    audit_learner_schema_identity(root, matrix, plan_text)
    audit_registry_append_parent_guard(root, matrix, plan_text)
    audit_advisor_campaign_writer_parent_guard(root, matrix, plan_text)
    claim_count = audit_evidence_graph(root, criteria_by_id)
    audit_selected_source_operated_run(root, matrix, plan_text)
    audit_local_recovery_drill(root, matrix, plan_text)
    audit_assisted_fresh_workspace(root, matrix, plan_text)
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
