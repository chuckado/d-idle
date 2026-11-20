#!/usr/bin/env python3
"""dmg_model.py: expected damage to a standing clvl-6 hero from N adjacent Den of Evil Fallen (agent dmgcheck).

NOT an emulator run. It is a Monte Carlo of the rules in docs/research/damage-intake-2026-11-15.md
(tags there): hit chance (V run in docs/research/monster-attack-trace.md, equal in d2d), the Fallen think
cadence (aidel 15, aip3 50 %, A1 10 frames; D + V run), and an optional 'flee while a kin is in DT' term
(FUN_005f02c0, V read; the time off is U). Use: python3 -I tools/emu/dmg_model.py
"""
import random, sys

def hit_chance(ar, dlvl_def, mlvl, clvl):
    c = ar * 100 / (ar + dlvl_def) * 2 * mlvl / (mlvl + clvl)
    return min(95.0, max(5.0, c)) / 100

def run(n_adj, secs, p_hit, dmg, aip3, flee_p, off_s, rng):
    """frames at 25/s. think every 15 frames after the mode ends; swing 10 frames.
    flee_p: chance that a think finds a unit in DT within 15 subtiles (then it is away off_s seconds)."""
    total = 0; swings = 0; zero = 0
    for _ in range(n_adj):
        t = rng.random() * 15  # frames, phase
        end = secs * 25
        while t < end:
            if rng.random() < flee_p:
                t += off_s * 25 + 15
                continue
            if rng.random() < aip3:
                swings += 1
                if rng.random() < p_hit: total += rng.randint(*dmg)
                t += 10 + 15
            else:
                t += 10 + 15 if rng.random() < 0.3 else 10 + 15  # wander 3 / s2 taunt / stand 10, each then +aidel
    return total, swings

def main():
    rng = random.Random(1)
    clvl, defense = 6, 28
    print("hero clvl %d defence %d; Den Fallen mlvl/AR/dmg: L1 8 1-2, L2 12 1-3 (D, docs/archive/monster-attack-trace-2026-10-30.md)" % (clvl, defense))
    for mlvl, ar, dmg in ((1, 8, (1, 2)), (2, 12, (1, 3))):
        p = hit_chance(ar, defense, mlvl, clvl)
        print("\nFallen mlvl %d: hit chance %.1f %%, swing mean %.1f" % (mlvl, 100 * p, sum(dmg) / 2))
        print("N_adj flee_p off_s | hits/20s(mean) dmg/20s(mean) P(zero dmg in 20 s) dmg/s")
        for n in (3, 5, 8):
            for flee_p, off in ((0.0, 0), (0.2, 3), (0.4, 3), (0.6, 4)):
                trials = 4000; z = 0; tot = 0
                for _ in range(trials):
                    d, _s = run(n, 20, p, dmg, 0.5, flee_p, off, rng)
                    tot += d; z += (d == 0)
                print("%d %.1f %d | %.2f %.1f %.1f %%  %.2f" % (n, flee_p, off, tot / trials / (sum(dmg) / 2), tot / trials, 100 * z / trials, tot / trials / 20))

if __name__ == "__main__":
    main()
