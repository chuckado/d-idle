#!/usr/bin/env python3
"""Tests of tools/backlog/catchup_done.py on a temp git repo. Run: python3 -I tools/backlog/test_catchup_done.py"""
import os
import re
import stat
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import catchup_done  # noqa: E402
import test_drift  # noqa: E402

Drift = test_drift.Drift
del test_drift.Drift  # keep unittest from collecting the parent's tests twice


class Catchup(Drift):
    def merged(self):
        self.put(1, "one", "Proposed", [1, 1])
        self.put(2, "two", "Proposed", [1, 0])
        self.put(3, "three", "Done", [1, 1])
        self.put(4, "four", "Proposed", [1, 1])  # not merged
        self.commit("base")
        for c in ("one", "two", "three"):
            subprocess.run(["git", "checkout", "-q", "-b", c], cwd=self.r, check=True)
            subprocess.run(["git", "commit", "-q", "--allow-empty", "-m", c], cwd=self.r, check=True,
                           env=dict(os.environ, GIT_AUTHOR_NAME="t", GIT_AUTHOR_EMAIL="t@t", GIT_COMMITTER_NAME="t", GIT_COMMITTER_EMAIL="t@t"))
            subprocess.run(["git", "checkout", "-q", "main"], cwd=self.r, check=True)
            subprocess.run(["git", "merge", "-q", "--no-ff", "-m", "Merge agent/%s: x" % c, c], cwd=self.r, check=True,
                           env=dict(os.environ, GIT_AUTHOR_NAME="t", GIT_AUTHOR_EMAIL="t@t", GIT_COMMITTER_NAME="t", GIT_COMMITTER_EMAIL="t@t"))

    def test_dry_run_lists_only_fully_ticked_merged_cards(self):
        self.merged()
        ready, held = catchup_done.candidates(self.r, "main")
        self.assertEqual([t["card"] for t in ready], ["one"])
        self.assertEqual([t["card"] for t in held], ["two"])

    def test_dry_run_changes_nothing_and_apply_calls_backlog_for_ready_only(self):
        self.merged()
        log = self.r / "calls"
        fake = self.r / "fakebacklog"
        fake.write_text('#!/bin/sh\necho "$*" >> %s\n' % log)
        fake.chmod(fake.stat().st_mode | stat.S_IXUSR)
        env = dict(os.environ, BACKLOG_CMD=str(fake))
        run = lambda *a: subprocess.run([sys.executable, str(Path(catchup_done.__file__)), "--repo", str(self.r), "--ref", "main", *a],
                                        capture_output=True, text=True, env=env)
        out = run().stdout
        self.assertIn("1 to set Done, 1 stay open (dry run", out)
        self.assertFalse(log.exists())
        out = run("--apply").stdout
        self.assertIn("applied", out)
        self.assertEqual(log.read_text().strip(), "task edit 1 -s Done")


if __name__ == "__main__":
    unittest.main()
