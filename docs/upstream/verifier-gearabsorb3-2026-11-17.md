# Verifier: gear absorb r3, the heal-back (2026-11-17)

> status: WIP, 2026-11-17 (card verifygearabsorb3, TASK-626, Opus, singly). Nothing posted.

WIP notes: rule re-derived (FUN_0057bf10, FUN_0057bf80, FUN_0057a980, FUN_0057c6c0, FUN_005541b0 read); game side of seeds 7 and 5001 reproduced byte for byte; hand cases with life above max run on game.exe (life brought down to max by the heal); r3 rebases clean on upstream 895f8a67; build of r3 running.
- r3 claim build rc 0; oracle 0 differ on seeds 7, 5001, 4242 (fresh); hand cases (life above max, flat revive) equal game.exe. README oracle row lacks the grep (test logs "the player died" to stdout: 7,310 lines on seed 7). r4 local (rebased on 895f8a67, README grep, merc wording, pins 10-12, ES pin): pr_guard 0/0, presubmit 0 FAIL 3 WARN.
