// SPDX-License-Identifier: GPL-3.0-or-later
// palskill: the d2d side of the Paladin differentials (tools/emu/pal_aura.py,
// pal_dmg.py, pal_flags.py). Built by hand against the libraries of a
// scratch upstream build (tools/emu/pal_build.sh), never part of upstream.
//   --skill-damage CASES   as tests/game/test_game.cpp's mode: `id min max len`
//   --aura CASES           `id range perdelay auralen a1..a6 p1..p5` (x: no stat)
//   --undead               `id undead demon` per game.exe MonStats unit id

#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <skills.hpp>

#include <array>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <fstream>
#include <map>
#include <sstream>
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
    if (mode == "--undead") {
        for (std::size_t id = 0; id < data->monsters.mon_bin.size(); ++id) {
            const auto row = data->monsters.mon_bin[id];
            if (row >= data->monsters.types.size()) continue;
            const auto& type = data->monsters.types[row];
            std::printf("%zu %d %d\n", id, int(type.undead), int(type.demon));
        }
        return 0;
    }
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        int id = 0, skill_id = 0, lvl = 0, clvl = 0, count = 0;
        std::array<int, 4> masteries{};
        fields >> id >> skill_id >> lvl >> clvl >> masteries[0] >> masteries[1] >> masteries[2] >> masteries[3] >> count;
        std::map<int, int> levels;
        for (int k = 0; k < count; ++k) { int other = 0, level = 0; fields >> other >> level; levels[other] = level; }
        levels[skill_id] = lvl;
        d2d::rules::CalcEnv env;
        env.clvl = clvl;
        env.base_level = env.level = [&](int other) { const auto found = levels.find(other); return found == levels.end() ? 0 : found->second; };
        env.stat = [&](int stat) { return stat >= d2d::d2s::kPassiveFireMastery && stat < d2d::d2s::kPassiveFireMastery + 4 ? masteries[std::size_t(stat - d2d::d2s::kPassiveFireMastery)] : 0; };
        const auto* skill = data->skills.get(skill_id);
        if (!skill) continue;
        if (mode == "--skill-damage") {
            std::printf("%d %d %d %d\n", id, d2d::rules::elem_damage(data->skills, *skill, env, lvl, false, 0, true),
                        d2d::rules::elem_damage(data->skills, *skill, env, lvl, true, 0, true), d2d::rules::elem_length(data->skills, *skill, env, lvl));
        } else if (mode == "--aura") {
            const auto value = [&](const d2d::rules::Calc& calc) { return d2d::rules::eval_calc(data->skills, calc, env, skill_id, lvl); };
            std::printf("%d %d %d %d", id, value(skill->aurarange), value(skill->perdelay_calc), value(skill->auralen));
            for (std::size_t k = 0; k < 6; ++k) { if (skill->aurastat[k] >= 0) std::printf(" %d", value(skill->aura_calc[k])); else std::printf(" x"); }
            for (std::size_t k = 0; k < 5; ++k) { if (skill->passive_stat[k] >= 0) std::printf(" %d", value(skill->passive_calc[k])); else std::printf(" x"); }
            std::printf("\n");
        }
    }
    return 0;
} catch (const std::exception& error) {
    std::printf("FAIL: %s\n", error.what());
    return 1;
}
