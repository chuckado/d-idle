// SPDX-License-Identifier: GPL-3.0-or-later
// Verifier mode for d2d's test_game (d2lod-bot card verifychill): a cold blow on one monster
// through the REAL d2d functions, not a copy of their arithmetic. Pasted into
// tests/game/test_game.cpp right after `character.panel = panel_stats(...)` by
// tools/emu/verify_chill.py's `patch` step, in a throwaway worktree only.
//   test_game --verify-chill CASES   -> one line per case: `id chill freeze` (ticks from now
//                                       to the end, or `-`)
// Phase 0: d2d::rules::missile_blow (cold, roll..roll in 256ths, the case's length) on
// Fight::target_of, then Fight::land. Phase 1: Fight::strike on a cold missile with a fixed
// element (elen 0) and the case's freeze length (the hit_func 13 path, glacial spike), so the
// freeze handling is d2d's own. Both: a fallen1 whose loaded MonStats coldeffect for the
// difficulty is the case's (restored after), its cold resistance through boss_res[4] (RESFAIL
// when target_of does not give the case's), the player's cold pierce (psum 335), an optional
// running chill, `boss` its Boss kind and `bcol` its MonStats boss column.
if (argc > 2 && std::string_view(argv[1]) == "--verify-chill") {
    const auto vc_row = data->monsters.by_id.find("fallen1");
    assert(vc_row != data->monsters.by_id.end());
    auto& vc_type = const_cast<d2d::rules::MonType&>(data->monsters.types[vc_row->second]);
    const auto vc_diff = vc_type.diff;
    const bool vc_bcol = vc_type.boss_column;
    const auto spike = data->skills.by_name.find("Glacial Spike");
    assert(spike != data->skills.by_name.end());
    GameData::MissileInfo vc_info;
    vc_info.name = "verifychill";
    vc_info.etype = 2;
    vc_info.to_hit = false;
    World vc_world(&*data, -1, -1);
    vc_world.enter(character);
    const World::Player vc_start = *vc_world.actor;
    d2d::rules::Rng vc_made_rng{ 5 };
    const Monster vc_base = make_monster(*data, int(vc_row->second), vc_world.actor->player.x, vc_world.actor->player.y, vc_made_rng, 0);
    d2d::rules::Rng vc_rng{ 11 };
    const std::uint32_t vc_now = 2000 * kTickMs;
    std::ifstream vc_cases(argv[2]);
    std::string vc_line;
    while (std::getline(vc_cases, vc_line)) {
        std::istringstream fields(vc_line);
        int id = 0, phase = 0, difficulty = 0, cold_effect = 0, res = 0, pierce = 0, roll = 0, length = 0, prior = 0, uflag = 0, bcol = 0, hire = 0, freeze = 0, boss = 0;
        if (!(fields >> id >> phase >> difficulty >> cold_effect >> res >> pierce >> roll >> length >> prior >> uflag >> bcol >> hire >> freeze >> boss)) continue;
        *vc_world.actor = vc_start;
        vc_world.actor->psum[d2d::d2s::kPassiveColdPierce] = pierce;
        vc_world.fight.game_difficulty = difficulty;
        vc_type.diff[std::size_t(difficulty)].cold_effect = cold_effect;
        vc_type.boss_column = bcol != 0;
        Monster monster = vc_base;
        monster.difficulty = difficulty;
        monster.boss = d2d::rules::Boss(boss);
        monster.stats.hit_points = monster.hit_points = 1000000;
        monster.life_frac = 0;
        monster.boss_res = {};
        monster.boss_res[4] = res - monster.target(*data).res[4];
        monster.chill_until = prior > 0 ? vc_now + std::uint32_t(prior) * kTickMs : 0;
        monster.stun_until = 0;
        vc_world.fight.monsters.clear();
        vc_world.fight.add_monster(monster);
        const auto target = vc_world.fight.target_of(0);
        if (phase == 0) {
            const auto blow = d2d::rules::missile_blow({ .etype = 2, .elo = roll, .ehi = roll, .elen = length }, target, vc_world.fight.pierce(), vc_rng);
            vc_world.fight.land(0, blow, true, vc_now, -1);
        } else {
            Missile missile;
            missile.info = &vc_info;
            missile.skill = int(spike->second);
            missile.fixed = roll;
            missile.friendly = true;
            vc_world.fight.strike(missile, 0, vc_now, freeze);
        }
        const auto& struck = vc_world.fight.monsters.back();
        const auto ticks = [&](std::uint32_t until) { return until > vc_now ? std::to_string((until - vc_now) / kTickMs) : std::string("-"); };
        std::printf("%d %s %s%s\n", id, ticks(struck.chill_until).c_str(), ticks(struck.stun_until).c_str(), target.res[4] == res ? "" : " RESFAIL");
        vc_type.diff = vc_diff;
        vc_type.boss_column = vc_bcol;
    }
    return 0;
}
