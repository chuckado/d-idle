# Vault equip: why it took many taps and left gear in the backpack (2026-11-23, card vaultequip)

User report: equipping from the Vault takes too many clicks, seems not to work, the gear ends up in the inventory, sometimes identified and sometimes not.

## Flow before (V by reading `internal/watch/web/vault.js`, `cmd/d2bot/vaultequip.go`, `internal/vaultkeep/equip.go`)
Open vault, tap the item (detail), Equip (opens the sheet), tick "Recall now" (otherwise the order waits for the next town visit), press Equip: 4 taps, plus the wait. With the bot in the field and the box unticked nothing visibly happens for minutes.

## Root causes, ranked
1. `loot.vaultEquip` defaulted to 0 (V: `internal/env/tuning_loot.go`). Then the page showed NO Equip button, only Take out (Keep), which creates the item in the BACKPACK and never wears it. This is the "gear ends up in the inventory". The page gave no reason. U on the sandbox: its knob value could not be read without the watch token; the live bot log (24 h, read only) holds no vault equip order or `equip:` line at all, only loot lines ("vault full") and a worn set that changed through pickups, which fits "Equip never ran".
2. With the knob on, the order waited for town; the recall box was unticked by default (V code), so the equip looked dead.
3. Identification: Equip does NOT reveal by design (user decision 2026-11-21, docs/vault.md "Equip needs no reveal"); Take out creates the item with whatever identified state the record has, and the gold-only reveal (`loot.vaultRevealMethods` 4) is a separate paid action. So an item looks identified only when it was revealed before. Not changed: it is a user decision; a reveal-before-wear step would need a new go.
4. Earlier real defect already fixed on main: a worn item taken off at a pickup with the same unit id (card recallsorc, `unwear`).

## Fix (this branch)
- `loot.vaultEquip` default 1 (still forced off in measured runs).
- The detail's Equip button is always shown; when off it is disabled and says why (and a tap toasts the reason).
- "Recall now" is ticked by default.
- One tap when the slot is empty: Equip fires the order straight away (`equipOneTap` in `vault-logic.js`); a swap, a ring or a two-hander that pushes a shield off keeps the sheet as the one confirm. Taps after the item is open: 1 (empty slot) or 2 (swap); before: 3 plus a hidden wait. Asserted in `internal/watch/testdata/vault_logic.mjs` (run by `go test ./internal/watch`).
- Server side (swap goes back to the vault, class and requirement refusals, Sorceress and Paladin tests) already existed: `go test ./internal/vaultkeep ./cmd/d2bot -run Equip`.

## Not verified live (U)
Real hero equip with the new default, toast wording on a real phone, whether the sandbox bot had the knob at 0.
