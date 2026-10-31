# Evidence pack: fix/vendor-prices-followon (for the user's go / no-go)

> status: current (agent DECOMPPRICE2, Opus, 2026-10-31). One page. Detail: [pr-price2-proof.md](pr-price2-proof.md); body: [pr-price2-draft.md](pr-price2-draft.md); base PR: #152 ([pr-price-evidence-pack.md](pr-price-evidence-pack.md)); standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claims (one commit each, stacked on #152's head 8a45cfc9).** game.exe 1.14d's `FUN_0062efb0`: D9 the npc.txt and quest multipliers divide by 1024 first when the price is above 0xffff (`0x62faac`, `0x62fae5`, `0x62fb1d`, quest loop `0x62fb70`); D10 a repair costs 0 unless the item has missing charges or its ItemTypes row has Repair and it restocks or has durability (`FUN_0062e660`, `FUN_00629930`): worn bows, crossbows and throwing potions are not repaired. d2d did neither. Both V (disassembly read, game.exe's function run).

**Proof summary.** game.exe's own price function in upstream's emulator (#152's `price_oracle.py`) against d2d's real `test_game --price` of four builds (upstream rules, #152 head, + D9, + D10). A new generator (`price_edge.py`) aims at the two terms; `price_model.py` equals game.exe on all 120,000 edge cases and names the columns each claim moves (`price_claims.py`).

**Numbers** (mismatches; edge set 120,000 cases + #152's set 100,000):

| Claim | before (#152 head) | after (HEAD 312272a7) |
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

**PROOF: incomplete** (independent verifier missing; everything else done).

**How to open it (recommendation).** Wait for #152 to merge, then rebase `fix/vendor-prices-followon` onto upstream main (`git rebase --onto upstream/main 8a45cfc9`), leaving the tooling commit and the two claims: 3 commits, 7 files, its own small PR. Reasons: #152 is green and cleared by CodeRabbit and only waits on the maintainer; adding two claims re-opens its review (a CodeRabbit round, new CI) for gold-level changes; the stacked commits touch the same `rules.hpp` hunks, so a separate PR before the merge would show #152's eight commits too. The alternative (push the two commits into #152 on the user's 'bundle', 11 files, well under 100) is fine for CodeRabbit limits if the maintainer prefers one PR.

Commands (not run; each needs the user's go):

```sh
git -C <PR checkout>/price3 fetch upstream && git -C <PR checkout>/price3 rebase --onto upstream/main 8a45cfc9 fix/vendor-prices-followon   # after #152 merges; re-run the commands above, pr_guard, tests
git -C <PR checkout>/price3 push origin fix/vendor-prices-followon
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/vendor-prices-followon --title "Vendor prices: npc mults divide first above 0xffff; no repair without a Repair type or durability" --body-file <body from pr-price2-draft.md>
```
