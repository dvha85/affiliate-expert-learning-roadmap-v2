"""Run BOOT in an isolated clone/cache; never edit the source checkout."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--source", default=str(ROOT))
    parser.add_argument("--go", default="go")
    args = parser.parse_args()
    go = shutil.which(args.go)
    if not go:
        parser.error("Go not found; install the toolchain before BOOT")
    with tempfile.TemporaryDirectory(prefix="br05-quickstart-") as temp:
        root = Path(temp)
        checkout = root / "repo"
        env = dict(os.environ, GOCACHE=str(root / "go-cache"), GOMODCACHE=str(root / "go-mod-cache"), GOTOOLCHAIN="local", GIT_TERMINAL_PROMPT="0")

        def run(command, cwd, expected=0, marker=None):
            print("RUN", " ".join(command), flush=True)
            result = subprocess.run(command, cwd=cwd, env=env, text=True, capture_output=True, timeout=300)
            print(result.stdout + result.stderr, end="", flush=True)
            if result.returncode != expected:
                raise RuntimeError(f"expected exit {expected}, got {result.returncode}")
            if marker and marker not in result.stdout + result.stderr:
                raise RuntimeError(f"missing expected output: {marker}")
            return result.stdout

        run(["git", "clone", "--no-hardlinks", "--quiet", args.source, str(checkout)], root)
        run(["git", "rev-parse", "HEAD"], checkout)
        run(["git", "--version"], checkout)
        run([go, "version"], checkout)
        if run(["git", "status", "--porcelain"], checkout).strip():
            raise RuntimeError("fresh clone is not clean")
        bot = checkout / "lab/affiliate-bot"
        run([go, "mod", "download"], bot)
        baseline = run([go, "run", "./cmd/bot"], bot, marker="RANK_SCENARIO")
        for marker in ("Evidence mode): synthetic", "Product B [B]", "score)=9.60", "Product A [A]", "Product C [C]"):
            if marker not in baseline:
                raise RuntimeError(f"quickstart expected output changed: {marker}")
        run([go, "test", "./...", "-count=1"], bot, marker="ok")
        run([go, "vet", "./..."], bot)
        code = bot / "cmd/bot/main.go"
        original = code.read_text(encoding="utf-8")
        good = "return result.Ranked[i].ProductName < result.Ranked[j].ProductName"
        if original.count(good) != 1:
            raise RuntimeError("BOOT tie-break changed; update guide and smoke together")
        print("INTENTIONAL FAIL at cmd/bot/main.go line", original[:original.index(good)].count("\n") + 1, flush=True)
        try:
            code.write_text(original.replace(good, good.replace(" < ", " > ")), encoding="utf-8")
            run([go, "test", "./cmd/bot", "-run", "^TestSameInputSameOutput$", "-count=1"], bot, expected=1, marker="expected deterministic A-first tie break")
        finally:
            code.write_text(original, encoding="utf-8")
        run([go, "test", "./...", "-count=1"], bot, marker="ok")
        for module in ("contracts", "lab/mission-runtime"):
            run([go, "test", "./...", "-count=1"], checkout / module, marker="ok")
        walkthrough = run([go, "run", "./cmd/demo", "O00"], checkout / "lab/mission-runtime", marker="DRY_RUN_ONLY")
        report = json.loads(walkthrough)["result"]
        if report["external_side_effects"] is not False or report["final_state"] != "DRY_RUN_ONLY":
            raise RuntimeError("O00 authority boundary changed")
        if run(["git", "status", "--porcelain"], checkout).strip():
            raise RuntimeError("smoke left unexpected changes")
        print("QUICKSTART SMOKE PASS: isolated clone + empty caches, run/test/intentional FAIL/fix/O00; no Mission PASS", flush=True)


if __name__ == "__main__":
    main()
