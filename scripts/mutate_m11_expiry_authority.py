"""Prove the canonical M11 gate rejects an exact lease expiry.

This mutation removes the strict lease-expiry comparison from a disposable
copy of core/m11, then runs the real canonical graph regression. The proof is
bounded offline mutation evidence for the graph validator; it does not cover
trusted external clocks, live execution, crash durability, or multi-host
coordination.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("core/m11/artifact_registry.go")
TEST = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
GUARD = "!evaluatedAt.Before(expiresAt)"
ASSERTION = "canonical gate at lease expiry was accepted"


def main():
    with tempfile.TemporaryDirectory(prefix="m11-expiry-authority-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            raise AssertionError("M11 lease-expiry guard anchor is missing or ambiguous")
        # Keep the parsed expiry variable live so the disposable copy remains
        # compile-valid while the strict comparison is semantically disabled.
        target.write_text(source.replace(GUARD, "expiresAt.Before(expiresAt)", 1), encoding="utf-8")

        result = subprocess.run(
            ["go", "test", "./m11", "-run", TEST, "-count=1"],
            cwd=temp_root / "core",
            env={**os.environ, "GOWORK": "off"},
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated M11 lease-expiry guard unexpectedly passed the real graph test")
        if ASSERTION not in output:
            raise AssertionError(
                "lease-expiry mutation caused an unrelated test failure; "
                f"missing {ASSERTION!r}:\n{output}"
            )

    print("M11 exact lease-expiry mutation detected by the real graph regression")


if __name__ == "__main__":
    main()
