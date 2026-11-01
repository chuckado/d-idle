// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// SKGROUND2: Blizzard through the REAL Fight (our copy; upstream path would
// be tests/game/test_sk_ground.cpp). Two oracle modes:
//   `--blizz SEED N`: N casts of Blizzard (a real Fight::fire at a point, the
//     centre stepped 100 frames by Fight::missile_tick) on an empty open
//     level; one line per shard `S<case> <frame> <dx> <dy>`: the subtile
//     offset of the shard from the centre's subtile. game.exe's offsets are
//     -6..5 at radius 7 (tools/emu/sk_ground_blizz.py, V run).
//   `--hit`: one standing `blizzard1` shard (Size 2) and one hostile monster
//     of each size 1..3 at every offset -7..7 subtiles; the monster has a
//     million life and the shard deals a fixed 10 points, so the monster's
//     life lost / 10 over the shard's frames is its hits. One line
//     `H <size> <class> <dx> <dy> <hits>` (game.exe's: tools/emu/sk_ground_hit.py).

#include <ai.hpp>
#include <character.hpp>
#include <character_store.hpp>
#include <combat.hpp>
#include <d2s_items.hpp>
#include <fight.hpp>
#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <level_ids.hpp>
#include <rules.hpp>
#include <world.hpp>

#include <algorithm>
#include <array>
#include <cassert>
#include <cmath>
#include <cstddef>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <string>
#include <string_view>
#include <utility>
#include <vector>

using namespace d2d::game;

