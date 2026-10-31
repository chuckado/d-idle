#!/usr/bin/env python3
"""game.exe 1.14d's monster-missile hit on the player as the oracle (d2lod-bot, card decompmon).

Runs game.exe's own code under unicorn through upstream's harness
(psi29a/d2decomped tools/emu/emu.py): copy this file into that directory
and run it from there:

    cd <upstream>/tools/emu
    D2_MPQ_DIR=<mpq dir> <venv>/bin/python mon_missile.py --exe <game.exe> [--tables DIR] [cases]

The missile's unit hit (FUN_005adf10, read): the to-hit roll FUN_0057d9b0
(owner, target, bonus = the missile's stat 19, missile flag 1) runs only when
the Missiles.txt row's ToHit (+0x18b) is set; a miss ends the missile. Then
FUN_005ad730 rolls the block / dodge with FUN_0057dfb0(game, owner; target,
missile 1, can-block = record +8 != 0, i.e. the missile carries physical
damage). Melee (FUN_0057ec10) passes can-block 1, missile 0.

- `block`: FUN_0057dfb0 run natively (FUN_00622d00, the "moving" test, too).
  Hooked: the block chance FUN_00622720 (the case's value), the dodge
  FUN_0057dd60 (0, so the block decision is isolated), the bookkeeping
  FUN_00554010 / FUN_005347c0. Diffed against two ports: the reading (no
  block when can-block is 0; a player in a moving mode other than WL blocks
  a third; roll = the defender's seed % 100) and d2d's
  rules::monster_blow (a third while running, no can-block gate).
- `roll`: FUN_005a89a0 (the missile's damage record), physical part with no
  target: min + rand(max - min) on the missile's stats 21 / 22 (256ths), top
  exclusive; d2d rolls `rng.range(a2_min, a2_max)` in whole points, inclusive
  (the printed mean gap is on random 256ths ends; on whole ends it is +1/512).
- `--tables DIR` (monstats.txt, missiles.txt, skills.txt from the 1.14d
  patch layer): every monster's MissA1 / MissA2 and every monster skill's
  srvmissilea row, by game.exe's two gates (ToHit, physical damage) against
  d2d's paths (MissA: always rolls, A1TH unless S1; a skill's missile:
  never rolls, always blockable).
"""
import csv
import os
import random
import sys

import emu


def lcg(seed):
    product = seed[0] * 0x6AC690C5 + seed[1]
    seed[0], seed[1] = product & 0xFFFFFFFF, product >> 32
    return seed[0]


def rand(seed, bound):
    if bound < 1: return 0
    v = lcg(seed)
    return v & (bound - 1) if bound & (bound - 1) == 0 else v % bound


MODES = (1, 1, 2, 2, 3, 3, 3, 0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18)


def tables(d):
    def rows(name):
        r = list(csv.reader(open(os.path.join(d, name), encoding="latin-1"), delimiter="\t"))
        return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != "Expansion"]
    missiles = {m["Missile"]: m for m in rows("missiles.txt")}
    skills = {s["skill"]: s for s in rows("skills.txt")}
    early = {"fallen1", "fallenshaman1", "zombie1", "quillrat1", "skeleton1", "sk_archer1", "brute2", "cr_archer1",
             "corruptrogue1", "fallen2", "quillrat2", "zombie2", "skeleton2", "sk_archer2", "fallenshaman2", "cr_archer2", "cr_lancer1"}
    out = {"missa_nohit": [], "a2_th": [], "skill_tohit": [], "noblock": [], "missa_rows": 0, "skill_rows": 0}
    for mon in rows("monstats.txt"):
        for col, th in (("MissA1", "A1TH"), ("MissA2", "A2TH")):
            m = missiles.get(mon.get(col, ""))
            if not m: continue
            out["missa_rows"] += 1
            tag = mon["Id"] + ("*" if mon["Id"] in early else "")
            if not m.get("ToHit"): out["missa_nohit"].append(f"{tag}:{m['Missile']}")
            if col == "MissA2" and mon.get("A2TH") and mon.get("A2TH") != mon.get("A1TH"):
                out["a2_th"].append(f"{tag}:{m['Missile']} A1TH {mon.get('A1TH')} A2TH {mon.get('A2TH')}")
        for k in range(1, 9):
            s = skills.get(mon.get(f"Skill{k}", ""))
            m = missiles.get(s.get("srvmissilea", "")) if s else None
            if not m: continue
            out["skill_rows"] += 1
            tag = mon["Id"] + ("*" if mon["Id"] in early else "")
            if m.get("ToHit") and s["skill"] != "Quick Strike": out["skill_tohit"].append(f"{tag}:{s['skill']}->{m['Missile']}")
            if not m.get("MinDamage") and not m.get("MaxDamage") and not m.get("SrcDamage"):
                out["noblock"].append(f"{tag}:{m['Missile']}")
    print(f"tables: {out['missa_rows']} MissA rows, {out['skill_rows']} monster-skill missile rows (* = early Act 1)")
    early_nohit = [x for x in out['missa_nohit'] if '*' in x]
    print(f"  MissA rows with ToHit blank (game: no roll; d2d MissA path: rolls): {len(out['missa_nohit'])} ({len(early_nohit)} early: {' '.join(early_nohit)}): {' '.join(out['missa_nohit'][:12])}")
    print(f"  MissA2 shooters with A2TH != A1TH (game: the owner's stat 19 at impact; d2d: A1TH): {len(out['a2_th'])}: {'; '.join(out['a2_th'][:8])}")
    print(f"  monster-skill missiles with ToHit set (game rolls; d2d missile_blow: never; Quick Strike excluded, d2d sends it down the MissA path): {len(out['skill_tohit'])}: {' '.join(sorted(set(out['skill_tohit']))[:12])}")
    print(f"  monster-skill missiles with no physical part (game: cannot be blocked; d2d: blockable): {len(set(out['noblock']))} rows, e.g. {' '.join(sorted(set(out['noblock']))[:10])}")


