import io
import json
import os
import sys
import unittest
from contextlib import redirect_stderr, redirect_stdout
from types import SimpleNamespace as NS

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import claim


class StubGh:
    def __init__(self, iss):
        self.iss, self.writes = iss, []

    def __call__(self, argv, stdin=None):
        if argv[0] == "issue" and argv[1] == "view":
            return 0, json.dumps(self.iss), ""
        self.writes.append((argv, stdin))
        return 0, "", ""


def issue(**kw):
    d = {"number": 3, "state": "OPEN", "labels": [{"name": "work-item"}], "assignees": [], "comments": []}
    d.update(kw)
    return d


def call(stub, **kw):
    a = NS(cmd="claim", item="PALSKILL", job="job x", note="Running, expected soon.", send=False, doc="", close=False)
    a.__dict__.update(kw)
    with redirect_stdout(io.StringIO()) as out, redirect_stderr(io.StringIO()):
        try:
            rc = claim.run(a, runner=stub, now="2026-11-02")
        except claim.Refused as e:
            rc = "refused: %s" % e
    return rc, out.getvalue()


class ClaimTests(unittest.TestCase):
    def setUp(self):
        self.log = os.path.join(claim.ROOT, claim.LOG)
        self.before = open(self.log).read() if os.path.exists(self.log) else None

    def tearDown(self):
        if self.before is None:
            if os.path.exists(self.log):
                os.remove(self.log)
        else:
            open(self.log, "w").write(self.before)

    def test_dry_run_posts_nothing_and_logs_nothing(self):
        s = StubGh(issue())
        rc, out = call(s)
        self.assertEqual(rc, 0)
        self.assertEqual(s.writes, [])
        self.assertIn("Claimed by the maintainers", out)
        self.assertEqual(os.path.exists(self.log), self.before is not None)

    def test_send_posts_comment_and_logs(self):
        s = StubGh(issue())
        self.assertEqual(call(s, send=True)[0], 0)
        self.assertEqual(len(s.writes), 1)
        self.assertIn("claim:claim:PALSKILL", s.writes[0][1])
        self.assertIn("PALSKILL\t3\tclaim", open(self.log).read())

    def test_human_claim_refused(self):
        s = StubGh(issue(comments=[{"body": "claiming", "author": {"login": "x"}}]))
        rc, _ = call(s, send=True)
        self.assertTrue(str(rc).startswith("refused: a human"))
        self.assertEqual(s.writes, [])

    def test_assigned_closed_unlabelled_refused(self):
        self.assertIn("assigned", call(StubGh(issue(assignees=[{"login": "x"}])), send=True)[0])
        self.assertIn("closed", call(StubGh(issue(state="CLOSED")), send=True)[0])
        self.assertIn("not labelled", call(StubGh(issue(labels=[])), send=True)[0])

    def test_second_claim_is_noop(self):
        s = StubGh(issue(comments=[{"body": "x\n<!-- claim:claim:PALSKILL -->", "author": {"login": "m"}}]))
        self.assertEqual(call(s, send=True)[0], 0)
        self.assertEqual(s.writes, [])

    def test_done_needs_existing_doc_and_closes(self):
        self.assertIn("does not exist", call(StubGh(issue()), cmd="done", doc="docs/nope.md", close=True, send=True)[0])
        s = StubGh(issue(state="CLOSED"))
        self.assertEqual(call(s, cmd="done", doc="docs/community/open-work.md", close=True, send=True)[0], 0)
        self.assertEqual([w[0][1] for w in s.writes], ["comment", "close"])
        self.assertIn("completed", s.writes[1][0])

    def test_unknown_item(self):
        self.assertIn("no issue", call(StubGh(issue()), item="NOPE")[0])


class WatchTests(unittest.TestCase):
    def test_classify(self):
        for body, want in (("claiming", "CLAIM"), ("can I take this", "CLAIM"), ("see https://x.example/y", "SUBMISSION"),
                           ("is it done?", "QUESTION"), ("I am releasing this", "WITHDRAWAL"), ("thanks", "OTHER")):
            self.assertEqual(claim.classify(body), want, body)

    def test_watch_lists_changes_only(self):
        import tempfile
        c1 = [{"body": "claiming", "createdAt": "2026-11-01T00:00:00Z", "author": {"login": "h"}}]
        data = [{"number": 1, "state": "OPEN", "title": "X: t", "comments": c1}, {"number": 2, "state": "OPEN", "title": "Y: t", "comments": []}]
        runner = lambda argv, stdin=None: (0, json.dumps(data), "")
        with tempfile.TemporaryDirectory() as d, redirect_stdout(io.StringIO()):
            cache = os.path.join(d, "c", "w.json")
            self.assertEqual([r[0] for r in claim.watch(runner, cache, update=True)], [1])
            self.assertEqual(claim.watch(runner, cache), [])
            c1.append({"body": "done? see PR", "createdAt": "2026-11-02T00:00:00Z", "author": {"login": "h"}})
            rows = claim.watch(runner, cache)
            self.assertEqual((rows[0][0], rows[0][3]), (1, "SUBMISSION"))


if __name__ == "__main__":
    unittest.main()
