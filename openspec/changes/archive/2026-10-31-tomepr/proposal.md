## Why

DECOMPTOME proved (docs/research/pr-tome-proof.md, 108,000 oracle cases, 22,395 differ) that game.exe 1.14d opens a Town Portal read from the pack on the click tick in any live mode, while d2d starts an SC cast, drops the use while busy and spends the charge at the click. The user asked for the upstream PR candidate ("build the tome PR", 2026-10-31). no observable behaviour change: research docs and oracle scripts in this repo; the d2d patch lives in the local upstream clone on the worker host, nothing pushed or posted.

## What Changes

- Local upstream branch `fix/tome-use` (worker host): a tooling commit (oracle `tools/emu/tome_use.py`, `test_game --tome-use`) and the fix commit (`World::read_portal` opens at once, old pair closed first, charge spent only on success), upstream tests.
- `docs/research/pr-tome-{proof,draft,evidence-pack}.md` finished: after-run, upstream checks, verifier report, limits.
- `tools/emu/tome_use.py` here kept in step with the branch copy.
- `docs/decomp-queue.md` item 2 status.
