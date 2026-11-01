# SPDX-License-Identifier: GPL-3.0-or-later
"""Monte Carlo: does a Lightning / Chain Lightning missile re-hit a monster?

Usage: python3 sk_light_rehit.py [N] [seed]   (default N=20000, seed 1)

The d2d parts (kMissileAim, missile_aim, MissileFlight launch/step/walk with
an open room, missile_touches) are line copies of upstream d2d main 72132b99
components/rules/missiles.hpp (the `points` path branch is left out).
The game.exe rule is a MODEL of the reading of FUN_005ae1f0 (one unit hit a
frame), FUN_0064a470 (skips the last unit hit) and FUN_005adb00 (state 0x56
justhit for NextDelay frames). Tag I (inferred), not a proof.

Scenario A: Lightning pass (Vel 30 -> path velocity 5760, Size 2, Range 25,
CollideKill 0, LastCollide 1): rule G (one hit a frame, never the last unit)
against rule D (every touching monster once, at most once each).
Scenario B: Chain Lightning hop (NextHit 1, NextDelay 4): is the first hit of
the hop the monster just struck (A)? D: yes if it touches first; G: A is
immune for NextDelay frames (window 4, and 3 to show the edge).
"""
import math
import random
import sys

KMISSILE_AIM = (
    (0, 4096), (32, 4095), (64, 4095), (96, 4094), (128, 4093), (161, 4092), (193, 4091), (225, 4089),
    (257, 4087), (289, 4085), (321, 4083), (353, 4080), (385, 4077), (417, 4074), (448, 4071), (480, 4067),
    (511, 4063), (543, 4059), (574, 4055), (606, 4050), (637, 4046), (668, 4041), (699, 4035), (729, 4030),
    (760, 4024), (791, 4018), (821, 4012), (851, 4006), (881, 3999), (911, 3993), (941, 3986), (971, 3979),
    (1000, 3971), (1030, 3964), (1059, 3956), (1088, 3948), (1117, 3940), (1145, 3932), (1174, 3924), (1202, 3915),
    (1230, 3906), (1258, 3897), (1286, 3888), (1313, 3879), (1340, 3870), (1368, 3860), (1394, 3851), (1421, 3841),
    (1448, 3831), (1474, 3821), (1500, 3811), (1526, 3800), (1552, 3790), (1577, 3780), (1602, 3769), (1627, 3758),
    (1652, 3747), (1677, 3736), (1701, 3725), (1725, 3714), (1749, 3703), (1773, 3692), (1796, 3680), (1820, 3669),
    (1843, 3657), (1866, 3646), (1888, 3634), (1911, 3622), (1933, 3610), (1955, 3599), (1977, 3587), (1998, 3575),
    (2020, 3563), (2041, 3551), (2062, 3539), (2082, 3526), (2103, 3514), (2123, 3502), (2143, 3490), (2163, 3478),
    (2183, 3465), (2202, 3453), (2221, 3441), (2240, 3428), (2259, 3416), (2278, 3403), (2296, 3391), (2314, 3379),
    (2332, 3366), (2350, 3354), (2368, 3341), (2385, 3329), (2402, 3317), (2419, 3304), (2436, 3292), (2453, 3279),
    (2469, 3267), (2486, 3255), (2502, 3242), (2518, 3230), (2533, 3218), (2549, 3205), (2564, 3193), (2580, 3181),
    (2595, 3169), (2609, 3156), (2624, 3144), (2639, 3132), (2653, 3120), (2667, 3108), (2681, 3096), (2695, 3084),
    (2709, 3072), (2722, 3060), (2736, 3048), (2749, 3036), (2762, 3024), (2775, 3012), (2788, 3000), (2800, 2988),
    (2813, 2977), (2825, 2965), (2837, 2953), (2849, 2942), (2861, 2930), (2873, 2919), (2884, 2907), (2896, 2896),
)

M32 = 0xFFFFFFFF
VEL = ((30 << 8) * 75) // 100  # 5760
RANGE = 25
MSIZE = 2


def s32(v):
    v &= M32
    return v - (1 << 32) if v & 0x80000000 else v


def missile_aim(x, y, to_x, to_y):
    west, north = to_x < x, to_y < y
    across = (x - to_x if west else to_x - x) & M32
    down = (y - to_y if north else to_y - y) & M32
    steep = s32(across) <= s32(down)
    major, minor = (down, across) if steep else (across, down)
    row = 0 if major == 0 else int(s32((minor * 0x7F) & M32) / s32(major))  # C++ truncates
    small, big = KMISSILE_AIM[max(0, min(127, row))]
    ax, ay = (small, big) if steep else (big, small)
    return (-ax if west else ax), (-ay if north else ay)


def missile_touches(x, y, ux, uy, size, missile_size=1):
    if missile_size < 1 or missile_size > 3 or size < 1:
        return False
    a, d = abs(x - ux), abs(y - uy)
    c = min(size, 3) - 1 + (missile_size - 1) * 3
    if c == 0:
        return a == 0 and d == 0
    if c in (1, 3):
        return a + d <= 1
    if c in (2, 6):
        return a <= 1 and d <= 1
    if c == 4:
        return a + d <= 2
    if c in (5, 7):
        return (a < 3 and d < 2) or (d < 3 and a <= 1)
    return a <= 2 and d <= 2


