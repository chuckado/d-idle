# Independent verifier: fix/nexthit (D-NEXTHIT, NextHit justhit) (VERIFYNEXTHIT, card verifynexthit, 2026-11-01)

> status: verifier record for the local upstream candidate `fix/nexthit` (claim 8abab0ed, tooling 7901fda2, base upstream main 72132b99) on the worker host. Author's proof and draft: `docs/research/pr-nexthit-proof.md` / `pr-nexthit-draft.md` (commit a9e92a2). Offline and read-mostly; nothing pushed or posted; the author's branch and worktrees untouched. Verdict at the end.

Tags: **V read** disassembled with upstream's `tools/emu/disasm.py` (my own dumps); **V run** game.exe's code run under upstream's emulator in my own oracle; **S** d2d source; **U** not traced.

Order of work: the claim commit and the disassembly first (the trace below), then my own oracle and d2d harness, then the author's proof and draft (they were read only after the trace and the first differential).

## 1. Where justhit ends: one frame later than the patch (V read, V run)

Set (V read, FUN_005adb00, called from FUN_005adf10 at 0x5ae07a / 0x5ae0ee): if the row's NextHit (+0x188, offset V by `txt_fields.py`) is set, a stat list is made by FUN_006251f0 (pool, flags 2, expiry frame = game +0xa8 + NextDelay (+0x189), unit type, unit id; the list's +0x18 is the expiry), FUN_005417d0 queues a timer of event type 0xc on the struck unit for that same frame, FUN_006252d0 puts state 0x56 in the list, FUN_00625ce0 sets the list's callback (+0x38) to 0x5adaf0, FUN_00626e10 attaches it to the unit, FUN_00639db0(unit, 0x56, 1) sets the bit.

Consumed (V read): FUN_005a8730 (behind the filter 0x5a8850; NextHit row: unit in 0x56 -> no) and FUN_005adf10 0x5adf84 (the same test again, then FUN_0064a470 last unit, FUN_0064a400). FUN_005ae1f0 walks the subtiles the missile entered this frame; FUN_00641cb0 returns the first unit that passes the filter; found: FUN_005adf10 and the missile's frame is over (0x5ae3a4..0x5ae3dd), so one unit a frame (V read). The 0x56 test in FUN_005a8730 does not depend on LastCollide; the last-unit test (FUN_0064a470) does (row +4 bit 0, 0x6ce268 = 1, V read).

Expired (V read): FUN_005417d0 -> FUN_005416b0 files the timer by `0x6e0b9c[unit type]`: player 0, monster 1, object 3, missile 2, item 4 (static table, V). The game frame FUN_0052d870 adds 1 to game +0xa8, then FUN_005414d0 runs that frame's timers group by group: **group 2 first (FUN_00541240, whose default handler is the missile dispatcher FUN_005adcc0 -> FUN_005adbb0 -> the row's pSrvDoFunc, i.e. every missile's move and collision), then group 0 (players, FUN_00540f60), group 1 (monsters, FUN_00541060 -> FUN_005a7f80), 3, 4**. A monster's type-0xc timer goes through FUN_005a7f80 (jump table: type 12 straight to 0x6e2490[12] = FUN_005a7ef0) -> FUN_00627460(unit, frame): every timed list (flag 2) with expiry <= frame -> FUN_00626c00 -> FUN_006269f0 -> the callback 0x5adaf0 -> FUN_00639db0(unit, 0x56, 0).

So a unit struck in frame f keeps 0x56 through the missile pass of frame f + NextDelay (its own timer runs after the missiles of that frame) and can be struck again from frame f + NextDelay + 1. For the novas (NextDelay 4): hit at f, blocked f+1..f+4, next hit at f+5 at the earliest. The patch (`now_ms < just_hit_until`, until = now + NextDelay x 40) and the author's oracle ("timer at the start of its frame") allow f+4. **The window is one frame short** (the author's limit 3 named both readings; the f+5 one is the game.exe one). The same holds for players and mercs as targets (group 0 also runs after group 2). V run: my oracle runs the expiry chain for real (FUN_005a7ef0 -> FUN_00627460 -> FUN_00626c00 -> FUN_006269f0 -> 0x5adaf0 -> FUN_00639db0 / FUN_00625a70 on a StatListEx at unit +0x5c), and a synthetic case (one monster, a fresh NextHit missile every frame) gives hits at frames 1, 6, 11 with the traced order and 1, 5, 9 with the patch's.

