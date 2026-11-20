---
status: accepted (Q4, Q5, Q7, Q13 since decided in 0021 to 0024)
date: 2026-11-19
decision-makers: user (popup answers on the Fable design review, 2026-11-19)
---
# 0020: Proving Ground first; the player spends every point; intent arbiter; class order

Amends [ADR 0018](0018-fun-loop-expedition-ledger-2026-11-07.md) item 6 (skills unlock by quest or find, one free-choice point) on who spends skill points. Builds on [ADR 0016](0016-automation-ladder.md) and [ADR 0017](0017-direction-fun-first-2026-11-07.md) without changing their text. Does NOT amend [ADR 0015](0015-goals-acceptance-on-game-exe.md): its A:B acceptance framing stands until Q7 is decided.

Source: the Fable design review, branch `agent/fablereview` (docs/design/fable-review-2026-11-19/, mergeq #967): [00-summary](../design/fable-review-2026-11-19/00-summary.md), [04-stat-skill-agency](../design/fable-review-2026-11-19/04-stat-skill-agency.md), [05-architecture-and-overhaul-options](../design/fable-review-2026-11-19/05-architecture-and-overhaul-options.md), [06-roadmap](../design/fable-review-2026-11-19/06-roadmap.md) section 7 (questions Q1 to Q13). Review claims are tagged V/S/U there; this ADR records only the user's choices.

## Context

The project builds far faster than it proves: much merged "offline, NOT verified live", the proven ground is Normal Act 1, and a quest chain broke in seven ways on first contact. Stat and skill points are spent from guide tables, not by the player. `runGoal` is an ordered chain with five separate loop guards patched on after incidents.

## Decision

1. **Proving Ground first (Q1).** Build scripted lockstep quest scenarios on d2d as a merge gate (roadmap R1, R2). The bot-change freeze is lifted ONLY for work that comes with a green Proving Ground scenario. New classes and new acts wait for it.
2. **Stat and skill agency, path D (Q2, Q3, Q10, Q11).** The player spends every stat and skill point. The hero derives its play from the skills it actually has. A generated capability table labels each skill supported, experimental or unsupported. Guide tracks are shareable build files. When the player is away the hero banks the points and asks; autopilot (follow a guide track) is a per-hero switch. Respec is a player event: the hero proposes it, the player decides. Unsupported skills are allowed with a clear label and one confirmation for attack skills.
3. **Arbiter and unit of work (Q6, Q9, Q12).** Replace `runGoal`'s chain by an intent arbiter, shadow first, one family at a time. Cards are vertical ("the hero does X on d2d, scenario green"), even when larger than a 400-line guide. The review's prototypes (arbiter, questspec checker, freebuild planner, build file) become real cards. No rewrite.
4. **Class order (Q8).** Sorceress second build, Necromancer Summoner, Hammerdin, Barbarian, ground-target builds with the Druid, Assassin, Amazon. Each waits for item 1.

## Consequences

- ADR 0018 item 6 is amended: skill points are the player's, not unlocked one at a time by quest or find; the "one free-choice point" is superseded. Whether the Codex keeps skill discovery as a pacing layer is NOT decided (second half of Q2): keep it out of cards until decided.
- Bot changes stop merging without a green scenario; the first card is the scenario runner with the stuck tree step as its first red line.
- Cards follow roadmap sections 2 and 3; agency P0 (stats to the player, per-level skill cap legality) is first.
- Build files and the capability table become data, per the build-flexibility rule in AGENTS.md.

## Open (not decided; do not treat as settled)

- Q4: ladder-only runewords (Spirit, Insight and more) in d.idle mode, or vanilla.
- Q5: what the hero may know (level room layout read verb, "reads never make a run CHEATED").
- Q7: fidelity acceptance (A:B verdict vs per-rule differentials). ADR 0015 acceptance stands until then.
- Q13: input rule (verbs, then keys, then verified clicks).
- Codex skill-unlock pacing layer (second half of Q2).
