# PR proof: a frozen monster thinks again at the freeze end + aidel (FREEZERESUME)

> status: BUILT, local candidate `fix/freeze-resume` on the worker host (base upstream main 72132b99); differential run before and after on real d2d builds; **NOT READY**: no independent verifier, no det state hashes. Nothing pushed or posted (card freezeresume, 2026-11-01). Follows the verifier's finding in `docs/research/pr-armorhit-verifier.md` (branch agent/verifyarmorhit, section 1: the end + aidel resume). Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-freezeresume-draft.md](pr-freezeresume-draft.md).

Tags: **V read** disassembled with upstream's `tools/emu/disasm.py` by this agent (addresses given); **V run** game.exe's code run under upstream's emulator in this agent's own oracle; **D** 1.14d tables from the MPQ; **S** d2d source at 72132b99; **I** inferred; **U** not traced.

## Answer

When a monster's freeze (state 1) ends, game.exe does not let it act at once: the state's end callback drops its pending AI events and files the next think at the end + MonStats aidel (13..15 frames for most rows, 6..15 over all rows by difficulty, 0 read as 15). While it is frozen every think that comes due is dropped. d2d let the monster think at the freeze end (or at its old `next_act`, whichever is later). On the differential, every case where game.exe freezes the monster (plain monsters and minions) differs on upstream main (12,503 of 12,548 and 12,459 of 12,516 on two seeds; the rest agree by coincidence of the old `next_act`); with the claim commit none does.

## Rule re-derived from the disassembly (V read, this agent)

