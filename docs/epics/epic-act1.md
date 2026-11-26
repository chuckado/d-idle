# epic-act1: Act I complete in Normal

> status: OPEN epic, written 2026-11-25. Tags: V verified (named run or file), S secondary, U unverified. Lab setups are CHEATED (stand-in level, gear, words): mechanics only, never evidence for the live hero.

Card: `epic-act1` (backlog/tasks). Parent bar: ADR 0016 T4 Questor and the Done-means bullet in docs/decisions.md (Current truth). Numbers are not copied here; follow the links.

## Quests (V data/questboard/act1.json)

| id (data/questboard/act1.json) | quest | kind | chunks | boss |
|---|---|---|---|---|
| `den_of_evil` | Den of Evil | optional | clear_den | - |
| `sisters_burial_grounds` | Sisters' Burial Grounds | optional | blood_raven | blood_raven |
| `search_for_cain` | The Search for Cain | optional | scroll, read, stones, free_cain, report | - |
| `forgotten_tower` | The Forgotten Tower | optional | countess | countess |
| `tools_of_the_trade` | Tools of the Trade | optional | malus | the_smith |
| `sisters_to_the_slaughter` | Sisters to the Slaughter | story | andariel | andariel |
| `act1_done` | Sail east to Act 2 | unlock | sail_east | - |

## Status (S, from the verdict files)

Drill verdicts today: `drill.act1.*` 1/1 GREEN, `step.act1.*` 22/23 GREEN (data/stepgrounds/verdicts/, lab only, every setup CHEATED).
Andariel drills (`drill.andariel.*`): 4 of 6 GREEN.

Acceptance 1 (quests defined with steps): V for the rows above, the board carries chunks and step lists for each; known gap list is in the Blockers section.

## Blockers (the real current REDs)

- Smith stand click (`step.act1.smith.stand_click_pickup` RED, matrix note 68)
- Tools of the Trade reward word 0x204e: the Charsi reward talk does not complete (`drill.tools.merc17` RED, `necro.act1-tools` RED, matrix note 70)
- Andariel live margin: a clvl 21 hero (life 280, merc 17) is refused as too hard; kill needs life 600 or the bold dial (`drill.andariel.live21*`, matrix note 69)

## Evidence pages

- docs/probes/quest-drill-matrix-2026-11-23.md
- data/stepgrounds/verdicts/ (drill.act1*.json, step.act1.json)
- docs/act1check.md, docs/boss.md, docs/act-plan.md

## Not yet proven (U)

Live fresh-hero run of this act with writes after start 0; readiness floors per leg; 3 h economy run; second class. Update this page when a card lands, do not hand-count here.
