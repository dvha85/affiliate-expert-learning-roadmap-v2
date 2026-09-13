import sys
import tempfile
import unittest
from pathlib import Path
from unittest import mock


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

from n8n_cli_preflight import command_prefix, validate_n8n_command


class N8nCLIPreflightTests(unittest.TestCase):
    def test_js_entrypoint_without_node_fails_before_runtime_work(self):
        with tempfile.TemporaryDirectory() as directory:
            entrypoint = Path(directory) / "n8n"
            entrypoint.write_text("#!/usr/bin/env node\n", encoding="utf-8")
            prefix = command_prefix(str(entrypoint), None)
            with mock.patch("n8n_cli_preflight.shutil.which", return_value=None):
                with self.assertRaisesRegex(SystemExit, "--n8n-node"):
                    validate_n8n_command(prefix, str(entrypoint), None)

    def test_explicit_node_allows_a_javascript_entrypoint_without_path_lookup(self):
        with tempfile.TemporaryDirectory() as directory:
            entrypoint = Path(directory) / "n8n"
            node = Path(directory) / "node"
            entrypoint.write_text("#!/usr/bin/env node\n", encoding="utf-8")
            node.write_text("#!/bin/sh\n", encoding="utf-8")
            prefix = command_prefix(str(entrypoint), str(node))
            with mock.patch("n8n_cli_preflight.shutil.which", return_value=None):
                validate_n8n_command(prefix, str(entrypoint), str(node))

    def test_non_node_entrypoint_does_not_require_node(self):
        with tempfile.TemporaryDirectory() as directory:
            entrypoint = Path(directory) / "n8n"
            entrypoint.write_text("#!/bin/sh\n", encoding="utf-8")
            prefix = command_prefix(str(entrypoint), None)
            with mock.patch("n8n_cli_preflight.shutil.which", return_value=None):
                validate_n8n_command(prefix, str(entrypoint), None)


if __name__ == "__main__":
    unittest.main()
