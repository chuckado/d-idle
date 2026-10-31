# fixmisc: DRILLS5 findings F3 and F4

> status: offline fix merged, live proof pending (card `openspec/changes/fixmisc`).

## F3 (P28): a poor melee character rests instead of drinking mana
Verdict: INTENDED, not a bug. `resource.PotionPolicy` (internal/resource/potion.go) documents it: a character with gold below `PoorGold` that has a melee or free fallback attack spends mana potions only when the pack danger is 2 or more; healing potions keep it alive, mana comes back by resting and melee. The D5.p28 control (gold 0, no mana drink) and the rich run (2 drinks) are that rule at work. No code change. `TestPoorMeleeRestsRichDrinksMana` (internal/resource/policy_test.go) pins gold 0 / weak pack = no drink, rich = drink, gold 0 / danger 2 = drink. Consequence: a soak of broke characters shows many `out of mana: backing off` lines; that is expected.

## F4: first Clear after a position jump says everything is unreachable
Root cause (read from code, then reproduced offline): `Fighter.assess` (internal/combat/reach.go) asks `Mover.PathTo`. `Navigator.PathTo` is pure planning on `World.Current()` and never advances the map tracker (it also runs from the guard goroutine, and the tracker is single-goroutine); only walks call `observe()`/`Update()`. After a jump the tracker still holds the previous level (or nothing), so there is no path, `FiringPosition` finds no valid point, the plan is `planNone` and `IgnoreUnreachable` is logged. After one walk the tracker is current, hence the drills' warm-up walk.
Jump kinds: `debug goto` (drill), town portal arrival, waypoint arrival, restart (new game, tracker reset). All are a position the tracker has not read yet, so one fix covers them: the first `Clear` or `Kill` after any of them used to fail the same way unless a walk came first.
Change: `Navigator.Sync()` (observe once; loading maps to `ErrNotInGame`) and the optional `combat.Syncer` interface; `Fighter.Clear` and `Kill` call it first, best effort (errors leave old behaviour). `*nav.Navigator` is what `combat.New` receives in d2bot and d2dcombat, so no wiring change. No spec number changed.
Tests: internal/nav/fixmisc_test.go (`TestSyncAfterSameLevelJump` shows PathTo refuses before Sync and plans after; `TestSyncAfterLevelChangeJump` for portal/waypoint arrival; `TestSyncDuringLoadingThenReady` for a restart), internal/combat/fixmisc_test.go (`TestClearSyncsMapAfterJump`, `TestKillSyncsMap`, fake Mover whose planner knows nothing until Sync).

## Verified
- VERIFIED OFFLINE (nav sim, combat harness): the above tests; the Kill test fails with the Sync call removed.
- UNVERIFIED LIVE: that the real `d2dworld.World.Update` right after `debug goto` has the new level's grid in one call (arrival link wait, `MinRefetch` 2 s rate limit apply: `due()` refetches on a level change at once, but within the same level a short jump under 60 subtiles or inside 2 s of the last fetch is not refetched). The drill premise (F4 note) is itself UNVERIFIED in the bot.

## Live proof recipe
On a hidden slot with the lab binary: `debug spawn` a pack, `debug goto <area>` and call `Fighter.Clear` with no warm-up walk (the DRILLS5 `F` group drill without its warm-up); pass when the result has no `IgnoreUnreachable` entries and returns in well under 4 s of game time for monsters in the open. Repeat after a town portal arrival and a waypoint arrival.
