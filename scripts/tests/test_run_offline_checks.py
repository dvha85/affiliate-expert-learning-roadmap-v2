import unittest


from scripts.run_offline_checks import OFFLINE_VALIDATORS, OPERATED_VALIDATORS, SMOKE_SCRIPTS, build_command_plan


class OfflineCheckPlanTests(unittest.TestCase):
    def test_operated_validators_are_explicitly_excluded(self):
        self.assertTrue(OPERATED_VALIDATORS)
        self.assertTrue(set(OPERATED_VALIDATORS).isdisjoint(OFFLINE_VALIDATORS))
        self.assertTrue(all("operated_execution" in path for path in OPERATED_VALIDATORS))

    def test_offline_plan_contains_the_review_baseline_steps(self):
        labels = {step.label for step in build_command_plan()}
        self.assertIn("Python regression suite", labels)
        self.assertIn("Readiness audit", labels)
        self.assertIn("Git whitespace check", labels)
        self.assertTrue(any(label.startswith("Go test contracts") for label in labels))
        self.assertTrue(any(label.startswith("Go vet lab/affiliate-bot") for label in labels))
        self.assertTrue(any(label.startswith("Offline validator scripts/validate_n8n_m07.py") for label in labels))
        self.assertTrue(any(label.startswith("Offline smoke scripts/smoke_br18b_backup_restore.py") for label in labels))

    def test_review_lists_are_non_empty_and_unique(self):
        self.assertEqual(len(OFFLINE_VALIDATORS), len(set(OFFLINE_VALIDATORS)))
        self.assertEqual(len(SMOKE_SCRIPTS), len(set(SMOKE_SCRIPTS)))
        self.assertGreaterEqual(len(OFFLINE_VALIDATORS), 10)
        self.assertGreaterEqual(len(SMOKE_SCRIPTS), 10)


if __name__ == "__main__":
    unittest.main()
