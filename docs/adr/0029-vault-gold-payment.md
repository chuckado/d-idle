---
status: accepted
date: 2026-11-21
decision-makers: user (popup)
---
# 0029: The vault gold reveal is paid at once from the stash gold (amends ADR 0014 for one purpose)

Source: user popup, 2026-11-21 ("Instant via a game-memory write"), on the vault gold reveal built in task 778 (the bot bought Scrolls of Identify at the next town visit). Amends [ADR 0014](0014-memory-writes-allowed-no-crutch.md) for ONE purpose; builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md) (answer listener orders).

## Context

The first gold reveal queued the bot to buy scrolls at the vendor and read them at the next town visit. The user wants the Reveal bar's Gold option to be instant: the cost leaves the STASH gold at once and the items are revealed in the same moment. Nothing in the game can subtract stash gold without either a trip to the stash and the vendor or a game-memory write; the user chose the write, knowing it conflicts with ADR 0014, which allows writes only for SETUP, INSTRUMENTATION, DIAGNOSIS and FIXTURE capture.

## Decision

ADR 0014 gains a fifth allowed write purpose: a PLAYER-PAID VAULT COST on an explicit player order. The only such write is the fork's verb `debug paygold <amount> <orderid>` (private lab fork, flag `--vault-pay`). Guard rails, all of them binding:

1. Never autonomous: the bot never pays on its own; the verb is called only in answer to a player order through the vault API (`POST /api/vault/revealgold`, answer-listener auth).
2. Never in a measured run: the CRUTCH RULE of ADR 0014 is unchanged. Gate, soak, baseline and fidelity runs never use it; the bot refuses when `D2BOT_RUN` names a run (`d2cheat runstart`), and the fork refuses when it runs in lockstep or its environment names a run (`D2BOT_RUN`, `D2D_MEASURED_RUN`).
3. Bounded: never more than the stash gold; never the purse or the inventory gold; amount at most the stash, at least 1.
4. Idempotent by order id: a repeated order id changes nothing (the fork answers `paid:0, already:true`).
5. Logged: every payment goes through `d2dctl.Pay`, which writes the cheat ledger line BEFORE the verb is sent with the reason `vault gold reveal order <id>`; it shows in `d2cheat ledger`. The fork logs its own line too.
6. Off by default: a stock d2d or a lab fork without `--vault-pay` refuses the verb (`--lab-verbs` alone does not pay); `caps` lists `vault.paygold` only with the flag; the sandbox launch passes the flag only when the knob `loot.vaultRevealGold` is on.
7. The bot's own play never depends on it. The reveal itself is the data unlock of the Ether path (affixes the vault already holds): no town visit, no scroll, no game step.

## Consequences

- The gold price is the vendor price of one Scroll of Identify per item (item tables), as before; the stash pays for as many items as it can in the player's list order and the order says so.
- Order chain on the Reveal bar: queued, paid, revealed (seconds), or failed with the reason.
- ADR 0014's CRUTCH RULE and the four earlier purposes stand; docs/cheats.md keeps the general rule, docs/vault.md describes the mechanism.
- Items without held affixes cannot be revealed with gold (same limit as Ether); scrolls, tome and Cain stay the other ways.

## Amended 2026-11-23: the vault sell path credits gold (user decision, vaultsell 2026-11-22)

Source: user approval 2026-11-23, on the instant vault sale built in the vaultsell work (docs/vault.md, "Sell is instant"). The fifth write purpose of this ADR (a player-ordered vault gold movement) now covers the other direction too: a vault SALE adds gold, not only a reveal subtracting it.

- Player-ordered Sell and Sell all credit the sale gold to the hero's STASH gold at once (never the purse). The amount is what the engine's own sell function pays, read with the fork's `item price` verb (and the `debug creditgold` call that adds it, one call per order). The verb is `debug creditgold <amount> <orderid>` (fork flag `--vault-pay`, caps `vault.creditgold`).
- It is a player-ordered action in the UI with a two-tap confirm: a sale worth `loot.vaultSellConfirm` (2000) gold or more needs a second confirming press (`ack`). The instant path is the knob `loot.vaultSellInstant` (default 1); 0 or a fork without the verbs keeps the old town-visit queue.
- All guard rails 1 to 7 above hold unchanged for the credit: never autonomous (the bot never sells on its own through this path), never in a measured run (fork and bot refuse), bounded (the stash may not pass 2,500,000; the sale then stops at the records that fit), idempotent by order id, ledgered BEFORE the verb (reason `player vault sale order <id>`), off without the flag, and the bot's own play never depends on it.
- Refusals leave every record in the Vault. The CRUTCH RULE of ADR 0014 is unchanged.

## Amended 2026-11-25: player-ordered equip through the direct item verbs (user decision, card equipv2)

Source: user request 2026-11-25 ("Equipping gear from the vault is still too slow, we may need to refactor the whole flow and cheat the items in/out of the character"). The fifth write purpose of this ADR (a player-ordered vault action) now also covers EQUIP: the fork's lab verbs `item equipdirect <hex> <slot> [identified=1]`, `item unequipdirect <slot>` and `item wearset <slot>=<hex> ...` (up to 12, one call; fork branch didle/directequip-20261125, caps `item.equipdirect`, `item.unequipdirect`, `item.wearset`) create the unit straight into the body slot and hand back the units that came off.

- Player-ordered only: the vault page's Equip and Wear set (answer listener auth) call them; the bot's own play never does (its gear logic, the town chores and the leveler keep the click and desk paths).
- Never in a measured run: the bot refuses when `D2BOT_RUN` names a run and the fork refuses a measured run (lockstep or its environment). The CRUTCH RULE of ADR 0014 is unchanged; the writes happen outside measured runs only.
- Logged: `d2dctl.DirectEquip` writes the cheat ledger line BEFORE the verb, reason `player vault equip order <id>`, and completes it with the result; an unwritable ledger stops the call.
- Nothing is lost: the equipped record leaves the Vault and every unit that came off becomes a Vault record at once (identified state kept, found zone "previously worn"); when the Vault cannot take it the item is created back on the hero (stash, else backpack under the keep mark). The order record keeps the removed bytes until the records exist (a restart finishes the step).
- No waiting for a safe moment: the verb is one atomic engine command on the shared control socket, so it may run while the hero fights, between two bot commands, in the field or in town. Unrevealed records are refused (`reveal it first`: the game will not wear an unidentified item); requirements (class, level, strength, dexterity, hands) are the engine's own answer, per piece.
- Knob `loot.vaultDirectEquip` (default 1; needs `loot.vault` 1). A fork build without the verbs (probed with `caps` at start, cached 60 s) or the knob at 0 keeps the order flow of `loot.vaultEquip` (docs/vault.md, Equip v2).

