import io
import json
import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import anthropic_http  # noqa: E402
import haiku_tick  # noqa: E402
import pr_review  # noqa: E402

FAKE_KEY = "sk-" + "ant-" + "fake" + "x" * 8  # built from fragments on purpose


def fdiff(path, body="+x\n"):
    return "diff --git a/%s b/%s\n--- a/%s\n+++ b/%s\n@@ -1 +1 @@\n%s" % (path, path, path, path, body)


class Filter(unittest.TestCase):
    def test_skips_bulk(self):
        d = fdiff("docs/probes/a.log") + fdiff("x.png") + fdiff("docs/research/n.md") + fdiff("internal/a.go")
        text, notes = pr_review.filter_diff(d)
        self.assertIn("internal/a.go", text)
        self.assertNotIn("probes", text)
        self.assertEqual(len(notes), 3)

    def test_truncation(self):
        big = fdiff("a.go", "+" + "y" * 80 + "\n") * 1 + fdiff("b.go", ("+" + "z" * 80 + "\n") * 500)
        text, notes = pr_review.filter_diff(big, max_bytes=10000)
        self.assertLessEqual(len(text.encode()), 10100)
        self.assertIn("[truncated]", text)
        self.assertTrue(any("truncated" in n for n in notes))

    def test_omitted_when_no_room(self):
        d = fdiff("a.go", ("+" + "y" * 80 + "\n") * 100) + fdiff("b.go")
        text, notes = pr_review.filter_diff(d, max_bytes=len(fdiff("a.go", ("+" + "y" * 80 + "\n") * 100)))
        self.assertNotIn("b.go", text)
        self.assertIn("omitted (size cap): b.go", notes)


class Marker(unittest.TestCase):
    def test_find(self):
        cs = [{"id": 1, "body": "hi"}, {"id": 2, "body": pr_review.MARKER + "\nold", "user": {"type": "Bot"}}]
        self.assertEqual(pr_review.find_marker_comment(cs), 2)
        self.assertIsNone(pr_review.find_marker_comment(cs[:1]))

    def test_user_marker_ignored(self):
        cs = [{"id": 3, "body": pr_review.MARKER, "user": {"type": "User"}}]
        self.assertIsNone(pr_review.find_marker_comment(cs))


class Prompt(unittest.TestCase):
    def test_build(self):
        p = pr_review.build_prompt("T", "B", "DIFF\n", ["n1"])
        self.assertIn("PR title: T", p)
        self.assertIn("- n1", p)
        self.assertIn("<diff>\nDIFF\n</diff>", p)
        self.assertIn("VERIFIED LIVE", pr_review.SYSTEM)
        self.assertIn("standard-library-only", pr_review.SYSTEM)


class Run(unittest.TestCase):
    def env(self, **kw):
        e = {"GITHUB_REPOSITORY": "o/r", "PR_NUMBER": "7", "GITHUB_TOKEN": "t", "ANTHROPIC_API_KEY": FAKE_KEY}
        e.update(kw)
        return e

    def make(self, changed=2, existing=None):
        calls = []

        def gh(token, method, path, body=None, accept=""):
            calls.append((method, path, body))
            if path.endswith("/comments?per_page=100"):
                return json.dumps(existing or []).encode()
            if method == "GET" and "/pulls/" in path:
                if "diff" in accept:
                    return fdiff("internal/a.go").encode()
                return json.dumps({"title": "t", "body": "b", "changed_files": changed}).encode()
            return b"{}"
        return gh, calls

    def test_posts_new_and_updates(self):
        sent = []

        def http(url, headers, body):
            sent.append((url, headers, body))
            return 200, {"content": [{"type": "text", "text": "looks fine"}]}
        gh, calls = self.make()
        self.assertEqual(pr_review.run(self.env(), gh, http), "posted")
        self.assertEqual(sent[0][2]["max_tokens"], 1500)
        self.assertEqual(sent[0][2]["model"], "claude-sonnet-5-5")
        self.assertEqual(sent[0][1]["x-api-key"], FAKE_KEY)
        self.assertEqual(calls[-1][0], "POST")
        gh, calls = self.make(existing=[{"id": 9, "body": pr_review.MARKER, "user": {"type": "Bot"}}])
        pr_review.run(self.env(REVIEW_MODEL="claude-haiku-5-5"), gh, http)
        self.assertEqual(sent[1][2]["model"], "claude-haiku-5-5")
        self.assertEqual(calls[-1][0], "PATCH")

    def test_too_many_files_no_api_call(self):
        def http(*a):
            raise AssertionError("no call expected")
        gh, calls = self.make(changed=500)
        self.assertEqual(pr_review.run(self.env(), gh, http), "too many files")
        self.assertIn("skipped", calls[-1][2]["body"])

    def test_no_key_comment(self):
        gh, calls = self.make()
        self.assertEqual(pr_review.run(self.env(ANTHROPIC_API_KEY=""), gh, None), "no key")
        self.assertIn("not set", calls[-1][2]["body"])

    def test_api_error_has_no_key(self):
        with self.assertRaises(RuntimeError) as cm:
            anthropic_http.message(FAKE_KEY, "m", "s", "u", 5, http=lambda *a: (401, {"error": {"type": "authentication_error"}}))
        self.assertNotIn(FAKE_KEY, str(cm.exception))
        self.assertIn("401", str(cm.exception))


class Tick(unittest.TestCase):
    def test_tick_and_dry_run(self):
        sent = []

        def http(url, headers, body):
            sent.append(body)
            return 200, {"content": [{"type": "text", "text": "merged: a, b"}]}
        facts = {"running": ["x"], "merged": ["a", "b"], "blocked": ""}
        self.assertEqual(haiku_tick.tick(facts, FAKE_KEY, http), "merged: a, b")
        self.assertEqual(sent[0]["model"], "claude-haiku-5-5")
        self.assertEqual(sent[0]["max_tokens"], 600)
        self.assertNotIn("blocked", sent[0]["messages"][0]["content"])
        out = io.StringIO()
        old, sys.stdout = sys.stdout, out
        try:
            haiku_tick.main([], io.StringIO(json.dumps(facts)))
        finally:
            sys.stdout = old
        self.assertIn("dry run", out.getvalue())


if __name__ == "__main__":
    unittest.main()
