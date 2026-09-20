import sys
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

from n8n_change_scope import requires_n8n_engine


class N8nChangeScopeTests(unittest.TestCase):
    def test_workflow_uses_shared_scope_helper(self):
        workflow = (SCRIPTS.parent / ".github" / "workflows" / "mission-agent-path-ci.yml").read_text(encoding="utf-8")
        self.assertIn("python scripts/n8n_change_scope.py", workflow)
        self.assertNotIn("grep -Eq", workflow)

    def test_direct_integration_dependencies_run_engine(self):
        paths = [
            "lab/n8n/M06-readonly-watcher.blueprint.json",
            "contracts/action-intent.schema.json",
            "core/m07/m07.go",
            "lab/affiliate-bot/go.mod",
            "lab/affiliate-bot/cmd/bot/watcher.go",
            "lab/affiliate-bot/internal/store/history.go",
            "scripts/n8n_runtime_env.py",
            "scripts/n8n_cli_preflight.py",
            "scripts/run_n8n_engine_regression.py",
            "scripts/run_n8n_m06_schedule_regression.py",
            "scripts/validate_n8n_m06_operated_execution.py",
        ]
        for path in paths:
            with self.subTest(path=path):
                self.assertTrue(requires_n8n_engine([path]))

    def test_docs_and_unrelated_labs_can_skip_engine(self):
        self.assertFalse(
            requires_n8n_engine(
                [
                    "README.md",
                    "docs/plans/REVIEW-REMEDIATION-PLAN.md",
                    "lab/affiliate-bot/README.md",
                    "lab/mission-runtime/cmd/demo/m11_chain.go",
                    "scripts/smoke_br18b_backup_restore.py",
                ]
            )
        )

    def test_whitespace_and_dot_prefixes_are_normalized(self):
        self.assertTrue(requires_n8n_engine(["  ./core/m06/offer_fixture.go  "]))
        self.assertTrue(requires_n8n_engine([".github/workflows/mission-agent-path-ci.yml"]))
        self.assertTrue(requires_n8n_engine([".github/workflows/curriculum-ci.yml"]))


if __name__ == "__main__":
    unittest.main()
