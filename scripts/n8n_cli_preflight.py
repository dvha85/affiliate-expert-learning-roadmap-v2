"""Fail fast when a JavaScript n8n entrypoint has no Node runtime."""

from pathlib import Path
import shutil
from typing import Optional


def command_prefix(n8n_cli: Optional[str], n8n_node: Optional[str]) -> list[str]:
    if n8n_node:
        if not n8n_cli:
            raise SystemExit("--n8n-node requires --n8n-cli")
        return [n8n_node, n8n_cli]
    if n8n_cli:
        return [n8n_cli]
    return ["n8n"]


def _entrypoint(path: str) -> Optional[Path]:
    candidate = Path(path)
    if candidate.is_file():
        return candidate
    resolved = shutil.which(path)
    return Path(resolved) if resolved else None


def _requires_node(entrypoint: Path) -> bool:
    try:
        with entrypoint.open("rb") as source:
            first_line = source.readline(256).lower()
    except OSError:
        return False
    return first_line.startswith(b"#!") and b"node" in first_line


def validate_n8n_command(prefix: list[str], n8n_cli: Optional[str], n8n_node: Optional[str]) -> None:
    """Validate the executable pair before a disposable runtime is created."""
    if shutil.which(prefix[0]) is None and not Path(prefix[0]).is_file():
        raise SystemExit(f"n8n command is unavailable: {prefix[0]}")
    if len(prefix) == 2 and not Path(prefix[1]).is_file():
        raise SystemExit(f"n8n CLI entrypoint is unavailable: {prefix[1]}")
    if n8n_cli and not n8n_node:
        entrypoint = _entrypoint(n8n_cli)
        if entrypoint and _requires_node(entrypoint) and shutil.which("node") is None:
            raise SystemExit(
                "n8n CLI JavaScript entrypoint requires node on PATH; "
                "pass --n8n-node /absolute/path/to/node"
            )
