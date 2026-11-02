// SPDX-License-Identifier: GPL-3.0-or-later
// monsterstats: the d2d side of tools/emu/ms_stats.py. Calls d2d's own
// rules::monster_stats / monster_ranges (components/rules/monsters.hpp) on the
// real loaded tables (load_game_data), for every MonStats row x difficulty x
// level 1..last MonLvl row (the type's level is set to the case level and no
// level_add, so the function runs as the game would at that level).
// Built by hand against a scratch upstream build (tools/emu/ms_build.sh), never
// part of upstream.
//   ms_d2d_test --stats     row diff level minhp maxhp ac th_a1 a1min a1max a2min a2max s1min s1max exp lo32 hi32
//                           (lo32 / hi32: the min / max of 32 rolled hit points, to check the range)
//   ms_d2d_test --res       row diff dr mr fire light cold poison  (d2d's base resistances, MonStats columns)
//   ms_d2d_test --meta      row id level_n level_h noratio
//   ms_d2d_test --players   n life_pct exp_pct
//   ms_d2d_test --th        row diff level a1th a2th s1th (d2d's three to-hit values)

#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <monsters.hpp>

#include <algorithm>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <string>
#include <string_view>

using namespace d2d::game;

int main(int argc, char** argv) try {
    if (argc < 2) return 2;
    const std::string_view mode = argv[1];
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir), patch ? fs::path(patch) : fs::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("FAIL: no patched data\n"); return 1; }
    auto& monsters = data->monsters;
    namespace R = d2d::rules;
    if (mode == "--players") {
        for (int n = 0; n <= 16; ++n) std::printf("%d %d %d\n", n, R::player_life_pct(n), R::player_exp_pct(n));
        return 0;
    }
    if (mode == "--bin") {   // game.exe's MonStats unit id -> d2d's row
        for (std::size_t id = 0; id < data->mon_bin.size(); ++id) std::printf("%zu %u\n", id, unsigned(data->mon_bin[id]));
        return 0;
    }
    const std::size_t levels = monsters.lvl.size();
    for (std::size_t row = 0; row < monsters.types.size(); ++row) {
        auto& type = const_cast<R::MonType&>(monsters.types[row]);
        if (mode == "--meta") { std::printf("%zu %s %d %d %d %d\n", row, type.id.c_str(), type.level[1], type.level[2], int(type.no_ratio), type.level[0]); continue; }
        for (int diff = 0; diff < 3; ++diff) {
            if (mode == "--res") {
                const auto& r = type.diff[std::size_t(diff)].res;
                std::printf("%zu %d %d %d %d %d %d %d\n", row, diff, r[0], r[1], r[2], r[3], r[4], r[5]);
                continue;
            }
            const int own = type.level[std::size_t(diff)];
            for (std::size_t level = 1; level < levels; ++level) {
                type.level[std::size_t(diff)] = int(level);
                R::Rng seed(1);
                const auto st = R::monster_stats(monsters, int(row), diff, seed);
                if (mode == "--th") {
                    std::printf("%zu %d %zu %d %d %d\n", row, diff, level, st.to_hit, st.to_hit2, st.to_hit_s1);
                    continue;
                }
                const auto rg = R::monster_ranges(monsters, int(row), diff);
                int lo = 1 << 30, hi = 0;
                for (std::uint32_t s = 1; s <= 32; ++s) {
                    R::Rng rng(s * 2654435761u);
                    const auto x = R::monster_stats(monsters, int(row), diff, rng);
                    lo = std::min(lo, x.hit_points); hi = std::max(hi, x.hit_points);
                }
                std::printf("%zu %d %zu %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d\n", row, diff, level, rg.min_hp, rg.max_hp, st.armor_class, st.to_hit,
                            st.a1_min, st.a1_max, st.a2_min, st.a2_max, st.s1_min, st.s1_max, st.exp, lo, hi, st.to_hit2, st.to_hit_s1);
            }
            type.level[std::size_t(diff)] = own;
        }
    }
    return 0;
} catch (const std::exception& error) {
    std::printf("FAIL: %s\n", error.what());
    return 1;
}
