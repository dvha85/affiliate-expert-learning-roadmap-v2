"""Offline end-to-end continuity smoke; synthetic fixtures only."""
import os
import subprocess
import sys
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
STEPS = [
    "scripts/smoke_quickstart.py",
    "scripts/smoke_br10a.py",
    "scripts/smoke_br10b.py",
    "scripts/smoke_br10c.py",
    "scripts/smoke_br12d.py",
    "scripts/smoke_br13b.py",
]


def main():
    env = dict(os.environ, GOWORK="off", GOCACHE=os.environ.get("GOCACHE", "/private/tmp/br16a-cache"), GO_BIN=os.environ.get("GO_BIN", "/usr/local/go/bin/go"))
    for step in STEPS:
        result = subprocess.run([sys.executable, step], cwd=ROOT, env=env, capture_output=True, text=True)
        if result.returncode:
            raise SystemExit(f"{step} failed: {result.stderr}")
    print("BR-16a PASS: offline continuity smoke M00→M06; synthetic fixtures, restart/replay and failure paths covered")


if __name__ == "__main__":
    main()
