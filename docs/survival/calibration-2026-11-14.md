# Survival calibration from the sandbox replay (card survcal, 2026-11-14)

> status: current (TASK-460). Acts on the findings of docs/survival/replay-2026-11-14.md. Tags: V verified (how), S secondary, I inferred, U unverified. Offline, not live.

## What changed (all constants are data in `internal/survival/params.json`, each tagged there)

| Fix | Constant | Before | After | Reason (tag) | Test |
|---|---|---|---|---|---|
| a | `predTakenFloorPerKill` | none | 1 life per kill | observed damage per kill was 0.3 to 3 life; a predicted value near 0 turned the ratio into the clamp (I) | `TestPredTakenFloor` |
| b | `kMax` | 8 | 3 | the hero lived 7.8 h with one death while k sat at 6 (I) | `TestParamsCalibration`, `TestKHeroUnreadIsBounded` |
| c | `mercPriorMinutes` | 40 | 25 | median of 11 measured merc lives, one hero (S) | `TestParamsCalibration` |
| d | `mercWindowMinutes` (the `either` row only) | 30 | 120 | five Tamoe merc deaths were 30 to 60 min apart; revive rule and chickens keep the 30 min window (I) | `TestMercDeathsAcrossTheWideWindow` |
| f | `kHeroUnreadMax` | none | 2 | with the weapon unread (power quality default) k_hero only measures how wrong the default kill speed is; it is capped, so the plan tables decide instead of "too hard" (I) | `TestKHeroUnreadIsBounded`, `TestUnreadOffenceDoesNotShutEveryZone` (Paladin, Sorceress) |
| e | zone entry line | no def, block, weapon | `def block dmg dmgpct ar` | additive logging, no behaviour change; only the terms the stat vector has | `TestZoneEntryAndSafetyLines` (Sorceress, Barbarian), `TestHeroOfEntryReadsTerms` |

`survival.Input.OffenceUnread` carries the quality flag from the hero (`power.Hero.OffenceUnread`, quality default); the bot wiring (`cmd/d2bot/survivalwire.go`) and the replay set it. `levelsim` reads the new fields when a log has them (offence derived, defence measured, block derived) and otherwise stays default and capped.

## Replay before and after (V: `levelsim -controller` on the committed fixture, the Tower builder, one job; before from the survreplay output file, after in the new `.out.txt`)

| Balanced dial | Before | After |
|---|---|---|
| S-B pass rule (bad stays under the enter line) | FAIL 6 of 14, quiet 8 of 8 | FAIL 6 of 14, quiet 8 of 8 (unchanged, see below) |
| `k_hero` at the end | 6.20 | 2.00 (the unread cap) |
| Minutes Cave Level 1 chosen (of 445) | 163 (cautious 212, bold 84) | 17 (cautious 91, bold 0) |
| Minutes judged `risky` | 249 | 86 |
| Planner picks the shadow agrees with | 2 of 45 | 5 of 45 (cautious 6, bold 6) |
| Minutes the shadow is where the hero was | 301 | 233 (it now also names Dark Wood 222, Black Marsh 76, Stony Field 62, Jail Level 1 39) |
| Black Marsh named again after the 15:27 death | 17:57 (2 h 30 min) | 17:57 (same) |
| Empty candidate sets | 0 | 0 (all dials) |
| Revive after the 3rd merc death (15:26) | refused (`merc_left_dead`) | refused; the 1st death at 15:22 now reads `yes` (the real revive came 54 s later) |
| Tamoe Highland (merc died 5 of 5 visits) chosen | not closed by cost | 3 minutes (19:04 and 19:34 to 35, the first visit and the next before the second death is on the books), then closed |
| Minutes with a dead merc where revive reads `yes` | 5 of 93 | 43 of 93 (the merc term now lifts more zones; the real hero revived 12 of 12) |

The S-B pass rule is unchanged because the stay table uses the zone entry hero and the margin alone (no ledger): none of the changed constants enter it, and this log still has no weapon or defence terms. It cannot pass or fail differently until a log carries the new fields. The enter lines were not touched; the 6.79 candidate line of the replay report is NOT adopted (it would close Stony Field and Dark Wood to a hero that farmed them for hours).

## What the pass rule can and cannot show with n = 1

- It shows a gross failure of the margin (a model that calls a zone `safe` where the hero lost a merc 11 times in 14 bad stays). It cannot tune a threshold: one hero, one build, the quiet class is two trivial fields.
- 11 of 14 bad stays are merc losses, which the pass rule counts the same as hero chickens. The merc term (level, hold, dps) is the least measured input; a pass or fail here mostly judges that term.
- The structural checks (no re-entry, no empty set, second revive refused) hold with and without the fixes, so they are not evidence of calibration. The agreement with the planner is NOT a target: the planner's picks lag the margin's by two to three levels.
- `k_hero` capped at 2 is a guard, not an estimate. Item 5 of the replay (Cave Level 1 and 2 predicted 11 XP per minute against 162 and 523 observed) is still open and sits in the Economy, not in these constants.

## Evidence needed before S-D

1. Fresh runs whose `zone entry` lines carry the new terms (a Paladin and at least one caster archetype), so `k_hero` is measured with a read weapon and the cap is not used.
2. The A/B (shadow against planner on the same seed) for chickens, deaths and merc deaths per hour; the replay cannot say whether following the shadow would have changed them.
3. A second hero with merc deaths in the same zone to confirm `mercWindowMinutes` and `either`; the merc life distribution (11 intervals) behind `mercPriorMinutes`.
4. A real revive quote and shop price from the `shop`/`npc` lines for `valuePerGold` (U).

Commands: `go run ./tools/levelsim/cmd/levelsim -controller internal/survival/testdata/replay-sandbox-clvl11to14.txt -install /gamedata -arch bruiser -dial balanced` (and `cautious`, `bold`; `LEVELSIM_ROWS=1` for rows). Full output: `internal/survival/testdata/replay-sandbox-clvl11to14.out.txt`.
