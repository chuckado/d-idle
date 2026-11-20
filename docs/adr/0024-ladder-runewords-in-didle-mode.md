---
status: accepted
date: 2026-11-19
decision-makers: user (popup, Fable review 2026-11-19)
---
# 0024: Ladder-only runewords enabled in d.idle mode

Source: user popup, 2026-11-19, on the Fable design review ([06-roadmap](../design/fable-review-2026-11-19/06-roadmap.md) section 7). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md), which listed this question as open.

## Context

Ladder-only runewords (Spirit, Insight and more) are part of normal play for the hero's builds but are not in the vanilla offline rules (Q4).

## Decision

They are enabled in d.idle mode as a visible rule in the fork. Gates and fidelity runs stay vanilla.

## Consequences

- The rule is a flag in the fork, shown to the player, off for gates, soaks and fidelity differentials.
- Build files may use these runewords only in d.idle mode; a gate scenario must not depend on them.