def main():
    argv, exe, tdir, args = sys.argv[1:], None, None, []
    while argv:
        a = argv.pop(0)
        if a == "--exe": exe = argv.pop(0)
        elif a == "--tables": tdir = argv.pop(0)
        else: args.append(a)
    cases = int(args[0]) if args else 2000
    if tdir: tables(tdir)
    rng = random.Random(11)
    e = emu.Emu(exe) if exe else emu.Emu()
    stats, block = {}, {}
    dodge_called = [0]
    e.hook(0x625480, lambda e: stats.get((e.arg(0), e.arg(1)), 0), 3)
    e.hook(0x625500, lambda e: stats.get((e.arg(0), e.arg(1)), 0), 3)
    e.hook(0x622720, lambda e: block.get(e.arg(0), 0), 2)
    e.hook(0x57dd60, lambda e: dodge_called.__setitem__(0, dodge_called[0] + 1) or 0, 1)
    e.hook(0x554010, lambda e: 0, 0)
    e.hook(0x5347c0, lambda e: 0, 4)
    for a in (0x63e940, 0x63e990): e.hook(a, lambda e: 0, 1)

    # ---- block ----
    game = e.alloc(0x200)
    moving = {}
    bad_read = bad_d2d = 0
    kinds = {}
    for case in range(cases):
        att, dfd = e.alloc(0x100), e.alloc(0x100)
        e.w32(att, 1); e.w32(dfd, 0)
        mode = rng.choice(MODES)
        e.w32(dfd + 0x10, mode)
        seed = [rng.getrandbits(32), rng.getrandbits(32)]
        e.w32(dfd + 0x20, seed[0]); e.w32(dfd + 0x24, seed[1])
        b = rng.choice((0, 0, 10, 25, 40, 50, 60, 75, rng.randrange(1, 100)))
        missile, canblock = rng.choice((0, 1)), rng.choice((0, 1, 1))
        block[dfd] = b
        got = e.call(0x57dfb0, dfd, missile, canblock, ecx=game, edx=att)
        roll = lcg(list(seed)) % 100
        # the reading: FUN_00622d00's moving modes found by the run itself are checked below
        want = 1 if canblock and b > 0 and roll < (b // 3 if mode in (3, 6) else b) else 0   # RN, TW (town walk)
        d2d = 1 if b > 0 and roll < (b // 3 if mode == 3 else b) else 0
        if got != want:
            bad_read += 1
            if bad_read <= 10: print(f"block case {case}: mode {mode} block {b} missile {missile} canblock {canblock} roll {roll}: game {got}, read {want}")
        if got != d2d:
            bad_d2d += 1
            k = ("canblock 0" if not canblock else f"mode {mode}") + (" missile" if missile else " melee")
            kinds[k] = kinds.get(k, 0) + 1
    print(f"block: {cases} cases, {bad_read} differ from the reading (a third in RN / TW, none when can-block 0); "
          f"{bad_d2d} differ from d2d's monster_blow: {dict(sorted(kinds.items()))}")

    # ---- roll ----
    bad = 0
    gap = 0.0
    for case in range(cases):
        mis = e.alloc(0x100)
        e.w32(mis, 3)
        seed = [rng.getrandbits(32), rng.getrandbits(32)]
        e.w32(mis + 0x20, seed[0]); e.w32(mis + 0x24, seed[1])
        lo = rng.randrange(1, 40) << 8 if rng.random() < 0.5 else rng.randrange(1, 40 << 8)
        hi = lo + (rng.randrange(0, 40) << 8 if rng.random() < 0.5 else rng.randrange(0, 40 << 8))
        stats.clear(); stats.update({(mis, 0x15): lo, (mis, 0x16): hi})
        rec = e.alloc(0x80)
        e.call(0x5a89a0, mis, 0, rec, ecx=game, edx=0)
        got = e.s32(rec + 8)
        want = lo + rand(list(seed), hi - lo)
        if got != want:
            bad += 1
            if bad <= 10: print(f"roll case {case}: {lo}-{hi}: game {got}, port {want}")
        # d2d rolls whole points inclusive of the row's ends cut to whole points
        game_mean = (lo + (lo + max(hi - lo - 1, 0))) / 2 / 256
        d2d_mean = ((lo >> 8) + max(hi >> 8, lo >> 8)) / 2
        gap += d2d_mean - game_mean
    print(f"roll: {cases} cases, {bad} differ (min + rand(max - min) in 256ths, top exclusive); "
          f"d2d whole-point inclusive mean minus game mean: {gap / cases:+.3f} hp a hit on these random 256ths ends")


if __name__ == "__main__": main()
