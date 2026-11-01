// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's champion / unique / minion stat mods on the cases tools/emu/umod_stats.py
// writes (d2lod-bot, card decompmods). Calls d2d's own rules::boss_stats
// (components/rules/uniques.hpp, included from an upstream checkout) and
// applies it the way game::make_boss does (components/game/ai.cpp, upstream
// main 72132b99, copied line for line below), on the case's fixed base (level,
// exp 1000, life 1000, defence 100, the row's resistances), so each field
// compares with game.exe's. A minion gets no mods, as d2d tags it
// (objgroups.cpp: tag(.., Boss::minion) with none).
//
//   g++ -O2 -std=c++20 -I<upstream>/components/rules -o umod_stats_harness umod_stats_harness.cpp
//   ./umod_stats_harness < game.tsv > d2d.tsv
//
// Fields not modelled by d2d print `*`: armor% (d2d folds it into def, the
// effective defence; --compare folds game.exe's the same way) and regen
// (d2d has no monster life regeneration).
#include "uniques.hpp"

#include <cstdio>
#include <sstream>
#include <string>
#include <vector>

int main() {
    using namespace d2d::rules;
    UMods umods;
    std::vector<std::array<int, 3>> dm;
    std::string line;
    char buf[4096];
    while (std::fgets(buf, sizeof buf, stdin)) {
        line = buf;
        std::istringstream in(line);
        std::string tag;
        in >> tag;
        if (tag == "#k") { for (auto& value : umods.constants) in >> value; continue; }
        if (tag == "#lvl") { int level; in >> level; if (std::size_t(level) >= dm.size()) dm.resize(std::size_t(level) + 1); in >> dm[std::size_t(level)][0] >> dm[std::size_t(level)][1] >> dm[std::size_t(level)][2]; continue; }
        if (tag.empty() || tag[0] == '#') continue;
        int row, diff, base_level, velocity, base_id;
        std::string kind_s, mods_s;
        std::array<int, 6> res{};
        in >> row >> diff >> kind_s >> mods_s >> base_level >> velocity >> base_id;
        for (auto& value : res) in >> value;
        MonType type;
        type.velocity = velocity;
        type.base = base_id;
        type.diff[std::size_t(diff)].res = res;
        std::vector<int> mods;
        if (mods_s != "-") { std::istringstream m(mods_s); std::string id; while (std::getline(m, id, '.')) mods.push_back(std::stoi(id)); }
        const Boss kind = kind_s == "c" ? Boss::champion : kind_s == "u" ? Boss::unique : kind_s == "s" ? Boss::superunique : Boss::minion;
        if (kind == Boss::minion) mods.clear();
        const auto boss = boss_stats(umods, type, kind, mods, diff);
        // make_boss (ai.cpp 2923..2958) on the fixed base.
        const int level = std::max(base_level + boss.level_add, 1);
        int hit_points = 1000, armor_class = 100, exp = 1000;
        hit_points += hit_points * boss.hp_pct / 100;
        hit_points = std::max(hit_points + hit_points * boss.life_after_pct / 100, 1);
        armor_class += armor_class * boss.defense_pct / 100;
        std::array<int, 6> boss_res{};
        if (boss.phys_resist >= 0) boss_res[0] = boss.phys_resist - res[0];
        exp *= boss.exp_mult;
        if (boss.double_defense) armor_class *= 2;
        if (boss.phys_resist < 0) boss_res = boss.res_add;
        else { const int set = boss_res[0]; boss_res = boss.res_add; boss_res[0] = set; }
        std::array<int, 6> el{};                    // fire lo hi, ltng lo hi, cold lo hi
        int cold_len = 0;
        const auto& level_row = dm[std::min<std::size_t>(std::size_t(level), dm.size() - 1)];
        const int damage_pct = level_row[std::size_t(diff)];
        if (boss.elem >= 0) {
            const int lo = damage_pct * boss.elem_min_pct / 100, hi = std::max(damage_pct * boss.elem_max_pct / 100, lo);
            const int slot = boss.elem == 0 ? 0 : boss.elem == 1 ? 2 : 4;
            el[std::size_t(slot)] = lo; el[std::size_t(slot) + 1] = hi;
            if (boss.elem == 2) cold_len = boss.elem_len;
        }
        int mana_lo = 0, mana_hi = 0;
        if (std::ranges::contains(mods, umod::manahit)) {
            const int base = kind == Boss::minion ? 16 : 28;
            mana_lo = damage_pct * umods.constants[std::size_t(base + diff)] / 100;
            mana_hi = std::max(damage_pct * umods.constants[std::size_t(base + 3 + diff)] / 100, mana_lo);
        }
        std::printf("%s\t%d\t%d\t%d\t%d\t%d\t%d", tag.c_str(), level, exp, hit_points << 8, boss.dmg_pct, boss.tohit_pct, boss.velocity_pct);
        for (int i = 0; i < 6; ++i) std::printf("\t%d", res[std::size_t(i)] + boss_res[std::size_t(i)]);
        std::printf("\t%d\t*\t*", armor_class);
        for (const int value : el) std::printf("\t%d", value);
        std::printf("\t%d\t%d\t%d\n", cold_len, mana_lo << 8, mana_hi << 8);
    }
}
