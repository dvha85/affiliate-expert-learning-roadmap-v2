"""Prove M11 validation rejects an execution without its reservation ledger link.

This mutation removes both duplicate offline restore/registry checks for the
same governed execution-to-reservation invariant from a disposable learner Bot
checkout. The focused core graph and backup graph regressions must then fail on
their checksum-valid orphan-reservation fixture. This is bounded offline
mutation evidence; it does not prove crash durability, multi-file atomicity,
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
CORE_TEST = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
BACKUP_TEST = "TestBackupRestoreRejectsExecutionMissingReservationLedger"
CORE_GUARD = '''\t\tif !reserved {
\t\t\treturn fmt.Errorf("production execution is orphaned from its reservation ledger")
\t\t}
'''
BACKUP_GUARD = '''\t\tif !reserved {
\t\t\treturn fmt.Errorf("M11 execution is orphaned from its restored reservation ledger")
\t\t}
'''
EXPECTED_FAILURE_MARKER = "M11 reservation-ledger guard"


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
        core_result = subprocess.run(
            ["go", "test", "./core/m11", "-run", f"^{CORE_TEST}$", "-count=1"],
            cwd=temp_root,
            env=environment,
            text=True,
            capture_output=True,
        )
        if core_result.returncode == 0:
            fail("mutated core M11 reservation-ledger guard unexpectedly passed its focused regression")

        backup_result = subprocess.run(
            ["go", "test", "./lab/affiliate-bot/cmd/bot", "-run", f"^{BACKUP_TEST}$", "-count=1"],
            cwd=temp_root,
            env=environment,
            text=True,
            capture_output=True,
        )
        output = backup_result.stdout + backup_result.stderr
        if backup_result.returncode == 0:
            fail("mutated backup M11 reservation-ledger guard unexpectedly passed its focused regression")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 reservation-ledger mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 reservation-ledger mutation detected by focused core and backup graph regressions")


if __name__ == "__main__":
    main()
