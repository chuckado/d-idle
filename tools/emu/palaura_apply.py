# SPDX-License-Identifier: GPL-3.0-or-later
"""palaura: run the friendly aura do (FUN_005cf010) and the Holy Fire / Shock do (FUN_005cf3a0) natively.

    tools/emu/run-on-worker.sh palaura_apply.py friendly [skill] [level]   # who gets which state and stats
    tools/emu/run-on-worker.sh palaura_apply.py hit [skill] [level] [seed] # enemy damage per pulse vs the element tables

World, fakes and what is hooked: palaura_area.py (real Blood Moor game, real player / monster units; owner block and
party list are hand-made; FUN_00554200 restored). Hooked for the calcs only, as skill_damage.py does: FUN_00643810
(a fake skill record per id) and FUN_006442a0 (its level: the case's); the caster's character level is the real stat
0xc and its mana the real stat 8. `friendly` calls the do as the pulse event does (game, caster, skill, level), then
reads each unit's states (FUN_00639df0) and the stats the aura sets. `hit` runs FUN_005cf3a0 on a caster with
Quill Rats in range, reads the rats' life (stat 6, 256ths) before and after, many pulses on fresh rats."""
import sys

from unicorn.x86_const import UC_X86_REG_ESI

import palaura_area as pa

FAKE = 0x30000000


