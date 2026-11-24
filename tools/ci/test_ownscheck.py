#!/usr/bin/env python3
"""Tests for ownscheck.py on throwaway repos. Run: python3 -I tools/ci/test_ownscheck.py"""
import datetime
import os
import shutil
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import ownscheck  # noqa: E402

SG = os.path.join(HERE, "..", "sharedgen", "sharedgen.py")


def card(n, name, owns, extra=""):
    return "---\nid: TASK-%d\ntitle: '%s: x'\nstatus: In Progress\nassignee:\n  - credit-runner\ncard: \"%s\"\nowns:\n%s%s---\n\n## Description\n" % (
        n, name, name, "".join('  - "%s"\n' % o for o in owns), extra)


class T(unittest.TestCase):
    def setUp(self):
        self.d = tempfile.mkdtemp()
        self.addCleanup(shutil.rmtree, self.d)
        os.makedirs(os.path.join(self.d, "tools", "sharedgen"))
        shutil.copy(SG, os.path.join(self.d, "tools", "sharedgen", "sharedgen.py"))
        self.g("init", "-q", "-b", "main")
        self.g("add", "-A")
        self.g("commit", "-qm", "base")

    def g(self, *a):
        subprocess.run(["git", "-C", self.d, "-c", "user.name=t", "-c", "user.email=t@t", *a], check=True, capture_output=True)

    def put(self, rel, text="x\n"):
        p = os.path.join(self.d, rel)
        os.makedirs(os.path.dirname(p), exist_ok=True)
        open(p, "w").write(text)

    def branch(self, name, owns, files, extra="", with_card=True):
        self.g("checkout", "-q", "-b", "agent/" + name, "main")
        if with_card:
            self.put("backlog/tasks/task-1 - %s.md" % name, card(1, name, owns, extra))
        for f in files:
            self.put(f)
        self.g("add", "-A")
        self.g("commit", "-qm", "work")
        return ownscheck.check(self.d, "main", "HEAD", "agent/" + name)

    def test_inside_owns_and_generated(self):
        f, w = self.branch("a", ["internal/a/**", "docs/a.md"], ["internal/a/x.go", "docs/a.md", "CHANGELOG.md", "docs/plan/STATUS.md", "backlog/merge-lines/a.txt"])
        self.assertEqual((f, w), ([], []))

    def test_outside_owns_fails(self):
        f, _ = self.branch("a", ["internal/a/**"], ["internal/a/x.go", "internal/b/y.go"])
        self.assertEqual(len(f), 1)
        self.assertIn("internal/b/y.go", f[0])

    def test_warn_window(self):
        self.branch("a", ["internal/a/**"], ["internal/b/y.go"])
        f, w = ownscheck.check(self.d, "main", "HEAD", "agent/a", "2026-12-01", datetime.date(2026, 11, 25))
        self.assertEqual((len(f), len(w)), (0, 1))
        f, w = ownscheck.check(self.d, "main", "HEAD", "agent/a", "2026-12-01", datetime.date(2026, 12, 2))
        self.assertEqual((len(f), len(w)), (1, 0))

    def test_adr_needs_allow_even_when_owned(self):
        f, _ = self.branch("a", ["docs/adr/**", "docs/decisions.md"], ["docs/adr/0030-x.md", "docs/decisions.md", "GLOSSARY.md"])
        self.assertEqual(len([x for x in f if "guarded path" in x]), 3)

    def test_adr_allowed_with_marker(self):
        f, _ = self.branch("a", ["docs/adr/**", "docs/decisions.md"], ["docs/adr/0030-x.md", "docs/decisions.md"], extra="allow-adr: true\n")
        self.assertEqual(f, [])

    def test_no_card_warns_but_guards(self):
        f, w = self.branch("a", [], ["docs/decisions.md", "internal/z.go"], with_card=False)
        self.assertEqual(len(f), 1)
        self.assertEqual(len(w), 1)

    def test_contracts_list_is_allowed(self):
        f, _ = self.branch("a", ["internal/a/**"], ["internal/a/x.go", "docs/contracts/c.md"], extra="contracts:\n  - docs/contracts/c.md\n")
        self.assertEqual(f, [])

    def test_main_is_skipped(self):
        self.assertEqual(ownscheck.main(["--root", self.d, "--base", "main", "--branch", "main"]), 0)


if __name__ == "__main__":
    unittest.main()
