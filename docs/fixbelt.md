# fixbelt: the Gate 1 rerun belt window

## Evidence
- d2dsoak counts consecutive minute lines outside town with packHP 0 and no "Healing" in beltBoxes (internal/d2dsoakstat/invariants.go). V: re-ran that rule over docs/probes/soak-logs/gate1-rerun-minutes.jsonl.gz on agent/gate1: two windows of 4 or more lines: lines 121 to 124 (4, clvl 6) and **lines 177 to 182 (6, clvl 7, wall 13:21:55 to 13:22:12)**. The 6 matches the reported limit breach. It is early game, not after the 14:27 death.
- V (minute lines): recovery ended at line 166 (gold 124, 6 potions); the bot went back to `Quest: fighting (Level up to 9)` and burned the belt in the Den (area 8) and Blood Moor: potions 6, 4, 0 (hpMin 47 percent, 5 potions used in two minutes). Town visit at line 173 to 174: gold 55 bought one Minor Healing (gold 25). Line 175 to 176: 1 potion left, 25 gold, still fighting in the Den. Line 177: 0 potions. Recovery restarted at line 181 (gold 14), exactly 15 minutes after line 166; a ground Light Healing Potion ended the window at line 183.
- I (inferred from code): `RecoveryCheck` returns "cooldown" for `recoveryCooldownMin` (15) after a recovery ends, unless `Force` (3 deaths in 30 minutes). Gold 25 < kit and shortfall > gold, so only the cooldown kept the quest running with 0 or 1 potion. Kit gate (Blood Moor and Cold Plains only) was not active either, hence the Den fight.
- Not the cause (V from minute lines): belt placement/shift-click/pickup (potions shown in beltBoxes correct, a picked-up potion lands in the belt at line 183); town rebuy (spent all it could: 30 gold per potion, 55 gold). The fork's belt rebase is not implicated by this window. U: whether rebuy would have bought a second potion with 25 gold (no, price 30).

## Cause
Poverty plus the recovery cooldown: a character with no money and a nearly empty belt was allowed to push the quest.

## Fix
`resource.KitLow(healing, t)`: healing potions (belt plus pack) under `recoveryPotionFloor` (default 2, 0 disables). `RecoveryInput.LowKit` bypasses the cooldown like `Force`; the "kit affordable" and "gold covers the kit" cases still win, so a character that can pay is not sent farming. Data-driven (existing knob), no class or level hardcoding. Tests: `TestRecoveryCheck`, `TestKitLow` (internal/resource), `TestLowKitStartsRecoveryInsideCooldown` (internal/bot, mem.Fake rig). The old cooldown test now sets the floor to 0.
Expected effect (U): recovery starts at line 175 (1 potion), the kit gate keeps the farm in Blood Moor or Cold Plains, so the last potion is not burned in the Den and gold reaches a potion price sooner. Risk (U): recovery may re-enter right after a cap end while poor; it still farms XP and gold, quest progress waits.

## Live recipe (not run)
1. Build the bot at this branch tip; use the fork binary of the rerun (lab2-freezeresume) and a fresh uncheated Sorceress in slot 0.
2. Rerun the Gate 1 soak (`d2dsoak run`/`compare -hours 22`) as in docs/soaks/gate1-rerun.md.
3. Pass: `d2dsoak report` shows BeltWindowMin at most 5. Also grep the log for `recovery: start` lines inside 15 minutes of a `recovery: end` (the override) and check quest progress is not starved: the XP/h and clvl at 22 h.

## Shared-file lines for the merger
- docs/BACKLOG.md: `- FIXBELT: Gate 1 belt window located (line 177 to 182, clvl 7, inside the 15 min recovery cooldown with 0 potions and 25 gold); low-kit override of the cooldown; awaits a Gate 1 soak (docs/fixbelt.md)`
- docs/INDEX.md: `- docs/fixbelt.md: Gate 1 rerun belt window: where, why, fix, live recipe`
- CHANGELOG.md (unreleased): `- fix: recovery starts inside its cooldown when healing potions are under the potion floor and the kit is unaffordable (Gate 1 belt window).`
