"""Prove the real M11 graph test detects removal of gate budget snapshots.

This mutation removes the execution-counter snapshot comparison from a
disposable copy of core/m11, then runs the real graph regression. It is bounded
mutation evidence for the offline validator, not production or provider
evidence.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("core/m11/artifact_registry.go")
TEST_NAME = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
EXPECTED_FAILURE = "gate with drifted ledger counters was accepted"
GUARD = " || ledgerState.ExecutionsTotal != x.ExecutionsTotalBefore ||"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-budget-snapshot-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            fail("M11 gate execution-counter snapshot guard anchor is missing or ambiguous")
        mutated = source.replace(
            GUARD,
            " ||\n\t\t\t// MUTATION: gate execution-counter snapshot enforcement removed.\n\t\t\t",
            1,
        )
        target.write_text(mutated, encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m11", "-run", TEST_NAME, "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated M11 gate budget snapshot guard unexpectedly passed the real graph test")
        if EXPECTED_FAILURE not in output:
            fail("budget snapshot mutation caused an unrelated test failure instead of the drift rejection")

    print("M11 gate budget snapshot mutation detected by the real graph regression")


if __name__ == "__main__":
    main()
