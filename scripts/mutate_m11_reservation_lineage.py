"""Prove M11 restore rejects an execution without its reservation ledger link.

This mutation removes both duplicate offline restore/registry checks for the
same governed execution-to-reservation invariant from a disposable learner
Bot checkout. The real BR-18b backup/restore smoke must then fail at the
checksum-valid orphan-reservation fixture. This is bounded offline mutation
evidence; it does not prove crash durability, multi-file atomicity,
multi-host locking, provider operation, or live execution.
"""
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CORE_TARGET = Path("core/m11/artifact_registry.go")
BACKUP_TARGET = Path("lab/affiliate-bot/cmd/bot/backup_command.go")
SMOKE = Path("scripts/smoke_br18b_backup_restore.py")
CORE_GUARD = '''\t\tif !reserved {
\t\t\treturn fmt.Errorf("production execution is orphaned from its reservation ledger")
\t\t}
'''
BACKUP_GUARD = '''\t\tif !reserved {
\t\t\treturn fmt.Errorf("M11 execution is orphaned from its restored reservation ledger")
\t\t}
'''
EXPECTED_FAILURE_MARKER = "orphan-m11-reservation guard"


def fail(message):
    raise AssertionError(message)


def remove_guard(path, guard, mutation_comment):
    source = path.read_text(encoding="utf-8")
    if source.count(guard) != 1:
        fail(f"reservation guard anchor is missing or ambiguous in {path}")
    path.write_text(source.replace(guard, mutation_comment, 1), encoding="utf-8")


def main():
    with tempfile.TemporaryDirectory(prefix="m11-reservation-lineage-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "examples", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)
        (temp_root / "scripts").mkdir()
        shutil.copy2(ROOT / SMOKE, temp_root / SMOKE)

        remove_guard(
            temp_root / CORE_TARGET,
            CORE_GUARD,
            "\t\t// MUTATION: core M11 reservation-ledger enforcement removed.\n",
        )
        remove_guard(
            temp_root / BACKUP_TARGET,
            BACKUP_GUARD,
            "\t\t// MUTATION: backup restore reservation-ledger enforcement removed.\n",
        )

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
            fail("mutated M11 reservation-ledger guards unexpectedly passed the real BR-18b smoke")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 reservation-ledger mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 reservation-ledger mutation detected by real BR-18b restore smoke")


if __name__ == "__main__":
    main()
