"""Prove M07 validates the model payload with its strict decoder."""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
TARGET = Path("core/m07/m07.go")
TEST = "TestValidateAgentOutputRejectsNonCanonicalJSONShape"
ASSERTION = "non-canonical model JSON was accepted"
OLD = '''\tif err := contracts.DecodeStrict(raw, &output); err != nil {
\t\treturn AgentOutput{}, fmt.Errorf("agent output schema: %w", err)
\t}
'''
NEW = '''\tif err := json.Unmarshal(raw, &output); err != nil {
\t\treturn AgentOutput{}, fmt.Errorf("agent output schema: %w", err)
\t}
'''


def main():
    with tempfile.TemporaryDirectory(prefix="m07-strict-output-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts"):
            shutil.copytree(ROOT / name, root / name)
        path = root / TARGET
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("M07 strict-output decoder anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(["go", "test", "./m07", "-run", TEST, "-count=1"], cwd=root / "core", env=env, text=True, capture_output=True)
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated M07 strict-output decoder unexpectedly passed the real regression")
        if ASSERTION not in output:
            raise AssertionError(f"M07 strict-output mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M07 strict output decoder mutation detected by real core regression")


if __name__ == "__main__":
    main()
