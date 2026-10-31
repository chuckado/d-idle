# Upstream bundle queue

> status: living list. Policy from the user (2026-10-31): open the tome PR when #147 merges; collect finished candidates into bundles; keep every bundle under CodeRabbit's single-PR limits.

## Rules
- At most ONE open upstream bundle at a time (the maintainer prefers bundles: "bundle multiple bits together for a larger PR"). A new bundle opens only after the previous one merges, unless the user says "bundle X into <PR>".
- One claim per commit; each claim keeps its own proof (differential against game.exe, independent verifier, `// ponytail:` notes) per docs/process/upstream-pr-standard.md.
- CodeRabbit limits: reviews are SKIPPED above 100 changed files per PR (seen on #148: 168 files, "68 over the limit of 100"); the free OSS review quota is rate limited (waits of ~20 min). So: at most 100 files per bundle, keep a margin (aim at 60 or fewer), and one `@coderabbitai review` trigger per push, never a burst.
- Opening any bundle needs the user's go (each bundle, except the queued items named below); CodeRabbit follow-ups run under the standing go.
- Re-check the base before opening: merge or replay onto the current upstream main (tree layout changed with #148); re-run pr_guard, upstream tests and the oracle.

## Queue (in order)
| Candidate | State | Commits | Notes |
|---|---|---|---|
| #147 belt scrolls | OPEN, being brought up to date with the new layout (UPSYNC147) | 3 + merge | waits for the maintainer |
| tome (inventory Town Portal opens at the click tick) | READY, local, queued by the user: open when #147 merges | 3 | branch fix/tome-use on the worker host; docs/research/pr-tome-*.md |
| shift-click belt / ctrl-click sell / inventory-open pickup | NOT STARTED (decomp-queue item 9) | ? | lab fork implements it; needs a differential proof against game.exe handlers |
| waypoint (decomp-queue item 3) | trace running (DECOMPWP) | ? | only if a deviation is proven |
| portal entry (ENTERPORTAL) | investigation running | ? | only if a d2d deviation is proven |

When the tome PR opens, add any other READY candidate to the same bundle only if the file count stays under the margin above.