class MissileFlight:
    def __init__(self, x, y, to_x, to_y, velocity=VEL, max_velocity=VEL, accel=0):
        if to_x == x and to_y == y:
            to_x += 1
            to_y += 1
        self.x = ((x << 16) + 0x8000) & M32
        self.y = ((y << 16) + 0x8000) & M32
        self.aim_x, self.aim_y = missile_aim(
            self.x, self.y, ((to_x << 16) + 0x8000) & M32, ((to_y << 16) + 0x8000) & M32)
        self.velocity, self.max_velocity, self.accel = velocity, max_velocity, accel
        self.accel_frames = 0
        self.crossed = []

    def sub(self):
        return self.x >> 16, self.y >> 16

    def step(self):  # open room: never blocked
        self.crossed = []
        if self.accel != 0:
            self.accel_frames += 1
            if self.accel_frames > 4:
                self.velocity += self.accel
                if self.velocity > self.max_velocity:
                    self.velocity, self.accel = self.max_velocity, 0
                elif self.velocity < 0:
                    self.velocity = 0
                self.accel_frames = 0
        speed = s32((self.velocity * 0x400) & M32) >> 6
        sx = s32((self.aim_x * speed) & M32) >> 12
        sy = s32((self.aim_y * speed) & M32) >> 12
        if sx == 0 and sy == 0:
            self.x = ((self.x & 0xFFFF0000) + 0x8000) & M32
            self.y = ((self.y & 0xFFFF0000) + 0x8000) & M32
            return False
        self._walk(sx, sy)
        self.x = (self.x + sx) & M32
        self.y = (self.y + sy) & M32
        return True

    def _walk(self, sx, sy):
        px, py = sx, sy
        while px > 0x10000 or px < -0x10000:
            px >>= 1
            py >>= 1
        while py > 0x10000 or py < -0x10000:
            px >>= 1
            py >>= 1
        end_x = ((self.x + sx) & M32) >> 16 & 0xFFFF
        end_y = ((self.y + sy) & M32) >> 16 & 0xFFFF
        ax, ay, hx, hy = self.x, self.y, self.x >> 16, self.y >> 16
        while hx != end_x or hy != end_y:
            nx, ny = (ax + px) & M32, (ay + py) & M32
            tx, ty = nx >> 16, ny >> 16
            if tx != hx or ty != hy:
                self.crossed.append((tx, ty))
                if len(self.crossed) >= 10:
                    break
            ax, ay, hx, hy = nx, ny, tx, ty


CX = CY = 1000


def rnd_dir(rng):
    a = rng.uniform(0, 2 * math.pi)
    return math.cos(a), math.sin(a)


def cast_A(rng, n=None, offset=None, sizes=None):
    """Return (hits_G, hits_D) lists for one cast."""
    n = n or rng.randint(2, 6)
    dx, dy = rnd_dir(rng)
    px, py = -dy, dx
    mons, seen = [], set()
    while len(mons) < n:
        t = rng.uniform(3, 30)
        o = rng.uniform(-3, 3) if offset is None else offset
        pos = (round(CX + dx * t + px * o), round(CY + dy * t + py * o))
        if pos in seen:
            continue
        seen.add(pos)
        mons.append((pos[0], pos[1], sizes or rng.randint(1, 3)))
    fl = MissileFlight(CX, CY, round(CX + dx * 40), round(CY + dy * 40))
    hg, hd, last, struck = [0] * n, [0] * n, -1, set()
    for _ in range(RANGE):
        fl.step()
        done = False
        for sx, sy in fl.crossed:
            for m, (mx, my, ms) in enumerate(mons):
                if m != last and missile_touches(sx, sy, mx, my, ms, MSIZE):
                    hg[m] += 1
                    last = m
                    done = True
                    break
            if done:
                break
        for m, (mx, my, ms) in enumerate(mons):
            if m not in struck and any(missile_touches(sx, sy, mx, my, ms, MSIZE) for sx, sy in fl.crossed):
                hd[m] += 1
                struck.add(m)
    return hg, hd


