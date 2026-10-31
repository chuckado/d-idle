# Evidence pack: fix/tome-use (for the user's go / no-go)

> status: current (agent TOMEPR, 2026-10-31; first pack DECOMPTOME). One page. Detail: [pr-tome-proof.md](pr-tome-proof.md); body: [pr-tome-draft.md](pr-tome-draft.md); trace: [town-portal-trace.md](town-portal-trace.md); standard: docs/process/upstream-pr-standard.md. posted: see the dated posted: lines below (PR #149 opened 2026-10-31).

**Claim.** A Town Portal scroll or a charged tome used from the pack opens the portal on the click tick in any live mode (no SC), closes the old pair first and spends the charge only if the portal opened; an item in hand refuses it (game.exe `FUN_0055e170` -> `FUN_005be290`). d2d cast SC, dropped the use while busy, spent the charge at the click and had no item-in-hand test.

**Proof summary.** game.exe's own `FUN_0054d750` chain (dispatcher, `FUN_0057eec0`, `FUN_0054b1e0`, `FUN_0055e170`, `FUN_005bf240`, `FUN_005be290`) runs in upstream's emulator (`tools/emu/tome_use.py`, in the branch; hooks listed in its docstring) against d2d's real `World::apply(cmd::UseItem)` (`test_game --tome-use`), same cases, joined by `diff`. Every link V (read with disasm.py and run).

**Numbers.**

| Scenario | Cases | before (29e75c76) | claim A (c5e2a0b1) | claim B (e35ab5ca) |
|---|---|---|---|---|
| Grid (tbk / tsc x 25 charges x town x in hand x modes 0..19 x spot) + 100,000 random, seed 7 | 108,000 | 29,551 | 23,127 | 0 (md5 equal) |
| 200,000 random + grid, seed 31 | 208,000 | 56,616 | 44,392 | 0 |

Seed 7 classes before: busy + spot (game opens now, d2d nothing) 8,360; busy, no spot (game closes the old pair, d2d nothing) 8,337; idle + spot (game now, d2d at tick 9) 3,194; idle, no spot (game keeps the charge and closes the old pair, d2d spends it and keeps the pair) 3,236; item in hand (game nothing, d2d casts) 6,424.

Same numbers on the pre-split run (d72b97a7 / 112f5de0 / c56670b9 on 27113a31). Their checks on e35ab5ca (upstream/main 03f42254): `test_game` rc 0 (two new OK blocks), `test_combat` rc 0, `test_rules` rc 0 (Release, `D2D_WERROR` on), `pr_guard.py upstream/main` 0 FAIL 0 WARN, `presubmit.py --quick` 0 FAIL 4 WARN (ruff absent; build by hand; two sibling questions answered in the body). Three commits authored by the user, body one line `Agent: Claude Code`, no trailer.

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python tome_use.py gen 100000 7 > cases.tsv && uv run python tome_use.py game cases.tsv > g.txt
../../build/tests/test_game --tome-use cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -     # empty
```

**What the bot gains.** On d2d a portal read under attack works as in game.exe (no dropped click while in get-hit, no charge lost to a cut cast): fewer failed trips and deaths in soaks (town-portal-trace.md section 3: 119 "tome still charged" failures fit the busy drop or a cut cast). Not measured; the bot fix (FIXPORTAL) must still handle today's d2d.

**Residual risks.**
- State 0x36 and level 0x88 not compared (ponytail); item mode, requirements, range, belt items and the game.exe client side outside the generator; `FUN_00628c20` and the spot search hooked; sounds not compared.
- `tools/d2ddet` hashes not run.
- Branch on upstream/main 03f42254 (rebased over the file split); fetch again right before a push.

**Verifier.** Independent Opus agent (own worktree): **VERDICT: holds with limits.** Re-read the gate (`FUN_0057eec0`: only state 0x36, mode 0 / 0x11 refuse), `FUN_0055e170` (no mode test, no SC, spend only after success) and `FUN_005be290` (close before make); reproduced 29,551 / 23,127 / 0, seed 31 0, fresh seed 1234 0; its own driver: 1 -> 0 charges, two reads in a row, walled spot, modes 0..19, a read during a skill-bar SC, co-op pairs: all as claimed. Limits: portal position not compared, sounds not compared, state 0x36 / level 0x88 / item mode / requirements / range not modelled, joined game not run; it flagged upstream's file split, since rebased (proof doc). Full report in the proof doc.

**PROOF: complete.** Recommendation: go, with the user's go per step (push `fix/tome-use`, open the PR with the draft body; or join a bundle as its own commits, per the maintainer's preference). Fetch upstream again right before a push.

**Push and PR commands (NOT run; each needs the user's go).**

```sh
git -C <pr checkout>/tome push origin fix/tome-use
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/tome-use --base main \
  --title "A Town Portal read from the pack opens at once, as game.exe does" --body-file <the body in pr-tome-draft.md>
```

posted: 2026-10-31 UTC upstream moved 5 commits since 03f42254 (#147 merge 0dc4c62a, f4c61afe); replayed the three commits locally onto upstream/main 0dc4c62a (rebase, branch not yet pushed, no force): one textual conflict in tools/emu/README.md (our oracle row next to #147's belt row, kept both); new hashes fd2dc256 (oracle and test mode), 64a84587 (claim A), a6d6fe51 (claim B), authored shatteredeagle, body 'Agent: Claude Code'; test_game keeps `--belt-pickup` and `--tome-use` side by side; pr_guard against upstream/main 0 FAIL 0 WARN, presubmit --quick 0 FAIL, test_game / test_combat / test_rules rc 0 (gcc 16 with -Wno-error=maybe-uninitialized and array-bounds), ctest 58 tests: not green and unrelated (same set as the #147 pack): test_mpq, smoke_d2d, quests_act1, quests_act4; oracle seed 7 (108,000): before 29,551, after 0; seed 31 (208,000): before 56,616, after 0; PR diff 8 files go: user 'Open the pr when 147 merges' 2026-10-31
posted: 2026-10-31 UTC pushed branch fix/tome-use-opens-portal-at-once to our fork (plain push), head a6d6fe51; the commit hashes quoted in the body were updated to the replayed ones go: user 'Open the pr when 147 merges' 2026-10-31
posted: 2026-10-31 UTC opened PR #149 https://github.com/psi29a/d2decomped/pull/149 against psi29a/d2decomped main, title and body from pr-tome-draft.md, 3 commits, 8 files, no other comment go: user 'Open the pr when 147 merges' 2026-10-31
posted: 2026-10-31 UTC review trigger not needed: CodeRabbit started its review on its own within a minute (no comment posted) go: user 'Open the pr when 147 merges' 2026-10-31
posted: 2026-10-31 UTC PR #149 now bundles the shift-click claim (97a590c7, ecb267db) on top of this one (head ecb267db, 5 commits, 19 files); the tome commits are unchanged (fd2dc256, 64a84587, a6d6fe51); re-proved on the combined branch: tome oracle seed 7 0 of 108,000 after, 29,551 with the two fix commits reverted; title and body edited go: user 'bundle click' 2026-10-31
posted: 2026-10-31 UTC pushed 3 commits to our fork branch fix/tome-use-opens-portal-at-once (plain fast-forward push ecb267db..0dc3e0d3): c8a9dca8 (the Town Portal read checks run before the patch-only return, in their own world; `--tome-use` case_id), 6a98903a (ponytail note: a pending portal click, take_portal, is kept), 0dc3e0d3 (ruff F841 in tome_use.py); pr_guard 0 FAIL 0 WARN, presubmit --quick 0 FAIL, test_game with and without D2_PATCH_INSTALLER, test_combat, test_rules, test_protocol rc 0; oracles unchanged: tome seed 7 0 of 108,000 mismatches, auto-belt 390,000 0; PR diff 20 files go: standing CodeRabbit go
posted: 2026-10-31 UTC replied in the two CodeRabbit inline threads of PR #149: https://github.com/psi29a/d2decomped/pull/149#discussion_r4432993489 (test_game guard: valid, fixed) and https://github.com/psi29a/d2decomped/pull/149#discussion_r4432993625 (take_portal: not changed, FUN_00535430 disassembled: unlinks the pair, writes nothing to the player; the walk to a clicked portal is the client's, outside the oracle; ponytail note) go: standing CodeRabbit go
