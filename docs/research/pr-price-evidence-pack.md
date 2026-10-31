# Evidence pack: fix/vendor-prices (for the user's go / no-go)

> status: current (agent PRICEPR, Opus, 2026-10-31). One page. Detail: docs/research/pr-price-proof.md; body: docs/research/pr-price-draft.md; trace: docs/research/price-trace.md; standard: docs/process/upstream-pr-standard.md. posted: nothing.

**Claims (one commit each).** game.exe 1.14d's `FUN_0062efb0`: D1 an armor is priced at cost x defence / maxac (buy, sell, repair); D2 a class item sells for a quarter; D3 quality extras count only on identified items; D4 an unidentified item cannot be repaired; D7 a broken ethereal item sells for 1. d2d did none of these.

**Proof summary.** game.exe's own price function runs in upstream's emulator on the loaded 1.14d tables (hooks: stat getters from the case, quest flags 0) against d2d's real `rules::item_price` / `rules::repair_cost` (`test_game --price`), 100,000 cases. The formula in Python (`price_model.py`) equals game.exe on all 100,000 and names the term behind each difference. Every address in the chain read and run (V), no U link.

**Numbers** (buy / sell / repair mismatches; clean subset = cases no unclaimed term touches: 84,030 / 93,095 / 75,112).

| Build | Clean subset | All 100,000 |
|---|---|---|
| upstream | 24,084 / 32,472 / 21,248 | 39,626 / 38,939 / 42,705 |
| + D1 armor | 3,933 / 20,070 / 7,464 | 19,475 / 26,537 / 28,920 |
| + D2 class | 3,933 / 4,189 / 7,464 | 19,475 / 10,648 / 28,920 |
| + D3 identified | 0 / 233 / 7,464 | 15,527 / 6,675 / 28,920 |
| + D4 repair gate | 0 / 233 / 0 | 15,527 / 6,675 / 20,695 |
| + D7 broken ethereal | **0 / 0 / 0** | 15,527 / 6,406 / 20,695 |

The rest of "all" is the unclaimed terms: whole-item repair (D5) 10,841, stacks (D6), divide-first NPC mult above 0xffff (D9, new) 9,232 / 138 / 1,608, repair gate's durability-column check (D10, new) 4,238.

