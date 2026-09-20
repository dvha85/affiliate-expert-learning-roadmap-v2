import sys
import argparse
import contextlib
import io
import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import Mock, patch


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

from n8n_runtime_env import isolated_n8n_environment
import run_n8n_engine_regression as engine
import run_n8n_m06_schedule_regression as schedule


class N8nRuntimeEnvironmentTests(unittest.TestCase):
    def test_external_database_queue_and_file_settings_are_removed(self):
        parent = {
            "PATH": "/usr/bin",
            "DB_TYPE": "postgresdb",
            "DB_TYPE_FILE": "/run/secrets/db-type",
            "DB_POSTGRESDB_HOST": "existing.example.invalid",
            "DB_POSTGRESDB_PASSWORD_FILE": "/run/secrets/password",
            "EXECUTIONS_MODE": "queue",
            "QUEUE_BULL_REDIS_HOST": "existing-redis.example.invalid",
            "N8N_TASK_RUNNERS_ENABLED": "true",
            "N8N_LOG_FILE_LOCATION": "/existing/n8n.log",
            "N8N_HOST": "existing.example.invalid",
            "N8N_RUNNERS_BROKER_PORT": "9999",
            "N8N_CONFIG_FILES": "/existing/config",
        }
        with tempfile.TemporaryDirectory() as directory:
            runtime = Path(directory)
            env = isolated_n8n_environment(runtime, broker_port=12345, base=parent)
        self.assertEqual(env["DB_TYPE"], "sqlite")
        self.assertEqual(env["EXECUTIONS_MODE"], "regular")
        self.assertEqual(env["N8N_RUNNERS_BROKER_PORT"], "12345")
        self.assertEqual(env["N8N_USER_FOLDER"], str(runtime / "n8n"))
        for key in parent:
            if key.startswith(("N8N_", "DB_", "QUEUE_", "TASK_RUNNERS_", "N8N_QUEUE_", "N8N_DATABASE_", "EXECUTIONS_")):
                self.assertNotEqual(env.get(key), parent[key], key)

    def test_node_path_is_prepended_without_losing_parent_path(self):
        with tempfile.TemporaryDirectory() as directory:
            runtime = Path(directory) / "runtime"
            node = Path(directory) / "node"
            env = isolated_n8n_environment(runtime, broker_port=123, base={"PATH": "/usr/bin"}, node_path=node)
        self.assertTrue(env["PATH"].startswith(str(node.parent.resolve())))
        self.assertIn("/usr/bin", env["PATH"])

    def test_both_runners_isolate_environment_before_first_import_child(self):
        class ImportBoundaryReached(Exception):
            pass

        keys = ("DB_TYPE", "DB_TYPE_FILE", "DB_POSTGRESDB_HOST", "DB_POSTGRESDB_PASSWORD_FILE", "EXECUTIONS_MODE", "QUEUE_BULL_REDIS_HOST", "N8N_CONFIG_FILES", "N8N_USER_FOLDER")
        hostile = dict(zip(keys, ("postgresdb", "/external/type", "existing.example.invalid", "/external/password", "queue", "redis.example.invalid", "/external/config", "/external/home")))
        for module in (engine, schedule):
            with self.subTest(runner=module.__name__), tempfile.TemporaryDirectory() as directory:
                runtime = Path(directory) / "runtime"
                runtime.mkdir()
                observed = []
                code = "import json,os; print(json.dumps({key:os.environ.get(key) for key in " + repr(keys) + "}))"
                prefix = [sys.executable, "-c", code]

                def fake_run(command, *, env, **kwargs):
                    if command[0] == "go":
                        return Mock(returncode=0)
                    self.assertIn("import:workflow", command)
                    child = subprocess.run(command, env=env, capture_output=True, text=True, check=True)
                    observed.append(json.loads(child.stdout))
                    raise ImportBoundaryReached()

                with contextlib.ExitStack() as stack:
                    stack.enter_context(patch.dict(os.environ, hostile))
                    stack.enter_context(patch.object(module.tempfile, "mkdtemp", return_value=str(runtime)))
                    stack.enter_context(patch.object(module, "run", side_effect=fake_run))
                    stack.enter_context(patch.object(module, "start_adapter", return_value=Mock()))
                    stack.enter_context(patch.object(module, "stop_adapter"))
                    stack.enter_context(patch.object(module, "choose_port", return_value=12345))
                    stack.enter_context(contextlib.redirect_stderr(io.StringIO()))
                    with self.assertRaises(ImportBoundaryReached):
                        if module is engine:
                            stack.enter_context(patch.object(module, "command_prefix", return_value=prefix))
                            stack.enter_context(patch.object(module, "validate_n8n_command"))
                            stack.enter_context(patch.object(sys, "argv", ["runner"]))
                            module.main()
                        else:
                            module.run_case(prefix, argparse.Namespace(keep_runtime=False, n8n_node=None), available_adapter=True)
                self.assertEqual(len(observed), 1)
                actual = observed[0]
                self.assertEqual(actual["DB_TYPE"], "sqlite")
                self.assertEqual(actual["EXECUTIONS_MODE"], "regular")
                self.assertEqual(Path(actual["N8N_USER_FOLDER"]).resolve(), runtime.resolve() / "n8n")
                for key in keys:
                    self.assertNotEqual(actual[key], hostile[key], key)


if __name__ == "__main__":
    unittest.main()
