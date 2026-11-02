// SPDX-License-Identifier: GPL-3.0-or-later
// mercstats: the d2d side of tools/emu/mst_oracle.py. Calls d2d's own rules (components/rules/rules.hpp,
// merc.hpp) on the real loaded tables (load_game_data). Built by hand against a scratch upstream build
// (tools/emu/mst_build.sh), never part of upstream.
//   mst_d2d_test hire FIRST-LAST   rules::merc_offer, as tools/drlg-dump hire prints it (hire.py's lines)
//   mst_d2d_test revive            rules::merc_resurrect_cost for merc level 0..255
//   mst_d2d_test levelup           rules::merc_stats (from the exact experience of each level) and
//                                  rules::merc_ski[l]l_levels for every LoD hireling Id x level 1..98

#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <merc.hpp>
#include <rules.hpp>

#include <algorithm>
#include <cstdio>
#include <cstdlib>
#include <exception>
#include <filesystem>
#include <format>
#include <fstream>
#include <iostream>
#include <set>
#include <sstream>
#include <string>
#include <string_view>

using namespace d2d::game;

int main(int argc, char** argv) try {
    if (argc < 2) return 2;
    const std::string_view mode = argv[1];
    namespace R = d2d::rules;
    if (mode == "revive") {
        for (int level = 0; level < 256; ++level) std::printf("%d %d\n", level, R::merc_resurrect_cost(level));
        return 0;
    }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(fs::path(mpq_dir), patch ? fs::path(patch) : fs::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("FAIL: no patched data\n"); return 1; }
    const auto& tables = data->rules;
    if (mode == "hire") {
        const std::string range = argv[2];
        const auto dash = range.find('-');
        const auto first = std::uint32_t(std::stoul(range.substr(0, dash), nullptr, 0));
        const auto last = dash == std::string::npos ? first : std::uint32_t(std::stoul(range.substr(dash + 1), nullptr, 0));
        for (auto index = first;; ++index) {
            const std::uint32_t seed = index * 0x9e3779b1u;
            for (int act : { 0, 1, 2 })
                for (int diff : { 0, 1, 2 })
                    for (int clvl : { 1, 9, 30, 75 }) {
                        std::printf("%08x %d %d %d: ", seed, act, diff, clvl);
                        const auto offer = R::merc_offer(tables, true, act, diff, clvl, R::HireEntry{ 0, seed });
                        if (!offer) { std::puts("none"); continue; }
                        std::printf("%d %d %d %d %d %d %u %d %d %d\n", offer->id, offer->level, offer->life, offer->str, offer->dex, offer->cost, offer->exp,
                                    offer->def, offer->dmg_min, offer->dmg_max);
                    }
            if (index == last) break;
        }
        return 0;
    }
    if (mode == "think") {   // stdin: tools/emu/mst_think.py lines; stdout: the same lines with d2d's effects
        std::string line;
        std::ifstream file;
        if (argc > 2) file.open(argv[2]);
        if (argc > 3 && !std::freopen(argv[3], "w", stdout)) return 1;   // results to a file: the loader's log stays out of it
        std::istream& source = argc > 2 ? static_cast<std::istream&>(file) : std::cin;
        while (std::getline(source, line)) {
            const auto cut = line.find(" ::: ");
            if (line.rfind("T ", 0) != 0 || cut == std::string::npos) continue;
            std::istringstream in(line.substr(2, cut - 2));
            R::MercView view;
            int on40 = 0, town = 0, split = 0;
            unsigned lo = 0, hi = 0;
            std::string bits;
            in >> view.cls >> view.size >> view.mode >> view.x >> view.y >> view.owner_x >> view.owner_y >> view.owner_mode >> view.end_x >> view.end_y
               >> view.cursor >> on40 >> town >> view.target >> split >> lo >> hi >> bits;
            for (auto& spot : view.ring) in >> spot.first >> spot.second;
            view.on_bit40 = on40 != 0; view.town = town != 0;
            R::Rng seed; seed.low = lo; seed.high = hi;
            std::string out;
            std::size_t tried = 0;
            const auto area_at = [&](int, int, int x, int) { return x >= split ? 2 : 1; };
            const auto try_move = [&](const R::MercMove& move) {
                const bool ok = tried < bits.size() && bits[tried++] == '1';
                if (!out.empty()) out += ';';
                out += std::format("move {} {} {} {} {} {} {}", move.x & 0xffff, move.y & 0xffff, move.mode, move.type, move.pct, move.steps, int(ok));
                return ok;
            };
            const auto act = R::hireable_think(view, seed, area_at, try_move);
            using Kind = R::MercAct::Kind;
            const std::string tail = act.kind == Kind::stand ? std::format("stand {}", act.frames) : act.kind == Kind::teleport ? "teleport"
                                   : act.kind == Kind::attack ? "attack" : "";
            if (!tail.empty()) { if (!out.empty()) out += ';'; out += tail; }
            std::printf("%s ::: %s ::: %u %u\n", line.substr(0, cut).c_str(), out.c_str(), seed.low, seed.high);
        }
        return 0;
    }
    if (mode == "attack") {   // a file of tools/emu/mst_attack.py lines (argv[2]), results to argv[3]
        std::string line;
        std::ifstream file(argv[2]);
        if (argc > 3 && !std::freopen(argv[3], "w", stdout)) return 1;
        while (std::getline(file, line)) {
            const auto cut = line.find(" ::: ");
            if (line.rfind("A ", 0) != 0 || cut == std::string::npos) continue;
            std::istringstream in(line.substr(2, cut - 2));
            R::MercFightView view;
            R::Hireling row;
            int melee = 0, growth = 0, row_level = 0, default_chance = 0;
            unsigned lo = 0, hi = 0;
            std::string bits;
            in >> view.cls >> view.level >> view.aip1 >> view.gap >> melee >> growth >> lo >> hi >> row_level >> default_chance;
            row.level = row_level; row.default_chance = default_chance;
            for (auto& id : view.ids) in >> id;
            for (auto& skill : row.skills) in >> skill.chance;
            for (auto& skill : row.skills) in >> skill.chance_per_level;
            for (auto& skill : row.skills) in >> skill.mode;
            for (auto& level : view.levels) in >> level;
            for (std::size_t k = 0; k < 6; ++k) { int flag = 0; in >> flag; view.aura[k] = flag != 0; }
            for (std::size_t k = 0; k < 6; ++k) { int flag = 0; in >> flag; view.running[k] = flag != 0; }
            in >> view.skill1 >> view.sk1mode >> bits;
            view.melee = melee != 0; view.row = &row; view.distance = view.gap + 1;
            R::Rng seed; seed.low = lo; seed.high = hi;
            std::string out;
            std::size_t tried = 0;
            const auto move = [&](int, int, int) {
                const bool ok = tried < bits.size() && bits[tried++] == '1';
                if (!out.empty()) out += ';';
                out += std::format("move {}", int(ok));
                return ok;
            };
            const auto result = R::merc_attack(view, growth, 5000, 5000, 5000 + view.gap + 1, 5000, seed, move);
            using Kind = R::MercAttack::Kind;
            std::string tail;
            switch (result.kind) {
            case Kind::skill: tail = std::format("skill {} {}", result.mode, result.skill); break;
            case Kind::aura: tail = std::format("aura {};stand 10", result.skill); break;
            case Kind::swing: tail = "swing"; break;
            case Kind::stand: tail = std::format("stand {}", result.frames); break;
            default: break;
            }
            if (!tail.empty()) { if (!out.empty()) out += ';'; out += tail; }
            std::printf("%s ::: %s ::: %d %u %u\n", line.substr(0, cut).c_str(), out.c_str(), growth, seed.low, seed.high);
        }
        return 0;
    }
    if (mode == "xp") {   // a file of tools/emu/mst_kill.py lines (argv[2]), results to argv[3]
        std::string line;
        std::ifstream file(argv[2]);
        if (argc > 3 && !std::freopen(argv[3], "w", stdout)) return 1;
        while (std::getline(file, line)) {
            const auto cut = line.find(" ::: ");
            if (line.rfind("K ", 0) != 0 || cut == std::string::npos) continue;
            std::istringstream in(line.substr(2, cut - 2));
            int hid = 0, blow = 0, merc_level = 0, owner_level = 0;
            long long exp = 0, award = 0;
            in >> hid >> exp >> award >> blow >> merc_level >> owner_level;
            const auto after = R::merc_kill_exp(std::uint32_t(exp), award, blow != 0, merc_level, owner_level);
            const int level = after == std::uint32_t(exp) ? merc_level : R::merc_stats(tables, hid, after).level;
            std::printf("%s ::: %u %d\n", line.substr(0, cut).c_str(), after, level);
        }
        return 0;
    }
    if (mode == "levelup") {
        std::set<int> ids;
        for (const auto& row : tables.hirelings) if (row.version == 100) ids.insert(row.id);
        for (const int id : ids) {
            const R::Hireling* first = nullptr;
            for (const auto& row : tables.hirelings) if (row.version == 100 && row.id == id) { first = &row; break; }
            for (int level = 1; level < 99; ++level) {
                const long long exp = (long long)(level + 1) * first->exp_per_level * level * level;
                const auto st = R::merc_stats(tables, id, std::uint32_t(exp));
                const auto* row = R::hireling_row(tables, id, level);
                std::array<int, 6> sk{}, levels{};
                if (row) {
                    for (std::size_t k = 0; k < 6; ++k) {
                        const auto found = data->skills.by_name.find(row->skills[k].name);
                        if (found != data->skills.by_name.end()) sk[k] = data->skills.get(found->second)->id;
                    }
                    levels = R::merc_skill_levels(*row, sk, level, [&](int sid) { const auto* s = data->skills.get(sid); return s ? s->reqlevel : 0x7fff; });
                }
                // d2d has no stat 0x4a for a merc: printed as 0 only so the columns line up (see D-regen in the trace).
                std::printf("%d %d str %d dex %d life %d def %d dmg %d %d ar %d res %d %d %d %d regen 0 skills %d %d %d %d %d %d lv %d\n", id, level, st.str, st.dex, st.life,
                            st.def, st.dmg_min, st.dmg_max, st.attack_rating, st.resist, st.resist, st.resist, st.resist, levels[0], levels[1], levels[2],
                            levels[3], levels[4], levels[5], st.level);
            }
        }
        return 0;
    }
    return 2;
} catch (const std::exception& error) {
    std::printf("FAIL: %s\n", error.what());
    return 1;
}
