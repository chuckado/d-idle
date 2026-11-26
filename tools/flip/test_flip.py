"""Tests for flip.py with a fake sandbox and a fake clock (no container, no host). Run: python3 -I tools/flip/test_flip.py"""
import json
import os
import sys
import tempfile
import unittest
from datetime import datetime, timedelta, timezone

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import flip as F  # noqa: E402

T0 = datetime(2026, 11, 26, 12, 0, 0, tzinfo=timezone.utc)
HARD = ["potion-not-drunk", "life-zero-alive", "refusal-storm", "xp-frozen", "mode-no-exit", "corpse-unresolved"]


class Clock:
    def __init__(self):
        self.t = T0

    def __call__(self):
        return self.t


class FakeSandbox:
    """Holds a tuning text and a status feed; failures are switches."""

    def __init__(self):
        self.tuning = json.dumps({"profile": "player", "overrides": {"combat.chickenPct": 22.0}})
        self.counts = {n: 0 for n in HARD}
        self.since = "2026-11-26T11:00:00Z"
        self.restarts = 0
        self.fail_restart = 0   # number of next restarts that fail
        self.fail_write = False
        self.fail_status = False
        self.log = []

    def read_tuning(self):
        return self.tuning

    def write_tuning(self, text):
        if self.fail_write:
            raise F.FlipError("write refused")
        self.tuning = text
        self.log.append("write")

    def remove_tuning(self):
        self.tuning = None

    def restart(self):
        if self.fail_restart:
            self.fail_restart -= 1
            raise F.FlipError("docker restart failed")
        self.restarts += 1
        self.since = (T0 + timedelta(minutes=self.restarts)).strftime("%Y-%m-%dT%H:%M:%SZ")
        self.counts = {n: 0 for n in HARD}  # a restart resets the counters
        self.log.append("restart")

    def status(self):
        if self.fail_status:
            raise F.FlipError("unreachable")
        inv = [{"name": n, "hard": True, "count": c, "since": self.since} for n, c in self.counts.items()]
        inv.append({"name": "town-thrash", "hard": False, "count": 99, "since": self.since})
        return {"invariants": inv}


