#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""A player's gear absorb (stats 142..149) on one elemental or magic field: the oracle for `monster_blow`'s element fields.

    uv run python gear_absorb.py gen 20000 7 > cases.tsv
    uv run python gear_absorb.py game cases.tsv > g.tsv          # game.exe, under the emulator
    g++ -O2 -std=c++23 -I../../components/rules -I../../components -I../../components/d2s -I../../components/txt -o gear_absorb_harness gear_absorb_harness.cpp && ./gear_absorb_harness < cases.tsv > d.tsv
    uv run python gear_absorb.py compare cases.tsv g.tsv d.tsv   # exit 1 on a difference, an empty file or a malformed line

A case line: `id type field res pct flat mdr`. `type` 0 fire, 1 lightning, 2 cold, 4 magic (d2d's element
index); `field` the damage in whole points, written into the record's field (<< 8); `res` the defender's
resistance after the cap (-100..75, stats 39 / 41 / 43; a magic field has none on a monster blow); `pct` and
`flat` the gear's absorb percent and flat (fire 142 / 143, lightning 144 / 145, magic 146 / 147, cold 148 / 149);
`mdr` the magic damage reduction (stat 35). `game` prints `id fine total`: the field after FUN_0057c1e0 and
the record total (+0x4c), both in 256ths.

game.exe runs FUN_0057c1e0 natively on a player defender in a classic Normal game (no difficulty penalty):
FUN_0057bf80 per field, whose absorb is FUN_0057bf10 (percent capped at 40, then the flat, both after the
resistance). Hooked, not the claim: the stat getter, the difficulty row, the pair percent (100), the unit tests the
state event calls make. Tags: the order and the cap are V run against game.exe's code.
"""
import os
import random
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import monster_damage  # noqa: E402  (the rig of the monster damage oracle)

# type -> (record offset, resist stat, max-resist stat, absorb percent stat, absorb flat stat)
ROWS = {0: (0x10, 0x27, 0x28, 0x8E, 0x8F), 1: (0x1C, 0x29, 0x2A, 0x90, 0x91), 2: (0x24, 0x2B, 0x2C, 0x94, 0x95), 4: (0x20, 0x25, 0x26, 0x92, 0x93)}
MDR_STAT = 35


def gen(count, seed):
    rng = random.Random(seed)
    lines = []
    for index in range(count):
        type_ = rng.choice((0, 1, 2, 4))
        field = rng.choice((1, 2, 5, 20, rng.randint(1, 60), rng.randint(1, 400)))
        res = 0 if type_ == 4 else rng.choice((0, rng.randint(-100, 75), rng.randint(0, 75)))
        pct = rng.choice((0, 0, rng.randint(1, 60), rng.randint(1, 45), 40, 41))
        flat = rng.choice((0, 0, rng.randint(1, 30), rng.randint(1, field + 5)))
        mdr = rng.choice((0, 0, rng.randint(1, 20), field, field + 1))
        lines.append(f"{index} {type_} {field} {res} {pct} {flat} {mdr}")
    return lines


def parse(line):
    words = line.split()
    if len(words) != 7:
        raise ValueError(f"bad case line {line!r}")
    return tuple(int(word) for word in words)


def read(path):
    with open(path, encoding="utf-8") as handle:
        return [line for line in handle.read().splitlines() if line.strip()]


def game(path):
    oracle = monster_damage.Oracle()
    e = oracle.e
    for line in read(path):
        index, type_, field, res, pct, flat, mdr = parse(line)
        offset, res_stat, max_stat, pct_stat, flat_stat = ROWS[type_]
        unit_addr = e.alloc(0x200)
        unit = e.view("unit", unit_addr)
        unit.type, unit["class"], unit.mode = 0, 0x10, 1
        oracle.stats.clear()
        oracle.stats.update({(unit_addr, res_stat): res, (unit_addr, pct_stat): pct, (unit_addr, flat_stat): flat, (unit_addr, MDR_STAT): mdr})
        oracle.cur["list"] = e.alloc(0x40)
        oracle.state.clear()
        record_addr = e.alloc(0x100)
        record = e.view("damage", record_addr)
        record.flags, record.physical = 1 | 4, 0
        e.w32(record_addr + offset, field << 8)
        oracle.out.clear()
        e.call(0x57C1E0, unit_addr, record_addr, ecx=oracle.game, edx=oracle.attacker)
        print(index, e.s32(record_addr + offset), e.s32(record_addr + 0x4C))


def compare(cases, gfile, dfile):
    case_lines, game_lines, d2d_lines = read(cases), read(gfile), read(dfile)
    if not case_lines:
        raise SystemExit("compare: empty case file")
    if not (len(case_lines) == len(game_lines) == len(d2d_lines)):
        raise SystemExit(f"compare: lengths differ ({len(case_lines)}, {len(game_lines)}, {len(d2d_lines)})")
    bad, by_kind = 0, {}
    for case_line, g_line, d_line in zip(case_lines, game_lines, d2d_lines, strict=True):
        case, g, d = parse(case_line), g_line.split(), d_line.split()
        if len(g) != 3 or len(d) != 3 or g[0] != d[0] or g[0] != str(case[0]):
            raise SystemExit(f"compare: malformed or out-of-step line near id {case[0]}")
        differ = g[1:] != d[1:]
        kind = "absorb" if case[4] > 0 or case[5] > 0 else "plain"
        row = by_kind.setdefault(kind, [0, 0])
        row[0] += 1
        row[1] += differ
        bad += differ
    for kind, (total, differ) in sorted(by_kind.items()):
        print(f"# {kind}: {total} cases, {differ} differ")
    print(f"# all: {len(case_lines)} cases, {bad} differ")
    sys.exit(1 if bad else 0)


def main():
    if len(sys.argv) < 2 or sys.argv[1] not in ("gen", "game", "compare"):
        raise SystemExit(__doc__)
    cmd = sys.argv[1]
    if cmd == "gen" and len(sys.argv) == 4:
        print("\n".join(gen(int(sys.argv[2]), int(sys.argv[3]))))
    elif cmd == "game" and len(sys.argv) == 3:
        game(sys.argv[2])
    elif cmd == "compare" and len(sys.argv) == 5:
        compare(*sys.argv[2:5])
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
