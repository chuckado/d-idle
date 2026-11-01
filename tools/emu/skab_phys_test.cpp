// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// `--phys CASES GAME`: a missile's physical (and optional fire) roll and the
// life d2d's REAL rules::missile_blow takes, the cases of tools/emu/skab_phys.py
// (game.exe's FUN_005a89a0 + FUN_0057c6c0), one line each, to compare:
// `id A B`. A = damage x 256 + frac on the case's seed; B = the same with
// the roll game.exe made (GAME is its output: id rp re lost), min = max.
// (Our copy; it lives in the fork checkout's tests/game next to the others.)

#include <combat.hpp>
#include <rules.hpp>

#include <array>
#include <cstdint>
#include <cstdio>
#include <exception>
#include <fstream>
#include <map>
#include <sstream>
#include <string>
#include <string_view>

int main(int argc, char** argv) try {
    if (argc < 4 || std::string_view(argv[1]) != "--phys") {
        std::printf("SKIP: an oracle mode, run `test_skab_phys --phys CASES GAME` (see tools/emu/skab_phys.py)\n");
        return D2D_TEST_SKIPPED;
    }
    std::map<long long, std::pair<int, int>> rolled;
    {
        std::ifstream game(argv[3]);
        std::string line;
        while (std::getline(game, line)) {
            std::istringstream fields(line);
            long long id = 0; int rp = 0, re = 0;
            if (fields >> id >> rp >> re) rolled[id] = { rp, re };
        }
    }
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        long long id = 0;
        int lo = 0, hi = 0, res = 0, elo = 0, ehi = 0, eres = 0;
        unsigned long long seed_low = 0, seed_high = 0;
        if (!(fields >> id >> lo >> hi >> res >> elo >> ehi >> eres >> seed_low >> seed_high)) continue;
        d2d::rules::Target target;
        target.res[0] = res;       // physical
        target.res[2] = eres;      // fire
        const auto run = [&](int plo, int phi, int e_lo, int e_hi, d2d::rules::Rng& rng) {
            d2d::rules::MissileDamage damage{ .phys_lo = plo, .phys_hi = phi, .etype = e_hi > 0 ? 0 : -1, .elo = e_lo, .ehi = e_hi };
            const auto blow = d2d::rules::missile_blow(damage, target, {}, rng);
            return blow.damage * 256 + blow.frac;
        };
        d2d::rules::Rng seeded;
        seeded.low = std::uint32_t(seed_low & 0xFFFFFFFFULL);
        seeded.high = std::uint32_t(seed_high & 0xFFFFFFFFULL);
        const int a = run(lo, hi, elo, ehi, seeded);
        const auto found = rolled.find(id);
        int b = -1;
        if (found != rolled.end()) {
            d2d::rules::Rng unused;
            b = run(found->second.first, found->second.first, found->second.second, found->second.second, unused);
        }
        std::printf("%lld %d %d\n", id, a, b);
    }
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "test_skab_phys: %s\n", error.what());
    return 1;
}
