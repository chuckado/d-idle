#!/usr/bin/env python3
"""Tests for knobbudget.py. Run: python3 -I tools/ci/test_knobbudget.py"""
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import knobbudget as K  # noqa: E402


def knobs(*keys):
    return "x\n<!-- knobdump:begin -->\n| key | default |\n|---|---|\n" + "".join("| %s | 1 |\n" % k for k in keys) + "<!-- knobdump:end -->\n"


class Knobs(unittest.TestCase):
    def test_parse(self):
        self.assertEqual(K.knob_keys(knobs("a.b", "c.d")), {"a.b", "c.d"})
        self.assertEqual(K.knob_keys(None), set())

    def test_net_rise_fails(self):
        self.assertEqual(len(K.knob_findings(knobs("a.b"), knobs("a.b", "c.d"), "")), 1)

    def test_add_with_removal_passes(self):
        self.assertEqual(K.knob_findings(knobs("a.b"), knobs("c.d"), ""), [])

    def test_owner_ok_passes(self):
        self.assertEqual(K.knob_findings(knobs("a.b"), knobs("a.b", "c.d"), "knob_net: +1 owner-ok\n"), [])

    def test_net_zero_text_is_not_owner_ok(self):
        self.assertEqual(len(K.knob_findings(knobs("a.b"), knobs("a.b", "c.d"), "knob_net: 0\n")), 1)


class Rows(unittest.TestCase):
    base = {"routes": [{"name": "r", "minLevel": 7, "evidence": 1, "evidence_ref": "single incident: needs seed set"}], "gap": 3}

    def run_rows(self, base, head):
        return K.row_findings("f.json", base, head)

    def test_unchanged_passes_even_with_n1(self):
        self.assertEqual(self.run_rows(self.base, self.base), [])

    def test_seeding_evidence_is_not_a_change(self):
        old = {"routes": [{"name": "r", "minLevel": 7}], "gap": 3}
        self.assertEqual(self.run_rows(old, self.base), [])

    def test_change_with_n1_fails(self):
        head = {"routes": [{"name": "r", "minLevel": 9, "evidence": 1, "evidence_ref": "single incident: needs seed set"}], "gap": 3}
        self.assertEqual(len(self.run_rows(self.base, head)), 1)

    def test_change_with_seed_set_passes(self):
        head = {"routes": [{"name": "r", "minLevel": 9, "evidence": 5, "evidence_ref": "docs/probes/seed-x.md"}], "gap": 3}
        self.assertEqual(self.run_rows(self.base, head), [])

    def test_change_with_margin_model_passes(self):
        head = {"routes": [{"name": "r", "minLevel": 9, "evidence": 1, "evidence_ref": "margin-model: zone 5 v2"}], "gap": 3}
        self.assertEqual(self.run_rows(self.base, head), [])

    def test_new_row_without_evidence_fails_top_level_row_too(self):
        head = {"routes": self.base["routes"], "gap": 4}
        self.assertEqual(len(self.run_rows(self.base, head)), 1)

    def test_ancestor_evidence_applies(self):
        head = {"evidence": 3, "evidence_ref": "docs/x.md", "farm": [{"area": 9, "goldPerMin": 8}]}
        self.assertEqual(self.run_rows(None, head), [])


if __name__ == "__main__":
    unittest.main()
