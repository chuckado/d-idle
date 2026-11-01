import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import coverage as c

DEN = """subsystem\tlabel\tkind\ttotal\tkeys\tsource
skills.sor\tSorceress\tints\t5\t36-40\tskills.txt
monster.ai\tAI\tnames\t3\tA,B,C\tmonstats.txt
treasure\tTC\tcount\t10\t\ttc.txt
"""
LED = "subsystem\trows\tstatus\tevidence\tcases\tseed_or_date\tnote\n"
GAPS = "rank\tsubsystem\trows\timpact\tsource\n"


def make(ledger_rows, gaps_rows="", doc="trace says 1,234 cases and D1 differs", extra_den=""):
    d = tempfile.mkdtemp()
    os.makedirs(os.path.join(d, "docs", "coverage"))
    with open(os.path.join(d, "docs", "t.md"), "w") as f:
        f.write(doc)
    for name, body in (("denominators.tsv", DEN + extra_den), ("ledger.tsv", LED + ledger_rows), ("gaps.tsv", GAPS + gaps_rows)):
        with open(os.path.join(d, "docs", "coverage", name), "w") as f:
            f.write(body)
    return d


def build(root):
    uni = c.load_universes(root)
    per = c.load_ledger(uni, root)
    gaps = c.load_gaps(uni, per, root)
    return uni, per, gaps


class LedgerTests(unittest.TestCase):
    def test_counts_and_render(self):
        root = make("skills.sor\t36-37\tTESTED-0-DIFF\tdocs/t.md\t1234\t2026-11-01\t\n"
                    "skills.sor\t38\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n"
                    "skills.sor\t39\tTESTED-DEVIATION\tdocs/t.md\t1234\tseed 7\tD1\n"
                    "treasure\t#4\tBOT-ONLY\tdocs/t.md\t\t2026-10-30\t\n",
                    "1\tskills.sor\tuntested\tbig\tdocs/t.md\n")
        uni, per, gaps = build(root)
        t = c.tally(uni, per)
        self.assertEqual(t["skills.sor"]["TESTED-0-DIFF"], 2)
        self.assertEqual(t["skills.sor"]["UNTESTED"], 1)
        self.assertEqual(t["treasure"]["BOT-ONLY"], 4)
        self.assertEqual(t["treasure"]["UNTESTED"], 6)
        text = c.render(uni, per, gaps)
        self.assertIn("| **total** | **18** | **2** | **1** | **1** | **4** | **10** |", text)
        self.assertIn("untested (1 of 5)", text)
        self.assertIn("as of 2026-11-01", text)

    def test_v_run_needs_count(self):
        root = make("skills.sor\t36\tTESTED-0-DIFF\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "no case count"):
            build(root)

    def test_count_must_appear_in_doc(self):
        root = make("skills.sor\t36\tTESTED-0-DIFF\tdocs/t.md\t999\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "does not appear"):
            build(root)

    def test_missing_evidence(self):
        root = make("skills.sor\t36\tREAD-ONLY\tdocs/nope.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "does not exist"):
            build(root)

    def test_deviation_needs_id(self):
        root = make("skills.sor\t36\tTESTED-DEVIATION\tdocs/t.md\t1234\t2026-11-01\tno id here\n")
        with self.assertRaisesRegex(c.CovError, "deviation id"):
            build(root)

    def test_overlap_and_unknown_rows(self):
        root = make("skills.sor\t36-38\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n"
                    "skills.sor\t38\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "already in another"):
            build(root)
        root = make("skills.sor\t99\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "not rows of"):
            build(root)

    def test_count_only_rows(self):
        root = make("treasure\t36\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "count-only"):
            build(root)
        root = make("treasure\t#11\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "exceed"):
            build(root)

    def test_gap_already_tested(self):
        root = make("skills.sor\t36\tREAD-ONLY\tdocs/t.md\t\t2026-11-01\t\n", "1\tskills.sor\t36\tx\tdocs/t.md\n")
        with self.assertRaisesRegex(c.CovError, "already have a status"):
            build(root)

    def test_bad_status_and_date(self):
        root = make("skills.sor\t36\tDONE\tdocs/t.md\t\t2026-11-01\t\n")
        with self.assertRaisesRegex(c.CovError, "status"):
            build(root)
        root = make("skills.sor\t36\tREAD-ONLY\tdocs/t.md\t\tyesterday\t\n")
        with self.assertRaisesRegex(c.CovError, "seed_or_date"):
            build(root)


class HelperTests(unittest.TestCase):
    def test_ints_roundtrip(self):
        ids = c.expand_ints("0-2,5,7-8", "t")
        self.assertEqual(ids, {0, 1, 2, 5, 7, 8})
        self.assertEqual(c.compact_ints(ids), "0-2,5,7-8")

    def test_tables(self):
        d = tempfile.mkdtemp()
        def w(name, text):
            with open(os.path.join(d, name + ".tsv"), "w") as f:
                f.write(text)
        w("skills", "skill\tId\tcharclass\tEType\nA\t0\t\t\nB\t36\tsor\tfire\nC\t37\tsor\t\n")
        w("missiles", "Missile\tId\nx\t0\ny\t1\n")
        w("monstats", "Id\thcIdx\tAI\nm\t0\tSkeleton\nn\t1\tSkeleton\nExpansion\t\t\n")
        w("states", "state\tid\nnone\t0\n")
        w("itemtypes", "ItemType\tCode\nNone\t\nSword\tswor\n")
        w("armor", "name\tcode\ncap\tcap\n")
        w("weapons", "name\tcode\naxe\taxe\nhax\thax\n")
        w("treasureclassex", "Treasure Class\tgroup\nGold\t\n")
        w("levels", "Name\tId\nNull\t0\nExpansion\t\n")
        u = c.universes_from_tables(d)
        self.assertEqual(u["skills.sor"], ("ints", {36, 37}))
        self.assertEqual(u["skills.monster"], ("ints", {0}))
        self.assertEqual(u["skillfx.damage"], ("ints", {36}))
        self.assertEqual(u["monster.ai"], ("names", {"Skeleton"}))
        self.assertEqual(u["monster.stats"], ("ints", {0, 1}))
        self.assertEqual(u["itemtypes"], ("names", {"swor"}))
        self.assertEqual(u["items.weapons"], ("count", 2))


class RepoTests(unittest.TestCase):
    def test_repo_inputs_valid_and_output_fresh(self):
        self.assertEqual(c.main(["--check"]), 0)


if __name__ == "__main__":
    unittest.main()
