# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's skill select (C->S 0x3c) as the oracle; d2d's World select beside it.

Runs natively in game.exe 1.14d: the server handler FUN_0054be70 (size 9, the
dword's bit 31 = the hand, the id = the low 31 bits below the Skills.txt row
count at DataTables +0xba0), the unit's skill-list lookup FUN_006439b0 (unit
+0xa8 -> +4 first entry; an entry: +0 the Skills.txt record, +4 next, +0x28
base level, +0x34 the owner guid, -1 = the character's own) and the level
FUN_006442a0(unit, entry, 1) (base + FUN_00644180's bonus for an own entry,
at least 0, at most FUN_00611830(0)). The handler returns 3 (a bad packet,
nothing changes) or calls FUN_005701b0(unit, hand, skill, guid) and then
FUN_00622f10 (left) / FUN_00622ea0 (right), and returns 0.

Hooked (not the claim): FUN_00644180 (the item bonus of an own entry: the
case's `bonus`), FUN_00611830 (the level cap: the case's `cap`),
FUN_005701b0 (the apply: recorded; it stores DataTables-record pointers at
the skill list +0x08 left / +0x0c right, starts or stops a right-hand aura
and ignores skill 5: read, not run) and FUN_00622f10 / FUN_00622ea0 (the
client-side cache refresh: recorded). DataTables is a zeroed block with only
the row count and the records' id words.

    uv run python skill_bind.py gen 100000 7 > cases.txt
    uv run python skill_bind.py game cases.txt > g.txt
    uv run python skill_bind.py d2d before cases.txt | diff g.txt - | grep -c '^>'
    uv run python skill_bind.py d2d after cases.txt | diff g.txt -

