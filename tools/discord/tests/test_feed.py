import io
import os
import subprocess
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
sys.path.insert(0, os.path.dirname(__file__))
import common as dc
import feed
import notify
from test_notify import Mock  # also points D2IDLE_DISCORD_HOME at an empty temp dir

ROOT = feed.ROOT
SOAK = os.path.join(ROOT, "docs/probes/postfix-soak1.md")
DRILLS = os.path.join(ROOT, "docs/probes/drills/README.md")
PR = "https://github.com/psi29a/d2decomped/pull/153"


def run(*argv, **kw):
    out, err = io.StringIO(), io.StringIO()
    with redirect_stdout(out), redirect_stderr(err):
        rc = feed.main(list(argv), **kw)
    return rc, out.getvalue(), err.getvalue()


class DryRun(unittest.TestCase):
    def test_soak_dry_run(self):
        rc, out, _ = run("soak", "--file", SOAK)
        self.assertEqual(rc, 0)
        self.assertIn("[dry run] channel #progress", out)
        self.assertIn("Verdict: INCONCLUSIVE", out)
        self.assertIn("deaths 0", out)
        self.assertIn("https://github.com/shatteredeagle/d-idle/blob/main/docs/probes/postfix-soak1.md", out)
        self.assertNotIn("/api/webhooks", out)
        self.assertLessEqual(len(out.strip().splitlines()), 7)

    def test_drill_dry_run(self):
        rc, out, _ = run("drill", "--file", DRILLS)
        self.assertEqual(rc, 0)
        self.assertIn("PASS", out)
        self.assertIn("Open findings", out)

    def test_drill_flags(self):
        rc, out, _ = run("drill", "--pass", "5", "--fail", "1", "--blocked", "0", "--finding", "D1.P1 belt column")
        self.assertEqual(rc, 0)
        self.assertIn("PASS 5, FAIL 1, BLOCKED 0", out)
        self.assertIn("D1.P1 belt column", out)

    def test_pr_status_table_and_bad_row(self):
        rc, out, _ = run("pr-status", "--pr", "153|Waypoint window|green|2|waiting")
        self.assertEqual(rc, 0)
        self.assertIn("#153 | Waypoint window | green | 2 | waiting", out)
        rc, _, err = run("pr-status", "--pr", "nonsense")
        self.assertEqual(rc, 2)
        self.assertIn("bad PR row", err)

    def test_mentions_neutralised(self):
        rc, out, _ = run("upstream", "--pr", "153", "--event", "merged", "--title", "ping @everyone", "--url", PR)
        self.assertEqual(rc, 0)
        self.assertNotIn("@everyone", out)


class Refusals(unittest.TestCase):
    def test_soak_without_verdict(self):
        with tempfile.TemporaryDirectory(dir=ROOT) as d:
            p = os.path.join(d, "soak.md")
            with open(p, "w") as f:
                f.write("# soak\nLength: 119 wall min\n0 deaths\n")
            rc, out, err = run("soak", "--file", p)
        self.assertEqual(rc, 2)
        self.assertIn("no verdict line", err)
        self.assertEqual(out, "")

    def test_soak_verdict_flag_ok(self):
        rc, out, _ = run("soak", "--verdict", "PASS", "--duration", "2 h", "--deaths", "0", "--link", SOAK_LINK)
        self.assertEqual(rc, 0)
        self.assertIn("Verdict: PASS", out)

    def test_upstream_url_validation(self):
        bad = ["https://github.com/psi29a/d2decomped/pull/153/files", "https://github.com/psi29a/d2decomped/pull/x",
               "http://github.com/psi29a/d2decomped/pull/153", "https://github.com/other/d2decomped/pull/153",
               "https://github.com/psi29a/d2decomped/pull/153 extra", "", "https://github.com/psi29a/d2decomped/pull/154"]
        for u in bad:
            rc, _, err = run("upstream", "--pr", "153", "--event", "opened", "--title", "t", "--url", u)
            self.assertEqual(rc, 2, u)
            self.assertIn("refused", err)
        rc, _, _ = run("upstream", "--pr", "153", "--event", "bogus", "--title", "t", "--url", PR)
        self.assertEqual(rc, 2)

    def test_release_refused_without_tag(self):
        with tempfile.TemporaryDirectory() as d:
            subprocess.run(["git", "init", "-q", d], check=True)
            url = feed.REPO_URL + "/releases/tag/v1"
            rc, _, err = run("release", "--tag", "v1", "--notes", "n", "--url", url, "--repo", d)
            self.assertEqual(rc, 2)
            self.assertIn("no release exists", err)
            subprocess.run(["git", "-C", d, "-c", "user.name=t", "-c", "user.email=t" + "@example.com", "commit", "-q",
                            "--allow-empty", "-m", "x"], check=True)
            subprocess.run(["git", "-C", d, "tag", "v1"], check=True)
            rc, out, _ = run("release", "--tag", "v1", "--notes", "first notes", "--url", url, "--repo", d)
            self.assertEqual(rc, 0)
            self.assertIn("#announcements", out)
            rc, _, err = run("release", "--tag", "v1", "--notes", "n", "--url", feed.REPO_URL + "/releases/tag/v2", "--repo", d)
            self.assertEqual(rc, 2)

    def test_sanitizer_rejects_private(self):
        for t in ("see /" + "home/x/file", "build on 10.0" + ".0.5", "mail a" + "@b.example", "host box.lan.example.lan"):
            rc, _, err = run("upstream", "--pr", "153", "--event", "opened", "--title", t, "--url", PR)
            self.assertEqual(rc, 2, t)
            self.assertIn("looks private", err)
        hook = "https://discord.com/api/" + "webhooks/123456789/abcDEF"
        rc, _, err = run("pr-status", "--pr", "1|%s|g|0|x" % hook)
        self.assertEqual(rc, 2)
        self.assertNotIn("abcDEF", err)

    def test_bad_channel(self):
        rc, _, err = run("soak", "--verdict", "PASS", "--channel", "general")
        self.assertEqual(rc, 2)


