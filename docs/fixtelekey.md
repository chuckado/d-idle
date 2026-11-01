# FIXTELEKEY: Teleport capability from the character's skills

> status: card fixtelekey, offline, not verified live. Source: BUILDFLEX audit row 5 / seam S3 (docs/research/build-flexibility-audit.md on branch agent/buildflex). Tags: V = read in code or run as a test here, S = secondary doc, U = unverified.

## 1. Call graph before the change (V: grep and read of the tree at origin/main)

Readers of `Navigator.TeleportKey` (non-test): `nav.hasTeleport()` (internal/nav/teleport.go) gates `Explore` (explore.go: `teleLeg`), `ExitTo` (exit.go: `teleportToward`), the probe walk (probe.go), `TeleportTo`/islands (teleport_islands.go) and the cast itself (`hopAt`: `in.CastAt(TeleportKey, cell)`). cmd/d2bot reads it in four places: `moverAdapter.CanTeleport` (feeds `task.World.CanTeleport`), `moverAdapter.TeleportToArea`, `newBossRunner` (`boss.Opts.CanTeleport`, read once at construction) and `farmLoop` (`FarmChar.Teleport`).

Setters: only `tools/rigs/act2live/main.go` (lines 70, 934-936, key "f2", cleared in town and the lair). No assignment in cmd/d2bot, internal/d2denv, internal/d2dcombat, internal/d2dbackend or tools/* (other than the rig). `d2.SkillTeleport` is used by no production code (V grep).

Other routes that bind or cast Teleport in production: the Fighter binds it. `profiles.json` layouts of the Sorceress profiles map Teleport to f2; `EnsureHotkeys` (Game.exe, after `prepare` and every leveler spend) and `d2dcombat.Rig.BindSkills` (d2d, pseudo key `sk54`) store the verified key in `Fighter.keys[54]`. Nothing ever read that entry for movement, and the Fighter itself never casts Teleport (`castgap` lists it only for timing). `cmd skill 54` appears only in the d2d input as the generic CastAt of any `sk<id>` key. Plan tables' `hotkeys` and `main_attack` stay unused in production (S1 of the audit, not this card).

What happened in production at clvl 18+ in Act 2 (Teleport learned and bound to f2): the key was bound, `Navigator.TeleportKey` stayed "", so `hasTeleport()` was false: every Explore and ExitTo walked, `moverAdapter.CanTeleport` returned false (task layer, Summoner and Council moat shortcuts off), boss runs skipped every `NeedsTeleport` stand point, and the farm table saw `Teleport=false`. Confirmed: a real production bug (finding is correct).

## 2. The change

- The key comes from the Fighter: `Fighter.TeleportKey()` = verified bind of the skill whose name is "Teleport" in the game tables (id from data, no literal 54). Skills the character does not know are never bound (`KnownSkills` needs effective level 1, or the skill list in d2d), so a Paladin, an Amazon, or a Sorceress that has not learned Teleport gets "" and nothing hops. A build whose layout binds Teleport to another key works unchanged. Enigma-granted Teleport on any class works once a layout or `BindSkills` binds it.
- `Navigator.TeleportKeyFn` is asked at every use, so learning Teleport at clvl 18 turns hops on at the next rebind without restarting. A fixed `TeleportKey` (rigs, tests) still wins.
- `wireTeleport` (cmd/d2bot/telewire.go) is called in both stacks (`buildBehaviors`, `buildD2DStack`). It also installs `Navigator.TeleportOK`: a level whose Levels.txt Teleport column is not 1, and every town, never gets a hop (`data.CanTeleportIn`, V read in docs/research/sk-teleport-trace.md). Duriel's Lair (73, rule 2, clear line only) is left to walking because the hop cell is chosen without a line check; the rig did the same. An area the tables do not know stays allowed.
- The cast path is unchanged: `CastAt(key, cell)` = select the hotkey, right click the cell (Game.exe), or `cmd skill <id>` (d2d). The key was verified by `checkKey` (docs/research/skill-bind-trace.md: the server stores a select only for a row with an entry of that id at level > 0), so the select cannot silently pick another skill.
- Boss: `Opts.CanTeleportFn` replaces the build-time bool in `newBossRunner`; the farm char and task adapter read `HasTeleport()` at call time.
- Refused hops stay bounded (already so, tests added): a cell where nothing happened is never tried again (`teleBad`), 3 consecutive non-moving hops end `teleportToward` and the explore leg falls back to walking (`teleBad`, `teleFails`, `teleportBadHops = 3`), mana waits are capped at 40 s. On d2d every refused hop still costs about 23.8 mana (DRILLS15 T3, V-d2d), so the bound is what limits the loss: at most 3 wasted casts per leg.

Not changed: which areas are visited, when the bot hops, hop range and mana constants, behaviour of a character without Teleport (key "" everywhere: byte-for-byte the old walk).

## 3. Tests

`go test -p 1 -count=1 ./internal/nav ./internal/combat ./internal/boss ./cmd/d2bot`: key from the layout for a Sorceress with Teleport, none without it or at skill level 0, none for a Paladin, another bound key honoured, key read lazily (learned later), hops use the function key and the right hotkey, no key walks without error, a refused level never casts, refused hops are bounded, boss `CanTeleportFn` wins and is live, `wireTeleport` follows the Fighter bind and the unknown-area rule.

## 4. Live proof recipe (not run; needs a hidden instance and the lab binary)

1. Throwaway Sorceress, offline d2d instance (cheat ledger entry via `cmd/d2cheat`). Raise to clvl 18 and give skill points with `debug points` so Teleport (and Static Field or Charged Bolt) are at level 1; set the leveler build to one whose layout lists Teleport.
2. Start `d2bot` (d2d stack) with the character in Lut Gholein; confirm in the log one `combat bind` or `BindSkills` line and that `HasTeleport` is true (a `nav teleport` Debug line appears on the first hop).
3. Run an Act 2 nav leg (for example Rocky Waste to Far Oasis through Dry Hills, `ExitTo`). Record: hops (count of `nav teleport` lines), `nav teleport did not move` lines (refused), mana before and after per hop (expect about 23.8 per hop at level 1, regen aside), walked distance versus the same leg without Teleport (set the skill to 0 or use a Paladin as the control: expect 0 hops, walk as before).
4. Boss path: run the Summoner (`NeedsTeleport` stop) and check the stand point is used; Duriel's Lair: expect no hops there.
5. d2d deviations that change the numbers (docs/probes/drills/README.md DRILLS15): T2 no room limit on d2d (hops of 120 subtiles all land; game.exe refuses outside its room list), T3 blocked target refused AND paid, T4 monster target lands on the monster (game.exe beside it), T7 a 25-cell hop is not refused at exactly 24 mana (mana 0.23 left). Treat refused-hop counts on d2d as a lower bound of game.exe's.
6. Expected: clvl 18+ Act 2 legs hop, refused hops under 3 per leg, no hops in town or the lair, a no-Teleport control walks.
