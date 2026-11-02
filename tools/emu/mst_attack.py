# SPDX-License-Identifier: GPL-3.0-or-later
"""mercstats: game.exe's merc attack think (FUN_005e5050 with its skill pick FUN_005e4d30) as the oracle for d2d's
C++ rules::merc_attack.

    mst_attack.py CASES SEED OUTFILE      one line per case:
        A cls level aip1 gap melee ai14 lo hi rowlevel default  ids[6] chance[6] per[6] mode[6] levels[6] aura[6] running[6]
          skill1 sk1mode BITS ::: effects ::: ai14_after lo hi

The cases are upstream's tools/emu/merc_attack.py's: a hireling Id, a level 1..98, the row game.exe picks (FUN_006562f0), the
skills it has at random levels, some auras already running, the gap to the foe (0..20), in melee or not, the AI control's +0x14
and the seed. The hooks are upstream's too (the stat read, the skill queries, the effect hand-offs). The effects are what
game.exe did: each move tried (`move 1` found a path), a skill (`skill mode id`), an aura (`aura id`, then `stand 10`), a swing,
a stand. tools/emu/mst_d2d_test.cpp `attack` replays the line through d2d's real rules::merc_attack and prints the line again.
"""
import random
import struct
import sys

import drlg
import merc_attack as ma
from unicorn.x86_const import UC_X86_REG_EDX


def fx(items):
    parts = []
    for it in items:
        if it[0] == "move":
            parts.append("move %d" % it[1])
        elif it[0] == "skill":
            parts.append("skill %d %d" % (it[1], it[2]))
        elif it[0] == "aura":
            parts.append("aura %d" % it[1])
        elif it[0] == "stand":
            parts.append("stand %d" % it[1])
        else:
            parts.append(it[0])
    return ";".join(parts)


