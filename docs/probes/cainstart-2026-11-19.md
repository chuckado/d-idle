# cainstart: why `cain.scroll` stayed RED from the camp (2026-11-19)

> status: current (card cainstart). Tags: V verified (run named), U unverified. CHEATED setup (clvl 17, waypoints, board record): mechanics evidence only.

## Finding (V, builder main lane, Paladin seed 11, lab verbs)

1. The chunk starts in town (`starting chunk`, `step Activate waypoint`) and walks on foot Blood Moor, Cold Plains, Stony Field, Underground Passage. A potion drink at 51 percent life cancels the travel step (`waiting: context canceled`), the board says `queued: waiting for the fight to end`, the level goal walks on. This part works: with a debug line per gate decision the chunk started again about 28 s later in Dark Wood (`preempt=true why=run`).
2. Thirteen seconds later the broke-recovery started (`recovery: start (gold 70 < kit 500)`). The scenario purse was 500 gold; the first town visit (low on mana potions) bought a Tome of Town Portal for 400. `boardPreempts` returns false while the recovery is active (docs/questboard.md: a recovery that has to refill the purse comes first), so the chunk waited behind about 20 virtual minutes of Stony Field and Cave farming and never reached the tree. Nothing was logged and the player's queued reason was empty.
3. Same scenario with a purse of 5000: GREEN (`cain.scroll` Paladin seed 11, run 20261119-144151-5f3a, rc 0, tree reached, scroll taken).

So the RED was a scenario setup defect (a purse below the kit price) plus a silent deferral, not a broken chunk start. Whether the live hero was in a recovery when it sat at the tree is U (the live log is not in this branch); this fix does not change what it does at the tree.

## Change

- Bot: `boardPreempts` now says why it waits (`quest board: start gate: ...` once per reason, and `QuestQueued` shows "Starts after the hero has refilled the purse and restocked." during a recovery); the queued reason is cleared when the chunk starts. Unit test `TestQuestBoardQueuedReasonWhileRecovering` (Sorceress and Paladin) fails before the change (no reason shown; a stale reason).
- Scenarios: the Cain scenarios get a purse of 5000.
- Sorceress seed 3 `cain.scroll` (run 20261119-145134-0332) is still RED for a different reason (the hero leaves for area 17 within four minutes and the game is left; U, not analysed): open.
