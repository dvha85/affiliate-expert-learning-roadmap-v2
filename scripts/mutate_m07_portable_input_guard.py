"""Prove M07 CLI portable inputs retain stable pathname identity.

The mutation replaces the bounded M07 input reader with os.ReadFile in a
disposable Bot copy. The real CLI regression must then accept a same-byte
post-open tool-result swap and fail at its dedicated assertion. This is narrow
local filesystem proof, not provider, crash, multi-host, or executor coverage.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/m07.go")
TEST = "TestM07RegisterToolResultRejectsPortableSymlinkSwapAfterOpen"
ASSERTION = "M07 registered a post-open tool-result swap"
OLD = "\tb, _, err := readStableRegularFileLimit(path, maxM07PortableInputBytes)\n\treturn b, err\n"
NEW = "\treturn os.ReadFile(path)\n"


def main():
    with tempfile.TemporaryDirectory(prefix="m07-portable-input-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("M07 portable-input stable-reader anchor is missing or ambiguous")
        if source.count('\t"io"\n') != 1:
            raise AssertionError("M07 import anchor is missing or ambiguous")
        source = source.replace('\t"io"\n', '\t"io"\n\t"os"\n', 1)
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
            raise AssertionError("mutated M07 portable-input guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"M07 portable-input mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M07 portable-input path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
