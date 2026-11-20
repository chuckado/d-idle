# Act 1 inventory and the next two scenario parts (card act1next, 2026-11-20)

> status: current. Tags: V verified (named run or test), S secondary, U unverified. Every scenario setup is CHEATED (see each `provenance`): mechanics evidence only, never balance or play evidence.

## 1. Inventory (S: read from `data/questboard/act1.json`, `internal/task/act1.go`, `data/questspec/act1.json`, `tools/provingground/scenarios`)

No chunk carries `unsupported` today (V: grep over `data/questboard`), so the board shows "not supported yet" for none; the real gaps are listed in the last column.

| quest | chunks | driven by board | scenario | gap |
|---|---|---|---|---|
| Den of Evil | clear_den | driven | act1.den.clear (this card) | none known |
| Sisters' Burial Grounds | blood_raven | driven | none | boss scenario; the reference hero already holds the title |
| Tools of the Trade | malus | driven | none | the Imbue item pick is `imbue:best` (docs/actbot.md: chooser NOT done); campaign skips imbue below clvl 20 or without a merc: out of scope |
| Search for Cain | scroll, read, stones, free_cain, report | driven | cain.* (9 scenarios, cain.chain green, #993) | none |
| Forgotten Tower | countess | driven | act1.tower.countess (this card) | runes are loot, not asserted |
| Sisters to the Slaughter | andariel | driven | none | clvl and potion gates; boss |
| Act 1 done | sail_east | driven | none | Warriv option name U |
