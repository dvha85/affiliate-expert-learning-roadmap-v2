"""Prove shared non-M07/M08-M11 CLI input paths retain stable identity.

The mutation replaces the common bounded reader with os.ReadFile in a
disposable Bot copy. The real M03 action-record command must then accept a
same-byte post-open swap and fail at its dedicated assertion. This is local
filesystem proof only; it does not cover provider, crash, multi-host, or live
business-operation behavior.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/portable_input.go")
TEST = "TestActionStoreRejectsPortableSymlinkSwapAfterOpen"
ASSERTION = "action store accepted a post-open input swap"
OLD = "\tb, _, err := readStableRegularFileLimit(path, maxGeneralPortableInputBytes)\n\treturn b, err\n"
NEW = "\treturn os.ReadFile(path)\n"


def main():
    with tempfile.TemporaryDirectory(prefix="general-portable-input-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("general portable-input stable-reader anchor is missing or ambiguous")
        if source.count("package main\n") != 1:
            raise AssertionError("general portable-input package anchor is missing or ambiguous")
        source = source.replace("package main\n", 'package main\n\nimport "os"\n', 1)
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
            raise AssertionError("mutated general portable-input guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"general portable-input mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("general CLI portable-input path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
