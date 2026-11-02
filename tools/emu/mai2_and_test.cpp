// SPDX-License-Identifier: GPL-3.0-or-later
// monsterai2: d2d's rules::andariel_think replayed on game.exe's own Andariel runs (tools/emu/mai2_six.py --dump-andariel).
// A line: diff melee aip1 aip2 aip3 aip4 skill1_set skill2_set seed_lo seed_hi | game act index, n, seed_lo, seed_hi after
// (ACTS: a1 = 1, idle = 0, skill = 4, walk = 5). d2d's function assumes both skills set (her MonStats row has both), so only those
// lines are compared; the others are counted apart. Prints `and cases compared differ skipped`. Built by tools/emu/mai2_build.sh.
#include <monsters.hpp>

#include <array>
#include <cstdio>
#include <fstream>
#include <sstream>
#include <string>

int main(int argc, char** argv) {
    if (argc < 2) return 2;
    std::ifstream in(argv[1]);
    int compared = 0, differ = 0, skipped = 0;
    for (std::string line; std::getline(in, line);) {
        std::istringstream fields(line);
        long long diff, melee, a1, a2, a3, a4, s1, s2, lo, hi, act, n, lo2, hi2;
        if (!(fields >> diff >> melee >> a1 >> a2 >> a3 >> a4 >> s1 >> s2 >> lo >> hi >> act >> n >> lo2 >> hi2)) continue;
        if (!s1 || !s2) { ++skipped; continue; }
        std::array<int, 8> aip{ int(a1), int(a2), int(a3), int(a4), 0, 0, 0, 0 };
        d2d::rules::Rng seed;
        seed.low = std::uint32_t(lo); seed.high = std::uint32_t(hi);
        const auto got = d2d::rules::andariel_think(melee != 0, aip, seed);
        using Act = d2d::rules::AndarielAct;
        const int want_act = got == Act::spray || got == Act::bolt ? 4 : got == Act::melee ? 1 : got == Act::idle ? 0 : 5;
        const int want_n = got == Act::spray ? 0 : got == Act::bolt ? 1 : got == Act::walk ? 7 : got == Act::idle ? 5 : 0;   // ai.cpp:1376-1380 gives idle 5, walk 7
        const bool same = want_act == act && want_n == n && seed.low == std::uint32_t(lo2) && seed.high == std::uint32_t(hi2);
        ++compared;
        if (!same) { ++differ; std::fprintf(stderr, "and mismatch: %s\n  d2d act %d n %d\n", line.c_str(), want_act, want_n); }
    }
    std::printf("and %d %d %d\n", compared, differ, skipped);
    return differ != 0;
}
