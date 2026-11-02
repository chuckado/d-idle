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


def main():
    mode = sys.argv[1]
    w = pa.World(5)
    w.real_callback()
    if mode == "friendly":
        friendly(w, int(sys.argv[2]) if len(sys.argv) > 2 else 98, int(sys.argv[3]) if len(sys.argv) > 3 else 1)


if __name__ == "__main__":
    main()