def prep(w, levels):
    e = w.e
    e.hook(0x643810, lambda e: FAKE + 0x40 * (e.mu.reg_read(UC_X86_REG_ESI) & 0xffff), 0)
    e.hook(0x6442a0, lambda e: levels.get((e.arg(1) - FAKE) // 0x40, 0), 3)


def states(e, u):
    return [s for s in range(0, 0xc4) if e.call(0x639df0, u, s) != 0]


def friendly(w, skill, level):
    e = w.e
    levels = {skill: level}
    prep(w, levels)
    cx, cy = w.free_spot()
    caster, _, _ = w.make(pa.PLAYER, 4, cx, cy)
    # FUN_00639db0 is the state setter (unit, state, on); a stat is set on a stat list: FUN_00627150(list, stat, value, layer)
    mana = int(sys.argv[5]) if len(sys.argv) > 5 else 100
    for stat, val in ((0xc, 30), (9, 100 << 8), (8, mana << 8)):
        e.call(0x627150, e.r32(caster + 0x5c), stat, val, 0)
    print("caster stats 6 7 8 9 0xc after the setup:", [e.call(0x625480, caster, s, 0) for s in (6, 7, 8, 9, 0xc)])
    ally, _, _ = w.make(pa.PLAYER, 1, cx + 3, cy)
    other, _, _ = w.make(pa.PLAYER, 2, cx - 3, cy)
    w.party([[e.r32(caster + 0xc), e.r32(ally + 0xc)]])
    rows = [("caster", caster), ("party ally player", ally), ("other player", other)]
    merc, _, _ = w.make(pa.MONSTER, 271, cx, cy + 4)
    w.own(merc, caster)
    pet, _, _ = w.make(pa.MONSTER, 63, cx, cy - 4)
    w.own(pet, caster)
    far, _, _ = w.make(pa.MONSTER, 63, cx, cy + 40)
    w.own(far, caster)
    foe, _, _ = w.make(pa.MONSTER, 63, cx + 5, cy)
    rows += [("merc", merc), ("pet", pet), ("pet 40 subtiles away", far), ("enemy", foe)]
    # who is hurt: stat 6 (life, 256ths) set to half of stat 7 (max life); players have no life yet, so give them 80 / 40 life
    hurt = set(sys.argv[4].split(",")) if len(sys.argv) > 4 else set()
    for l, u in rows:
        if l in ("caster", "party ally player", "other player"):
            e.call(0x627150, e.r32(u + 0x5c), 7, 80 << 8, 0)
            e.call(0x627150, e.r32(u + 0x5c), 6, (40 if l in hurt else 80) << 8, 0)
        elif l in hurt:
            e.call(0x627150, e.r32(u + 0x5c), 6, e.call(0x625480, u, 7, 0) // 2, 0)
    life = lambda u: e.call(0x625480, u, 6, 0) / 256
    row = e.call(0x0, ecx=0) if False else None
    rec = e.r32(e.r32(0x744304) + 0xb98) + skill * 0x23c
    print(f"skill {skill} level {level}: aurastate {e.r16(rec + 0x80)} auratargetstate {e.r16(rec + 0x82)} aurafilter {e.r32(rec + 0x50):#x}"
          f" stats {[e.r16(rec + 0x54 + 2 * k) for k in range(6)]}")
    before = {l: (states(e, u), e.call(0x625480, u, 25, 0), life(u)) for l, u in rows}
    if len(sys.argv) > 6 and sys.argv[6] == "state85":
        e.call(0x639db0, caster, 0x55, 1)
    print("before pulse: mana", e.call(0x625480, caster, 8, 0) / 256, "state 85", e.call(0x639df0, caster, 0x55) != 0)
    from unicorn import UC_HOOK_CODE
    from unicorn.x86_const import UC_X86_REG_EAX, UC_X86_REG_EDX, UC_X86_REG_ECX

    def reg(r): return e.mu.reg_read(r)
    trace = []
    def at(addr, what):
        e.mu.hook_add(UC_HOOK_CODE, lambda mu, a, s, _: trace.append((what, reg(UC_X86_REG_EAX), reg(UC_X86_REG_EDX), reg(UC_X86_REG_ECX))),
                      begin=addr, end=addr)
    at(0x5cf090, "56cd50 returned (length base)")
    at(0x5cee3d, "56e970 returned (stat list)")
    at(0x5cedc0, "callback entry")
    at(0x5cee38, "call 56e970")
    at(0x56e970, "56e970 entry")
    at(0x5cee44, "callback returns early (no stat list)")
    at(0x5cee4a, "callback has the stat list")
    at(0x5cef8d, "stat loop step")
    at(0x5cf143, "before 56b740 (self apply)")
    from unicorn.x86_const import UC_X86_REG_EBP
    def local(what, off):
        e.mu.hook_add(UC_HOOK_CODE, lambda mu, a, s, _: trace.append((what, e.s32(reg(UC_X86_REG_EBP) + off), 0, 0)), begin=0x5cf229 if off == -0x10 else 0x5cf239, end=0x5cf229 if off == -0x10 else 0x5cf239)
    local("cost (256ths) at the payment", -0x10)
    local("stats applied counter at the payment", -0x24)
    print("do result", e.call(0x5cf010, skill, level, ecx=w.game.g, edx=caster, limit=20000000))
    for t in trace: print("   trace", t[0], [hex(v) for v in t[1:]])
    for l, u in rows:
        st = states(e, u)
        print(f"   {l:22s} states {st} stat25 {e.call(0x625480, u, 25, 0)} (was {before[l][1]}) life {before[l][2]} -> {life(u)}")
    print("after pulse: mana", e.call(0x625480, caster, 8, 0) / 256, "state 85", e.call(0x639df0, caster, 0x55) != 0)


def regen(w):
    """FUN_005806f0 (EDI = unit; the per-frame mana regeneration, SKMANA's) with state 85 off / on, a flat stat 26 and a stat 27 %."""
    e = w.e
    cx, cy = w.free_spot()
    caster, _, _ = w.make(pa.PLAYER, 4, cx, cy)
    lst = e.r32(caster + 0x5c)
    for stat, val in ((9, 100 << 8),):
        e.call(0x627150, lst, stat, val, 0)
    print("max mana 100 points = 25600; per-frame base = max(25600 / 3000, 1) =", 25600 // 3000, "(256ths)")
    for flat, pct, st85 in ((0, 0, 0), (0, 0, 1), (512, 0, 0), (512, 0, 1), (0, 100, 0), (0, 100, 1), (512, 100, 0), (512, 100, 1)):
        e.call(0x627150, lst, 8, 0, 0)
        e.call(0x627150, lst, 26, flat, 0)
        e.call(0x627150, lst, 27, pct, 0)
        e.call(0x639db0, caster, 0x55, st85)
        e.call(0x5806f0, regs={"edi": caster})
        print(f"state 85 {'on ' if st85 else 'off'} stat 26 {flat:4d} stat 27 {pct:3d}: mana after one frame {e.call(0x625480, caster, 8, 0)} (256ths)")


def hit(w, skill, level, pulses, do=0x5cf3a0):
    """enemy damage per pulse: FUN_005cf3a0 (do 66) or FUN_005d0920 (do 81) on Quill Rats in range, rat life before / after."""
    e = w.e
    levels = {skill: level}
    prep(w, levels)
    cx, cy = w.free_spot()
    caster, _, _ = w.make(pa.PLAYER, 4, cx, cy)
    for stat, val in ((0xc, 30), (9, 100 << 8), (8, 100 << 8), (7, 100 << 8), (6, 100 << 8)):
        e.call(0x627150, e.r32(caster + 0x5c), stat, val, 0)
    lo = e.call(0x644d50, caster, skill, level, 1)
    hi = e.call(0x644e40, caster, skill, level, 1)
    rec = e.r32(e.r32(0x744304) + 0xb98) + skill * 0x23c
    rng = e.call(0x646ca0, caster, e.r32(rec + 0x64), skill, level)
    print(f"skill {skill} level {level}: element min {lo} max {hi} (256ths, FUN_00644d50 / 00644e40 flag 1), range {rng}, aurafilter {e.r32(rec + 0x50):#x},"
          f" etype {e.read(rec + 0x1dc, 1)[0]}, auratargetstate {e.r16(rec + 0x82)}")
    # the unit's own seed (FUN_0045c3e0 rolls it): a fresh unit's is 0 and then the roll is stuck at the minimum
    e.w32(caster + 0x20, w.rng.getrandbits(32) | 1)
    e.w32(caster + 0x24, 666)
    import monstats
    rows = [r for r in monstats.table("monstats") if r.get("id") and r["id"] != "Expansion"]
    col = {1: "resfi", 2: "resli", 3: "resma", 4: "resco", 5: "respo"}[e.read(rec + 0x1dc, 1)[0]]
    # one Quill Rat (63, resist 0) and rows with other values of the element's resist column (MonStats, normal)
    picks = [63]
    for want in (50, 75, 100, -20):
        for i, r in enumerate(rows):
            if r.get(col, "") == str(want) and i < 700 and r.get("npc") != "1" and r.get("enabled") == "1" and i not in picks:
                picks.append(i)
                break
    resist = {63: 0}
    resist.update({i: int(rows[i][col]) for i in picks[1:]})
    foes = []
    for k, cls in enumerate(picks):
        u, _, _ = w.make(pa.MONSTER, cls, cx + 2, cy - 2 + k)
        foes.append(u)
    print(f"element column {col}; targets (class: resist %):", {i: resist[i] for i in picks})
    far, _, _ = w.make(pa.MONSTER, 63, cx + rng + 3, cy)
    ally_pet, _, _ = w.make(pa.MONSTER, 63, cx, cy + 3)
    w.own(ally_pet, caster)
    npc, _, _ = w.make(pa.MONSTER, 148, cx, cy - 3)
    allv = foes + [far, ally_pet, npc]
    for u in allv:
        e.call(0x627150, e.r32(u + 0x5c), 7, 100000 << 8, 0)
        e.call(0x627150, e.r32(u + 0x5c), 6, 100000 << 8, 0)
    drops = []
    for p in range(pulses):
        before = [e.call(0x625480, u, 6, 0) for u in allv]
        e.call(do, skill, level, ecx=w.game.g, edx=caster, limit=50000000)
        after = [e.call(0x625480, u, 6, 0) for u in allv]
        d = [b - a for a, b in zip(after, before)]
        drops.append(d)
        for u, b in zip(allv, after): e.call(0x627150, e.r32(u + 0x5c), 6, 100000 << 8, 0)
    names = [f"class {c} (resist {resist[c]})" for c in picks] + ["rat out of range", "caster's pet", "NPC (Akara)"]
    for i, n in enumerate(names):
        vals = [d[i] for d in drops]
        print(f"   {n:28s} damage per pulse (256ths) min {min(vals)} max {max(vals)} distinct {len(set(vals))} first {vals[:6]}")
    rolls = [d[0] for d in drops]
    print(f"{len(rolls)} pulses: roll (resist 0 rat) within [{lo}, {hi}]:", sum(lo <= v <= hi for v in rolls), "of", len(rolls),
          "; distinct rolls", len(set(rolls)), "; mean", sum(rolls) / len(rolls), "(table mean", (lo + hi) / 2, ")")
    for j, c in enumerate(picks[1:], 1):
        r = resist[c]
        same = sum(d[j] == d[0] * (100 - r) // 100 for d in drops)
        same_md = sum(d[j] == (d[0] * (100 - r) + 50) // 100 for d in drops)
        print(f"   class {c} resist {r}: damage == roll * (100 - r) // 100 in {same} of {len(drops)} pulses (rounded: {same_md})")
    print("targets outside the circle, owned, or NPC took damage in", sum(1 for d in drops if any(d[len(picks):])), "pulses")


def main():
    mode = sys.argv[1]
    w = pa.World(5)
    w.real_callback()
    if mode == "regen":
        return regen(w)
    if mode == "hit":
        s = int(sys.argv[2])
        return hit(w, s, int(sys.argv[3]), int(sys.argv[4]) if len(sys.argv) > 4 else 20, 0x5d0920 if s == 114 else 0x5cf3a0)
    if mode == "friendly":
        friendly(w, int(sys.argv[2]) if len(sys.argv) > 2 else 98, int(sys.argv[3]) if len(sys.argv) > 3 else 1)


if __name__ == "__main__":
    main()
