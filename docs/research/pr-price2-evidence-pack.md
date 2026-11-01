# Evidence pack: fix/vendor-prices-followon-v2 (for the user's go / no-go)

> status: current (agent DECOMPPRICE2, Opus, 2026-10-31; updated by card prfixes: branch `fix/vendor-prices-followon-v2`, head da80d0ab, independently verified HOLDS WITH LIMITS in [verify-price2.md](verify-price2.md); the old `fix/vendor-prices-followon` (312272a7) is untouched). One page. Detail: [pr-price2-proof.md](pr-price2-proof.md); body: [pr-price2-draft.md](pr-price2-draft.md); base PR: #152 ([pr-price-evidence-pack.md](pr-price-evidence-pack.md)); standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claims (one commit each, stacked on #152's head 8a45cfc9).** game.exe 1.14d's `FUN_0062efb0`: D9 the npc.txt and quest multipliers divide by 1024 first when the price is above 0xffff (`0x62faac`, `0x62fae5`, `0x62fb1d`, quest loop `0x62fb70`); D10 a repair costs 0 unless the item has missing charges or its ItemTypes row has Repair and it restocks or has durability (`FUN_0062e660`, `FUN_00629930`): worn bows, crossbows and throwing potions are not repaired. d2d did neither. Both V (disassembly read, game.exe's function run).

