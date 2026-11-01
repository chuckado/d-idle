# FIXBUFF: armor recast inside a long fight, armor states, d2d buff states

Card `openspec/changes/fixbuff`. Findings B1-B3 of docs/research/sk-armor-trace.md (SKARMOR). Builds on docs/fixrecast.md (`buffDue` / `buffLimit`).

Tags: V verified (how), S secondary, U unverified.

## What changed

- **B1.** `Fighter.Buff` ran only on need-buff edges (new game, back from town, goal reset, death; internal/bot/goal.go), so inside one clear longer than 120 s the armor lapsed (V read of the call sites). The fight loop (internal/combat/fight.go step 1c) now calls `fightBuff`: the same `buffDue` decision (state if readable, else the clock and the real duration minus `buffMarginSec`) on the snapshot the loop already has, no extra game read per tick, a recast costs one cast. Guards: it sits after the survival action, the hard rules and the lost-fight check (so a potion, portal or chicken always comes first), and is skipped when `crowdAdjacentMax` or more monsters are within `combat.crowdAdjacent` (the rule `meleeCastGuard` uses for the portal cast; a hit cutting the cast is U). Their decisions are unchanged. Between fights (roaming) nothing recasts: the next fight tick does, and the sweep's own edge still covers town trips. `Buff` is now read + aura + `buffPass`; its behaviour on an edge is unchanged.
- **B2.** `d2.StateChillingArmor` 20, `d2.StateEnergyShield` 30, `d2.StateShiverArmor` 88 and their `buffStates` entries. Ids are States.txt rows as game.exe loads them (V read: sk-armor-trace Q1). Not seen switching live (U); Frozen 10 is S/V as in fixrecast, Bone Armor 14 V live. Energy Shield ends at 0 mana in game.exe, so the state reading it as gone recasts it after the 4 s grace (V read, trace Q5).
- **B3.** The dump has no state list. The control socket's bare, read-only `debug states` (docs/forkdbg.md: `buffs=<skill>:<level>,`) lists the player's self buffs by skill id: `d2dctl.Client.Buffs` parses it, `d2dsrc.Options.BuffStates` (on in `d2denv.Build`, one more verb in the prefetch batch) maps the skills 40, 50, 58, 60, 68 to the state ids and sets `Player.States`. Other states (poison, curses) stay unknown. A binary that refuses the verb switches the read off after one note, then the clock rule runs as before. UNVERIFIED live: the exact reply line is taken from forkdbg.md (a lab fork doc, S); the parser takes the `buffs=` token from any line. Skill ids 50, 58, 60, 68 are stock Skills.txt rows (S). No fork verb was added. If the production d2d lacks the bare verb the read is simply off.

## Tests (offline)

`go test -p 1 -count=1 ./internal/combat ./internal/d2 ./internal/d2dctl ./internal/d2dsrc ./internal/d2denv`: `TestFightBuffRecastsInLongFight` (600 s fight, levels 1 and 2, clock-only and with fake states: never down, 5 to 6 casts), `TestFightBuffStateLostRecastsAtOnce`, `TestFightBuffNotWhenSurrounded`, `TestBuffStates`, `TestBuffStatesRefused`, `TestParseBuffs`; the fixrecast tests still pass.

## Live proof recipe (not run)

DRILLS13's armor recast proof becomes valid for the long-fight case once this merges; rerun it. On a fork slot (`debug points 40 1`, bind, `debug mana full`, cast, note the tick): keep a monster alive in a fight (`debug mon set <id> life 1000 maxlife 1000 dmg 0 0` next to her, no need-buff edge), step lockstep to about 114 s (126 s at level 2): the log must show `buff skill=Frozen Armor` from inside the fight (`fight buff` has no own line), `debug states` shows `buffs=40:N` throughout and never empty before 120 s (132 s). Also check that `Player.States` has bit 10 on the d2d backend (a snapshot dump) and that `debug states` is accepted by the production binary. Without this change the same run shows the armor gone at 120 s (SKARMOR A1).

## Shared-file lines for the merger

- BACKLOG: fixbuff live proof (long fight armor recast on a fork slot, rerun DRILLS13's armor proof); confirm the exact `debug states` reply and ids 20, 30, 88 live.
- INDEX: docs/fixbuff.md, armor recast inside long fights, d2d buff states.
- CHANGELOG: the bot recasts Frozen Armor and the other armors during a long fight (before: only after a town trip or new game), and reads the armor state on d2d.
