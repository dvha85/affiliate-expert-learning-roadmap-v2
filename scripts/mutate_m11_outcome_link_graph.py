"""Prove the core M11 outcome-to-evaluation graph guard cannot be removed.

This mutation removes only the core graph check that requires a checksum-valid
production ledger OutcomeID link to match the offline evaluation reached from
the same execution. The real core graph regression must then fail on its
swapped-outcome-ID fixture. This is bounded offline/read-only mutation
evidence; it does not prove ledger durability, crash or power-loss recovery,
atomic multi-file publication, distributed locking, provider operation, live
execution, or production readiness.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CORE_TARGET = Path("core/m11/artifact_registry.go")
CORE_TEST = "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks"
CORE_GUARD = '''\t// If a ledger outcome and its offline evaluation are both present, the
\t// outcome identity is part of the same immutable link. Keep accepting the
\t// historical ledger-before-evaluation intermediate, but reject a later
\t// evaluation that exposes a checksum-valid outcome-ID swap.
\tfor _, ledger := range ledgers {
\t\tfor _, link := range ledger.OutcomeLinks {
\t\t\tevaluationID, evaluationOK := evaluationExecutions[link.ExecutionID]
\t\t\tif !evaluationOK {
\t\t\t\tcontinue
\t\t\t}
\t\t\tif evaluation := evaluations[evaluationID]; evaluation.OutcomeID != link.OutcomeID {
\t\t\t\treturn fmt.Errorf("production ledger outcome link does not match its evaluation")
\t\t\t}
\t\t}
\t}
'''
EXPECTED_FAILURE_MARKER = "ledger outcome link with a swapped outcome ID was accepted"


def fail(message):
    raise AssertionError(message)


def main():
    with tempfile.TemporaryDirectory(prefix="m11-outcome-link-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / CORE_TARGET
        source = target.read_text(encoding="utf-8")
        if source.count(CORE_GUARD) != 1:
            fail("M11 outcome-link graph guard anchor is missing or ambiguous")
        target.write_text(
            source.replace(
                CORE_GUARD,
                "\t// MUTATION: core M11 outcome-link graph enforcement removed.\n",
                1,
            ),
            encoding="utf-8",
        )

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", "./m11", "-run", f"^{CORE_TEST}$", "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail("mutated M11 outcome-link graph guard unexpectedly passed its focused regression")
        if EXPECTED_FAILURE_MARKER not in output:
            fail(
                "M11 outcome-link mutation caused an unrelated failure; "
                f"missing {EXPECTED_FAILURE_MARKER!r}:\n{output[-6000:]}"
            )
    print("M11 outcome-link graph mutation detected by the real core graph regression")


if __name__ == "__main__":
    main()
