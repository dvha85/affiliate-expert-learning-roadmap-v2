"""Guard the checked-in, deliberately simple aggregate gate YAML contract."""
import os
from pathlib import Path
import re
import shutil
import subprocess
import unittest

ROOT = Path(__file__).resolve().parents[2]
BASH = (str(Path(os.environ.get("ProgramFiles", "C:/Program Files")) / "Git/bin/bash.exe")
        if os.name == "nt" else shutil.which("bash"))


def job_blocks(text):
    text = text.split("\njobs:\n", 1)[1]
    matches = list(re.finditer(r"^  ([a-z0-9-]+):\s*$", text, re.MULTILINE))
    return {match.group(1): text[match.end():matches[index + 1].start() if index + 1 < len(matches) else len(text)]
            for index, match in enumerate(matches)}


class CIGateContractTests(unittest.TestCase):
    def test_every_workflow_job_is_a_required_fail_closed_dependency(self):
        governance = (ROOT / "docs/governance/REPOSITORY-GOVERNANCE.md").read_text(encoding="utf-8")
        for filename, gate in (("curriculum-ci.yml", "curriculum-gate"), ("mission-agent-path-ci.yml", "mission-gate")):
            with self.subTest(workflow=filename):
                workflow = (ROOT / ".github/workflows" / filename).read_text(encoding="utf-8")
                jobs = job_blocks(workflow)
                block = jobs[gate]
                self.assertIn("if: ${{ always() }}", block)
                self.assertIn(f"name: {gate}", block)
                needs_block = re.search(r"    needs:\n((?:      - [^\n]+\n)+)", block)
                self.assertIsNotNone(needs_block, "use an explicit needs list for the aggregate gate")
                needs = re.findall(r"      - ([a-z0-9-]+)", needs_block[1])
                self.assertCountEqual(needs, set(jobs) - {gate})
                guarded = re.findall(r'test "\$\{\{ needs\.([a-z0-9-]+)\.result \}\}" = success', block)
                self.assertCountEqual(guarded, needs)
                self.assertNotRegex(workflow, r"continue-on-error:\s*true")
                self.assertRegex(governance, rf"\| `{re.escape(filename)}`[^\n]*\| `{gate}` \|")

    @unittest.skipUnless(BASH and Path(BASH).is_file(), "bash required to execute the actual workflow gate")
    def test_actual_gate_shell_rejects_failed_cancelled_and_skipped_children(self):
        for filename, gate in (("curriculum-ci.yml", "curriculum-gate"), ("mission-agent-path-ci.yml", "mission-gate")):
            block = job_blocks((ROOT / ".github/workflows" / filename).read_text(encoding="utf-8"))[gate]
            script = "\n".join(line[10:] for line in block.splitlines() if line.startswith("          test "))
            needs = re.findall(r"needs\.([a-z0-9-]+)\.result", script)
            self.assertTrue(needs)
            for failed_job, result in [(None, "success")] + [(job, status) for job in needs for status in ("failure", "cancelled", "skipped")]:
                with self.subTest(gate=gate, job=failed_job, result=result):
                    command = re.sub(r"\$\{\{ needs\.([a-z0-9-]+)\.result \}\}", lambda match: result if match[1] == failed_job else "success", script)
                    completed = subprocess.run([BASH, "-e", "-c", command], capture_output=True, text=True)
                    self.assertEqual(completed.returncode == 0, failed_job is None)

    @unittest.skipUnless(BASH and Path(BASH).is_file(), "bash required to execute the actual workflow selection")
    def test_scope_helper_error_never_becomes_docs_only_skip(self):
        workflow = (ROOT / ".github/workflows/mission-agent-path-ci.yml").read_text(encoding="utf-8")
        selection = workflow.split("      - name: Select n8n engine coverage", 1)[1].split("      - uses:", 1)[0]
        script = "\n".join(line[10:] for line in selection.split("        run: |\n", 1)[1].splitlines())
        # Replace only data acquisition/helper execution; execute the actual
        # checked-in decision branches. No GitHub or external runtime is used.
        script = script.replace('changed="$(git diff --name-only "$BASE_SHA" "$GITHUB_SHA")"', 'changed="README.md"')
        script = script.replace(' >> "$GITHUB_OUTPUT"', '')
        for status, decision in ((0, "run"), (1, "skip"), (1, ""), (0, ""), (2, ""), (127, "")):
            with self.subTest(exit=status, decision=decision):
                command = script.replace("python scripts/n8n_change_scope.py", f"(printf '{decision}'; exit {status})")
                result = subprocess.run([BASH, "-e", "-o", "pipefail", "-c", command], env=dict(os.environ, EVENT_NAME="pull_request", GITHUB_OUTPUT="/dev/stdout"), capture_output=True, text=True)
                if decision:
                    self.assertEqual(result.returncode, 0, result.stderr)
                    self.assertIn("run=" + ("true" if status == 0 else "false"), result.stdout)
                else:
                    self.assertNotEqual(result.returncode, 0, result.stdout)
                    self.assertNotIn("run=false", result.stdout)


if __name__ == "__main__":
    unittest.main()
