"""Prove M08 retains stable identity for a persisted M07 proposal."""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/mission_command.go")
TEST = "TestMissionM08IntentRejectsM07ProposalSymlinkSwapAfterOpen"
ASSERTION = "M08 intent accepted a same-byte M07 proposal symlink swap"
OLD = "\traw, _, err := readStableRegularFile(proposalPath)\n"
NEW = "\traw, err := os.ReadFile(proposalPath)\n"


def main():
    with tempfile.TemporaryDirectory(prefix="m08-m07-proposal-path-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("M08 M07-proposal stable-reader guard anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(["go", "test", "./cmd/bot", "-run", TEST, "-count=1"], cwd=root / "lab/affiliate-bot", env=env, text=True, capture_output=True)
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated M08 M07-proposal guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"M08 M07-proposal mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M08 M07-proposal path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
