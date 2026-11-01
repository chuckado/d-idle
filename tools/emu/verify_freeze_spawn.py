# SPDX-License-Identifier: GPL-3.0-or-later
"""Verifier (card verifyfreeze): which monster-data flag bits game.exe's own spawn path gives
each kind of monster, and what game.exe's own FUN_0057b230 then does to it.

    PYTHONPATH=<d2decomped>/tools/emu python verify_freeze_spawn.py [SEEDS] [SEED0] > spawn.txt

Runs upstream's monsters.py population (FUN_0052d0f0 over every room of a level, the real
spawn path: FUN_00555230, the boss / champion / minion placement FUN_005a43e0, FUN_005a48c0,
FUN_005a4850, FUN_005a0320, the act bosses FUN_005b1cf0) for SEEDS map seeds from SEED0 on
levels with champions, uniques and superuniques in three difficulties. Afterwards, for every
monster unit FUN_00555230 returned: its monster-data word +0x16 (read from the unit), its MonUMod
list (+0x1c), whether it got FUN_00639db0(unit, 0x76, ...) (state 118), and the outcome of the
REAL FUN_0057b230(eax = unit; attacker, 100) with its MonStats coldeffect set to -50 for the call
(so that a plain monster can be frozen at all): `freeze` (state 1 set), `chill` (state 0xb set)
or `none`. Hooked: FUN_00554010 (the attacker's game) and the timer events FUN_005417d0 /
FUN_00540e60 (not the claim); the state list calls run for real on the real unit.

Output: one `unit` line per monster, then a `table` of (flags word, kind guess, outcome) counts.
"""
import collections
import sys

import capstone
from unicorn import UC_HOOK_CODE
from unicorn.x86_const import UC_X86_REG_EAX

import drlg
import monsters

LEVELS = (3, 4, 8, 17, 25, 37, 41, 73, 79, 102, 108)   # Cold Plains, Stony Field, Den, Burial Grounds, Tower Cellar 5, Catacombs 4, Rocky Waste, Duriel, Spider Forest, Durance 3, Chaos Sanctuary


def rets(e, start, n=0x600):
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
    out = []
    for ins in md.disasm(bytes(e.read(start, n)), start):
        if ins.mnemonic == "ret": out.append(ins.address)
        if ins.mnemonic == "int3" and out: break
    return out


def main():
    nseeds = int(sys.argv[1]) if len(sys.argv) > 1 else 4
    seed0 = (int(sys.argv[2], 0) if len(sys.argv) > 2 else 1) & 0xFFFFFFFF
    e = drlg.boot()
    made, state118 = [], set()

    def on_ret(mu, addr, size, _):
        unit = mu.reg_read(UC_X86_REG_EAX)
        if unit and e.r32(unit) == 1: made.append(unit)

    def on_state(mu, addr, size, _):
        if e.arg(1) == 0x76: state118.add(e.arg(0))

    hooks = [e.mu.hook_add(UC_HOOK_CODE, on_ret, begin=a, end=a) for a in rets(e, 0x555230)]
    hooks.append(e.mu.hook_add(UC_HOOK_CODE, on_state, begin=0x639db0, end=0x639db0))
    table = collections.Counter()
    attacker = e.alloc(0x200)
    e.w32(attacker, 0)
    for difficulty in range(3):
        for k in range(nseeds):
            seed = (seed0 + 7919 * k) & 0xFFFFFFFF
            for lid in LEVELS:
                del made[:]
                state118.clear()
                try:
                    monsters.dump(e, seed, lid, difficulty)
                except SystemExit:
                    continue
                g = monsters_game[0] if monsters_game else None
                for unit in list(dict.fromkeys(made)):
                    data = e.r32(unit + 0x14)
                    if not data: continue
                    flags = e.read(data + 0x16, 2)
                    flags = flags[0] | flags[1] << 8
                    mods = bytes(e.read(data + 0x1c, 9)).rstrip(b"\0")
                    cls = e.r32(unit + 4)
                    outcome = freeze_outcome(e, unit, attacker, cls, difficulty, g)
                    print(f"unit d{difficulty} s{seed} l{lid} class {cls} flags {flags:#06x} mods {'.'.join(map(str, mods)) or '-'}"
                          f"{' state118' if unit in state118 else ''} -> {outcome}")
                    table[(flags & 0xff, outcome, unit in state118)] += 1
    for h in hooks: e.mu.hook_del(h)
    for (flags, outcome, s118), n in sorted(table.items()):
        print(f"table flags {flags:#04x} bit8 {1 if flags & 8 else 0} state118 {int(s118)} -> {outcome}: {n}")


monsters_game = []
_new_game = monsters.new_game


def _capture(e, seed, difficulty=0):
    g = _new_game(e, seed, difficulty)
    monsters_game[:] = [g]
    return g


monsters.new_game = _capture


patched, current_game = {}, []


def freeze_outcome(e, unit, attacker, cls, difficulty, g):
    tables = e.r32(0x744304)
    row = e.r32(tables + 0xa78) + cls * 0x1a8
    saved = bytes(e.read(row + 0x168, 3))
    e.mu.mem_write(row + 0x168, bytes([0xce, 0xce, 0xce]))     # coldeffect -50 in every difficulty
    current_game[:] = [g]
    if not patched:                                           # one stub slot each, swapped in and out per call
        for addr, fn, n in ((0x554010, lambda e: current_game[0], 0), (0x5417d0, lambda e: 0, 4), (0x540e60, lambda e: 0, 2)):
            code = e.read(addr, 5)
            e.hook(addr, fn, n)
            patched[addr] = (code, e.read(addr, 5))
    hooks = [(addr, code) for addr, (code, _) in patched.items()]
    for addr, (_, jump) in patched.items():
        e.mu.mem_write(addr, jump)
        e.mu.ctl_remove_cache(addr, addr + 5)
    before = (e.call(0x639df0, unit, 1), e.call(0x639df0, unit, 0xb))
    try:
        e.call(0x57b230, attacker, 100, regs={"eax": unit})
        after = (e.call(0x639df0, unit, 1), e.call(0x639df0, unit, 0xb))
    finally:
        e.mu.mem_write(row + 0x168, saved)
        for addr, code in hooks:                               # undo the hooks: the next population needs them real
            e.mu.mem_write(addr, code)
            e.mu.ctl_remove_cache(addr, addr + 5)
    if before != (0, 0): return "prior-state"
    if after[0] and not after[1]: return "freeze"
    if after[1] and not after[0]: return "chill"
    return "none" if after == (0, 0) else "both"


if __name__ == "__main__": main()
