"""Prove each M10 canonical identity guard is exercised by its real Go test.

This bounded mutation regression edits one disposable source copy at a time,
then requires the existing forged-identity assertion in core/m10 to fail.  It
does not claim mutation coverage for mutable ledger state, crash recovery,
multi-host concurrency, or operated execution.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("core/m10/artifact_registry.go")
MUTATIONS = (
    (
        "gate",
        "\t\tif gate.GateID != expectedGateID {\n",
        "\t\tif expectedGateID != expectedGateID {\n",
        "TestCanaryAuthorizationBindsGateWithoutExecuting",
        "graph accepted a forged canary gate ID",
    ),
    (
        "authorization",
        "\t\tif authorization.AuthorizationID != expectedAuthorizationID {\n",
        "\t\tif expectedAuthorizationID != expectedAuthorizationID {\n",
        "TestCanaryAuthorizationBindsGateWithoutExecuting",
        "graph accepted a forged authorization ID",
    ),
    (
        "execution",
        "\t\tif record.ExecutionID != terminalExecutionID(authorization, record.AttemptedAt, record.Status, record.Error) {\n",
        "\t\tif record.ExecutionID != record.ExecutionID {\n",
        "TestCancelledCanaryExecutionRecordHasNoSideEffect",
        "graph accepted a forged execution ID",
    ),
)


def fail(message):
    raise AssertionError(message)


def run_mutation(name, guard, disabled_guard, test_name, forged_assertion):
    with tempfile.TemporaryDirectory(prefix="m10-identity-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if source.count(guard) != 1:
            fail(f"canonical M10 {name}-ID guard anchor is missing or ambiguous")
        target.write_text(source.replace(guard, disabled_guard, 1), encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m10", "-run", test_name, "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail(f"mutated M10 {name}-ID guard unexpectedly passed the real graph test")
        if forged_assertion not in output:
            fail(f"{name}-ID mutation caused an unrelated test failure instead of its forged-ID rejection")


def main():
    for mutation in MUTATIONS:
        run_mutation(*mutation)
    print("M10 canonical gate, authorization and execution ID mutations detected by real graph regressions")


if __name__ == "__main__":
    main()
