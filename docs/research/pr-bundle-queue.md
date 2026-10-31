# Upstream bundle queue

> status: living list. Policy from the user (2026-10-31): open the tome PR when #147 merges; collect finished candidates into bundles; keep every bundle under CodeRabbit's single-PR limits.

## Rules
- At most ONE open upstream bundle at a time (the maintainer prefers bundles: "bundle multiple bits together for a larger PR"). A new bundle opens only after the previous one merges, unless the user says "bundle X into <PR>".
- One claim per commit; each claim keeps its own proof (differential against game.exe, independent verifier, `// ponytail:` notes) per docs/process/upstream-pr-standard.md.
- CodeRabbit limits: reviews are SKIPPED above 100 changed files per PR (seen on #148: 168 files, "68 over the limit of 100"); the free OSS review quota is rate limited (waits of ~20 min). So: at most 100 files per bundle, keep a margin (aim at 60 or fewer), and one `@coderabbitai review` trigger per push, never a burst.
- Two open PRs by the user's explicit exception 2026-10-31: #149 stable, price PR (#152) for the CodeRabbit round.
- Opening any bundle needs the user's go (each bundle, except the queued items named below); CodeRabbit follow-ups run under the standing go.
- Re-check the base before opening: merge or replay onto the current upstream main (tree layout changed with #148); re-run pr_guard, upstream tests and the oracle.

## Queue (in order)
| Candidate | State | Commits | Notes |
|---|---|---|---|
| #147 belt scrolls | MERGED upstream 2026-10-31T17:45Z (0dc4c62a) | 3 + merge | done |
| tome (inventory Town Portal opens at the click tick) | OPEN #149 https://github.com/psi29a/d2decomped/pull/149 (2026-10-31, user go); bundles 2 claims (tome + shift-click belt) since 2026-10-31 (user 'bundle click'), 19 files, head ecb267db | 5 | branch fix/tome-use-opens-portal-at-once; docs/research/pr-tome-*.md; the one open bundle until it merges |
| shift-click belt (decomp-queue item 9, docs/research/pr-click-*.md) | BUNDLED into #149 (2026-10-31, user 'bundle click'): commits 97a590c7 (claim) and ecb267db (tooling); 390,000 cases 129,283 -> 0 re-proved on the combined branch | 2 commits (in the 5 of #149) | ctrl-click sell and inventory-open pickup: client-only, no PR (docs/research/click-handlers-trace.md) |
| waypoint (decomp-queue item 3, docs/research/pr-wp-*.md) | local, proof incomplete, needs the independent verifier | 6 files, 1 commit | branch fix/waypoint-hostile-window on the worker host |
| vendor prices (decomp-queue item 5, docs/research/pr-price-proof.md) | OPEN #152 (2026-10-31, user go): 5 claims (D1 armor, D2 class sell, D3 identified gate, D4 no unidentified repair, D7 broken ethereal), clean subset 24,084 / 32,472 / 21,248 -> 0 / 0 / 0 of 100k; upstream tests green (5 ctest failures identical on upstream); pr_guard 0/0; Opus verifier: all hold; docs/research/pr-price-evidence-pack.md | 6 commits, 8 files | branch fix/vendor-prices on upstream/main d0ea14b7 on the worker host, head 1bfff15c; pushed to the fork and opened as #152 |
| walk velocity (decomp-queue item 12, docs/research/pr-walk-*.md) | local, oracle proof complete (712,362 cases, 425,665 -> 0), upstream tests green, pr_guard 0/0; independent verifier NOT run | 3 commits, 5 files | branch fix/walk-velocity on upstream d0ea14b7 on the worker host (96953bfe monster RN at Velocity, 7e8feb87 chill adds coldeffect + floor 25, 97d55732 player velocitypercent raw) |
| portal entry (ENTERPORTAL) | investigation running | ? | only if a d2d deviation is proven |
| monster missile gates (decomp-queue item 7, docs/research/monster-attack-trace.md) | deviations proven (MissA ToHit gate: 94 of 142 rows; can-block = physical: 10,503 of 100,000 oracle cases); no branch, no fix, no verifier | 2 claims planned | no early-game effect; low priority bundle filler |

When the tome PR opens, add any other READY candidate to the same bundle only if the file count stays under the margin above.
