"""Prove the real M11 graph test detects a ledger that predates activation.

This mutation removes the production ledger-to-activation timing predicate from
a disposable copy of core/m11, then runs the real graph regression. It is
bounded mutation evidence for the offline validator, not production or
provider evidence.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("core/m11/artifact_registry.go")
TEST_NAME = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
EXPECTED_FAILURE = "ledger before its activation was accepted"
GUARD = "windowStartedAt.Before(activatedAt)"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-ledger-activation-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            fail("M11 ledger activation guard anchor is missing or ambiguous")
        mutated = source.replace(
            GUARD,
            "windowStartedAt.Equal(windowStartedAt) /* MUTATION: ledger activation timing enforcement removed. */",
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
            fail("mutated M11 ledger activation guard unexpectedly passed the real graph test")
        if EXPECTED_FAILURE not in output:
            fail(
                "ledger activation mutation caused an unrelated test failure instead of the timing rejection:\n"
                + output
            )

    print("M11 ledger activation mutation detected by the real graph regression")


if __name__ == "__main__":
    main()
