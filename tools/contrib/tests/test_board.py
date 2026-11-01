import io
import os
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import board

ITEM = """# x
## AAA-1: first thing
- size: M
- status: ready
- kind: decomp
- source: queue:20
- question: q
- start: s
- deliverable: d
- needs: n
## BBB-2: second | pipe
- size: S
- status: done
- kind: verify
- source: backlog:FOO
- question: q
- start: s
- deliverable: d
- needs: n
"""
QUEUE = "| id | q |\n|---|---|\n| 20 | a | b | M | proposed (x) |\n| 3 | a | b | M | done 2026 |\n"


def run(*argv):
    out, err = io.StringIO(), io.StringIO()
    with redirect_stdout(out), redirect_stderr(err):
        rc = board.main(list(argv))
    return rc, out.getvalue(), err.getvalue()


class Parse(unittest.TestCase):
    def test_parse_and_validate(self):
        items = board.parse(ITEM)
        self.assertEqual([i["id"] for i in items], ["AAA-1", "BBB-2"])
        self.assertEqual(board.validate(items, QUEUE, "FOO"), [])

    def test_errors(self):
        items = board.parse(ITEM.replace("queue:20", "queue:3"))
        self.assertTrue(any("already done" in e for e in board.validate(items, QUEUE, "FOO")))
        items = board.parse(ITEM.replace("queue:20", "queue:99"))
        self.assertTrue(any("no row 99" in e for e in board.validate(items, QUEUE, "FOO")))
        items = board.parse(ITEM.replace("- size: M\n", ""))
        self.assertTrue(any("missing size" in e for e in board.validate(items)))
        items = board.parse(ITEM)
        self.assertTrue(any("FOO" in e for e in board.validate(items, QUEUE, "nothing")))
        with self.assertRaises(board.BoardError):
            board.parse("## no colon here\n")

    def test_verify_needs_done_row(self):
        items = board.parse(ITEM.replace("queue:20", "queue:20").replace("kind: decomp", "kind: verify"))
        self.assertTrue(any("needs a done queue row" in e for e in board.validate(items, QUEUE, "FOO")))

    def test_render(self):
        items = board.parse(ITEM)
        t = board.render_board(items, {"AAA-1": "https://github.com/o/r/issues/5"})
        self.assertIn("| AAA-1 | first thing | M | decomp | ready |", t)
        self.assertIn("[issue](https://github.com/o/r/issues/5)", t)
        self.assertIn("second / pipe", t)
        self.assertIn("1 of 2 items are open.", t)

    def test_issue_body(self):
        b = board.issue_body(board.parse(ITEM)[0])
        for s in ("### Question", "### Deliverable", "### Evidence checklist", "Comment `claiming`", "NO game files"):
            self.assertIn(s, b)

    def test_issues_tsv_rejects_non_github(self):
        with tempfile.NamedTemporaryFile("w", dir=board.ROOT, suffix=".tsv", delete=False) as f:
            f.write("A\thttps://github.com/o/r/issues/1\nB\thttp://evil.example/x\nC\tnotaurl\n")
        try:
            self.assertEqual(board.load_issues(os.path.relpath(f.name, board.ROOT)), {"A": "https://github.com/o/r/issues/1"})
        finally:
            os.unlink(f.name)


class Real(unittest.TestCase):
    def test_real_source_is_valid_and_board_current(self):
        rc, _, err = run("--check")
        self.assertEqual(rc, 0, err)

    def test_dry_run_writes_nothing_and_issues_dir(self):
        rc, out, _ = run()
        self.assertEqual(rc, 0)
        self.assertIn("dry run", out)
        with tempfile.TemporaryDirectory() as d:
            rc, out, _ = run("--write-issues", d)
            self.assertEqual(rc, 0)
            files = sorted(os.listdir(d))
            self.assertIn("SKMELEE.md", files)
            self.assertNotIn("BBB-2.md", files)
            self.assertIn("nothing was sent", out)


if __name__ == "__main__":
    unittest.main()
