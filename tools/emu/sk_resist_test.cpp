// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// `--resist CASES`: a monster's resistance after a curse or aura and the
// element damage it lets through, the cases of tools/emu/sk_resist.py
// (game.exe's FUN_005c3540, FUN_0057be00, FUN_0057c6c0), one line each, to
// diff: `id adj effres dealt`. adj and effres follow d2d's code line for
// line (Fight::apply_target_stats adds the aurastat value as is;
// rules::missile_blow takes the pierce from fire..poison when res < 100,
// floored at -100 there); dealt is the REAL rules::missile_blow (256ths:
// damage x 256 + frac) for magic, fire, lightning and cold; physical and
// poison print "-" (whole-point / per-tick paths, not this claim).
// (Our copy; upstream path would be tests/game/test_resist.cpp.)

#include <combat.hpp>
#include <rules.hpp>

#include <algorithm>
#include <array>
#include <cstddef>
#include <cstdio>
#include <exception>
#include <fstream>
#include <sstream>
#include <string>
#include <string_view>

int main(int argc, char** argv) try {
    if (argc < 3 || std::string_view(argv[1]) != "--resist") {
        std::printf("SKIP: an oracle mode, run `test_resist --resist CASES` (see tools/emu/sk_resist.py)\n");
        return D2D_TEST_SKIPPED;
    }
    // case elem: 0 physical, 1 magic, 2 fire, 3 lightning, 4 cold, 5 poison
    static constexpr std::array<std::size_t, 6> kIndex{ 0, 1, 2, 3, 4, 5 };   // -> Target::res (ResDm..ResPo)
    static constexpr std::array<int, 6> kEtype{ -1, 4, 0, 1, 2, 3 };          // -> Skill::etype order
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        long long case_id = 0;
        int elem = 0, base = 0, value = 0, pierce = 0, roll = 0;
        if (!(fields >> case_id >> elem >> base >> value >> pierce >> roll) || elem < 0 || elem > 5) continue;
        d2d::rules::Target target;
        target.res[kIndex[std::size_t(elem)]] = base;
        target.res[kIndex[std::size_t(elem)]] += value;            // Fight::apply_target_stats: `target.res[i] += value`
        const int etype = kEtype[std::size_t(elem)];
        int res = target.res[kIndex[std::size_t(elem)]];
        if (etype >= 0 && etype < 4 && res < 100) res = std::max(res - pierce, -100);
        std::printf("%lld %d %d ", case_id, value, res);
        if (etype < 0 || etype == 3) { std::printf("-\n"); continue; }
        std::array<int, 4> pierces{};
        if (etype < 4) pierces[std::size_t(etype)] = pierce;
        d2d::rules::Rng rng;
        const auto damage = d2d::rules::MissileDamage{ .etype = etype, .elo = roll, .ehi = roll };
        const auto blow = d2d::rules::missile_blow(damage, target, pierces, rng);
        std::printf("%d\n", blow.damage * 256 + blow.frac);
    }
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "test_resist: %s\n", error.what());
    return 1;
}
