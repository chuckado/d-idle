# SPDX-License-Identifier: GPL-3.0-or-later
"""Frost Nova's ring: game.exe's own code (FUN_0056d400, FUN_0059fa30, the
missile's srvdofunc 1) against d2d's nova (components/game/fight/missiles.cpp
`Fight::fire`, do 22: 64 even angles, a way not a point).

    python sk_nova_ring.py [spots] [seed] [monsters]

Runs from upstream's tools/emu (imports missiles.py, drlg.py, monsters.py;
needs D2_MPQ_DIR, D2_PATCH_INSTALLER, D2_BUILD like missiles.py).

A. FUN_0056d400 called natively for row 119 (frostnova), FUN_0059fa30
   hooked to record the 64 params blocks it builds: count, flags, start,
   offsets (the tables at 0x6e1288 / 0x6e1388), the velocity override.
B. For free spots in a Blood Moor game (missiles.py's Game), the 64 ring
   missiles made by the REAL FUN_0059fa30 from params built as FUN_0056d400
   builds them, each stepped frame by frame by its real srvdofunc (no foe:
   FUN_00554200 hooked to "hostile to nobody"). Each frame's 16.16 position
   is compared with (a) missiles.py's port fed the table offset (the reading)
   and (b) the same port aimed as d2d aims a nova missile:
   missile_aim_along(cos, sin) of the k x 2 pi / 64 angle, and follow set
   (d2d: `follow = !aimed`).
C. A MODEL (tag I, not game code): monsters of size 1..3 at random subtiles
   within 13 of the caster, the game.exe flights of B, and two hit rules:
   game.exe (FUN_005ae1f0 / FUN_005adf10 / FUN_005adb00 as read: one unit a
   missile a frame, the missile's last unit skipped, state 0x56 `justhit`
   on the unit for NextDelay 4 frames, skipped by every NextHit missile)
   and d2d (`skill_missile_hits`: per-missile `hit_at`, every touching
   monster in the frame). Prints hits per monster per cast by distance.
"""
import math
import random
import struct
import sys

import missiles

ROW, SKILL = 119, 44            # frostnova, Frost Nova
NEXT_DELAY = 4


def aim_along(table, dx, dy):
    """rules::missile_aim_along (components/rules/missiles.hpp), line for line."""
    across, down = abs(dx), abs(dy)
    steep = across <= down
    major, minor = (down, across) if steep else (across, down)
    row = 0 if major == 0 else max(0, min(127, int(minor * 127 / major + 1e-4)))
    small, big = table[row]
    ax, ay = (small, big) if steep else (big, small)
    return (-ax if dx < 0 else ax), (-ay if dy < 0 else ay)


def f32(v):
    return struct.unpack("<f", struct.pack("<f", v))[0]


def d2d_dirs():
    """Fight::fire's nova: angle = float(k) * 2 * 3.14159265f / 64 in floats; cos / sin."""
    out = []
    for k in range(64):
        a = f32(f32(f32(float(k) * 2) * f32(3.14159265)) / 64)
        out.append((f32(math.cos(a)), f32(math.sin(a))))
    return out


def free_disc(game, x, y, r):
    for dx in range(-r, r + 1):
        for dy in range(-r, r + 1):
            if dx * dx + dy * dy <= r * r:
                c = game.collision(x + dx, y + dy)
                if c == 0xFFFF or c & 0x1f0f: return False
    return True


def record_params(game, src, sx, sy, lvl):
    """A: FUN_0056d400 natively with FUN_0059fa30 hooked to read each params block."""
    e = game.e
    seen = []
    # The hook sees edx = params (fastcall); read it from the register.
    from unicorn.x86_const import UC_X86_REG_EDX

    def on_make2(e):
        p = e.mu.reg_read(UC_X86_REG_EDX)
        seen.append({"flags": e.r32(p), "owner": e.r32(p + 4), "row": e.r32(p + 0x10), "x": e.s32(p + 0x14), "y": e.s32(p + 0x18),
                     "dx": e.s32(p + 0x1c), "dy": e.s32(p + 0x20), "vel": e.s32(p + 0x28), "skill": e.r32(p + 0x2c), "lvl": e.r32(p + 0x30)})
        return 0
    e.hook(0x59fa30, on_make2, 0)
    vel = e.call(0x663270, ROW, lvl)
    e.call(0x56d400, ROW, SKILL, lvl, vel, ecx=game.g, edx=src)
    return seen, vel


