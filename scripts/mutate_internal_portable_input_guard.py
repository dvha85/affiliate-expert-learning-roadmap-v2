"""Prove learner internal portable inputs retain stable pathname identity.

The mutation makes the internal shared reader fall back to os.ReadFile in a
disposable learner copy. Its real post-open same-byte symlink regression must
then fail at the dedicated assertion. This is local filesystem proof, not a
provider, crash, multi-host, or live-operation claim.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/internal/store/history.go")
TEST = "TestReadPortableInputRejectsSameByteSymlinkSwapAfterOpen"
ASSERTION = "portable reader accepted a post-open same-byte symlink swap"
OLD = "\treturn ReadStableRegularFileLimit(path, MaxPortableInputBytes)\n"
NEW = "\treturn os.ReadFile(path)\n"


def main():
    with tempfile.TemporaryDirectory(prefix="internal-portable-input-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("internal portable-input stable-reader anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./internal/store", "-run", TEST, "-count=1"],
            cwd=root / "lab/affiliate-bot",
            env=env,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated internal portable-input guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"internal portable-input mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("internal learner portable-input path guard mutation detected by real regression")


if __name__ == "__main__":
    main()
