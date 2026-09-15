#!/usr/bin/env python3
"""Run one deterministic, exhaustive shard of learner Bot top-level tests.

The Bot package has a large set of integration-style tests.  Running every
top-level Test* function in one process is correct but serialises independent
temporary-runtime scenarios.  CI invokes this helper once for each shard.  It
discovers the actual tests from Go, assigns each name by a stable SHA-256
digest, and invokes one normal ``go test`` command for its assignment.  Thus a
new test is automatically covered by exactly one shard rather than silently
falling out of CI.
"""

import argparse
import hashlib
from pathlib import Path
import re
import subprocess
import sys


TEST_NAME = re.compile(r"^Test[A-Za-z0-9_]+$")


def select_shard(names, shard_index, shard_count):
    """Return sorted unique valid names assigned to one deterministic shard."""
    if shard_count < 2:
        raise ValueError("shard count must be at least two")
    if not 0 <= shard_index < shard_count:
        raise ValueError("shard index must be in [0, shard count)")
    unique = sorted(set(names))
    if any(not TEST_NAME.fullmatch(name) for name in unique):
        raise ValueError("go test listed an invalid top-level test name")
    return [
        name
        for name in unique
        if int.from_bytes(hashlib.sha256(name.encode("utf-8")).digest()[:8], "big") % shard_count == shard_index
    ]


def list_tests(go, working_directory):
    result = subprocess.run(
        [go, "test", "./cmd/bot", "-list", "^Test"],
        cwd=working_directory,
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode:
        sys.stderr.write(result.stdout + result.stderr)
        raise RuntimeError("could not list learner Bot tests")
    return [line.strip() for line in result.stdout.splitlines() if TEST_NAME.fullmatch(line.strip())]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--shard-index", type=int, required=True)
    parser.add_argument("--shard-count", type=int, required=True)
    parser.add_argument("--go", default="go")
    parser.add_argument(
        "--working-directory",
        default=str(Path(__file__).resolve().parents[1] / "lab" / "affiliate-bot"),
    )
    args = parser.parse_args()
    names = list_tests(args.go, args.working_directory)
    selected = select_shard(names, args.shard_index, args.shard_count)
    if not names or not selected:
        raise RuntimeError("learner Bot test shard is empty")
    pattern = "^(" + "|".join(selected) + ")$"
    print(
        f"learner Bot test shard {args.shard_index + 1}/{args.shard_count}: "
        f"{len(selected)}/{len(set(names))} top-level tests",
        flush=True,
    )
    subprocess.run([args.go, "test", "./cmd/bot", "-run", pattern], cwd=args.working_directory, check=True)


if __name__ == "__main__":
    main()
