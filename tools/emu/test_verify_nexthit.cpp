// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// VERIFYNEXTHIT: independent d2d side of verify_nexthit.py. `--vnh SEED N [edge]`.
// Nova / Frost Nova casts (one or two) on 1..6 monsters (standing or moving)
// through the real Fight::fire / Fight::world. Lines:
//   K<case> <skill> <lvl> <n> <cast2 frame or 0> <size>@<x>,<y>/<vx>,<vy> ...
//   I<case> <frame> <missile>:<m>.<m> ...   contacts, monster index order
//   P<case> <frame> <missile>:<m>.<m> ...   contacts, crossed-subtile (path) order
//   D<case> <frame> <m>x<hits> ...          d2d's hits that frame
//   S<case> <hits per monster>

#include <ai.hpp>
#include <character.hpp>
#include <character_store.hpp>
#include <d2s_items.hpp>
#include <fight.hpp>
#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <level_ids.hpp>
#include <rules.hpp>
#include <world.hpp>

#include <algorithm>
#include <array>
#include <cmath>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <string>
#include <string_view>
#include <vector>

using namespace d2d::game;

int main(int argc, char** argv) try {
    if (argc < 4 || std::string_view(argv[1]) != "--vnh") { std::printf("SKIP\n"); return D2D_TEST_SKIPPED; }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    if (!mpq_dir) { std::printf("SKIP\n"); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir), patch ? fs::path(patch) : fs::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("SKIP\n"); return D2D_TEST_SKIPPED; }
    d2d::rules::Rng rng(11);
    auto made = new_character(*data, 1, "Verifier", false, true, rng);
    Character character;
    character.character_class = 1; character.name = "Verifier";
    character.header = made.header; character.stats = made.stats; character.items = made.items;
    character.panel = panel_stats(*data, character.header, character.items, character.stats);
    World world(&*data, -1, -1);
    world.enter(character);
    Level open = *data->level(d2d::rules::level_ids::kBloodMoor);
    std::ranges::fill(open.walk, std::uint16_t(0));
    world.level = &open;
    world.fight.enter(&open);
    // caster subtile with missile rooms 20 subtiles all round
    int cx = -1, cy = -1;
    for (int y = 30; y < 400 && cx < 0; ++y)
        for (int x = 30; x < 400 && cx < 0; ++x) {
            bool ok = true;
            for (int ox = -4; ox <= 4 && ok; ox += 4)
                for (int oy = -4; oy <= 4 && ok; oy += 4)
                    ok = !missile_rooms(*data, open, float(x + ox) + 0.5f, float(y + oy) + 0.5f).empty();
            if (ok) cx = x * 5 + 2, cy = y * 5 + 2;
        }
    if (cx < 0) { std::fprintf(stderr, "no spot\n"); return 1; }
    world.actor->player.x = (float(cx) + 0.5f) / 5;
    world.actor->player.y = (float(cy) + 0.5f) / 5;
    // a MonStats row per size: the first hostile-looking row of that size from my own list, else any
    std::array<int, 4> kind{ -1, -1, -1, -1 };
    for (const char* name : { "quillrat1", "fallen1", "skeleton1", "zombie1", "brute1", "wendigo1", "fetish1", "sandmaggot1" })
        if (auto f = data->monsters.by_id.find(name); f != data->monsters.by_id.end()) {
            const int size = data->monsters.types[f->second].size;
            if (size >= 1 && size <= 3 && kind[std::size_t(size)] < 0) kind[std::size_t(size)] = int(f->second);
        }
    for (std::size_t row = 0; row < data->monsters.types.size(); ++row)
        if (const int size = data->monsters.types[row].size; size >= 1 && size <= 3 && kind[std::size_t(size)] < 0) kind[std::size_t(size)] = int(row);
    for (int s = 1; s <= 3; ++s) std::fprintf(stderr, "size %d row %d %s\n", s, kind[std::size_t(s)], data->monsters.types[std::size_t(kind[std::size_t(s)])].id.c_str());
    const int nova = data->skills.by_name.at("Nova"), frost = data->skills.by_name.at("Frost Nova");
    const Crowd crowd;
    std::uint32_t now = 5000 * kTickMs;
    struct Mon { int size; float x, y, vx, vy; };   // subtiles from the caster, subtiles per frame
    const auto run = [&](const std::string& tag, int skill_id, int lvl, std::vector<Mon> mons, int cast2) {
        const auto& skill = *data->skills.get(skill_id);
        std::printf("K%s %s %d %zu %d", tag.c_str(), skill_id == nova ? "nova" : "frostnova", lvl, mons.size(), cast2);
        for (const auto& m : mons) std::printf(" %d@%.2f,%.2f/%.2f,%.2f", m.size, m.x, m.y, m.vx, m.vy);
        std::printf("\n");
        world.fight.monsters.clear(); world.fight.missiles.clear(); world.fight.pending.clear();
        d2d::rules::Rng spawn_rng{ 9 };
        for (const auto& m : mons) {
            Monster monster = make_monster(*data, kind[std::size_t(m.size)], (float(cx) + 0.5f + m.x) / 5, (float(cy) + 0.5f + m.y) / 5, spawn_rng, 0);
            monster.stats.hit_points = monster.hit_points = 2000000;
            monster.life_frac = 0; monster.next_act = ~0u;
            world.fight.add_monster(monster);
        }
        for (std::size_t i = 0; i < world.fight.monsters.size(); ++i) {
            auto& monster = world.fight.monsters[i];
            monster.boss_res = {};
            const auto res = world.fight.target_of(i).res;
            for (std::size_t k = 0; k < monster.boss_res.size() && k < res.size(); ++k) monster.boss_res[k] = -res[k];
        }
        Fight::Caster caster;
        caster.x = world.actor->player.x; caster.y = world.actor->player.y; caster.level = lvl;
        caster.to_x = caster.x; caster.to_y = caster.y + 1;
        const auto cast = [&] {
            world.fight.fire(skill, now, caster);
            for (auto& missile : world.fight.pending) missile.fixed = 10 << 8;
        };
        cast();
        std::vector<int> total(mons.size(), 0);
        for (int frame = 1; frame <= 60; ++frame) {
            if (frame == cast2) cast();
            if (world.fight.missiles.empty() && world.fight.pending.empty() && frame > 1 && frame > cast2) break;
            now += kTickMs;
            std::vector<int> before;
            for (std::size_t i = 0; i < world.fight.monsters.size(); ++i) {
                auto& monster = world.fight.monsters[i];
                monster.next_act = ~0u;
                monster.unit.x += mons[i].vx / 5; monster.unit.y += mons[i].vy / 5;
                before.push_back(monster.hit_points);
            }
            world.fight.world(true, now, 0.04f, crowd);
            std::string by_index, by_path;
            for (const auto& missile : world.fight.missiles) {
                if (missile.skill != skill_id) continue;
                std::string a, b;
                std::vector<int> seen;
                const int msize = missile.info ? missile.info->size : 1;
                for (const auto& spot : missile.flight.crossed)
                    for (std::size_t i = 0; i < world.fight.monsters.size(); ++i) {
                        const auto& monster = world.fight.monsters[i];
                        if (!monster.targetable() || std::ranges::find(seen, int(i)) != seen.end()) continue;
                        const int fx = open.world_x * 5 + int(std::floor(monster.unit.x * 5)), fy = open.world_y * 5 + int(std::floor(monster.unit.y * 5));
                        if (d2d::rules::missile_touches(spot.first, spot.second, fx, fy, data->monsters.types[std::size_t(monster.type)].size, msize)) seen.push_back(int(i));
                    }
                for (const int i : seen) b += (b.empty() ? "" : ".") + std::to_string(i);
                std::vector<int> sorted = seen; std::ranges::sort(sorted);
                for (const int i : sorted) a += (a.empty() ? "" : ".") + std::to_string(i);
                if (!seen.empty()) { by_index += " " + std::to_string(missile.sound) + ":" + a; by_path += " " + std::to_string(missile.sound) + ":" + b; }
            }
            if (!by_index.empty()) { std::printf("I%s %d%s\n", tag.c_str(), frame, by_index.c_str()); std::printf("P%s %d%s\n", tag.c_str(), frame, by_path.c_str()); }
            std::string hits;
            for (std::size_t i = 0; i < world.fight.monsters.size(); ++i) {
                const int lost = before[i] - world.fight.monsters[i].hit_points;
                if (lost <= 0) continue;
                total[i] += lost / 10;
                hits += " " + std::to_string(i) + "x" + std::to_string(lost / 10) + (lost % 10 ? "?" : "");
            }
            if (!hits.empty()) std::printf("D%s %d%s\n", tag.c_str(), frame, hits.c_str());
        }
        std::printf("S%s", tag.c_str());
        for (const int h : total) std::printf(" %d", h);
        std::printf("\n");
        now += 200 * kTickMs;
    };
    std::uint32_t state = std::uint32_t(std::strtoul(argv[2], nullptr, 0)) & 0xffffffffu;
    const auto next = [&state] { state = (state * 1664525u + 1013904223u) & 0xffffffffu; return state >> 8; };
    int id = 0;
    if (argc > 4 && std::string_view(argv[4]) == "edge")
        for (const int skill_id : { nova, frost })
            for (int size = 1; size <= 3; ++size)
                for (const int d : { 0, 1, 2, 3, 4, 5, 6, 7, 9, 11, 13, 15, 17 })
                    for (int h = 0; h < 8; ++h) {
                        const double a = h * 3.14159265358979 / 4;
                        run("E" + std::to_string(id++), skill_id, 1, { { size, float(std::lround(d * std::cos(a))), float(std::lround(d * std::sin(a))), 0, 0 } }, 0);
                    }
    static constexpr std::array<float, 7> kVel{ -1.0f, -0.6f, -0.3f, 0.0f, 0.3f, 0.6f, 1.0f };
    for (int c = 0; c < std::atoi(argv[3]); ++c) {
        const int skill_id = next() % 2 ? nova : frost;
        const int lvl = 1 + int(next() % 30);
        const int count = 1 + int(next() % 6);
        const bool moving = next() % 2;
        const int cast2 = next() % 5 == 0 ? 1 + int(next() % 8) : 0;
        std::vector<Mon> mons;
        for (int k = 0; k < count; ++k) {
            Mon m{};
            m.size = 1 + int(next() % 3);
            const double r = double(next() % 1700) / 100.0, a = double(next() % 3600) / 3600.0 * 6.28318530718;
            m.x = float(std::lround(r * std::cos(a) * 4)) / 4; m.y = float(std::lround(r * std::sin(a) * 4)) / 4;
            if (moving) { m.vx = kVel[next() % kVel.size()]; m.vy = kVel[next() % kVel.size()]; }
            mons.push_back(m);
        }
        run("R" + std::to_string(c), skill_id, lvl, mons, cast2);
    }
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "%s\n", error.what());
    return 1;
}
