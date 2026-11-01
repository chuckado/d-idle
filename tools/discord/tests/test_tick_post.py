import io
import json
import os
import sys
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, ".."))
import tick_post  # noqa: E402

FAKE_KEY = "sk-" + "ant-" + "fake" + "y" * 8  # built from fragments on purpose
FACTS = {"running": ["navi", "combi"], "merged": ["fixrecast"], "launched": [], "waiting": "go for the release"}
GOOD = "running: navi, combi\nmerged: fixrecast\nwaiting: go for the release"


class FakeHTTP:
    def __init__(self, *texts, status=200, stop="end_turn"):
        self.stop = stop
        self.texts, self.status, self.calls, self.seen = list(texts), status, 0, []

    def __call__(self, url, headers, body, *a, **kw):
        self.calls += 1
        self.seen.append(body)
        if self.status != 200:
            return self.status, {"error": {"type": "overloaded_error"}}
        t = self.texts.pop(0)
        if isinstance(t, Exception):
            raise t
        return 200, {"content": [{"type": "text", "text": t}], "stop_reason": self.stop,
                     "usage": {"input_tokens": 100, "output_tokens": 40}}


def run(facts, http, *argv, key=lambda: FAKE_KEY, sender=None):
    out, err = io.StringIO(), io.StringIO()
    rc = tick_post.main(list(argv), stdin=io.StringIO(json.dumps(facts)), http=http, sender=sender, key_reader=key,
                        now=1793543400.0, out=out, err=err)
    return rc, out.getvalue(), err.getvalue()


class Post(unittest.TestCase):
    def test_normal_dry_run(self):
        h = FakeHTTP(GOOD)
        rc, out, err = run(FACTS, h)
        self.assertEqual(rc, 0)
        self.assertIn("source=haiku", out)
        self.assertIn("input_tokens=100 output_tokens=40", out)
        self.assertIn("- Running: navi, combi", out)
        self.assertEqual(h.calls, 1)
        self.assertEqual(h.seen[0]["model"], "claude-haiku-5-5")
        self.assertEqual(h.seen[0]["max_tokens"], 1200)
        self.assertEqual(err, "")

    def test_send_uses_sender_once(self):
        sent = []
        rc, out, _ = run(FACTS, FakeHTTP(GOOD), "--send", sender=lambda *a: (sent.append(a) or (True, "posted to #progress")))
        self.assertEqual(rc, 0)
        self.assertEqual(len(sent), 1)
        self.assertEqual(sent[0][0], "progress")
        self.assertIn("posted to #progress", out)

    def test_guard_reword(self):
        bad = GOOD + "\nblocked: some tiny word runs like this one here ok"
        facts = dict(FACTS, blocked="nothing big")
        ok = GOOD + "\nblocked: some tiny word runs, like this one here, ok"
        h = FakeHTTP(bad, ok)
        rc, out, err = run(dict(FACTS, blocked="some tiny word runs like this one here ok"), h)
        self.assertEqual(h.calls, 2)
        self.assertIn("one reword", err)
        self.assertIn("source=haiku", out)
        self.assertIn("input_tokens=200", out)
        self.assertIn("some tiny word runs like this one here ok", h.seen[1]["messages"][0]["content"])
        del facts

    def test_five_short_words_fallback_when_reword_fails(self):
        phrase = "some tiny word runs like this"
        f = dict(FACTS, blocked=phrase)
        bad = GOOD + "\nblocked: " + phrase
        h = FakeHTTP(bad, bad)
        rc, out, err = run(f, h)
        # the plain rendering has the same run; the separator repair now makes it postable (it used to be refused)
        self.assertEqual(rc, 0, err)
        self.assertIn("guard run found", err)
        self.assertIn("some tiny word, runs like this", out)

    def test_fallback_on_http_error(self):
        rc, out, err = run(FACTS, FakeHTTP(status=529))
        self.assertEqual(rc, 0)
        self.assertIn("source=deterministic", out)
        self.assertIn("plain rendering", err)
        self.assertIn("http 529", err)

    def test_fallback_on_timeout_and_no_key(self):
        rc, out, err = run(FACTS, FakeHTTP(TimeoutError("slow")))
        self.assertIn("source=deterministic", out)
        def nokey():
            raise FileNotFoundError("nope")
        rc, out, err = run(FACTS, FakeHTTP(GOOD), key=nokey)
        self.assertEqual(rc, 0)
        self.assertIn("source=deterministic", out)

    def test_fallback_when_facts_dropped(self):
        rc, out, err = run(FACTS, FakeHTTP("running: navi, combi\nwaiting: go for the release"))
        self.assertIn("source=deterministic", out)
        self.assertIn("- Merged: fixrecast", out)

    def test_thinking_disabled_and_budget(self):
        h = FakeHTTP(GOOD)
        run(FACTS, h)
        self.assertEqual(h.seen[0]["thinking"], {"type": "disabled"})
        self.assertGreaterEqual(h.seen[0]["max_tokens"], 1200)
        self.assertNotIn("temperature", h.seen[0])

    def test_truncated_response_falls_back_and_logs(self):
        rc, out, err = run(FACTS, FakeHTTP(GOOD[:30], stop="max_tokens"))
        self.assertEqual(rc, 0)
        self.assertIn("source=deterministic", out)
        self.assertIn("stop_reason max_tokens", err)

    def test_reworded_but_complete_accepted(self):
        f = dict(FACTS, merged=["fixrecast (PR 41)"], waiting="go for the release 1.2")
        h = FakeHTTP("Running: Navi, Combi\nmerged: fixrecast (PR 41)\nWaiting: The release 1.2 needs a go")
        rc, out, err = run(f, h)
        self.assertIn("source=haiku", out, err)

    def test_dropped_or_invented_number_rejected(self):
        f = dict(FACTS, merged=["fixrecast (PR 41)"])
        rc, out, err = run(f, FakeHTTP("running: navi, combi\nmerged: fixrecast\nwaiting: go for the release"))
        self.assertIn("source=deterministic", out)
        self.assertIn("dropped 41", err)
        rc, out, err = run(f, FakeHTTP("running: navi, combi\nmerged: fixrecast (PR 41), PR 42\nwaiting: go for the release"))
        self.assertIn("invented number 42", err)

    def test_bad_line_logged_without_key(self):
        rc, out, err = run(FACTS, FakeHTTP("Here is your tick: sure thing"))
        self.assertIn("first bad line", err)
        self.assertNotIn(FAKE_KEY, err)

    def test_idle_refused_without_api(self):
        h = FakeHTTP(GOOD)
        rc, out, err = run({"running": ["navi"], "merged": [], "waiting": "none"}, h)
        self.assertEqual(rc, 2)
        self.assertIn("idle tick", err)
        self.assertEqual(h.calls, 0)

    def test_key_never_in_output(self):
        for h in (FakeHTTP(GOOD), FakeHTTP(status=401), FakeHTTP(TimeoutError(FAKE_KEY))):
            rc, out, err = run(FACTS, h, "--send", sender=lambda *a: (True, "posted"))
            self.assertNotIn(FAKE_KEY, out + err)
            self.assertNotIn(FAKE_KEY, json.dumps(h.seen[0]["messages"]) if h.seen else "")

    def test_no_api_flag(self):
        h = FakeHTTP(GOOD)
        rc, out, _ = run(FACTS, h, "--no-api")
        self.assertEqual(h.calls, 0)
        self.assertIn("source=deterministic", out)




