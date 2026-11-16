# Survival slice S-D: bot.zoneModel (card survsd, 2026-11-16)

> status: current (TASK-552). Offline: nothing here ran against a live game. Tags: V verified (how), S secondary (a doc), U unverified. Design: [design-2026-11-14.md](design-2026-11-14.md) sections 5.3, 10, 12; gate: [ab-harness.md](ab-harness.md). The knob DEFAULT STAYS 0: the shipped default flips only after the A/B passes and the user says go.

## 1. What the knobs do

| Knob | Values | Meaning |
|---|---|---|
| `bot.zoneModel` | 0 (default), 1 | 0: the plan tables choose the leveling zone, exactly as before. The shadow lines (`zone shadow:`, `zone choice:`) keep running. 1: the survival controller (`survival.Ledger.Choose`) chooses among the plan's zones by margin and expected value. |
| `bot.zoneDial` | 0 cautious, 1 balanced (default), 2 bold | the enter and stay lines of the margin (`internal/power/profiles.json` dials). Read only with `bot.zoneModel` 1. The knob is a number because every knob is; the ids are the table `zoneDialIDs` in `internal/bot/zonemodel.go`. |

## 2. How the choice is made (V: code read, tests below)

- `PickOptions.Model` (`internal/bot/plan.go`) is the hook `Plan.Pick` asks first. It is shown the eligible steps (acts, kit, exclusions already applied); the plan tables stay the candidate list. The model names an area (`ok`), asks for a fresh game (`fresh`) or declines.
- Declining is the fallback, never a stall: a controller error (`zone model: error, the plan tables decide`, one line per distinct error), no candidate the profile covers, an area the plan did not offer, or no hook (no game tables) all leave the plan rule to decide.
- Candidates: the eligible steps plus the zones cleared in this game, marked `Depleted`. The controller's own rule decides when a depleted best zone means a fresh game (`FreshRatio`, `FreshMinutes` in `internal/survival/params.json`, U).
- The log line is the controller's `zone choice:` line (numbers allowed in the log); with the model on the shadow `zone choice:` is not repeated.
- Hysteresis: the controller's stay line, switch margin and ledger step-down apply, and the goal keeps a pick for the dwell `zoneRecheck` (3 min, design 11, U) before it asks again; the pick is also re-asked at once on a level change, a cleared or banned zone, as before.

## 3. Boss floors

The margin of the hero against a boss (`power.BossPack`, `Model.BossWord`: the boss alone as one melee monster, minions not in the pack, U) can replace the Andariel class floor, boss by boss; Andariel is the first and only boss wired (the other floors stay legacy until their data is shown; add a boss by calling `zoneBossFloor` for its key).

- Margin data exists (`danger.Table.FindBoss`, a readable hero): word "safe" makes the level floors met. "risky" or "too hard" keep the legacy floor and switch on the below-recommended warning (`AndarielWarn`): warn, never wall. The margin never raises a floor.
- No margin data, an explicit `boss.andarielMinLevel*` knob, or `bot.zoneModel` 0: the legacy floor, unchanged.

## 4. Merc and fresh games

- Merc dead policy: with the model on, `mercWanted` returns nothing for a "revive" while the latest verdict (younger than 12 min) says the merc is dead and `survival.Ledger.Revive` judged it not worth it (the zone kills him, the kit comes first, or he adds too little). Log: `merc: left dead by the zone model`. Hire is untouched. Reviving is itself off by default (`bot.mercReviveUnverified`), so today this only matters for builds that turn it on.
- Fresh game as a zone tool: when the verdict is `Fresh` (best zone cleared and a new game beats walking to a poorer one) and there is a Cycler and a cleared zone, `Pick` reports nothing eligible; the existing path (`freshGameToLevel`, then the one guarded `startFreshGame` with its in-flight and debounce guard, `freshAllowed` no-gain gap and the thrash cap) starts the game. No new call site.
- Farm loop (`docs/farmloop.md`): untouched. The controller's depleted notion is the planner's cleared set (`loop.cleared`); the farm reset still acts on coverage inside a zone.

## 5. Words only

`/api/survival` (the glance block) carries the margin word (safe, risky, too hard) and the why sentence, and the new field `model: true` when the block is the zone in effect instead of a shadow. `Runner.ZoneMargin()` gives the same word and sentence to the Status line. No number leaves the log.

## 6. Tests (V: run on this branch)

`internal/bot/zonemodel_test.go` (Sorceress and Paladin): model arm picks a lower zone than legacy; error, unknown area, knob off fall back; fresh verdict and depleted candidates; dwell; merc left dead and dial table; boss floor (safe lowers, risky and too hard warn, no data and knob keep legacy); knob defaults. `internal/survival/zonedial_test.go` (both builds): the dial changes the choice and never inverts; no zone flapping over a plateau. `internal/power/boss_test.go` (both builds): boss word ordering. `cmd/d2bot/powerwire_test.go`: no tables, no hooks.

## 7. A/B

`tools/survival/abrun.sh` refuses a binary without the string `bot.zoneModel`; the built `d2bot` now contains it (V: `strings` on a build of this branch). The run itself is the A/B card's job ([ab-harness.md](ab-harness.md) section 4 flip rule); nothing was run here. Not done: per-boss floors beyond Andariel; the `zone.why` in the Status line is words from the controller, the UI sentence of Status is the watch card's.
