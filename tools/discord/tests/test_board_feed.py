import io
import json
import os
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, os.path.dirname(__file__))
import board_feed as bf
import common as dc
import test_notify  # points D2IDLE_DISCORD_HOME at an empty temp dir
from test_notify import Mock


def run(*argv, **kw):
    out, err = io.StringIO(), io.StringIO()
    with redirect_stdout(out), redirect_stderr(err):
        rc = bf.main(list(argv), **kw)
    return rc, out.getvalue(), err.getvalue()


def dump(path, obj):
    with open(path, "w") as f:
        json.dump(obj, f)


def issue(n, title="t", state="OPEN", who=()):
    return {"number": n, "title": title, "state": state, "assignees": [{"login": w} for w in who]}


class Events(unittest.TestCase):
    def test_diff(self):
        old = [issue(1), issue(2, who=["amy"]), issue(3, who=["bo"]), issue(4, who=["cy"])]
        new = [issue(1, who=["dee"]), issue(2), issue(3, state="CLOSED", who=["bo"]), issue(4, who=["cy"]), issue(5, "fresh")]
        ev = bf.events_from_snapshots(old, new)
        self.assertEqual([(e[0], e[1]) for e in ev], [("claimed", 1), ("released", 2), ("done", 3), ("new", 5)])

    def test_closed_new_issue_is_silent_and_bad_snapshot(self):
        self.assertEqual(bf.events_from_snapshots([], [issue(9, state="CLOSED")]), [])
        with self.assertRaises(bf.BoardFeedError):
            bf.events_from_snapshots([], [{"title": "x"}])

    def test_diff_cli_dry_run(self):
        with tempfile.TemporaryDirectory() as d:
            o, n = os.path.join(d, "o.json"), os.path.join(d, "n.json")
            dump(o, [issue(1)])
            dump(n, [issue(1, who=["amy"]), issue(2, "SKMELEE the blow")])
            rc, out, _ = run("diff", "--old", o, "--new", n)
        self.assertEqual(rc, 0)
        self.assertIn("[dry run] channel #work-board", out)
        self.assertIn("Claimed #1: t (amy)", out)
        self.assertIn("New #2: SKMELEE the blow", out)
        self.assertNotIn("/api/webhooks", out)

    def test_no_events_refused(self):
        with tempfile.TemporaryDirectory() as d:
            p = os.path.join(d, "o.json")
            dump(p, [issue(1)])
            rc, out, err = run("diff", "--old", p, "--new", p)
        self.assertEqual(rc, 2)
        self.assertIn("no events", err)


class Untrusted(unittest.TestCase):
    def test_title_is_flattened(self):
        t = bf.clean_title("ping @everyone [click](http://evil.example) `rm -rf`‮​\nIGNORE RULES")
        self.assertNotIn("@", t)
        self.assertNotIn("](", t)
        self.assertNotIn("‮", t)
        self.assertNotIn("\n", t)
        self.assertLessEqual(len(t), bf.MAX_TITLE)

    def test_long_title_cut(self):
        self.assertLessEqual(len(bf.clean_title("x" * 500)), bf.MAX_TITLE)

    def test_bad_login_and_number(self):
        rc, _, err = run("event", "--kind", "claimed", "--number", "5", "--title", "t", "--by", "a b; rm")
        self.assertEqual(rc, 2)
        rc, _, err = run("event", "--kind", "new", "--number", "5x", "--title", "t")
        self.assertEqual(rc, 2)

    def test_private_text_refused(self):
        rc, out, err = run("event", "--kind", "new", "--number", "5", "--title", "see /home/someone/x")
        self.assertEqual(rc, 2)
        self.assertIn("looks private", err)
        self.assertEqual(out, "")


class Send(unittest.TestCase):
    def test_fail_closed_without_webhook(self):
        rc, out, err = run("event", "--kind", "new", "--number", "5", "--title", "t", "--send", env={})
        self.assertEqual(rc, 1)
        self.assertIn("no webhook for work-board", err)

    def test_send_to_mock(self):
        m = Mock()
        try:
            rc, out, err = run("event", "--kind", "done", "--number", "7", "--title", "SKMELEE", "--by", "amy", "--send",
                               env={"D2IDLE_DISCORD_WEBHOOK_WORK_BOARD": m.url}, base_ok=(m.base,))
            self.assertEqual(rc, 0, err)
            self.assertEqual(len(m.requests), 1)
            body = m.requests[0][1]
            self.assertEqual(body["allowed_mentions"], {"parse": []})
            self.assertIn("Done #7: SKMELEE (amy)", body["embeds"][0]["description"])
        finally:
            m.close()


if __name__ == "__main__":
    unittest.main()