def ring_game(game, src, sx, sy, lvl, offs, limit=40):
    """B: the 64 missiles by the real FUN_0059fa30, stepped by their srvdofunc."""
    e = game.e
    vel = e.call(0x663270, ROW, lvl)
    fn = e.r32(0x73c768 + 4 * game.row(ROW, 0xc, 2))
    out = []
    for k in range(64):
        p = e.alloc(0x60)
        for i in range(0x18): e.w32(p + 4 * i, 0)
        fields = {0: 3 | 4, 1: src, 4: ROW, 5: sx, 6: sy, 7: offs[k][0] & 0xFFFFFFFF, 8: offs[k][1] & 0xFFFFFFFF, 0xa: vel, 0xb: SKILL, 0xc: lvl}
        for i, v in fields.items(): e.w32(p + 4 * i, v)
        m = e.call(0x59fa30, ecx=game.g, edx=p)
        if not m:
            out.append(None)
            continue
        path = e.r32(m + 0x2c)
        rng = e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16
        track = []
        for _ in range(limit):
            if e.s32(e.r32(m + 0x14) + 0x10) << 16 >> 16 < 1: break
            e.call(fn, ecx=game.g, edx=m)
            track.append((e.r32(path), e.r32(path + 4)))
        out.append((rng, track))
    return out, vel


def port_track(game, rooms, sx, sy, to, vel, rng, aim=None, follow=None):
    fl = missiles.Flight(game.table, sx, sy, to[0], to[1], vel, 24 << 8, -1000, None, rooms)
    if aim is not None: fl.dx, fl.dy = aim
    if follow is not None: fl.follow = follow
    blocked = lambda x, y: bool(game.collision(x, y) & 4) or game.collision(x, y) == 0xFFFF
    track, crossed = [], []
    for _ in range(rng):
        moved, entered = fl.step(blocked)
        track.append((fl.x, fl.y))
        crossed.append(entered)
        if not moved: break
    return track, crossed


def crossed_of(track, sx, sy):
    """The subtiles each frame entered, rebuilt from the 16.16 track (walk order)."""
    out, x, y = [], (sx << 16) + 0x8000, (sy << 16) + 0x8000
    for nx, ny in track:
        sub_x, sub_y = nx - x, ny - y
        sub_x = missiles.s32(sub_x); sub_y = missiles.s32(sub_y)
        while not -0x10000 <= sub_x <= 0x10000: sub_x >>= 1; sub_y >>= 1
        while not -0x10000 <= sub_y <= 0x10000: sub_x >>= 1; sub_y >>= 1
        here, end, ent = (x >> 16, y >> 16), (nx >> 16, ny >> 16), []
        ax, ay = x, y
        guard = 0
        while here != end and guard < 12:
            ax, ay = ax + sub_x, ay + sub_y
            there = (ax >> 16, ay >> 16)
            if there != here: ent.append(there)
            here = there
            guard += 1
        out.append(ent)
        x, y = nx, ny
    return out


def hits_game(flights, foes, window=NEXT_DELAY):
    """Model of the game.exe rule (I): per frame, missiles in creation order."""
    n = len(foes)
    hits, just_until, last = [0] * n, [-1] * n, [-1] * len(flights)
    frames = max(len(c) for c in flights)
    for f in range(frames):
        for k, cr in enumerate(flights):
            if f >= len(cr) - 1: continue           # the frame its range ends: no unit test (S, SKMISS)
            done = False
            for spot in cr[f]:
                for j, (ux, uy, size) in enumerate(foes):
                    if not missiles.touches(*spot, ux, uy, size): continue
                    if just_until[j] > f or last[k] == j: continue
                    hits[j] += 1; just_until[j] = f + window; last[k] = j
                    done = True
                    break
                if done: break
    return hits