int main(int argc, char** argv) try {
    const std::string_view mode = argc > 1 ? argv[1] : "";
    if (mode != "--blizz" && mode != "--hit") { std::printf("SKIP: needs --blizz SEED N or --hit (see tools/emu/sk_ground_blizz.py, sk_ground_hit.py)\n"); return D2D_TEST_SKIPPED; }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* home = std::getenv("HOME");
    const fs::path data_dir = mpq_dir ? fs::path(mpq_dir) : fs::path(home ? home : "") / "Workspace/private/diablo2";
    if (!fs::exists(data_dir / "d2data.mpq")) { std::printf("SKIP: no d2data.mpq in %s\n", data_dir.string().c_str()); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");

    auto data = load_game_data(data_dir, patch ? fs::path(patch) : fs::path{}, 0x1234);
    assert(data && data->town.id == 1 && !data->town.walk.empty());
    if (!data->patched) { std::printf("SKIP: needs the 1.14d data (D2_PATCH_INSTALLER)\n"); return D2D_TEST_SKIPPED; }
    d2d::rules::Rng rng(7);
    constexpr int kSorceress = 1;
    auto made = new_character(*data, kSorceress, "Headless", false, true, rng);
    Character character;
    character.character_class = kSorceress;
    character.name = "Headless";
    character.header = made.header;
    character.stats = made.stats;
    character.items = made.items;
    character.panel = panel_stats(*data, character.header, character.items, character.stats);
    World world(&*data, -1, -1);
    world.enter(character);
    Level open = *data->level(d2d::rules::level_ids::kBloodMoor);
    std::ranges::fill(open.walk, std::uint16_t(0));
    world.level = &open;
    world.fight.enter(&open);
    int at_x = -1, at_y = -1;
    for (int y = 20; y < 400 && at_x < 0; ++y)
        for (int x = 20; x < 400 && at_x < 0; ++x)
            if (std::ranges::all_of(std::array<std::pair<int, int>, 5>{ { { 0, 0 }, { 30, 0 }, { -30, 0 }, { 0, 30 }, { 0, -30 } } },
                                    [&](auto off) { return !missile_rooms(*data, open, float(x + off.first) + 0.5f, float(y + off.second) + 0.5f).empty(); }))
                at_x = x * 5 + 2, at_y = y * 5 + 2;
    assert(at_x >= 0);
    world.actor->player.x = (float(at_x) + 0.5f) / 5;
    world.actor->player.y = (float(at_y) + 0.5f) / 5;
    const Crowd crowd;
    std::uint32_t now = 1000 * kTickMs;
    const int blizzard = data->skills.by_name.at("Blizzard");
    const auto& skill = *data->skills.get(blizzard);

    if (mode == "--blizz") {
        std::uint32_t state = std::uint32_t(argc > 2 ? std::atoi(argv[2]) : 1);
        if (state == 0) state = 1;
        const auto next = [&state] { state ^= state << 13; state ^= state >> 17; state ^= state << 5; return state; };
        const int cases = argc > 3 ? std::atoi(argv[3]) : 100;
        long shards = 0;
        for (int i = 0; i < cases; ++i) {
            world.fight.monsters.clear(); world.fight.missiles.clear(); world.fight.pending.clear();
            world.fight.missile_frame = -1;
            const int cx = at_x + int(next() % 41) - 20, cy = at_y + int(next() % 41) - 20;   // the centre's subtile
            Fight::Caster caster;
            caster.x = world.actor->player.x; caster.y = world.actor->player.y; caster.level = 1 + int(next() % 20);
            caster.to_x = (float(cx) + 0.5f) / 5; caster.to_y = (float(cy) + 0.5f) / 5;
            world.fight.fire(skill, now, caster);
            for (int frame = 1; frame <= 110; ++frame) {
                now += kTickMs;
                world.fight.world(true, now, 0.04f, crowd);
                for (const auto& missile : world.fight.missiles)
                    if (missile.info && missile.info->name == "blizzard1" && missile.born == now) {
                        ++shards;
                        std::printf("S%d %d %d %d\n", i, frame, int(std::floor(missile.x * 5)) - cx, int(std::floor(missile.y * 5)) - cy);
                    }
            }
            now += 100 * kTickMs;
        }
        std::fprintf(stderr, "casts %d shards %ld\n", cases, shards);
        return 0;
    }

    // --hit
    std::array<int, 4> kind{ -1, -1, -1, -1 };
    for (const char* name : { "flyingscimitar", "maggotegg1", "fallen1", "zombie1", "bighead1", "brute2", "wendigo1" })
        if (const auto found = data->monsters.by_id.find(name); found != data->monsters.by_id.end())
            if (const int size = data->monsters.types[found->second].size; size >= 1 && size <= 3 && kind[std::size_t(size)] < 0) kind[std::size_t(size)] = int(found->second);
    for (std::size_t row = 0; row < data->monsters.types.size(); ++row)
        if (const int size = data->monsters.types[row].size; size >= 1 && size <= 3 && kind[std::size_t(size)] < 0) kind[std::size_t(size)] = int(row);
    const auto& shard_row = data->missiles.at("blizzard1");
    for (int size = 1; size <= 3; ++size) {
        if (kind[std::size_t(size)] < 0) { std::printf("SKIP: no size %d row\n", size); return D2D_TEST_SKIPPED; }
        std::fprintf(stderr, "size %d: MonStats row %d %s\n", size, kind[std::size_t(size)], data->monsters.types[std::size_t(kind[std::size_t(size)])].id.c_str());
        for (int dx = -7; dx <= 7; ++dx)
            for (int dy = -7; dy <= 7; ++dy) {
                world.fight.monsters.clear(); world.fight.missiles.clear(); world.fight.pending.clear();
                world.fight.missile_frame = -1;
                d2d::rules::Rng spawn_rng{ 5 };
                Monster monster = make_monster(*data, kind[std::size_t(size)], (float(at_x + dx) + 0.5f) / 5, (float(at_y + dy) + 0.5f) / 5, spawn_rng, 0);
                monster.stats.hit_points = monster.hit_points = 1000000;
                monster.life_frac = 0;
                monster.next_act = ~0u;
                world.fight.add_monster(monster);
                monster = world.fight.monsters[0];
                {
                    auto& mon = world.fight.monsters[0];
                    mon.boss_res = {};
                    const auto res = world.fight.target_of(0).res;
                    for (std::size_t k = 0; k < mon.boss_res.size() && k < res.size(); ++k) mon.boss_res[k] = -res[k];
                }
                world.fight.launch(shard_row, skill, 1, (float(at_x) + 0.5f) / 5, (float(at_y) + 0.5f) / 5, 0, 0, shard_row.range, now);
                for (auto& missile : world.fight.pending) missile.fixed = 10 << 8;
                int hits = 0;
                for (int frame = 1; frame <= shard_row.range + 1; ++frame) {
                    now += kTickMs;
                    const int before = world.fight.monsters[0].hit_points;
                    world.fight.monsters[0].next_act = ~0u;
                    world.fight.world(true, now, 0.04f, crowd);
                    hits += (before - world.fight.monsters[0].hit_points) / 10;
                }
                std::printf("H %d %d %d %d %d\n", size, kind[std::size_t(size)], dx, dy, hits);
                now += 20 * kTickMs;
            }
    }
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "%s\n", error.what());
    return 1;
}