def main():
    cases, seed0, path = int(sys.argv[1]), int(sys.argv[2]), sys.argv[3]
    rng = random.Random(seed0)
    outfile = open(path, "w")
    e = drlg.boot()
    rows = ma.hirelings(e)
    ids = sorted({r["id"] for r in rows if r["version"] == 100})
    t = e.r32(0x744304)
    skill_count, state_count = e.r32(t + 0xba0), e.r32(t + 0xc4)
    aura_mask = e.r32(0x6ce27c)
    monstats = e.r32(t + 0xa78)
    sk_cache = {}

    def sk(sid):
        if sid not in sk_cache:
            sk_cache[sid] = ma.skill_info(e, sid)
        return sk_cache[sid]
    cur = {}

    def out(item):
        cur["out"].append(item)
        return 0
    e.hook(0x625480, lambda e: cur["s"]["level"] if e.arg(1) == 0xc else 0, 3)

    def record(e):
        rec = e.alloc(0x20)
        e.w32(rec + 8, cur["s"]["hid"])
        return rec
    e.hook(0x574bd0, record, 0)
    e.hook(0x620510, lambda e: cur["sizes"].get(e.arg(0), 1), 1)
    e.hook(0x6439b0, lambda e: e.mu.reg_read(UC_X86_REG_EDX) if cur["s"]["have"].get(e.mu.reg_read(UC_X86_REG_EDX), 0) > 0 else 0, 1)
    e.hook(0x6442a0, lambda e: cur["s"]["have"].get(e.arg(1), 0), 3)
    e.hook(0x639df0, lambda e: int(e.arg(1) in cur["s"]["states"]), 2)
    e.hook(0x622c40, lambda e: int(cur["s"]["melee"]), 3)
    e.hook(0x5dead0, lambda e: out(("skill", e.arg(0) & 0xff, e.arg(1))), 5)
    e.hook(0x5701b0, lambda e: out(("aura", e.arg(0))), 2)
    e.hook(0x5ddf90, lambda e: out(("swing",)), 2)
    e.hook(0x5de080, lambda e: out(("stand", e.arg(0))), 1)
    e.hook(0x540e60, lambda e: 0, 2)

    def setoff(e):
        ok = cur["moves"].random() >= cur["s"]["fails"]
        out(("move", int(ok)))
        return int(ok)
    e.hook(0x5a7c20, setoff, 3)
    e.hook(0x5a6260, lambda e: 0, 2)
    mark = e.brk
    for case in range(cases):
        if e.brk - mark > 0x2000000:
            e.brk = mark
        hid = rng.choice(ids)
        lvl = rng.randint(1, 98)
        row = ma.row_for(rows, hid, lvl)
        cls = row["cls"]
        msrow = monstats + cls * 0x1a8
        diff = rng.randint(0, 2)
        aip1 = struct.unpack("<h", e.read(msrow + 0x56 + diff * 2, 2))[0]
        have = {}
        for sid in row["skills"]:
            if 0 < sid < skill_count and rng.random() < 0.85:
                have[sid] = rng.randint(1, 30)
        states = set()
        for sid in have:
            info = sk(sid)
            if info["aura"] == 1 and info["state"] >= 0 and rng.random() < 0.4:
                states.add(info["state"])
        gap = rng.choice((0, 1, 2, 3, 5, 9, 20))
        s = {"cls": cls, "hid": hid, "level": lvl, "ai14": rng.choice((0, 0, 10, 20, 30, 60)), "aip1": aip1, "gap": gap,
             "melee": rng.random() < 0.5, "have": have, "states": states, "fails": rng.random(),
             "skill1": struct.unpack("<h", e.read(msrow + 0x170, 2))[0], "sk1mode": e.read(msrow + 0x180, 1)[0],
             "seed": (rng.getrandbits(32), rng.getrandbits(32))}
        game = e.alloc(0x2000)
        e.mu.mem_write(game + 0x6d, bytes([diff]))
        e.w32(game + 0x70, 1)
        merc_u, tgt, owner, mpath, tpath, params, ai, ms2 = (e.alloc(n) for n in (0x200, 0x200, 0x200, 0x200, 0x200, 0x40, 0x100, 0x40))
        for a, n in ((merc_u, 0x200), (tgt, 0x200), (mpath, 0x200), (tpath, 0x200), (params, 0x40), (ai, 0x100), (ms2, 0x40)):
            e.mu.mem_write(a, bytes(n))
        mx, my = 5000, 5000
        tx, ty = mx + gap + 1, my
        e.w32(merc_u, 1); e.w32(merc_u + 4, cls); e.w32(merc_u + 0xc, 77); e.w32(merc_u + 0x2c, mpath)
        e.w32(merc_u + 0x20, s["seed"][0]); e.w32(merc_u + 0x24, s["seed"][1])
        e.w32(mpath, mx << 16 | 0x8000); e.w32(mpath + 4, my << 16 | 0x8000); e.w32(mpath + 0x30, merc_u)
        e.w32(tgt, 1); e.w32(tgt + 0x2c, tpath); e.w32(tpath, tx << 16 | 0x8000); e.w32(tpath + 4, ty << 16 | 0x8000); e.w32(tpath + 0x30, tgt)
        e.w32(ai + 0x14, s["ai14"])
        e.w32(params, ai); e.w32(params + 0x1c, msrow); e.w32(params + 0x20, ms2)
        cur.update(s=s, out=[], moves=random.Random(case), sizes={merc_u: 1, tgt: 1})
        e.call(0x5e5050, cls, owner, tgt, merc_u + 0x20, params, ecx=game, edx=merc_u)
        effects, ai14_after, seed_after = cur["out"], e.r32(ai + 0x14), (e.r32(merc_u + 0x20), e.r32(merc_u + 0x24))
        rm = random.Random(case)
        bits = "".join("1" if rm.random() >= s["fails"] else "0" for _ in range(100))
        skill_ids = [sid if 0 < sid < skill_count else 0 for sid in row["skills"]]
        levels = [have.get(sid, 0) for sid in skill_ids]
        auras = [int(bool(sk(sid)["flags"] & aura_mask)) if sid else 0 for sid in skill_ids]
        running = [int(bool(sid) and sk(sid)["aura"] == 1 and sk(sid)["state"] in states) for sid in skill_ids]
        j = lambda xs: " ".join(str(x) for x in xs)
        print("A %d %d %d %d %d %d %d %d %d %d  %s  %s  %s  %s  %s  %s  %s  %d %d %s ::: %s ::: %d %d %d" % (
            cls, lvl, aip1, gap, int(s["melee"]), s["ai14"], s["seed"][0], s["seed"][1], row["level"], row["default"],
            j(skill_ids), j(row["chance"]), j(row["per"]), j(row["mode"]), j(levels), j(auras), j(running),
            s["skill1"], s["sk1mode"], bits, fx(effects), ai14_after, seed_after[0], seed_after[1]), file=outfile)
    outfile.close()


main()
