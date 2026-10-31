## Why

docs/process/upstream-pr-standard.md section 5 requires an independent adversarial verifier for every upstream PR candidate before the coordinator may ask the user for a go. Two local candidates have none for their new claims: `fix/walk-velocity-v2` (walkfix: the merc and pet commits, the reviewer-runnable mode, the replay on upstream main) and `fix/vendor-prices-followon` (D9 the npc / quest multipliers over 0xffff, D10 the repair gate). This card is that verifier: re-derive game.exe's rule from the disassembly without the authors' oracles, run an own differential against the REAL d2d builds before and after, sweep d2d and game.exe for sibling code paths, and run upstream's CI-class checks.

## What Changes

- `tools/emu/verify_walk_mercpet.py` + `verify_walk_mercpet_mode.inc`, `tools/emu/verify_price_followon.py` + `verify_price_followon_mode.inc`: own generators, game.exe emulator sides and d2d test modes that drive the real d2d functions.
- A verifier section appended to `docs/research/pr-walk-proof.md`; `docs/research/verify-price2.md` (pr-price2-proof.md belongs to the open card decompprice2).
- `docs/research/pr-bundle-queue.md`: the walk and price follow-on rows (READY / NOT READY with reasons); one `docs/BACKLOG.md` line.

owns: tools/emu/verify_walk_mercpet.py, tools/emu/verify_walk_mercpet_mode.inc, tools/emu/verify_price_followon.py, tools/emu/verify_price_followon_mode.inc, docs/research/pr-walk-proof.md, docs/research/verify-price2.md, docs/research/pr-bundle-queue.md, docs/BACKLOG.md, openspec/changes/verifyprs2/**
exclusive: none (offline; own throwaway worktrees of the candidate branches on the worker host, no game instance, no d2d slot; nothing pushed or posted)
gates: none (verification; no spec numbers change)
