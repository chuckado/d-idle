// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// VERIFYARMORHIT's own d2d side: `--varmor CASES` runs the real Fight::buff_events
// on the cases of verify_armor_hit.py (zombie1 as the striker, explicit Shiver /
// Chilling points, prior freeze / chill): `id chill freeze calc1`, frames from now.

#include <ai.hpp>
#include <character.hpp>
#include <character_store.hpp>
#include <combat.hpp>
#include <d2s_items.hpp>
#include <fight.hpp>
#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <montypes.hpp>
#include <rules.hpp>
#include <uniques.hpp>
#include <world.hpp>

#include <cstdint>
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
    if (argc < 3 || std::string_view(argv[1]) != "--varmor") { std::printf("SKIP: oracle mode\n"); return D2D_TEST_SKIPPED; }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir ? mpq_dir : "."), patch ? fs::path(patch) : fs::path{}, 0x4242);
    if (!data || !data->patched) { std::printf("SKIP: need 1.14d\n"); return D2D_TEST_SKIPPED; }
    d2d::rules::Rng rng(11);
    auto made = new_character(*data, 1, "Varmor", false, true, rng);
    Character character;
    character.character_class = 1;
    character.name = "Varmor";
    character.header = made.header;
    character.stats = made.stats;
    character.items = made.items;
    character.panel = panel_stats(*data, character.header, character.items, character.stats);
    const auto zombie = data->monsters.by_id.find("zombie1");
    if (zombie == data->monsters.by_id.end()) return 1;
    auto& type = const_cast<d2d::rules::MonType&>(data->monsters.types[zombie->second]);
    const auto real_diff = type.diff;
    const bool real_bcol = type.boss_column;
    const auto id_of = [&](const char* name) { return int(data->skills.by_name.find(name)->second); };
    const int frozen = id_of("Frozen Armor"), shiver = id_of("Shiver Armor"), chilling = id_of("Chilling Armor");
    const auto* row = data->skills.get(frozen);
    World world(&*data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    d2d::rules::Rng spawn_rng{ 9 };
    const Monster base = make_monster(*data, int(zombie->second), world.actor->player.x + 2, world.actor->player.y, spawn_rng, 0);
    const std::uint32_t now = 1500 * kTickMs;
    int shiver_pts = 0, chilling_pts = 0;
    world.fight.skill_base = [&](int id) { return id == shiver ? shiver_pts : id == chilling ? chilling_pts : 0; };
    std::ifstream in(argv[2]);
    std::string line;
    while (std::getline(in, line)) {
        std::istringstream f(line);
        int id, diff, kind, ce, res, pierce, level, sh, ch, phys, aidel, pf, pc;
        if (!(f >> id >> diff >> kind >> ce >> res >> pierce >> level >> sh >> ch >> phys >> aidel >> pf >> pc)) continue;
        *world.actor = start;
        world.actor->psum[d2d::d2s::kPassiveColdPierce] = pierce;
        world.actor->self_states.push_back({ frozen, level, now + 5000 * kTickMs });
        shiver_pts = sh;
        chilling_pts = ch;
        world.fight.game_difficulty = diff;
        type.diff[std::size_t(diff)].cold_effect = ce;
        type.boss_column = kind == 5;
        Monster m = base;
        m.difficulty = diff;
        m.boss = kind == 5 ? d2d::rules::Boss::none : d2d::rules::Boss(kind);
        m.stats.hit_points = m.hit_points = 1 << 20;
        m.boss_res = {};
        m.boss_res[4] = res - m.target(*data).res[4];
        m.stun_until = pf ? now + std::uint32_t(pf) * kTickMs : 0;
        m.chill_until = pc ? now + std::uint32_t(pc) * kTickMs : 0;
        world.fight.monsters.clear();
        world.fight.add_monster(m);
        const int got_res = world.fight.target_of(0).res[4];
        Foe foe;
        foe.melee_by.push_back(world.fight.monsters.back().id);
        if (phys > 0) foe.damaged_by.push_back(world.fight.monsters.back().id);   // ai.cpp:2698 feeds this from taken.damage
        world.fight.buff_events(foe, now);
        const auto& s = world.fight.monsters.back();
        const auto rel = [&](std::uint32_t until) { return until > now ? std::to_string((until - now) / kTickMs) : std::string("-"); };
        std::printf("%d %s %s %d%s\n", id, rel(s.chill_until).c_str(), rel(s.stun_until).c_str(), world.fight.calc(*row, row->calc[0], level), got_res == res ? "" : " RESFAIL");
        type.diff = real_diff;
        type.boss_column = real_bcol;
    }
    return 0;
} catch (const std::exception& e) {
    std::fprintf(stderr, "%s\n", e.what());
    return 1;
}
