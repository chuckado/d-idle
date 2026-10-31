// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's armor absorb attribution and per-blow get-hit on tools/emu/hit_absorb.py's cases (d2lod-bot, PR #134 round 2).
// Default (AFTER): the REAL d2d::rules::absorb_blows + d2d::rules::struck_big of the patched combat.hpp (no copy),
//   the blows built as Foe::take builds them (damage = physical + fire / lightning / cold, mark = an element or poison).
// --before: PR #134 head 5466cd17, quoted: Fight::absorb (fight.cpp:1404..1416) takes min(pool, the frame's physical
//   or fire / lightning / cold, the frame's damage) off the sum; the lambda struck_big (fight.cpp:4873..4880) cuts that
//   amount off the blows in landing order whatever their kind, each blow's mark as Foe::take set it; rules::get_hit.
// Prints `id gh pool_left damage_left` (256ths), as hit_absorb.py's game side.
//
//   g++ -O2 -std=c++2c -I<d2d>/components/rules -I<d2d>/components/d2s -I<d2d>/components/txt -I<d2d>/components \
//       -o hit_absorb_harness hit_absorb_harness.cpp && ./hit_absorb_harness [--before] < cases.tsv > d2d.tsv
#include "combat.hpp"
#include <cstdio>
#include <cstring>
#include <vector>

int main(int argc, char** argv) {
    const bool before = argc > 1 && std::strcmp(argv[1], "--before") == 0;
    char line[512];
    std::vector<d2d::rules::StruckBlow> blows;
    long frame_id = 0, frame_pool = 0, frame_armor = 0, frame_max = 0;
    unsigned long frame_lo = 0, frame_hi = 0;
    while (std::fgets(line, sizeof line, stdin)) {
        long id, kind, maxlife, pool, armor, phys, fire, cold, ltng, poison, hc, next;
        unsigned long lo, hi;
        if (line[0] == '#' || std::sscanf(line, "%ld %ld %ld %ld %ld %ld %ld %ld %ld %ld %ld %lu %lu %ld", &id, &kind, &maxlife, &pool, &armor, &phys, &fire, &cold, &ltng,
                                           &poison, &hc, &lo, &hi, &next) != 14) continue;
        if (blows.empty()) { frame_id = id; frame_pool = pool; frame_armor = armor; frame_max = maxlife; frame_lo = lo; frame_hi = hi; }
        const int physical = int(phys >> 8), elemental = int((fire + cold + ltng) >> 8);
        blows.push_back({ physical + elemental, int(hc), elemental > 0 || poison > 0, physical, elemental, poison > 0 });
        if (next != 0) continue;
        int pool_left = int(frame_pool >> 8), damage = 0;
        for (const auto& blow : blows) damage += blow.damage;
        d2d::rules::Rng seed;
        seed.low = std::uint32_t(frame_lo); seed.high = std::uint32_t(frame_hi);
        bool gh = false;
        if (before) {
            int part = 0;
            for (const auto& blow : blows) part += frame_armor ? blow.elemental : blow.physical;
            int absorbed = pool_left > 0 && damage > 0 ? std::min({ pool_left, part, damage }) : 0;
            pool_left -= absorbed; damage -= absorbed;
            for (const auto& blow : blows) {
                const int cut = std::min(absorbed, blow.damage);
                absorbed -= cut;
                if (blow.damage - cut > 0 && d2d::rules::get_hit(frame_max, std::int64_t(blow.damage - cut) << 8, blow.hit_class, blow.marked, seed)) gh = true;
            }
        } else {
            if (pool_left > 0) damage -= d2d::rules::absorb_blows(blows, pool_left, frame_armor != 0);   // Fight::absorb's call
            gh = d2d::rules::struck_big(blows, frame_max, seed);
        }
        std::printf("%ld %d %d %d\n", frame_id, gh ? 1 : 0, pool_left << 8, damage << 8);
        blows.clear();
    }
}
