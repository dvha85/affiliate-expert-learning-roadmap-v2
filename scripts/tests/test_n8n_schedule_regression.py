import json
import sys
import unittest
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(SCRIPTS))

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


if __name__ == "__main__":
    unittest.main()
