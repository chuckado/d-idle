// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's gear element roll of a melee blow on d35_elem.py's cases: the REAL make_fighter (mastery scaling) and player_blow
// (the roll), a fighter with no physical damage, a target with no resistance and defense 0; the seed is the case's two words.
// Prints `id hit whole`: the blow's damage in whole points (hit 0: the to-hit roll missed).
#include <combat.hpp>
#include <d2s_items.hpp>
#include <rules.hpp>
#include <cstdio>

using namespace d2d::rules;

int main() {
    char line[512];
    namespace s = d2d::d2s;
    static constexpr int kMin[5] = { s::kFireMinDamage, s::kLightningMinDamage, s::kColdMinDamage, 0, s::kMagicMinDamage };
    static constexpr int kMax[5] = { s::kFireMaxDamage, s::kLightningMaxDamage, s::kColdMaxDamage, 0, s::kMagicMaxDamage };
    static constexpr int kMast[5] = { s::kPassiveFireMastery, s::kPassiveLightningMastery, s::kPassiveColdMastery, 0, s::kPassiveMagicMastery };
    while (std::fgets(line, sizeof line, stdin)) {
        long id, s0, s1, el, mn, mx, mast;
        if (line[0] == '#' || std::sscanf(line, "%ld %ld %ld %ld %ld %ld %ld", &id, &s0, &s1, &el, &mn, &mx, &mast) != 7) continue;
        Tables tables;
        StatSum sum{}, wsum{};
        d2d::d2s::Stats stats;
        stats.values[s::kLevel] = 1;
        sum[kMin[el]] = mn; sum[kMax[el]] = mx; sum[kMast[el]] = mast;
        Fighter fighter = make_fighter(tables, nullptr, nullptr, sum, wsum, stats, {}, 0, {});
        fighter.phys_lo = fighter.phys_hi = 0; fighter.phys_pct = 0;
        fighter.ar_base = 1 << 20; fighter.ar_pct = 0;
        Target target; target.hit_points = target.max_hp = 1 << 20;
        Blow blow;
        // the to-hit roll comes first on d2d's seed; game.exe's roll is the first draw, so run d2d from the state one step before
        Rng rng; rng.low = std::uint32_t(s0); rng.high = std::uint32_t(s1);
        const Rng start = rng;
        (void)start;
        blow = player_blow(fighter, target, 1, rng);
        std::printf("%ld %d %d\n", id, blow.hit ? 1 : 0, blow.damage);
    }
}
