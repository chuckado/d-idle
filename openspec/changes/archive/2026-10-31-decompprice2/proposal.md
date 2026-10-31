## Why

Upstream PR #152 (vendor prices) left two proven-but-unclaimed gaps between d2d and game.exe 1.14d's price function `FUN_0062efb0`: D9, the NPC price multipliers divide by 1024 first above 0xffff, and D10, the repair gate's ItemTypes Repair / durability-column check (decomp-queue items 10 and 11). This card re-derives both from game.exe's code, measures them with an edge-case generator against real d2d builds and builds a local follow-on branch stacked on #152. No observable behaviour change in the bot: research, docs and emulator scripts only; no Go code, no game instance; nothing pushed or posted upstream.

## What Changes

- `docs/research/pr-price2-proof.md`, `pr-price2-draft.md`, `pr-price2-evidence-pack.md`: the claims with addresses and tags, the differential counts per claim on real d2d builds, verifier-ready commands, the draft body, the opening recommendation.
- `tools/emu/price_edge.py`, `tools/emu/price_claims.py`: the edge-case generator and the per-claim counter (copies of the upstream tooling commit's files).
- `docs/decomp-queue.md` rows 10 and 11 and result lines; `docs/research/pr-bundle-queue.md` row; `docs/research/README.md` and `docs/INDEX.md` lines; one `docs/BACKLOG.md` line.

owns: docs/research/pr-price2-*.md, tools/emu/price_edge.py, tools/emu/price_claims.py, openspec/changes/decompprice2/**, docs/decomp-queue.md, docs/research/pr-bundle-queue.md, docs/research/README.md, docs/INDEX.md, docs/BACKLOG.md
exclusive: none (offline; no game instance, no d2d slot, no Go changes; its own upstream worktree on the worker host)
gates: none (research)
