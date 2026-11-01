# NextHit v2: posted log

> status: current (card opennexthit, 2026-11-01). Evidence: [pr-nexthit-v2-addendum.md](pr-nexthit-v2-addendum.md), [pr-nexthit-proof.md](pr-nexthit-proof.md), [pr-nexthit-verifier.md](pr-nexthit-verifier.md). Body: the draft plus the addendum's five edits.

- posted: 2026-11-01 PR #158 opened https://github.com/psi29a/d2decomped/pull/158 (title 'game: a NextHit hit marks the unit justhit, one unit a frame (Nova / Frost Nova, differential proof vs game.exe)'; footer `Agent: Claude Code`), head 6b3964fa = branch fix/nexthit-justhit on our fork (2 commits, 8 files, base upstream main 72132b99, unchanged), pushed as a new branch (plain push, no force), opened with `gh pr create -R psi29a/d2decomped --head shatteredeagle:fix/nexthit-justhit`. go: user, 2026-11-01 (popup 'Open the hit-once PR (Recommended)'). No CodeRabbit trigger posted.
- posted: 2026-11-01 final gates on 6b3964fa (the worker host): pr_guard 0 FAIL / 0 WARN (range and each commit); test_rules, test_combat, test_protocol, test_game, test_monsters rc 0; presubmit --quick 0 FAIL / 4 WARN (answered in the body); gcc -Wshadow -Wunused -Wsign-compare 0 warnings on added lines; clang-tidy 136 TUs, 0 findings in touched files; ruff with the upstream rules plus B905 clean; seed is 32-bit in both generators; 8 files.
- posted: 2026-11-01 differentials on the pushed head: author's harness 100,198 cases, T 0 / H 0 differ; verifier's harness 100,624 cases (4 seeds), 0 differ in all slices (I order); contact order P changes 6 cases (the stated limit).
- posted: 2026-11-01 shared files: the head merges clean with #152, #153, #155, #156, #157 and with #155 + #156 together, and #155 and #156 merge clean with the head (stated in the body).
- posted: 2026-11-01 Discord upstream notice for #158 sent to #d2d-upstream (one post).

## Shared-file lines for the merger

- docs/BACKLOG.md: `- NEXTHIT v2: OPEN upstream as #158 (head 6b3964fa, fork fix/nexthit-justhit); open: the coordinator's CodeRabbit trigger, review answers (docs/research/pr-nexthit-posted.md)`
- docs/INDEX.md, research: `- docs/research/pr-nexthit-posted.md: NextHit v2 posted log (PR #158, final gates, differentials)`
- CHANGELOG.md: none.
