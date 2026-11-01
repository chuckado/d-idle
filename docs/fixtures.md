# Recorded fixtures (.snap, .d2mi)

Five small binary files under `testdata/` are recordings of the running game by our own tooling. They contain no game files, no Blizzard code and no data tables.

| File | Size | What it is |
|---|---|---|
| `internal/d2/testdata/town.snap` | 24 KB | `D2SN` pages: gzip of selected memory pages of the game process, throwaway offline character "Testy", level 1, standing in the Rogue Encampment |
| `internal/d2/testdata/town_inv.snap` | 24 KB | same character with items in the inventory |
| `internal/questmem/testdata/quests-i5.snap` | 9 KB | pages for the quest-flag reader, throwaway character "Multi" in town |
| `internal/d2map/testdata/rogue_encampment.d2mi` | 18 KB | gzip map image: the collision grid and room list of the one town level, read through our map reader |
| `internal/d2map/testdata/rogue_encampment_b.d2mi` | 14 KB | the same town, second recording (`D2MI1B` header) |

## What is inside (V: inspected 2026-11-01 by gunzip and a string scan, `strings`-style, of every file)

- The snaps decompress to 172 KB (town) and 49 KB (quests): raw process pages (heap, structures, pointers, tile cache bytes). Printable strings: the character name, the PE type-info name `.?AUTALLOCREC@@`, the install path of `Game.exe` (a default Windows path), a few internal archive path names of tile files (`DATA\GLOBAL\TILES\...dt1`), and three-letter item codes (`hth`, `stf`, `tbk`, `hp1`). No text blob, no table rows, no code.
- The `.d2mi` files decompress to 167 KB and 153 KB of binary cells and room records with a 6-byte tag; no strings besides the tag.
- Nothing here is a verbatim game table or a game file. The recordings are state of one running game, and they are small on purpose (the CI guard caps docs files at 256 KB; `tools/ci/check.sh` rejects `.mpq`, `.d2s`, `.key`).
- U: raw memory pages may carry tiny runtime fragments derived from game data (the tile cache bytes). We read them as incidental state, not as a copy of a file; if the owner of the game objects to even that, the five files move to the private ops repo and the tests skip when absent (history would still hold them).

## Regenerate

Needs a live offline game with a throwaway character (lab only, see `AGENTS.md` and `docs/replay.md`):

- `.snap`: `cmd/d2snap record=<path>[,key]` (the legacy writer) or `cmd/d2rec record` (envelope format, `docs/replay.md`), with `D2BOT_INSTANCE` set.
- `.d2mi`: made once from the map reader of `internal/d2map` (`docs/map.md`, fixture note at line 42); U: no checked-in command regenerates it, the recording tool was an ad-hoc run.
- Envelope format and loader: `docs/replay.md`, `internal/replay`, `internal/mem/envelope.go`.

Offline, the tests only read these files; no step needs the game.
