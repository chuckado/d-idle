// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's Crushing Blow of a player's blow on tools/emu/d35_crush.py's cases: the REAL d2d::rules::player_blow (no copy of its lines).
// A case is `id life phys res div`: the target's life in 256ths (a whole number of points), the weapon's physical in 256ths
// (a fixed roll, before the target's physical resistance `res`), the target's Crushing Blow divisor (4, 8, 10). The weapon
// has Crushing Blow 100 %; the seed is advanced until the blow hits and crushes. Prints `id crush` in 256ths: the blow's
// damage in 256ths less its physical after resistance (what the crush added).
//
//   g++ -O2 -std=c++23 -I<d2d>/components/rules -I<d2d>/components -I<d2d>/components/d2s -I<d2d>/components/txt \
//       -o d35_crush_harness d35_crush_harness.cpp
//   ./d35_crush_harness < cases.tsv > d2d.tsv
#include "combat.hpp"
#include <cstdio>

int main() {
    char line[512];
    while (std::fgets(line, sizeof line, stdin)) {
        long id, life, phys, res, div;
        if (line[0] == '#' || std::sscanf(line, "%ld %ld %ld %ld %ld", &id, &life, &phys, &res, &div) != 5) continue;
        d2d::rules::Fighter fighter;
        fighter.phys_lo = fighter.phys_hi = int(phys);
        fighter.ar_base = 1 << 20;
        fighter.crushing = 100;
        d2d::rules::Target target;
        target.hit_points = int((life + 255) >> 8);
        target.max_hp = int((life + 255) >> 8);
        target.res[0] = int(res);
        target.crush_divisor = int(div);
        [&](auto& tgt) { if constexpr (requires { tgt.life_fine; }) tgt.life_fine = int(life); }(target);
        d2d::rules::Blow blow;
        for (std::uint32_t seed = std::uint32_t(id) * 7919u + 1; !blow.hit; ++seed) {
            d2d::rules::Rng rng{ seed };
            blow = d2d::rules::player_blow(fighter, target, 30, rng);
        }
        std::printf("%ld %d\n", id, blow.damage * 256 + blow.frac - blow.phys_fine);
    }
}
