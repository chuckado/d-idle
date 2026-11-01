# Game-code guard, labpath class, path scrubber

Why: the repo holds notes about the game (address tags, offsets, short formulas), never pasted game code or data. A one-off audit (2026-11-01, card auditfix) found nothing pasted; these checks keep it so.

## tools/ci/gamecode_guard.sh (runs inside `tools/ci/check.sh`, also `--hygiene-only`)
Scans every tracked or untracked-not-ignored text file (not `testdata/`, images, `tools/ci/gamecode_*`). Fails on:
- a run of 6 or more lines that start with an x86 mnemonic (mov push pop call jmp cmp lea test xor add sub jcc) with a register or `0x` operand, optionally after an address and hex bytes;
- a Ghidra pseudo-C block: 4 or more lines with `undefined1..8`, `uVarN`, `iVarN`, `param_N`, `local_N` (or short non-table lines with `LAB_`/`DAT_`), gaps of up to 2 lines allowed;
- 8 or more lines of a hex dump (`00401000  55 8b ec 83 ...`);
- a tab-separated header with 4 or more of: skill, charclass, lvlmana, minac, maxac, MonStats, Level, Treasure Class;
- a tracked file under `docs/` over 256 KB that is not in `tools/ci/gamecode_allow.txt` (path, then a `#` reason).
Allowed: one-line backtick instruction quotes, `FUN_xxxxxxxx` tags, tables that name `LAB_`/`DAT_` addresses in prose.
Tune patterns in the script, never exempt a file just to pass. Counter-test: `tools/ci/gamecode_test.sh` (fake 7 and 10 line listings, address+bytes listing, pseudo-C, hex dump and table header must fail; inline quotes and tags pass); it runs in `tools/ci/scripts.sh`. `tools/ci/gamecode_guard.sh FILE...` checks given files.

## leakcheck class labpath
`cmd/leakcheck` flags `~/d2d-lab`, `~/labrun`, `~/d2lod-wt`, `~/d2lod-ops` under `docs/probes/` only (other docs name `~/d2d-lab` as a documented working directory). The existing `homepath` class already covers `/home/<name>` and `/Users/`. Tests: `go test ./cmd/leakcheck` (`TestLabPathScopedToProbes`).

## tools/ci/scrub_paths.py
`tools/ci/scrub_paths.py [--check] [ROOT]` (default `docs/probes`) rewrites in place: `~/d2d-lab` to `<lab>`, `~/labrun/<host>` to `<labrun>`, `~/d2lod-wt` to `<wt>`, `~/d2lod-ops` to `<ops>`, `/home/<name>/` to `<home>/`, and prints a count per file. Only that text changes, so values and line counts stay; files only shrink. Run it on every new drill or soak log before committing (leakcheck fails otherwise). First run (2026-11-01): 801 paths in 105 files, 100 files under `docs/probes/drills/` and 5 in navsoak, navsoak2, navsoak3 and soakfull.
