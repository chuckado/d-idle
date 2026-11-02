# SPDX-License-Identifier: GPL-3.0-or-later
"""verifywaypoint: game.exe's 0x49 refusal (run natively through wp_refuse.Oracle) against two rules.

    tools/emu/run-on-worker.sh --ttl 3000 wpv_cmp.py author 7 100000     # the trace's grid + 100000 random, seed 7
    tools/emu/run-on-worker.sh --ttl 3000 wpv_cmp.py own 11 100000       # my own generator, seed 11

Rules compared per case (now, active, level-change times, hostile times):
  R_game  = the Oracle (game.exe code)
  R_d2d   = not active, or (now - latest level change) mod 2^32 < 10000
  R_hostile = not active, or (now - latest hostile declaration) mod 2^32 < 10000   (my reading of game.exe)
Prints the count of cases and the disagreement counts of each rule vs the Oracle,
split by direction, and a hash of the case list.
"""
import hashlib
import io
import random
import sys
from contextlib import redirect_stdout

import wp_refuse


def own_cases(n, seed):
    rng = random.Random(seed)
    cases = []
    for _ in range(n):
        now = rng.randrange(70_000, 1 << 31)
        nl = rng.randrange(1, 5)
        levels = sorted(now - rng.choice((rng.randrange(0, 15000), rng.randrange(0, 900000))) for _ in range(nl))
        hostiles = sorted(now - rng.choice((rng.randrange(0, 15000), rng.randrange(0, 90000))) for _ in range(rng.choice((0, 1, 1, 2, 3))))
        cases.append((now, int(rng.random() < 0.85), levels, hostiles))
    return cases


def author_cases(n, seed):
    buf = io.StringIO()
    with redirect_stdout(buf):
        wp_refuse.gen(n, seed, False)
    out = []
    for line in buf.getvalue().splitlines():
        if line.strip():
            _i, now, active, levels, hostiles = wp_refuse.parse(line)
            out.append((now, active, levels, hostiles))
    return out


def main():
    mode, seed, n = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
    cases = author_cases(n, seed) if mode == "author" else own_cases(n, seed)
    h = hashlib.sha256(repr(cases).encode()).hexdigest()[:16]
    o = wp_refuse.Oracle()
    M = 1 << 32
    cnt = dict(total=0, d2d_ref_game_go=0, game_ref_d2d_go=0, hos_ref_game_go=0, game_ref_hos_go=0, game_refused=0)
    for now, active, levels, hostiles in cases:
        g = o.refused(now, active, hostiles)
        d = 1 if (not active or (now - max(levels)) % M < 10000) else 0
        hr = 1 if (not active or (hostiles and (now - max(hostiles)) % M < 10000)) else 0
        cnt["total"] += 1
        cnt["game_refused"] += g
        cnt["d2d_ref_game_go"] += d == 1 and g == 0
        cnt["game_ref_d2d_go"] += g == 1 and d == 0
        cnt["hos_ref_game_go"] += hr == 1 and g == 0
        cnt["game_ref_hos_go"] += g == 1 and hr == 0
    print(mode, "seed", seed, "n", n, "casehash", h, cnt)
    print("d2d differs:", cnt["d2d_ref_game_go"] + cnt["game_ref_d2d_go"], "hostile-rule differs:", cnt["hos_ref_game_go"] + cnt["game_ref_hos_go"])


main()
