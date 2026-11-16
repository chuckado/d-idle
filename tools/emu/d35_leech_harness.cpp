// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's life / mana leech of a player's melee blow on tools/emu/d35_leech.py's cases: the REAL d2d::rules::player_blow
// (no copy of its lines). The weapon's physical is the case's `phys` (256ths, a fixed roll), the target has no
// resistance, defense 0, no block; the case's seed is advanced until the blow hits. Prints `id life256 mana256 phys256`
// with life256 = what Fight::land credits (`blow.life << 8`, strike.cpp), against a target holding `life` (256ths).
//
//   g++ -O2 -std=c++20 -I<d2d>/components/rules -I<d2d>/components -I<dir of d2s_items.hpp> -o d35_leech_harness d35_leech_harness.cpp
//   ./d35_leech_harness < cases.tsv > d2d.tsv
#include "combat.hpp"
#include <cstdio>

int main() {
    char line[512];
    while (std::fgets(line, sizeof line, stdin)) {
        long id, diff, sl, sm, drain, phys, life;
        if (line[0] == '#' || std::sscanf(line, "%ld %ld %ld %ld %ld %ld %ld", &id, &diff, &sl, &sm, &drain, &phys, &life) != 7) continue;
        d2d::rules::Fighter fighter;
        fighter.phys_lo = fighter.phys_hi = int(phys);
        fighter.ar_base = 1 << 20;
        fighter.life_steal = int(sl);
        fighter.mana_steal = int(sm);
        d2d::rules::Target target;
        target.hit_points = int(life >> 8);
        target.max_hp = int(life >> 8);
        target.drain = int(drain);
        d2d::rules::Blow blow;
        for (std::uint32_t seed = std::uint32_t(id) * 7919u + 1; !blow.hit; ++seed) {
            d2d::rules::Rng rng{ seed };
            blow = d2d::rules::player_blow(fighter, target, 30, rng);
        }
        std::printf("%ld %d %d %d\n", id, blow.life * 256, blow.mana * 256, blow.phys_fine);
    }
}
