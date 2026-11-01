// SPDX-License-Identifier: GPL-3.0-or-later
// test-links: game
// `--sk-mana CASES`: a player's skill mana check and payment, the cases of
// tools/emu/sk_mana.py (game.exe's FUN_0056c160 check, then FUN_0056bfe0's
// payment), one line each, to diff: `id check paid` with d2d's rule (every
// cast path in components/game/fight: `lvl > 0 && mana >= cost_of`, then
// `pay` takes rules::mana_cost) on the loaded patched Skills.txt.

#include <gamedata.hpp>
#include <gamedata_load.hpp>
#include <skills.hpp>

#include <cassert>
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
    if (argc < 3 || std::string_view(argv[1]) != "--sk-mana") {
        std::printf("SKIP: an oracle mode, run `test_sk_mana --sk-mana CASES` (see tools/emu/sk_mana.py)\n");
        return D2D_TEST_SKIPPED;
    }
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    const char* home = std::getenv("HOME");
    const fs::path data_dir = mpq_dir ? fs::path(mpq_dir) : fs::path(home ? home : "") / "Workspace/private/diablo2";
    if (!fs::exists(data_dir / "d2data.mpq")) { std::printf("SKIP: no d2data.mpq in %s\n", data_dir.string().c_str()); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = load_game_data(data_dir, patch ? fs::path(patch) : fs::path{}, 0x1234);
    assert(data && data->patched);

    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        long long case_id = 0, mana = 0;
        int skill_id = 0, lvl = 0;
        fields >> case_id >> skill_id >> lvl >> mana;
        const auto* skill = data->skills.get(skill_id);
        int check = 0;
        long long paid = 0;
        if (skill) {
            const int cost = d2d::rules::mana_cost(*skill, lvl);
            check = lvl > 0 && mana >= cost ? 1 : 0;
            paid = check ? cost : 0;
        }
        std::printf("%lld %d %lld\n", case_id, check, paid);
    }
    return 0;
} catch (const std::exception& error) {
    std::printf("FAIL: %s\n", error.what());
    return 1;
}
