# Why the zone model never picks the Tower Cellars (toolstrade39, 2026-11-24)

Author: toolstrade39 (internal agent)

V live `zone choice` lines (read-only docker logs, dial bold, trial 19:40 to 21:42):
`zone choice: Dark Wood margin 17.8 (need 1.2; low 12.7) k 1.50 xp/min 32 dial bold ... | not Tower Cellar Level 4: margin_low, margin 1.4 (low 1.0, need 1.2, k 3.00) xp/min 94 | not Tower Cellar Level 1: margin_low, margin 1.1 (low 0.8, need 1.2, k 3.00) xp/min 100 | not Tower Cellar Level 2: margin_low, margin 1.6 (low 1.2, need 1.2, k 3.00) xp/min 56 | ... Level 3: margin 1.1 (low 0.8) k 3.00 xp/min 61`.

- Not an explore/exploit problem: eligibility looks only at the margin (low / k against the dial's need); xp (model or realized) only orders the zones that are already eligible. Cellar is in the candidate list, no exclusion applies.
- Cause (V from the lines, arithmetic): k = k_zone x k_hero. The hero k is at its clamp 3 (Tamoe k 6.62 = 2.2 x 3, Dark Wood 1.50 = 0.5 x 3). A zone without kills has k_zone 1, so it carries the whole 3; the visited fields have k_zone down to 0.5. The raw cellar margins are 4.2/3.0/3.3/3.3 (low margin 3.0 and up, against need 1.2); divided by 3 they fall under the need, and the risk route is closed (k 3.00 is at or above riskMaxK 2.9).
- The hero k is a kill-time ratio (U whether it should scale damage); the damage actually taken per prediction in the visited zones is half of it (k 1.5).
- Fix: dial row flag `poolUnvisitedK` (params.json, bold only): a zone without kills takes the kill-weighted k of the zones with kills when that is lower. Self-limiting: after the first visit the zone has its own k, its chickens (x1.5) and deaths (x2.5) raise it, and the dial's death row (bold: 1 death) forces it out. Cellar low margins would read about 2.0 at k 1.5 (U: live).
- Risk: the hero (about 280 life) chickened 7 times in 107 min in the k 1.5 fields; the cellars are harder than the model's 1.5 may say. Trial: set the bold dial, watch the first cellar visit.
