## Why

The freeze-rules candidate `fix/freeze-rules` was independently verified (card verifyfreeze: C5 and C3 hold with limits, C4 holds) but was not ready to open: commit bodies over the three-line standard (and typos), PR body wording (Frozen Armor, flag table, residual, det note), and the fork's #155 head moved from ddc8281d to 8c56e96a so the branch was no longer a fast-forward of it.

## What Changes

- On the worker host (local only, nothing pushed or posted): branch `fix/freeze-rules-v2` from the fork's current `fix/static-field` head, the four commits cherry-picked with bodies cut to the standard and the typos fixed, authored and committed with the same identity as the fork's #155 commits.
- Re-run the gates on the new head (pr_guard, differentials, test binaries, static-field and cold-length modes, clang-tidy, gcc warnings, ruff, size).
- Update docs/research/pr-freeze-draft.md (body), pr-freeze-proof.md (hashes, gate results), the freeze row of pr-bundle-queue.md and one BACKLOG line. The det state hashes need a d2d slot and stay open.

owns: docs/research/pr-freeze-draft.md, docs/research/pr-freeze-proof.md, docs/research/pr-bundle-queue.md, docs/BACKLOG.md, openspec/changes/freezepolish/**
exclusive: none (offline; the fork branch fix/freeze-rules-v2 and throwaway worktrees on the worker host, no game instance, no d2d slot)
gates: none (polish of a local candidate; no spec numbers change)
