// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// VERIFYCURSEIMMUNE's own d2d side (independent of the author's test_sk_resist):
//   --vcurse CASES  the real Fight::target_of on zombie1 with the case's six base
//                   resistances (boss_res pins them), a REAL curse skill / level on
//                   the monster and a REAL enemy aura / level on the player (the
//                   skill rows' own aurastat calcs, no value fed in):
//                   `id r0 r1 r2 r3 r4 r5 [AURAOUT] [BASEFAIL]`
//   --vlist         every Skills row with a resistance aurastat (36 37 39 41 43 45) or srvdofunc 30:
//                   `L id srvdofunc level stat:value ...` for levels 1..99, values from
//                   d2d's eval_calc (the same call apply_target_stats makes).
// Case line (verify_curse_immune.py gen): id class b0 b1 b2 b3 b4 b5 curse clvl aura alvl
// (skill ids, -1 / 0 none). Built as tests/game/test_verify_curse_immune.cpp.

#include <ai.hpp>
#include <character.hpp>
#include <character_store.hpp>
#include <combat.hpp>
#include <fight.hpp>
#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <rules.hpp>
#include <skills.hpp>
#include <world.hpp>

#include <array>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <fstream>
#include <sstream>
#include <string>
#include <string_view>

using namespace d2d::game;

int main(int argc, char** argv) try {
    const std::string_view mode = argc > 1 ? std::string_view(argv[1]) : std::string_view();
    if (mode != "--vcurse" && mode != "--vlist") { std::printf("SKIP: oracle mode\n"); return D2D_TEST_SKIPPED; }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir ? mpq_dir : "."), patch ? fs::path(patch) : fs::path{}, 0x4242);
    if (!data || !data->patched) { std::printf("SKIP: need 1.14d\n"); return D2D_TEST_SKIPPED; }
    d2d::rules::Rng rng(13);
    auto made = new_character(*data, 3, "Vcurse", false, true, rng);   // a Paladin (Conviction)
    Character character;
    character.character_class = 3;
    character.name = "Vcurse";
    character.header = made.header;
    character.stats = made.stats;
    character.items = made.items;
    character.panel = panel_stats(*data, character.header, character.items, character.stats);
    World world(&*data, -1, -1);
    world.enter(character);
    auto& fight = world.fight;
    if (mode == "--vlist") {
        for (const auto& row : data->skills.rows) {
            bool res = false;
            for (int s : row.aurastat) res = res || s == 36 || s == 37 || s == 39 || s == 41 || s == 43 || s == 45;
            if (!res && row.srvdofunc != d2d::rules::ServerDoFunction::kCurse) continue;
            for (int lvl = 1; lvl <= 99; ++lvl) {
                std::printf("L %d %d %d", row.id, int(row.srvdofunc), lvl);
                for (std::size_t k = 0; k < row.aurastat.size(); ++k)
                    if (row.aurastat[k] >= 0)
                        std::printf(" %d:%d", row.aurastat[k], d2d::rules::eval_calc(data->skills, row.aura_calc[k], fight.calc_env(), row.id, lvl));
                std::printf("\n");
            }
        }
        return 0;
    }
    if (argc < 3) return 1;
    const auto zombie = data->monsters.by_id.find("zombie1");
    if (zombie == data->monsters.by_id.end()) return 1;
    d2d::rules::Rng spawn_rng{ 9 };
    const Monster proto = make_monster(*data, int(zombie->second), world.actor->player.x + 1, world.actor->player.y, spawn_rng, 0);
    int aura_level = 1;
    fight.skill_level = [&](int) { return aura_level; };
    std::ifstream in(argv[2]);
    std::string line;
    while (std::getline(in, line)) {
        std::istringstream f(line);
        int id = 0, cls = 0, curse = -1, clvl = 0, aura = 0, alvl = 0;
        std::array<int, 6> base{};
        if (!(f >> id >> cls >> base[0] >> base[1] >> base[2] >> base[3] >> base[4] >> base[5] >> curse >> clvl >> aura >> alvl)) continue;
        Monster m = proto;
        m.difficulty = 0;
        m.boss_res = {};
        const auto plain = m.target(*data).res;
        for (std::size_t i = 0; i < 6; ++i) m.boss_res[i] = base[i] - plain[i];
        m.curse = {};
        m.cry = {};
        if (curse >= 0) { m.curse.skill = curse; m.curse.level = clvl; m.curse.until = 0xFFFFFFF0u; }
        fight.monsters.clear();
        fight.add_monster(m);
        world.actor->aura = aura;
        aura_level = alvl;
        const bool base_ok = fight.monsters.back().target(*data).res == base;
        const bool aura_out = aura > 0 && !fight.in_aura(fight.monsters.back());
        const auto got = fight.target_of(0).res;
        std::printf("%d %d %d %d %d %d %d%s%s\n", id, got[0], got[1], got[2], got[3], got[4], got[5], aura_out ? " AURAOUT" : "", base_ok ? "" : " BASEFAIL");
    }
    return 0;
} catch (const std::exception& e) {
    std::fprintf(stderr, "%s\n", e.what());
    return 1;
}
