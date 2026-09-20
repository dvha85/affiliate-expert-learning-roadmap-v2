"""Decide whether a change needs the disposable n8n engine regression.

The workflow keeps the expensive engine job scoped on pull requests, but the
scope must follow the actual dependency boundary rather than a hand-maintained
list of individual adapter files.  GitHub passes changed repository-relative
paths on stdin, one per line; the command exits zero when engine coverage is
needed and one otherwise.
"""

from __future__ import annotations

import sys
from typing import Iterable


_PATH_PREFIXES = (
    "lab/n8n/",
    "contracts/",
    "core/",
    "lab/affiliate-bot/cmd/bot/",
    "lab/affiliate-bot/internal/",
)
_EXACT_PATHS = {
    ".github/workflows/curriculum-ci.yml",
    ".github/workflows/mission-agent-path-ci.yml",
    "lab/mission-runtime/go.mod",
    "lab/mission-runtime/go.sum",
    "lab/affiliate-bot/go.mod",
    "lab/affiliate-bot/go.sum",
}
_N8N_SCRIPT_PREFIXES = (
    "scripts/n8n_",
    "scripts/run_n8n_",
    "scripts/validate_n8n_",
)


def requires_n8n_engine(paths: Iterable[str]) -> bool:
    """Return whether any changed path can affect the n8n integration."""

    for raw_path in paths:
        path = raw_path.strip()
        while path.startswith("./"):
            path = path[2:]
        if not path:
            continue
        if path in _EXACT_PATHS or any(path.startswith(prefix) for prefix in _PATH_PREFIXES):
            return True
        if any(path.startswith(prefix) and path.endswith(".py") for prefix in _N8N_SCRIPT_PREFIXES):
            return True
    return False


def main() -> int:
    return 0 if requires_n8n_engine(sys.stdin) else 1


if __name__ == "__main__":
    raise SystemExit(main())
