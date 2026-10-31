# Draft PR body: "Town Portal from the pack opens at once, as game.exe's 0x20 handler does" (agent DECOMPTOME, 2026-10-31)

> status: draft, NOT ready (proof incomplete: [pr-tome-proof.md](pr-tome-proof.md) section 4). Nothing pushed, posted or filed. Branch to be: `fix/tome-use` off upstream/main, author shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>, no trailer, body `Agent: Claude Code`.

---

## What

Reading a Scroll of Town Portal, or a Tome of Town Portal with charges, from the inventory opens the portal pair on the same tick, with no cast, in any mode a live player is in. Today d2d starts an SC cast for it (`World::read_portal` -> `Fight::cast_scroll`): the use is dropped while the player attacks, casts, blocks or is in hit recovery, the charge is spent at the click, and a hit during the cast loses it with no portal.

## game.exe

| Address | What it shows |
|---|---|
| `FUN_0054d750` | packet 0x20 refused only when `FUN_0057eec0` says dead (mode 0 / 0x11) or state 0x36 |
| `FUN_0055e170` | a book with stat 0x46 <= 0 returns at once; after a successful use stat 0x46 -1 (`LAB` at `0x55e29b`) or the scroll removed |
| `FUN_005be290` | pSpell 2: town / level 0x88 refused (voice 0x18), `FUN_00535430` old pair closed, `FUN_0056d130` opens; no mode set |
| `FUN_005bf3d0` | the SC cast (srvdofunc 113) is the skill-bar path only; it re-enters `FUN_0055e170` |

## Evidence

`tools/emu/tome_use.py` (attached, not in the PR): `FUN_0054d750` run natively, 108,000 cases; d2d before: 22,395 differ; after: TODO.

## Limits

TODO (proof section 4.5).

Agent: Claude Code
