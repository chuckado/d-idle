# Survival: user answers to the open questions (2026-11-14)

> status: current. Records the user's answers to docs/survival/design-2026-11-14.md section 12, given 2026-11-14 in the coordinator session. Binding for slices S-A to S-F. The proposed text for docs/decisions.md and the ADR amendment is the coordinator's to write (guard rule); this file is the source they cite.

| # | Question (design section 12) | Answer |
|---|---|---|
| 1 | Dial | Cautious, Balanced, Bold; default Balanced |
| 2 | Margin on the watch screen | Words only (safe, risky, too hard); never a number |
| 3 | Margin replaces boss clvl floors | Yes, boss by boss. Warn, never wall. The legacy floor stays visible until the margin has matched the logged outcomes (ADR 0011 amendment pending, coordinator) |
| 4 | Merc may stay dead when the zone kills him | Yes; the reason goes into the Status sentence |
| 5 | Fresh games as a zone tool | Allowed |
| 6 | Archetypes | Five |
| 7 | Overrides | Five per class, each with evidence and an expiry (enforced in internal/danger: `MaxPerClass`, evidence file check, expiry date) |
| 8 | Second live drill class | Sorceress |
| 9 | Readiness package | TASK-217 and TASK-218 fold into S-A and S-B; TASK-219 to TASK-223 are closed as superseded (coordinator edits those cards) |
| 10 | Release | The S-D engine ships default ON after its A/B passes (this replaces the design's default-off proposal) |

Consequences for S-A: the `/api/danger` endpoint carries numbers because it is for the engine and for developers; the watch UI never shows them (item 2).
