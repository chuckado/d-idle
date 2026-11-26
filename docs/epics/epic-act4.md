# epic-act4: Act IV complete in Normal

> status: OPEN epic, written 2026-11-25. Tags: V verified (named run or file), S secondary, U unverified. Lab setups are CHEATED (stand-in level, gear, words): mechanics only, never evidence for the live hero.

Card: `epic-act4` (backlog/tasks). Parent bar: ADR 0016 T4 Questor and the Done-means bullet in docs/decisions.md (Current truth). Numbers are not copied here; follow the links.

## Quests (V data/questboard/act4.json)

| id (data/questboard/act4.json) | quest | kind | chunks | boss |
|---|---|---|---|---|
| `izual` | The Fallen Angel | optional | izual | izual |
| `hellforge` | Hell's Forge | optional | hammer, smash | hephasto |
| `terrors_end` | Terror's End | story | seals, diablo | diablo |
| `act4_done` | Take Tyrael's portal to Act 5 | unlock | portal | - |

## Status (S, from the verdict files)

Drill verdicts today: `drill.act4.*` 0/2 GREEN, `step.act4.*` 0/5 GREEN (data/stepgrounds/verdicts/, lab only, every setup CHEATED).

Acceptance 1 (quests defined with steps): V for the rows above, the board carries chunks and step lists for each; known gap list is in the Blockers section.

## Blockers (the real current REDs)

- Izual reward bit not set (`drill.act4.izual` RED, word 0x000c)
- Hellforge: no travel or Hammer step ran (`drill.act4.hellforge` RED)
- Seals: the stand-in dies to De Seis, Infector and Vizier (kit too weak; `step.act4.terrors_end.seals.*` RED)
- River of Flame and Chaos Sanctuary travel RED (the hero dies on the way)

## Evidence pages

- docs/probes/quest-drill-matrix-2026-11-23.md
- data/stepgrounds/verdicts/ (drill.act4*.json, step.act4.json)
- docs/act4.md

## Not yet proven (U)

Live fresh-hero run of this act with writes after start 0; readiness floors per leg; 3 h economy run; second class. Update this page when a card lands, do not hand-count here.