class Tick(unittest.TestCase):
    NOW = "1793543400"  # 2026-11-01 14:30 UTC

    def tick(self, *argv):
        return run("tick", "--now", self.NOW, *argv)

    def test_dry_run(self):
        rc, out, _ = self.tick("--running", "navi,combi", "--merged", "fixrecast,feedtick", "--waiting", "go for X")
        self.assertEqual(rc, 0)
        self.assertIn("[dry run] channel #progress", out)
        self.assertIn("dispatcher tick 2026-11-01 14:30 UTC", out)
        self.assertIn("- Running: navi, combi", out)
        self.assertIn("- Merged: fixrecast, feedtick", out)
        self.assertIn("- Waiting on the maintainer: go for X", out)

    def test_empty_and_idle_refused(self):
        for argv in ((), ("--running", "navi"), ("--merged", "none", "--blocked", "-"), ("--upstream", "no change")):
            rc, out, err = self.tick(*argv)
            self.assertEqual(rc, 2, argv)
            self.assertIn("idle tick", err)
            self.assertEqual(out, "")

    def test_file(self):
        with tempfile.NamedTemporaryFile("w", suffix=".txt", delete=False) as f:
            f.write("Running: a, b\nMerged: c\nBlocked: none\n")
        self.addCleanup(os.unlink, f.name)
        rc, out, _ = self.tick("--file", f.name)
        self.assertEqual(rc, 0)
        self.assertIn("- Merged: c", out)
        self.assertNotIn("Blocked", out)
        with open(f.name, "w") as g:
            g.write("gibberish line\n")
        rc, _, err = self.tick("--file", f.name)
        self.assertEqual(rc, 2)
        self.assertIn("unknown line", err)

    def test_hostile_strings(self):
        rc, out, _ = self.tick("--merged", "ping @everyone")
        self.assertEqual(rc, 0)
        self.assertNotIn("@everyone", out)
        hook = "https://discord.com/api/" + "webhooks/123456789/abcDEF"
        for t in ("see /" + "home/x/file", "build on 10.0" + ".0.5", "mail a" + "@b.example", "host box.lan.example.lan", hook):
            rc, out, err = self.tick("--blocked", t)
            self.assertEqual(rc, 2, t)
            self.assertIn("looks private", err)
            self.assertNotIn("abcDEF", err + out)

    def test_length_cap(self):
        rc, out, _ = self.tick("--merged", ",".join("agent%d" % i for i in range(200)), "--blocked", "x" * 2000,
                               "--waiting", "y" * 2000)
        self.assertEqual(rc, 0)
        body = out.split("\n", 2)[2].strip()
        self.assertLessEqual(len(body), feed.TICK_MAX)
        self.assertLess(feed.TICK_MAX, 4096)

    def test_long_tick_is_whole_below_cap(self):
        text = "z" * 2500
        rc, out, _ = self.tick("--blocked", text, "--waiting", "w" * 1000)
        self.assertEqual(rc, 0)
        self.assertIn(text, out)
        self.assertNotIn("(cut:", out)

    def test_cut_at_line_boundary_with_marker(self):
        lines = ["- Blocked: " + "b" * 1500, "- Waiting on the maintainer: " + "w" * 1500, "- Upstream: " + "u" * 1500]
        text = "\n".join(lines)
        got = feed.tick_cut(text)
        self.assertLessEqual(len(got), feed.TICK_MAX)
        self.assertEqual(got.split("\n")[:2], lines[:2])
        self.assertEqual(got.split("\n")[-1], "(cut: %d chars)" % (len(text) - len("\n".join(lines[:2]))))
        self.assertNotIn("u" * 10, got)
        self.assertEqual(feed.tick_cut("a" * feed.TICK_MAX), "a" * feed.TICK_MAX)
        self.assertIn("(cut: ", feed.tick_cut("a" * (feed.TICK_MAX + 1)))

    def test_channel_is_progress_on_send(self):
        m = Mock()
        self.addCleanup(m.close)
        home = tempfile.TemporaryDirectory()
        self.addCleanup(home.cleanup)
        env = {dc.HOME_ENV: home.name, notify.env_name("progress"): m.url}
        rc, out, err = run("tick", "--now", self.NOW, "--merged", "x", "--send", store=dc.MemoryStore(), env=env,
                           sleep=lambda s: None, base_ok=(m.base,))
        self.assertEqual(rc, 0, err)
        self.assertIn("posted to #progress", out)
        self.assertEqual(len(m.requests), 1)
        self.assertIn("dispatcher tick", m.requests[0][1]["embeds"][0]["title"])


