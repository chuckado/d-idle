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
        self.assertNotIn("BBB-2", t)  # done items are not on the board
        self.assertIn("1 of 1 listed items are open", t)
        self.assertIn("work-archive.md", t)

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


ARCH = board.ARCHIVE_HEAD


class Archive(unittest.TestCase):
    def moved(self, extra=""):
        return board.do_archive(ITEM + extra, "", "2026-01-02", "someone")

    def test_done_section_moves_with_metadata(self):
        src, arc, moved = self.moved()
        self.assertEqual(moved, ["BBB-2"])
        self.assertEqual([i["id"] for i in board.parse(src)], ["AAA-1"])
        item = board.parse(arc)[0]
        self.assertEqual((item["id"], item["done"], item["by"]), ("BBB-2", "2026-01-02", "someone"))
        self.assertTrue(item["result"].startswith("docs/"))
        self.assertEqual(board.validate_archive([item], set()), [])

    def test_idempotent_and_keeps_existing_metadata(self):
        src, arc, _ = self.moved("- done: 2025-05-05\n- by: octocat\n- result: docs/research/z.md\n")
        src2, arc2, moved2 = board.do_archive(src, arc, "2030-01-01")
        self.assertEqual((src2, arc2, moved2), (src, arc, []))
        item = board.parse(arc)[0]
        self.assertEqual((item["done"], item["by"], item["result"]), ("2025-05-05", "octocat", "docs/research/z.md"))

    def test_table_between_markers_and_missing_markers(self):
        _, arc, _ = self.moved()
        t = board.render_archive(arc, board.parse(arc), {"BBB-2": "https://github.com/o/r/issues/9"})
        self.assertIn("| BBB-2 | second / pipe | verify | S |", t)
        self.assertIn("[issue](https://github.com/o/r/issues/9)", t)
        self.assertEqual(board.render_archive(t, board.parse(t), {"BBB-2": "https://github.com/o/r/issues/9"}), t)
        with self.assertRaises(board.BoardError):
            board.render_archive("no markers", [])

    def test_archive_validation(self):
        _, arc, _ = self.moved()
        item = board.parse(arc)[0]
        self.assertTrue(any("both" in e for e in board.validate_archive([item], {"BBB-2"})))
        bad = dict(item, status="ready", done="soon")
        errs = board.validate_archive([bad], set())
        self.assertTrue(any("status done" in e for e in errs) and any("YYYY" in e for e in errs))


class Real(unittest.TestCase):
    def test_real_source_is_valid_and_board_current(self):
        rc, _, err = run("--check")
        self.assertEqual(rc, 0, err)

    def test_real_archive_holds_only_done_and_board_holds_none_of_them(self):
        opens, archived = board.load_all()
        self.assertTrue(archived)
        self.assertEqual([i["id"] for i in opens if i["status"] == "done"], [])
        board_text = board.read(board.OUT)
        for it in archived:
            self.assertEqual(it["status"], "done")
            self.assertNotIn("| %s |" % it["id"], board_text)
            self.assertIn("| %s |" % it["id"], board.read(board.ARCHIVE))
            self.assertTrue(os.path.exists(os.path.join(board.ROOT, it["result"])), it["result"])

    def test_dry_run_writes_nothing_and_issues_dir(self):
        rc, out, _ = run()
        self.assertEqual(rc, 0)
        self.assertIn("dry run", out)
        with tempfile.TemporaryDirectory() as d:
            rc, out, _ = run("--write-issues", d)
            self.assertEqual(rc, 0)
            files = sorted(os.listdir(d))
            self.assertTrue(files and all(f.endswith(".md") for f in files))
            self.assertNotIn("BBB-2.md", files)
            self.assertIn("nothing was sent", out)


if __name__ == "__main__":
    unittest.main()
