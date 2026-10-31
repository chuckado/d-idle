#!/usr/bin/env python3
"""game.exe 1.14d's walk / run path velocity as the oracle against d2d (d2lod-bot, card decompwalk).

The velocity a moving mode gets is set once at the mode's start by
FUN_00623f50 (0x624113..0x6241c9): walk % = FUN_00621740(unit, row 4: stat
96 faster run/walk, 150 v / (150 + v) for any v != 0) + stat 0x43
(velocitypercent), at least 25; path velocity = FUN_00621360(type, class) x
walk % / 100 (a player's CharStats WalkVelocity << 8, a monster's MonStats
Velocity << 8 in ANY moving mode, RN included), stored by FUN_00648690 at
path +0x7c. The per-tick step is velocity / 4096 subtiles (FUN_0064fe40,
docs/research/movement-and-recovery-trace.md), so the velocity IS the step.

    gen   OUT              write the cases (exhaustive grids, see below)
    game  CASES > game.tsv run game.exe's code from 0x624113 to 0x6241ce on each case
                           (unicorn, upstream tools/emu/emu.py on PYTHONPATH,
                           D2_MPQ_DIR set, --exe the read-only game.exe copy);
                           prints `id velocity`
    cmp   CASES game.tsv d2d.tsv   d2d.tsv from walk_speed_harness.cpp (`id before after`);
                           counts mismatches (|d2d - game| >= 1 unit of 1/4096
                           subtile a tick) before and after, per term

Inputs the grids need: monstats.txt / monstats2.txt (patched 1.14d, mpq-cat
recipe docs/process/mpq-data-reading.md) in --data DIR.

Case line: `id kind cls mode vel run d frw chill cold walkv runv`
  kind 0 player (mode 2 WL / 3 RN), 1 monster (mode 2 WL / 15 RN);
  vel / run: monster MonStats Velocity / Run, player CharStats WalkVelocity / RunVelocity;
  d: monster the non-chill velocitypercent deltas (aura, boss, AI pace),
     player the velocitypercent deltas (Vigor, Burst of Speed, ...);
  frw: player stat 96; chill 0/1; cold: monster MonStats coldeffect (Normal), player -50.
game.exe gets stat 0x43 = 100 + d + chill x cold (+ RunVelocity x 100 / WalkVelocity - 100
for a running player, FUN_00620e80) and stat 96 = frw.
"""
import csv
import os
import sys

TERMS = ("M1_rn_uses_run", "M2_chill_multiplies", "M3_floor_10", "P1_vp_diminished", "P2_negative_frw")


def rows(data):
    ms = list(csv.DictReader(open(os.path.join(data, "monstats.txt"), encoding="latin1"), delimiter="\t"))
    m2 = {x["Id"]: x for x in csv.DictReader(open(os.path.join(data, "monstats2.txt"), encoding="latin1"), delimiter="\t")}
    out = []
    for i, x in enumerate(ms):
        if not x.get("Id") or x["Id"] == "Expansion":
            continue
        num = lambda k: int(x[k]) if x.get(k, "").lstrip("-").isdigit() else 0
        rn = m2.get(x.get("MonStatsEx", ""), {}).get("mRN", "") == "1"
        out.append((i, num("Velocity"), num("Run"), num("coldeffect"), rn))
    return out


def gen(path, data):
    n = 0
    with open(path, "w") as f:
        f.write("# id kind cls mode vel run d frw chill cold walkv runv\n")
        for (i, vel, run, cold, rn) in rows(data):
            for mode in ((2, 15) if rn else (2,)):
                for d in range(-100, 201):
                    for chill in (0, 1):
                        f.write(f"{n} 1 {i} {mode} {vel} {run} {d} 0 {chill} {cold} 0 0\n")
                        n += 1
        for cls in range(7):                        # every class: WalkVelocity 6, RunVelocity 9 (D)
            for mode in (2, 3):
                for frw in range(-60, 201):
                    for vp in range(-50, 101, 5):
                        for chill in (0, 1):
                            f.write(f"{n} 0 {cls} {mode} 6 9 {vp} {frw} {chill} -50 6 9\n")
                            n += 1
    print(n, "cases", file=sys.stderr)


def read_cases(path):
    for line in open(path):
        if line.startswith("#") or not line.strip():
            continue
        yield [int(v) for v in line.split()]