SOAK_LINK = feed.REPO_URL + "/blob/main/docs/probes/postfix-soak1.md"


class Send(unittest.TestCase):
    def setUp(self):
        self.m = Mock()
        self.addCleanup(self.m.close)
        self.home = tempfile.TemporaryDirectory()
        self.addCleanup(self.home.cleanup)
        self.env = {dc.HOME_ENV: self.home.name}
        self.store = dc.MemoryStore()

    def go(self, *argv, **kw):
        return run(*argv, store=self.store, env=self.env, sleep=lambda s: None, base_ok=(self.m.base,), **kw)

    def test_channel_without_webhook_fails_closed_dry_run_still_works(self):
        rc, out, err = self.go("upstream", "--pr", "153", "--event", "opened", "--title", "t", "--url", PR, "--send")
        self.assertEqual(rc, 1)
        self.assertIn("no webhook for d2d-upstream", err)
        self.assertEqual(self.m.requests, [])
        rc, out, _ = self.go("upstream", "--pr", "153", "--event", "opened", "--title", "t", "--url", PR)
        self.assertEqual(rc, 0)
        self.assertIn("#d2d-upstream", out)

    def test_d2d_upstream_posts_once_its_webhook_exists(self):
        self.env[notify.env_name("d2d-upstream")] = self.m.url
        rc, out, _ = self.go("upstream", "--pr", "153", "--event", "opened", "--title", "t", "--url", PR, "--send")
        self.assertEqual(rc, 0, out)
        rc, out, _ = self.go("pr-status", "--pr", "153|t|green|1|waiting", "--channel", "d2d-upstream", "--send")
        self.assertEqual(rc, 0, out)
        self.assertEqual(len(self.m.requests), 2)
        for _, body in self.m.requests:
            self.assertEqual(body["allowed_mentions"], {"parse": []})

    def test_pr_status_goes_to_progress_and_other_channels_fail_closed(self):
        self.env[notify.env_name("progress")] = self.m.url
        rc, out, _ = self.go("pr-status", "--pr", "153|t|green|1|waiting", "--send")
        self.assertEqual(rc, 0, out)
        self.assertEqual(len(self.m.requests), 1)
        rc, _, err = self.go("pr-status", "--pr", "153|t|green|1|waiting", "--channel", "stream", "--send")
        self.assertEqual(rc, 1)
        self.assertIn("no webhook for stream", err)
        self.assertEqual(len(self.m.requests), 1)

    def test_rate_limit_retried_then_posts(self):
        self.m.script = [(429, {"retry_after": 0.01}), (204, None)]
        self.env[notify.env_name("progress")] = self.m.url
        rc, out, _ = self.go("drill", "--pass", "1", "--fail", "0", "--blocked", "0", "--send")
        self.assertEqual(rc, 0)
        self.assertEqual(len(self.m.requests), 2)

    def test_rate_limit_gives_up(self):
        self.m.script = [(429, {"retry_after": 0.01})] * 3
        self.env[notify.env_name("progress")] = self.m.url
        rc, _, err = self.go("drill", "--pass", "1", "--fail", "0", "--blocked", "0", "--send")
        self.assertEqual(rc, 1)
        self.assertIn("gave up", err)

    def test_default_is_dry_run_no_network(self):
        self.env[notify.env_name("progress")] = self.m.url
        self.go("soak", "--verdict", "PASS")
        self.assertEqual(self.m.requests, [])


if __name__ == "__main__":
    unittest.main()
