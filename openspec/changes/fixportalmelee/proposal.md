## Why

Gate 1 soak 2's one death (clvl 4, Burial Grounds): four healing potions were drunk against five undead, then the town portal cast was cut by melee hits and the scroll was spent (docs/probes/postfix-soak2-analysis.md section 3). A cast takes game time; with several attackers adjacent it is not an escape.

## What Changes

- internal/combat Decide: a TownPortal decision in a fight with 3 or more adjacent hostiles becomes a potion if one is ready, else Chicken.
- internal/resource AbortFight: a PortalOut with 3 or more adjacent hostiles becomes LeaveGame.
- Adjacent count is always filled now (it was only filled when the crowd rule was on). No knob, no numeric default changed: the 3 is the crowd rule's existing constant.

## Impact

Owns the files in tasks.md and docs/fixportalmelee.md. Offline only; live recipe in the doc. No spec delta (no numeric gate changes).
