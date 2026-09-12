"""Prove backup graph validation retains stable M07 tool-sidecar identity."""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/backup_command.go")
TEST = "TestM07BackupGraphRejectsToolSidecarSymlinkSwapAfterOpen"
ASSERTION = "backup graph accepted a same-byte M07 tool sidecar symlink swap"
OLD = "\t\traw, _, err := readStableRegularFile(path)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tif parts[0] == \"tool-results\" {"
NEW = "\t\traw, err := os.ReadFile(path)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tif parts[0] == \"tool-results\" {"


def main():
    with tempfile.TemporaryDirectory(prefix="m07-backup-sidecar-path-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("M07 backup tool-sidecar stable-reader guard anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(["go", "test", "./cmd/bot", "-run", TEST, "-count=1"], cwd=root / "lab/affiliate-bot", env=env, text=True, capture_output=True)
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated M07 backup tool-sidecar guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"M07 backup sidecar mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M07 backup graph tool-sidecar path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
