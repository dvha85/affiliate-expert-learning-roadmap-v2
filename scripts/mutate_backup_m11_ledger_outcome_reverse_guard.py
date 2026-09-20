"""Prove backup restore cannot lose the M11 ledger-to-outcome reverse link.

This mutation removes the reverse ledger-outcome cardinality guard from a
disposable learner Bot checkout. The real BR-18b backup/restore smoke must then
fail at its checksum-valid snapshot with the ledger outcome link removed. This
is bounded offline fixture mutation evidence; it does not prove crash
durability, multi-file atomicity, multi-host locking, provider operation, live
execution, or business outcome.
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
GUARD = '''\tfor outcomeID := range outcomesByID {
\t\tif !ledgerOutcomeLinks[outcomeID] {
\t\t\treturn fmt.Errorf("M11 fixture outcome is absent from its restored ledger")
\t\t}
\t}
'''
EXPECTED_FAILURE_MARKER = "orphan-ledger-outcome-restored"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="backup-m11-ledger-outcome-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "examples", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)
        (temp_root / "scripts").mkdir()
        shutil.copy2(ROOT / SMOKE, temp_root / SMOKE)

        target = temp_root / TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            fail("M11 ledger-outcome reverse cardinality guard anchor is missing or ambiguous")
        mutation = "\t// MUTATION: M11 ledger-outcome reverse cardinality enforcement removed.\n"
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
            fail("mutated M11 ledger-outcome reverse guard unexpectedly passed the real BR-18b smoke")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 ledger-outcome mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 ledger-outcome reverse mutation detected by real BR-18b restore smoke")


if __name__ == "__main__":
    main()
