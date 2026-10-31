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

Their checks: `test_rules`, `test_combat`, `test_game`, `test_protocol` green; full build rc 0, ctest CTEST_SHORT; `pr_guard.py` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL (1 WARN: build skipped by --quick, done by hand). 6 commits authored by the user, body `Agent: Claude Code`, no trailer; 8 files.

**Commands (reviewer re-runs on their copy).**

```sh
python3 tools/pr_guard.py && cmake --build build && ctest --test-dir build --output-on-failure
cd tools/emu && uv run python price_oracle.py gen 100000 7 > cases.tsv && uv run python price_oracle.py game cases.tsv > g.txt
../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d.txt && uv run python price_terms.py cases.tsv g.txt d.txt   # last line 0 / 0 / 0
```

**What the bot gains.** The bot's own price helpers are fixed separately (card fix-price, findings B1-B6). On d2d the vendor now charges and pays what game.exe would for armor, class items, unidentified loot and broken ethereal items, so drills that sell or repair on d2d see the original's numbers.

**Residual risks.** No stat list in the oracle (magic stat costs, staff mods, charges, sockets not compared); no quests or reduced prices; nothing checked live on the original for these terms; base upstream/main d0ea14b7: fetch again right before a push. Bundle rule: #149 is the open bundle; this waits for it to merge or for the user's "bundle".

**Verifier.** VERIFIER_SHORT

**PROOF: PROOF_STATE.** Recommendation: RECOMMENDATION

Push and PR commands (not run; each needs the user's go):

```sh
git -C <PR checkout>/price2 fetch upstream && git -C <PR checkout>/price2 rebase upstream/main   # only if upstream moved; re-run the oracle, pr_guard, tests
git -C <PR checkout>/price2 push origin fix/vendor-prices
gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/vendor-prices --title "Vendor prices as game.exe prices them" --body-file <body from pr-price-draft.md>
```
