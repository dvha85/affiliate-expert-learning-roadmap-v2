"""Prove a repeated M11 STOP cannot overwrite the durable STOP record.

The mutation removes the real command's existing-STOP guard in a disposable
Bot checkout. Its byte-identity regression must then observe that a second
stop rewrites the reason and fail at its dedicated assertion. This is local
runtime proof only; it does not cover multi-file crash recovery or locking
across hosts.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/mission_command.go")
TEST = "TestMissionStopCannotOverwriteDurableReason"
ASSERTION = "second stop did not fail closed"
GUARD = '''\tcase "m11-stop", "stop":
\t\tif len(args) != 3 {
\t\t\treturn emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-stop STATE_DIR REASON"), 2)
\t\t}
\t\ts, err := loadMissionState(args[1])
\t\tif err != nil {
\t\t\treturn emit("STATE_ERROR", nil, err, 1)
\t\t}
\t\tif s.Stop {
\t\t\treturn emit("STOPPED", s, fmt.Errorf("durable STOP: %s", s.StopReason), 1)
\t\t}
'''
DISABLED_GUARD = '''\tcase "m11-stop", "stop":
\t\tif len(args) != 3 {
\t\t\treturn emit("USAGE_ERROR", nil, fmt.Errorf("usage: bot mission m11-stop STATE_DIR REASON"), 2)
\t\t}
\t\ts, err := loadMissionState(args[1])
\t\tif err != nil {
\t\t\treturn emit("STATE_ERROR", nil, err, 1)
\t\t}
'''


def main():
    with tempfile.TemporaryDirectory(prefix="mission-stop-immutability-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(GUARD) != 1:
            raise AssertionError("durable STOP immutability guard anchor is missing or ambiguous")
        path.write_text(source.replace(GUARD, DISABLED_GUARD, 1), encoding="utf-8")
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
            raise AssertionError("mutated durable STOP immutability guard unexpectedly passed the real Bot regression")
        if ASSERTION not in output:
            raise AssertionError(f"STOP mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("durable STOP immutability mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
