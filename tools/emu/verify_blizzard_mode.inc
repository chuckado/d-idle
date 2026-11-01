// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_game (d2lod-bot card verifyblizzard): Blizzard / Eruption cast through
// the REAL Fight::fire and Fight::missile_tick (components/game/fight/missiles.cpp), not a copy of the
// shard arithmetic. Pasted into tests/game/test_game.cpp right after `character.panel = panel_stats(...)`
// by tools/emu/verify_blizzard.py's `patch` step, in a throwaway clone only.
//   test_game --vb-blizzard CASES  -> one line per shard: `id skill frame dx dy` (subtiles from the centre)
//                                     and one `# id centre_frames shards rng_low rng_high` line per case
// A case: `id skill_name rng_low rng_high level target_dx target_dy` (target in cells from the player).
// The world rng is set to {rng_low, rng_high} first; d2d's shards roll the shared world rng.
// Per frame: Fight::missile_tick, then the pending missiles join and the spent ones (now >= dies) go,
// as Fight::world does (tick.cpp) without flight, hits or the town / Blood Moor split.
if (argc > 2 && std::string_view(argv[1]) == "--vb-blizzard") {
    World world(&*data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        int id = 0, lvl = 1;
        std::string name;
        std::uint32_t low = 0, high = 0;
        float tdx = 0, tdy = 0;
        if (!(fields >> id >> name >> low >> high >> lvl >> tdx >> tdy)) continue;
        const auto found = data->skills.by_name.find(name);
        assert(found != data->skills.by_name.end());
        const auto* skill = data->skills.get(found->second);
        *world.actor = start;
        auto& player = *world.actor;
        world.fight.missiles.clear();
        world.fight.pending.clear();
        world.fight.monsters.clear();
        world.fight.missile_frame = -1;
        world.rng.low = low;
        world.rng.high = high;
        const std::uint32_t t0 = 1000 * kTickMs;
        Fight::Caster caster{ player.player.x, player.player.y, lvl, -1, player.player.x + tdx, player.player.y + tdy };
        world.fight.fire(*skill, t0, caster);
        const GameData::MissileInfo* centre_info = nullptr;
        for (const auto& m : world.fight.pending)
            if (m.info && (m.info->srv_do == 10 || m.info->srv_do == 25)) centre_info = m.info;
        if (!centre_info) { std::printf("# %d nocentre\n", id); continue; }
        const auto sub = data->missiles.find(centre_info->sub);
        assert(sub != data->missiles.end());
        const GameData::MissileInfo* shard_info = &sub->second;
        int centre_frames = 0, shards = 0;
        for (int f = 0; f < 1000; ++f) {
            const std::uint32_t now = t0 + std::uint32_t(f) * kTickMs;
            const Missile* centre = nullptr;
            for (const auto& m : world.fight.missiles) if (m.info == centre_info) centre = &m;
            for (const auto& m : world.fight.pending) if (m.info == centre_info) centre = &m;
            if (!centre) break;
            const float cx = centre->x, cy = centre->y;
            ++centre_frames;
            world.fight.missile_tick(now);
            for (const auto* list : { &world.fight.missiles, &world.fight.pending })
                for (const auto& m : *list)
                    if (m.info == shard_info && m.born == now) {
                        ++shards;
                        std::printf("%d %s %d %ld %ld\n", id, name.c_str(), f, std::lround((m.x - cx) * 5), std::lround((m.y - cy) * 5));
                    }
            // what Fight::world does after missile_tick, minus flight and hits: the new join, the spent go
            std::ranges::move(world.fight.pending, std::back_inserter(world.fight.missiles));
            world.fight.pending.clear();
            std::erase_if(world.fight.missiles, [&](const Missile& m) { return now >= m.dies; });
        }
        std::printf("# %d %d %d %u %u\n", id, centre_frames, shards, world.rng.low, world.rng.high);
    }
    return 0;
}
