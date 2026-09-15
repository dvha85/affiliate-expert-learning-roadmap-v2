import importlib.util
from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("learner_bot_shard", ROOT / "scripts" / "run_learner_bot_test_shard.py")
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class LearnerBotTestShardTests(unittest.TestCase):
    def test_partitions_every_test_once_and_stably(self):
        names = [f"TestCase{index}" for index in range(80)]
        first = [MODULE.select_shard(names, index, 2) for index in range(2)]
        second = [MODULE.select_shard(reversed(names), index, 2) for index in range(2)]
        self.assertEqual(first, second)
        self.assertEqual(sorted(first[0] + first[1]), sorted(names))
        self.assertTrue(all(group for group in first))

    def test_rejects_invalid_partition_parameters_and_names(self):
        with self.assertRaises(ValueError):
            MODULE.select_shard(["TestGood"], 0, 1)
        with self.assertRaises(ValueError):
            MODULE.select_shard(["TestGood"], 2, 2)
        with self.assertRaises(ValueError):
            MODULE.select_shard(["not-a-test"], 0, 2)
