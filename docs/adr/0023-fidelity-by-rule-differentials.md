---
status: accepted, amends 0015 (acceptance)
date: 2026-11-19
decision-makers: user (popup, Fable review 2026-11-19)
---
# 0023: Fidelity acceptance: per-rule differentials replace the A:B verdict; no play on Game.exe

Source: user popup, 2026-11-19, on the Fable design review ([06-roadmap](../design/fable-review-2026-11-19/06-roadmap.md) section 7). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md), which listed this question as open.

## Context

ADR 0015 accepts d2d by a statistical A:B comparison (G-D: 6 game.exe lanes against 6 d2d lanes, XP/h and kills/h at 20 percent). Runs are noisy and seeds cannot be matched (Q7).

## Decision

Fidelity is accepted per rule: a zero-difference differential for each rule, or a listed shim in docs/d2d-deviations.md. This replaces the statistical A:B verdict of ADR 0015. We stop planning to PLAY on the Game.exe backend; Game.exe stays the oracle for differentials only.

## Consequences

- ADR 0015 header is marked amended by this ADR; its Gate A text about the A:B verdict no longer applies.
- Fidelity work becomes a list of rules with a differential or a shim each, instead of one aggregate verdict.
- Gate A keeps the 24 h unattended Act 1 run on d2d; only its A:B part is dropped.
