// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_game (d2lod-bot card verifyprs): a unit's walk % through the REAL
// d2d code of the build under test. Pasted into tests/game/test_game.cpp right after
// `character.panel = panel_stats(...)` in a throwaway worktree only (VERIFY_VP in the after build).
//   test_game --verify-walk CASES  -> `id pct` per case (pct to 3 decimals) or `id skip why`
// kind 0, a monster: one frame of the real monster_update (components/game/ai.cpp) on a monster in
//   WL or RN with a far spot to walk to in the Blood Moor's bounds with every subtile open; its walk % is the distance it moved
//   over MonStats Velocity's (cells_per_sec x 0.04 s). MonStats Velocity and coldeffect of the row
//   are overridden by the case (the loaded row, restored after).
// kind 1, a player: the real World::walk_percent(Player, now) with faster run/walk and
//   velocitypercent set as update_fighters leaves them (before the fix velocitypercent is folded
//   into frw there: strike.cpp:185-186 of upstream main, so the before build gets frw + vp).
if (argc > 2 && std::string_view(argv[1]) == "--verify-walk") {
    World world(&*data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    Level open = *data->level(d2d::rules::level_ids::kBloodMoor);   // its size, every subtile open
    std::ranges::fill(open.walk, std::uint16_t(0));
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        int id = 0, kind = 0;
        if (!(fields >> id >> kind)) continue;
        if (kind == 0) {
            std::string row_id;
            int mode = 0, difficulty = 0, velocity = 0, cold = 0, speed_pct = 0, boss_speed = 0, move_pct = 0, laid = 0, chilled = 0;
            if (!(fields >> row_id >> mode >> difficulty >> velocity >> cold >> speed_pct >> boss_speed >> move_pct >> laid >> chilled)) continue;
            const auto row = data->monsters.by_id.find(row_id);
            if (row == data->monsters.by_id.end()) { std::printf("%d skip norow\n", id); continue; }
            auto& type_info = const_cast<GameData&>(*data).monsters.types[std::size_t(row->second)];
            const auto saved_type = type_info;
            type_info.velocity = velocity;
            type_info.diff[std::size_t(difficulty)].cold_effect = cold;
            d2d::rules::Rng rng{ std::uint32_t(id + 1) };
            std::vector<Monster> pack{ make_monster(*data, int(row->second), 10.f, 10.f, rng, difficulty) };
            auto& monster = pack[0];
            monster.difficulty = difficulty;
            const std::uint32_t now = 1000 * 40;
            set_mode(*data, monster, mode == 15 ? "RN" : "WL", now - 40);
            monster.speed_pct = speed_pct; monster.boss_speed = boss_speed; monster.move_pct = move_pct;
            monster.laid_until = laid ? now + 100000 : 0;
            monster.chill_until = chilled ? now + 100000 : 0;
            monster.wandering = true; monster.aware = false; monster.chase = -1;
            monster.steps = { { int(std::floor(monster.unit.x * 5)) + 250, int(std::floor(monster.unit.y * 5)) } };
            monster.step = 0; monster.budget = 0x14; monster.end_x = monster.steps[0].first; monster.end_y = monster.steps[0].second;
            monster.next_act = now + 100000; monster.mode_until = 0;
            const std::string mode_before(monster.mode);
            std::array<Foe, 1> foes{};
            foes[0].x = 400.f; foes[0].y = 400.f; foes[0].alive = false;
            const Crowd crowd;
            std::vector<Missile> missiles;
            const float x0 = monster.unit.x, y0 = monster.unit.y;
            monster_update(*data, open, monster, foes, rng, now, 0.04f, crowd, missiles, pack);
            const float moved = std::hypot(monster.unit.x - x0, monster.unit.y - y0);
            const float unit_pct = cells_per_sec(float(velocity)) * 0.04f;
            type_info = saved_type;
            if (mode_before != (mode == 15 ? "RN" : "WL")) { std::printf("%d skip mode %s\n", id, mode_before.c_str()); continue; }
            if (velocity <= 0 || unit_pct <= 0) { std::printf("%d skip vel\n", id); continue; }
            std::printf("%d %.3f\n", id, double(moved / unit_pct * 100.f));
        } else {
            int cls = 0, run = 0, frw = 0, vp = 0, chilled = 0;
            if (!(fields >> cls >> run >> frw >> vp >> chilled)) continue;
            *world.actor = start;
            auto& player = *world.actor;
            player.character.character_class = cls;
            player.running = run != 0;
            player.character.stats.values[d2d::d2s::kStamina] = 100 << 8;
            const std::uint32_t now = 1000 * 40;
            player.chilled = chilled ? now + 100000 : 0;
#ifdef VERIFY_VP
            player.player_combat.frw = frw;
            player.player_combat.velocity_pct = vp;
#else
            player.player_combat.frw = frw + vp;
#endif
            std::printf("%d %d.000\n", id, world.walk_percent(player, now));
        }
    }
    return 0;
}
