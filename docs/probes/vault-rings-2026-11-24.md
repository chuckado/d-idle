# Why the Vault never holds a ring (2026-11-24)

Question (user): "there's never rings in the vault: are we not recording them or just none dropped?"

## Answer
(a) Rings drop and the bot sees them. (b) It picks them up into the backpack and never records them in the vault. (c) The page is not hiding them: nothing is there. One rule bug, fixed here; a second, separate cause (the vault is full) also blocks all new records.

## Evidence (V unless tagged)
Sandbox bot log, read-only (`docker logs`, no token, no save), window 2026-11-24 05:44Z to 15:29Z (about 9.75 h, the whole container life):
- `loot: clicked` pickups of jewelry and charms: 15 Ring (14 magic, 1 rare), 5 Amulet, 14 charms (7 Small, 4 Grand, 3 Large). All were pick-up clicks into the backpack (`pickOne`), none went through the vault path (`vaultOne`). Zero `vault` records of a ring, amulet or charm.
- Jewels (the one jewelry type with a vault rule that fired): 2 `loot shadow: ground Jewel ... g-jewels`, and 7 `vault full: magic Jewel` refusals.
- The rings picked up then hit `sell Ring today=sell rules=keep s-jewelry-stash ... Kept: the stash is full`: they stay in the pack for the stash chore, not the vault.
- `vault full: new drops are not recorded until room is made` at 05:45:48Z and 440 `vault full: <item>` skips since (V from the log; the count of records and the capacity are U: the vault stats need the watch token, not read).

## Cause (V: code and a failing test)
`lootrules.VaultPreset` turned the Starter's ground `stash` rules into `vault`, but the rules that match a drop on the floor are `g-jewelry-id`, `g-small-charm-id`, `g-big-charm-id` (action `keep`, "unidentified jewelry"). A drop is always unidentified, so those fire first (prio 55 and 70); `g-good-jewelry` (stash, needs identified) can never match on the ground. The preset kept `keep` rules as pickups, so the ring went to the pack for the town identify, then sell or stash.

## Fix
`VaultPreset` now turns those three `keep` rules into `vault` (an unidentified magic or rare ring, amulet, circlet or charm is peeked and recorded; the gold reveal opens it in the vault, as for any magic item). Test `TestVaultPresetUnidentifiedJewelry` (Sorceress and Paladin, record-all on and off): fails without the change (ring, amulet, small and grand charm decided `pickup` by `g-jewelry-id` / `g-small-charm-id`), passes with it. Not run live: U that a recorded unidentified ring shows on the page (magic weapons already do the same path).

## Not fixed
The vault being full (placeholder capacity, no trim by default per the 2026-11-21 decision) will still refuse new rings until room is made or capacity grows; that is a product decision, not a bug.
