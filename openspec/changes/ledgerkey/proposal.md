## Why

The cheat ledger is keyed by character name, so five copies of one save (or one name in two prefixes) share one identity and one cheat poisons all of them (docs/research/character-factory.md 4.5; Palli exists in two prefixes with a shared count). The baseline copies and the memrules 'writes after start' count need per-prefix identity.

## What Changes

- internal/cheat/ledger.go: optional `char_id`, `Filter.Key`, `Ledger.CheatedKey`, `Summary.ByKey`; bookkeeping kinds (setup-copy, runstart) never count as cheats.
- internal/cheat/ledgerkey.go: `Entry.Key()` (CharID, else instance/name, else name), `KindSetupCopy`, `CharKey`.
- cmd/d2cheat: `ledger --key INST/NAME --inst I`, key column.

## Impact

No spec delta; old ledger lines need no migration (they already carry instance and name; lines without an instance match by name, fail closed). Overlap (declared): card memrules (agent/memrules, unmerged) edits internal/cheat/ledger.go (adds `Run` after `Backup`, separate hunk) and cmd/d2cheat/main.go (`runstart`, `--since-run`, separate hunks); merge order is free, ledgerkey was built on main. Overlap (declared): card snaplib writes the setup-copy entries.
