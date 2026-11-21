---
status: accepted
date: 2026-11-21
decision-makers: user (popups, Fable review 2026-11-21, combat engine)
---
# 0026: Combat engine: one capability table, generated tiers, delivery grammar, reads a player can see

Source: user popups, 2026-11-21, on [combat-engine.md](../design/fable-review-2026-11-21/combat-engine.md) section 6 (Q1 to Q11); appendices [A](../design/fable-review-2026-11-21/combat-engine-a-today.md), [B](../design/fable-review-2026-11-21/combat-engine-b-grammar.md), [C](../design/fable-review-2026-11-21/combat-engine-c-tests.md), [D](../design/fable-review-2026-11-21/combat-engine-d-cards.md). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md) (capability table, no freeze lift without a green scenario), [ADR 0021](0021-input-rule-verb-key-click.md) (input rule) and [ADR 0022](0022-hero-knowledge-reads-never-cheat.md) (reads). Fidelity per [ADR 0023](0023-fidelity-by-rule-differentials.md).

## Context

The engine is one loop for "attack one monster"; summons, curses and traps live in class-switched side loops with skill names in Go; three tables classify the same skills and disagree. The capability tier is hand-typed and gates nothing.

## Decision

1. `data/capabilities/capabilities.json` is the ONLY hand-edited skill table. `skills.json` and `skillKinds` are generated from it, then removed (Q1).
2. Tiers are generated: unsupported (no valid delivery or no driver), experimental (valid and a driver exists), supported (plus evidence from a green scenario row). Some rows drop from supported to experimental until their scenario exists (Q2).
3. A fused `cmd skill` verb (select and cast in one) is allowed as input under ADR 0021 only if a differential shows it equals select-then-cast (Q3).
4. Flag-gated READ verbs `minions`, `skillstate` and `combat events` are added to the fork; reads fall under ADR 0022 (Q4).
5. The hero may ACT only on reads a player can see (cooldown shade, minion count, curse icon). Hidden reads are test-only. Each read row is marked `visible` or `hidden` (Q5).
6. Experimental skills are used in the rotation behind the expectation breaker; unsupported skills never (Q6).
7. The upkeep budget is per-archetype data; defaults: 1 cast in 4, mana reserve 10 percent (Q7).
8. Blessed Hammer is a self-centred `insidePack` stance, after one live look (Q8).
9. Whirlwind and shapeshift forms are parked until fidelity items A6 to A8 have verdicts (Q9).
10. No new pixel work: a verb path first (Q10).
11. Foundation cards, which change no behaviour, may use a golden parity test over the simulator and scenario corpus as their gate (Q11).
12. Delivery: 16 cards, FOUNDATION first, pause after each group for the user's look.

## Consequences

- No skill id, name or class check in combat Go code (build-flexibility rule); the engine branches on grammar enums.
- Read verbs never make a run CHEATED; hidden reads may be used only in tests and scenario assertions.
- The old tables and the class switch are deleted in the cleanup card.