def scenario_A(N, rng):
    print("== Scenario A: Lightning pass (N=%d) ==" % N)
    st = {k: dict(c=0, hg=0, hd=0, diff=0, mons=0, g2=0, d_not_g=0, mx=0) for k in range(2, 7)}
    for _ in range(N):
        hg, hd = cast_A(rng)
        s = st[len(hg)]
        s["c"] += 1
        s["hg"] += sum(hg)
        s["hd"] += sum(hd)
        s["diff"] += hg != hd
        s["mons"] += len(hg)
        s["g2"] += sum(h >= 2 for h in hg)
        s["d_not_g"] += sum(d > 0 and g == 0 for g, d in zip(hg, hd))
        s["mx"] = max(s["mx"], max(hg))
    tot = {k: sum(st[n][k] for n in st) for k in ("c", "hg", "hd", "diff", "mons", "g2", "d_not_g")}
    tot["mx"] = max(st[n]["mx"] for n in st)
    print("%-5s %7s %9s %9s %9s %10s %12s %6s" % (
        "n", "casts", "meanG", "meanD", "G!=D %", "mon G>=2 %", "D>0,G=0 %", "maxG"))
    for k, s in list(st.items()) + [("all", tot)]:
        if not s["c"]:
            continue
        print("%-5s %7d %9.3f %9.3f %9.2f %10.2f %12.2f %6d" % (
            k, s["c"], s["hg"] / s["c"], s["hd"] / s["c"], 100 * s["diff"] / s["c"],
            100 * s["g2"] / s["mons"], 100 * s["d_not_g"] / s["mons"], s["mx"]))
    for sz in (1, 2, 3):  # sanity: lone monster on the line is hit once by both
        for _ in range(300):
            hg, hd = cast_A(rng, n=1, offset=0, sizes=sz)
            assert hg == [1] and hd == [1], (sz, hg, hd)
    print("sanity: lone monster on the line hit exactly once by G and D (sizes 1..3, 300 casts each)")
    print()


def angle_bin(ax, ay, bx, by):
    if (ax, ay) == (0, 0) or (bx, by) == (0, 0):
        return "n/a"
    a = abs(math.degrees(math.atan2(ax * by - ay * bx, ax * bx + ay * by)))
    return "%3d-%3d" % (min(int(a // 45), 3) * 45, min(int(a // 45), 3) * 45 + 45)


def scenario_B(N, rng):
    print("== Scenario B: Chain Lightning hop (N=%d) ==" % N)
    rows = []
    for _ in range(N):
        sA, sB = rng.randint(1, 3), rng.randint(1, 3)
        dx, dy = rnd_dir(rng)
        t = rng.uniform(3, 30)
        A = (round(CX + dx * t), round(CY + dy * t))
        if A == (CX, CY):
            continue
        fl = MissileFlight(CX, CY, A[0], A[1])
        spot = None
        for _f in range(RANGE):
            fl.step()
            if any(missile_touches(sx, sy, A[0], A[1], sA, MSIZE) for sx, sy in fl.crossed):
                spot = fl.sub()
                break
        if spot is None:
            continue
        while True:
            bx, by = rng.randint(-20, 20), rng.randint(-20, 20)
            B = (spot[0] + bx, spot[1] + by)
            if bx * bx + by * by <= 400 and B != A:
                break
        fh = MissileFlight(spot[0], spot[1], B[0], B[1])
        first = {"D": None, "G4": None, "G3": None}
        dframe = None
        for k in range(1, RANGE + 1):
            fh.step()
            ta = any(missile_touches(sx, sy, A[0], A[1], sA, MSIZE) for sx, sy in fh.crossed)
            tb = any(missile_touches(sx, sy, B[0], B[1], sB, MSIZE) for sx, sy in fh.crossed)
            for key, win in (("D", 0), ("G4", 4), ("G3", 3)):
                if first[key] is None:
                    if ta and k > win:
                        first[key] = "A"
                        if key == "D":
                            dframe = k
                    elif tb:
                        first[key] = "B"
            if all(v for v in first.values()):
                break
        ang = angle_bin(A[0] - spot[0], A[1] - spot[1], B[0] - spot[0], B[1] - spot[1])
        rows.append((sA, ang, first, dframe))
    def table(title, keyf):
        groups = {}
        for r in rows:
            groups.setdefault(keyf(r), []).append(r)
        print(title)
        print("%-9s %7s %8s %8s %8s %8s %8s" % ("group", "hops", "A|D %", "A|G4 %", "A|G3 %", "none|D %", "D frame"))
        for g in sorted(groups):
            rs = groups[g]
            c = len(rs)
            f = lambda key, v: 100 * sum(r[2][key] == v for r in rs) / c
            fr = [r[3] for r in rs if r[3]]
            print("%-9s %7d %8.2f %8.2f %8.2f %8.2f %8.2f" % (
                g, c, f("D", "A"), f("G4", "A"), f("G3", "A"), f("D", None), sum(fr) / len(fr) if fr else 0))
    table("all hops", lambda r: "all")
    table("by size of A", lambda r: "sA=%d" % r[0])
    table("by angle A-spot-B (deg)", lambda r: r[1])
    print()


def main():
    N = int(sys.argv[1]) if len(sys.argv) > 1 else 20000
    seed = int(sys.argv[2]) if len(sys.argv) > 2 else 1
    assert missile_touches(0, 0, 2, 0, 2, 2) and not missile_touches(0, 0, 3, 0, 2, 2)
    print("sanity: missile_touches(0,0,2,0,2,2) True, (0,0,3,0,2,2) False; seed=%d" % seed)
    rng = random.Random(seed)
    scenario_A(N, rng)
    scenario_B(N, rng)


if __name__ == "__main__":
    main()
