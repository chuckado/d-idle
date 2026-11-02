#!/usr/bin/env python3
"""Does a carried shield's damage reach the player's stats? game.exe 1.14d's own equip code under unicorn (card skshield).

Runs next to upstream's harness (tools/emu: emu.py, drlg.py). The data tables are game.exe's own loader's
(drlg.boot: FUN_00619300), the item rows are real (Items.txt rows by code); the unit, its inventory and the
item units are minimal structures built here. Game code that runs: FUN_004c0d20 (the equip attach),
FUN_00627910 -> FUN_00626e10 -> FUN_00626920 / FUN_006266c0 (attach and stat merge), FUN_00627260 (set a
stat), FUN_00535d10 (the hand toggle the hit code runs round a melee hit), FUN_0063bef0 (the hand weapon),
FUN_00629bb0 (item type test), FUN_006235a0 (dual wield class), FUN_00625480 (read a stat), FUN_006280a0
(item flag test). Hooked: FUN_00663cc0 (inventory bookkeeping), FUN_0040b430 / FUN_0040b4b0 (pool allocator),
FUN_00535bc0 (the weapon the skill uses; the scenario's `skill`).

    python skshield_run.py            # prints one line per scenario and a summary
Scenarios are an enumeration, not random: classes x hands x shield stat sets x item states x toggle flag x skill
hand (no seed). The "model" is this card's reading of the code (docs/research/skshield-trace.md Q2), applied per
scenario; a DIFF is a place where the reading and the run disagree.
"""
import itertools
import sys

from unicorn.x86_const import UC_X86_REG_EDX

import drlg

PALADIN, BARB, ASSASSIN, SORC = 3, 4, 6, 1
DAMAGE_STATS = (17, 18, 21, 22, 93, 25)        # damagerelated = 1 in ItemStatCost (93 = IAS, 25 = damagepercent)
OTHER_STATS = (31, 80)                          # armorclass, item_magicbonus: not damage related
ALL_STATS = DAMAGE_STATS + OTHER_STATS
OWNER = 0x1234
W_STATS = ((21, 3), (22, 7))
W2_STATS = ((21, 1), (22, 2))
SHIELDS = {
    "none": (), "flat": ((21, 2), (22, 5)), "ed": ((17, 40), (18, 40)), "flat+ed": ((21, 2), (22, 5), (17, 40), (18, 40)),
    "ias": ((93, 20),), "plain": ((31, 10), (80, 15)), "dmg%": ((25, 30),),
}


class World:
    def __init__(self):
        self.e = e = drlg.boot()
        self.codes = []
        while rec := e.call(0x6335f0, len(self.codes)):
            self.codes.append(e.read(rec + 0x80, 4).decode("latin-1").rstrip(" \0"))
        self.skill = 0
        # pool allocator: FUN_0040b430 (fastcall ecx pool, edx size; 3 stack args), FUN_0040b4b0 (realloc; ecx pool, edx old, 4 stack args)
        e.hook(0x40b430, lambda e: e.alloc(e.mu.reg_read(UC_X86_REG_EDX) + 16), 3)

        def realloc(e):
            new = e.alloc(e.arg(0) + 16)
            old = e.mu.reg_read(UC_X86_REG_EDX)
            if old: e.mu.mem_write(new, e.read(old, min(e.arg(0), 0x400)))
            return new
        e.hook(0x40b4b0, realloc, 4)
        e.hook(0x663cc0, lambda e: 1, 4)
        e.hook(0x535bc0, lambda e: self.skill, 0)

    def lst(self, flags):
        return self.e.call(0x6251f0, 0, flags, 0, 0, 0)

    def unit(self, cls):
        e = self.e
        u = e.alloc(0x200)
        e.w32(u, 0); e.w32(u + 4, cls)
        e.w32(u + 0x5c, self.lst(0x80000000))
        e.w32(e.r32(u + 0x5c) + 0x44, u)
        inv = e.alloc(0x40)
        e.w32(inv, 0x1020304); e.w32(inv + 0x1c, OWNER)
        grid = e.alloc(0x20)
        e.w32(inv + 0x14, grid); e.w32(inv + 0x18, 1)
        e.mu.mem_write(grid + 8, e.read(0x74479c, 2))
        slots = e.alloc(0x40)
        e.w32(grid + 0xc, slots)
        e.w32(u + 0x60, inv)
        return u, slots

    def item(self, code, flags, stats, bodyloc):
        e = self.e
        it = e.alloc(0x200)
        e.w32(it, 4); e.w32(it + 4, self.codes.index(code)); e.w32(it + 0xc, OWNER)
        data = e.alloc(0x100)
        e.w32(data + 0x18, flags)
        e.mu.mem_write(data + 0x44, bytes([bodyloc]))
        e.w32(it + 0x14, data)
        e.w32(it + 0x5c, self.lst(0))
        for sid, val in stats:
            e.call(0x627260, it, sid, val, 0)
        return it

    def stat(self, u, sid):
        v = self.e.call(0x625480, u, sid, 0)
        return v - (1 << 32) if v >= 1 << 31 else v


