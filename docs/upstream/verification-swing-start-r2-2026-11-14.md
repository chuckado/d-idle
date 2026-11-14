# Independent verification of the swing-start PR r2 (2026-11-14)

> status: current, 2026-11-14 (TASK-418, agent prverify6). Verdict **NEEDS-WORK, text only**: the code, the test and the evidence PASS; two wording edits (the oracle commit's message, one stale sentence in the draft) before posting. No third verifier pass is needed: the coordinator checks the two edits and that the tree is unchanged. Nothing posted, nothing pushed. Consumer: TASK-417 (prswing2); the user's go is still needed for any post.

Verifier: d2-reviewer (Opus), singly, second pass. First pass: [verification-swing-start-2026-11-14.md](verification-swing-start-2026-11-14.md) (r1, NEEDS-WORK, five edits). Candidate: `fid/swing-start-r2` on the private fork remote, head `8626e244`, two commits on upstream main `f9b8cfed` (oracle `4a18fa9e`, claim `8626e244`); upstream main is still `f9b8cfed` (`git ls-remote`, 2026-11-14), so no rebase is needed. Draft: [drafts/swing-start.md](drafts/swing-start.md) (`e813fdb8`), queue row 1c. Tags: V read, V run (on the worker host), S, U.

Harness (scratch, not for upstream): `~/d2d-lab/pv6/` on the worker host. Clean detached worktrees `wt-pv6-{p,o,c}` at `f9b8cfed`, `4a18fa9e`, `8626e244` (porcelain empty after every run), builds `build-pv6-{p,o,c}` (GCC, Release, `D2D_WERROR=OFF`); r1 from pass one's clean `build-pv5-c` (`ebeb4268`). Own probe `pv6_swing.cpp` (new grid, below), own game side `pv6_oracle.py` (reuses only pass one's emulator glue), `summ6.py`.

## Verdict

**NEEDS-WORK (text only).** All five first-pass edits are in and correct; the held-weapon fix is pinned by a test that fails on r1 and passes on r2; a new 78,172-cell grid through d2d's real attack path gives **0 differences on r2** in every comparable cell (parent: 18,064 of 51,056 reachable cells differ). The class-restricted exclusion of the author's run is legitimate by d2d's own equip rules, and no other difference hides in it. Two text edits remain (A, B); C to E are optional.

## Required edits

