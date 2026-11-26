import sys
import tempfile
import unittest
from datetime import date
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import plan  # noqa: E402

TODAY = date(2026, 11, 25)


def card(n, name, status="Proposed", extra="", prio="medium"):
    return (f"---\nid: TASK-{n}\ntitle: \"{name}: x\"\nstatus: {status}\nassignee:\n  - credit-runner\nlabels:\n  - bot\n  - feature\n"
            f"priority: {prio}\ncard: \"{name}\"\nowns:\n  - zz/{name}\n{extra}---\n")


class T(unittest.TestCase):
    def sec(self, cards, **kw):
        with tempfile.TemporaryDirectory() as d:
            t = Path(d) / "backlog" / "tasks"
            t.mkdir(parents=True)
            for i, c in enumerate(cards, 1):
                (t / f"task-{i} - x.md").write_text(c)
            return plan.build(d, TODAY, **kw)

    def ids(self, sec, k):
        return [x["card"] for x in sec[k]]

    def test_default_lanes_and_done_hidden(self):
        s = self.sec([card(1, "ip", "In Progress"), card(2, "p"), card(3, "d", "Deferred"), card(4, "x", "Done")])
        self.assertEqual(self.ids(s, "now"), ["ip"])
        self.assertEqual(self.ids(s, "next"), ["p"])
        self.assertEqual(self.ids(s, "parked"), ["d"])

    def test_dependency_blocks_until_done(self):
        s = self.sec([card(1, "a"), card(2, "b", extra="dependencies:\n  - TASK-1\n")])
        self.assertEqual(self.ids(s, "blocked"), ["b"])
        self.assertEqual(s["blocked"][0]["open_deps"], ["TASK-1"])
        s = self.sec([card(1, "a", "Done"), card(2, "b", extra="dependencies:\n  - TASK-1\n")])
        self.assertEqual(self.ids(s, "next"), ["b"])

    def test_order_priority_then_due(self):
        s = self.sec([card(1, "lo", prio="low"), card(2, "hi", prio="high", extra="due: 2026-12-30\n"), card(3, "hi2", prio="high", extra="due: 2026-12-01\n")])
        self.assertEqual(self.ids(s, "next"), ["hi2", "hi", "lo"])

    def test_overdue_and_owner_lane_and_epic_filter(self):
        s = self.sec([card(1, "late", extra="due: 2026-11-01\n"), card(2, "w", extra="lane: waiting-owner\nlane_note: ask\n"), card(3, "kid", extra="epic: epic-x\n")])
        self.assertEqual(self.ids(s, "overdue"), ["late"])
        self.assertEqual(self.ids(s, "waiting-owner"), ["w"])
        s = self.sec([card(3, "kid", extra="epic: epic-x\n"), card(4, "other")], epic="epic-x")
        self.assertEqual(self.ids(s, "next"), ["kid"])


if __name__ == "__main__":
    unittest.main()