Their checks: `test_rules`, `test_combat`, `test_game`, `test_protocol` green; full build rc 0, ctest's five failures (test_mpq, smoke_d2d, pad_use_d2d, quests_act1/4) identical on upstream d0ea14b7: unchanged; `pr_guard.py` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL (1 WARN: build skipped by --quick, done by hand). 6 commits authored by the user, body `Agent: Claude Code`, no trailer; 8 files.

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python price_oracle.py gen 100000 7 > cases.tsv && uv run python price_oracle.py game cases.tsv > g.txt
../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d.txt && uv run python price_terms.py cases.tsv g.txt d.txt   # last line 0 / 0 / 0
```

**What the bot gains.** The bot's own price helpers are fixed separately (card fix-price, findings B1-B6). On d2d the vendor now charges and pays what game.exe would for armor, class items, unidentified loot and broken ethereal items, so drills that sell or repair on d2d see the original's numbers.

**Residual risks.** No stat list in the oracle (magic stat costs, staff mods, charges, sockets not compared); no quests or reduced prices; nothing checked live on the original for these terms; base upstream/main d0ea14b7: fetch again right before a push. Bundle rule: #149 is the open bundle; this waits for it to merge or for the user's "bundle".

**Verifier.** Independent Opus agent (read-only): **all five claims HOLD**, no counterexample. It re-read every address, compared game.exe's own armo and Class tests with d2d's on all 659 item codes (0 differ), ran 13,920 hand-made edge cases (quality 0..9 x identified x ethereal x 6 durability states x 4 defences x all 17 vendors x 3 difficulties): 0 mismatches outside the unclaimed terms, reproduced the HEAD output byte for byte, and showed each claim's unit test fails when that claim alone is reverted. Limits: stacks not tested in isolation; starter items (flag 0x20000, price 1) and replenish (stat 252) repair are further unclaimed terms; sockets read, not run; base stat 31 = saved defence inferred, not observed. Full report in the proof doc.

**PROOF: complete.** Recommendation: go, when #149 has merged (or on the user's 'bundle' into it, 8 files fits), with the user's go per step. Re-fetch upstream first and re-run the oracle, pr_guard and tests.

Push and PR commands (as run 2026-10-31, see the posted lines below):

```sh
git -C <PR checkout>/price2 fetch upstream && git -C <PR checkout>/price2 rebase upstream/main   # only if upstream moved; re-run the oracle, pr_guard, tests
git -C <PR checkout>/price2 push origin fix/vendor-prices
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/vendor-prices --title "Vendor prices as game.exe prices them" --body-file <body from pr-price-draft.md>
```

posted: 2026-10-31 go: user "push the price PR" (exception to one open bundle: #149 stays open and untouched).
posted: 2026-10-31 fetch upstream: main still d0ea14b7, no replay; pr_guard upstream/main HEAD 0 FAIL / 0 WARN, 8 files, head 1bfff15c (proof as above unchanged).
posted: 2026-10-31 push fix/vendor-prices to the fork, plain push: 1bfff15c (commits d8dae7b6 fcfcd82f 3566b42d 95557754 02e77e74 1bfff15c).
posted: 2026-10-31 PR opened: https://github.com/psi29a/d2decomped/pull/152 (title and body as drafted, ends "Agent: Claude Code").
posted: 2026-10-31 review trigger: not posted; CodeRabbit had commented on its own at once.
posted: 2026-10-31 CodeRabbit review of #152 (2 inline findings, review body held no more): both valid, in scope (go: standing CodeRabbit go). Pushed 70e54f57 to the fork branch fix/vendor-prices, plain fast-forward push from 1bfff15c: test_game --price names (case_id, defense, range-for) and the price_model.py docstring path (re/store.md). No rule change: `--price` output on the 100,000 cases byte-identical to the 1bfff15c run, so the claims' counts are unchanged; test_rules, test_combat, test_game, test_protocol green; pr_guard 0 FAIL 0 WARN; presubmit --quick 0 FAIL.
posted: 2026-10-31 replies in the two CodeRabbit threads (`@coderabbitai valid`): https://github.com/psi29a/d2decomped/pull/152#discussion_r4433287717 and https://github.com/psi29a/d2decomped/pull/152#discussion_r4433287791
posted: 2026-10-31 Windows (MSVC) CI of #152 failed on 70e54f57: tests/test_rules.cpp(509) C4456 'unknown' hides the declaration at line 67 (ours, 95557754; line 509 is upstream's). Renamed our local to unid_magic, one commit a50bdbf3 (body `Agent: Claude Code`), plain fast-forward push to the fork branch fix/vendor-prices (go: standing go for our own PR's CI break). test_rules ok on the worker host; pr_guard 0 FAIL 0 WARN against upstream/main. No comment on the PR. Branch is behind upstream/main 72132b99 and a merge conflicts in tools/emu/README.md (see the report; not rebased).
posted: 2026-10-31 force-push of #152 (go "rebase 152"): fix/vendor-prices rebased onto upstream main after #149 merged (8 commits, one conflict only, tools/emu/README.md: kept our price-oracle row and upstream's tome and auto-belt rows). Old head a50bdbf3, new head 8a45cfc9, `--force-with-lease` pinned to a50bdbf3. Re-verified: test_rules, test_game, test_combat, test_protocol green; price harness 100,000 cases 0 differ; pr_guard 0 FAIL 0 WARN against upstream/main; presubmit --quick 0 FAIL; 8 files in the diff. No comment posted.
