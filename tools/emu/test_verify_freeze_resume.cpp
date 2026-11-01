// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// VERIFYFREEZERESUME d2d probe (verifier's own; not the candidate's test).
// `--vfr CASES ORDER`: replays tools/emu/verify_freeze_resume.py cases on a
// real d2d build. Each case: a zombie1 body (its think swings when its foe
// stands in melee) carrying the case row's own loaded aidel for the case
// difficulty, frozen through the REAL Fight::strike (freeze argument) at the
// case frames, monster_update run every tick from before the first event.
// ORDER n: d2d's own tick order (Fight::world runs monster_update, then the
// player's missiles strike: fight/tick.cpp); ORDER g: game.exe's order
// (FUN_005414d0 runs the missile group before the monster group).
// A prior stun is written as the one stun site writes it (strike.cpp).
// Prints `id swing think`: the first frame (relative to the first freeze)
// the monster swings, and the first frame its next_act moves for any reason
// other than the claim's refile (`-` none within the horizon).

#include <ai.hpp>
#include <character.hpp>
#include <character_store.hpp>
#include <fight.hpp>
#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <level_ids.hpp>
#include <rules.hpp>
#include <uniques.hpp>
#include <world.hpp>

#include <array>
#include <cassert>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <fstream>
#include <sstream>
#include <string>
#include <string_view>
#include <vector>

using namespace d2d::game;

namespace {
template <class M> void write_stun(M& monster, std::uint32_t until) {
    if constexpr (requires { monster.stunned_until; }) monster.stunned_until = until;
    monster.stun_until = until;
}
}  // namespace

int main(int argc, char** argv) try {
    if (argc < 4 || std::string_view(argv[1]) != "--vfr") {
        std::printf("SKIP: verifier probe, run `test_verify_freeze_resume --vfr CASES n|g`\n");
        return D2D_TEST_SKIPPED;
    }
    const bool game_order = std::string_view(argv[3]) == "g";
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    if (!mpq_dir) { std::printf("SKIP: D2_MPQ_DIR\n"); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir), patch ? fs::path(patch) : fs::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("SKIP: needs the 1.14d patch\n"); return D2D_TEST_SKIPPED; }

    d2d::rules::Rng rng(11);
    auto made = new_character(*data, 1, "Probe", false, true, rng);
    Character character;
    character.character_class = 1;
    character.name = "Probe";
    character.header = made.header;
    character.stats = made.stats;
    character.items = made.items;
    character.panel = panel_stats(*data, character.header, character.items, character.stats);
    World world(&*data, -1, -1);
    world.enter(character);

    const auto body_it = data->monsters.by_id.find("zombie1");
    const auto spike = data->skills.by_name.find("Glacial Spike");
    if (body_it == data->monsters.by_id.end() || spike == data->skills.by_name.end()) { std::printf("SKIP: rows\n"); return D2D_TEST_SKIPPED; }
    auto& body = const_cast<d2d::rules::MonType&>(data->monsters.types[body_it->second]);
    const auto saved_diff = body.diff;
    const bool saved_boss_column = body.boss_column;
    GameData::MissileInfo info;
    info.name = "vfr";
    info.etype = 2;
    info.to_hit = false;
    Missile missile;
    missile.info = &info;
    missile.skill = int(spike->second);
    missile.fixed = 0;
    missile.friendly = true;
    Level here{};
    here.id = d2d::rules::level_ids::kBloodMoor;
    const Crowd crowd;
    std::vector<Missile> shots;
    constexpr int kBase = 3000, kHorizon = 1100;

    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        if (line.empty() || line[0] == '#') continue;
        std::istringstream in(line);
        int id, diff, kind, pending, stun_at, stun_len;
        std::string row;
        if (!(in >> id >> diff >> row >> kind >> pending >> stun_at >> stun_len)) continue;
        std::vector<std::pair<int, int>> freezes;
        for (int at, len; in >> at >> len;) freezes.emplace_back(at, len);
        const auto src = data->monsters.by_id.find(row);
        if (src == data->monsters.by_id.end()) { std::printf("%d NOROW -\n", id); continue; }
        body.diff[std::size_t(diff)].aidel = data->monsters.types[src->second].diff[std::size_t(diff)].aidel;
        body.boss_column = kind == 5;
        int start = pending;
        if (stun_len > 0) start = std::min(start, stun_at);
        for (const auto& f : freezes) start = std::min(start, f.first);
        start -= 2;
        d2d::rules::Rng spawn{ 0x51e00000u ^ std::uint32_t(id) };
        Monster monster = make_monster(*data, int(body_it->second), 10.f, 10.f, spawn, diff);
        monster.boss = kind >= 1 && kind <= 4 ? d2d::rules::Boss(kind) : d2d::rules::Boss::none;
        monster.stats.hit_points = monster.hit_points = 1000000;
        monster.life_frac = 0;
        monster.aware = true;
        monster.next_act = std::uint32_t(kBase + pending) * kTickMs;
        world.fight.game_difficulty = diff;
        world.fight.monsters.clear();
        world.fight.add_monster(monster);
        auto& m = world.fight.monsters[0];
        std::array<Foe, 1> foes{};
        foes[0].x = 10.4f;
        foes[0].y = 10.f;
        d2d::rules::Rng think_rng{ 0x7a110000u ^ std::uint32_t(id) };
        int swing = 0, think = 0;
        bool have_swing = false, have_think = false;
        const auto strikes = [&](int frame, std::uint32_t now) {
            if (stun_len > 0 && frame == stun_at) write_stun(m, now + std::uint32_t(stun_len) * kTickMs);
            for (const auto& f : freezes)
                if (f.first == frame) world.fight.strike(missile, 0, now, f.second < 0 ? 0 : f.second);   // -1: zero frames after the divisor
        };
        for (int frame = start; frame <= kHorizon && !(have_swing && have_think); ++frame) {
            const std::uint32_t now = std::uint32_t(kBase + frame) * kTickMs;
            if (game_order) strikes(frame, now);
            const std::string_view before_mode = m.mode;   // mode names are static strings
            const std::uint32_t before_next = m.next_act, before_stun = m.stun_until;
            foes[0].alive = true;
            monster_update(*data, here, m, foes, think_rng, now, 0.04f, crowd, shots, world.fight.monsters);
            const int aidel = data->monsters.types[body_it->second].diff[std::size_t(diff)].aidel;
            const bool refile = before_stun != 0 && now >= before_stun && m.next_act == before_stun + std::uint32_t(aidel > 0 ? aidel : 15) * kTickMs;
            if (!have_swing && m.mode != before_mode && (m.mode == "A1" || m.mode == "A2")) { swing = frame; have_swing = true; }
            if (!have_think && m.next_act != before_next && !refile) { think = frame; have_think = true; }
            if (!game_order) strikes(frame, now);
        }
        const auto out = [](bool have, int v) { return have ? std::to_string(v) : std::string("-"); };
        std::printf("%d %s %s\n", id, out(have_swing, swing).c_str(), out(have_think, think).c_str());
        body.diff = saved_diff;
        body.boss_column = saved_boss_column;
    }
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "%s\n", error.what());
    return 1;
}
