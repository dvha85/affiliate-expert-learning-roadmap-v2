"""Prove the core M11 reverse-ledger graph guard cannot be removed.

This mutation removes only the core graph check that requires a checksum-valid
ledger reconciliation link to resolve to the same stopped execution and lease.
The real core graph regression must then fail on its orphan-resolution fixture.
This is bounded offline/read-only mutation evidence; it does not prove ledger
durability, crash or power-loss recovery, atomic multi-file publication,
distributed locking, provider operation, live execution, or production
readiness.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CORE_TARGET = Path("core/m11/artifact_registry.go")
CORE_TEST = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
CORE_GUARD = '''\t// A resolved UNKNOWN transition is published as a new stopped-ledger head
\t// after the resolution artifact. The append-only registry therefore permits
\t// the resolution-only intermediate, but any ledger that claims the
\t// resolution must resolve it back to the same execution and lease. Without
\t// this reverse check a checksum-valid ledger could invent a resolution ID or
\t// attach a real resolution from another lifecycle.
\tfor _, ledger := range ledgers {
\t\tseenResolutionIDs := map[string]bool{}
\t\tfor _, resolutionID := range ledger.ReconciliationResolutionIDs {
\t\t\tif seenResolutionIDs[resolutionID] {
\t\t\t\treturn fmt.Errorf("production ledger has duplicate reconciliation resolution link")
\t\t\t}
\t\t\tseenResolutionIDs[resolutionID] = true
\t\t\tresolution, resolutionOK := resolutionsByID[resolutionID]
\t\t\texecution, executionOK := executions[resolution.ExecutionID]
\t\t\tif !resolutionOK || !executionOK || resolution.LeaseID != ledger.LeaseID || resolution.LeaseVersion != ledger.LeaseVersion || resolution.LeaseHash != ledger.LeaseHash || execution.ProductionLeaseID != ledger.LeaseID || execution.ProductionLeaseVersion != ledger.LeaseVersion || execution.ProductionLeaseHash != ledger.LeaseHash || ledger.ControlMode != "STOPPED" || ledger.ReconciliationRequired || ledger.StopReason != "RECOVERY_REVIEW_REQUIRED" {
\t\t\t\treturn fmt.Errorf("production ledger reconciliation link is orphaned or mismatched")
\t\t\t}
\t\t}
\t}
'''
EXPECTED_FAILURE_MARKER = "ledger with an orphan reconciliation resolution link was accepted"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-reverse-ledger-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / CORE_TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(CORE_GUARD) != 1:
            fail("M11 reverse-ledger graph guard anchor is missing or ambiguous")
        target.write_text(
            source.replace(
                CORE_GUARD,
                "\t// MUTATION: core M11 reverse-ledger graph enforcement removed.\n",
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
            fail("mutated M11 reverse-ledger graph guard unexpectedly passed its focused regression")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 reverse-ledger mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 reverse-ledger graph mutation detected by the real core graph regression")


if __name__ == "__main__":
    main()
