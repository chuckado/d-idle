# Search for Cain, CAIN-5b: free Cain and report to Akara (2026-11-18)

> status: current, offline work (agent cain5b), NOT run live. Tags: V verified (a test, or the L2 lab drill fixtures), S source read, U unverified. Inputs: `search-for-cain-l2-2026-11-18.md` (V on the d2d lab build, CHEATED character: mechanics only), `search-for-cain-45-2026-11-18.md` (stones and portal).

## What changed

- `internal/bot/task.go`: `Interact("Cain's Cage")` is `freeCain`. (1) `clearGuards`: while a hostile lives within `bot.clearRadius`, call the normal `Fighter.Clear` (its survival and chicken logic stays active; no class branch), bounded by 150 s and by 3 Clear calls that kill nothing; then a Danger fault, no click. (2) Click the cage through the object adapter, then poll the quest reader up to 6 s for the word of `search_for_cain` to read "reward pending" (bit 1, V: 0x0010 to 0x2012, instantly); up to 3 clicks. Success comes from the word, not from the click. With no live reader (save only) one click counts.
- `cmd/d2bot/objects.go`: the cage is `objClick` (one cage) and the character first walks to a cell 11 subtiles west and 3 north of the gibbet (`objApproach`, V seed 11: the cell the drill walked to, (25,27)); a bare interact from the arrival stalls at (22.7, 24.7). The object is found by class 26 every time, never by index.
- `internal/task/act1.go`: the whole-level `StepClear` of Tristram is gone (Griswold and Bile stand 12+ cells away, not in the way); the cage step does the guards and the click, budget 6 min. The return is the existing `home` node (`ToTown`: TP scroll, `TakePortal(0)`, V L2); the report is the existing board group `search_for_cain_board` (Akara, pick `cain report`).
- `data/questboard/act1.json`: `free_cain` ("Free Cain from the cage in Tristram", steps `clear_cage_guards`, `free_cain`, `return_camp`, nodes `cage`, `home`) and `report` are supported (no `unsupported`).

## Evidence on the board (V, offline)

`free_cain` done = the quest word reads Started+Pending or Done (bit 1 or bit 0), i.e. what the reader maps to Cleared; `report` done = word Done (0x2011). Both come from the quest reader (d2d "dump quests", a shipped verb), not from clicks. `TestCainWordsFromFixtures` reads the three words out of the L2 fixtures; `TestCageApproachFromFixtures` the gibbet row (class 26, NU to ON) and the approach cell.

## Does the sandbox d2d build (devp10/devp11) suffice?

- Needed and shipped: objects with class (presets, V), the quest word (`dump quests`, V: the bot reads it every poll), `npcs`, `move`, `interact`, `debug portals`, TP and `TakePortal`. Nothing here needs a lab-only `debug` verb: the cage state (`debug objects` ON) is NOT used, the word is the proof.
- Only `debug objects` is lab-only/U (the stones use it with a blind fallback, CAIN-4); the cage does not.
- So no private-fork addition is needed for CAIN-5b. U: whether the preset of the gibbet streams in at the arrival (it did on the lab build: row 31 right at arrival, V L2; the adapter explores otherwise).

## Not done / live drill checklist (needs a hidden instance, a clvl 15+ hero)

1. Arrive in Tristram with the three archers awake: expect `campaign clear` style fighting, no click until they are dead; the hero must not die (U: a clvl 17 Paladin; the guard has no level gate).
2. After the click the word reads 0x2012 within 6 s; Cain's NPC row is ignored.
3. `home`: TP scroll in the pack; without one the permanent portal (-2003) to Stony Field is NOT wired (U), only mentioned.
4. In town the report: Akara `Talk` once the board's hand-in tap runs; the word must read 0x2011 (L2: the greeting comes after 3 to 5 interact calls; the menu may be empty, so the poll is on the word).
5. Fighting after the approach walk: the walk itself is guarded by the runner's safety logic only (no second clear).
