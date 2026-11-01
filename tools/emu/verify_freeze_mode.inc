// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// Verifier (card verifyfreeze), copied to tests/game/test_verify_freeze.cpp in a throwaway
// build tree of each commit: `--verify-freeze SEED N`, the cases of tools/emu/verify_freeze.py
// (same generator, seed masked to 32 bits) through the REAL Fight::strike with a freeze length
// (the hit_func 13 path), one line each: the case and `chill freeze`, the ticks from now each
// ends, or `-`. Builds on every commit of the range: the freeze divisor is only set where
// GameData has it.

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

#include <array>
#include <cstdint>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <string>
#include <string_view>
#include <vector>

using namespace d2d::game;

namespace {

struct Case {
    std::string id;
    int diff = 0, cold_effect = 0, bcol = 0, kind = 0, half = 0, res = 0, pierce = 0, roll = 0, freeze = 0, prior_chill = 0, prior_freeze = 0, fdiv = 1, cdiv = 1;
};

constexpr std::array<int, 3> kFdiv{ 1, 2, 4 }, kCdiv{ 1, 2, 4 };

struct Xorshift {
    std::uint32_t x;
    std::uint32_t operator()() { x ^= x << 13; x ^= x >> 17; x ^= x << 5; return x; }
};

std::vector<Case> cases(std::uint32_t n, std::uint32_t seed) {
    std::vector<Case> out;
    int i = 0;
    auto edge = [&](int diff, int ce, int bcol, int kind, int half, int res, int pierce, int roll, int fr, int pc, int pf, int fdiv, int cdiv) {
        out.push_back({ "E" + std::to_string(i++), diff, ce, bcol, kind, half, res, pierce, roll, fr, pc, pf, fdiv, cdiv });
    };
    std::vector<std::array<int, 2>> kinds_half{ { 0, 0 }, { 1, 0 }, { 2, 0 }, { 4, 0 }, { 5, 0 } };
    for (const int kind : { 3, 6 }) for (const int half : { 0, 1 }) kinds_half.push_back({ kind, half });
    for (int diff = 0; diff < 3; ++diff) {
        const auto fd = kFdiv[std::size_t(diff)], cd = kCdiv[std::size_t(diff)];
        for (const int ce : { -50, -1, 0, 1, 25 })
            for (const auto& [kind, half] : kinds_half)
                for (const int bcol : { 0, 1 })
                    for (const int res : { -150, -100, -1, 0, 1, 50, 99, 100, 101 })
                        for (const int fr : { 0, 1, 2, 3, 4, 5, 7, 50, 100, 375 }) edge(diff, ce, bcol, kind, half, res, 0, 256, fr, 0, 0, fd, cd);
        for (const int pierce : { 1, 40, 100 })
            for (const int res : { 50, 99, 100, 150 })
                for (const int fr : { 50, 150 })
                    for (const int kind : { 0, 1 }) edge(diff, -50, 0, kind, 0, res, pierce, 256, fr, 0, 0, fd, cd);
        for (const int pc : { 0, 1, 37, 1000 })
            for (const int pf : { 0, 1, 37, 1000 })
                for (const int fr : { 0, 2, 50, 375 })
                    for (const int kind : { 0, 1, 4 })
                        for (const int ce : { -50, 25 }) edge(diff, ce, 0, kind, 0, 0, 0, 256, fr, pc, pf, fd, cd);
    }
    for (const int fdiv : { 1, 2, 3, 5, 7 })
        for (const int cdiv : { 1, 2, 3, 4 })
            for (const int fr : { 1, 2, 3, 4, 5, 9, 50, 101 })
                for (const int ce : { -50, -1, 25 })
                    for (const int kind : { 0, 1 }) edge(0, ce, 0, kind, 0, 0, 0, 256, fr, 0, 0, fdiv, cdiv);
    for (const int fr : { 0xFFFFF, 0x100000, 0x100001, 10000000 })
        for (const int res : { -100, -50, 0, 50 })
            for (const int kind : { 0, 1 }) edge(0, -50, 0, kind, 0, res, 0, 256, fr, 0, 0, 1, 1);
    for (const int kind : { 0, 1, 4 }) edge(1, -50, 0, kind, 0, 0, 0, 0, 50, 0, 0, kFdiv[1], kCdiv[1]);
    Xorshift rand{ seed != 0 ? seed : 1 };
    constexpr std::array<int, 8> kEffects{ -50, -50, -50, -33, -25, 0, 10, -100 };
    for (std::uint32_t k = 0; k < n; ++k) {
        Case c;
        c.id = "R" + std::to_string(k);
        c.diff = int(rand() % 3);
        if (rand() % 4 != 0) c.cold_effect = kEffects[rand() % 8];
        else c.cold_effect = int(rand() % 201) - 100;
        c.bcol = rand() % 8 == 0 ? 1 : 0;
        c.kind = int(rand() % 7);
        if (c.kind == 3 || c.kind == 6) c.half = int(rand() % 2);
        else c.half = rand() % 16 == 0 ? 1 : 0;
        const auto r = rand() % 100;
        if (r < 10) c.res = 100 + int(rand() % 51);
        else if (r < 25) c.res = -int(rand() % 151);
        else c.res = int(rand() % 100);
        if (c.res >= 100 || rand() % 4 != 0) c.pierce = 0;
        else c.pierce = int(rand() % 61);
        c.roll = 1 + int(rand() % 0x4000);
        if (rand() % 10 == 0) c.freeze = 0;
        else c.freeze = 1 + int(rand() % 500);
        if (rand() % 10 < 7) c.prior_chill = 0;
        else c.prior_chill = 1 + int(rand() % 600);
        if (rand() % 10 < 7) c.prior_freeze = 0;
        else c.prior_freeze = 1 + int(rand() % 600);
        if (rand() % 5 != 0) c.fdiv = kFdiv[std::size_t(c.diff)];
        else c.fdiv = 1 + int(rand() % 8);
        if (rand() % 5 != 0) c.cdiv = kCdiv[std::size_t(c.diff)];
        else c.cdiv = 1 + int(rand() % 8);
        out.push_back(c);
    }
    return out;
}

template <typename Data> void set_freeze_divisor(Data& data, int diff, int value) {
    if constexpr (requires { data.freeze_divisor; }) data.freeze_divisor[std::size_t(diff)] = value;
}

// game.exe's flag words per kind map to these d2d kinds (verify_freeze.py's FLAGS).
constexpr std::array<d2d::rules::Boss, 7> kBoss{ d2d::rules::Boss::none, d2d::rules::Boss::champion, d2d::rules::Boss::unique, d2d::rules::Boss::superunique,
                                                 d2d::rules::Boss::minion, d2d::rules::Boss::champion, d2d::rules::Boss::unique };

}  // namespace

