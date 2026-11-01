## Why

The independent verifier (card verifyprs2, docs/research/verify-price2.md and the verifier section of docs/research/pr-walk-proof.md) found two small gaps in the local upstream candidates: two pet paths in `fix/walk-velocity-v2` still follow the owner at RunVelocity x 1.1 with no `// ponytail:` note, and `presubmit` with ruff fails on `fix/vendor-prices-followon` (B905). This card fixes both in NEW branches (`fix/walk-velocity-v3`, `fix/vendor-prices-followon-v2`) built from the candidates' heads, re-runs the CI-class checks and the differentials, and updates our docs with the new heads. Local only: nothing pushed upstream, nothing posted.

## What Changes

- Worker host (not this repo): new branches `fix/walk-velocity-v3` (two ponytail notes, fixup into the pet commit) and `fix/vendor-prices-followon-v2` (`strict=True`, fixup into the tooling commit).
- Docs: `pr-walk-draft.md` and `pr-walk-evidence-pack.md` (limits, new head), `pr-price2-draft.md` and `pr-price2-evidence-pack.md` (new head), and the walk and price follow-on rows of `pr-bundle-queue.md`.

owns: docs/research/pr-walk-draft.md, docs/research/pr-walk-evidence-pack.md, docs/research/pr-price2-draft.md, docs/research/pr-price2-evidence-pack.md, docs/research/pr-bundle-queue.md, openspec/changes/prfixes/**
exclusive: none (offline; own worktrees on the worker host, no game instance, no d2d slot; nothing pushed or posted)
gates: none (no spec numbers change)