**Proof summary.** game.exe's own price function in upstream's emulator (#152's `price_oracle.py`) against d2d's real `test_game --price` of four builds (upstream rules, #152 head, + D9, + D10). A new generator (`price_edge.py`) aims at the two terms; `price_model.py` equals game.exe on all 120,000 edge cases and names the columns each claim moves (`price_claims.py`).

**Numbers** (mismatches; edge set 120,000 cases + #152's set 100,000):

| Claim | before (#152 head) | after (HEAD da80d0ab; v1 312272a7 gave the same) |
|---|---|---|
| D9 buy / sell / repair | 46,730 / 545 / 14,356 + 9,232 / 154 / 1,608 | 0 / 0 / 0 + 0 / 0 / 0 |
| D10 repair | 15,519 + 5,163 | 0 + 0 |
| clean subset (no stack, no whole-item term) | 46,730 / 545 / 28,473 + 9,232 / 154 / 5,947 | **0 / 0 / 0 + 0 / 0 / 0** |

Newly wrong: 0 for each commit. D10 changes no buy or sell line. The queue's numbers hold; two were single-term buckets: D9 sell is 154 (138 single-term), D10 is 5,163 (4,238 single-term).

**Their checks.** `test_rules`, `test_combat`, `test_protocol` ok, `test_game` rc 0 (game data dir); clang-tidy with the repo config over the touched TUs: 0 findings in our lines (4 pre-existing in `components/drlg/outdoor.hpp`, untouched); gcc `-Wshadow -Wunused -Wsign-compare`: 0 warnings; `pr_guard.py upstream/main HEAD` 0 FAIL / 0 WARN; `presubmit.py --quick` 0 FAIL, 1 WARN (build skipped, done by hand). 3 commits authored and committed by the user's numeric-ID noreply identity, body `Agent: Claude Code`, no trailer. Stack vs upstream/main 72132b99: 11 commits, 11 files (this change alone: 7 files).

**Commands (reviewer re-runs on their copy).**

```sh
cmake --build build --target test_game test_rules test_combat test_protocol && build/tests/test_rules && python3 tools/pr_guard.py
cd tools/emu && uv run python price_edge.py 120000 11 > cases.tsv && uv run python price_oracle.py game cases.tsv > g.txt
../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d.txt && uv run python price_claims.py cases.tsv g.txt d.txt   # bigmul 0/0/0, nodur repair 0, clean 0/0/0
```

**What the bot gains.** On d2d, dear items (Act 5 vendors x 2, uniques, high-defence armor) and bows / crossbows now cost what game.exe charges; drills that buy, sell or repair those see the original's numbers. The bot's own estimates are card fix-price's (B1-B6).

**Residual risks.** No stat list (charges, staff mods, stat costs, sockets); charge-only repairs of items without durability add the full base in game.exe (not claimed, ponytail); quest mults covered by the unit test and the disassembly, not by the differential; 32-bit wrap not modelled (no real price near it); stat 152 / 252 details as in #152; nothing live. **No independent verifier yet.**

**v2 re-run (head da80d0ab, worker host).** `test_game --price` on the same 120,000 edge cases (seed 11) and #152's 100,000: bigmul 0 / 0 / 0, nodur repair 0, clean 0 / 0 / 0 on both sets (220,000 cases, as the author's); the d2d output is byte-identical to the v1 head's (312272a7) on the edge set. The `nodur` sell column's 825 / 759 and the all-case buckets are the whole-item and stack terms d2d still lacks (not claimed). pr_guard 0 FAIL / 0 WARN (range against upstream/main and each of the 11 commits); `presubmit --quick` with ruff on PATH: 0 FAIL on the follow-on's own lines (see below for #152's two B905 lines); built with `-Wshadow -Wunused -Wsign-compare`: 0 warnings on added lines; test_rules, test_combat, test_protocol, test_game (data dir) rc 0; clang-tidy over the 135 selected TUs: 0 findings in the touched files (pre-existing ones in untouched files only); tooling unit tests OK.

**Verifier and fixes (card prfixes, 2026-10-31).** The independent verifier (card verifyprs2) found D9 and D10 HOLD WITH LIMITS and one gap: `presubmit` with ruff failed on `tools/emu/price_claims.py:41` (B905). v2 fixes that one line (`zip(m.price(case), normal, strict=True)`, fixed up into the tooling commit 777df9d7; the tree differs from 312272a7 by that line only). ruff (upstream rule set F,E9,B,BLE001,PLW1510,S602,S604) is clean on `price_claims.py` and `price_edge.py`; `presubmit --quick --base 8a45cfc9` (the follow-on's own lines) 0 FAIL. Against upstream/main today the same ruff 0.17 also flags B905 on #152's `price_oracle.py:39` and `price_terms.py:47` (not our lines, not touched; they pass with `--target-version py39`, so an older ruff (presubmit's uvx fallback pins 0.8.6; not run here) probably passes them, and they drop out of the added lines once #152 is merged). Open steps: #152 must merge first, then the rebase recipe above, then re-run the checks; the det state-hash step does not apply to prices.

**PROOF: complete for the offline checks** (independent verifier HOLDS WITH LIMITS); waits for #152.

**How to open it (recommendation).** Wait for #152 to merge, then rebase `fix/vendor-prices-followon-v2` onto upstream main (`git rebase --onto upstream/main 8a45cfc9`), leaving the tooling commit and the two claims: 3 commits, 7 files, its own small PR. Reasons: #152 is green and cleared by CodeRabbit and only waits on the maintainer; adding two claims re-opens its review (a CodeRabbit round, new CI) for gold-level changes; the stacked commits touch the same `rules.hpp` hunks, so a separate PR before the merge would show #152's eight commits too. The alternative (push the two commits into #152 on the user's 'bundle', 11 files, well under 100) is fine for CodeRabbit limits if the maintainer prefers one PR.

Commands (not run; each needs the user's go):

```sh
git -C <PR checkout>/price4 fetch upstream && git -C <PR checkout>/price4 rebase --onto upstream/main 8a45cfc9 fix/vendor-prices-followon-v2   # after #152 merges; re-run the commands above, pr_guard, tests
git -C <PR checkout>/price4 push origin fix/vendor-prices-followon-v2
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/vendor-prices-followon-v2 --title "Vendor prices: npc mults divide first above 0xffff; no repair without a Repair type or durability" --body-file <body from pr-price2-draft.md>
```