REFUSED_TODAY = {"running": "FORKFREEZE: freeze resume fix on the fork",
                 "merged": "Freeze-resume fix proof and independent verifier (READY WITH CHANGES, held for your go) and the Frozen Armor recast drill held",
                 "waiting": ["Nothing."]}


class Repair(unittest.TestCase):
    def guard(self, text):
        return tick_post.notify.check_text("", text, "")

    def test_repair_text_passes_real_guard_and_keeps_tokens(self):
        t = "- Blocked: aaaa bbbbb cccccc dddd eeeee ffff gggg hhhh iiii jjjj 12.5 PR-41"
        self.assertEqual(self.guard(t), ["cdkey"])
        fixed, runs = tick_post.repair_runs(t)
        self.assertEqual(self.guard(fixed), [])
        self.assertTrue(runs)
        self.assertEqual(fixed.replace(",", ""), t)  # only commas were inserted
        self.assertIn("12.5 PR-41", fixed)

    def test_hyphen_run_and_clean_text(self):
        fixed, runs = tick_post.repair_runs("aaaa-bbbb-cccc-dddd-eeee")
        self.assertEqual(self.guard(fixed), [])
        self.assertEqual(fixed.replace(", ", "-"), "aaaa-bbbb-cccc-dddd-eeee")
        clean = "- Running: navi, combi (PR 41)"
        self.assertEqual(tick_post.repair_runs(clean), (clean, []))

    def test_gives_up_after_five(self):
        t = " ".join(["word"] * 40)
        fixed, runs = tick_post.repair_runs(t)
        self.assertEqual(len(runs), 5)

    def test_refused_today_facts_post_plain_after_repair(self):
        v = tick_post.facts_to_vals(REFUSED_TODAY)
        self.assertEqual(self.guard(tick_post.render(v, 1793543400.0)[1]), ["cdkey"])  # refused before the repair
        for facts in (REFUSED_TODAY, dict(REFUSED_TODAY, blocked="aaaa bbbb cccc dddd eeee 7")):
            rc, out, err = run(facts, FakeHTTP(status=529), "--no-api")
            self.assertEqual(rc, 0, err)
            self.assertEqual(self.guard(out), [])
            self.assertIn("Freeze-resume", out)
            self.assertNotIn("Waiting", out)

    def test_haiku_text_repaired_after_failed_reword(self):
        f = dict(REFUSED_TODAY, blocked="aaaa bbbb cccc dddd eeee 7")
        bad = ("running: FORKFREEZE: freeze resume fix on the fork\nmerged: " + REFUSED_TODAY["merged"] +
               "\nblocked: aaaa bbbb cccc dddd eeee 7")
        h = FakeHTTP(bad, bad)
        rc, out, err = run(f, h)
        self.assertEqual(rc, 0, err)
        self.assertIn("source=haiku", out)
        self.assertIn("guard run found", err)
        self.assertIn("aaaa bbbb cccc, dddd eeee 7", out)

    def test_waiting_nothing_not_invented(self):
        good = ("running: FORKFREEZE: freeze resume fix on the fork\nmerged: " + REFUSED_TODAY["merged"] +
                "\nwaiting: Nothing is waiting on you right now")
        rc, out, err = run(REFUSED_TODAY, FakeHTTP(good))
        self.assertEqual(rc, 0, err)
        self.assertNotIn("invented", err)
        self.assertNotIn("Waiting", out)

if __name__ == "__main__":
    unittest.main()