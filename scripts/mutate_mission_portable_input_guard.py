"""Prove M08-M11 command inputs retain stable pathname identity.

The mutation replaces the shared bounded portable-input reader with os.ReadFile
inside a disposable Bot copy. The real M10 authorization-reservation regression
must then accept the same-byte post-open symlink swap and fail at its dedicated
assertion. This is narrow local filesystem proof, not crash, multi-host or
executor coverage.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/mission_command.go")
TEST = "TestMissionM10ReserveAuthorizationRejectsPortableSymlinkSwapAfterOpen"
ASSERTION = "M10 reserve accepted a post-open authorization swap"
OLD = "\tb, _, err := readStableRegularFileLimit(path, maxMissionPortableInputBytes)\n\treturn b, err\n"
NEW = "\treturn os.ReadFile(path)\n"


def main():
    with tempfile.TemporaryDirectory(prefix="mission-portable-input-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("mission portable-input stable-reader anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./cmd/bot", "-run", TEST, "-count=1"],
            cwd=root / "lab/affiliate-bot",
            env=env,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated mission portable-input guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"mission portable-input mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M08-M11 portable-input path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
