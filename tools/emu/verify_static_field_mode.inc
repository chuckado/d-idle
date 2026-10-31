// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_game (d2lod-bot card verifyprs): Static Field through the REAL
// Fight::spot (components/game/fight/strike.cpp), not a copy of its arithmetic. Pasted into
// tests/game/test_game.cpp right after `character.panel = panel_stats(...)` by
// tools/emu/verify_static_field.py's `patch` step, in a throwaway worktree only.
//   test_game --verify-static CASES   -> one line per case: `id life_after` (256ths, 0 = dead)
// A case: one monster (fallen1) on the player's spot, its max life, life (256ths), raw lightning
// resistance (boss_res[3] makes Monster::target's res[3] equal it), the player's lightning pierce
// (psum stat 334), the game's difficulty and expansion flag, the skill level; phase 1 also sets
// Skills.txt Param4 (calc1, the %) and Param3 (calc2, the least) on the loaded row, and with
// VERIFY_SFMIN (the after build) DifficultyLevels StaticFieldMin for the difficulty.
if (argc > 2 && std::string_view(argv[1]) == "--verify-static") {
    const auto sf_found = data->skills.by_name.find("Static Field");
    assert(sf_found != data->skills.by_name.end());
    auto* sf = const_cast<d2d::rules::Skill*>(data->skills.get(sf_found->second));
    const auto real_par = sf->par;
    const auto fallen_row = data->monsters.by_id.find("fallen1");
    assert(fallen_row != data->monsters.by_id.end());
    World world(&*data, -1, -1);
    world.enter(character);
    const World::Player start = *world.actor;
    d2d::rules::Rng made_rng{ 5 };
    const Monster base_monster = make_monster(*data, int(fallen_row->second), world.actor->player.x, world.actor->player.y, made_rng, 0);
    std::ifstream cases(argv[2]);
    std::string line;
    while (std::getline(cases, line)) {
        std::istringstream fields(line);
        int id = 0, phase = 0, difficulty = 0, expansion = 0, max_hp = 0, life = 0, res = 0, pierce = 0, slvl = 1, pct = 0, low = 0, floor_pct = 0;
        if (!(fields >> id >> phase >> difficulty >> expansion >> max_hp >> life >> res >> pierce >> slvl >> pct >> low >> floor_pct)) continue;
        *world.actor = start;
        auto& player = *world.actor;
        sf->par = real_par;
        if (phase == 1) { sf->par[3] = pct; sf->par[2] = low; }
#ifdef VERIFY_SFMIN
        const auto saved_min = data->static_field_min;
        if (phase == 1) const_cast<GameData&>(*data).static_field_min[std::size_t(difficulty)] = floor_pct;
#endif
        if (expansion) player.character.header.status |= 0x20; else player.character.header.status &= ~0x20;
        world.fight.game_difficulty = difficulty;
        player.psum[d2d::d2s::kPassiveLightningPierce] = pierce;
        Monster monster = base_monster;
        monster.difficulty = difficulty;
        monster.stats.hit_points = max_hp;
        monster.hit_points = (life + 255) >> 8;
        monster.life_frac = monster.hit_points * 256 - life;
        monster.boss_res = {};
        monster.boss_res[3] = res - monster.target(*data).res[3];
        world.fight.monsters.clear();
        world.fight.add_monster(monster);
        const int got_res = world.fight.monsters.back().target(*data).res[3];
        Fight::Caster caster{ player.player.x, player.player.y, slvl, -1, player.player.x, player.player.y };
        world.fight.spot(*sf, 1000 * kTickMs, caster);
        const auto& struck = world.fight.monsters.back();
        std::printf("%d %d%s\n", id, struck.alive() ? struck.hit_points * 256 - struck.life_frac : 0, got_res == res ? "" : " RESFAIL");
#ifdef VERIFY_SFMIN
        const_cast<GameData&>(*data).static_field_min = saved_min;
#endif
    }
    sf->par = real_par;
    return 0;
}
