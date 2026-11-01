"""The draft server delta (docs/community/discord/server-work.json) planned offline against a server that already matches server.json."""
import json
import os
import sys
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import apply_server as ap

DOCS = os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs", "community", "discord")
GUILD = "5550001"


class FakeApi:
    """Only GETs are answered; any write is a test failure (a dry run must not write)."""

    def __init__(self, spec):
        self.roles = [{"id": GUILD, "name": "@everyone"}] + [{"id": str(10 + i), "name": r["name"]} for i, r in enumerate(spec["roles"])]
        self.channels, n = [], 100
        for cat in spec["categories"]:
            n += 1
            self.channels.append({"id": str(n), "type": ap.CATEGORY, "name": cat["name"]})
            for c in cat["channels"]:
                n += 1
                self.channels.append({"id": str(n), "type": ap.KINDS[c.get("type", "text")], "name": c["name"]})

    def call(self, method, path, body=None):
        assert method == "GET", "dry run wrote: %s %s" % (method, path)
        return self.roles if path.endswith("/roles") else self.channels


class WorkDelta(unittest.TestCase):
    def setUp(self):
        with open(os.path.join(DOCS, "server.json")) as f:
            self.base = json.load(f)
        with open(os.path.join(DOCS, "server-work.json")) as f:
            self.delta = json.load(f)

    def test_dry_run_creates_exactly_the_new_channels(self):
        lines = []
        actions = ap.plan_and_apply(FakeApi(self.base), GUILD, self.delta, apply=False, out=lines.append)
        self.assertEqual(actions, ["create category WORK", "create channel #work-board in WORK", "create channel #agents in WORK"])

    def test_delta_channels_not_already_in_base(self):
        names = {c["name"] for cat in self.base["categories"] for c in cat["channels"]}
        self.assertFalse(names & {"work-board", "agents"})

    def test_permissions_are_known_and_agents_has_no_attach(self):
        for cat in self.delta["categories"]:
            for c in cat["channels"]:
                for o in c["overwrites"]:
                    ap.perm_bits(o.get("allow"))
                    ap.perm_bits(o.get("deny"))
                    self.assertNotIn("attach_files", o.get("allow", []))
                    self.assertNotIn("mention_everyone", o.get("allow", []))

    def test_work_board_read_only_for_everyone(self):
        wb = self.delta["categories"][0]["channels"][0]
        self.assertIn("send_messages", wb["overwrites"][0]["deny"])
        self.assertEqual(wb["overwrites"][0]["role"], "@everyone")

    def test_agents_hidden_from_everyone(self):
        ag = self.delta["categories"][0]["channels"][1]
        self.assertIn("view_channel", ag["overwrites"][0]["deny"])


if __name__ == "__main__":
    unittest.main()
