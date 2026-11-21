---
status: accepted
date: 2026-11-21
decision-makers: user (popups, Fable review 2026-11-21, quest board)
---
# 0025: Quest board lifecycle: stored runs, accept chain, stop cascade, restart-persistent parks

Source: user popups, 2026-11-21, on [quest-board.md](../design/fable-review-2026-11-21/quest-board.md) section 7 (Q1 to Q11). Amends design 1.5 ("never handler progress") of [quest-board-2026-11-16](../design/quest-board-2026-11-16.md). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md) (arbiter, Proving Ground scenarios) and [ADR 0022](0022-hero-knowledge-reads-never-cheat.md) (evidence is read, not written).

## Context

Six board bugs share one cause: the board derives everything from game facts each tick and stores no player intent, so a stop, a park or a "seen done" is forgotten. Quest evidence can also move backwards (a finished part regresses).

## Decision

1. The board stores the lifecycle of a run (status, hold or park, completion latch) in the hero's plan file (Q1). Evidence stays derived and is an input, never the state. Knob `quests.boardSM`; the plan file is dual-written while it is staged (Q11).
2. Start on a later part of a quest accepts the unfinished earlier parts; the toast names them (Q2).
3. Stop on a part cascades to the parts that wait for it; the toast names them (Q3).
4. If exact evidence regresses for a finished part while later parts are accepted, the run PARKS AND ASKS ("Run it again" or "Stop"); field work is never redone unasked (Q4).
5. A parked run stays parked after a bot restart (Q5).
6. The hero waits at the giver for the hand-in tap 30 game minutes (`quests.handinIdleMin`); then the chunk stays `awaiting` and the tap sends the hero back (Q6).
7. One quest at a time: a Start while another quest runs is refused and the refusal names the running one. Revisit after Act 2 (Q7).
8. Priority unchanged (Q8): survival, player orders and Pause, town chores, gold recovery and the Farm gold order, then the chunk, then level and farm goals.
9. A guard refusal that a town visit fixes (potions, repair, stash gold) triggers an immediate restock trip and the run continues; no failure is counted (Q9). Relates to [ADR 0028](0028-one-supply-need-town-planner.md).
10. A hold the bot cannot resolve in time parks with a reason and a Needs-you row: 20 game minutes for chores, 60 for training holds (Q10).
11. The coordinator turns the review's section 5 into the ten cards and runs them (Q11).

## Consequences

- The plan file gains run records and latches; "stored intent, derived facts" replaces "never stored".
- Each rule needs a board scenario in the Proving Ground (appendix B of the review) before the knob default moves.
- A second quest cannot be queued until the post-Act 2 revisit.
