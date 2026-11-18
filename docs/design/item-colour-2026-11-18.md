# Item colour in the inventory art: chain and slice plan (2026-11-18)

> status: design, task-634 (card itemcolour). Follow-up to jewelcolor (task-632), which removed the invented wash. Tags: V = read in code or data and checked, S = secondary, U = unverified. No game art in the repo; nothing here needs the game or the fork build to read.

## 1. How game.exe colours an item picture

Source: the private fork's decomp notes (`docs/research/re/compcode.md`, V: written against game.exe 1.14d and checked against 20 real saves for the character-side tints) and the fork's working implementation (`components/compcode/compcode.hpp` `Colours::of`, `apps/d2d/scene.hpp` `Scene::item_map`, `gamedata_load.cpp` `load_tables`; V: read, and d2d draws inventory items with it).

An item's picture is the grey base DC6 (`invfile`) drawn through ONE colormap, chosen by two numbers:

1. Colormap SET = the base's `InvTrans` column (armor/weapons/misc.txt). Only 1, 2 and 5..8 have a set (`tints_with`); 0, 3, 4 mean no tint. Sets live in the install: `data\global\items\palette\<name>.dat` (grey, grey2, gold, brown, greybrown, invgrey, invgrey2, invgreybrown = sets 1..8), 21 x 256 bytes each (V: the fork requires >= 21*256).
2. Colour index 0..20 (Colors.txt row, whit .. bwht) from FUN_0062c100 with `inv` set:
   - unique: `UniqueItems.invtransform` by unique id; set item: `SetItems.invtransform` by set id.
   - magic or rare: the FIRST suffix with a `transformcolor` (magic: suffix; rare: the three rare suffixes in order), else the first prefix with one, else the class automod (`AutoMagic.transformcolor`, 1-based id).
   - anything else: if socketed, the first socketed gem's `gems.txt transform` (not runes or jewels), else the automod.
   - Affix ids index the raw table rows including separator rows (the save's ids; V compcode.md).

Apply: colormap bytes `[colour*256 .. +256)` map each DC6 pixel's palette index to another palette index; then the normal palette (`pal.dat`) gives RGB. Index 0 stays transparent (`Scene::mapped`: `entries[i] = base[i==0 ? 0 : map[i]]`). Ground items and the paper doll use `Transform` / `chrtransform` instead (inv = false); the inventory, stash, store, belt and cursor use `InvTrans` / `invtransform`.

Jewels: all share base `jew` art. A magic jewel's colour comes from its suffix or prefix `transformcolor`, a rare jewel's from its first coloured rare affix, a unique's from `UniqueItems.invtransform`. The same chain colours magic/rare rings, amulets and all tinted gear. U: the actual `jew` InvTrans value and which affix rows carry a colour need the real tables (no install on this box); the slice-1 test below prints them.

## 2. What the bot payload carries

- Inventory and stash (what the watch shows) come from `dump items` rows (`internal/d2dsrc/items.go dumpItem`): id, code, loc, slot, x, y, quality, ilvl, sockets, qty, flags (V: read). NO prefix, suffix, rare ids, unique/set id, and no colour. `status.Item` has no field for them either.
- `item full` / `ground full` (vault verbs, `internal/d2dctl/vault_verbs.go VaultRecord`) DO carry `Quality`, `Prefix`, `Suffix`, `Rare1/Rare2`, `Affixes` (the six rare affix ids), `SetID`, `UniqueID` and the raw hex (V: struct and `docs/d2d-vault-verbs.md`; sample lines in the vault spike, `jew q=4 suf=285`). So a vault card could derive the colour bot-side from the tables, but the inventory path cannot, and a Go copy of the rules would duplicate fork logic and the separator-row indexing.
- Missing from both: the resolved colour and the base's InvTrans.

## 3. Smallest fork change

In the private fork (`apps/d2d/devctl_dump.hpp` rows, plus the `item full` / `ground full` view), behind `--item-verbs` / `--vault-verbs` (default off, like the existing additions), write two integers computed by the code the renderer already uses:

- `"inv_tint":T` = `item_pieces[code].inv_transform` when `tints_with(T)`, else 0.
- `"inv_colour":C` = `item_colours.of(quality, unique_id, set_id, prefix, suffix, affixes, class_affix, socket0code, true)`, or -1.

That is a ~10-line helper next to `Scene::item_map` (factor `item_map` so it returns the pair, then the dump prints it). Document it in the fork's control-channel doc; fork branch pushed to the private remote the same day; never upstream without the user's go. Needs the fork build (devp11). No change to tables or rules. Alternative rejected: expose only the affix ids (bigger payload and the bot re-implements `Colours::of` plus separator indexing).

## 4. Art pipeline and UI

Today `internal/art` decodes a DC6 with the ACT1 palette to a PNG on demand with a disk cache (`store.go`, key grammar `item-<invfile>`, V: read), from the user's own install.

Do the shift SERVER-side, per (base, tint, colour), lazily, not on the client and not a palette-indexed PNG:
- New key `item-<invfile>~t<T>c<C>` (extend `itemKey`; `ValidKey` stays strict; T 1..8, C 0..20). The store reads `data\global\items\palette\<set>.dat` from the install (same archive order), maps `Pix` through the 256 entries of colour C, then the normal `Image(pal)`. A missing set file or a bad index returns the untinted icon (never an error to the page).
- Cost: at most 21 colours x the bases actually owned, each a small PNG, cached on disk under the existing sha-keyed cache. No need to pre-render; the page only asks for what it shows.
- `status.Item` gains `tint string json:"tint,omitempty"` (e.g. `t6c12`) next to `art`; `cmd/d2bot/display.go itemArt` sets it from the dump fields; the web layer appends `~<tint>` to the art URL when present. Items without a tint are unchanged (cache and URLs stable).
- Legal rule kept: tinted PNGs are produced at runtime from the user's install into the cache; none are committed. Tests use synthetic DC6 and a synthetic 21x256 map built in the test.

## 5. Slices (each testable offline)

| # | Slice | Where | Offline test | Needs fork build |
|---|---|---|---|---|
| 1 | Fork: `inv_tint` / `inv_colour` in `dump items`, `item full`, `ground full`; fork doc | fork (private), builder | fork `tests/item_verbs_d2d.py` extension on the builder: rolled magic jewel prints colour >= 0, plain normal item -1; table print of jew InvTrans | yes (devp11) |
| 2 | Bot parse: `dumpItem` reads the two fields into `d2.Item`; `VaultRecord` parses them; `status.Item.Tint` | internal/d2dsrc, d2dctl, status, cmd/d2bot | fixture JSON lines in `internal/d2dsrc/testdata`, `Tint` string golden; missing fields = empty tint | no (fixture first) |
| 3 | Art: `item-<invfile>~tTcC` key, colormap reader, remap | internal/art | synthetic DC6 + synthetic map: golden PNG differs from untinted exactly at mapped pixels; bad T/C falls back; `ValidKey` table | no |
| 4 | UI: append `~tint` to the art URL, fixture `fx=phone` jewels with distinct tints, regression: no CSS colour on cells | internal/watch/web | extend `artembed_test.go`: N jewels with N tints request N distinct keys | no |
| 5 | Live check (checklist) | worker host | none | yes |

Order: 3 and 4 can land before the fork (tint empty = today's picture), slice 2 after a fixture from slice 1; user-visible only once 1 to 4 are in.

## 6. What the user should see afterwards

Magic and rare jewels, rings, amulets and coloured gear show the in-game tint (a jewel's egg takes its affix colour; unique jewels and set items their fixed colour); plain normal items stay grey/white. Two jewels with different suffix colours look different; the page does not invent any colour of its own.

Live checklist (not run, no game here): build devp11 with slice 1; `item full` on a hero holding a magic jewel, a rare jewel, a unique jewel and a plain jewel: colours differ and match the game window; `ground full` shows the same item with `inv_colour` equal to the carried one; compare one tinted armor piece with a screenshot of the real inventory.

## 7. Open points (U)

- The `jew` InvTrans value and the exact colour per jewel affix: print from the install in slice 1.
- Belt, merc and cursor items use the same `inv` rule (V compcode.md); the page's merc equipment will pick the tint up once `mk` gets the field.
- Ground pictures in the watch map (if ever drawn) use Transform/chrtransform: a separate `tint` pair, not part of this plan.
