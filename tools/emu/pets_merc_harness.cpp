// SPDX-License-Identifier: GPL-3.0-or-later
// d2d's side of tools/emu/pets_merc.py: reads its case lines, runs d2d's
// own rules code (components/rules/monsters.hpp: search_pick, search_sight,
// sight_choice) and prints the lines game.exe's side prints.
//   S: ai.cpp search_target's evil-monster branch (ai.cpp:546..563 at
//      upstream 72132b99): a dead pet entry scores 0x7fffffff, then
//      rules::search_pick over the foes in fight/tick.cpp's order (player,
//      merc, pets oldest first). `after` tries the pets newest first
//      (game.exe's FUN_005b1900 inserts each right after the player), the
//      proposed order; `before` keeps d2d's.
//   M: merc.cpp's search (merc.cpp:302..330): rules::search_sight, then
//      sight_choice with the case's path answer.
// Build: g++ -O2 -std=c++20 -I<d2d>/components -I<d2d>/components/d2s pets_merc_harness.cpp
#include "rules/monsters.hpp"

#include <cstdio>
#include <iostream>
#include <sstream>
#include <string>
#include <vector>

int main(int argc, char** argv) {
    const bool after = argc > 1 && std::string(argv[1]) == "after";
    std::string line;
    while (std::getline(std::cin, line)) {
        std::istringstream in(line);
        std::string kind;
        long long index = 0;
        in >> kind >> index;
        if (kind == "S") {
            int best = 0, sight = 0, count = 0;
            in >> best >> sight >> count;
            std::vector<int> ids;
            std::vector<d2d::rules::SearchFoe> foes;
            for (int k = 0; k < count; ++k) {
                int id = 0, distance = 0, pet = 0, away = 0, dead = 0, blocked = 0;
                in >> id >> distance >> pet >> away >> dead >> blocked;
                d2d::rules::SearchFoe foe{ distance, pet != 0, away != 0, dead != 0, false, 0 };
                if (foe.pet && foe.dead) foe.distance = 0x7fffffff;    // ai.cpp:558
                foe.blocked = sight && (foe.list || foe.distance < best) && blocked;   // ai.cpp:559 (sight_blocked from the case)
                ids.push_back(id);
                foes.push_back(foe);
            }
            if (after && foes.size() > 2) {                          // player, then the pets newest first
                std::reverse(foes.begin() + 1, foes.end());
                std::reverse(ids.begin() + 1, ids.end());
            }
            const auto pick = d2d::rules::search_pick(foes, best, sight != 0);
            std::printf("S %lld %d %d\n", index, pick.target >= 0 ? ids[std::size_t(pick.target)] : 0, pick.target >= 0 ? pick.best : pick.nearest);
        } else if (kind == "M") {
            int path = 0, count = 0;
            in >> path >> count;
            std::vector<int> ids;
            std::vector<d2d::rules::SightFoe> foes;
            for (int k = 0; k < count; ++k) {
                int id = 0, distance = 0, threat = 0, enemy = 0, blocked = 0;
                in >> id >> distance >> threat >> enemy >> blocked;
                d2d::rules::SightFoe foe{ distance, threat, enemy != 0, false };
                foe.blocked = foe.enemy && foe.distance < 0x31 && blocked;   // merc.cpp:312
                ids.push_back(id);
                foes.push_back(foe);
            }
            int distance = 0x7fffffff;
            const int chosen = d2d::rules::sight_choice(std::span<const d2d::rules::SightFoe>(foes), d2d::rules::search_sight(foes), [&](int) { return path != 0; }, distance);
            std::printf("M %lld %d %d\n", index, chosen >= 0 ? ids[std::size_t(chosen)] : 0, chosen >= 0 ? distance : 0x7fffffff);
        }
    }
}
