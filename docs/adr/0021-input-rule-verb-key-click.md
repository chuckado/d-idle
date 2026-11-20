---
status: accepted
date: 2026-11-19
decision-makers: user (popup, Fable review 2026-11-19)
---
# 0021: Input rule: verb, then key, then verified click

Source: user popup, 2026-11-19, on the Fable design review ([06-roadmap](../design/fable-review-2026-11-19/06-roadmap.md) section 7). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md), which listed this question as open.

## Context

The hero can act on d2d through lab verbs, hotkeys or mouse clicks; clicks are the least reliable path. Review: [07-input-layer](../design/fable-review-2026-11-19/07-input-layer.md) (Q13).

## Decision

The fork's action verbs are the game's own client-to-server messages, one verb per message. The bot prefers a direct verb, then a hotkey, then a verified click. Item, gold and identify verbs move from the lab verb set to the input set after one differential each. A generated hotkey map, built from the planner slots, is written into the hero save.

## Consequences

- New actions are added as verbs first; a click path needs a stated reason.
- Each verb that moves from lab to input needs its own differential against the game's message.
- The hotkey map is data generated per hero, not hand-kept (build-flexibility rule in AGENTS.md).
