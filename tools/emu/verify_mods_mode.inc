// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier (card verifymods): the d2d side of tools/emu/verify_mods.py. Paste into a test
// (tests/game/test_verify_mods.cpp: `// test-links: game`, includes ai.hpp gamedata.hpp
// gamedata_load.hpp rules.hpp uniques.hpp) and call verify_mods(argc, argv) from main.
// `--verify-mods SEED N`: the same cases from SEED & 0xffffffff, each through the REAL
// make_boss (rules::boss_stats, rules::monster_stats) or, for a plain one, monster_stats;
// the unit seed is the Rng's low word, as the spawn's life roll draws it first.
// Prints `i row diff kind mods seed | level life def exp vel`, or `| -` when the row has no
// Level for that difficulty.

namespace verify_mods_impl {
inline std::uint32_t xorshift(std::uint32_t& x) {
    x ^= x << 13;
    x ^= x >> 17;
    x ^= x << 5;
    return x;
}
}  // namespace verify_mods_impl

inline int verify_mods(int argc, char** argv) {
    using verify_mods_impl::xorshift;
    if (argc < 4 || std::string_view(argv[1]) != "--verify-mods") return -1;
    std::uint32_t state = std::uint32_t(std::stoull(argv[2], nullptr, 0) & 0xffffffffULL);
    if (state == 0) state = 1;
    const long count = std::stol(argv[3]);
    const char* mpq_dir = std::getenv("D2_MPQ_DIR");
    if (!mpq_dir) { std::printf("SKIP: set D2_MPQ_DIR\n"); return D2D_TEST_SKIPPED; }
    const char* patch = std::getenv("D2_PATCH_INSTALLER");
    auto data = d2d::game::load_game_data(std::filesystem::path(mpq_dir), patch ? std::filesystem::path(patch) : std::filesystem::path{}, 0x1234);
    if (!data || !data->patched) { std::printf("SKIP: needs the 1.14d data (D2_PATCH_INSTALLER)\n"); return D2D_TEST_SKIPPED; }
    static constexpr std::string_view kinds = "uscmp";
    static constexpr std::array<int, 5> champ{ 16, 36, 37, 38, 39 };
    static constexpr std::array<int, 9> umods{ 5, 6, 8, 9, 17, 18, 25, 27, 28 };
    static constexpr std::array<int, 16> edge_rows{ 0, 1, 2, 3, 5, 19, 156, 229, 250, 300, 400, 500, 571, 650, 700, 733 };
    constexpr std::uint32_t rows = 734;
    const long edges = 3 * 16 * long(edge_rows.size());
    for (long i = 0; i < count; ++i) {
        const int row = i < edges ? edge_rows[std::size_t(i) % edge_rows.size()] : int(xorshift(state) % rows);
        const int diff = i < edges ? int((i / long(edge_rows.size())) % 3) : int(xorshift(state) % 3);
        const char kind = kinds[xorshift(state) % kinds.size()];
        std::vector<int> mods;
        if (kind == 'c') {
            mods.push_back(champ[xorshift(state) % champ.size()]);
        } else if (kind == 'u' || kind == 's') {
            const std::uint32_t picks = xorshift(state) % 4;
            for (std::uint32_t pick = 0; pick < picks; ++pick) {
                const int mod = umods[xorshift(state) % umods.size()];
                if (std::ranges::find(mods, mod) == mods.end()) mods.push_back(mod);
            }
        }
        const std::uint32_t unit_seed = xorshift(state);
        std::string mod_text;
        for (const int mod : mods) mod_text += (mod_text.empty() ? "" : ".") + std::to_string(mod);
        std::printf("%ld %d %d %c %s %u | ", i, row, diff, kind, mod_text.empty() ? "-" : mod_text.c_str(), unit_seed);
        if (std::size_t(row) >= data->monsters.types.size() || data->monsters.types[std::size_t(row)].level[std::size_t(diff)] == 0) {
            std::printf("-\n");
            continue;
        }
        d2d::rules::Rng look_rng{ 1 };
        auto monster = d2d::game::make_monster(*data, row, 0, 0, look_rng, diff, false);
        d2d::rules::Rng rng{ unit_seed };
        if (kind == 'p') {
            monster.stats = d2d::rules::monster_stats(data->monsters, row, diff, rng);
            monster.hit_points = monster.stats.hit_points;
        } else {
            const auto boss = kind == 'c' ? d2d::rules::Boss::champion : kind == 'u' ? d2d::rules::Boss::unique
                            : kind == 's' ? d2d::rules::Boss::superunique : d2d::rules::Boss::minion;
            d2d::game::make_boss(*data, monster, boss, mods, -1, 0, diff, rng);
        }
        std::printf("%d %d %d %d %d\n", monster.stats.level, monster.hit_points, monster.stats.armor_class, monster.stats.exp, monster.boss_speed);
    }
    return 0;
}