def model(cls, hand, sname, state, flag, skill, wstats):
    """This card's reading of FUN_004c0d20 + FUN_00535d10, as code."""
    dual = cls in (BARB, ASSASSIN)
    stats = {"W": wstats, "W2": W2_STATS, "S": SHIELDS[sname]}
    held = [k for k in ("W", "W2", "S") if k in hand.split("+")]
    prim = "W2" if "W2" in held else "W" if "W" in held else None      # FUN_0063bef0: the left slot's weapon first
    other = "W" if prim == "W2" else "W2" if "W2" in held else None    # FUN_0063bf90 (a weapon only)
    active = {k: True for k in held}
    if dual:
        for k in ("W", "W2"):
            if k in held and k != prim: active[k] = False             # FUN_004c0d20: not the primary weapon -> attached off
    attached = [k for k in held if not (k == "S" and state != "ok")]
    if flag:                                                           # FUN_00535d10 block 1: both hand weapons off
        for k in (prim, other):
            if k: active[k] = False
    if dual:                                                           # block 2
        sk = {"W": "W", "W2": "W2", "none": None}[skill]
        sk = sk if sk in held else None
        if sk is None:
            if prim: active[prim] = True
        elif sk == prim:
            active[sk] = True
            if other: active[other] = False
        else:
            active[sk] = True
            if prim: active[prim] = False
    want = {sid: 0 for sid in ALL_STATS}
    for k in attached:
        for sid, val in stats[k]:
            if sid in OTHER_STATS or k == "S" or active[k]:
                want[sid] += val
    return want


def main():
    w = World(); e = w.e
    n = bad = 0
    rows = []
    cases = []
    for cls, sname, state in itertools.product((PALADIN, BARB, ASSASSIN, SORC), SHIELDS, ("ok", "broken", "noequip")):
        if sname == "none" and state != "ok": continue
        for flag in (0, 1):
            cases.append((cls, "S", sname, state, flag, "W"))
            cases.append((cls, "W+S", sname, state, flag, "W"))
    for cls in (PALADIN, BARB, ASSASSIN, SORC):
        cases.append((cls, "empty", "none", "ok", 0, "none"))
        for flag in (0, 1):
            cases.append((cls, "W", "none", "ok", flag, "W"))
            cases.append((cls, "W", "none", "ok", flag, "none"))
    for cls, flag, skill in itertools.product((BARB, ASSASSIN, PALADIN), (0, 1), ("W", "W2", "none")):
        cases.append((cls, "W+W2", "none", "ok", flag, skill))
    for cls, hand, sname, state, flag, skill in cases:
        flagbits = {"ok": 0, "broken": 0x100, "noequip": 0x4000}[state]
        u, slots = w.unit(cls)
        items = {}
        if "W2" in hand.split("+"): items["W2"] = w.item("ssd", 0, W2_STATS, 5)
        if "W" in hand.split("+"): items["W"] = w.item("ssd", 0, W_STATS, 4)
        if "S" in hand.split("+"): items["S"] = w.item("spk", flagbits, SHIELDS[sname], 5)
        if "W" in items: e.w32(slots + 4 * 4, items["W"])
        if "W2" in items: e.w32(slots + 4 * 5, items["W2"])
        if "S" in items: e.w32(slots + 4 * 5, items["S"])
        for k in ("W", "W2", "S"):
            if k in items: e.call(0x4c0d20, 0, regs={"edi": u, "esi": items[k]})
        w.skill = items.get(skill, 0)
        e.call(0x535d10, ecx=u, edx=flag)
        got = {sid: w.stat(u, sid) for sid in ALL_STATS}
        want = model(cls, hand, sname, state, flag, skill, W_STATS)
        n += 1
        diff = {sid: (got[sid], want[sid]) for sid in got if got[sid] != want[sid]}
        bad += bool(diff)
        rows.append(f"cls {cls} hand {hand:5} shield {sname:8} {state:7} flag {flag} skill {skill:4}: " + " ".join(f"{sid}={got[sid]}" for sid in got if got[sid] or want[sid]) + (f"  DIFF(got,model) {diff}" if diff else ""))
    # observation only (no model): how the weapon's own ED and a shield's ED sit in the player's list (the op 13 question)
    obs = []
    for hand, wst, sst in (("W", W_STATS + ((17, 50), (18, 50)), ()), ("W+S", W_STATS, ((17, 40), (18, 40))), ("W+S", W_STATS + ((17, 50), (18, 50)), ((17, 40), (18, 40))), ("W+S", W_STATS, ((21, 2), (22, 5)))):
        u, slots = w.unit(PALADIN)
        wi = w.item("ssd", 0, wst, 4); e.w32(slots + 16, wi); e.call(0x4c0d20, 0, regs={"edi": u, "esi": wi})
        if "S" in hand:
            si = w.item("spk", 0, sst, 5); e.w32(slots + 20, si); e.call(0x4c0d20, 0, regs={"edi": u, "esi": si})
        w.skill = 0
        e.call(0x535d10, ecx=u, edx=0)
        obs.append(f"OBS paladin {hand}: weapon stats {wst}, shield stats {sst}: player 17/18/21/22 = " + "/".join(str(w.stat(u, s)) for s in (17, 18, 21, 22)))
        wl = e.r32(wi + 0x5c)
        obs.append("    weapon's own list: " + "/".join(str(e.call(0x625480, wi, s, 0)) for s in (17, 18, 21, 22)))
    print("\n".join(rows))
    print("\n".join(obs))
    print(f"scenarios {n}, differ from the model {bad}")


if __name__ == "__main__":
    sys.exit(main())
