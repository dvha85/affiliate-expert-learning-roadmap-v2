from __future__ import annotations

import importlib.util
import json
import tempfile
import shutil
import subprocess
import os
import re
import unittest
from pathlib import Path
from unittest.mock import patch


ROOT = Path(__file__).resolve().parents[2]
VALIDATOR_PATH = ROOT / "scripts" / "validate_learner_walkthroughs.py"
SPEC = importlib.util.spec_from_file_location("validate_learner_walkthroughs", VALIDATOR_PATH)
assert SPEC and SPEC.loader
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class LearnerWalkthroughValidationTests(unittest.TestCase):
    def test_private_learner_evidence_path_is_actually_gitignored(self):
        text = (ROOT / "curriculum/PRACTICE-M00-M05.md").read_text(encoding="utf-8")
        self.assertIn("workspace/learner/M01…M05", text)
        examples = ["workspace/learner/M01/evidence.md", "workspace/learner/M05/evidence.md", "workspace/learner/CA-T18.md"]
        result = subprocess.run(
            ["git", "check-ignore", "--no-index", "--stdin", "-z"], cwd=ROOT,
            input=b"\0".join(path.encode("utf-8") for path in examples) + b"\0",
            capture_output=True,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout.decode("utf-8").split("\0")[:-1], examples)

    def test_published_posix_blocks_parse(self):
        bash = Path("C:/Program Files/Git/bin/bash.exe") if os.name == "nt" else Path(shutil.which("bash") or "/missing-bash")
        if not bash.is_file():
            self.skipTest("Bash is not installed")
        for relative in ("curriculum/PRACTICE-M00-M05.md", "curriculum/M06/M06.3-n8n-readonly-workflow.md"):
            text = (ROOT / relative).read_text(encoding="utf-8")
            for index, match in enumerate(re.finditer(r"([`~]{3})(?:sh|bash)\n(.*?)\n\1", text, re.S)):
                with self.subTest(document=relative, block=index):
                    result = subprocess.run([str(bash), "-n"], input=match[2], text=True, capture_output=True)
                    self.assertEqual(result.returncode, 0, result.stderr)

    @unittest.skipUnless(os.name == "nt", "PowerShell parser is verified on Windows")
    def test_published_powershell_blocks_parse(self):
        for relative in ("curriculum/PRACTICE-M00-M05.md", "curriculum/M06/M06.3-n8n-readonly-workflow.md"):
            text = (ROOT / relative).read_text(encoding="utf-8")
            for index, match in enumerate(re.finditer(r"([`~]{3})powershell\n(.*?)\n\1", text, re.S)):
                with self.subTest(document=relative, block=index):
                    source = match[2].replace("'", "''")
                    command = "$tokens=$null; $errors=$null; [void][System.Management.Automation.Language.Parser]::ParseInput('" + source + "',[ref]$tokens,[ref]$errors); if ($errors.Count) { $errors | Out-String | Write-Output; exit 1 }"
                    result = subprocess.run(["powershell", "-NoProfile", "-Command", command], text=True, capture_output=True)
                    self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_manual_n8n_python_launchers_use_isolated_environment(self):
        text = (ROOT / "curriculum/M06/M06.3-n8n-readonly-workflow.md").read_text(encoding="utf-8")
        snippets = re.findall(r"(?m)^import sys, subprocess\n(.*?)^raise SystemExit\((.*?)\)\s*$", text, re.S)
        self.assertEqual(len(snippets), 2)
        for content, call in snippets:
            script = "import sys, subprocess\n" + content + "raise SystemExit(" + call + ")"
            with tempfile.TemporaryDirectory() as directory:
                runtime = Path(directory)
                with patch.object(subprocess, "call", return_value=0) as child, patch("sys.argv", ["inline", str(runtime), str(runtime / "node.exe"), str(runtime / "n8n")]), patch.dict(os.environ, {"DB_TYPE": "postgresdb", "N8N_USER_FOLDER": "unrelated-runtime"}):
                    with self.assertRaises(SystemExit) as outcome:
                        exec(compile(script, "documented-n8n-launcher", "exec"), {})
                self.assertEqual(outcome.exception.code, 0)
                environment = child.call_args.kwargs["env"]
                self.assertEqual(environment["DB_TYPE"], "sqlite")
                self.assertEqual(environment["N8N_USER_FOLDER"], str(runtime / "n8n"))
                self.assertEqual(environment["N8N_LISTEN_ADDRESS"], "127.0.0.1")
                self.assertEqual(environment["N8N_BLOCK_ENV_ACCESS_IN_NODE"], "false")

    def test_blueprint_negative_endpoint_ack_and_bypass_paths(self):
        path = "lab/n8n/M06-readonly-watcher.blueprint.json"
        for mutation, expected in (("endpoint", "endpoint"), ("ack", "ACK guard"), ("bypass", "ACK flow")):
            with self.subTest(mutation=mutation):
                data = json.loads((ROOT / path).read_text(encoding="utf-8"))
                nodes = {node["name"]: node for node in data["nodes"]}
                if mutation == "endpoint":
                    nodes["Build and Append Canonical M06 Adapter"]["parameters"]["url"] = "http://127.0.0.1/wrong"
                elif mutation == "ack":
                    nodes["Require Canonical Store ACK"]["parameters"]["jsCode"] = "return $input.all();"
                else:
                    data["connections"]["Build and Append Canonical M06 Adapter"]["main"][0].append({"node": "Report Canonical M06 Result"})
                self.assertTrue(any(expected in e for e in MODULE.validate_blueprint(data, path)))

    def test_equivalent_execution_table_heading_is_accepted(self):
        original_read = MODULE._read
        def read(root, relative):
            value = original_read(root, relative)
            if relative == "curriculum/PRACTICE-M00-M05.md":
                value = value.replace("Expected output", "Kết quả dự kiến").replace("Input, output, failure và evidence cho từng chặng", "Bảng kết quả")
            if relative == "curriculum/M06/M06.3-n8n-readonly-workflow.md":
                value = value.replace("persistence: APPENDED | EXACT_DUPLICATE | HANDOFF_ERROR", "Persistence: HANDOFF_ERROR, EXACT_DUPLICATE, APPENDED")
            return value
        with patch.object(MODULE, "_read", side_effect=read):
            self.assertEqual(MODULE.validate(ROOT), [])

    def test_removed_synthetic_credit_boundary_is_rejected(self):
        original_read = MODULE._read
        def read(root, relative):
            value = original_read(root, relative)
            if relative == "curriculum/PRACTICE-M00-M05.md":
                value = "\n".join(line for line in value.splitlines() if not (line.startswith("|") and "synthetic rehearsal" in line))
            return value
        with patch.object(MODULE, "_read", side_effect=read):
            self.assertIn("M00-M05 practice lacks explicit synthetic/learner-credit boundary", MODULE.validate(ROOT))

    def test_review_blueprint_auth_cannot_be_satisfied_by_note(self):
        with tempfile.TemporaryDirectory() as directory:
            isolated = Path(directory)
            paths = set(MODULE.PRACTICE_LINK_TARGETS) | set(MODULE.DOCS_WITH_ADAPTER_SETUP) | {
                "curriculum/PRACTICE-M00-M05.md", "starter-kits/M06-readonly-watcher/CHECKPOINTS.md",
                "lab/affiliate-bot/cmd/bot/watcher.go",
                "lab/n8n/M06-readonly-watcher.blueprint.json", "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
            }
            for relative in paths:
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(ROOT / relative, target)
            target = isolated / "lab/n8n/M06-readonly-watcher.blueprint.json"
            blueprint = json.loads(target.read_text(encoding="utf-8"))
            node = next(node for node in blueprint["nodes"] if node["name"] == "Build and Append Canonical M06 Adapter")
            node["parameters"]["headerParameters"] = {"parameters": []}
            node["notes"] = "CANONICAL_ADAPTER_TOKEN is mentioned but not wired"
            target.write_text(json.dumps(blueprint), encoding="utf-8")
            self.assertTrue(any("Authorization" in e for e in MODULE.validate(isolated)))

    def test_review_history_and_negated_warning_do_not_fail(self):
        original_read = MODULE._read
        def read(root, relative):
            value = original_read(root, relative)
            if relative in MODULE.DOCS_WITH_ADAPTER_SETUP:
                value += "\n## Lịch sử\nTên cũ: AFFILIATE_ADAPTER_TOKEN.\n"
            if relative == "curriculum/PRACTICE-M00-M05.md":
                value += "\nDo not skip the Mission gate.\n"
            return value
        with patch.object(MODULE, "_read", side_effect=read):
            self.assertEqual(MODULE.validate(ROOT), [])

    def test_review_m06_exercise_has_persistence_cases(self):
        text = (ROOT / "curriculum/M06/M06.3-n8n-readonly-workflow.md").read_text(encoding="utf-8")
        for case, status in (("first_import", "APPENDED"), ("exact_retry", "EXACT_DUPLICATE"), ("identity_conflict", "HANDOFF_ERROR"), ("new_event", "APPENDED")):
            self.assertRegex(text, rf"\| {case} \|[^\n]*`{status}`")

    def test_review_walkthrough_has_concrete_json_inputs(self):
        text = (ROOT / "curriculum/PRACTICE-M00-M05.md").read_text(encoding="utf-8")
        for name in ("action.json", "outcome.json", "advisor-config.json", "evaluation-config.json", "proposal.json", "review.json"):
            self.assertIn(f"<!-- rehearsal:{name} -->", text)
        self.assertNotIn("<history.jsonl>", text)
        self.assertNotIn("<observed-time>", text)

    def test_all_planned_entrypoints_link_the_shared_walkthrough(self) -> None:
        required_documents = (
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
        for relative in required_documents:
            with self.subTest(document=relative):
                text = (ROOT / relative).read_text(encoding="utf-8")
                self.assertIn("PRACTICE-M00-M05.md", text)

    def test_guard_rejects_a_removed_practice_link(self) -> None:
        target = "curriculum/M01/M01.1-deterministic-contract.md"
        original_read = MODULE._read

        def read_without_target_link(root: Path, relative: str) -> str:
            content = original_read(root, relative)
            if relative == target:
                return content.replace("PRACTICE-M00-M05.md", "", 1)
            return content

        with patch.object(MODULE, "_read", side_effect=read_without_target_link):
            errors = MODULE.validate(ROOT)
        self.assertEqual(errors, [f"{target} does not link the practice bridge"])

    def test_guard_rejects_a_broken_walkthrough_href(self) -> None:
        target = "curriculum/M01/M01.1-deterministic-contract.md"
        original_read = MODULE._read

        def read_with_broken_href(root: Path, relative: str) -> str:
            content = original_read(root, relative)
            if relative == target:
                return content.replace(
                    "../PRACTICE-M00-M05.md",
                    "../missing/PRACTICE-M00-M05.md",
                    1,
                )
            return content

        with patch.object(MODULE, "_read", side_effect=read_with_broken_href):
            errors = MODULE.validate(ROOT)
        self.assertEqual(errors, [f"{target} has an invalid walkthrough link target"])

    def test_guard_requires_input_output_failure_evidence_table(self) -> None:
        original_read = MODULE._read

        def read_without_expected_output_header(root: Path, relative: str) -> str:
            content = original_read(root, relative)
            if relative == "curriculum/PRACTICE-M00-M05.md":
                return content.replace("Expected output", "Observed output", 1)
            return content

        with patch.object(MODULE, "_read", side_effect=read_without_expected_output_header):
            errors = MODULE.validate(ROOT)
        self.assertEqual(
            errors,
            ["M00-M05 practice bridge lacks the structured input/expected-output/failure/evidence table"],
        )

    def test_guard_rejects_wrong_m06_expected_status(self) -> None:
        original_read = MODULE._read

        def read_with_wrong_change_status(root: Path, relative: str) -> str:
            content = original_read(root, relative)
            if relative == "curriculum/M06/M06.3-n8n-readonly-workflow.md":
                return content.replace(
                    "change detection: NEW | UNCHANGED | CHANGED",
                    "change detection: NEW | UNCHANGED | APPENDED",
                    1,
                )
            return content

        with patch.object(MODULE, "_read", side_effect=read_with_wrong_change_status):
            errors = MODULE.validate(ROOT)
        self.assertEqual(
            errors,
            ["M06 lesson has incorrect change-detection or persistence expected statuses"],
        )

    def test_guard_rejects_fixture_as_real_or_gate_bypass_instruction(self) -> None:
        original_read = MODULE._read

        forbidden_additions = (
            "Fixture is real E1 and counts as M00 PASS.",
            "Skip the Mission gate after the smoke passes.",
            "Update PROGRESS.md to PASS after CI smoke.",
        )
        expected = "M00-M05 practice bridge contains a learner-credit or Mission-gate bypass instruction"
        for addition in forbidden_additions:
            with self.subTest(instruction=addition):
                def read_with_bypass_instruction(root: Path, relative: str) -> str:
                    content = original_read(root, relative)
                    if relative == "curriculum/PRACTICE-M00-M05.md":
                        return content + f"\n{addition}\n"
                    return content

                with patch.object(MODULE, "_read", side_effect=read_with_bypass_instruction):
                    errors = MODULE.validate(ROOT)
                self.assertEqual(errors, [expected])

    def test_current_repository_walkthroughs_are_aligned(self) -> None:
        self.assertEqual(MODULE.validate(ROOT), [])

    def test_retired_token_name_is_rejected_in_isolated_copy(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            isolated = Path(directory)
            for relative in MODULE.DOCS_WITH_ADAPTER_SETUP:
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(
                    "CANONICAL_ADAPTER_TOKEN\nAFFILIATE_ADAPTER_TOKEN\n",
                    encoding="utf-8",
                )
            for relative in (
                "lab/n8n/M06-readonly-watcher.blueprint.json",
                "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
            ):
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(
                    json.dumps({"value": "CANONICAL_ADAPTER_TOKEN"}),
                    encoding="utf-8",
                )
            (isolated / "lab/affiliate-bot/cmd/bot").mkdir(parents=True)
            (isolated / "lab/affiliate-bot/cmd/bot/watcher.go").write_text(
                'os.Getenv("CANONICAL_ADAPTER_TOKEN")', encoding="utf-8"
            )
            errors = MODULE.validate(isolated)
            self.assertTrue(any("retired token" in error for error in errors))

    def test_m06_status_layers_are_not_satisfied_by_one_status(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            isolated = Path(directory)
            for relative in MODULE.DOCS_WITH_ADAPTER_SETUP:
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("CANONICAL_ADAPTER_TOKEN", encoding="utf-8")
            for relative in (
                "lab/n8n/M06-readonly-watcher.blueprint.json",
                "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
            ):
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(
                    json.dumps({"value": "CANONICAL_ADAPTER_TOKEN"}),
                    encoding="utf-8",
                )
            watcher = isolated / "lab/affiliate-bot/cmd/bot/watcher.go"
            watcher.parent.mkdir(parents=True, exist_ok=True)
            watcher.write_text('os.Getenv("CANONICAL_ADAPTER_TOKEN")', encoding="utf-8")
            errors = MODULE.validate(isolated)
            self.assertTrue(any("change state NEW" in error for error in errors))
            self.assertTrue(any("persistence status APPENDED" in error for error in errors))

    def test_m03_to_m05_learner_cli_markers_are_required(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            isolated = Path(directory)
            for relative in MODULE.DOCS_WITH_ADAPTER_SETUP:
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text("CANONICAL_ADAPTER_TOKEN", encoding="utf-8")
            for relative in (
                "lab/n8n/M06-readonly-watcher.blueprint.json",
                "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
            ):
                target = isolated / relative
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(
                    json.dumps({"value": "CANONICAL_ADAPTER_TOKEN"}),
                    encoding="utf-8",
                )
            watcher = isolated / "lab/affiliate-bot/cmd/bot/watcher.go"
            watcher.parent.mkdir(parents=True, exist_ok=True)
            watcher.write_text('os.Getenv("CANONICAL_ADAPTER_TOKEN")', encoding="utf-8")
            (isolated / "curriculum").mkdir(parents=True, exist_ok=True)
            (isolated / "curriculum/README.md").write_text(
                "PRACTICE-M00-M05.md", encoding="utf-8"
            )
            (isolated / "curriculum/PRACTICE-M00-M05.md").write_text(
                "M00 packet M01 M02 M03 M04 M05 synthetic Mission PASS PROGRESS.md "
                "action record outcome import advisor mock evaluation create "
                "proposal import review import ./cmd/bot .\\cmd\\bot",
                encoding="utf-8",
            )
            errors = MODULE.validate(isolated)
            self.assertTrue(any("learner command marker" in error for error in errors))


if __name__ == "__main__":
    unittest.main()
