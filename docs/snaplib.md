# Snapshot library (d2restore)

Cards: snaplib (tool) and ledgerkey (ledger identity). Design source: `docs/research/character-factory.md` section 4 (not merged when this was written). Rules: ADR 0014 (a restore is a logged SETUP action; a copy of an honest character is `copy-of-honest`, not cheated). Tags: V = verified (here: by unit test on synthetic files, or by the cited live finding), S = secondary, U = unverified. Nothing in this card was run against the game or a real save.

## What it is

`internal/saves/snaplib.go` (logic) and `cmd/d2restore` (CLI). A library is a directory `<lib>/<id>/` with the character's whole file set (`<Name>.d2s .key .map .ma0..9`, `.d2x` if present) and `manifest.json`. Library default: `$D2RESTORE_LIB` or `<data>/snaplib`. Separate from the rotating per-instance backups of `d2saves`.

Manifest v1 (V by test): id, name, class, clvl, xp, gold, stash gold, active difficulty and act, hardcore, expansion, map id, `kit_hash` (sha256 of the .d2s from the skill block on: skills, items, corpse; header and stats excluded), sha256 per file, date, bot commit, `ledger_status` (honest or cheated, read from the cheat ledger at snapshot time), source run, source instance. Stash: offline 1.14d keeps it inside the .d2s (V: no .d2x in any prefix listing, character-factory.md); a `.d2x` is copied when present (U).

## Commands

```
d2restore snapshot --name Fresha --inst i5 --at-menu [--id ID] [--run R] [--bot-commit H]
d2restore list [--json]
d2restore verify ID
d2restore place --snapshot ID --as Freshb --inst i4 --reason "baseline late run 3" [--run R] [--allow-cheated]
```
`--dir D` replaces the save directory of `--inst` (tests, odd layouts). `--ledger FILE` and `--lib DIR` override the defaults.

- `snapshot`: needs `--at-menu` (fail closed, as `d2saves`); refuses a corrupt .d2s, a file that changes while being read, an existing id. The ledger status is looked up by `<inst>/<name>`; without `--inst` any cheat of that name counts.
- `verify`: every file re-hashed against the manifest (no missing, no extra), .d2s parsed (signature, size, checksum, markers, stat block), header fields compared with the manifest. It does NOT prove the game loads it (live, below).
- `place`: copy into ANOTHER prefix under a NEW name (must differ from the source name, letters with at most one `-` or `_`, 2 to 15 characters; S: the game's own rule, the tool is stricter than the file format). Preflight, in order: `--reason` given; target is not i0; the ledger is writable; the prefix has no running game and every instance lock (`paths.Layout.GameLocks`, the files `tools/live.sh` and the launcher flock) can be taken without waiting, and stays held during the copy; the snapshot verifies; the source's ledger status is honest (else refused unless `--allow-cheated`, provenance `copy-of-cheated`); no file of the new name exists in the target (case-insensitive, never an overwrite). Files are staged under dot names, the .d2s is renamed last, then the placed .d2s is re-read and verified. A refusal after the arguments are valid is recorded in the ledger with `success:false`.
- Rename (V by test on synthetic files, offsets from `internal/cheat/d2s`, V on real files there): name field 16 bytes NUL padded at 0x14, size field unchanged, checksum at 0x0C recomputed with `d2s.Checksum` (rotate-left-1 plus byte, checksum bytes zeroed), result must pass `d2s.Parse`. Only name and checksum bytes differ (tested). That no other copy of the name exists in the file, and that `.key`, `.map`, `.ma*` need only the new file name, is S/U: a live check.

## Ledger line

`place` appends one `setup-copy` entry (kind `cheat.KindSetupCopy`) to the cheat ledger: actor, instance = the target prefix, character = the new name, reason, and params: snapshot, source name, provenance (`copy-of-honest`), sha256 of source and placed .d2s, file list, source ledger status, clvl, run. It prints `setup: copy-of-honest <id>` for the run record. `setup-copy` and the memrules `runstart` marker are bookkeeping kinds: `Ledger.Cheated` and `CheatedKey` ignore them, so a placed copy is honest. Placed AFTER a run-start marker it is still a successful write after start, so `WritesAfterStart` counts it and the run is CHEATED, as ADR 0014 says (copies before the marker, never after).

## Ledger identity (ledgerkey)

Entries are keyed `Entry.Key()`: `CharID` if set (optional field `char_id`), else `<instance>/<name>`, else the bare name for entries written before instances were recorded. Old lines need no migration: they already carry instance and character. `Filter.Key` and `Ledger.CheatedKey(since, "i4/Freshb")` match case-insensitively; a legacy entry without an instance still counts against the same name in every prefix (fail closed, it cannot be attributed). `Filter.Character` and `Ledger.Cheated(since, name)` keep their old name-only meaning. `Summarize` adds `ByKey`; `d2cheat ledger` takes `--key INST/NAME` and `--inst I` and prints the key. A cheat on `i5/Fresha` no longer poisons `i4/Freshb` (renamed copy) nor `i4/Fresha`.

Merge order with card memrules (`agent/memrules`, not merged when this was written): both touch `internal/cheat/ledger.go` (struct fields in separate hunks), `cmd/d2cheat/main.go` (`cmdLedger` cases in separate hunks) and `docs/cheats.md` is not touched here. The run id of a setup-copy goes in `params.run`, not in memrules' `Entry.Run`, so either order compiles; after both merge, `d2restore` may set `Entry.Run` too. `WritesAfterStart` scopes by the marker's `Instance`, which already separates prefixes; start the marker with the instance set (`d2cheat runstart` does, via `$D2BOT_INSTANCE`). The gate judge and `soak24.sh preflight` / `ledgercheck.sh` (private tools, not edited) still key by name: move them to `<inst>/<name>` (U, coordinator).

## Live preflight: what the first live restore must show (none done yet)

1. Pick a free hidden instance (never i0); the user's approval of file copies covers Fresha for the baseline runs only. Check the target prefix is stopped and its lock is free.
2. `d2restore snapshot --name Fresha --inst i5 --at-menu`, then `verify`; compare `list` with the character screen: clvl, class, difficulty.
3. `d2restore place ... --as Freshb --inst iN`; check the ledger line and that `Fresha` files in the source prefix are byte-identical to before.
4. Start the instance through the normal tooling: the character list must show Freshb with the right class and level, and the game must LOAD it (checksum and name rewrite accepted). If the list shows the old name or refuses the file, the rename is incomplete (S/U above): stop, report the bytes.
5. Enter the game: read name, level, gold through the normal readers; the bot continues (starts a soak step, moves, fights) with no "new character" behaviour. Save and Exit; `d2savecheck Freshb Sorceress` and `d2restore snapshot` of the copy must give the same kit hash and xp.
6. The `.map` seed question (U): the header map id (0xAB) equals the `.map` seed and is identical in every copy (V stable across loads, character-factory.md 4.6). Compare `.map` before and after the game (the id should stay), and, if cheap, a level dump (`d2mapdump`) of one Act I area in two copies: equal rectangles mean the layout is pinned and the copies are less independent. Decide on other difficulty entry or other source characters only after that.
7. Record in the run note: `start kit: copy-of-honest <id>`, `writes after start: 0`.

## Not done

Wiring the baseline launcher and `soak24.sh` to call `d2restore` (private tools); loads-and-continues smoke (live); `ledgercheck.sh` keying; hardcore and d2d placement (d2d saves live under XDG dirs, the same file set, not tested).