int main(int argc, char** argv) try {
    if (argc < 4 || std::string_view(argv[1]) != "--verify-freeze") {
        std::printf("SKIP: an oracle mode, run `test_verify_freeze --verify-freeze SEED N`\n");
        return D2D_TEST_SKIPPED;
    }
    const auto seed = std::uint32_t(std::strtoull(argv[2], nullptr, 0) & 0xFFFFFFFFULL);
    const auto n = std::uint32_t(std::strtoul(argv[3], nullptr, 0));
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    if (!mpq_dir || !fs::exists(fs::path(mpq_dir) / "d2data.mpq")) { std::printf("SKIP: set D2_MPQ_DIR\n"); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir), patch ? fs::path(patch) : fs::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("SKIP: set D2_PATCH_INSTALLER\n"); return D2D_TEST_SKIPPED; }
    auto& game_data = const_cast<GameData&>(*data);
    d2d::rules::Rng rng(7);
    constexpr int kSorceress = 1;
    auto made = new_character(game_data, kSorceress, "Verifier", false, true, rng);
    Character character;
    character.character_class = kSorceress;
    character.name = "Verifier";
    character.header = made.header;
    character.stats = made.stats;
    character.items = made.items;
    character.panel = panel_stats(game_data, character.header, character.items, character.stats);
    const auto plain = game_data.monsters.by_id.find("fallen1");
    const auto spike = game_data.skills.by_name.find("Glacial Spike");
    if (plain == game_data.monsters.by_id.end() || spike == game_data.skills.by_name.end()) { std::fprintf(stderr, "no fallen1 / Glacial Spike\n"); return 1; }
    auto& type = game_data.monsters.types[plain->second];
    const auto saved_diff = type.diff;
    const bool saved_bcol = type.boss_column;
    const auto saved_cdiv = game_data.cold_divisor;
    GameData::MissileInfo info;
    info.name = "verifyfreeze";
    info.etype = 2;
    info.to_hit = false;
    World world(&game_data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    d2d::rules::Rng spawn_rng{ 5 };
    const Monster base = make_monster(game_data, int(plain->second), world.actor->player.x, world.actor->player.y, spawn_rng, 0);
    const std::uint32_t now = 3000 * kTickMs;
    for (const auto& c : cases(n, seed)) {
        *world.actor = start;
        world.actor->psum[d2d::d2s::kPassiveColdPierce] = c.pierce;
        world.fight.game_difficulty = c.diff;
        type.diff = saved_diff;
        type.diff[std::size_t(c.diff)].cold_effect = c.cold_effect;
        type.boss_column = c.bcol != 0;
        game_data.cold_divisor[std::size_t(c.diff)] = c.cdiv;
        set_freeze_divisor(game_data, c.diff, c.fdiv);
        Monster monster = base;
        monster.difficulty = c.diff;
        monster.boss = kBoss[std::size_t(c.kind)];
        monster.half_freeze = c.half != 0;
        monster.stats.hit_points = monster.hit_points = 100000000;
        monster.life_frac = 0;
        monster.boss_res = {};
        monster.boss_res[4] = c.res - monster.target(game_data).res[4];
        monster.chill_until = c.prior_chill > 0 ? now + std::uint32_t(c.prior_chill) * kTickMs : 0;
        monster.stun_until = c.prior_freeze > 0 ? now + std::uint32_t(c.prior_freeze) * kTickMs : 0;
        world.fight.monsters.clear();
        world.fight.add_monster(monster);
        const bool res_ok = world.fight.target_of(0).res[4] == c.res;
        Missile missile;
        missile.info = &info;
        missile.skill = int(spike->second);
        missile.fixed = c.roll;
        missile.friendly = true;
        world.fight.strike(missile, 0, now, c.freeze);
        const auto& struck = world.fight.monsters.back();
        const auto ticks = [&](std::uint32_t until) { return until > now ? std::to_string((until - now) / kTickMs) : std::string("-"); };
        std::printf("%s %d %d %d %d %d %d %d %d %d %d %d %d %d %s %s%s\n", c.id.c_str(), c.diff, c.cold_effect, c.bcol, c.kind, c.half, c.res, c.pierce, c.roll, c.freeze,
                    c.prior_chill, c.prior_freeze, c.fdiv, c.cdiv, ticks(struck.chill_until).c_str(), ticks(struck.stun_until).c_str(), res_ok ? "" : " RESFAIL");
    }
    type.diff = saved_diff;
    type.boss_column = saved_bcol;
    game_data.cold_divisor = saved_cdiv;
    return 0;
} catch (const std::exception& error) {
    std::fprintf(stderr, "%s\n", error.what());
    return 1;
}
