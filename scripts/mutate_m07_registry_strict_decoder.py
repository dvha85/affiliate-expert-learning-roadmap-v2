"""Prove the learner reads M07 registry policy with its strict decoder."""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("lab/affiliate-bot/cmd/bot/m07.go")
TEST = "TestLoadM07RegistryRejectsNonCanonicalPolicyJSON"
ASSERTION = "non-canonical registry policy was accepted"
OLD = "contracts.DecodeStrict(raw, &registry)"
NEW = "json.Unmarshal(raw, &registry)"
IMPORT = '\n\t"github.com/dvha85/affiliate-expert-learning-roadmap-v2/contracts"\n'


def main():
    with tempfile.TemporaryDirectory(prefix="m07-strict-registry-mutation-") as directory:
        root = Path(directory)
        for name in ("contracts", "core", "lab/affiliate-bot"):
            source = ROOT / name
            destination = root / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copytree(source, destination)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1 or source.count(IMPORT) != 1:
            raise AssertionError("M07 strict-registry decoder anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1).replace(IMPORT, "", 1), encoding="utf-8")
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
            raise AssertionError("mutated M07 strict-registry decoder unexpectedly passed the real regression")
        if ASSERTION not in output:
            raise AssertionError(
                "M07 strict-registry mutation caused an unrelated failure; "
                f"missing {ASSERTION!r}:\n{output}"
            )
    print("M07 strict registry decoder mutation detected by real learner regression")


if __name__ == "__main__":
    main()
