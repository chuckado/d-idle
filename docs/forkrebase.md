# FORKREBASE: the private fork stack rebuilt on upstream main 52362d52 (card forkrebase, 2026-11-01)

> status: current. Branches `lab2-debug-skills`, `lab2-armorhit`, `lab2-freezeresume` of the private d2d fork replace the `lab-*` stack as the base of soaks and drills. The old `lab-*` branches and binaries stay untouched as history. Patches never live in this repo (GPL, docs/decisions.md). Standing user approval 2026-10-31: functionality in the private fork when documented and vetted; never in an upstream pull request without the user's go.

Tags: **V run** measured on a real build in this card; **S** source or git history; **I** inferred; **U** not verified.

## Why

The maintainer merged all six of our pull requests (#152, #153, #155, #156, #157, #158; upstream main head `52362d52`). The old stack (`lab-debug-skills` < `lab-nexthit` < `lab-armorhit` < `lab-freezeresume`) sat on the pre-merge main and conflicted with it in two files, so it could not simply be merged. The new stack is built on `52362d52` with fresh branch names; nothing was force-pushed and no old ref moved. S: [upstreamsync.md](upstreamsync.md) section 2 had the plan.

## The new stack (pushed to the private remote the same day)

| Branch | Tip | Holds |
|---|---|---|
| `lab2-debug-skills` | `7ca4a22ccd096abb94597699e9da6c04d175138d` | the lab patch set (lockstep, dump, click modifiers, belt and cheat verbs, debug verbs for skill drills) on upstream main |
| `lab2-armorhit` | `8c70eb09609d6528661d7654c65bf89eb2d74127` | + the Frozen Armor freeze-rule claim and the `armor_hit.py` / `test_armor_hit` tooling |
| `lab2-freezeresume` | `ee5e06d8ef990200b699e61779c64ded1451c33d` | + the freeze-end resume claim and the `freeze_resume.py` / `test_freeze_resume` tooling |

The base is `52362d52cfa03bd998f195414182169eb7dcf3d6`. There is no `lab2-nexthit`: see the next section.

## What changed against the old stack

1. **The lab patch set commit (old `3a629371`, new `2ad4099d` in the stack) conflicted** with upstream in `components/rules/rules.hpp` and `apps/d2d/town/update.cpp`. Resolution: upstream's belt code wins (`rules::belt_box`, `rules::auto_belt`, the shift flag handed to `item_cursor_command`; upstream #147 and #149). The lab `belt_slot` function and the lab shift-click branch are dropped. Only the lab ctrl-click sell is kept: with the cursor empty, a ctrl-left-click on an inventory item sends the sell command to the open store (FUN_0048ffe0, C->S 0x33). S, git diff of the fork.
2. **One extra commit** (`7ca4a22c`, on `lab2-debug-skills`): the lab `give <code> belt` verb called `belt_slot`; it now calls upstream's `belt_box`, the same rule. Without it the build fails.
3. **The three lab-nexthit commits are redundant and dropped** (old `27a172fc` claim, `98951707` tooling, `2c6c4a1a` the end-of-justhit fix). Upstream merged the same work as #158 in two commits (`b9e7a3d5` claim, `6b3964fa` tooling). S: the `components` hunks of the fork's three commits and of the merged pair are identical apart from line offsets; the docs, `test_next_hit.cpp` and `next_hit.py` files are byte-identical at upstream main and at `2c6c4a1a`. V: `test_next_hit --next-hit 1 100000` on the new build, below. So the stack carries no hit-once code of its own any more.
4. **Other merged PRs**: #152 (vendor prices), #153 (waypoint window), #155 (Static Field and cold length), #156 (walk velocity), #157 (boss stats) arrive with the new base. The fork never carried a patch for any of them (S, upstreamsync.md), so nothing else was dropped. Armor and freeze-resume are not merged upstream yet, so their commits stay; they cherry-picked onto the new base with no conflict.
5. The old `lab-click-modifier` branch was not in the stack (its content was the lab patch set commit) and is untouched.

## Binary and the naming rule

- `bin/d2d.lab2-freezeresume` on the worker host: Release, `D2D_WERROR=OFF` (GCC 16 rejects `ui.cpp` under `-Werror`, as for every lab build), from `lab2-freezeresume` `ee5e06d8`. Build script `build-fr2.sh`, worktree `wt-lab2s`, build dir `build-fr2` under the fork directory; the verifiers' `test_verify_armor_hit.cpp` and `test_verify_freeze_resume.cpp` (from `tools/emu/` on main) sit untracked in `wt-lab2s/tests/game/`. `lab2-armorhit` and `lab2-nexthit` binaries were not built (not needed: the freezeresume binary is a superset).
- The old binaries (`d2d.lab-nexthit2`, `d2d.lab-armorhit`, `d2d.lab-freezeresume`, `d2d.lab-debug-skills`, ...) were not touched.
- **RULE: from now on soaks and drills use `bin/d2d.lab2-freezeresume` and name the binary in their report.** Select it with `D2D_BIN=<fork>/bin/d2d.lab2-freezeresume tools/d2dslot.sh run -n N -- d2drill ...` or `d2drill -bin <fork>/bin/d2d.lab2-freezeresume`. Results measured on an older lab binary stay valid for that binary; do not mix them in one comparison without saying so.

## Vetting (V run, worker host, `lab2-freezeresume` `ee5e06d8`)

Game.exe's side of each differential is the recorded run of the proofs (it does not depend on the d2d build); only the d2d side was re-run. The results equal the ones in [forkfreeze.md](forkfreeze.md) and [forkarmor.md](forkarmor.md) on the old stack, plus upstream's own changes.

| Check | Result |
|---|---|
| Hit-once, `test_next_hit --next-hit 1 100000` (the arguments are count then seed) | 100,198 cases, **0 of 100,198** T lines differ, 0 H lines differ |
| Armor, the author's `armor_hit.py` / `test_armor_hit` (seed 0x41524d48) | **0 of 108,506** differ, calc1 0, RESFAIL 0 |
| Armor, the verifier's harness (`verify_armor_hit.py`, `test_verify_armor_hit --varmor`, four case files of 44,400) | chill and freeze end frames differ in **0** of every file; the known resume column (game.exe resume minus d2d stun end) differs in 14,513, 12,710, 12,745, 12,746 (52,714), the same counts as on the old stack |
| Freeze resume, the author's `freeze_resume.py` / `test_freeze_resume` (37,776 cases per seed) | frozen in game.exe: **0 of 12,548** (seed 1) and **0 of 12,516** (seed 2) differ, every aidel class 0 |
| Freeze resume, the verifier's harness (`verify_freeze_resume.py`, `test_verify_freeze_resume --vfr`, 31,484 cases per seed, order g = game.exe's) | single, prior freeze, refreeze (inside, on the end frame, in the delay window, after, on the think frame) and think due on the freeze frame: **0 differ on both seeds**. Residuals as before: prior stun 397 and 422, zero-length 308 and 320, specials. Order n (d2d's own tick order): prior freeze 19, delay window 2, think frame 38 (47 on seed 2), think-on-freeze 253 (239): the known tick-order residual |
| Unit tests, `D2D_WERROR=OFF` build | `test_rules`, `test_game`, `test_game_rules`, `test_combat`, `test_monsters`, `test_monster_ai`, `test_protocol`, `test_devctl`, `test_click_keys`: all exit 0 |

Note for whoever re-runs the verifier's freeze harness: the d2d output has log lines on stdout, so filter result lines with `grep -E "^[0-9]+ "`, and the last argument of `compare` is a column number (1), not a seed.

## Live check (V run, one hidden slot, `--no-save` throwaway Sorceress through `d2drill`, every process through `agentrun.sh`)

`D2D_AGENT=forkrebase tools/agentrun.sh forkrebase --ttl 300 -- tools/d2dslot.sh run -n 1 -t 300 -- d2drill -group T -repeat 1 -bin <fork>/bin/d2d.lab2-freezeresume -label forkrebase -v -only D22-C,D22-A` (capacity before: idle 50 plus nice 1 percent, iowait 38). `d2drill` was built from this card's worktree (main `59d4bec`).

- D22-C PASS: walking legs 254 / 578 / 708 ticks, no key, no hops. D22-A PASS: Teleport Stony Tomb leg **211** ticks (3 hops, mana 150 to 87.9, hop ticks 18 18 18), Halls 533, border 717: the DRILLS25 ticks, identical.
- `-group D1 -only D1.P1s,D1.P1`: D1.P1 PASS (belt fill, the bot's own function), **D1.P1s PASS**: the drill's `click x y shift` on a pack potion now runs through upstream's auto-belt and still moves exactly one potion per click (belt 0:hp1, then 0:hp1 4:hp1; pack hp1x2 to none).

So the rebased binary boots, drives the bot and passes the drills that touch the changed code. Not run live: the ctrl-click sell (U; the code path is the old lab one, compiled, `test_click_keys` passes).

## Belt and shift-click: what the bot and the drills must change

Nothing now. Grep of the bot repo for the lab modifiers: the only user is `cmd/d2drill/live.go` (`ShiftClickPack`, verb `click x y shift`, drill D1.P1s in `internal/drill/d1.go`). The `click` verb with modifiers is still in the lab binary (`lab2-debug-skills`), and shift now takes upstream's `auto_belt` path, which is the rule the lab `belt_slot` implemented, so D1.P1s passes unchanged. The bot itself uses `cmd belt put`, not the click. One difference to check in a later card: the lab function refused town and identify scrolls that are not autobelt (`isc`, `tsc`); upstream's `belt_box` states its own rule for scrolls (a scroll goes only to a column whose bottom is the same scroll). U: not compared case by case here; no drill depends on scrolls.

## How to drop branches when upstream merges more

- When upstream merges the armor-hit PR (and the freeze-resume PR, held for the user's go), run `git cherry -v <new upstream main> lab2-freezeresume` in the fork: the commits marked `-` are redundant. Compare the merged trees as for hit-once above (identical hunks, or say exactly how the maintainer's version differs and keep upstream's), then build `lab3-*` branches the same way (new names, no force-push). `lab2-debug-skills` stays: it is lab-only and never upstream.
- If both merge, the fork reduces to `lab2-debug-skills`; `bin/d2d.lab2-freezeresume` is then replaced by a build of that branch and the rule above names the new binary.
- The old `lab-*` branches stay as history; nothing deletes them.

## Shared-file lines for the merger

- docs/BACKLOG.md: `- FORKREBASE: the private fork stack is rebuilt on upstream main 52362d52 as lab2-debug-skills < lab2-armorhit < lab2-freezeresume (docs/forkrebase.md); soaks and drills use bin/d2d.lab2-freezeresume and name it; the hit-once commits are redundant (merged as #158); next: drop armor and freeze commits when upstream merges them`
- docs/INDEX.md: `- docs/forkrebase.md: the private fork stack rebuilt on upstream main (lab2-*): what conflicted, what became redundant, vetting, binary and naming rule`
- CHANGELOG.md (unreleased, docs): `- docs: FORKREBASE: the private d2d fork is rebuilt on upstream main as lab2-freezeresume (docs/forkrebase.md).`
- GLOSSARY.md: none missing.
