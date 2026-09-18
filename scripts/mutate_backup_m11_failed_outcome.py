"""Prove backup restore rejects a FAILED M11 execution without its outcome.

This mutation removes the production guard requiring every restored FAILED M11
execution to have its fixture outcome. The real BR-18b backup/restore smoke
must then fail at the checksum-valid snapshot with that outcome removed. This
is bounded offline fixture mutation evidence; it does not prove crash
durability, multi-file atomicity, multi-host locking, provider operation, or
live execution.
"""
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/backup_command.go")
SMOKE = Path("scripts/smoke_br18b_backup_restore.py")
GUARD = '''\t\tif record.Status == "FAILED" && !linked[record.ExecutionID] {
\t\t\treturn fmt.Errorf("failed M11 execution is missing restored fixture outcome")
\t\t}
'''
EXPECTED_FAILURE_MARKER = "missing-failed-outcome"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="backup-m11-failed-outcome-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "examples", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)
        (temp_root / "scripts").mkdir()
        shutil.copy2(ROOT / SMOKE, temp_root / SMOKE)

        target = temp_root / TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            fail("M11 failed-execution outcome guard anchor is missing or ambiguous")
        mutation = "\t\t// MUTATION: M11 failed-execution fixture-outcome enforcement removed.\n"
        target.write_text(source.replace(GUARD, mutation, 1), encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            [sys.executable, str(temp_root / SMOKE)],
            cwd=temp_root,
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated M11 failed-execution outcome guard unexpectedly passed the real BR-18b smoke")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 failed-outcome mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 failed-execution fixture-outcome mutation detected by real BR-18b restore smoke")


if __name__ == "__main__":
    main()
