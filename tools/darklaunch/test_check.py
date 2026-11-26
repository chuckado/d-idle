#!/usr/bin/env python3
"""Tests: python3 -I tools/darklaunch/test_check.py"""
import os
import sys
import unittest
from datetime import date

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import check  # noqa: E402
import ledger as L  # noqa: E402

TODAY = date(2026, 12, 1)


def item(**kw):
    d = dict(name="a.b", built_by="x (1)", dark_since="2026-11-01", why_off="w", enable_criteria="c",
             owner="coordinator", review_by="2026-12-15", status="dark", criteria_met="U")
    d.update(kw)
    return d


KNOBS = {"a.b": ("0", "| a.b | 0 | x [dark-launch] |"), "c.d": ("0", "| c.d | 0 | plain |")}


class T(unittest.TestCase):
    def test_ok(self):
        self.assertEqual(check.problems([item()], KNOBS, TODAY), [])

    def test_marked_knob_without_row_fails(self):
        ps = check.problems([], KNOBS, TODAY)
        self.assertEqual(len(ps), 1)
        self.assertIn("a.b", ps[0])

    def test_unmarked_default_zero_is_not_flagged(self):
        self.assertEqual(check.problems([item()], KNOBS, TODAY), [])

    def test_stale_row_fails(self):
        ps = check.problems([item(review_by="2026-11-30")], KNOBS, TODAY)
        self.assertTrue(any("past" in p for p in ps))

    def test_stale_row_with_status_change_passes(self):
        self.assertEqual(check.problems([item(review_by="2026-11-30", status="retired")], KNOBS, TODAY), [])

    def test_missing_field_and_bad_status(self):
        self.assertTrue(check.problems([item(enable_criteria="")], KNOBS, TODAY))
        self.assertTrue(any("status" in p for p in check.problems([item(status="x")], KNOBS, TODAY)))

    def test_row_for_vanished_knob_fails(self):
        self.assertTrue(check.problems([item(name="z.z")], KNOBS, TODAY))

    def test_real_ledger_and_knobs_parse(self):
        items = L.load_ledger(check.os.path.join(check.ROOT, "data", "darklaunch.json"))
        self.assertGreater(len(items), 20)
        self.assertEqual(check.problems(items, L.knob_rows(os.path.join(check.ROOT, "docs", "knobs.md")), date(2026, 11, 25)), [])


if __name__ == "__main__":
    unittest.main()
