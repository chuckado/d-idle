# FIXMISSILE: d2d missiles and the cast gap

> status: current. Card openspec/changes/fix-missile. Source: docs/research/monster-attack-trace.md Q5 (B1, B3). Offline only; nothing run live.

## B1 d2d gives the bot no missiles (d2d-only fix, bot side done, fork verb requested)
What the d2d dev interface offers (V read: docs/d2d-backend.md, docs/d2d-fork.md, `internal/d2dctl`, `internal/d2dsrc`): `state missiles=<count>` only. The `dump` sections (`player,stats,skills,items,quests,ui,merc,time`) do not include missiles ("not in this first cut"), and no `debug` verb lists them. So there is NO listing.

Request for the fork agent (not built here): `debug missiles`, one row per missile in flight, tab separated, then `ok` (empty list: just `ok`):
`<unit> <missile id> <x> <y> <dx> <dy> <speed> <owner type> <owner unit> <ttl>`
- unit: unit id; missile id: missiles.txt row; x y: level-relative cells like `monsters`; dx dy: unit vector of flight; speed: cells per second; owner type: d2 unit type of the shooter (0 player, 1 monster), owner unit: its unit id (the owner is required: without it our own bolts would read as hostile); ttl: ticks left, -1 unknown.

Bot side: `d2dctl.Client.Missiles` / `ParseMissile`, `d2dsrc.Source.FillMissiles` (missiles.go, called from `d2denv.Reader.Snapshot`; items.go/source.go are owned by the open card fix-obj) fills `Snapshot.Missiles` (Class, OwnerType, OwnerID, subtile X Y; dx dy speed ttl are parsed but unused: the dodge only needs position and owner). One extra round trip per snapshot (not prefetched); the first `err` latches it off (stock d2d: no cost). Game.exe's reader is untouched. Shared vs d2d-only: d2d-only.

Note: the sidestep (`MissileDodge`) stays off by default (existing knob, policy); the radius dodge in the fight loop is on. Turning the sidestep on is held for the live pass.

## B3 CastGapMs below the cast (shared default change)
Evidence: the early Sorceress casts Fire Bolt / Charged Bolt in 560 ms; docs/d2d-casts-real.md (V-d2d) measured 1.6 to 1.75 commands per real cast at `CastGapMs` 450 (dropped mid-cast) and 1.02 to 1.04 at 650; the trace counts 5.4 cast commands per kill vs 1.1 on Game.exe. New default 650, knob range unchanged (250..1200), a constant not computed from FCR (the knob is the single place; faster casters can lower it). It applies to every class (shared); other classes add their own waits on top. Three simulator tests whose fixtures were tuned to 450 pin it (`oldCastGap`). docs/d2d-combat.md still says 450 (not owned here).

## Held (policy, after a live pass)
B2 run policy into packs (`internal/d2dinput/run.go`); B4 `TownPotionMin` 2 and chicken exits with an empty belt.

## Live proof (after the fork verb and merge)
Drill D5 (monster fight): `debug spawn` a ranged monster (Corrupt Archer, Fallen Shaman) next to a throwaway character, run the fighter, count `dodging` lines and cast commands per kill. Pass: dodges > 0 with the verb, commands per cast about 1.0.
