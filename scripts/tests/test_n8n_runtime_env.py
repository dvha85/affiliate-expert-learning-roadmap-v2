import sys
import tempfile
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

from n8n_runtime_env import isolated_n8n_environment


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


if __name__ == "__main__":
    unittest.main()