class FlipTest(unittest.TestCase):
    def setUp(self):
        self.d = tempfile.TemporaryDirectory()
        self.addCleanup(self.d.cleanup)
        self.lp = os.path.join(self.d.name, "darklaunch.json")
        row = {"name": "town.plan", "built_by": "x", "dark_since": "2026-11-20", "why_off": "w", "enable_criteria": "c",
               "owner": "coordinator", "review_by": "2026-12-09", "status": "dark", "criteria_met": "U"}
        with open(self.lp, "w") as f:
            json.dump({"version": 1, "items": [row]}, f)
        self.sb = FakeSandbox()
        self.clock = Clock()
        self.out = []
        self.f = F.Flipper(lambda inst: self.sb, F.Store(os.path.join(self.d.name, "st")), F.Ledger(self.lp), clock=self.clock,
                           out=self.out.append, env={})

    def row(self):
        return F.Ledger(self.lp).row("town.plan")

    def flip(self, window=3600):
        return self.f.flip("sb1", "town.plan", 2.0, "scenario green", window)

    def test_flip_writes_backup_knob_and_ledger(self):
        w = self.flip()
        self.assertEqual(json.loads(self.sb.tuning)["overrides"], {"combat.chickenPct": 22.0, "town.plan": 2.0})
        self.assertEqual(self.sb.restarts, 1)
        self.assertTrue(w["tag"].startswith("bak-flip-"))
        self.assertEqual(self.row()["status"], "trial")
        self.assertEqual(self.row()["trial_started"], "2026-11-26T12:00:00Z")

    def test_breach_in_window_reverts(self):
        self.flip()
        self.clock.t += timedelta(minutes=10)
        self.assertEqual(self.f.poll("sb1"), "open")
        self.sb.counts["potion-not-drunk"] = 1
        self.assertEqual(self.f.poll("sb1"), "reverted")
        self.assertEqual(json.loads(self.sb.tuning), {"profile": "player", "overrides": {"combat.chickenPct": 22.0}})
        self.assertEqual(self.sb.restarts, 2)
        r = self.row()
        self.assertEqual((r["status"], r["revert_reason"].split(":")[0]), ("dark", "hard invariant breach"))
        self.assertIn("potion-not-drunk+1", r["revert_reason"])
        self.assertIsNone(self.f.store.get("sb1"))
        self.assertTrue(any(o.startswith("REVERT ") for o in self.out))

    def test_soft_counter_does_not_revert(self):
        self.flip()  # town-thrash is 99 in the feed and not hard
        self.clock.t += timedelta(minutes=1)
        self.assertEqual(self.f.poll("sb1"), "open")

    def test_no_breach_closes_row(self):
        self.flip(window=600)
        self.clock.t += timedelta(minutes=5)
        self.assertEqual(self.f.poll("sb1"), "open")
        self.clock.t += timedelta(minutes=6)
        self.assertEqual(self.f.poll("sb1"), "closed")
        self.assertEqual(self.row()["status"], "enabled")
        self.assertIn("overrides", self.sb.tuning)
        self.assertEqual(json.loads(self.sb.tuning)["overrides"]["town.plan"], 2.0)
        self.assertIsNone(self.f.store.get("sb1"))

    def test_breach_after_window_end_still_reverts_first(self):
        self.flip(window=600)
        self.clock.t += timedelta(minutes=30)
        self.sb.counts["xp-frozen"] = 2
        self.assertEqual(self.f.poll("sb1"), "reverted")

    def test_flip_restart_failure_restores_and_reports(self):
        self.sb.fail_restart = 1
        orig = self.sb.tuning
        with self.assertRaises(F.FlipError) as c:
            self.flip()
        self.assertIn("restored", str(c.exception))
        self.assertEqual(self.sb.tuning, orig)  # never left half set
        self.assertEqual(self.row()["status"], "dark")
        self.assertIsNone(self.f.store.get("sb1"))

    def test_flip_restart_and_restore_both_fail_stays_pending(self):
        self.sb.fail_restart = 2
        with self.assertRaises(F.FlipError) as c:
            self.flip()
        self.assertIn("pending", str(c.exception))
        self.assertEqual(self.sb.tuning, json.dumps(json.loads(self.sb.tuning)))  # a whole file
        self.assertIsNotNone(self.f.store.get("sb1"))
        self.assertEqual(self.row()["status"], "dark")

    def test_revert_restart_failure_is_reported_and_retried(self):
        self.flip()
        self.sb.counts["life-zero-alive"] = 1
        self.sb.fail_restart = 1
        self.assertEqual(self.f.poll("sb1"), "open")
        self.assertTrue(any(o.startswith("REVERT FAILED") and "restart" in o for o in self.out))
        self.assertEqual(json.loads(self.sb.tuning)["overrides"], {"combat.chickenPct": 22.0})  # file already restored
        self.assertEqual(self.row()["status"], "trial")
        self.assertEqual(self.f.poll("sb1"), "reverted")  # next poll retries
        self.assertEqual(self.row()["status"], "dark")

    def test_revert_write_failure_is_reported(self):
        self.flip()
        self.sb.counts["mode-no-exit"] = 1
        self.sb.fail_write = True
        self.assertEqual(self.f.poll("sb1"), "open")
        self.assertTrue(any("REVERT FAILED" in o and "write" in o for o in self.out))
        self.sb.fail_write = False
        self.assertEqual(self.f.poll("sb1"), "reverted")

    def test_unreadable_status_reverts_after_limit(self):
        self.flip()
        self.sb.fail_status = True
        for _ in range(F.HARD_MISS_LIMIT - 1):
            self.assertEqual(self.f.poll("sb1"), "open")
        self.assertEqual(self.f.poll("sb1"), "reverted")
        self.assertIn("unreadable", self.row()["revert_reason"])

    def test_restart_inside_window_rebaselines(self):
        self.flip()
        self.sb.restart()  # counters back to zero, new since
        self.assertEqual(self.f.poll("sb1"), "open")
        self.sb.counts["corpse-unresolved"] = 1
        self.assertEqual(self.f.poll("sb1"), "reverted")

    def test_guards(self):
        self.flip()
        with self.assertRaises(F.FlipError):
            self.flip()  # one window per sandbox
        f2 = F.Flipper(lambda i: self.sb, F.Store(os.path.join(self.d.name, "st2")), F.Ledger(self.lp), clock=self.clock,
                       out=self.out.append, env={"D2BOT_RUN": "r1"})
        with self.assertRaises(F.FlipError):
            f2.flip("sb2", "town.plan", 1.0, "x", 60)
        with self.assertRaises(F.FlipError):
            self.f.flip("sb3", "no.such", 1.0, "x", 60)

    def test_watch_once_and_loop(self):
        self.flip(window=120)
        slept = []

        def sleep(s):
            slept.append(s)
            self.clock.t += timedelta(seconds=s)
        self.f.sleep = sleep
        self.assertEqual(self.f.watch(interval=60), ["closed"])
        self.assertGreaterEqual(len(slept), 1)

    def test_no_backup_file_restores_by_removal(self):
        self.sb.tuning = None
        self.sb.read_tuning = lambda: None
        self.flip()
        self.sb.counts["xp-frozen"] = 1
        self.f.poll("sb1")
        self.assertIsNone(self.sb.tuning)

    def test_parse_window(self):
        self.assertEqual(F.parse_window("24h"), 86400)
        self.assertEqual(F.parse_window("90s"), 90)
        with self.assertRaises(F.FlipError):
            F.parse_window("soon")


if __name__ == "__main__":
    unittest.main()
