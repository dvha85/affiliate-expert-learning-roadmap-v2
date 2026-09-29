"""Validate learner-facing walkthrough contracts against the current runtime.

This is a documentation/contract guard. It does not execute Go, n8n, a model,
an external provider, or a live action. Runtime and operated evidence remain
separate acceptance layers.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


EXPECTED_TOKEN = "CANONICAL_ADAPTER_TOKEN"
WRONG_TOKEN = "AFFILIATE_ADAPTER_TOKEN"
M06_SUCCESS_STATUSES = ("APPENDED", "EXACT_DUPLICATE")
M06_CHANGE_STATES = ("NEW", "UNCHANGED", "CHANGED")
PRACTICE_BYPASS_PATTERNS = (
    re.compile(r"(?im)\b(?:skip|bypass|ignore)\b[^\n]{0,60}\b(?:mission\s+)?gate\b"),
    re.compile(r"(?im)\b(?:bỏ qua|vượt qua)\b[^\n]{0,60}\b(?:mission\s+)?gate\b"),
    re.compile(
        r"(?im)\b(?:synthetic|fixture)\b[^\n]{0,80}\b(?:counts as|is|becomes)\s+"
        r"(?:real|E1|M00 PASS|Mission PASS|Reality PASS|Operated PASS)\b"
    ),
    re.compile(
        r"(?im)\b(?:mark|set|update)\b[^\n]{0,50}\bPROGRESS\.md\b"
        r"[^\n]{0,80}\b(?:smoke|fixture|CI)\b"
    ),
)
PRACTICE_COMMAND_MARKERS = (
    "action record",
    "outcome import",
    "advisor mock",
    "evaluation create",
    "proposal import",
    "review import",
    "./cmd/bot",
    ".\\cmd\\bot",
)

PRACTICE_LINK_TARGETS = (
    "curriculum/README.md",
    "curriculum/M01/M01.1-deterministic-contract.md",
    "curriculum/M01/M01.4-failure-first-operated-proof.md",
    "curriculum/M02/M02.4-restart-query-operated-proof.md",
    "starter-kits/M01-deterministic-bot/README.md",
    "starter-kits/M01-deterministic-bot/CHECKPOINTS.md",
    "starter-kits/M02-history-replay/README.md",
    "starter-kits/M02-history-replay/CHECKPOINTS.md",
    "curriculum/M03/M03.3-compliance-and-operated-proof.md",
    "curriculum/M04/M04.3-ai-eval-safe-profile.md",
    "curriculum/M05/M05.2-review-and-rollback.md",
    "starter-kits/M03-tracked-human-action/README.md",
    "starter-kits/M03-tracked-human-action/CHECKPOINTS.md",
    "starter-kits/M03-tracked-human-action/M03-OPERATED-EVIDENCE-TEMPLATE.md",
    "starter-kits/M04-grounded-ai-advisor/README.md",
    "starter-kits/M04-grounded-ai-advisor/CHECKPOINTS.md",
    "starter-kits/M04-grounded-ai-advisor/M04-OPERATED-EVIDENCE-TEMPLATE.md",
    "starter-kits/M05-reviewed-improvement/README.md",
    "starter-kits/M05-reviewed-improvement/CHECKPOINTS.md",
    "starter-kits/M05-reviewed-improvement/M05-OPERATED-EVIDENCE-TEMPLATE.md",
    "starter-kits/CONTINUITY-CHECKPOINT.md",
)

DOCS_WITH_ADAPTER_SETUP = (
    "curriculum/M06/M06.3-n8n-readonly-workflow.md",
    "curriculum/M07/M07.3-n8n-evidence-agent.md",
    "starter-kits/M06-readonly-watcher/README.md",
)
CODE_BLOCK_RE = re.compile(r"(?m)^([`~]{3,})[^\n]*\n(.*?)^\1\s*$", re.DOTALL)
MARKDOWN_LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)")


def _read(root: Path, relative: str) -> str:
    path = root / relative
    return path.read_text(encoding="utf-8") if path.is_file() else ""


def _code_blocks(text: str) -> str:
    """Return fenced command/content blocks, excluding surrounding prose."""

    return "\n".join(match.group(2) for match in CODE_BLOCK_RE.finditer(text))


def active_instructions(text: str) -> str:
    """Exclude explicitly historical/negative-example sections, not live steps."""
    ignored_level = None
    lines = []
    fence = None
    for line in text.splitlines():
        if re.match(r"^(`{3,}|~{3,})", line):
            fence = None if fence else line[:3]
        heading = re.match(r"^(#{1,6})\s+(.+)", line) if not fence else None
        if heading:
            level = len(heading[1])
            if ignored_level is not None and level <= ignored_level:
                ignored_level = None
            if re.search(r"lịch sử|historical|ví dụ sai|negative example", heading[2], re.I):
                ignored_level = level
        if ignored_level is None:
            lines.append(line)
    return "\n".join(lines)


def execution_table_present(text: str) -> bool:
    roles = (("chặng", "stage"), ("input", "đầu vào"),
             ("expected output", "kết quả dự kiến"), ("failure", "ca lỗi"),
             ("evidence", "bằng chứng"))
    return any(all(any(word in line.casefold() for word in aliases) for aliases in roles)
               for line in text.splitlines() if line.startswith("|"))


def load_rehearsal_inputs(root: Path) -> dict:
    text = _read(root, "curriculum/PRACTICE-M00-M05.md")
    pattern = re.compile(r"<!-- rehearsal:([\w-]+\.json) -->\s*([`~]{3})json\s*\n(.*?)\n\2", re.S)
    result = {}
    for match in pattern.finditer(text):
        if match[1] in result:
            raise ValueError(f"duplicate rehearsal input {match[1]}")
        value = json.loads(match[3])
        if not isinstance(value, dict):
            raise ValueError(f"rehearsal input must be an object: {match[1]}")
        result[match[1]] = value
    expected = {"action.json", "outcome.json", "advisor-config.json", "evaluation-config.json", "proposal.json", "review.json"}
    if set(result) != expected:
        raise ValueError("rehearsal requires six concrete JSON input samples")
    return result


def validate_blueprint(blueprint: dict, relative: str) -> list[str]:
    errors = []
    if not isinstance(blueprint, dict) or not isinstance(blueprint.get("nodes"), list):
        return [f"invalid blueprint nodes: {relative}"]
    nodes = {node.get("name"): node for node in blueprint["nodes"] if isinstance(node, dict)}
    if len(nodes) != len(blueprint["nodes"]):
        errors.append(f"duplicate/invalid blueprint nodes: {relative}")
    m06 = "M06-" in relative
    endpoints = ({"Build and Append Canonical M06 Adapter": "/v1/m06/fixture-import"} if m06 else {
        "Fetch and Register Tool Adapter": "/v1/m07/fetch-and-register",
        "Canonical M07 Context Adapter": "/v1/m07/context",
        "Validate Grounding Adapter": "/v1/m07/validate",
        "Persist Agent Proposal Adapter": "/v1/m07/register-proposal",
    })
    for name, endpoint in endpoints.items():
        params = nodes.get(name, {}).get("parameters", {})
        headers = params.get("headerParameters", {}).get("parameters", [])
        authorization = [item.get("value", "") for item in headers if item.get("name", "").lower() == "authorization"]
        if params.get("sendHeaders") is not True or len(authorization) != 1 or not re.fullmatch(
            r"=\{\{\s*['\"]Bearer ['\"]\s*\+\s*\$env\.CANONICAL_ADAPTER_TOKEN\s*\}\}", authorization[0]
        ):
            errors.append(f"{relative}: {name} Authorization is not wired to canonical token")
        if params.get("method") != "POST" or endpoint not in params.get("url", ""):
            errors.append(f"{relative}: {name} has wrong endpoint/method")
    chain = (["Schedule Trigger", "M06 Adapter Input", "Build and Append Canonical M06 Adapter", "Require Canonical Store ACK", "Report Canonical M06 Result"] if m06 else [
        "Manual Trigger", "M07 Adapter Input", "Fetch and Register Tool Adapter", "Require Registered Tool ACK",
        "Canonical M07 Context Adapter", "Require Canonical M07 Context", "Read-only Evidence Agent",
        "Require Raw Model JSON Text", "Validate Grounding Adapter", "Require Grounded Proposal",
        "Persist Agent Proposal Adapter", "Require Persisted Agent Proposal ACK", "Report Persisted M07 Proposal",
    ])
    for source, target in zip(chain, chain[1:]):
        branches = blueprint.get("connections", {}).get(source, {}).get("main", [])
        destinations = [item.get("node") for branch in branches for item in branch]
        if source not in nodes or target not in nodes or destinations != [target]:
            errors.append(f"{relative}: broken/bypassed ACK flow {source} -> {target}")
    if m06:
        ack = nodes.get("Require Canonical Store ACK", {}).get("parameters", {}).get("jsCode", "")
        for marker in ("status!=='APPENDED'", "status!=='EXACT_DUPLICATE'", "canonical_history_ack!==true", "canonical_history_persisted!==true", "!response.record_id"):
            if marker not in ack:
                errors.append(f"{relative}: missing ACK guard {marker}")
        fields = {item.get("name"): item.get("value") for item in nodes.get("Report Canonical M06 Result", {}).get("parameters", {}).get("assignments", {}).get("assignments", [])}
        for field in ("record_id", "canonical_history_handoff", "canonical_history_persisted", "execution_permitted"):
            if fields.get(field) != "={{ $json." + field + " }}":
                errors.append(f"{relative}: report loses {field}")
    else:
        for name, markers in {
            "Require Registered Tool ACK": ("response.status!=='ACK'", "response.artifact_id", "response.evidence_raw_json"),
            "Require Canonical M07 Context": ("response.status!=='VALID'", "response.artifact.record_id!==expected"),
            "Require Grounded Proposal": ("response.artifact.state!=='HUMAN_REVIEW'", "response.execution_permitted!==false"),
            "Require Persisted Agent Proposal ACK": ("response.status!=='ACK'", "response.artifact_id", "execution_permitted:false"),
        }.items():
            code = nodes.get(name, {}).get("parameters", {}).get("jsCode", "")
            for marker in markers:
                if marker not in code:
                    errors.append(f"{relative}: missing ACK guard {name}: {marker}")
    return errors


def validate(root: Path) -> list[str]:
    errors: list[str] = []

    for relative in DOCS_WITH_ADAPTER_SETUP:
        text = active_instructions(_read(root, relative))
        if not text:
            errors.append(f"missing learner document: {relative}")
            continue
        if EXPECTED_TOKEN not in text:
            errors.append(f"{relative} does not name {EXPECTED_TOKEN}")
        if WRONG_TOKEN in text:
            errors.append(f"{relative} retains retired token name {WRONG_TOKEN}")

    m06_lesson = _read(root, "curriculum/M06/M06.3-n8n-readonly-workflow.md")
    m06_starter = _read(root, "starter-kits/M06-readonly-watcher/README.md")
    m06_checkpoint = _read(root, "starter-kits/M06-readonly-watcher/CHECKPOINTS.md")
    m06_bundle = "\n".join((m06_lesson, m06_starter, m06_checkpoint))
    for marker in M06_CHANGE_STATES:
        if marker not in m06_bundle:
            errors.append(f"M06 learner path does not explain change state {marker}")
    for marker in M06_SUCCESS_STATUSES:
        if marker not in m06_bundle:
            errors.append(f"M06 learner path does not explain persistence status {marker}")
    for marker in ("change detection", "persistence", "HANDOFF_ERROR", "history bytes"):
        if marker.casefold() not in m06_bundle.casefold():
            errors.append(f"M06 learner path lacks distinction marker: {marker}")
    m06_lesson_code = _code_blocks(m06_lesson)
    for label, states in (("change detection", set(M06_CHANGE_STATES)), ("persistence", set(M06_SUCCESS_STATUSES) | {"HANDOFF_ERROR"})):
        match = re.search(rf"(?im)^{label}:\s*([^\n]+)", m06_lesson_code)
        if not match or set(re.findall(r"\b[A-Z][A-Z_]+\b", match[1])) != states:
            errors.append("M06 lesson has incorrect change-detection or persistence expected statuses")
            break
    for case, status in (("first_import", "APPENDED"), ("exact_retry", "EXACT_DUPLICATE"), ("identity_conflict", "HANDOFF_ERROR"), ("new_event", "APPENDED")):
        if not re.search(rf"(?m)^\|\s*{case}\s*\|[^\n]*`{status}`", m06_lesson):
            errors.append(f"M06 exercise lacks expected persistence case {case}: {status}")
    for marker in ("isolated_n8n_environment", "N8N_BLOCK_ENV_ACCESS_IN_NODE", "New-Item", "mkdir", "127.0.0.1:8787"):
        if marker not in m06_lesson_code:
            errors.append(f"M06 setup lacks isolated runtime step: {marker}")

    blueprint_paths = (
        "lab/n8n/M06-readonly-watcher.blueprint.json",
        "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    )
    for relative in blueprint_paths:
        path = root / relative
        try:
            blueprint = json.loads(path.read_text(encoding="utf-8"))
        except FileNotFoundError:
            errors.append(f"missing blueprint: {relative}")
            continue
        except json.JSONDecodeError as exc:
            errors.append(f"invalid blueprint {relative}: {exc}")
            continue
        errors.extend(validate_blueprint(blueprint, relative))

    watcher_source = _read(root, "lab/affiliate-bot/cmd/bot/watcher.go")
    if f'os.Getenv("{EXPECTED_TOKEN}")' not in watcher_source:
        errors.append("watcher runtime does not read the canonical adapter token")
    if f'os.Getenv("{WRONG_TOKEN}")' in watcher_source:
        errors.append("watcher runtime still reads the retired adapter token")

    workflow = _read(root, "curriculum/README.md")
    practice = _read(root, "curriculum/PRACTICE-M00-M05.md")
    practice_code = _code_blocks(practice)
    if "PRACTICE-M00-M05.md" not in workflow:
        errors.append("curriculum README does not link the M00-M05 practice bridge")
    for marker in (
        "M00 packet",
        "M01",
        "M02",
        "M03",
        "M04",
        "M05",
        "synthetic",
        "Mission PASS",
        "PROGRESS.md",
    ):
        if marker not in practice:
            errors.append(f"M00-M05 practice bridge lacks marker: {marker}")
    for marker in PRACTICE_COMMAND_MARKERS:
        if marker not in practice_code:
            errors.append(f"M00-M05 practice bridge lacks learner command marker: {marker}")
    if not execution_table_present(practice):
        errors.append(
            "M00-M05 practice bridge lacks the structured input/expected-output/failure/evidence table"
        )
    prose = CODE_BLOCK_RE.sub("", active_instructions(practice))
    actionable_lines = [line for line in prose.splitlines() if not re.search(
        r"(?:do not|must not|never|don't|không)\s+(?:skip|bypass|ignore|update|mark|set|bỏ qua|vượt qua)", line, re.I)]
    if any(pattern.search(line) for line in actionable_lines for pattern in PRACTICE_BYPASS_PATTERNS):
        errors.append("M00-M05 practice bridge contains a learner-credit or Mission-gate bypass instruction")
    boundary_rows = [line for line in practice.splitlines() if line.startswith("|") and "synthetic rehearsal" in line]
    if not any(re.search(r"không|not|no ", line, re.I) and all(x in line for x in ("Mission PASS", "Reality PASS", "Operated PASS")) for line in boundary_rows):
        errors.append("M00-M05 practice lacks explicit synthetic/learner-credit boundary")
    try:
        load_rehearsal_inputs(root)
    except (ValueError, json.JSONDecodeError) as error:
        errors.append(f"M00-M05 practice inputs: {error}")
    if re.search(r"<(?:history|actions|outcomes|evaluations|proposals|reviews)\.jsonl>|<observed-time>", practice_code):
        errors.append("M00-M05 runnable commands contain unresolved store/time placeholders")

    for relative in PRACTICE_LINK_TARGETS:
        text = _read(root, relative)
        if "PRACTICE-M00-M05.md" not in text:
            errors.append(f"{relative} does not link the practice bridge")
            continue
        targets = MARKDOWN_LINK_RE.findall(CODE_BLOCK_RE.sub("", text))
        practice_targets = [
            target for target in targets
            if Path(target.split("#", 1)[0]).name == "PRACTICE-M00-M05.md"
        ]
        expected_path = (root / "curriculum/PRACTICE-M00-M05.md").resolve()
        if not practice_targets or any(
            not (root / relative).parent.joinpath(target.split("#", 1)[0]).resolve() == expected_path
            for target in practice_targets
        ):
            errors.append(f"{relative} has an invalid walkthrough link target")

    return errors


def main(argv: list[str] | None = None) -> int:
    args = argv if argv is not None else sys.argv[1:]
    root = Path(args[0]).resolve() if args else Path(__file__).resolve().parents[1]
    errors = validate(root)
    if errors:
        print("LEARNER WALKTHROUGH VALIDATION FAILED")
        for error in errors:
            print(f"- {error}")
        return 1
    print("LEARNER WALKTHROUGH VALIDATION PASS: token, M06 semantics and M00-M05 bridge are aligned")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
