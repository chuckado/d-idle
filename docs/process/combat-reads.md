# Combat read verbs (dlv-reads, TASK-721)

> status: current 2026-11-21. Fork side (private forge, branch didle/combat-reads-20261121, tip 81b191a94) built and tested headless on the builder, ticks per second measured (no difference with the journal on); bot side is the typed client and fake only, no bot consumer yet (the consumers are later cards: dlv-upkeep-minions, dlv-expect follow-ups, the scenario assertion kinds). Tags: V = verified (how is said), S = read in code or docs, U = unverified. Decision: ADR 0026 Q4 (the three reads are flag-gated and are reads under ADR 0022: never CHEATED) and Q5 (the hero acts only on `visible` reads).

## What the fork exposes

Three READ verbs of the private d2d fork, gated by the d2d flag `--read-combat` (`--lab-verbs` implies it). Without the flag they answer `err read combat off: debug <verb> needs d2d --read-combat`; with it they are listed in `caps` as `read.minions`, `read.skillstate`, `read.combat-events`. Each answers ONE JSON line, then `ok`. They write no game state (V: the fork test compares the unit list before and after repeated reads). The journal behind `combatlog` records only while the flag is on (V: the fork source; one relaxed atomic load per site when off); frame cost: 6000 lockstep steps with 12 fallens, 6 zombies and a Clay Golem took 0.80 to 1.03 s on the fork without the journal and 0.73 to 1.11 s with it (builder under load, 4 runs each, V: measured), no measurable difference.

Go client: `internal/d2dctl/combat_reads.go` (`Client.Minions`, `SkillState`, `CombatLog(n)`; types carry `Visible` and `Hidden` marks; every failure wraps `ErrCombatRead` and names `debug <verb>`), the fake in `internal/d2dctl/fake` (`Fixtures.Minions|SkillState|CombatLog`), tests in `combat_reads_test.go` (a Necromancer's army, Paladin and Sorceress skill states, a Sorceress kill).

## Fields and marks

| verb | field | visible or hidden | why |
|---|---|---|---|
| `debug minions` | id, name, owner, skill, life, maxlife, x, y, mode, here | visible | pets stand on screen with a life bar; the class tells the skill |
| | ttl (ticks of 40 ms left of a timed summon, -1 never) | hidden | the screen shows no timer |
| `debug skillstate` | id, name, base, level, mana, cooldown, usable, left, right, charges | visible | skill page and tooltip, the icon greyed out, the bound buttons, charge counters |
| | status (engine refusal code: 0 ok, 1 mana, 2 weapon, 3 no row, 4 form, 5 passive, 6 aura, 7 no level, 8 delay) | hidden | a code, not a thing on screen (the voice line and the grey icon are what a player sees) |
| `debug combatlog` | kind, source, target, skill, by_hero | visible | a swing, a death, a block animation, who did it |
| | amount, detail (damage, mana paid, life left, blows, refusal code, missile row) | hidden | exact numbers are not on screen (exact monster resist and damage numbers: tests only) |

Event kinds: `cast` (amount mana paid, detail level), `refuse` (amount engine status), `hit` (amount whole damage the unit took, detail its life left), `miss`, `block`, `kill` (detail 1: the hero's credit), `hurt` (the hero took `amount` in `detail` blows), `missile_hit` (amount Missiles.txt id, detail missile level). The ring holds 1024 events; `combatlog [n]` returns the last n (default 64) oldest first; `total` counts all events since the start, so a scenario takes a bookmark (`Total`) and later reads `Since(bookmark)`.

## How to use in a scenario

1. Decide with visible fields only (Q5): cooldown, usable, mana, minion count and life, kind and source of events. A bot rule that reads a hidden field is a defect; the type's `IsHidden(name)` is the check a consumer can assert on.
2. Verify with anything: a scenario assertion may use hidden fields (damage amounts, status codes) because it judges the engine, not the hero.
3. Assert on counts, not life deltas: `l.Count("kill")`, the `cast` events of a skill id, the minions' `Skill` and `Owner`. Take `Total` before the action and read `Since` after.
4. Always feature-detect: `HasCap(caps, CapReadMinions)` etc.; a fork without the flag answers an error naming the verb, never an empty list.

## Not done (follow-ups)

- Corpses and states in `monsters` (appendix C section 4 of the combat engine review): not part of this card.
- Scenario runner assertion kinds and `tools/provingground` consumers: later cards.
- A live differential of the journal against game.exe.