def hits_d2d(flights, foes, window=NEXT_DELAY):
    """Model of d2d's rule (S read, missiles.cpp:584-596, tick.cpp:394-411)."""
    n = len(foes)
    hits = [0] * n
    frames = max(len(c) for c in flights)
    hit_at = [dict() for _ in flights]
    for f in range(frames):
        for k, cr in enumerate(flights):
            if f >= len(cr) - 1: continue           # same last-frame rule as above, to isolate the NextHit rule
            for j, (ux, uy, size) in enumerate(foes):
                if not any(missiles.touches(*spot, ux, uy, size) for spot in cr[f]): continue
                if j in hit_at[k] and f < hit_at[k][j] + window: continue
                hit_at[k][j] = f
                hits[j] += 1
    return hits


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    spots = int(args[0]) if args else 4
    rng = random.Random(int(args[1]) if len(args) > 1 else 1)
    nmon = int(args[2]) if len(args) > 2 else 4000
    game = missiles.Game()
    e = game.e
    offs = [(e.s32(0x6e1288 + 4 * k), e.s32(0x6e1388 + 4 * k)) for k in range(64)]
    print("table", offs)
    dirs = d2d_dirs()
    found, tries = [], 0
    while len(found) < spots and tries < 20000:
        tries += 1
        room = rng.choice(game.rooms)
        x, y = room[1] + rng.randrange(room[3]), room[2] + rng.randrange(room[4])
        if free_disc(game, x, y, 14): found.append((x, y))
    print(f"spots {found}")
    all_game_flights = []
    total = {"a": 0, "b": 0, "bframes": 0, "frames": 0, "subtiles_b": 0, "flights": 0}
    rows_diff = []
    for (x, y) in found:
        u = game.unit(63, x, y)
        if not u: continue
        src, sx, sy, _ = u
        rooms = missiles.Rooms(game.rect, game.near, e.call(0x620bb0, src))
        for lvl in (1, 20):
            got, vel = ring_game(game, src, sx, sy, lvl, offs)
            pvel = missiles.missile_velocity(24, 0, lvl)
            for k in range(64):
                if got[k] is None: continue
                rng_g, track = got[k]
                total["flights"] += 1
                ta, _ = port_track(game, rooms, sx, sy, (sx + offs[k][0], sy + offs[k][1]), pvel, rng_g)
                ax, ay = aim_along(game.table, *dirs[k])
                tb, _ = port_track(game, rooms, sx, sy, (sx + offs[k][0], sy + offs[k][1]), pvel, rng_g, aim=(ax, ay), follow=True)
                if ta[:len(track)] != track: total["a"] += 1
                if tb[:len(track)] != track:
                    total["b"] += 1
                    total["bframes"] += sum(1 for p, q in zip(track, tb) if p != q)
                    sg = set(s for c in crossed_of(track, sx, sy) for s in c)
                    sb = set(s for c in crossed_of(tb, sx, sy) for s in c)
                    total["subtiles_b"] += len(sg ^ sb)
                    if lvl == 1 and (x, y) == found[0]:
                        ga = missiles.aim(game.table, 0, 0, offs[k][0] << 16, offs[k][1] << 16)
                        rows_diff.append((k, ga, (ax, ay), (track[-1][0] - ((sx << 16) + 0x8000)) / 65536, (track[-1][1] - ((sy << 16) + 0x8000)) / 65536,
                                          (tb[len(track) - 1][0] - ((sx << 16) + 0x8000)) / 65536, (tb[len(track) - 1][1] - ((sy << 16) + 0x8000)) / 65536))
                total["frames"] += len(track)
            if lvl == 1:
                all_game_flights.append((sx, sy, [crossed_of(g[1], sx, sy) if g else [] for g in got], [len(g[1]) if g else 0 for g in got],
                                         [g[1][-1] if g and g[1] else None for g in got]))
    print(f"B: {total['flights']} ring flights; reading port differs in {total['a']}; d2d aim differs in {total['b']} "
          f"({total['bframes']} of {total['frames']} frames, {total['subtiles_b']} subtiles entered by only one)")
    print("B first spot lvl 1, directions where d2d's aim differs: k, game unit vector, d2d unit vector, game end (dx, dy), d2d end (dx, dy)")
    for r in rows_diff: print("  ", r[0], r[1], r[2], f"{r[3]:.3f},{r[4]:.3f}", f"{r[5]:.3f},{r[6]:.3f}")
    if all_game_flights:
        sx, sy, cr, lens, ends = all_game_flights[0]
        reach = [math.hypot((p[0] >> 16) - sx, (p[1] >> 16) - sy) for p in ends if p]
        print(f"B frames per flight {sorted(set(lens))}; end distance (subtiles) min {min(reach):.2f} max {max(reach):.2f}")
    # A: the params FUN_0056d400 builds (hook installed last: B needs the real FUN_0059fa30)
    x, y = found[0]
    u = game.unit(63, x + 1, y)
    if u:
        seen, vel = record_params(game, u[0], u[1], u[2], 1)
        same = all(s["dx"] == offs[i][0] and s["dy"] == offs[i][1] for i, s in enumerate(seen))
        print(f"A: FUN_0056d400 made {len(seen)} calls; flags {sorted(set(s['flags'] for s in seen))}; vel arg {sorted(set(s['vel'] for s in seen))} "
              f"(FUN_00663270 {vel}); offsets = tables: {same}; start = caster: {all((s['x'], s['y']) == (u[1], u[2]) for s in seen)}; "
              f"skill {sorted(set(s['skill'] for s in seen))} lvl {sorted(set(s['lvl'] for s in seen))} row {sorted(set(s['row'] for s in seen))}")
    # C: the model
    if not all_game_flights: return
    sx, sy, cr, _, _ = all_game_flights[0]
    flights = [c for c in cr if c]
    for window in (NEXT_DELAY, NEXT_DELAY - 1):
        by = {}
        for _ in range(nmon):
            pack = rng.choice((1, 1, 3, 6))
            foes = []
            while len(foes) < pack:
                d, a = rng.uniform(0.5, 13), rng.uniform(0, 2 * math.pi)
                fx, fy = sx + round(d * math.cos(a)), sy + round(d * math.sin(a))
                if (fx, fy) == (sx, sy) or any((fx, fy) == (f[0], f[1]) for f in foes): continue
                foes.append((fx, fy, rng.choice((1, 2, 3))))
            hg, hd = hits_game(flights, foes, window), hits_d2d(flights, foes, window)
            for j, (fx, fy, size) in enumerate(foes):
                band = min(int(math.hypot(fx - sx, fy - sy)), 13)
                key = (band, size, pack > 1)
                s = by.setdefault(key, [0, 0, 0, 0, 0])
                s[0] += 1; s[1] += hg[j]; s[2] += hd[j]; s[3] += hg[j] > 1; s[4] += hg[j] == 0 and hd[j] > 0
        print(f"C window {window}: distance band, size, pack: monsters, hits/cast game.exe rule, d2d rule, game.exe 2+ hits, hit by d2d only")
        for key in sorted(by):
            n, g, d, two, only = by[key]
            print(f"  d{key[0]:>2} s{key[1]} {'pack' if key[2] else 'lone'} n={n:>5} game {g / n:.3f} d2d {d / n:.3f} game2+ {two / n:.3f} d2donly {only / n:.3f}")
        tot = [sum(v[i] for v in by.values()) for i in range(5)]
        print(f"  all n={tot[0]} game {tot[1] / tot[0]:.3f} d2d {tot[2] / tot[0]:.3f} ratio {tot[2] / max(tot[1], 1):.2f}")


if __name__ == "__main__":
    main()
