"""Prove canonical runtime-store loaders reject external symlink paths.

This narrow composite mutation changes a disposable Bot copy so state/STOP,
M10 registry/cost bounds, and the M11 registry use ordinary path-following
reads. The real runtime-store symlink regression must then fail at every
matching assertion. It does not cover every canonical JSONL store, races,
crashes, multi-host filesystems, or operated runtime evidence.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MISSION_PATH = Path("lab/affiliate-bot/cmd/bot/mission_command.go")
M11_PATH = Path("lab/affiliate-bot/cmd/bot/m11_registry.go")
TEST_NAME = "TestCanonicalRuntimeStoresRejectExternalSymlinkPaths"
ASSERTIONS = (
    "status accepted external mission-state symlink",
    "accepted external M10 artifact registry symlink",
    "accepted external trusted cost-bound registry symlink",
    "accepted external M11 artifact registry symlink",
    "accepted external durable STOP symlink",
)


def fail(message):
    raise AssertionError(message)


def replace_once(path, guard, replacement, label):
    source = path.read_text(encoding="utf-8")
    if source.count(guard) != 1:
        fail(f"{label} guard anchor is missing or ambiguous")
    path.write_text(source.replace(guard, replacement, 1), encoding="utf-8")


def main():
    with tempfile.TemporaryDirectory(prefix="runtime-store-path-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        mission = temp_root / MISSION_PATH
        source = mission.read_text(encoding="utf-8")
        for guard, replacement, label in (
            ("b, _, err := readStableRegularFile(path)", "b, err := os.ReadFile(path)", "canonical state/STOP"),
            ("raw, _, err := readStableRegularFile(m10ArtifactRegistryPath(dir))", "raw, err := os.ReadFile(m10ArtifactRegistryPath(dir))", "M10 registry"),
            ("raw, _, err := readStableRegularFile(trustedCostBoundsPath(dir))", "raw, err := os.ReadFile(trustedCostBoundsPath(dir))", "M10 cost-bound registry"),
        ):
            if source.count(guard) != 1:
                fail(f"{label} guard anchor is missing or ambiguous")
            source = source.replace(guard, replacement, 1)
        mission.write_text(source, encoding="utf-8")

        m11 = temp_root / M11_PATH
        replace_once(
            m11,
            "raw, _, err := readStableRegularFile(m11ArtifactRegistryPath(dir))",
            "raw, err := os.ReadFile(m11ArtifactRegistryPath(dir))",
            "M11 registry",
        )

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./cmd/bot", "-run", TEST_NAME, "-count=1"],
            cwd=temp_root / "lab/affiliate-bot",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated runtime-store path guards unexpectedly passed the real symlink regression")
        missing = [assertion for assertion in ASSERTIONS if assertion not in output]
        if missing:
            fail(f"runtime-store path mutation caused an unrelated test failure; missing {missing}:\n{output}")
    print("canonical runtime-store path composite mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
