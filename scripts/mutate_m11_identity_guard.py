"""Prove each M11 canonical identity guard is exercised by its real Go test.

This is deliberately a narrow mutation test.  It changes a disposable copy of
the implementation, then requires the existing graph test to fail at its
matching forged-identity assertion.  It does not claim broad mutation coverage
or operated runtime evidence.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
GUARD_PATH = Path("core/m11/artifact_registry.go")
MUTATIONS = (
    (
        "gate",
        """\t\t\tif x.GateID != ComputeProductionGateID(lease, x.IntentID, x.IntentHash, snapshot, bound, ledger, x.EvaluatedAt) {
\t\t\t\treturn fmt.Errorf(\"production gate has a non-canonical gate ID\")
\t\t\t}
""",
        "graph accepted a forged production gate ID",
    ),
    (
        "authorization",
        """\t\t\tif x.AuthorizationID != ComputeProductionAuthorizationID(gate.GateID, x.ExecutorID) {
\t\t\t\treturn fmt.Errorf(\"production authorization has a non-canonical authorization ID\")
\t\t\t}
""",
        "graph accepted a forged production authorization ID",
    ),
    (
        "execution",
        """\t\t\tif x.ExecutionID != ComputeProductionExecutionID(auth.AuthorizationID) {
\t\t\t\treturn fmt.Errorf(\"production execution has a non-canonical execution ID\")
\t\t\t}
""",
        "graph accepted a forged production execution ID",
    ),
)


def fail(message):
    raise AssertionError(message)


def run_mutation(name, guard, forged_assertion):
    with tempfile.TemporaryDirectory(prefix="m11-identity-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / GUARD_PATH
        source = target.read_text(encoding="utf-8")
        if source.count(guard) != 1:
            fail(f"canonical M11 {name}-ID guard anchor is missing or ambiguous")
        mutation = f"\t\t\t// MUTATION: {name} ID integrity enforcement removed for regression proof.\n"
        target.write_text(source.replace(guard, mutation, 1), encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m11", "-run", "TestArtifactGraphAcceptsExactProductionLifecycleLinks", "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail(f"mutated M11 {name}-ID guard unexpectedly passed the real graph test")
        if forged_assertion not in output:
            fail(f"{name}-ID mutation caused an unrelated test failure instead of its forged-ID rejection")


def main():
    for name, guard, forged_assertion in MUTATIONS:
        run_mutation(name, guard, forged_assertion)

    print("M11 canonical gate, authorization and execution ID mutations detected by real graph regression")


if __name__ == "__main__":
    main()
