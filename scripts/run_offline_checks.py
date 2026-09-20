"""Run the complete offline/read-only verification contract from a clean checkout.

This entrypoint deliberately excludes operated validators and the real n8n
engine.  Those commands need an execution artifact or a pinned external
runtime and are documented as separate acceptance steps.
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Optional, Sequence


ROOT = Path(__file__).resolve().parents[1]

GO_MODULES = (
    "contracts",
    "core",
    "lab/affiliate-bot",
    "lab/mission-runtime",
)

OFFLINE_VALIDATORS = (
    "scripts/validate_repo.py",
    "scripts/validate_missions.py",
    "scripts/validate_artifact_spine.py",
    "scripts/validate_continuity.py",
    "scripts/validate_language_policy.py",
    "scripts/validate_agent_semantics.py",
    "scripts/validate_semantic_contracts.py",
    "scripts/validate_m11.py",
    "scripts/validate_n8n_m06.py",
    "scripts/validate_n8n_m06_cases.py",
    "scripts/validate_n8n_m06_selected_source.py",
    "scripts/validate_n8n_m07.py",
    "scripts/validate_n8n_m07_adversarial.py",
    "scripts/validate_n8n_m07_output_cases.py",
)

OPERATED_VALIDATORS = (
    "scripts/validate_n8n_m06_operated_execution.py",
    "scripts/validate_n8n_m07_operated_execution.py",
)

SMOKE_SCRIPTS = (
    "scripts/smoke_br08.py",
    "scripts/smoke_br09.py",
    "scripts/smoke_br10a.py",
    "scripts/smoke_br10b.py",
    "scripts/smoke_br10c.py",
    "scripts/smoke_br10d_accesstrade.py",
    "scripts/smoke_br11a.py",
    "scripts/smoke_br12d.py",
    "scripts/smoke_br13b.py",
    "scripts/smoke_br16a_offline.py",
    "scripts/smoke_br18b_backup_restore.py",
)


@dataclass(frozen=True)
class CommandStep:
    label: str
    command: tuple[str, ...]
    cwd: Path
    environment: tuple[tuple[str, str], ...] = ()


def _python_script(root: Path, relative_script: str) -> tuple[str, ...]:
    return (sys.executable, str(root / relative_script))


def build_command_plan(root: Path = ROOT) -> tuple[CommandStep, ...]:
    """Return the ordered offline verification plan without executing it."""

    root = root.resolve()
    steps: list[CommandStep] = []
    gowork_off = (("GOWORK", "off"),)

    for module in GO_MODULES:
        module_root = root / module
        steps.append(
            CommandStep(
                label=f"Go test {module}",
                command=("go", "test", "-count=1", "./..."),
                cwd=module_root,
                environment=gowork_off,
            )
        )
        steps.append(
            CommandStep(
                label=f"Go vet {module}",
                command=("go", "vet", "./..."),
                cwd=module_root,
                environment=gowork_off,
            )
        )

    steps.append(
        CommandStep(
            label="Learner Bot race regression",
            command=("go", "test", "-race", "-count=1", "./..."),
            cwd=root / "lab/affiliate-bot",
            environment=gowork_off,
        )
    )
    steps.append(
        CommandStep(
            label="Python regression suite",
            command=(sys.executable, "-m", "unittest", "discover", "-s", "scripts/tests", "-v"),
            cwd=root,
        )
    )

    for relative_script in OFFLINE_VALIDATORS:
        steps.append(
            CommandStep(
                label=f"Offline validator {relative_script}",
                command=_python_script(root, relative_script),
                cwd=root,
            )
        )

    for relative_script in SMOKE_SCRIPTS:
        steps.append(
            CommandStep(
                label=f"Offline smoke {relative_script}",
                command=_python_script(root, relative_script),
                cwd=root,
            )
        )

    steps.append(
        CommandStep(
            label="Readiness audit",
            command=_python_script(root, "scripts/audit_readiness.py"),
            cwd=root,
        )
    )
    steps.append(
        CommandStep(
            label="Git whitespace check",
            command=("git", "diff", "--check"),
            cwd=root,
        )
    )
    return tuple(steps)


def required_executables(plan: Sequence[CommandStep]) -> tuple[str, ...]:
    """Return missing command names before any partial verification starts."""

    missing = set()
    for step in plan:
        executable = step.command[0]
        if Path(executable).is_absolute():
            available = Path(executable).is_file()
        else:
            available = shutil.which(executable) is not None
        if not available:
            missing.add(executable)
    return tuple(sorted(missing))


def _print_plan(plan: Sequence[CommandStep]) -> None:
    for index, step in enumerate(plan, start=1):
        print(f"{index:02d}. {step.label}: {' '.join(step.command)}")


def run_command_plan(plan: Sequence[CommandStep]) -> int:
    for index, step in enumerate(plan, start=1):
        print(f"[{index:02d}/{len(plan)}] {step.label}", flush=True)
        environment = os.environ.copy()
        environment.update(dict(step.environment))
        try:
            result = subprocess.run(step.command, cwd=step.cwd, env=environment, check=False)
        except OSError as error:
            print(f"BLOCKED: {step.command[0]} could not start: {error}", file=sys.stderr)
            return 127
        if result.returncode != 0:
            print(f"FAILED: {step.label} (exit {result.returncode})", file=sys.stderr)
            return result.returncode
    print("OFFLINE CHECKS PASS: offline/read-only contract completed")
    return 0


def main(argv: Optional[Sequence[str]] = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--list", action="store_true", help="print the plan without executing commands")
    parser.add_argument("--root", type=Path, default=ROOT, help="repository root (default: current checkout)")
    args = parser.parse_args(argv)

    root = args.root.resolve()
    plan = build_command_plan(root)
    if args.list:
        _print_plan(plan)
        print("Excluded operated validators: " + ", ".join(OPERATED_VALIDATORS))
        print("Excluded real n8n engine: scripts/run_n8n_engine_regression.py and scripts/run_n8n_m06_schedule_regression.py")
        return 0

    missing = required_executables(plan)
    if missing:
        print(
            "OFFLINE CHECKS BLOCKED: missing required executable(s): " + ", ".join(missing),
            file=sys.stderr,
        )
        print("Install the declared local toolchain or run this entrypoint in hosted CI; no checks were started.", file=sys.stderr)
        return 2
    return run_command_plan(plan)


if __name__ == "__main__":
    raise SystemExit(main())
