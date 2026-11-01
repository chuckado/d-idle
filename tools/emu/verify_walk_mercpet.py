#!/usr/bin/env python3
"""Independent verifier for the merc and pet claims of upstream candidate fix/walk-velocity-v2
(d2lod-bot card verifyprs2: commits 455e366f merc, 85d50c78 pet).

Written without reading the author's oracle (tools/emu/walk_velocity.py). game.exe side re-derived
from disasm.py reads:
- FUN_00623f50 (0x624113..0x6241c9): a unit in a moving mode with a path: pct = FUN_00621740 row 4
  (stat 96, diminished) + stat 0x43, at least 0x19; path velocity = FUN_00621360(type, class) x pct
  / 100 (imul, truncating) to FUN_00648690. Type and class come from unit +0 / +4 through
  FUN_00645270, which remaps them only when unit +0xc8 bit 3 is set and the unit is in a state of
  States gfxtype 1 or 2 (a Shadow's gfxtype 2: player type, class = gfxclass; not claimed).
- FUN_00621360: type 1 (any monster unit: a merc and a pet are monster units) -> MonStats +0x32
  (Velocity) << 8 whatever the mode; type 0 -> CharStats +0x40 (WalkVelocity) << 8.
- FUN_006214a0: a monster's mode 2 (WL) and 15 (RN) is a moving mode both by MonStats +0xd bit 0
  and by the mode tables at 0x6e8b90 (hcIdx < 0x19a) / 0x6e8cd0 (the act 5 mercs 0x230, 0x231),
  row +4 = 1 for modes 2 and 15 in both.
- The pace: FUN_005a6260 stores an AI's pct at control +0x1c; FUN_005a63f0 moves it to +0x10 for
  path types other than 100 (WL / RN / 13 / 14 get 101 -> 13 from the table at 0x6e23d0);
  FUN_005a6380 (from FUN_005a7c20 at 0x5a7dbf) puts a non-zero +0x10 on stat 0x43 in a stat list
  (0x4004) and calls FUN_00623f50. The merc's run at the owner's footsteps paces 0x3c
  (FUN_005e52d0); its run at a foe (FUN_005ded00 -> FUN_005deb60, mode 0xf) sets no pace.

game side: game.exe's own FUN_00623f50, FUN_00645270 (unit +0xc8 = 0: no remap), FUN_006214a0,
FUN_006216e0, FUN_00621740 and FUN_00621360 run natively under unicorn (upstream tools/emu/emu.py)
against a fake data-table block at [0x744304] (+0xa78 MonStats rows of 0x1a8, Velocity at +0x32,
+0xd flag byte from the case). Hooked: FUN_00620f00 (no-op), FUN_006213d0 (animation rate, 256),
FUN_00625480 / FUN_00625500 (stat reads: 0x43 and 96 from the case), FUN_00648690 (records the
velocity). The inputs of stat 0x43 are d2d's model, taken as given: a merc's 100 + move_pct, a
pet's 100 + move_pct + speed_pct (its summon's velocitypercent) + buff velocitypercent; stat 96
(faster run/walk) is 0 for both (d2d models none for a merc or a pet).

d2d side: the REAL Fight::merc_turn / Fight::pet_walking of the build under test
(`test_game --verify-mercpet`, verify_walk_mercpet_mode.inc pasted into a throwaway worktree).

    python verify_walk_mercpet.py gen N SEED > cases.tsv
    python verify_walk_mercpet.py game cases.tsv > game.tsv
    build/tests/test_game --verify-mercpet cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
    python verify_walk_mercpet.py compare cases.tsv game.tsv d2d.tsv

cases.tsv kind 2: id 2 merc_type mode(2 WL / 15 RN) velocity move flag
          kind 3: id 3 monstats-id skill-id mode velocity move speed buff flag
game.tsv: id pct velocity (velocity = what FUN_00648690 gets).
"""
import random
import struct
import sys

