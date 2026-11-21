// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's elemental or magic field of one monster blow on a player, on tools/emu/gear_absorb.py's cases: the REAL
// d2d::rules::monster_blow (no copy of its lines). The record is built without a roll (`exact`: one El row of min == max
// whole points, no physical), so the field is the case's `field`; the defender's resistance, flat magic reduction and
// (from the claim on) gear absorb come from the case. Prints `id fine total`: the field after the reduction, the resistance
// and the absorb in 256ths, and the record total before it is clamped.
//
//   g++ -O2 -std=c++23 -I<d2d>/components/rules -I<d2d>/components -I<d2d>/components/d2s -I<d2d>/components/txt \
//       -o gear_absorb_harness gear_absorb_harness.cpp
//   ./gear_absorb_harness < cases.tsv > d2d.tsv
#include "combat.hpp"
#include <cstdio>

int main() {
    char line[512];
    while (std::fgets(line, sizeof line, stdin)) {
        long id, type, field, res, pct, flat, mdr;
        if (line[0] == '#' || std::sscanf(line, "%ld %ld %ld %ld %ld %ld %ld", &id, &type, &field, &res, &pct, &flat, &mdr) != 7) continue;
        d2d::rules::Fighter defender;
        defender.mdr = int(mdr);
        if (type < 4) defender.res[std::size_t(type)] = int(res);
        // The gear's absorb exists from the claim on; the parent has none.
        [&](auto& unit) {
            if constexpr (requires { unit.absorb_pct; }) {
                unit.absorb_pct[std::size_t(type)] = int(pct);
                unit.absorb_flat[std::size_t(type)] = int(flat);
            }
        }(defender);
        d2d::rules::MonStats attacker;
        attacker.elements[0] = { int(type), 100, int(field), int(field), 0, "A1" };
        d2d::rules::Rng rng{ 1 };
        const auto taken = d2d::rules::monster_blow(defender, 1, d2d::rules::Motion::standing, attacker, false, rng, false, true, true);
        std::printf("%ld %d %d\n", id, taken.element_fine[std::size_t(type)], taken.sum.value_or(0));
    }
}
