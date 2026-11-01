# fixmanamodel: the bot's mana model (regen, cost booking, potions)

> status: card fixmanamodel, offline, not verified live. Evidence: docs/research/sk-mana-trace.md (B1-B3, trace of game.exe 1.14d), docs/probes/drills/README.md DRILLS13 (M1) and DRILLS15 (M2, M4). Tags: V read = disassembly, V-d2d = measured on d2d, S = secondary, U = unverified.

## What changed

| Part | Before | Now | Evidence |
|---|---|---|---|
| Regen | `max / 100 s x (1 + bonus/100)` | `ManaPerFrame256 = max(max x 256 / 3000, 1) x (100 + bonus) / 100` (integers), x 25 frames / 256 per second. 45 max: 3 a frame, 0.29 / s, 153 s to fill (was 0.45 / s); 100 max: 0.78 / s; 250: 2.05 / s | V read FUN_005806f0; V-d2d DRILLS15 M4: 45 max, 750 ticks = +2250/256, Warmth 1 the same (test `TestManaRegenMeasured`) |
| Backend switch for regen | n/a | none: d2d and game.exe use the same rule, so one function | V read + V-d2d (equal) |
| Tracker booking | cost booked at the command | `ManaBook`: d2d `BookD2D` pays 1 tick (0.04 s) after the command and a cut cast still pays; game.exe `BookGameExe` pays 0.32 s (8 ticks, Fire Bolt 1, no FCR) after the command and a cut cast (player mode 4, hit recovery) never pays | V-d2d DRILLS13 M1 (1 tick; action frame 7 ticks later), DRILLS15 M2 (7 of 7 cut casts paid on d2d); game.exe: V read trace D1/D3. Backend from `Fighter.D2D` |
| Sim potion | +40 at once | +40 spread over 128 frames (5.12 s), a second potion joins what is left | V read (misc.txt len, FUN_005be3f0); the join rule U |

Code: `internal/resource/manamodel.go` (named constants `ManaRegenFrames`, `FramesPerSec`, `ManaPotSecs`, models `BookD2D`, `BookGameExe`), `internal/combat/hard.go` (`manaBooking`, `modeGetHit`), `fight.go` and `castfirst.go` (book at send). `Scenario.ManaFullSecs` still overrides the sim regen (sensitivity test). Tests: `internal/resource/manamodel_test.go`.

## Unverified or not modelled

- game.exe action-frame delay of 0.32 s is only for Fire Bolt 1 at no FCR; other skills and FCR move it (U). The four start-paying skills (Inferno, Telekinesis, Thunder Storm, Hydra) pay at once on game.exe (V read) and are booked late here: a small, short-lived error only on the Hell flee/chicken lines.
- Cut detection reads the player mode once per fight tick; a hit shorter than a tick is missed (U). Mode 4 = hit recovery is S (D2 PlayerMode) plus V-d2d pmode reads.
- Flat mana recovery (stat 26) and Sorceress potion doubling (x2) and rejuvenation timing are not modelled; the sim's +40 stays the sim's own figure.
- Live d2d confirmation of the booking model is only DRILLS13/15 for Fire Bolt; there is no game.exe run.

## Follow-ups (thresholds that read the model; NOT retuned here)

`Input.RestSecs` feeds `restMaxSec` (rest vs drink / melee / go on, preengage.go, town.go starved test with `restStartManaPct`). With the real regen an early Sorceress (45 max) now rests 1.5x longer for the same top-up, so more packs hit "rest too long". The unit tests `TestPreEngage` (mana 12 became 17) and `TestRestsBeforeAFarPack` (start mana 15 became 22) needed inputs moved for that reason. Review: `restMaxSec`, `restStartManaPct`, `castsPerMonster`, and the Hell `ManaBurn*` constants against the new curve with a soak, not by hand.

## Live proof idea

A drill (group M style) that casts Fire Bolt at a target dummy with a throwaway Sorceress at level 8 for 60 s, logs per tick the bot's modelled mana (start mana + `ManaPerSec` x t - booked costs from `ManaBook.Take`) next to `dump stats` raw[8]; pass when the difference stays inside one cast cost on both a d2d slot and a game.exe instance, with a second run where a `debug mon set dmg` blow cuts casts (expect no refund on d2d, refund on game.exe).

## Shared-file lines for the merger

- INDEX: docs/fixmanamodel.md: regen per-frame whole 256ths, backend mana booking, potion over time.
- CHANGELOG: the bot's mana model follows game.exe/d2d: early pools regen 1.2x-1.8x slower than the old 100 s line, casts are booked at payment, mana potions arrive over 5 s.
- BACKLOG: live proof of the mana model; review of rest thresholds against the new curve.
