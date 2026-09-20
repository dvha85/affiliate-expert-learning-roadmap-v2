"""Prove the core M11 cycle graph guard cannot be removed.

This mutation removes only the core graph check that requires a CLOSED M11
cycle to retain the exact evaluation outcome and execution lineage. The real
core graph regression must then fail on its mismatched-evaluation-outcome
fixture. This is bounded offline/read-only mutation evidence; it does not
prove ledger durability, crash or power-loss recovery, atomic multi-file
publication, distributed locking, provider operation, live execution, or
production readiness.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CORE_TARGET = Path("core/m11/artifact_registry.go")
CORE_TEST = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
CYCLE_GUARD = '''\t\tcase *ProductionCycleRecord:
\t\t\tlease, leaseOK := leases[x.LeaseID]
\t\t\tgate, gateOK := gates[x.GateID]
\t\t\tauth, authOK := authorizations[x.AuthorizationID]
\t\t\texecution, executionOK := executions[x.ExecutionID]
\t\t\tevaluation, evaluationOK := evaluations[x.EvaluationID]
\t\t\tclosedAt, closedErr := time.Parse(time.RFC3339, x.ClosedAt)
\t\t\tevaluatedAt, evaluatedErr := time.Parse(time.RFC3339, evaluation.EvaluatedAt)
\t\t\tif !leaseOK || !gateOK || !authOK || !executionOK || !evaluationOK || closedErr != nil || evaluatedErr != nil || x.Status != "CLOSED" || x.CorrelationID != lease.CorrelationID || lease.LeaseVersion != x.LeaseVersion || lease.LeaseHash != x.LeaseHash || gate.IntentID != x.IntentID || gate.IntentHash != x.IntentHash || auth.AuthorizationID != x.AuthorizationID || auth.ProductionGateID != x.GateID || execution.AuthorizationID != x.AuthorizationID || execution.ProductionGateID != x.GateID || execution.ExecutionID != x.ExecutionID || execution.AttemptedAt != x.OpenedAt || execution.CorrelationID != x.CorrelationID || evaluation.LeaseID != x.LeaseID || evaluation.ExecutionID != x.ExecutionID || evaluation.OutcomeID != x.OutcomeID || closedAt.Before(evaluatedAt) {
\t\t\t\treturn fmt.Errorf("production cycle has an orphaned or mismatched link")
\t\t\t}
'''
EXPECTED_FAILURE_MARKER = "cycle with mismatched evaluation outcome was accepted"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-cycle-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / CORE_TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(CYCLE_GUARD) != 1:
            fail("M11 cycle graph guard anchor is missing or ambiguous")
        target.write_text(
            source.replace(
                CYCLE_GUARD,
                "\t\tcase *ProductionCycleRecord:\n\t\t\t// MUTATION: core M11 cycle graph enforcement removed.\n",
                1,
            ),
            encoding="utf-8",
        )

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m11", "-run", f"^{CORE_TEST}$", "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated M11 cycle graph guard unexpectedly passed its focused regression")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 cycle mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 cycle graph mutation detected by the real core graph regression")


if __name__ == "__main__":
    main()
