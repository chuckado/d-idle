# fixdenbounce: the Den of Evil bounce (evidence stage, no fix yet)

> status: EVIDENCE ONLY, root cause not established, no code changed. Origin: [postfix-soak2-analysis.md](probes/postfix-soak2-analysis.md) section 5 and ranked row 3 ("FIXDENBOUNCE"). Tags: V = read from the soak 2 bot log (`logs/bot-0.log` in the postfix2 run directory on the worker host, 62,496 lines, read-only grep), S = secondary, U = unverified.

## What DENBOUNCE is
The bot enters a cave level (Den of Evil 8 from Blood Moor 2, Cave Level 1 9 from Cold Plains 3) and is back in the outdoor level a few milliseconds of wall time later, with no walk, no kill and no log line in between. Repeated, it is a shuttle that only the border-loop breaker (10 crossings per minute, 180 s skip) ends.

## Evidence (V)
- Counts over the run: `Entered Den of Evil` 248, `quest.den_of_evil/to-den` step runs 121, `area skipped ... border-loop` 110 (analysis, V).
- Cycle at 21:12:29 to 21:12:30 (log lines 4800 to 4866), repeated 8 times in 0.55 wall s until `area skipped Den of Evil border-loop` at line 4862. One cycle, in order: `state clearing -> traveling (Quest: travelling Go to the Den of Evil)`; `nav exit to 8 gate (135,140) border=false`; `Entered Den of Evil`; `task step quest.den_of_evil/to-den outcome ran`; `state traveling -> clearing (Quest: fighting Clear the Den of Evil)`; `d2dworld: arrived in level 8`; `Entered Blood Moor` 10 to 20 ms later; `action cancelled: re-plan, exploring crossed an area border (target area 8)`. No `nav exit to 2`, no `nav: stepping away from a cave mouth`, no `combat`, no `clear killed nothing` line between `Entered Den of Evil` and `Entered Blood Moor`. So the character leaves the Den without `ExitTo` and without the mouth step-away.
- Same shape for Cave Level 1 (lines 1560 to 1586): `nav exit to 9 gate (95,260)`, `Entered Cave Level 1`, `Entered Cold Plains` about 0.15 s later, 10 cycles, then `area skipped Cave Level 1 border-loop`. Both levels log `arrived in level N: 0 adjacent entries` (V: 215 of 241 Den arrivals and 407 of 467 Cave 1 arrivals; "adjacent" counts border-linked outdoor levels, d2dworld/world.go:374, so 0 is normal for a cave and is NOT the cause).
- Fallout (V): `exit to Cold Plains: nav: the level changed during the walk: left level 2 for 3 but ended in 8` blocks step `level.9` (line 4876): the walk to the Cold Plains exit ends in the Den.
- The first Den visit (line 238, 21:02:30) was an accident of a roam in Blood Moor (`exploring crossed an area border (target area 2)`), and the bot left through `nav exit to 2 gate (35,135)` correctly.
- The quest word: `den_of_evil=0x0010` from 21:02:30 until 22:13:59 (never `cleared`), then `0x2012`, `0x2001` and `Quest done: den_of_evil` at 22:14:00 (line 38790). The planner still picks `run=den_of_evil kind=quest area=8` at 22:16, 22:25 and 22:44 (lines 40371, 46034, 51533): `runFor` (internal/bot/leveling.go) maps area 8 to a quest run whether or not the quest is done, and `to-den` is gated by `stageAtLeast(StageCleared)` only in the task graph (internal/task/act1.go denNodes), not in the leveling picker. The analysis' "UNKNOWN why the quest is not marked done" is answered: the word stayed 0x0010 because the objective was not met until 22:14.

## Root cause (what is and is not established)
- Established (V): the leave is silent and does not go through nav.ExitTo or the mouth step-away, and it happens right after the arrival at the return tile of a cave (d2d places the character inside the mouth's trigger area, docs/d2d-loop.md, S for this doc).
- Not established (U): which call walks the character off the cave arrival cell and fires the return trigger. Candidates read in code, none provable from INFO lines: (a) `Runner.sweepPass` first calls `touchWaypointHere`, then `Fighter.Clear` (approach to a hostile near the arrival cell crossing the return tile; combat's `nearMouth` guard is off unless `MouthGuard` is set, config.go:56), then `Mover.Explore`; (b) `nav.walk`'s mouth step only reacts to exits listed in the level map and skips an exit the character stands within 0.5 subtile of (walk.go mouthStep), so an arrival ON the return tile is neither stepped away from nor penalised; (c) the d2d mover itself on a first click. Writing a fix for any of these now would be a guess.
- The 25 `ended in 8` blocks and the 705 step-aways are the nav side of the same trigger problem: penalizeExits (nav.go) is a cost not a wall and skips exits within exitAvoidR+2 of the goal.

## What is needed to fix it (the question)
1. A debug-level slice of one cycle: the nav `nav walk callers=` debug line (walk.go) names the caller of every walk. Run the live recipe below with the bot log at debug and read which of (a), (b), (c) leaves the Den.
2. Policy question for the user (not for this primitive card): should the leveling picker skip `run=den_of_evil` once `Quest done` is logged (answer from the log is yes, 3 re-picks after done), and should it keep re-picking a quest area after `questSweepRetries`? Both are picker policy and paused cards cover them.

## Live recipe (not run)
Per docs/probes/drills/README.md, worker host, one hidden slot, throwaway Sorceress, lockstep, level 5, no Den cleared:
`D2D_AGENT=fixdenbounce tools/agentrun.sh fixdenbounce --ttl 900 -- tools/d2dslot.sh run -t 900 -- d2drill -group DW -only DW.m1 -v -label fixdenbounce-pre`
DW.m1 and DW.m2 (internal/drill/dfield.go fMouth, PlaceAtMouth) already stand the character in a cave mouth trigger of Cold Plains; add DW.m3 for Blood Moor to Den of Evil (area 2, 8): `PlaceAtMouth(outer 2, cave 8)`, then run the bot's `quest.den_of_evil/to-den` and `clear` steps, pass when the character stays in area 8 for 60 virtual seconds with at most 2 `Entered Blood Moor` lines. Capture with the bot log at debug level and grep `nav walk` callers and `Entered`.

## Shared-file lines for the merger
- docs/BACKLOG.md: FIXDENBOUNCE stays queued; note "evidence in docs/fixdenbounce.md; leave path unproven; needs a debug-level live slice (DW.m3)".