# (MonStats Id, Skills.txt summoning skill) of the non-Shadow pets; revives use ordinary rows. The skill only
# selects pet_walking's Shadow branch (its aurastate), so the vines' exact skill row does not matter.
PETS = [("necroskeleton", "Raise Skeleton"), ("necromage", "Raise Skeletal Mage"), ("claygolem", "Clay Golem"),
        ("bloodgolem", "BloodGolem"), ("irongolem", "IronGolem"), ("firegolem", "FireGolem"), ("valkyrie", "Valkyrie"),
        ("spiritwolf", "Summon Spirit Wolf"), ("fenris", "Summon Fenris"), ("druidbear", "Summon Grizzly"),
        ("plaguepoppy", "Vines"), ("cycleoflife", "VineCycler"), ("vinecreature", "Vines"),
        ("oaksage", "Oak Sage"), ("heartofwolverine", "Heart of Wolverine"), ("spiritofbarbs", "Spirit of Barbs"),
        ("zombie1", "Revive"), ("fallen1", "Revive"), ("skeleton1", "Revive"), ("goatman1", "Revive"),
        ("brute1", "Revive"), ("wraith1", "Revive"), ("bloodlord1", "Revive"), ("doomknight1", "Revive")]
PACES = (0, 0, 0x3c, 0xf, 30, 120)            # a merc's / pet's AI paces (0x3c the merc's run after its owner)


def tab(name):
    sys.path.insert(0, ".")
    from emu import mpq_cache
    t = mpq_cache().read(name).decode("latin1").split("\r\n")
    h = t[0].split("\t")
    return h, [r.split("\t") for r in t[1:] if r and len(r.split("\t")) > 3]


def tables():
    h, rows = tab("data\\global\\excel\\monstats.txt")
    mon = {r[0]: int(r[h.index("hcIdx")]) for r in rows if r[h.index("hcIdx")].strip().lstrip("-").isdigit()}
    h, rows = tab("data\\global\\excel\\hireling.txt")
    by_hc = {v: k for k, v in mon.items()}
    mercs = {}
    for r in rows:
        if r[h.index("Version")] != "100":
            continue
        mid, cls = int(r[h.index("Id")]), int(r[h.index("Class")])
        if cls in by_hc:
            mercs.setdefault(mid, cls)
    h, rows = tab("data\\global\\excel\\skills.txt")
    skills = {r[h.index("skill")]: int(r[h.index("Id")]) for r in rows if r[h.index("Id")].isdigit()}
    return mon, mercs, skills


def gen(n, seed):
    mon, mercs, skills = tables()
    rng = random.Random(seed)
    pets = [(m, skills[s]) for m, s in PETS if m in mon and s in skills]
    missing = [m for m, s in PETS if m not in mon or s not in skills]
    if missing:
        print("missing pets:", missing, file=sys.stderr)
    rows = []
    sums = (-30000, -400, -101, -100, -76, -75, -74, 0, 9, 10, 11, 24, 25, 26, 99, 100, 101, 115, 159, 160, 161, 220, 1000, 30000)
    vels = (1, 2, 3, 5, 6, 7, 8, 9, 13, 60, 255)
    for mid in sorted(mercs):
        for mode in (2, 15):
            for s in sums:
                for v in vels:
                    rows.append((2, mid, mode, v, s - 100, rng.choice((0, 1))))
    for name, sk in pets:
        for mode in (2, 15):
            for s in sums:
                for v in (1, 3, 6, 9, 13, 60, 255):
                    pace = rng.choice(PACES)
                    move, speed, buff = rng.choice(((s - 100, 0, 0), (0, s - 100, 0), (0, 0, s - 100), (pace, s - 160 - pace, 60)))
                    rows.append((3, name, sk, mode, v, move, speed, buff, rng.choice((0, 1))))
    n += len(rows)
    while len(rows) < n:
        v = rng.choice((rng.randint(1, 15), rng.randint(1, 60), rng.randint(1, 255)))
        mode = rng.choice((2, 15))
        if rng.random() < 0.5:
            move = rng.choice((rng.choice(PACES), rng.randint(-200, 200), rng.randint(-30000, 29900)))
            rows.append((2, rng.choice(sorted(mercs)), mode, v, move, rng.choice((0, 1))))
        else:
            name, sk = rng.choice(pets)
            move = rng.choice(PACES + (rng.randint(-150, 150),))
            speed = rng.choice((0, 0, rng.randint(-100, 100), rng.randint(0, 300)))
            buff = rng.choice((0, 0, rng.randint(1, 200), rng.randint(-100, 100), rng.randint(-20000, 20000)))
            if 100 + move + speed + buff > 30100:
                buff = 30000 - move - speed
            rows.append((3, name, sk, mode, v, move, speed, buff, rng.choice((0, 1))))
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


