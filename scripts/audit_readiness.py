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
    if not isinstance(record, dict) or "six required parallel jobs" not in record.get("scope", ""):
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
        "\n  deterministic-smokes-and-mutations:\n",
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
    audit_n8n_engine_runtime_compatibility(root, matrix, plan_text)
    audit_deterministic_runtime_sharding(root, matrix, plan_text)
    audit_accesstrade_pending_import_recovery(root, matrix, plan_text)
    audit_campaign_result_visible_ack(root, matrix, plan_text)
    audit_m11_recovery_admission_approval_guard(root, matrix, plan_text)
    audit_m11_fixture_outcome_execution_cardinality(root, matrix, plan_text)
    audit_m11_recovery_handoff_strict_decode(root, matrix, plan_text)
    audit_m06_history_handoff_strict_decode(root, matrix, plan_text)
    audit_m06_adapter_visible_append(root, matrix, plan_text)
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
