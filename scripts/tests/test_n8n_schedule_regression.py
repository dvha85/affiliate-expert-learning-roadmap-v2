import json
import ast
import inspect
from contextlib import closing
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import Mock, patch


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

import run_n8n_m06_schedule_regression as runner
from run_n8n_m06_schedule_regression import decode_flatted_execution


class N8nScheduleRegressionTests(unittest.TestCase):
    def test_flatted_numeric_strings_remain_literals(self):
        raw = json.dumps([{"value": "1", "nested": ["2"]}, "1", "2"])
        self.assertEqual(
            decode_flatted_execution(raw),
            {"value": "1", "nested": ["2"]},
        )

    def test_flatted_nested_references_are_resolved(self):
        raw = json.dumps([{"child": "1"}, {"value": "2"}, "literal"])
        self.assertEqual(
            decode_flatted_execution(raw),
            {"child": {"value": "literal"}},
        )

    def test_flatted_cycles_are_marked_without_recursing_forever(self):
        raw = json.dumps([{"self": "0"}])
        self.assertEqual(
            decode_flatted_execution(raw),
            {"self": {"$flatted_ref": 0}},
        )

    def test_flatted_shared_references_are_not_cycles(self):
        raw = json.dumps([{"a": "1", "b": "1"}, {"values": "2"}, ["3", "4", 5], "007", "text"])
        child = {"values": ["007", "text", 5]}
        self.assertEqual(decode_flatted_execution(raw), {"a": child, "b": child})

    def test_restart_requires_new_finished_trigger_after_shutdown_watermark(self):
        with tempfile.TemporaryDirectory() as directory:
            database = Path(directory) / "database.sqlite"
            log = Path(directory) / "server.log"
            log.write_text("fixture", encoding="utf-8")
            with closing(sqlite3.connect(database)) as connection, connection:
                connection.execute("CREATE TABLE execution_entity (id INTEGER, workflowId TEXT, mode TEXT, status TEXT, finished INTEGER)")
                connection.execute("CREATE TABLE execution_data (executionId INTEGER, data TEXT)")
                connection.executemany("INSERT INTO execution_entity VALUES (?, ?, ?, ?, ?)", [
                    (2, "workflow", "trigger", "success", 1),
                    (3, "workflow", "trigger", "success", 1),
                    (99, "other-workflow", "trigger", "success", 1),
                ])
                connection.execute("INSERT INTO execution_data VALUES (3, ?)", (json.dumps([{"value": "1"}, "old"]),))
            watermark = runner.latest_execution_id(database, "workflow")
            self.assertEqual(watermark, 3)
            self.assertEqual(runner.latest_execution_id(database, "missing"), 0)
            server = Mock()
            server.poll.return_value = None
            # No wall-clock sleep: exactly one polling iteration, then timeout.
            with patch.object(runner.time, "monotonic", side_effect=[0, 0, 2]), patch.object(runner.time, "sleep"):
                with self.assertRaisesRegex(AssertionError, "timed out"):
                    runner.wait_for_execution(database, "workflow", "success", server, log, after_id=watermark, timeout=1)
            with closing(sqlite3.connect(database)) as connection, connection:
                connection.executemany("INSERT INTO execution_entity VALUES (?, ?, ?, ?, ?)", [
                    (4, "workflow", "manual", "success", 1),
                    (5, "workflow", "trigger", "success", 0),
                    (6, "workflow", "trigger", "success", 1),
                ])
                for execution_id in (4, 5, 6):
                    connection.execute("INSERT INTO execution_data VALUES (?, ?)", (execution_id, json.dumps([{"value": "1"}, "new"])))
            self.assertEqual(runner.wait_for_execution(database, "workflow", "success", server, log, after_id=watermark), (6, {"value": "new"}))

    def test_restart_watermark_is_taken_after_stop_and_used_for_replay(self):
        tree = ast.parse(inspect.getsource(runner.run_case))
        calls = sorted((node for node in ast.walk(tree) if isinstance(node, ast.Call) and isinstance(node.func, ast.Name)), key=lambda node: node.lineno)
        watermark = next(node for node in calls if node.func.id == "latest_execution_id")
        stops = [node for node in calls if node.func.id == "stop_n8n" and node.lineno < watermark.lineno]
        self.assertTrue(stops)
        restart = next(node for node in calls if node.func.id == "start_n8n" and node.lineno > watermark.lineno)
        replay = next(node for node in calls if node.func.id == "wait_for_execution" and node.lineno > restart.lineno)
        after_id = next(keyword.value for keyword in replay.keywords if keyword.arg == "after_id")
        self.assertIsInstance(after_id, ast.Name)
        self.assertEqual(after_id.id, "restart_watermark")


if __name__ == "__main__":
    unittest.main()
