"""Prove M10/M11 recovery readers retain stable regular-file identity.

The disposable Bot copy replaces the recovery-journal readers with ordinary
path-following reads. The real regression swaps the named journal for an
external symlink with identical bytes after the reader opens it; each mutated
reader must then fail at the dedicated assertion. This narrow proof excludes
crash/power-loss, multi-host filesystems, and other runtime stores.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MISSION_PATH = Path("lab/affiliate-bot/cmd/bot/mission_command.go")
TEST_NAME = "TestRecoveryJournalReadersRejectSymlinkSwapAfterOpen"
ASSERTION = "recovery journal reader accepted a same-byte symlink swap"


def fail(message):
    raise AssertionError(message)


def replace_once(source, old, new, label):
    if source.count(old) != 1:
        fail(f"{label} guard anchor is missing or ambiguous")
    return source.replace(old, new, 1)


def main():
    with tempfile.TemporaryDirectory(prefix="recovery-journal-path-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / MISSION_PATH
        source = target.read_text(encoding="utf-8")
        for old, new, label in (
            (
                """func readM10ExecutionJournal(path string) ([]byte, error) {
\traw, _, err := readStableRegularFile(path)
\tif err != nil {
\t\treturn nil, err
\t}
\treturn raw, nil
}""",
                """func readM10ExecutionJournal(path string) ([]byte, error) {
\treturn os.ReadFile(path)
}""",
                "M10 recovery journal",
            ),
            (
                """func readM11Journal(path string) ([]byte, error) {
\traw, _, err := readStableRegularFile(path)
\tif err != nil {
\t\treturn nil, err
\t}
\treturn raw, nil
}""",
                """func readM11Journal(path string) ([]byte, error) {
\treturn os.ReadFile(path)
}""",
                "M11 recovery journal",
            ),
        ):
            source = replace_once(source, old, new, label)
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
            fail("mutated recovery-journal path guards unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            fail(f"recovery-journal mutation caused an unrelated test failure; missing {ASSERTION!r}:\n{output}")
    print("M10/M11 recovery-journal path guard mutations detected by real Bot regression")


if __name__ == "__main__":
    main()