Can the window end before the next ring missile arrives? Yes, often: Nova and Frost Nova live 13 / 14 frames at 24 a frame, and a monster stays under the ring for several frames when it is large, close, moving into the ring, or under a second cast; then the frame f+4 contact is a hit in d2d (and in the author's oracle) and a miss in game.exe. Counts in section 2.

Other rows: 48 Missiles.txt rows have NextHit (V, the 1.14d Missiles.txt read through `txt_rows`), NextDelay 3 to 25: besides the novas, Chain Lightning, Multiple Shot / Strafe arrows and bolts (Pierce, CollideKill), Battle Cry / War Cry / Battle Orders / Battle Command, Fissure cracks, Twister / Tornado, Volcano, Shock Field, Blade Sentinel's stand-in, Dragon Breath, Claws of Thunder, the monster and boss novas. The patch changes every one of them that goes through `skill_missile_hits`; only the novas are measured.

## 2. Own differential (own generator, own oracle)

- d2d side: `tools/emu/test_verify_nexthit.cpp` (built as `tests/game/test_verify_nexthit.cpp` in throwaway detached worktrees): the real `Fight::fire` / `Fight::world` on an open Blood Moor copy; MonStats rows chosen by a different rule than the author's (first targetable row of each size: rat, fetish1, brute1); quarter-subtile positions in a disc of 17 subtiles, 1..6 monsters, skill level 1..30, Nova or Frost Nova, half the cases with every monster MOVING (0.3 to 1 subtile a frame on each axis, moved before each frame), a fifth with a SECOND cast of the same skill 1..8 frames after the first. Generator: an LCG (`x * 1664525 + 1013904223`, masked to 32 bits, top 24 bits used), seeds 0x5101, 0x6202, 0x7303, 0x8404 (strtoul, masked to 32 bits); 25,000 random cases a seed plus 624 edge cases on the first seed (one monster of each size at 0..17 subtiles in 8 headings). It prints the contacts in two orders: monster index order (I, as d2d's frame loop and the author's oracle) and the order of the subtiles crossed (P, as FUN_005ae1f0 walks them).
- game side: `tools/emu/verify_nexthit.py`, my own: the filter, FUN_005adf10, FUN_005adb00 and the stat-list functions run for real on a StatListEx per unit (hooks in its docstring: allocator, owner, hostility, stat getters, the States.txt record, the timer insert (recorded), hit event, damage roll (counted), apply); the expiry chain FUN_005a7ef0 -> ... run for real at the timer's frame, either after the frame's missiles (**post**, the traced order) or before them (**pre**, the patch's / author's order).
- compare: `tools/emu/verify_nexthit_compare.py` (per-frame hit lines and per-monster totals per case).

100,624 cases (4 x 25,000 + 624), per build; each build replayed on its own contacts.

| build | oracle order | cases differing | monster-casts with a different total (of 197,459 touched) | d2d hits / touched | game.exe |
|---|---|---|---|---|---|
| upstream main 72132b99 | post (traced) | 82,876 | 173,783 | 10.28 | 1.21 |
| upstream main 72132b99 | pre | 82,800 | 173,196 | 10.28 | 1.30 |
| fix/nexthit head 7901fda2 | pre (author's order) | **0** | 0 | 1.30 | 1.30 |
| fix/nexthit head 7901fda2 | post (traced) | **25,252** | **17,115** | 1.30 | 1.21 |
| head + `<=` (the `fix` variant, one character in missiles.cpp) | post (traced) | **0** | 0 | 1.21 | 1.21 |
| head + `<=` | pre | 25,252 | 17,115 | 1.21 | 1.30 |

The `fix` variant is 0 on every slice under the traced order (Nova, Frost Nova, standing, moving, one and two casts; 6 cases in the subtile order, as for the head under its own order). So the whole residual of the head is the end frame, and `<=` removes it.

Head vs the traced order, by slice (cases differing / cases; d2d vs game.exe hits per touched monster):

| slice | differing | d2d / touched | game.exe |
|---|---|---|---|
| Nova | 12,069 / 50,077 | 1.29 | 1.20 |
| Frost Nova | 13,183 / 50,547 | 1.30 | 1.22 |
| standing | 6,065 / 50,302 | 1.16 | 1.11 |
| moving | 19,187 / 50,322 | 1.51 | 1.36 |
| one cast | 16,105 / 80,548 | 1.18 | 1.11 |
| two casts | 9,147 / 20,076 | 1.77 | 1.59 |

Under the author's order my harness reproduces the author's result (0 differing on the head, moving monsters and second casts included), so the patch matches its own model everywhere; the residual is the end frame alone.

Contact order: replaying the subtile (P) order instead of the index (I) order changes 6 cases of 100,624 on the head (pre): FUN_00641cb0's order matters but rarely (the units' order inside one subtile is still U).

Contacts base vs head: identical in all 100,624 cases of my harness (I lines). The author's "1,645 contact lines differ" is not reproduced; its cause stays untraced (U) and it does not affect the comparison (each build is replayed on its own contacts).

## 3. Isolation (each commit built alone)

Four throwaway builds of my harness: base (72132b99), tool (72132b99 + cherry-pick of 7901fda2 alone), claim (8abab0ed alone), head (7901fda2). md5 of every case, contact and hit line (K / I / P / D / S, log lines dropped), per seed:

| seed | base | tool alone | claim alone | head |
|---|---|---|---|---|
| 0x5101 | c404e16dfa6a | c404e16dfa6a | cf2cb90b1bff | cf2cb90b1bff |
| 0x6202 | 0937a86a0772 | 0937a86a0772 | d134e972902d | d134e972902d |
| 0x7303 | 04ad986c0bb3 | 04ad986c0bb3 | 85a6f8097457 | 85a6f8097457 |
| 0x8404 | 373d3913120a | 373d3913120a | d9aba07f0b53 | d9aba07f0b53 |

The tooling commit changes nothing in d2d's behaviour (tool = base, byte for byte); the claim commit alone carries the whole change (claim = head, and the oracle outputs of claim and head are identical). The `fix` variant differs from the head only in the cases section 2 counts.

## 4. CI-class checks (head 7901fda2, my runs on the worker host)

- pr_guard: 0 FAIL / 0 WARN on the range against upstream/main and on each commit (8abab0ed, 7901fda2).
- presubmit --quick: rc 0, 0 FAIL, **4 WARN** (siblings unit kinds; siblings hit paths; docs-sync; build skipped). My run shows no fifth WARN; the author's "fifth WARN not read" is not reproduced on the committed head (it was probably from an amended, uncommitted state). The body must answer the three real ones: (a) unit kinds: game.exe puts 0x56 on any unit (players and mercs too, timer group 0 also after the missiles); d2d's monster-missile path does not read NextHit, so monster novas (Mephisto's, Baal's, Andariel-class poison novas, teeth) still multi-hit the player: say so; (b) hit paths: only missiles carry NextHit: fine; (c) docs-sync: re/skills.md:928 ("strike a unit again NextDelay frames on") is wrong in both readings; correct it in the PR.
- ruff with upstream's rule set (presubmit: F, E9, B, BLE001, PLW1510, S602, S604) and `--select B905`: clean. Ruff 0.17.3 with its own defaults flags I001 (import order) and SIM115 (open without a context manager) in next_hit.py: not upstream CI, a one-line `--fix` if wanted.
- gcc -Wshadow -Wunused -Wsign-compare build of test_game, test_rules, test_combat, test_protocol, test_monsters, test_next_hit: 0 warnings in the touched files.
- test_rules, test_combat, test_protocol, test_game, test_monsters: rc 0 (Release, -Wno-error=maybe-uninitialized -Wno-error=array-bounds, MPQ dir set).
- clang-tidy (repo .clang-tidy, tidy-pr-select: 136 TUs that include a touched file): 0 findings in the touched files (29 elsewhere, untouched files).
- Replay: upstream main is still 72132b99 (`gh api repos/psi29a/d2decomped/commits/main`, 2026-11-01); cherry-pick of both commits clean.
- Shared files with our open PRs: **not none** (the brief expected none): #155 (fix/static-field, head 8c56e96a) touches `gamedata.hpp` and `gamedata_load.cpp`, #156 (fix/walk-velocity, head 3f418788) touches `fight/tick.cpp`. Textual merges of the head with each of #152, #153, #155, #156, #157 and with #155 + #156 together are clean (V, `git merge` in a throwaway worktree). Semantic overlap: #155's chill and freeze rules act per Frost Nova hit, so their effect per cast falls with this patch (expected, see section 5).
- Size: 7 files, 2 commits, +20 / -8 in the claim.

## 5. Det state hashes (base vs patched, one hidden d2d slot)

Builds: the DETHASH2 lab tree (upstream 72132b99 + lab commits, 5e0a51bc) as base (its binary, sha256 70c7b3694173, the one docs/research/det-hashes-2026-11-01b.md used) and the same tree + cherry-pick of the claim 8abab0ed (`nexthit`, sha256 f01f2fe206e9; the tooling commit adds no d2d code, section 3). One hidden slot (`tools/d2dslot.sh run`, slot 0, under `agentrun.sh verifynexthit`), capacity checked first (idle + nice 83 %), the DETHASH2 driver unchanged: `det_scenarios_run.py run BINDIR OUTDIR <14 scenarios> base,nexthit`, seed 3, lockstep, a fresh d2d per run, 28 runs.

The base hashes reproduce DETHASH2's base column in all 14 scenarios (mon and state). Results (V):

| scenario | nexthit vs base | class |
|---|---|---|
| passive, fight, cold, static, cold2 | same | no NextHit missile |
| hit_minion | same | the Frost Nova reaches no minion (as in DETHASH2) |
| merc, pet, follow | same | no NextHit missile |
| hit_plain | differs, only the Frost Nova target #5 and the next target #6 | EXPECTED: Frost Nova point blank hits once: hp 106/133 -> 131/133, last chill 40 ms earlier (14000 -> 13960); then #6's spawn roll comes from a shifted RNG stream (max life 119 -> 93); Glacial Spike's chill / stun deadlines unchanged |
| hit_champion | same two targets | EXPECTED: 303/312 -> 310/312 (chill same); #6's Glacial Spike damage roll shifted (20 -> 24 lost), deadlines unchanged |
| hit_unique | same two targets | EXPECTED: 494/520 -> 518/520, chill 14000 -> 13960; #6 is another unique (spawn roll shifted: Spine Shank 560 -> Steel Shank the Unclean 708), deadlines unchanged |
| hit_boss (Blood Raven) | same two targets | EXPECTED: 109/113 -> 112/113, chill 10160 -> 10120; #6's damage 10 -> 9 lost, deadlines unchanged |
| hit_immune (Hell zombie) | only target #10 | EXPECTED (RNG): the cold-immune #9 loses nothing either way (fewer hit draws though), so #10's spawn roll shifts (max life 4495 -> 5918); chill / stun deadlines unchanged |

DET-CHANGES-EXPLAINED: 9 of 14 scenarios identical, 5 differ only by the claimed Frost Nova hit count (damage, the last chill's frame) and the RNG stream shift it causes for the next spawned target; 0 unexplained. Coverage: no Nova and no moving or multi-target Frost Nova in the scenario set; no other NextHit row (Chain Lightning, Multiple Shot, the war cries) is cast; the `fix` variant (`<=`) was not run here (one more binary, the same harness).

## 6. Code review notes (S unless tagged)

1. HIGH: the end frame (section 1): `now_ms < struck_unit.just_hit_until` must be `<=` (or until = now + (NextDelay + 1) x 40). Built and run as the `fix` variant (one character), see section 2. The `// ponytail:` comment ("justhit ends at the start of its last frame (the timer's place in game.exe's frame is not traced)") becomes "justhit ends after the missiles of frame f + NextDelay (FUN_005414d0 runs the missile timers, group 2, before the monsters', group 1)".
2. MEDIUM: scope. The change applies to all 48 NextHit rows that reach `skill_missile_hits` (Chain Lightning, Multiple Shot / Strafe, the war cries, Fissure, Twister / Tornado, Volcano, Shock Field, Blade Sentinel, Dragon Breath, Claws of Thunder ...), only the novas are measured. For Multiple Shot and Strafe this is a large behaviour change (one arrow of a volley per monster per 5 frames). It is the game.exe rule (the filter has no row exception, V read) but the body must list these rows as "same rule, not compared".
3. LOW: `Missile::last_hit` is a monster INDEX (as `hit_at` was): if the monster vector is compacted while a missile flies, it points at another monster. Pre-existing pattern; a unit id would be safer.
4. LOW: `just_hit_until` lives on `Monster` only; players, mercs and pets as targets have no justhit (see the presubmit answer above).
5. LOW: the frame-stop test in tick.cpp (`missile.last_hit_ms == now_ms`) relies on `skill_missile_hits` stamping `last_hit_ms` only on a real hit: true today (the stamp follows both early returns); a comment saying so would help.

## 7. Verdict

- **C1a (a NextHit hit puts justhit on the unit, every NextHit missile of any cast passes it by): HOLDS WITH LIMITS.** The mechanism is right (V read, V run). The window is one frame short: game.exe blocks frames f+1..f+NextDelay and allows f+NextDelay+1 (FUN_005414d0 runs the monsters' timers after the missiles); the patch allows f+NextDelay. On my 100,624 cases the head differs from game.exe in **25,252 cases / 17,115 monster-casts** (one cast, standing: 6,065 of 50,302; moving 19,187 of 50,322; two casts 9,147 of 20,076), d2d 1.30 vs game.exe 1.21 hits per touched monster; upstream main differs in 82,876 (10.28 vs 1.21). With `<=` (one character) the residual is **0 of 100,624** on every slice.
- **C1b (a LastCollide row skips its last unit): HOLDS** (V read: FUN_0064a470 compares only when row +4 bit 0 is set, the missile data's one slot; run inside the oracle; no case where it alone decides: the novas' 0x56 test always fires first).
- **C1c (one unit a frame): HOLDS WITH LIMITS** (V read FUN_005ae1f0; restated, not run, as the author says); within a frame the order of units differs between d2d (monster index) and game.exe (subtiles crossed, then FUN_00641cb0's room order, U): 6 of 100,624 cases.
- Isolation: HOLDS (tool alone = base, claim alone = head, md5 equal on all four seeds).
- CI-class: clean (pr_guard 0/0 per commit, presubmit 0 FAIL / 4 WARN, tests rc 0, gcc and clang-tidy 0 in touched files, ruff clean with upstream rules); shared files with #155 and #156 (merges clean).
- Det hashes: DET-CHANGES-EXPLAINED, 0 unexplained.

**NOT READY** as it stands: the end frame is a correctness bug the differential can see (a quarter of my cases). READY once (1) `now_ms < struck_unit.just_hit_until` becomes `<=` in missiles.cpp and the `// ponytail:` note is replaced by the traced rule; (2) the author's oracle `next_hit.py` runs the expiry after the frame's contacts (it runs it before: the 0-differ was by construction on that point) and the differential is re-run (my `fix` run: 0 of 100,624); (3) the draft edits below. No other code change needed.

Draft edits (pr-nexthit-draft.md):
1. Change: "for NextDelay frames" -> "until the missile pass of frame + NextDelay has run (FUN_005414d0 runs the missile timers, group 2, before the monsters', group 1): struck at f, struck again at f + NextDelay + 1 at the earliest".
2. Limits: drop "the frame justhit ends ... is not traced"; add: moving monsters and second casts are covered by the verifier's run (100,624 cases), standing-only no longer; Nova / Frost Nova only, the other 46 NextHit rows (Chain Lightning, Multiple Shot / Strafe, the war cries, Fissure, Twister / Tornado, Volcano, Shock Field, Blade Sentinel, Dragon Breath, Claws of Thunder) get the same rule, not compared; Chain Lightning hop search and piercing Lightning not claimed; units' order inside one frame (subtile order) not compared (6 of 100,624); players, mercs and pets as targets keep no justhit and monster novas still multi-hit the player (presubmit "unit kinds" answer).
3. Evidence table: replace the counts by the re-run after the fix (author's 100,198 and the verifier's 100,624), and say the oracle runs the stat-list expiry chain (FUN_005a7ef0 -> FUN_00627460 -> 0x5adaf0) for real if that is adopted from verify_nexthit.py.
4. docs-sync answer: correct re/skills.md:928 in the same PR.
5. Optional: `ruff check --fix` for I001 / SIM115 in next_hit.py (ruff 0.17 defaults; not upstream's rules).

## Commands (worker host)

```sh
# throwaway detached worktrees of upstream main 72132b99 (base), 8abab0ed (claim), 7901fda2 (head),
# 72132b99 + cherry-pick 7901fda2 (tool), 7901fda2 with `<=` (fix); each gets tests/game/test_verify_nexthit.cpp
build/tests/test_verify_nexthit --vnh 0x5101 25000 edge > d.txt          # and 0x6202 / 0x7303 / 0x8404 without edge
PYTHONPATH=<upstream>/tools/emu python verify_nexthit.py d.txt I post > g.txt   # I|P order, post|pre timer order
python verify_nexthit_compare.py d-0x5101.txt g-0x5101.txt d-0x6202.txt g-0x6202.txt ...
```

## Shared-file lines for the merger

- docs/BACKLOG.md: `- NEXTHIT: verifier (VERIFYNEXTHIT, docs/research/pr-nexthit-verifier.md): end frame one short (f+5 not f+4, FUN_005414d0 group order); fix '<=' in missiles.cpp, re-run, then the draft edits listed there`
- docs/INDEX.md, research: `- docs/research/pr-nexthit-verifier.md: independent verifier of fix/nexthit (justhit end frame traced, own differential, det hashes)`
- CHANGELOG.md: none.
- GLOSSARY.md: `justhit` (state 0x56) still missing; proposed text: "justhit: game.exe unit state 0x56, put on a unit struck by a NextHit missile until the missile pass of frame + NextDelay has run; while on, every NextHit missile passes the unit by."