- **A. MEDIUM (oracle commit message, `4a18fa9e`).** Its body still says "12,460 cells (7 classes, 9 weapons, 5 modes, 10 IAS, 7 WSM): 815 differ on the swing's end, 2,101 on the strike tick". That was r1's grid. This commit now skips claws for every class but the Assassin, and its own grid gives `11270 cells  2011 differ (end 815, strike 2010)` (V run: `test_swing_start --swing-start` from `build-pv6-o`, `swing_start.py compare`). Reword it to: `11,270 cells (7 classes, 9 weapons, claws for the Assassin only, 5 modes, 10 IAS, 7 WSM): 815 differ on the swing's end, 2,010 on the strike tick.` The claim commit's message already matches this. The amend changes both hashes; check that `git diff 8626e244 <new head>` is empty and that the subjects and trailers are unchanged.
- **B. MEDIUM (draft, "What changes", the `Fighter::attack_start` sentence).** It still says "the start frame of the weapon in hand (worn or not: U that game.exe's held weapon ignores the strength and level requirements; the COF record follows the worn weapon the same way)". That is r1's comment, and r2's code, combat.md and the draft's own table say the opposite. Replace it with: "the start frame of the usable weapon in hand (`player_fighter`'s `weapon`: slot 4, else slot 5, broken and requirement-short ones skipped)".

## Optional (not blocking)

- C. LOW (draft, the 19,800-cell paragraph). "r2: 0 of 19,170. The other 630 cells are gear no hero of that class can wear" mixes two counts. In the author's `cmp-r2s.txt` (V run, re-counted here), 660 cells hold class-restricted gear and 630 of them differ. The other 19,140 cells, all reachable, give 0 differences. Say "0 of the 19,140 reachable cells; 630 of the 660 class-restricted cells differ". The cause is also more precise than "d2d does not count class-restricted gear as worn": d2d's `meets_requirements` refuses a class item (`rules.hpp:1199`), so `can_wear` will not equip it and `settle_requirements` flags a worn one 0x4000. Pass one's oracle took the held weapon from the gear tag and not from that flag.
- D. LOW (draft, Limits): "is I" is not one of the V/S/U tags; write U.
- E. LOW (draft, upstream state): a new upstream PR, #529 (another contributor, Lightning Fury's burst), touches `tools/emu/README.md` and `fight/missiles.cpp`. `git merge-tree` with r2 is clean (V run), and their README rows are in different places. Name it in the post note.
- (Carried from pass one, still optional) The "about 12 percent of her melee damage per second" line is arithmetic (S). It stays in the internal section only.

## The five first-pass edits on r2

| Edit | r2 | Check |
|---|---|---|
| 1 HIGH held weapon | `player_fighter` reuses its own `weapon` (`strike.cpp:109`). That variable comes from the loop over `gear_active` items, and `gear_active` requires `rules::item_usable` (0x4100 clear, `character.hpp:112`). The choice is kind 2, the right arm first, else the left: game.exe's slot 4 then slot 5 rule. A weapon with no base record leaves `attack_start` at 0, as r1 did. The "worn or not" comment is gone. r1 to r2 in code: only this block (V read, `git diff ebeb4268 8626e244 -- components`). | V run: r2's `test_swing_start` (no argument) prints `34 cells match game.exe`. The same test compiled against r1's library (`-DNDEBUG`, so its first assert does not stop it) prints `FAIL: unusable 'ssd': swing 18 strike 10, game.exe's 19 and 11` and `FAIL: unusable 'gsd': swing 22 strike 12, game.exe's 23 and 13`; with asserts on, it aborts on the unknown-key assert (edit 5b). The parent lacks `attack_start_frame`, so the test does not compile there. |
| 2 counts | combat.md: 815 of 840 cells with a start frame; 1,068 of 8,120 cells of the other classes; 1,176 of the 10,430 cells without a start frame; 11,270 cells. | V run: the claim's grid on the oracle commit gives `11270 cells 2011 differ (end 815, strike 2010)`, and 1,068 + 108 = 1,176 matches the draft's table. On r2: `11270 cells  same`. |
| 3 Fend | combat.md and the draft list Fend (A1) among the skills that take the start frame; Jab and Impale (SQ) keep their length. | V run (probe): an Amazon's Fend with an axe goes from 13 to 11 ticks, start 2. |
| 4 throws | The draft's collateral: TH strikes on `attack_action_tick` with start 0, earlier in 56 of 360 cells, length unchanged; combat.md says the same. | Pass one's T lines (V run), not repeated. In the new probe the throw skills reach `start_swing` only as d2d's plain-attack fallback (`swing_skill` 0), so those 480 cells are plain A1 swings and match game.exe's A1 on r2 (0 differ). That is not the missile path. |
| 5 wording | `FUN_006235a0` is described as a class test (code comment, combat.md, docstring). `attack_start_frame` returns row 0 for an unknown key (`zzz` gives Amazon 1, Sorceress 1, Paladin 0, asserted). The end's floor of 2 is written in combat.hpp, combat.md, the docstring and the draft. Subject 63 characters. | V read. |

`swing_start.py table` on r2 = pass one's raw dump, row for row (V run).

## New differential (V run)

Probe `pv6_swing.cpp`, not used by the author or in pass one. It goes through d2d's real attack path (`start_swing` -> `set_pmode` -> `player_modes` ticked every 40 ms -> `strike` and the mode's end), against a Fallen with 500,000 life in the Blood Moor. The grid:
- **every** weapons.txt base (all tiers);
- two heroes per class: S (level 99, Str/Dex 255) and F (fresh level-1 stats, so heavy weapons are requirement-short through d2d's own `settle_requirements`: 31,092 such cells, Amazon 6,836 and Sorceress 732);
- every fifth weapon also broken (0x100);
- shields and broken shields in the left hand, a quiver with bows and crossbows;
- the Barbarian and Assassin dual-wielding with a broken right hand, or holding a weapon in the left hand only;
- the weapon's **own** WSM, not a forced one;
- IAS from {3, 17, 33, 48, 66, 85, 110, 140, 190};
- every class skill whose `anim` is A1, A2, S1 to S4, KK or TH, plus the plain attack.

The game side, per cell, applies game.exe's held-weapon rule to the flags d2d actually holds after entering (slot 4 if a weapons.txt item without 0x100/0x4000, else slot 5, else none). It then runs `FUN_00623b10` and `FUN_005539b0` natively on the record's real AnimData flags at d2d's rate. Both trees get identical inputs: the 17 input fields of all 78,172 `R` lines match between the parent and r2.

| Cells (78,172 printed; 3,520 have no record and are skipped) | n | parent differ | r2 differ |
|---|---|---|---|
| reachable (wearable gear, a usable skill, single swing) | 51,056 | **18,064** | **0** |
| of which Amazon / Sorceress | 11,152 | 9,216 | 0 |
| of which the other five classes | 39,904 | 8,848 | 0 |
| of which the right hand broken or requirement-short | 28,734 | 10,326 | 0 |
| class-restricted gear (d2d's `can_wear` false: am*, ob*, claws on the wrong class) | 7,280 | 2,134 | 0 |
| plain-attack fallback of throw skills (TH) | 480 | 480 | 0 |
| wereform skills in human form, passives (unreachable in play) | 10,560 | 5,916 | 0 |
| chains: Zeal, Fury, Dragon Talon (the probe times the whole chain against one swing) | 5,276 | 5,000 | 4,976 |

The chain rows are a probe artefact on both trees (for example, Zeal ends at 30 ticks in d2d against one 10-tick swing in game.exe). Zeal's per-hit gaps are covered by `test_zeal_chain` (passes on both trees, V run) and by the author's `zg_grid.py` (draft, not re-run here). The r1 run of this grid is in progress at the time of writing; see the addendum.

**The author's exclusion (V run, `pv5/cmp-r2s.txt` re-counted).** Every difference except Jab is in an Amazon or Sorceress cell holding gear d2d's `can_wear` refuses that class: Amazon with ob1/wrb/skr, Sorceress with am1/am3/am5/wrb/skr. That is 630 cells, all within those 660 class-restricted cells; the other 19,140 cells have 0 differences. The new probe's `W` lines list `can_wear` for every weapon and class. The refused set is exactly: the Amazon's am* items for every other class, the Sorceress's orbs (ob*) for every other class, and claws (katars) for every class but the Assassin. With the flags read the way the new oracle reads them, those cells also give 0 differences on r2. Whether game.exe's `FUN_0055dbc0` flags such an item is S (d2d's model). It does not matter here, since neither side can equip one.

## Regressions, builds, tests (V run)

- Code paths: unchanged from pass one's review except the held-weapon block. `set_pmode` callers, monsters, mercs and pets (`npc_timing`, start 0), bows and crossbows (row 0 for those weapon classes), Whirlwind, sequences and A2 were re-read; the new grid covers bow skills (A1, start 0), Smite, kicks, Assassin traps (S2), Barbarian and Assassin dual wield, and Fend. All give r2 = game.exe.
- Full `ninja` of every target at each commit alone, Linux GCC: `f9b8cfed`, `4a18fa9e`, `8626e244` all rc 0. macOS and Windows were checked by reading only (nothing platform-specific; `<string_view>` is included).
- `ctest -j6` with the data and the 1.14d installer: parent 299 tests, r2 300 (`test_swing_start` passes). Both fail the same four (`test_mpq`, `smoke_d2d`, `quests_act1`, `quests_act4`: no video MPQs on this box). `test_zeal_chain` passes on both.
- `pr_guard.py f9b8cfed HEAD`: 0 FAIL, 0 WARN; `--sorted`: 0 FAIL. `presubmit.py --quick --base f9b8cfed`: 0 FAIL and 5 WARN (two siblings WARNs, two error-path WARNs on `swing_start.py`, build skipped because of `--quick`), all answered in the draft; ruff ok (from a venv). Full presubmit (strict build, clang-tidy) was not run; the full build and ctest above stand in for its build and test parts, and clang-tidy is U.
- Hygiene: one topic, 10 files, 478 insertions and 8 deletions (the code is 6 files). The oracle commit comes first and builds alone. Subjects are 69 and 63 characters, with 2 and 3 body lines, then a blank line and `Agent: Claude Code`. Author and committer are the user's GitHub noreply identity; there is no Co-Authored-By. The diff has no lab hosts, paths or names (grep, 0 hits). The draft claims no kill-rate effect ("no kill-rate change is measurable", internal section), says the bot A/B was not repeated for r2, and keeps the client drawing of skipped frames as U.

## Upstream state (V run, read-only `gh`, 2026-11-14)

Upstream main is `f9b8cfed`, unchanged. There are 4 open upstream PRs (#529, #456, #433, #236), none of them ours, and our open PR count is 0. #529 (opened today) and #456 touch files this branch touches: `git merge-tree` is clean against #529, and pass one found it clean against #456. No open issue or PR mentions `set_pmode`, swing or attack speed. CodeRabbit last commented today on others' PRs; with 0 open PRs of ours, the post stays within the limit of at most 3.

## Not done

No live run, no sandbox run, nothing posted or pushed. Not re-run here: the throw (missile-path) count, `zg_grid.py`, and full presubmit with clang-tidy.
