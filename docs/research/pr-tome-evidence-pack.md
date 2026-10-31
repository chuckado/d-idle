# Evidence pack: fix/tome-use (for the user's go / no-go)

> status: current (agent TOMEPR, 2026-10-31; first pack DECOMPTOME). One page. Detail: [pr-tome-proof.md](pr-tome-proof.md); body: [pr-tome-draft.md](pr-tome-draft.md); trace: [town-portal-trace.md](town-portal-trace.md); standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claim.** A Town Portal scroll or a charged tome used from the pack opens the portal on the click tick in any live mode (no SC), closes the old pair first and spends the charge only if the portal opened; an item in hand refuses it (game.exe `FUN_0055e170` -> `FUN_005be290`). d2d cast SC, dropped the use while busy, spent the charge at the click and had no item-in-hand test.

**Proof summary.** game.exe's own `FUN_0054d750` chain (dispatcher, `FUN_0057eec0`, `FUN_0054b1e0`, `FUN_0055e170`, `FUN_005bf240`, `FUN_005be290`) runs in upstream's emulator (`tools/emu/tome_use.py`, in the branch; hooks listed in its docstring) against d2d's real `World::apply(cmd::UseItem)` (`test_game --tome-use`), same cases, joined by `diff`. Every link V (read with disasm.py and run).

**Numbers.**

| Scenario | Cases | before (d72b97a7) | claim A (112f5de0) | claim B (c56670b9) |
|---|---|---|---|---|
| Grid (tbk / tsc x 25 charges x town x in hand x modes 0..19 x spot) + 100,000 random, seed 7 | 108,000 | 29,551 | 23,127 | 0 (md5 equal) |
| 200,000 random + grid, seed 31 | 208,000 | 56,616 | 44,392 | 0 |

Seed 7 classes before: busy + spot (game opens now, d2d nothing) 8,360; busy, no spot (game closes the old pair, d2d nothing) 8,337; idle + spot (game now, d2d at tick 9) 3,194; idle, no spot (game keeps the charge and closes the old pair, d2d spends it and keeps the pair) 3,236; item in hand (game nothing, d2d casts) 6,424.

Their checks on c56670b9: `test_game` rc 0 (29 OK, two new blocks), `test_combat` rc 0, `test_rules` rc 0 (Release, `D2D_WERROR` on), `pr_guard.py upstream/main` 0 FAIL 0 WARN, `presubmit.py --quick` 0 FAIL 4 WARN (ruff absent; build by hand; two sibling questions answered in the body). Three commits authored by the user, body one line `Agent: Claude Code`, no trailer.

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python tome_use.py gen 100000 7 > cases.tsv && uv run python tome_use.py game cases.tsv > g.txt
../../build/tests/test_game --tome-use cases.tsv | grep -E '^[0-9]+ ' | diff g.txt -     # empty
```

**What the bot gains.** On d2d a portal read under attack works as in game.exe (no dropped click while in get-hit, no charge lost to a cut cast): fewer failed trips and deaths in soaks (town-portal-trace.md section 3: 119 "tome still charged" failures fit the busy drop or a cut cast). Not measured; the bot fix (FIXPORTAL) must still handle today's d2d.

**Residual risks.**
- State 0x36 and level 0x88 not compared (ponytail); belt / stash items and the game.exe client side outside the generator; `FUN_00628c20` and the spot search hooked.
- `tools/d2ddet` hashes not run.
- Branch on upstream/main 27113a31; fetch again right before a push.

**Verifier.** PENDING.

**PROOF: PENDING.**

**Push and PR commands (NOT run; each needs the user's go).**

```sh
git -C <pr checkout>/tome push origin fix/tome-use
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/tome-use --base main \
  --title "A Town Portal read from the pack opens at once, as game.exe does" --body-file <the body in pr-tome-draft.md>
```
