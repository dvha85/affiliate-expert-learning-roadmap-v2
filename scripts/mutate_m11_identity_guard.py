"""Prove the M11 canonical gate-ID guard is exercised by its real Go test.

This is deliberately a narrow mutation test.  It changes a disposable copy of
the implementation, then requires the existing graph test to fail at its
forged-gate assertion.  It does not claim broad mutation coverage or operated
runtime evidence.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("core/m11/artifact_registry.go")
GUARD = """\t\t\tif x.GateID != ComputeProductionGateID(lease, x.IntentID, x.IntentHash, snapshot, bound, ledger, x.EvaluatedAt) {
\t\t\t\treturn fmt.Errorf(\"production gate has a non-canonical gate ID\")
\t\t\t}
"""
MUTATION = """\t\t\t// MUTATION: gate ID integrity enforcement removed for regression proof.
"""
FORGED_ASSERTION = "graph accepted a forged production gate ID"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-identity-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            fail("canonical M11 gate-ID guard anchor is missing or ambiguous")
        target.write_text(source.replace(GUARD, MUTATION, 1), encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m11", "-run", "TestArtifactGraphAcceptsExactProductionLifecycleLinks", "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated M11 gate-ID guard unexpectedly passed the real graph test")
        if FORGED_ASSERTION not in output:
            fail("mutation caused an unrelated test failure instead of the forged-gate rejection")

    print("M11 canonical gate-ID mutation detected by real graph regression")


if __name__ == "__main__":
    main()
