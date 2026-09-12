"""Prove backup refuses a same-byte source symlink swap in the real Bot test.

This deliberately narrow composite mutation removes all stable regular-file
identity checks in a disposable source copy. The real source-swap regression
must then fail at its dedicated assertion. It does not claim race, crash,
multi-host, or operated-runtime mutation coverage.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("lab/affiliate-bot/cmd/bot/backup_command.go")
TEST_NAME = "TestBackupRejectsSourceSymlinkSwapAfterInventory"
TEST_ASSERTION = "backup followed source symlink swapped after inventory"
MUTATIONS = (
    (
        "pre-open regular-file check",
        "if !before.Mode().IsRegular() {",
        "if !before.Mode().IsRegular() && false { // MUTATION: pre-open regular-file check removed for regression proof.",
    ),
    (
        "open identity check",
        "if !opened.Mode().IsRegular() || !os.SameFile(before, opened) {",
        "if (!opened.Mode().IsRegular() || !os.SameFile(before, opened)) && false { // MUTATION: open identity check removed for regression proof.",
    ),
    (
        "post-read identity check",
        "if err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after) {",
        "if (err != nil || !after.Mode().IsRegular() || !os.SameFile(opened, after)) && false { // MUTATION: post-read identity check removed for regression proof.",
    ),
)


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="backup-source-guard-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        for name, guard, disabled_guard in MUTATIONS:
            if source.count(guard) != 1:
                fail(f"backup {name} guard anchor is missing or ambiguous")
            source = source.replace(guard, disabled_guard, 1)
        target.write_text(source, encoding="utf-8")

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
            fail("mutated backup source identity guards unexpectedly passed the real source-swap test")
        if TEST_ASSERTION not in output:
            fail(f"backup source-identity mutation caused an unrelated test failure:\n{output}")
    print("backup source identity composite mutation detected by the real Bot regression")


if __name__ == "__main__":
    main()
