# Per-zone ledger and the shadow controller (survival slice S-C)

> status: current (card survC). Implements design-2026-11-14.md sections 5.3 to 5.6, 8 and 10 for slice S-C: log and API only, NO behaviour change. Tags: V verified (how), S secondary, I inferred, U unverified. Nothing here ran against a live game or a recorded real log.

## What shipped

- `internal/survival` (data in `params.json`, every constant tagged U): `Ledger` (per (area, difficulty) rows with exponential decay over field time, half-life 20 min; pooled `k_hero`), `KZone`, `KHero`, hard events (chicken x1.5, merc death x1.5, death x2.5, relaxing with time away), the power-change rule (a row learned at another power, more than 10 percent off, is dropped), `ForcedOut` (the dial's "observed cost that forces a step down"), `MarkLeft` / `Locked` (a zone left for cost stays closed until the hero's power moves by more than 10 percent), `Choose` (design 5.3: candidates, eligibility with the enter line for a new zone and the stay line for the current one, score `xpPerMin + lambda x netGoldPerMin` amortised over the stay, switching margin 15 percent, depleted zone -> next safe one or a fresh game, empty set -> first field plus "needs you"), `Revive` (design 5.5), `Economy` (xp and net gold per minute), `GlanceOf` (words only).
- Persistence: its own versioned file, `ledger.json` in the instance directory next to `deaths.json` (`paths.Instance.Ledger`, `store.File`, version 1). Additive: a bot without this package never reads it; a missing file is an empty ledger; unknown fields load; a file from a newer build is refused and never overwritten (the wire then starts empty and does not save). The brief said "in the state file": a separate file keeps the old bot's state files untouched, which is the same promise.
- Bot hooks (passive): `Deps.ZoneChoice` (called after the `zone shadow:` line, logs `zone choice:` and, with a dead merc, `zone revive:`), `Deps.ZoneObserve` (area summary, chicken, death, merc alive-to-dead edge) and the fighter's `OnSummary` chain (`survFight`) for the per-fight damage that `k_zone` and `k_hero` compare with the model. Wiring in `cmd/d2bot/survivalwire.go`, both backends.
- API: `GET /api/survival` (`internal/watch/survivalapi.go`, `Hub.SetSurvivalSource`): `{"enabled":false}` until the bot reports a choice, then `{"enabled":true,"survival":{schema,zone{id,name,margin,dial,why{code,text}},alternatives[],merc?,fresh?,needsYou?}}`. WORDS ONLY (user answer 2): `margin` is `safe`, `risky` or `too hard`; a test walks the JSON and fails on any number outside an `id` and on a digit in any `text`. Codes are the closed list in `internal/survival/why.go`.

## Interface notes for the UI slices (not edited here)

- `/api/glance` does not exist on this base (slice S2 `uis2` owns `internal/watch/glance.go`). The brief S2 and S3 list no survival fields. Proposal: S2 embeds the block of `/api/survival` as `survival` (same shape) and the Status sentence reads `survival.zone.why.text`; "Why not X" is `survival.alternatives[]`; the dead-merc reason is `survival.merc.text`. The design's example used `zone.margin` as a number and `xpPerMin`; both are dropped here by user answer 2. Until S2 lands the UI reads `/api/survival` behind its feature flag. Mismatch to settle with S2: the design names the block `zone`, S2's reply shape has no such key.

## Deliberate limits (shadow)

- Candidates are the plan's steps of the difficulty (no quest or waypoint reachability, no navigation faults): I.
- Depleted = the planner's `cleared` mark. The fresh-game rule: the best eligible zone is depleted and the next safe zone is worth less than `freshRatio` (0.5, U) of it, or none is safe: a fresh game on the same zone. The user's answer allows fresh games as a zone tool; "the next is above the line" is read as "the next safe zone".
- Revive quote is the constant 400 (U; swap7 paid 360). Potion price 50, travel 2 min, gold per kill from the plain-kill table of internal/data: all U.
- Taken per kill: `FightSummary.Damage` is the sum of life drops between samples (healing not netted out; V in dodge.go), predicted taken per kill is the margin model's typical pack. The ratio is therefore biased up whenever potions heal in a fight (U); `k_zone` is clamped to 0.5 to 8.
- The hero power scalar for the 10 percent rule is max life, times 1.25 with a living merc (I, U). It ignores the belt on purpose.

## Tests

`go test -p 1 ./internal/survival ./internal/watch ./cmd/d2bot ./internal/bot -run 'Both|NoSecond|Dial|KZone|Depleted|Hole|Revive|Ledger|Glance|Unknown|Survival|Zone'`. Two builds (Paladin on `bruiser`, Sorceress on `caster`): the choice ratchets up with level and never picks Black Marsh at clvl 3; after two merc deaths in Black Marsh the hero leaves and does not re-enter at 1 minute, 45 minutes and 6 hours at unchanged power, and may when its life grew by 30 percent; dial step-down per dial; `k_zone` and `k_hero` do not double-count; revive worth it at 1 gold, refused with a reason at millions and when the kit reserve is short, worth less after merc deaths, closed after two in the window; ledger save, reload, old document, unknown fields, newer version; glance is words only.

## Replay of real logs (NOT run: the logs are not in the repository)

`git ls-files | grep -iE 'swap7|hero14|h4'` is empty. On the host that holds them (S-B report): `go run ./tools/levelsim/cmd/levelsim -shadow swap7.log -install <game dir> -arch bruiser` gives the margin per minute (S-B). The controller replay is the same stream fed to `survival.Ledger`: for each `zone entry` call `Choose` with the candidate margins of that minute, for each `merc: died` / `character died` / chicken `safety decision` call `Observe`. A `-controller` mode for `levelsim` is the S-D prerequisite; until then the synthetic test `TestNoSecondBlackMarshEntry` is the stand-in. Pass rule (design 10): swap7 shows no second Black Marsh entry and one revive (the second quote refused or the zone closed), hero14 never produces an empty candidate set. One sandbox hour in shadow compared minute by minute: grep `zone choice:` against `planner` in the bot log.

## What blocks S-D

1. Real-log replay and the sandbox hour (above) have not run; every number in `params.json` is U.
2. Live: `ledger.json` appears in the instance directory and `/api/survival` answers on a hidden instance (checklist: start a hidden slot, wait for one `planner` line, `grep 'zone choice:' bot.log`, `curl .../api/survival`).
3. S-B limits stay (defence, block and spell damage unverified, merc constants).
4. Gold side: real revive quote and potion price are not read (constants).
5. `levelsim -controller` mode and the abmerc A/B harness for the knob `bot.zoneModel`.
