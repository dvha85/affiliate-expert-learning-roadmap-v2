"""Prove the real M11 graph test detects removal of exact gate lineage checks.

This mutation removes the complete authorization-to-gate artifact-lineage
guard from a disposable copy of core/m11, then runs the real graph regression.
It is bounded mutation evidence for the offline validator, not production or
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
EXPECTED_FAILURE = "authorization switched to health/cost artifacts not evaluated by its gate"
LINEAGE_GUARDS = (
    " || gate.LeaseID != x.ProductionLeaseID",
    " || gate.LeaseVersion != x.ProductionLeaseVersion",
    " || gate.LeaseHash != x.ProductionLeaseHash",
    " || gate.HealthSnapshotID != x.ProductionHealthSnapshotID",
    " || gate.HealthSnapshotHash != x.ProductionHealthSnapshotHash",
    " || gate.CostBoundID != x.ProductionCostBoundID",
    " || gate.CostBoundHash != x.ProductionCostBoundHash",
    " || gate.CostBoundMinor != x.ProductionCostBoundMinor",
)


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-lineage-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if any(source.count(guard) != 1 for guard in LINEAGE_GUARDS):
            fail("M11 authorization gate lineage guard anchor is missing or ambiguous")
        mutation = "\n\t\t\t// MUTATION: authorization-to-gate exact lineage enforcement removed.\n"
        mutated = source
        for guard in LINEAGE_GUARDS:
            mutated = mutated.replace(guard, "", 1)
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
            fail("mutated M11 authorization gate lineage guard unexpectedly passed the real graph test")
        if EXPECTED_FAILURE not in output:
            fail("lineage mutation caused an unrelated test failure instead of the forged-lineage rejection")

    print("M11 authorization-to-gate exact lineage mutation detected by the real graph regression")


if __name__ == "__main__":
    main()
