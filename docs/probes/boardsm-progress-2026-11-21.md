# Quest board state machine: migration progress (2026-11-21)

> status: current. One line per merged card (design `docs/design/fable-review-2026-11-21/quest-board.md` section 5). Live evidence: `docs/probes/boardsm-scenarios-2026-11-21.md` (CHEATED setups: mechanics only).

- cards 1 to 5 (bsm1 scenarios, bsm2 machine, bsm3 store, bsm4 exec + knob `quests.boardSM`, bsm5 holds): merged before this file; see the scenario probe above for their live state.
- bsm6 (TASK-780, merged 2026-11-21): replies carry `changed`, chunk rows carry `reason {code,text,remedy}`, `handin_ready`; live stop-cascades GREEN (Paladin 11, Sorceress 3), reason-always GREEN Sorceress 3 (Paladin 11 BLOCKED twice: lab pace, run cut); handin.* stay pending (the lab quest word 0x0002 does not derive awaiting, U). CHEATED setups, mechanics only.
- bsm7 (TASK-781, this card): kit plan under quests.boardSM, chores trip reports its hold, a clean trip that changes nothing is a hold; live hold.restock and restock-stash-gold GREEN on Paladin 11 (CHEATED setups); Sorceress 3 BLOCKED by the lab pace; broke, too-weak, timeout still pending.