def tdiv(a, b):
    """C / (x86 idiv): truncates toward zero."""
    q = abs(a) // abs(b)
    return q if (a >= 0) == (b > 0) else -q


def game(path):
    mon, mercs, _ = tables()
    from emu import Emu
    emu = Emu()
    nmon = max(mon.values()) + 1
    dt = emu.alloc(0x1000)
    monstats = emu.alloc(0x1a8 * nmon)
    emu.w32(0x744304, dt)
    emu.w32(dt + 0xa78, monstats)
    emu.w32(dt + 0xa80, nmon)
    unit = emu.alloc(0x200)
    path_ = emu.alloc(0x100)
    st, got = {}, {}
    emu.hook(0x620f00, lambda e: 0, 4)
    emu.hook(0x6213d0, lambda e: 256, 0)
    emu.hook(0x625480, lambda e: st.get(e.arg(1), 0) & 0xFFFFFFFF, 3)
    emu.hook(0x625500, lambda e: st.get(e.arg(1), 0) & 0xFFFFFFFF, 3)

    def setvel(e):
        got["vel"] = struct.unpack("<i", struct.pack("<I", e.arg(1)))[0]
        return 0

    emu.hook(0x648690, setvel, 4)
    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, kind = int(f[0]), int(f[1])
        if kind == 2:
            mid, mode, vel, move, flag = map(int, f[2:7])
            cls = mercs[mid]
            s43 = 100 + move
        else:
            name, _, mode, vel, move, speed, buff, flag = f[2], *map(int, f[3:10])
            cls = mon[name]
            s43 = 100 + move + speed + buff
        row = monstats + cls * 0x1a8
        emu.mu.mem_write(row + 0x32, struct.pack("<h", vel))
        emu.mu.mem_write(row + 0xd, bytes([flag]))
        for off in range(0, 0x200, 4):
            emu.w32(unit + off, 0)
        emu.w32(unit, 1)
        emu.w32(unit + 4, cls)
        emu.w32(unit + 0x10, mode)
        emu.w32(unit + 0x2c, path_)
        st.clear()
        st[0x43], st[96] = s43, 0
        got.clear()
        emu.call(0x623f50, unit, 0, 0)
        v = got.get("vel")
        pct = max(s43, 25)                      # the rule as read (stat 96 is 0 here)
        ok = v == tdiv((vel << 8) * pct, 100)
        print(cid, pct, v, "" if ok else "VELFAIL")


def compare(cases, a, b):
    rows = {ln.split()[0]: ln.split() for ln in open(cases) if ln.strip()}
    ga = {ln.split()[0]: ln.split() for ln in open(a) if ln.strip()}
    gb = {ln.split()[0]: ln.split() for ln in open(b) if ln.strip()}
    velfail = sum(1 for v in ga.values() if "VELFAIL" in v or v[2] == "None")
    from collections import Counter
    skipped = Counter(" ".join(v[1:]) for v in gb.values() if v[1] == "skip")
    bad, per, seen = [], Counter(), Counter()
    for k, r in rows.items():
        if k not in gb or gb[k][1] == "skip" or k not in ga:
            continue
        key = ("merc " if r[1] == "2" else "pet ") + ("RN" if (r[3] if r[1] == "2" else r[4]) == "15" else "WL")
        seen[key] += 1
        g, d = int(ga[k][1]), float(gb[k][1])
        if abs(g - d) > 0.01 + 1e-5 * abs(g):        # d2d steps in float: 4e-6 relative seen at walk % 12,599 and up
            bad.append(k)
            per[key] += 1
    print(f"cases {len(rows)} compared {sum(seen.values())} {dict(seen)} skipped {sum(skipped.values())} "
          f"mismatches {len(bad)} {dict(per)} velfail {velfail}")
    for k in bad[:10]:
        print("  ", " ".join(rows[k]), "game", ga[k][1], "d2d", gb[k][1])
    print("  skip reasons", skipped.most_common(5))


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        compare(*sys.argv[2:5])
    else:
        sys.exit(__doc__)
