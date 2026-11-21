# Act 1 boss scenarios: Blood Raven, Tools of the Trade, Andariel (card act1more, 2026-11-20)

> status: current. Tags: V verified (named run or test), S secondary, U unverified. Every setup is CHEATED (level, gold, waypoints, potions by SETUP verbs): mechanics evidence only, never balance or play evidence.

## 1. What was added (V: `go run ./cmd/d2scenario validate`, `go test -race -count=20 ./internal/questspec`)

- `tools/provingground/scenarios/act1more.json`: `act1.burial.blood_raven` (quest record 2), `act1.tools.malus` (record 3, U), `act1.andariel.slaughter` (record 6). Board-taps driver, Paladin and Sorceress, seeds 11 and 3; death and chicken bounds as `log_line` assertions (`min: 0`).
- `internal/questspec/act1boss_test.go`: the three boss chains kill their boss in an area entered first, with a kill-or-cleared post and `park` on failure; the Malus (`hdm`) is picked before the hand-in; the title "Bane of Blood Raven" exists in `internal/watch/hall_titles.json` as a kill rule on Blood Raven x1 (data check only: no Hall reader in a run); the scenario file covers each boss quest once.
- Not used: `api_reply` / `trigger.api` (only on the unmerged pgcover2 branch, not in main).

## 2. Runs (builder, main lane, bot of this branch, lab d2d)

| scenario | class/seed | verdict | job |
|---|---|---|---|
| act1.burial.blood_raven (bankFirst default) | Paladin/11 | RED (see finding 1; also my assertion defect: `max` without `min: 0` demanded a death) | 20261120-184924-bb3e |

## 3. Findings

1. FINDING (V, log of 20261120-184924-bb3e): a hero given 8000 gold has the bank chore deposit 6900 into the stash and the shop spend 800 in the first minute (`town budget ... byChore bank -6900`), leaving 130 to 300. After the Start tap the chunk's first step `quest.sisters_burial/chores` "ran but is still needed (evidence 2)" for 28 wall minutes (quest board state `needs_attention`, record 2 never written, no trip to the Burial Grounds): the purse is empty while the stash holds the gold and no chore withdraws it for potions. Same family as Den finding 1 of `provingground-act1next-2026-11-20.md`. Bot defect (U which side: chore withdraw or the bank keep line); card: `town: withdraw stash gold when the kit chore is short`.
