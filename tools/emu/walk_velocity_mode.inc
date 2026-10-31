// SPDX-License-Identifier: GPL-3.0-or-later
// Copy of the test_game mode in branch fix/walk-velocity-v2 (tests/game/test_game.cpp, after `character.panel = ...`); see tools/emu/walk_velocity.py.
    // `--walk-velocity CASES`: a unit's walk % through the REAL d2d code, the
    // cases of tools/emu/walk_velocity.py (game.exe's own FUN_00623f50). Per
    // case `num pct` (the walk %, 3 decimals) or `num skip why`. A monster:
    // one frame of monster_update on a WL / RN monster with a far spot to
    // walk to; a player: World::walk_percent; a merc: one Fight::merc_turn;
    // a pet: one Fight::pet_walking. The % is the distance moved over the
    // Velocity's step (cells_per_sec x 0.04 s). The level is a copy of the
    // Blood Moor with every subtile open; the case's Velocity (and a
    // monster's coldeffect) replace the MonStats row's, restored after.
    if (argc > 2 && std::string_view(argv[1]) == "--walk-velocity") {
        World world(&*data, -1, -1);
        world.enter(character);
        const World::Player start = *world.actor;
        Level open = *data->level(d2d::rules::level_ids::kBloodMoor);
        std::ranges::fill(open.walk, std::uint16_t(0));
        world.level = &open;
        std::ifstream cases(argv[2]);
        std::string line;
        const std::uint32_t now = 1000 * 40;
        const float tick = 0.04f;
        const Crowd crowd;
        while (std::getline(cases, line)) {
            std::istringstream fields(line);
            int num = 0, kind = 0;
            if (!(fields >> num >> kind)) continue;
            if (kind == 1) {
                int klass = 0, run = 0, frw = 0, vel_pct = 0, chilled = 0;
                if (!(fields >> klass >> run >> frw >> vel_pct >> chilled)) continue;
                *world.actor = start;
                auto& player = *world.actor;
                player.character.character_class = klass;
                player.running = run != 0;
                player.character.stats.values[d2d::d2s::kStamina] = 100 << 8;
                player.chilled = chilled ? now + 100000 : 0;
                player.player_combat.frw = frw;
                player.player_combat.velocity_pct = vel_pct;
                std::printf("%d %d.000\n", num, world.walk_percent(player, now));
                continue;
            }
            // 0 a monster `row mode difficulty velocity cold speed boss move laid chilled`; 2 a merc `merc_type
            // mode velocity move`; 3 a pet `row mode velocity move speed buff` (mode 2 WL, 15 RN)
            std::string row_id;
            int mode = 0, difficulty = 0, velocity = 0, cold = 0, speed_pct = 0, boss_speed = 0, move_pct = 0, laid = 0, chilled = 0, buff = 0, merc_type = 0;
            bool parsed = false;
            if (kind == 0) parsed = bool(fields >> row_id >> mode >> difficulty >> velocity >> cold >> speed_pct >> boss_speed >> move_pct >> laid >> chilled);
            else if (kind == 2) parsed = bool(fields >> merc_type >> mode >> velocity >> move_pct);
            else if (kind == 3) parsed = bool(fields >> row_id >> mode >> velocity >> move_pct >> speed_pct >> buff);
            if (!parsed) continue;
            if (kind == 2) {
                const auto found = data->mercs.find(merc_type);
                if (found == data->mercs.end()) { std::printf("%d skip nomerc\n", num); continue; }
                row_id = found->second.npc.id;
            }
            const auto row = data->monsters.by_id.find(row_id);
            if (row == data->monsters.by_id.end()) { std::printf("%d skip norow\n", num); continue; }
            auto& type_info = const_cast<GameData&>(*data).monsters.types[std::size_t(row->second)];
            const auto saved_type = type_info;
            type_info.velocity = velocity;
            if (kind == 0) type_info.diff[std::size_t(difficulty)].cold_effect = cold;
            const std::string_view want = mode == 15 ? "RN" : "WL";
            const float unit_pct = cells_per_sec(float(velocity)) * tick;
            float moved = 0;
            bool walked = false;
            d2d::rules::Rng case_rng{ std::uint32_t(num + 1) };
            if (kind == 0) {
                std::vector<Monster> pack{ make_monster(*data, int(row->second), 10.f, 10.f, case_rng, difficulty) };
                auto& monster = pack[0];
                monster.difficulty = difficulty;
                set_mode(*data, monster, want, now - 40);
                monster.speed_pct = speed_pct; monster.boss_speed = boss_speed; monster.move_pct = move_pct;
                monster.laid_until = laid ? now + 100000 : 0;
                monster.chill_until = chilled ? now + 100000 : 0;
                monster.wandering = true; monster.aware = false; monster.chase = -1;
                monster.steps = { { int(std::floor(monster.unit.x * 5)) + 250, int(std::floor(monster.unit.y * 5)) } };
                monster.step = 0; monster.budget = 0x14; monster.end_x = monster.steps[0].first; monster.end_y = monster.steps[0].second;
                monster.next_act = now + 100000; monster.mode_until = 0;
                std::array<Foe, 1> foes{};
                foes[0].x = 400.f; foes[0].y = 400.f; foes[0].alive = false;
                std::vector<Missile> missiles;
                const float from_x = monster.unit.x, from_y = monster.unit.y;
                monster_update(*data, open, monster, foes, case_rng, now, tick, crowd, missiles, pack);
                walked = monster.mode == want;
                moved = std::hypot(monster.unit.x - from_x, monster.unit.y - from_y);
            } else if (kind == 2) {
                *world.actor = start;
                auto& header = world.actor->character.header;
                header.merc_type = std::uint16_t(merc_type); header.merc_seed = 7; header.merc_dead = false;
                world.spawn_merc();
                if (world.actor->merc && world.actor->merc->type == int(row->second)) {
                    auto& merc = *world.actor->merc;
                    merc.unit.x = merc.unit.y = 10.5f;
                    merc.mode = want; merc.unit.walking = true; merc.unit.mode_ms = now - 40; merc.mode_until = 0; merc.next_act = now + 100000;
                    merc.unit.path = { { merc.unit.x + 40, merc.unit.y } };
                    merc.move_pct = move_pct;
                    world.actor->merc_chase = false; world.actor->merc_target = -1;
                    world.fight.merc_turn(now, tick, crowd, false);
                    walked = merc.mode == want;
                    moved = std::hypot(merc.unit.x - 10.5f, merc.unit.y - 10.5f);
                }
            } else {
                Fight::Pet pet;
                pet.monster = make_monster(*data, int(row->second), 10.f, 10.f, case_rng, 0);
                pet.where = &open;
                set_mode(*data, pet.monster, want, now - 40);
                pet.monster.unit.x = pet.monster.unit.y = 10.5f;
                pet.monster.unit.walking = true;
                pet.monster.unit.path = { { 50.5f, 10.5f } };
                pet.monster.move_pct = move_pct; pet.speed_pct = speed_pct;
                pet.buff_stats[d2d::d2s::kVelocityPercent] = buff;
                world.fight.pet_walking(pet, now, tick, crowd);
                walked = pet.monster.mode == want;
                moved = std::hypot(pet.monster.unit.x - 10.5f, pet.monster.unit.y - 10.5f);
            }
            type_info = saved_type;
            if (!walked) { std::printf("%d skip mode\n", num); continue; }
            if (velocity <= 0 || unit_pct <= 0) { std::printf("%d skip vel\n", num); continue; }
            std::printf("%d %.3f\n", num, double(moved / unit_pct * 100.f));
        }
        return 0;
    }
