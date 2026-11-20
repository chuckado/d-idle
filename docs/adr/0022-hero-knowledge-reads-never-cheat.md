---
status: accepted
date: 2026-11-19
decision-makers: user (popup, Fable review 2026-11-19)
---
# 0022: Hero knowledge: engine facts exposed as reads never make a run CHEATED

Source: user popup, 2026-11-19, on the Fable design review ([06-roadmap](../design/fable-review-2026-11-19/06-roadmap.md) section 7). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md), which listed this question as open.

## Context

What the hero may know about a level was open (Q5). Without level layout the hero explores blind, and quest-object rooms cost long searches.

## Decision

K1: the fork may expose the level room layout and the quest-object rooms on entry as a READ-only verb. General rule: any engine fact exposed as a read never makes a run CHEATED. Writes still follow ADR 0014 and the crutch rule.

## Consequences

- The crutch rule in AGENTS.md is unchanged for writes; reads of engine state are allowed in measured runs.
- The read verb is part of the fork and flag-gated like other verbs; it is recorded in the run record.
