"""Prove the canonical M07 tool-result loader keeps stable path identity."""
import os
import shutil
import subprocess
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
WATCHER = Path("lab/affiliate-bot/cmd/bot/watcher.go")
TEST = "TestM07ToolArtifactRejectsSymlinkSwapAfterOpen"
ASSERTION = "M07 tool loader accepted a same-byte symlink swap"
OLD = """func loadM07ToolArtifact(historyPath, id string, registry []corem07.ToolSpec, recordID string) (corem07.RegisteredToolResult, error) {
\tpath, err := m07ArtifactPath(historyPath, \"tool-results\", id)
\tif err != nil {
\t\treturn corem07.RegisteredToolResult{}, err
\t}
\traw, _, err := readStableRegularFile(path)
\tif err != nil {
\t\treturn corem07.RegisteredToolResult{}, fmt.Errorf(\"registered M07 tool artifact is not a stable regular file: %w\", err)
\t}
\treturn corem07.ValidateRegisteredToolResult(raw, registry, recordID)
}"""
NEW = """func loadM07ToolArtifact(historyPath, id string, registry []corem07.ToolSpec, recordID string) (corem07.RegisteredToolResult, error) {
\tpath, err := m07ArtifactPath(historyPath, \"tool-results\", id)
\tif err != nil {
\t\treturn corem07.RegisteredToolResult{}, err
\t}
\traw, err := os.ReadFile(path)
\tif err != nil {
\t\treturn corem07.RegisteredToolResult{}, err
\t}
\treturn corem07.ValidateRegisteredToolResult(raw, registry, recordID)
}"""


def main():
    with tempfile.TemporaryDirectory(prefix="m07-tool-path-mutation-") as directory:
        root = Path(directory)
        for name in ("core", "contracts", "lab/affiliate-bot"):
            shutil.copytree(ROOT / name, root / name)
        path = root / WATCHER
        source = path.read_text(encoding="utf-8")
        if source.count(OLD) != 1:
            raise AssertionError("M07 tool stable-reader guard anchor is missing or ambiguous")
        path.write_text(source.replace(OLD, NEW, 1), encoding="utf-8")
        env = os.environ.copy()
        env["GOWORK"] = "off"
        result = subprocess.run(["go", "test", "./cmd/bot", "-run", TEST, "-count=1"], cwd=root / "lab/affiliate-bot", env=env, text=True, capture_output=True)
        output = result.stdout + result.stderr
        if result.returncode == 0:
            raise AssertionError("mutated M07 tool path guard unexpectedly passed the real symlink-swap regression")
        if ASSERTION not in output:
            raise AssertionError(f"M07 tool path mutation caused an unrelated failure; missing {ASSERTION!r}:\n{output}")
    print("M07 canonical tool-result path guard mutation detected by real Bot regression")


if __name__ == "__main__":
    main()
