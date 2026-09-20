"""Protect fixture chronology without weakening the process-owned clock."""
import sys
import unittest
from datetime import datetime, timedelta, timezone
from pathlib import Path
from unittest.mock import patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
import smoke_br16a_offline as br16
import smoke_br18b_backup_restore as br18


class SmokeFixtureTimeTests(unittest.TestCase):
    def test_whole_second_intervals_and_historical_dates_are_preserved(self):
        for module in (br16, br18):
            with self.subTest(module=module.__name__):
                earlier = datetime.fromisoformat(module.fixture_time("2026-09-07T00:00:00Z"))
                later = datetime.fromisoformat(module.fixture_time("2026-09-08T00:00:00Z"))
                self.assertEqual(later - earlier, timedelta(days=1))

    def test_fractional_authorizations_always_precede_reservation(self):
        # A process started near a second rollover previously shifted the
        # second AND reattached the unshifted fraction. All 24 authorizations
        # could then be later than their reservation, causing zero winners.
        now = datetime(2026, 9, 21, 0, 0, 0, 999500, tzinfo=timezone.utc)
        for module in (br16, br18):
            with self.subTest(module=module.__name__), patch.object(module, "_FIXTURE_NOW", now):
                base = datetime.fromisoformat(module.fixture_time("2026-09-08T00:00:00Z"))
                reserved = datetime.fromisoformat(module.fixture_time("2026-09-08T00:00:01Z"))
                for index in range(1, 25):
                    authorized = datetime.fromisoformat(module.fixture_time(f"2026-09-08T00:00:00.{index:03d}Z"))
                    self.assertLess(authorized, reserved)
                    self.assertEqual(authorized - base, timedelta(milliseconds=index))

    def test_nanosecond_fraction_is_preserved_without_carry(self):
        now = datetime(2026, 9, 21, 0, 0, 0, 999999, tzinfo=timezone.utc)
        for module in (br16, br18):
            with self.subTest(module=module.__name__), patch.object(module, "_FIXTURE_NOW", now):
                self.assertEqual(module.fixture_time("2026-09-08T00:00:00.999999999Z"), "2026-09-21T00:00:00.999999999Z")


if __name__ == "__main__":
    unittest.main()
