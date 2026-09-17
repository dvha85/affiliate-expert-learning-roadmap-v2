"""Prove direct M10/M11 graph callers cannot bypass envelope validation.

This bounded mutation regression removes the public graph envelope check in a
disposable source copy, then requires the matching real core regression to
fail on forged artifact metadata or non-canonical bytes.  It does not claim
mutation coverage for mutable ledgers, crash recovery, multi-host locking,
provider operation, or live execution.
"""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MUTATIONS = (
    (
        "M10",
        Path("core/m10/artifact_registry.go"),
        "TestCanaryAuthorizationBindsGateWithoutExecuting",
        "graph accepted a forged M10 registry",
    ),
    (
        "M11",
        Path("core/m11/artifact_registry.go"),
        "TestArtifactGraphAcceptsAndRejectsExactProductionLifecycleLinks",
        "graph accepted a forged M11 registry",
    ),
)


def fail(message):
    raise AssertionError(message)


def remove_graph_envelope_guard(source, module):
    signature = "func ValidateArtifactGraph(entries []ArtifactEntry)"
    if source.count(signature) != 1:
        fail(f"canonical {module} graph function anchor is missing or ambiguous")
    prefix, graph = source.split(signature, 1)
    guard = (
        "if entry.ArtifactID != expected.ArtifactID || entry.ContentHash != expected.ContentHash || "
        "!bytes.Equal(entry.Artifact, expected.Artifact)"
    )
    if graph.count(guard) != 1:
        fail(f"canonical {module} graph envelope guard anchor is missing or ambiguous")
    graph = graph.replace(guard, 'if false && expected.ArtifactID == ""', 1)
    return prefix + signature + graph


def run_mutation(module, relative, test_name, forged_assertion):
    with tempfile.TemporaryDirectory(prefix="registry-graph-envelope-mutation-") as temp_dir:
        temp_root = Path(temp_dir)
        for directory in ("core", "contracts"):
            shutil.copytree(ROOT / directory, temp_root / directory)

        target = temp_root / relative
        target.write_text(remove_graph_envelope_guard(target.read_text(encoding="utf-8"), module), encoding="utf-8")

        environment = os.environ.copy()
        environment["GOWORK"] = "off"
        result = subprocess.run(
            ["go", "test", f"./{relative.parts[1]}", "-run", test_name, "-count=1"],
            cwd=temp_root / "core",
            env=environment,
            text=True,
            capture_output=True,
        )
        output = result.stdout + result.stderr
        if result.returncode == 0:
            fail(f"mutated {module} graph envelope guard unexpectedly passed the real graph test")
        if forged_assertion not in output:
            diagnostic = output[-4000:].strip()
            fail(
                f"{module} envelope mutation caused an unrelated test failure instead of its forged-envelope rejection; "
                f"go test output tail:\n{diagnostic}"
            )


def main():
    for mutation in MUTATIONS:
        run_mutation(*mutation)
    print("M10 and M11 registry graph envelope mutations detected by real graph regressions")


if __name__ == "__main__":
    main()
