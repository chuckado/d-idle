# Model A/B on decomposition item SKRESIST: Opus vs Sonnet vs Haiku (card modelab)

> status: card modelab, 2026-11-01, docs only (no game, no emulator run, nothing pushed or posted). Compares three finished results for decomp row 21 (SKRESIST, monster resists and immunities by difficulty). The two experiment arms (docs and tools) stay on their own branches and are NOT merged; this file is the only thing that lands.

Tags: **V** a claim I checked by reading the arm documents against each other and against the independent verifier's record (not by running anything: no new emulator runs); **R** the arm's own reported evidence (a run it says it did; not re-run here); **I** my inference; **U** unverified. Reference for "game.exe facts": the Opus trace `sk-resist-trace.md` plus the independent verifier `pr-curseimmune-verifier.md` (own oracle, own harness, verdict READY: 0 of 616,160 compared cases differ after the fix, 5 case classes). That reference is itself not infallible; where it is the only source I say so.

Arms: **Opus** (reference; `sk-resist-trace.md`, `tools/emu/sk_resist.py`, `sk_resist_test.cpp`, then `pr-curseimmune-draft.md` / `-verifier.md`; became upstream PR #160). **Sonnet** (blind; `sk-resist-trace-ab-sonnet.md`, `ab_resist*`, branch `agent/skresistab`). **Haiku** (blind; `sk-resist-trace-ab-haiku.md`, `abh_resist_*`, branch `agent/skresistab-haiku`). All three started from the commit before the Opus work (base bbbfddd; the Opus merge parent).

## 1. Verdict in one table

| | Opus | Sonnet | Haiku |
|---|---|---|---|
| Rules matched to Opus (of Q1..Q10) | (reference) | 9 same, 1 extra (Q11 data range, equal to Opus Q2) | 5 same, 1 partial, 4 missing |
| The curse / aura 1/5 rule (the real bug) | found (Q6, D1) | **found** (Q8, D1), with the "own list" nuance | **missed** (tagged U, not run) |
| d2d side | REAL build (`missile_blow` real; curse adj copied line for line) | REAL `missile_blow` and `boss_stats`; curse delta copied; player chain copied | a Python port of d2d's rule only |
| d2d deviations found | D1 (+D1b), D2, D3 (melee pierce/floor, read), D4 | D1, D2, D3 (physical whole points, extra), D4 / D5 (game side), D6 equal | none proven (D1..D5 all "no deviation" or U) |
| Headline counts | D1: 88,101 / 147,628 reduction, 14,355 damage | D1: 3,943 / 5,922 non-physical damage cases | 0 of 127,108, 0 of 2,688, 0 of 10,500 (against a port / the tables) |
| Wrong claims vs reference | none found; 3 LOW text slips caught by the verifier | 1 (Fire / Lightning Mastery as pierce, in B2), 1 doubtful (D3 size) | 0 outright; one misleading headline ("no deviation") |
| Blindness kept | n/a | not logged in the doc (U) | yes, with a candid log; 2 name-level glimpses |
| Effort (ledger) | 218,048 tokens, 68 tools, 17.5 min | **351,087**, 112 tools (the brief said about 85), 26.7 min | 243,901, 96 tools, 15.6 min |
| Trust as PR-candidate basis after an independent verifier | yes (it was; READY) | yes, with caveats (section 6) | no |

## 2. Rule by rule (Opus Q1..Q10)

"same" = same conclusion, equivalent evidence level or better; "diff" = different; "miss" = absent.

| Opus rule | Sonnet | Haiku | Note |
|---|---|---|---|
| Q1 base = MonStats column of difficulty at spawn (FUN_00573cb0), hireling reads Normal | same (Q1, Q2) plus V run of the loader: 13,212 values equal d2d's loaded table (R). Opus had V read only, so Sonnet is stronger here | same for the column; **miss** the hireling-Normal and the clamp; V run 1,750 rows, 10,500 values (R), against the tables not d2d | Haiku's count covers rows with Level > 0 only (582 / 584 / 584), Sonnet's all 734 rows (R) |
| Q2 which rows are immune per difficulty | same (Q11; counts identical to Opus Q2 for rows >= 100: Hell 20 / 11 / 113 / 105 / 137 / 131) | partial: only "618 values at 100 or more" in total | |
| Q3 boss / champion / unique / minion adds | same and stronger: V run 100,000 cases vs real `boss_stats`, 0 differ; minions run with flag 0 so no resist adds (Opus: S only) | S only, "U for this arm" | Sonnet's six-stat immunity count and "no add once two exist" agree with Opus's two-immunity limit |
| Q4 hit-time resistance: pierce below 100, no monster penalty/cap, floor -100 | same (Q3, Q4, Q5, Q7), V run 100,000 cases | same (Q2, Q3, Q4 player branch), V run 127,108 + 2,688 (R) | all three agree on pierce skipped at >= 100 for a plain monster, floor -100 on all |
| Q5 damage = MulDiv, flat reduction first, >= 100 gives 0 | same (Q6), adds that absorb is not run (U) | partial: S only, "not re-run" | |
| Q6 curses and auras divided by 5 on a base >= 100 (FUN_005c3540) | **same** (Q8: stat set 36 / 37 / 39 / 41 / 43 / 45, truncation toward 0, players and hirelings whole; V run, all 64 stat ids) | **miss** (Q5: Lower Resist, Conviction "U") | the decisive rule |
| Q7 curse values (Lower Resist 91, Conviction 123, Amplify, Decrepify) | partial: names the skills and formulas (R7: `-dm56`, `-min(ln34,150)`), no L1 = -31 figure | miss (U) | |
| Q8 masteries: Cold Mastery = pierce 335; Fire / Lightning Mastery are damage %, not pierce | **diff** on Fire / Lightning (B2 says stats 333 / 334 are Fire / Lightning Mastery's); same on Cold Mastery (20 + 5 per level) | miss (U) | see section 5, E1 |
| Q9 what an immune hit still does (no get-hit, no chill) | partial: cold length cut (V run), get-hit consequence U (agrees with Opus's U) | miss | |
| Q10 resist penalty only for non-monsters: 0 / -40 / -100 | same (Q3), plus classic -20 / -50 and penalty-before-cap order | same (Q4), same order | |

Count (Q1..Q10, partial = half): Sonnet about 9 of 10, Haiku about 5 of 10 (I).

## 3. Deviations: found / missed / extra against the Opus set

Opus: D1 curse cut missing; D1b first-stat-0 curse not applied (read only, unreachable); D2 no -100 floor on magic / physical; D3 melee elements without pierce / floor (read, no harness); D4 Sanctuary undead DR (read). Label caution: **D3 means different things in the three documents** (Opus: melee pierce / floor; Sonnet: physical whole points; Haiku: pierce callers, open).

| Opus deviation | Sonnet | Haiku |
|---|---|---|
| D1 curse / aura 1/5 on an immune | found, real build (R: 5,922 non-physical cases, 3,943 differ in damage or length, 0 after the delta is cut first; 6,966 all-element cases differ in the resist seen) | missed |
| D1b first aurastat 0 means no curse | missed (not mentioned) | missed |
| D2 no -100 floor, magic / physical | found (R: 3,890 synthetic cases), correctly labelled latent | missed (its port covers the four elements only) |
| D3 melee pierce / floor (read, S on d2d side) | missed (says `player_blow` not diffed, U) | partial: names the question as open (U), no answer |
| D4 Sanctuary undead DR | missed | partial: mentions the special path, not run |
| extra: whole-point physical roll (Sonnet D3) | see section 4 | n/a |
| extra: hireling chain and player magic resist (Sonnet D4, D5; game side only) | not in Opus; the verifier's class 5 (V run, game side) agrees hirelings and players take the full value | n/a |

Counts differ because the case generators and sizes differ (Opus 147,628 grid + random across all six types; verifier 616,160 over real skills and levels; Sonnet 100,000 per chain; Haiku 127,108 against a port). **No direction disagrees**: wherever Opus and Sonnet both measured D1 and D2, both say d2d deviates and both give the same trigger ("base >= 100 and a negative change" for D1; "total under -100 on magic / physical" for D2). The verifier's class counts (82,257 / 85,357 immunity verdicts differ on base, 0 on the fixed head) are a third independent confirmation of D1's direction (R, reference).

## 4. Extras the arms found that Opus did not

- **Sonnet D3, physical rolls cut to whole points before the resist (R: 3,480 of 6,437 physical cases after D1; 3,559 of the residual in the everything bucket).** Same effect or new? **Same effect class as `sk-melee2-trace.md` D1** (d2d keeps physical in whole points where game.exe keeps 256ths; that card measured it on the player's blow `player_blow`, 80,250 of 80,557 weapon cases differ in 256ths, +0.498 points a landed blow, V run on the real build). **New location**: Sonnet measured it on the missile path (`missile_blow` rolls physical, shifts to whole points, then `resisted()`). Opus saw the same seam and skipped it on purpose (physical and poison print "-" in its harness: "whole-point / per-tick paths, not this claim"). So it is real in kind (I, consistent with a V run on a second code path), new for missiles, and unverified by a second harness (the verifier did not test physical `dealt`). Open tension (U): melee2's table notes `Blow::frac` is carried "only from missiles" (combat.hpp:507), which would mean missiles already keep a fraction; Sonnet says the physical missile roll is whole-point. Settle by reading `missile_blow` before any fix. Reach: only physical missile skills (not the Sorceress plan).
- **Sonnet bot B4, `DoctorLine` caps the raw resist before the penalty** (game.exe adds the penalty to the raw sum and caps after; Hell expansion raw 100 gives 0, not -25). Consistent with Opus Q10's order (penalty then cap, V read); a real bot finding Opus did not list (I; the bot code line is the arm's S, not re-read).
- **Sonnet B3, six-stat immunity count in `FUN_005a1370`, nothing at all added once two exist** (more precise than Opus B3's "skips adds per third immunity; counts magic": compatible, Sonnet states the exact rule).
- **Sonnet, several curses stacking:** the cut tests the unit's OWN list (FUN_006253b0 -> list +0x24), so two curses on a base-100 monster are each cut. The verifier's stacked class (3,024 cases, 0 differ after the fix) is the independent support (R).
- **Haiku B1, the bot's gear score caps resist at 75 while game.exe caps min(75 + max-resist stat, 95)** for a player (I; impact undervalues +max-resist gear above 75). Opus did not look at the player side of the gear score. A modest, plausible extra (not verified).
- **Haiku Q4 / Sonnet Q4:** the player branch (penalty, cap, floor) verified by both against d2d's panel expression (Sonnet 9,890 cases, Haiku 2,688, 0 differ; both are comparisons with a copied or ported expression, not the real function).

## 5. Wrong or doubtful claims (against the reference)

| id | arm | claim | against reference | who is right |
|---|---|---|---|---|
| E1 | Sonnet B2 | Fire / Lightning Mastery carry pierce stats 333 / 334, so the bot under-rates fire / lightning too | Opus Q8 (via `sk-passives-trace.md` Q3 / Q4): Fire / Lightning Mastery are damage percent, only Cold Mastery gives pierce (335); the item stats 333 / 334 / 336 are separate | Opus likely right (S, not V read by me); Sonnet's own V run only proves the pierce rule, not which skill sets 333 / 334 (U). Treat Sonnet's B2 text as wrong for fire / lightning until checked |
| E2 | Haiku headline | "no proven d2d deviation", D2 / D5 "no deviation" | D1 is real on the build (Opus V run, Sonnet R, verifier) | Haiku's evidence is true for what it compared (element missile rule vs a transcription of that rule) but the port cannot contain `apply_target_stats`, so it was blind to D1 by construction. The doc says so in its Limits; the headline does not |
| E3 | Haiku Q5 | the 1.14d patch stack was approximated with the installed MPQs (no installer) | Opus / Sonnet used the patch installer path | doubtful (I): zero base mismatches argues for equal tables, which Haiku states as inference; Sonnet's 13,212 values vs d2d is the stronger check |
| E4 | Sonnet D1 evidence | real build | the curse delta is copied (`target.res += value`), not the real `Fight::target_of`; Opus's `adj` also follows d2d line for line | both weaker than the verifier, which ran the real `Fight::target_of` with real skills (616,160 cases). Direction of D1 unaffected (V, reference) |
| E5 | Sonnet D3 size | 3,480 of 6,437 | one path, one harness, not re-run | U until a second harness (section 4) |
| E6 | Brief vs ledger | Sonnet "about 85 tool calls" | ledger row: 112 tool uses | ledger is right (used above) |

Opus slips the verifier caught (for symmetry, V from `pr-curseimmune-verifier.md` section 6): first-stat "-25 -> -5" should be -31 -> -6 (L1); "ResDm min 10 gives -90" should be min 0 giving -100 (equal to the floor); pet defenders not mentioned. All LOW, text only.

## 6. Tags, evidence, blindness

- **Opus:** V read with addresses, V run on the real build, numbers reproduced by an independent oracle (the verifier did not open the author's harness). Strongest chain. Weakness: Q1 and Q3 only V read / S; melee D3 read-only.
- **Sonnet:** V read + V run throughout, real `missile_blow` / `boss_stats` / loader tables, explicit "copied, not the real function" notes, and a clean split of what is U (absorb, poison, hireling d2d side). Honest tags. Weakness: E1 (an unchecked inference tagged inside a V run bullet), no blindness log (I can only say its base commit is the pre-Opus one and it cites a different d2d pin, 7d5b3a5c, than Opus's 72132b99, so it did its own clone work).
- **Haiku:** candid limits, a "V port" tag that tells the reader the d2d side is a transcription, and a blindness log. Its tags are accurate; its conclusions simply cover less. Blindness: it saw the NAMES of two other agents' directories and another arm's worktree branch name (a directory whose name contains the resist-fix label); contents not opened. That is a mild leak (the name of a fix hints that a deviation exists) and did not stop it from reporting "no deviation"; I rate blindness kept in practice, with an asterisk. Sonnet's doc claims blindness but gives no log (U).
- **Templates and how much they helped.** The shared tree (base bbbfddd) already held the SKMELEE, SKMISS, SKMANA, SKSTATIC and SKPASS tooling, including `sk_missile_hit.py` / `sk_missile_hit_test.cpp` (a REAL `missile_blow` differential), `hit_rule.py`, `boss_level.py`, `static_field.py`, `sk_mana.py`, `sk_melee.py`, `sk_pass_tbl.py` and `monster_hit.py` (but no SKRESIST file and not sk_melee2 / the PR proofs). Haiku's log names `boss_level.py` as its harness pattern, `hit_rule.py` as template, and monstats / life_apply headers; Sonnet's test cpp mirrors the `sk_missile_hit` shape (inference from file naming and its header; the doc does not say so). The Opus run was the first on the item and had the same templates, so the scaffolds did not favour the later arms over Opus; what helped later arms was **knowing the scaffold works**, plus the resist-adjacent docs they were allowed to read (monster-mods-trace, sk-passives-trace, sorc-skill-matrix row 8). The matrix's row 8 text names Cold Mastery and Lower Resist; this may have steered Sonnet toward Q8 but Haiku had the same text and still missed it.

## 7. Effort

| arm | tokens | tool uses | wall | findings per 100k tokens (I, crude: "real build deviations proven") |
|---|---|---|---|---|
| Opus | 218,048 | 68 | 17.5 min | D1 + D2 proven = 0.9 |
| Sonnet | 351,087 | 112 | 26.7 min | D1 + D2 + D3 = 0.9 |
| Haiku | 243,901 | 96 | 15.6 min | 0 |

Haiku spent more tokens than Opus and found no deviation; the cheap model was not cheaper here (the ledger counts tokens, not price, so the rate card is not applied; I did not compute cost). Sonnet cost about 60 percent more tokens than Opus for a comparable deviation set. The follow-on Opus-line steps are separate rows: SKRESISTFIX (Sonnet, 170,651 tokens, 13.7 min) built the fix candidate and an Opus verifier rated it READY.

## 8. Per-arm trust verdict (as a basis for a PR candidate, followed by an independent verifier)

- **Opus:** yes. Already proven that way (PR #160).
- **Sonnet:** yes with conditions. D1 is correct, specific and reproducible, and the harness is close to what the PR candidate used; the fix would have been scoped right. Conditions: drop or re-check B2's Fire / Lightning claim; rerun D1 with the real `Fight::target_of` (as the verifier did) before calling it proven; keep D3 out of the PR until a second harness and the `Blow::frac` question are settled. It also correctly refused to ship a candidate itself ("harness is this card's own, no independent verifier ran").
- **Haiku:** no. Its output would have produced no PR candidate, and the "no deviation" headline would have closed the item wrongly. Useful as a verified spec of the element path (pierce under 100, floor -100, player cap) and as a table-equality check; not as a deviation hunt.
- **What each missed that mattered:** Haiku missed the curse rule, which is the whole real bug (D1; 82,000+ immunity verdicts differ). Sonnet missed nothing decisive: it missed D1b, D3-melee and D4 (all unreachable, read-only or Paladin-only in Opus). Opus missed (and chose to skip) the physical whole-point path, and Q1 / Q3 V runs that Sonnet did.

## 8b. Head-to-head: Sonnet vs Haiku (n = 1 item, one run each)

| | Sonnet | Haiku |
|---|---|---|
| Rules (Q1..Q10 vs the Opus set) | about 9 of 10 | about 5 of 10 |
| Curse / aura 1/5 rule | found (Q8), V read + V run, all 64 stat ids | missed (U) |
| d2d side | real `missile_blow`, real `boss_stats`, loaded tables; curse delta and player chain copied | Python port of d2d's element rule only |
| Deviations | D1 (3,943 of 5,922), D2 (3,890 synthetic, latent), D3 whole-point physical (3,480 of 6,437, to re-check), D4 / D5 game side | none; D3 open (U) |
| Equality results | boss adds 100,000, loader 13,212 values, player chain 9,890, all 0 differ | base 10,500 values, element rule 127,108, player 2,688, all 0 differ (against tables / a port) |
| Errors | B2 Fire / Lightning Mastery as pierce (likely wrong); no blindness log | misleading "no deviation" headline; no outright false rule |
| Effort (ledger) | 351,087 tokens, 112 tool calls, 26.7 min | 243,901 tokens, 96 tool calls, 15.6 min |
| Per useful finding (I, crude: deviations proven on a real build) | about 117k tokens and 9 min each (3 findings) | no finding to divide by; 244k tokens and 15.6 min for a clean spec of the element path |

**Haiku did well:** about 30 percent fewer tokens and 42 percent less wall time; accurate, exhaustive, byte-compared runs of what it was pointed at (base column read R1, effective resist R2/R3, player branch R4: 0 differ in every grid); candid limits and a blindness log. Its element-path facts agree with Opus and Sonnet on every point it covers (pierce skipped at >= 100, floor -100, no monster penalty, player cap order).

**Sonnet did that Haiku did not:** asked what else writes the monster's resist stat (the curse filter, Q8), built against the real d2d code so deviations carry real-build counts, found D1 / D2 / D3, ran the boss-mod adds against the real function, and refused to claim a PR candidate without an independent verifier.

**Could Haiku's port-only approach ever have found the real deviation?** No, not by construction (I). A port is a transcription of the rule the author already believes; D1 lives in `apply_target_stats`, a function the port did not include, and the game-side oracle never ran the curse filter either. A port can only confirm that the transcribed part equals game.exe. A real-build harness, or a port made from a differential on the whole path, is the minimum; and the brief has to make someone ask the curse question.

**Recommendation (honest, one item):** Haiku is fit for mechanical, fully specified sub-tasks: run an existing differential or oracle with given inputs, dump tables, grid-run a named game.exe function against a given expectation, bot-side grep findings. It is NOT fit for choosing which game.exe functions matter, deciding what a "full" comparison must cover, or concluding that no deviation exists. Those need Sonnet (first-pass decomposition) with Opus as the independent verifier. One run per arm: a second Haiku run might have found the curse rule; this is evidence, not a rate.

## 9. What this says about which model for which decomposition task

1. **Open-ended "find what differs" tracing needs at least Sonnet.** The deciding step was noticing that curses go through a separate filter function before they reach the stat list. Sonnet got there; Haiku read the right functions (FUN_0057be00, FUN_00573cb0) and verified them exhaustively but never asked what else writes the stat. That is a coverage-of-question failure, not a carefulness failure: Haiku's evidence on what it did cover is accurate.
2. **A ported oracle cannot find a deviation outside its port.** Haiku's design (game.exe vs a Python transcription of d2d's missile rule) guaranteed "0 differ" for the part of d2d it transcribed. The real-build harness is what makes a deviation count meaningful; any model should be told to use it. This is a brief and template issue as much as a model one: the earlier tooling pattern (the real-build test cpp) was in the tree.
3. **Sonnet is a credible replacement for Opus as the first-pass decomposer on this class of item** (matched the rules, found the main bug, added real extras, honest tags), with Opus kept as the independent verifier. This matches the current agent configuration (Sonnet decomp, Opus verifier).
4. **Haiku fits the mechanical sub-steps** (exhaustive table equality, grid runs, bot-side grep findings), not the question-setting. Its candour about limits makes its output safe to read but not safe to rely on for a conclusion of absence.
5. **Effort is not monotone in model size.** The weaker model used more tokens than Opus; Sonnet's blind run used the most. Cost per verified finding favours the model that already knows which functions matter; blind runs pay a discovery cost the Opus run paid first (and then later arms still paid, because blindness removed it).

Honest limits: **n = 1 item, one run per arm**; variance between runs of the same model is unknown (a second Haiku run might find the curse rule; a second Sonnet run might miss it). The Opus arm ran with a different brief and scaffold (first, no blindness constraint, a hand-built harness, the later Sonnet fix and verifier on top), so its numbers are not a clean control. The Opus arm's evidence is also partly what I used as reference, which biases toward agreement with it; where I could, I leaned on the independent verifier instead. The arms had different pins of the upstream source (72132b99, 7d5b3a5c, 173b79cc), so a differing line number is not a disagreement. No new emulator or build runs were made for this comparison: every count above is quoted from an arm document (R) or the reference (V), none re-measured (U for re-measurement). Token and tool counts are from the usage ledger; the ledger row for the Sonnet arm shows 112 tool uses where the brief quoted about 85.

## 10. Proposed shared-file lines

See `openspec/changes/modelab/merge-lines.txt`. Proposed glossary term (not written): "blind arm": a decomposition re-run from a tree before the reference result, for a model comparison.
