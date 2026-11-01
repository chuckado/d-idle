// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_game (d2lod-bot card verifyprs2): a merc's and a pet's walk % through the
// REAL d2d code of the build under test. Pasted into tests/game/test_game.cpp right after
// `character.panel = panel_stats(...)` in a throwaway worktree only.
//   test_game --verify-mercpet CASES  -> `id pct` per case (pct to 4 decimals) or `id skip why`
// kind 2, a merc: World::spawn_merc for the hireling Id (header.merc_type), then one real
//   Fight::merc_turn with the merc in WL or RN on a path 40 cells long.
// kind 3, a pet: a Fight::Pet of the MonStats row with its summoning skill's id (no Shadow: its base is
//   not claimed), one real Fight::pet_walking on a path 40 cells long.
// The level is the Blood Moor with every subtile open. The case's Velocity replaces the MonStats row's
// (restored after). The step is chosen so the unit moves about 4 cells whatever the build's rule (so a
// huge walk % never reaches the path's end); the % is the distance moved over Velocity's distance in it.
if (argc > 2 && std::string_view(argv[1]) == "--verify-mercpet") {
    World world(&*data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    Level open = *data->level(d2d::rules::level_ids::kBloodMoor);
    std::ranges::fill(open.walk, std::uint16_t(0));
    world.level = &open;
    std::ifstream cases(argv[2]);
    std::string line;
    const std::uint32_t now = 2000 * 40;
    const Crowd crowd;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        int id = 0, kind = 0;
        if (!(fields >> id >> kind)) continue;
        std::string row_id;
        int merc_type = 0, mode = 0, velocity = 0, move_pct = 0, speed_pct = 0, buff = 0, skill_id = -1;
        bool parsed = false;
        if (kind == 2) parsed = bool(fields >> merc_type >> mode >> velocity >> move_pct);
        else if (kind == 3) parsed = bool(fields >> row_id >> skill_id >> mode >> velocity >> move_pct >> speed_pct >> buff);
        if (!parsed) { std::printf("%d skip parse\n", id); continue; }
        if (kind == 2) {
            const auto found = data->mercs.find(merc_type);
            if (found == data->mercs.end()) { std::printf("%d skip nomerc\n", id); continue; }
            row_id = found->second.npc.id;
        }
        const auto row = data->monsters.by_id.find(row_id);
        if (row == data->monsters.by_id.end()) { std::printf("%d skip norow\n", id); continue; }
        if (velocity <= 0) { std::printf("%d skip vel\n", id); continue; }
        auto& type_info = const_cast<GameData&>(*data).monsters.types[std::size_t(row->second)];
        const auto saved_type = type_info;
        type_info.velocity = velocity;
        const std::string_view want = mode == 15 ? "RN" : "WL";
        // a step that moves about 4 cells under the old rule (Run, floor 10) and the new (Velocity, floor 25)
        const int sum = std::max(100 + move_pct + speed_pct + buff, 10);
        const float fastest = cells_per_sec(float(std::max({ velocity, type_info.run, 1 }))) * float(sum) / 100.f;
        const float tick = std::clamp(4.f / fastest, 1e-5f, 8.f);   // longer than a frame for slow units: float precision
        float moved = 0;
        bool walked = false;
        if (kind == 2) {
            *world.actor = start;
            auto& header = world.actor->character.header;
            header.merc_type = std::uint16_t(merc_type); header.merc_seed = 11; header.merc_dead = false;
            world.spawn_merc();
            if (!world.actor->merc || world.actor->merc->type != int(row->second)) { type_info = saved_type; std::printf("%d skip spawn\n", id); continue; }
            auto& merc = *world.actor->merc;
            merc.unit.x = merc.unit.y = 20.5f;
            merc.mode = want; merc.unit.walking = true; merc.unit.mode_ms = now - 40; merc.mode_until = 0; merc.next_act = now + 100000;
            merc.unit.path = { { merc.unit.x + 40.f, merc.unit.y } };
            merc.move_pct = move_pct;
            world.actor->merc_chase = false; world.actor->merc_target = -1;
            world.fight.merc_turn(now, tick, crowd, false);
            walked = merc.mode == want;
            moved = std::hypot(merc.unit.x - 20.5f, merc.unit.y - 20.5f);
        } else {
            d2d::rules::Rng pet_rng{ std::uint32_t(id * 7 + 3) };
            Fight::Pet pet;
            pet.monster = make_monster(*data, int(row->second), 20.f, 20.f, pet_rng, 0);
            pet.skill = skill_id;
            pet.where = &open;
            set_mode(*data, pet.monster, want, now - 40);
            pet.monster.unit.x = pet.monster.unit.y = 20.5f;
            pet.monster.unit.walking = true;
            pet.monster.unit.path = { { 60.5f, 20.5f } };
            pet.monster.move_pct = move_pct; pet.speed_pct = speed_pct;
            if (buff != 0) pet.buff_stats[d2d::d2s::kVelocityPercent] = buff;
            world.fight.pet_walking(pet, now, tick, crowd);
            walked = pet.monster.mode == want;
            moved = std::hypot(pet.monster.unit.x - 20.5f, pet.monster.unit.y - 20.5f);
        }
        type_info = saved_type;
        if (!walked) { std::printf("%d skip mode\n", id); continue; }
        const double unit_dist = double(cells_per_sec(float(velocity))) * double(tick);
        std::printf("%d %.4f\n", id, double(moved) / unit_dist * 100.0);
    }
    return 0;
}