| Address | What it does | Tag |
|---|---|---|
| FUN_0057b230 0x57b2bc..0x57b3e9 (EAX unit; attacker, length) | for a plain monster at coldeffect < 0: length / MonsterFreezeDivisor; end = frame + length; no state-1 list: a new one (FUN_006251f0) with end, callback **0x57b170** (FUN_00625ce0 at 0x57b36f, the only place that callback is registered in the image), attached (FUN_00626e10), state bit 1 on (FUN_00639db0); a list already there: its end is raised to the new end only when later (FUN_00625330 / FUN_00625310), else kept; timer 0xc at the list's end (new or kept); then **FUN_00540e60(unit, 2)** removes every pending type-2 event of the unit and event 2 is filed at frame + length + 1 | V read |
| 0x57b170 (state 1's callback, not a function entry in upstream's game-functions.tsv; ECX unit, EDX state; the list) | a dying unit with FUN_0063a4a0 set: nothing; else state off (FUN_00639db0(unit, state, 0)), FUN_00553570; for a monster (type 1) with a MonStats row (monster data +0 at unit +0x14): aidel = byte row +0x4f, or row +0x4f + difficulty (game +0x6d) when game +0x6a or game +0x74 is non-zero; aidel 0 -> 15 (`jg`); FUN_00540e60(unit, 2) and event 2 filed at game +0xa8 (the frame) + aidel. Only event 2 is refiled; no other type | V read 0x57b170..0x57b21e |
| game +0x6a | the game type, 3 for a single player game (docs/research/re/net-join.md, packet 0x67 +0x11), so offline the difficulty column (aidel / aidel(N) / aidel(H), MonStats +0x4f / +0x50 / +0x51) is read | V read (upstream doc), I for the lab |
| FUN_005a7f80 (ECX game, EDX unit; type, a, b) | type > 0xe: nothing; types 3..12 through the byte table 0x5a7fe8 (00 00 00 01 01 00 01 01 01 00) and the jump table 0x5a7fe0 (0x5a7fbf direct, 0x5a7fa8 gated): **gated are types 0, 1, 2, 6, 7, 9, 10 (and 13, 14, which have no handler)**; ungated 3, 4, 5, 8, 12. A gated event on a unit with state 1 (FUN_00639df0) that is not dying (FUN_005541b0 = 0) is dropped, not refiled. Handlers (table 0x6e2490): 0 0x5a7ba0 / 1 0x5a7be0 (the mode's own events), 2 0x5b1740 (the AI driver), 6 0x5a7f00, 7 0x5a4370, 9 0x56fe40, 10 0x5a7f70 (the skill-set target's end), 12 0x5a7ef0 (state lists) | V read; table bytes V run (read from the image) |
| 0x5a7ef0 -> FUN_00627460(unit, frame) | walks the unit's timed lists (+0x10 & 2) and expires those whose end (+0x18) <= frame (FUN_00626c00 -> FUN_006269f0, which runs the callback +0x38 with ECX the owner +4, EDX the state +0x14, the list pushed). So a list whose end was raised does not expire at the old timer | V read |
| FUN_005416b0 (the queue's file) | an event filed for a frame <= the current frame (game +0xa8) goes to the current + 1; later frames stay | V read 0x5416e3..0x5416ed |
| FUN_0057aae0 (state 21, stun) | callback 0x56e900: state off, client update, no AI event touched | V read |
| FUN_0057af80 (state 0xb, chill) | callback 0x57ad80: states 0xb and 0x6b off; nothing on the AI; FUN_005a7f80 gates on state 1 only. **The chill drops and refiles nothing** | V read |

Consequences (I from the rows above, V run in the oracle):

- The end + 1 event FUN_0057b230 files never runs as a think: at the end the callback removes it; when an earlier, longer freeze is running it comes due while frozen and is dropped. The next think is always end + aidel.
- A second freeze during the freeze raises the end (when later) and removes the pending events; the callback at the final end files end + aidel. A second freeze in the delay window (between the end and end + aidel) removes the pending end + aidel think; its own end + aidel is the next think.
- A zero-length freeze (length / divisor = 0: lengths under 2 / 4 in Nightmare / Hell) makes a list ending now; its timer is moved to frame + 1 (FUN_005416b0), so the monster's next think is frame + 1 + aidel (I; d2d upstream never makes such a freeze, the freeze-rules candidate skips them). Not claimed.
- Only freezes (state 1) carry this. Champions, uniques, superuniques, MonStats boss rows and hirelings are never frozen by FUN_0057b230 (they are chilled), so in game.exe they never get the delay.
- Other events dropped while frozen and never refiled: types 0, 1, 6, 7, 9 and 10 (for instance 10, the skill-set target's end, FUN_00573120). Not claimed (ponytail).

aidel values (D, MonStats.txt, aidel / aidel(N) / aidel(H)): 15/14/13 on 393 rows (skeletons, zombies, fallen...), 15/15/15 on 255 (NPCs, critters), 15/13/12 on 30 (Diablo, council), empty (0 -> 15) on 18, 15/13/11 on 13, 1/1/1 on the 5 hireling rows, 15/9/7 on 5, 15/13/10 on 4 (Blood Raven, Griswold, Summoner), 15/12/9 on 3, 15/11/9 Andariel, 15/10/6 Mephisto, 15/12/8 Izual, 6/6/6 the Valkyrie, 15/14/12 Baal's clone. No champion or unique modifier: 0x57b170 reads the class row only.

## d2d (S, 72132b99)

- `monster_update` (components/game/ai.cpp:1831): while `now_ms < stun_until` the monster stands (NU) unless in GH / BL; from the first tick at or after `stun_until` it goes on, and `think()` runs as soon as `now_ms >= next_act` (ai.cpp:653).
- `stun_until` is set by the freezes: Glacial Spike's hit (missiles.cpp:662), Frozen Armor (states.cpp:160) and the Act 3 merc's Frozen Armor (states.cpp:194), all as `max(stun_until, end)`; and by the stun (state 21, strike.cpp:448) as a plain assignment. Nothing refiles a think at a freeze's end.

## The claim commit (c6cc4f12)

- `Monster::stunned_until` (ai.hpp): the stun's end, written next to `stun_until` at the one stun site (strike.cpp:448). A stand whose end differs from it was last set by a freeze.
- `monster_update` (ai.cpp, right after the dead-monster return): at the first tick at or after `stun_until`, if it was a freeze, `next_act = stun_until + aidel x 40 ms` (aidel of the monster's row and difficulty, 0 -> 15); then both ends are cleared (the state is off). Replacing `next_act` (not `max`) is FUN_00540e60's removal of the pending events.
- docs/research/re/monster-ai.md: five lines on the gate and the refile.
- `// ponytail:` the other dropped event types; a freeze that ends under a longer stun (d2d keeps one stand end).

Why `stunned_until` and not a `freeze_until` written at the freeze sites: the three freeze sites are each rewritten by an open or local candidate (missiles.cpp:662 by fix/freeze-rules-v2, states.cpp:160 by fix/armor-hit); marking the stun instead leaves them untouched, so the candidates merge cleanly and their freeze rules get the resume for free.

## Differential (game.exe natively vs real d2d builds)

- game side: `tools/emu/freeze_resume.py game` (copy in this repo). Run natively: FUN_0057b230 (every freeze, both the first and the second), FUN_005a7f80 (every event as it comes due), 0x5a7ef0 -> FUN_00627460 (the expiry walk), the callback 0x57b170, and inside them FUN_0057af30, FUN_0063e9f0, FUN_005a0180, FUN_0063ee90. Hooked: the queue (FUN_005417d0 files, FUN_00540e60 removes by type; dispatch at the frame filed), the state bits, the list allocation / lookup / end calls (laid out as FUN_00627460 reads them), FUN_00626c00 (unlinked; the callback then run as FUN_006269f0 does), the chill FUN_0057af80, one MonStats row (the case row's aidel bytes, zombie1's coldeffect, the boss byte for kind 5), DifficultyLevels' MonsterFreezeDivisor (D), game +0x6a = 3. The AI driver 0x5b1740 is hooked: its frame is the output (it then waits aidel). The raw freeze length passed is the case's frames x the divisor, so the end is exact.
- d2d side: `test_freeze_resume --freeze-resume CASES` (new tests/game/test_freeze_resume.cpp, copy tools/emu/freeze_resume_mode.inc): a zombie1 (Zombie AI: with its foe in melee every think is a swing, so the think is observable as the first A1 / A2) carrying the case row's real aidel for the case's difficulty (copied from that row's loaded MonStats data), Boss kind / boss column set, `next_act` = the pending think, a prior freeze as `stun_until`; frozen through the REAL `Fight::strike` (Glacial Spike's freeze path, missiles.cpp:662) for the case's frames; then the real `monster_update` a tick at a time (a second `Fight::strike` at second_at). Prints `id first second`.
- cases: `freeze_resume.py gen 30000 SEED` (LCG `x * 1664525 + 1013904223`, seed masked to 32 bits, top 24 bits): 7,776 boundary cases (3 difficulties x 16 rows covering every aidel triple of MonStats x 6 kinds x 27 shapes: lengths 1 / 2 / 13 / 50 / 250 with pending thinks 1 / 14 / 60, prior freezes 1 / 30 / 300, second freezes inside the freeze, at its end, inside the delay window (31..45 for a 30-frame freeze and aidel 15) and after) + 30,000 random (length 1..250, pending 1..80, a quarter with a prior freeze 1..300, a third with a second freeze 1..length + 60, 1..120 frames). Seeds 0xF2E50001 (cases md5 065d3dd3260a, game ccc09ad55409) and 0x51A70002 (48fcf5ae4a58, 45ab193bf3d5); the d2d harness seeds its spawn / think `Rng`s with 32-bit values per case.

Kinds: game.exe freezes plain monsters and minions only; champions, uniques, superuniques and MonStats boss rows are chilled by FUN_0057b230 (no state 1, no delay: their next think is the pending one). d2d upstream freezes them all (the freeze-rules deviation C3, not this claim), so those rows differ before and after and are counted apart.

| build | seed | frozen in game.exe (plain + minion) differ | not frozen in game.exe (specials) differ |
|---|---|---|---|
| before: 72132b99 + the tooling commit's two files | 0xF2E50001 | **12,503 of 12,548** | 21,895 of 25,228 |
| after: fix/freeze-resume head efacc7f4 | 0xF2E50001 | **0 of 12,548** | 25,136 of 25,228 (freeze-rules C3) |
| before | 0x51A70002 | **12,459 of 12,516** | 21,802 of 25,260 |
| after | 0x51A70002 | **0 of 12,516** | 25,148 of 25,260 (C3) |

Per class (seed 0xF2E50001, before -> after): plain 6,201 -> 0 of 6,223; minion 6,302 -> 0 of 6,325; champion 5,429 -> 6,227 of 6,248; unique 5,423 -> 6,222 of 6,245; superunique 5,509 -> 6,313 of 6,335; boss column 5,534 -> 6,374 of 6,400 (the specials are the C3 rows above; they agree before only where d2d's freeze end happened to fall before the pending think).

Per aidel value, frozen cases (seed 0xF2E50001, before -> after; seed 0x51A70002 in brackets):

| aidel | cases | before | after |
|---|---|---|---|
| 1 (hireling rows) | 800 [775] | 794 [767] | 0 [0] |
| 6 | 1,009 [998] | 1,009 [995] | 0 [0] |
| 7 | 242 [244] | 241 [243] | 0 [0] |
| 8 | 283 [264] | 281 [264] | 0 [0] |
| 9 | 791 [820] | 791 [818] | 0 [0] |
| 10 | 496 [510] | 492 [506] | 0 [0] |
| 11 | 520 [536] | 520 [535] | 0 [0] |
| 12 | 1,060 [1,011] | 1,051 [999] | 0 [0] |
| 13 | 1,606 [1,595] | 1,591 [1,579] | 0 [0] |
| 14 | 1,009 [989] | 1,004 [987] | 0 [0] |
| 15 (MonStats) | 3,912 [3,971] | 3,910 [3,963] | 0 [0] |
| 15 (blank row, 0 -> 15) | 820 [803] | 819 [803] | 0 [0] |

The before mismatch is game.exe later by aidel (the d2d monster swings at the end) or, with a pending think later than the end, d2d at the old pending frame and game.exe at end + aidel (the freeze removed it).

Composition with the freeze-rules candidate (V run): a detached tree at fix/freeze-rules-v2 (409903a6) merged with fix/freeze-resume, the 12,567 Normal cases of seed 0xF2E50001 (MonsterFreezeDivisor 1, so the harness length equals game.exe's): frozen cases 0 of 4,153; specials' first think 0 of 6,544 (no prior freeze) against 5,383 on upstream main and 6,531 with this claim alone. The rest of the specials' rows in that run are harness artefacts, not deviations: a prior freeze set on a special (game.exe cannot freeze one) and the think after the first swing (the oracle's hook waits aidel; d2d waits for the swing's end).

## Isolation

- before tree = upstream 72132b99 with only the tooling commit's two files added (worktree `fresume-before`, files byte-identical to the head's, checked with `cmp`): its output is the "before" rows.
- claim alone: c6cc4f12 has no test; the head = claim + the two files, so head - before = the claim. The tooling commit adds no d2d code.

## Where the oracle and this reading could share a mistake (for the independent verifier)

1. **The dispatch frame.** The queue is hooked: an event filed for frame X runs at frame X in the oracle, and d2d thinks at the tick where `now_ms >= next_act`. FUN_005416b0 moves an event for a frame <= now to now + 1 (V read), but the per-frame walk (FUN_005414d0 -> FUN_00541240 / FUN_00540f60 ..., buckets frame & 63) and its comparison were not run. If game.exe ran a timer filed for X at X + 1, both the callback and its refile would shift and the resume would be end + aidel + 1 or + 2. A verifier should run the real queue (FUN_005416b0 and the walk) for one freeze.
2. **Order inside a frame.** The oracle runs a frame's due events in filing order and applies a second freeze before that frame's events; d2d strikes before the tick's `monster_update`. A second freeze landing exactly on the first freeze's end frame is the one case where the order matters (both sides keep the freeze going: the list is raised before the old timer fires).
3. **What counts as "acts".** game.exe: the AI driver 0x5b1740 is entered (the driver may then stand, so a think is not always visible on screen). d2d: the zombie's first swing. A d2d think that only stands would not be seen by the harness; with its foe in melee the Zombie think always swings (monsters.hpp:1409), so here the two coincide. The claim moves the think, not what the think does.
4. **game +0x6a.** Set to 3 (single player). With 0 (and +0x74 0) game.exe would read the Normal aidel at every difficulty; the offline lab is single player (I).
5. **Inputs the oracle fakes.** One MonStats row (aidel bytes from MonStats.txt, the coldeffect of zombie1), the monster never dying, state 0x36 off. The d2d body is always a zombie1 with the case row's aidel copied in (the row's own AI is not run on either side).

## Residuals and limits (not claimed)

- The specials are not frozen in game.exe at all (freeze-rules C3, fix/freeze-rules-v2; Frozen Armor: fix/armor-hit). Until those land, d2d still freezes them and now also delays them by aidel.
- A zero-length freeze (Nightmare / Hell, length under the divisor) delays the next think to frame + 1 + aidel in game.exe (I); d2d upstream never makes one, and the freeze-rules candidate skips them.
- d2d lets a get-hit (GH / BL) that ends during a freeze think once (ai.cpp:1831 lets GH play out); game.exe drops that think too (types 0 / 1 / 2 gated). Not changed; the claim replaces `next_act` at the end, so the end + aidel think is right either way.
- A freeze that ends while a longer stun (state 21) is on: game.exe refiles at the freeze end + aidel (then the stunned driver stands); d2d keeps one stand end and refiles nothing. Rare (Barbarian / item stun on a frozen monster).
- Events of types 0, 1, 6, 7, 9 and 10 dropped while frozen are not modelled.
- Players are chilled, never frozen, by FUN_0057b230 (player branch); mercs and pets are hirelings or not monsters of this path: only monsters are claimed.
- Not run: independent verifier, det state hashes (need a d2d slot; Glacial Spike scenarios in DETHASH2 would show it), Windows / MSVC build.

## Candidate

Branch `fix/freeze-resume` in the worker host's PR checkout (worktree `fresume`), base upstream main 72132b99 (= `gh api repos/psi29a/d2decomped/commits/main` on 2026-11-01). Author and committer the numeric-ID noreply form, bodies two lines plus `Agent: Claude Code`, no Claude trailer.

| commit | what | files |
|---|---|---|
| c6cc4f12 | claim: a frozen monster thinks again at the freeze end + aidel | ai.cpp (+13), ai.hpp (+1), fight/strike.cpp (1 line), docs/research/re/monster-ai.md (+5) |
| efacc7f4 | tooling: `tools/emu/freeze_resume.py`, `tests/game/test_freeze_resume.cpp` | 2 new files |

## CI-class checks (head efacc7f4, worker host, offline)

- pr_guard: upstream/main..head 0 FAIL / 0 WARN; each commit (c6cc4f12, efacc7f4) 0 / 0. (A first run warned that FUN_0057b170 / FUN_005a7ef0 are not function entries in upstream's game-functions.tsv: they are a callback and a table handler, now written as addresses.)
- presubmit --quick: rc 0, 0 FAIL, 2 WARN: `siblings: unit kinds` (answered in the draft: players and hirelings are chilled, never frozen, by FUN_0057b230) and `build: skipped (--quick)` (the build below). The data-gating WARN of the first run is fixed (the mode skips without the 1.14d patch).
- Tooling unittests (test_tooling, test_gaps, test_pr_guard, test_presubmit, test_tidy_pr_select): pass.
- gcc `-Wshadow -Wunused -Wsign-compare` Release build of test_game, test_rules, test_combat, test_protocol, test_monsters, test_freeze_resume: 0 warnings on added lines.
- Tests (MPQ dir and patch): test_monsters, test_rules, test_combat, test_protocol, test_game rc 0.
- clang-tidy via tidy-pr-select (136 TUs): 0 findings in ai.cpp, ai.hpp, strike.cpp, test_freeze_resume.cpp (the rest are upstream's own `x` / `y` bindings, also on upstream main); direct run on the three touched TUs: only outdoor.hpp's pre-existing bindings.
- ruff with presubmit's rules (F, E9, B, BLE001, PLW1510, S602, S604) plus `--select B905`: clean. (ruff's default rule set also reports SIM115 on three `open()` calls, the same pattern as upstream's freeze_rules.py; not in the upstream rules.)
- Replay: upstream main is still 72132b99 (`gh api repos/psi29a/d2decomped/commits/main`, 2026-11-01): both commits apply as they are.
- Size: claim 4 files +20 / -1; tooling 2 new files. test_game.cpp untouched (new test file).
- Before / after re-run on the final head (same case and game files): identical counts (12,503 / 12,459 -> 0).

## Dependencies (textual, `git merge-tree`, both orders)

Clean with #155 (head 8c56e96a), #156 (3f418788), #157 (cf85634a), #158 (6b3964fa; it adds `just_hit_until` next to `stun_until`: the new field sits on the other side of `stun_until` so the two insertions do not touch), the local fix/freeze-rules-v2 (409903a6) and fix/armor-hit (d5bd19c1). Semantically the two freeze candidates set `stun_until` with `max` as before, so their freezes get the resume without change; the merged v2 tree was run (above).

## Bot implication

Every freeze on d2d (Glacial Spike, Frozen Armor, the merc's Frozen Armor; Freezing Arrow and Blades of Ice are chills on d2d anyway) lets the monster act aidel frames earlier than game.exe: 15 frames (0.6 s) per freeze in Normal, 13..15 in Nightmare, 13 for most rows in Hell (bosses 6..12). Direction and size of the three cold deviations together, for a Sorceress:

| deviation | d2d vs game.exe | size (estimate) |
|---|---|---|
| this resume (all freezes) | **pessimistic**: monsters act sooner after a freeze | Glacial Spike's freeze is ln34 = 50 + 3 (lvl - 1) frames (D, x Blizzard synergy); at slvl 1 in Normal the monster is out 50 frames on d2d vs 65 in game.exe (game.exe 30 % longer), at slvl 20 107 vs 122 (14 %) |
| freeze divisor and specials (freeze-rules C5 / C3, unmerged) | **optimistic**: d2d freezes 2x / 4x too long in Nightmare / Hell and freezes champions / uniques / bosses at all | Nightmare: d2d 50 vs game.exe 25 + 14 = 39 frames at slvl 1; Hell 50 vs 12 + 13 = 25; specials: any freeze vs none. Larger than the resume outside Normal |
| Frozen Armor freeze-everything (fix/armor-hit, unmerged) | **optimistic**: every damaging striker, bosses and cold immunes included, stands calc1 frames | 30+ frames per damaging hit on Andariel and champion packs, where game.exe at most chills |
| hit-once over-hit (#158, open) | **optimistic**: NextHit missiles strike again a frame early, more hits per cast | per cast; see docs/research/pr-nexthit-proof.md |

Net: in Normal with Glacial Spike alone, d2d kill and death numbers are pessimistic by the resume (the monster is free about 15 frames sooner per freeze); outside Normal and with Frozen Armor up the optimistic deviations dominate. Mark cold-skill soak numbers "d2d, D-FREEZERESUME" (with D-ARMORHIT / D-FREEZERULES where they apply) until the fixes run in the lab binary.

## Fork proposal (not done)

Under the standing fork authorisation: carry the claim commit privately as `lab-freezeresume` on the private fork after the lab-nexthit / lab-armorhit commits (it merges clean with both candidates), build `bin/d2d.lab-freezeresume`, re-run `test_freeze_resume` against the oracle on that build (expect 0 frozen-case mismatches), and use it for cold-skill soaks. Proposed only; the coordinator or the user decides.

## Commands (worker host)

```sh
cd <PR checkout>/fresume                                   # fix/freeze-resume, head efacc7f4
# build test_freeze_resume as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
cd tools/emu
uv run python freeze_resume.py gen 30000 0xF2E50001 > c1.tsv
D2_MPQ_DIR=<dir with bin/game.exe> uv run python freeze_resume.py game c1.tsv > g1.tsv
D2_MPQ_DIR=<MPQ dir> D2_PATCH_INSTALLER=<patch> <build>/tests/test_freeze_resume --freeze-resume c1.tsv > d1.tsv
uv run python freeze_resume.py compare c1.tsv g1.tsv d1.tsv   # "frozen in game.exe ... 0 differ"; before (72132b99 + the 2 files): 12,503
```

A verifier should write their own generator and harness, re-derive 0x57b170 / FUN_005a7f80 / FUN_005416b0 and run the real queue (point 1 above) before reading this doc.

## Shared-file lines for the merger

- docs/research/pr-bundle-queue.md (owned by the open card armorhit, so not edited here), append a row or line: `- FREEZERESUME (2026-11-01, docs/research/pr-freezeresume-proof.md): local fix/freeze-resume (claim c6cc4f12, tooling efacc7f4, on upstream main 72132b99) **BUILT, NOT READY** (independent verifier and det state hashes needed): a frozen monster thinks again at the freeze end + MonStats aidel (0x57b170; FUN_005a7f80 drops thinks while frozen); own oracle + harness, 2 x 37,776 cases: frozen (plain + minion) 12,503 / 12,459 -> 0; specials are freeze-rules C3 (0 when merged with fix/freeze-rules-v2); pr_guard 0/0 per commit, presubmit 0 FAIL / 2 WARN answered, tests green; merges clean both orders with #155..#158, freeze-rules-v2, armor-hit`
- docs/BACKLOG.md: `- FREEZERESUME: independent verifier + det state hashes for fix/freeze-resume (freeze end + aidel resume, docs/research/pr-freezeresume-proof.md; verifier: run the real event queue FUN_005416b0 / FUN_005414d0 for one freeze); proposed lab-freezeresume fork carry`
- docs/INDEX.md, research: `- docs/research/pr-freezeresume-proof.md / pr-freezeresume-draft.md: freeze end + aidel resume (0x57b170, FUN_005a7f80) local upstream candidate and differential (FREEZERESUME)`
- CHANGELOG.md (unreleased, docs): `- docs: FREEZERESUME differential: d2d lets a frozen monster act at the freeze end, game.exe at the end + aidel; local upstream candidate fix/freeze-resume brings the frozen cases from 12,503 to 0 (docs/research/pr-freezeresume-proof.md).`
- GLOSSARY.md: none missing (aidel is a MonStats column name).