def game(exe, path, data):
    import struct
    import emu
    from unicorn import UC_HOOK_CODE
    e = emu.Emu(exe)
    mu = e.mu
    table = rows(data)
    nmon = max(i for (i, *_r) in table) + 1
    dt = e.alloc(0x1000)
    mon = e.alloc(nmon * 0x1a8)
    for (i, vel, run, cold, rn) in table:
        rec = mon + i * 0x1a8
        mu.mem_write(rec + 0x32, struct.pack("<hhhh", vel, run, 256, 256))
    chars = e.alloc(7 * 0xc4)
    for c in range(7):
        mu.mem_write(chars + c * 0xc4 + 0x40, bytes((6, 9)))
    e.w32(dt + 0xa78, mon)
    e.w32(dt + 0xa80, nmon)
    e.w32(dt + 0xbc4, chars)
    e.w32(dt + 0xbc8, 7)
    e.w32(0x744304, dt)
    stat = {96: 0, 0x43: 100}
    e.hook(0x625500, lambda e_: stat[96] & 0xFFFFFFFF, 3)      # FUN_00621740's getter (row 4: stat 96)
    e.hook(0x625480, lambda e_: stat[0x43] & 0xFFFFFFFF, 3)    # stat 0x43 velocitypercent
    hit = {"stop": 0}

    def notmoving(mu_, addr, size, _):
        hit["stop"] = 1
        mu_.emu_stop()
    mu.hook_add(UC_HOOK_CODE, notmoving, begin=0x6241d7, end=0x6241d7)
    unit = e.alloc(0x100)
    path_ = e.alloc(0x100)
    anim = e.alloc(0x40)
    frame = emu.STACK + emu.STACK_SIZE - 0x2000
    for (cid, kind, cls, mode, vel, run, d, frw, chill, cold, walkv, runv) in read_cases(path):
        s43 = 100 + d + (cold if chill else 0)
        if kind == 0 and mode == 3:
            s43 += runv * 100 // walkv - 100
        stat[96], stat[0x43] = frw, s43
        e.w32(unit, kind)
        e.w32(unit + 0x2c, path_)
        e.w32(path_ + 0x7c, 0xDEADBEEF)
        e.w32(frame - 4, cls)
        e.w32(frame - 8, mode)
        e.w32(frame - 0xc, anim)
        e.w32(frame + 8, kind)
        hit["stop"] = 0
        mu.reg_write(emu.UC_X86_REG_EBP, frame)
        mu.reg_write(emu.UC_X86_REG_ESP, frame - 0x60)
        mu.reg_write(emu.UC_X86_REG_EDI, unit)
        mu.reg_write(emu.UC_X86_REG_EBX, kind)
        mu.reg_write(emu.UC_X86_REG_EAX, mode)
        mu.emu_start(0x624113, 0x6241ce, count=4000)
        v = e.s32(path_ + 0x7c)
        print(cid, "notmoving" if hit["stop"] or v == -559038737 else v)


def cmp(cases, gpath, dpath):
    g = {}
    for line in open(gpath):
        a = line.split()
        g[int(a[0])] = a[1]
    dd = {}
    for line in open(dpath):
        a = line.split()
        dd[int(a[0])] = (float(a[1]), float(a[2]))
    n = before = after = notmoving = 0
    per = {t: 0 for t in TERMS}
    per_kind = {0: [0, 0, 0], 1: [0, 0, 0]}
    examples = []
    for c in read_cases(cases):
        cid, kind, cls, mode, vel, run, d, frw, chill, cold, walkv, runv = c
        gv = g.get(cid)
        if gv is None or gv == "notmoving":
            notmoving += 1
            continue
        gv = int(gv)
        b, a = dd[cid]
        n += 1
        per_kind[kind][0] += 1
        if abs(b - gv) >= 1:
            before += 1
            per_kind[kind][1] += 1
            if len(examples) < 12 and (kind == 1 and mode == 15 or len(examples) < 6):
                examples.append((c, gv, b))
            if kind == 1:
                if mode == 15 and run != vel:
                    per["M1_rn_uses_run"] += 1
                elif chill and cold and 100 + d + cold >= 25:
                    per["M2_chill_multiplies"] += 1
                else:
                    per["M3_floor_10"] += 1
            else:
                if d != 0:
                    per["P1_vp_diminished"] += 1
                else:
                    per["P2_negative_frw"] += 1
        if abs(a - gv) >= 1:
            after += 1
            per_kind[kind][2] += 1
    print(f"cases {n} (not a moving mode: {notmoving}); before {before} differ; after {after} differ")
    for k, (t, b, a) in per_kind.items():
        print(f"  {'monster' if k else 'player'}: {t} cases, before {b}, after {a}")
    for t in TERMS:
        print(f"  term {t}: {per[t]}")
    for (c, gv, b) in examples:
        print("  e.g. case", c, "game", gv, "d2d", round(b, 2))


def main():
    args = sys.argv[1:]
    opt = {}
    for k in ("--exe", "--data"):
        if k in args:
            at = args.index(k)
            opt[k] = args[at + 1]
            del args[at:at + 2]
    data = opt.get("--data", ".")
    if not args:
        raise SystemExit(__doc__)
    if args[0] == "gen":
        gen(args[1], data)
    elif args[0] == "game":
        game(opt.get("--exe"), args[1], data)
    elif args[0] == "cmp":
        cmp(args[1], args[2], args[3])
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