A case is `i size id hand guid cap n (eid eguid ebase ebonus)*n`: the packet
size, the skill id and hand, the packet's owner guid, the level cap, then the
unit's entries. `game` prints `i set hand id guid` (set 0: refused, nothing
stored). `d2d before` is components/game/world/commands.cpp:352-357 at
upstream 72132b99 (every select stored); `d2d after` is the proposed rule:
stored when the id is a Skills.txt row and the character's level in it is
above 0, read for an own select from d2d's `skill_level` (points + gear) and,
for an item's, from the charged entry with that item. The mapping from d2d's
character to game.exe's entry list (an own entry exists when d2d's level
model gives the skill a level; a charged entry per item) is the bridging
assumption; d2d's cmd::SelectSkill has no item field, so the `item` cases
are a limit, not part of a d2d claim.
"""
import random
import sys

import emu
from unicorn.x86_const import UC_X86_REG_EDX

ROWS = 357                      # Skills.txt rows in the patched 1.14d stack (ids 0..356)
DT = 0x744304                   # the DataTables pointer
OWN = 0xFFFFFFFF


def gen(n_random, seed):
    rng = random.Random(seed)
    cases = []
    # exhaustive: every id 0..ROWS+3, both hands, own select; the entry none / own with base -1..3 and bonus 0..2
    for sid in range(ROWS + 4):
        for hand in (0, 1):
            cases.append((9, sid, hand, OWN, 20, []))
            for base in (-1, 0, 1, 2, 3):
                for bonus in (0, 1, 2):
                    cases.append((9, sid, hand, OWN, 20, [(sid, OWN, base, bonus)]))
            # an item's entry (guid 7): the right guid, a wrong guid, an own select of it
            for base in (0, 1, 5):
                cases.append((9, sid, hand, 7, 20, [(sid, 7, base, 0)]))
                cases.append((9, sid, hand, 8, 20, [(sid, 7, base, 0)]))
                cases.append((9, sid, hand, OWN, 20, [(sid, 7, base, 0)]))
    for size in (0, 8, 10, 13):                                   # the length check
        cases.append((size, 36, 0, OWN, 20, [(36, OWN, 1, 0)]))
    for cap in (0, 1, 20, 99):                                    # the cap's edge
        cases.append((9, 36, 0, OWN, cap, [(36, OWN, 1, 0)]))
    for _ in range(n_random):
        sid = rng.choice((rng.randrange(ROWS), rng.randrange(ROWS + 8), rng.randrange(1 << 31)))
        entries = []
        for _k in range(rng.randrange(0, 5)):
            eid = rng.choice((sid, sid, rng.randrange(ROWS)))
            eg = rng.choice((OWN, OWN, OWN, rng.randrange(1, 40)))
            entries.append((eid, eg, rng.randrange(-2, 6), rng.randrange(0, 4) if eg == OWN else 0))
        guid = rng.choice((OWN, OWN, OWN) + tuple(e[1] for e in entries if e[1] != OWN) + (rng.randrange(1, 40),))
        cases.append((rng.choice((9, 9, 9, 9, 8, 10)), sid, rng.randrange(2), guid, rng.choice((20, 20, 99, 1)), entries))
    for i, (size, sid, hand, guid, cap, entries) in enumerate(cases):
        print(i, size, sid, hand, guid, cap, len(entries), *[x for e in entries for x in e])


def parse(line):
    p = [int(x) for x in line.split()]
    i, size, sid, hand, guid, cap, n = p[:7]
    entries = [tuple(p[7 + 4 * k:11 + 4 * k]) for k in range(n)]
    return i, size, sid, hand, guid, cap, entries


class Oracle:
    def __init__(self):
        self.e = e = emu.Emu()
        self.cur = {}
        self.out = {}
        dt = e.alloc(0x1000)
        self.records = e.alloc(ROWS * 0x23C)
        e.w32(DT, dt)
        e.w32(dt + 0xBA0, ROWS)
        e.w32(dt + 0xB98, self.records)
        for sid in range(ROWS):
            e.mu.mem_write(self.records + sid * 0x23C, sid.to_bytes(2, "little"))
        self.dummy = e.alloc(0x23C)                 # a record for an id past the table (never matches a valid id)
        self.unit, self.skills = e.alloc(0x200), e.alloc(0x40)
        self.nodes = [e.alloc(0x40) for _ in range(8)]
        self.packet = e.alloc(16)
        e.w32(self.unit + 0xA8, self.skills)

        def bonus(_e):
            node = e.arg(0)
            return self.cur["bonus"].get(node, 0)
        e.hook(0x644180, bonus, 1)
        e.hook(0x611830, lambda _e: self.cur["cap"], 1)

        def apply(_e):
            self.out["set"] = (e.mu.reg_read(UC_X86_REG_EDX), e.arg(0), e.arg(1))
            return 0
        e.hook(0x5701B0, apply, 2)
        e.hook(0x622F10, lambda _e: self.out.__setitem__("cache", "L") or 0, 1)
        e.hook(0x622EA0, lambda _e: self.out.__setitem__("cache", "R") or 0, 1)

    def select(self, size, sid, hand, guid, cap, entries):
        e = self.e
        self.out.clear()
        self.cur = {"cap": cap, "bonus": {}}
        prev = 0
        for k, (eid, eg, base, bon) in enumerate(entries):
            node = self.nodes[k]
            e.mu.mem_write(node, b"\0" * 0x40)
            e.w32(node, self.records + eid * 0x23C if 0 <= eid < ROWS else self.dummy)
            e.w32(node + 0x28, base)
            e.w32(node + 0x34, eg)
            self.cur["bonus"][node] = bon
            if prev:
                e.w32(prev + 4, node)
            else:
                e.w32(self.skills + 4, node)
            prev = node
        if prev:
            e.w32(prev + 4, 0)
        else:
            e.w32(self.skills + 4, 0)
        word = (sid & 0x7FFFFFFF) | (0x80000000 if hand else 0)
        e.mu.mem_write(self.packet, bytes([0x3C]) + word.to_bytes(4, "little") + (guid & 0xFFFFFFFF).to_bytes(4, "little"))
        ret = e.call(0x54BE70, self.packet, size, ecx=0, edx=self.unit)
        s = self.out.get("set")
        if ret != 0 or s is None:
            return (0, hand, sid, guid)
        assert self.out.get("cache") == ("L" if s[0] else "R"), self.out
        return (1, s[0], s[1], s[2])


def d2d_before(size, sid, hand, guid, cap, entries):
    # commands.cpp:352-357 (upstream 72132b99): the hand takes any id (only the co-op Resurrect id 0x7ffe is gated;
    # it is past ROWS and never generated as such). The wire form has no size check against 9 bytes or guid.
    return (1, hand, sid, guid)


def level(entry, cap):
    _eid, eg, base, bon = entry
    v = base + (bon if eg == OWN else 0)
    return min(max(v, 0), cap)


def d2d_after(size, sid, hand, guid, cap, entries):
    # Proposed: stored only when 0 <= id < rows and the character's level in the skill (own: points + gear, the
    # entry for that id and owner) is above 0. size is game.exe's wire check (d2d's own codec has none).
    if size != 9 or not 0 <= sid < ROWS:
        return (0, hand, sid, guid)
    for e in entries:
        if e[0] == sid and e[1] == guid:
            return (1, hand, sid, guid) if level(e, cap) > 0 else (0, hand, sid, guid)
    return (0, hand, sid, guid)


def main():
    a = sys.argv[1:]
    if len(a) >= 3 and a[0] == "gen":
        gen(int(a[1]), int(a[2]))
    elif len(a) >= 2 and a[0] == "game":
        o = Oracle()
        with open(a[1]) as f:
            for line in f:
                if line.strip():
                    i, *c = parse(line)
                    print(i, *o.select(*c))
    elif len(a) >= 3 and a[0] == "d2d":
        fn = {"before": d2d_before, "after": d2d_after}[a[1]]
        with open(a[2]) as f:
            for line in f:
                if line.strip():
                    i, *c = parse(line)
                    print(i, *fn(*c))
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
