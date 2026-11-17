# resistorder: the bot's resist model in game.exe order

> status: current (2026-11-17, card resistorder, TASK-611). Fixes bot finding B1 and uses T1 of [decomp37-combat.md](decomp37-combat.md) (merged with the decomp37 branch). Offline: no game.

## Rule applied (game.exe 1.14d)

1. Add the difficulty penalty (Normal 0, Nightmare -40, Hell -100) to the raw stat sum. V read FUN_0057be00 `0x57be5b..0x57be72`, V run model-ab-skresist Q10.
2. Cap at 75 + the element's max-resist stat (magic: stat 0x26). V read `0x57be8f..0x57bec8`. The absolute ceiling of 95 is S (published limit), U (not traced).
3. Floor at -100 (V, sk-resist-trace).
4. Percent absorb (stats 142/144/146/148, capped at 40) applies after the resistance, then the flat absorb. V read and V run (decomp37 T1, Q2..Q5). Poison and physical have no absorb row.

A raw 90 in Nightmare is 50 (the bot used 35); a raw 120 in Hell is 20 (the bot used -25); a raw 175 in Hell is 75 (the bot used -25).

## Where it lives now

`internal/d2/resist.go`: `d2.ResistTable` (a data table: per-difficulty penalty, base cap, cap limit, floor, absorb cap, `FoldAbsorb` flag) with `Effective(raw, maxBonus, difficulty)` and `Intake(res, absPct)`. `Difficulty.ResistPenalty` reads the table. The player read (`internal/d2/player.go`) fills `Player.MaxResist` (stats 38/40/42/44/46) and `Player.Absorb` (142..149); an absent stat reads 0, so a hero without absorb or max-resist gear gets exactly the old numbers except for the order fix. Stat ids 142..149 and 38 are V read from the game.exe table; 40/42/44/46 are S (neighbours in itemstatcost order, U live); whether the player's merged stat list exposes any of them is U.

## Sites that compute effective resist (file, before this card)

| Site | Before | Now |
|---|---|---|
| `internal/combat/rules.go` `PlayerResists.Effective` (was :311) | `max(min(v,75)+p,-100)` | `d2.ResistTable`: penalty, cap + max bonus, floor, percent absorb. Callers: `combat/hard.go` packDanger/hellCtx, `ResistDanger`, `bot/loop.go:947` (antidote count), all via `combat.ResistsOf(p)` |
| `internal/readiness/model.go` `worstResist` (was :357) | `min(v,cap)` then penalty | same table with the check's `cap` param as base cap; `Hero.MaxBonus`, `Hero.AbsorbPct` (filled in `bot/readinessfacts.go`); a `ResistWithPenalty` hero skips the penalty |
| `internal/d2/difficulty.go` `DoctorLine` | cap before penalty | table |
| `internal/leveler/resistgate.go` `ResistGate` | penalty then cap 75 (already right) | uses the table; max bonus not read here |
| `internal/gear/score.go` `ResistValue` | `capRaw = 75 - penalty` (already the game.exe order) | base cap from the table; max-resist bonus U (gear scoring is advice only) |
| `internal/power/margin.go:261`, `power/hero.go` | cap 75 on the panel value | cap 75 + max bonus; percent absorb folded into `Hero.Resist`. The panel value is read as shown (no penalty applied): U TASK-216 |
| `internal/resource/gear.go:42`, `internal/danger/generate.go` | read "as displayed" or area data, no cap logic | unchanged |

Not done, and why: flat absorb (needs the hit size; `d2.Absorb` carries it for a later damage-intake model); magic resist (B2: the readiness elements have no magic row, stat 37 is read but no check uses it); gear policy never equips, so absorb gear scoring (B3) is not added.

## Interplay (documented, unchanged)

Max-resist gear and charms move the cap, not the value. Amplify Damage and Lower Resist subtract from a monster's resist, not the player's; a curse on the player (the bot does not model it) would enter at step 1 as a raw change. A negative effective resist is floored at -100, which is also the floor game.exe uses, so a deep hole stays comparable between the models.

## Diffs in outputs

- `readiness` resist check: Hell raw 120 was Fail (-25), now Warn (20); raw 175 was Fail, now Pass. Nightmare raw above 75 reads higher. Test `TestResistBothBasisCases` updated, table test `TestWorstResistGameOrder` added (sorceress and paladin, three difficulties).
- `combat` `Effective` Hell fire 80: -25 -> -20 (test `TestResistPenaltyAndDanger`, `TestEffectiveGameOrder`). `d2doctor` DoctorLine: Hell raw 90 fire -25 -> -10.
- No golden file under power, survival or danger changed. No UI line changed meaning: the hero sheet shows the raw stat (`cmd/d2bot/readywire.go:50`), the readiness line shows the new value with the same wording.

## Not verified

No live run: the player's stat list may or may not already contain the penalty (TASK-216), and may not expose stats 38/40..46 or 142..149; each reads 0 when absent. Live check: `d2cheat`/dump stats of a Nightmare character with a known amulet, compare `DoctorLine` with the character panel.
