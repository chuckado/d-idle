# Revive mercenary as a player command (card botmerc, API watch.merc/1)

> status: built offline, NOT verified live (2026-11-14). Design: `docs/ui/redesign-2026-11-14/C-orders.md` (`revive_merc`), brief B3. Tags: V verified by the offline tests named, S secondary, U unverified live.

## What it is
A request the player gives (button "Revive mercenary"): at the next safe moment the bot walks home and the EXISTING revive step runs there. Nothing about the policy changes: the gold-only gate of `docs/mercrevive.md` (price x `merc.reviveGoldMult` + the revive reserve, carried plus stash gold, hourly cap) is the single rule, so the button and the town step cannot disagree. The merc policy's thresholds, hire and gear are untouched (gear is not possible yet, `docs/research/merc-gear.md`).

## Priority and safety
Same ladder as the other player commands (`docs/inventory-actions.md`, "Priority ladder"): safety first (chicken, flee, crowd rules, a fight within the clear radius), then player commands, then chores. The request is looked at only where those commands are: `bot.playerKinds` (`ErrandsWaiting`) starts the errand trip at the repair trip's safe moment (`Runner.safeMoment`), `Pending` runs it in town. A safety action cancels the walk home like any command and the request stays queued. V: the generic `TestErrandTripGoesHomeAtASafeMoment`, `TestSafetyInterruptsAPlayerCommandInProgress` (internal/bot) run on the same hooks; the merc glue is `TestPlayerReviveDeadWithGoldGoesHomeAndRevives`.

## Request model (`internal/merc/playerrevive.go`)
- One request at a time. States `queued` -> `doing` -> `done` | `failed` (reason); `cancelled` only from `queued`. In memory only: a restart forgets it (it is cheap to ask again; U whether that should be persisted).
- The trip is wanted only while the policy would pay NOW (`PlayerRevive.Ready`): if the gold is spent on the way the request waits and no trip is made for nothing. V: `TestPlayerReviveGoldGoneBeforeTheTripMakesNoTrip`.
- In town the existing `reviveStep` decides and pays (quote probe, stash fetch, backoff included). Its outcome ends the request: revived or already alive -> `done`; refused for gold, cap or no reviver -> `failed` with the reason, NO cooldown; a failed attempt (hireling unreachable, price changed) or the step's own 2 minute backoff -> `failed` with the reason and a cooldown of `merc.PlayerReviveCooldown` (5 minutes, a constant, U which value suits the NPC walk). V: `TestPlayerReviveFailureKeepsTheReasonAndACooldown` (both classes).
- A fight or cancelled context in town puts the request back to `queued` (`Postpone`).

## Button rules (like the repair button)
`merc.PlayerRevive.View`: off with a reason when no merc is hired ("no mercenary is hired"), not read yet, alive ("the mercenary is alive"), level unknown, the hourly cap reached, gold short ("need N more gold", N = the policy's bar minus carried plus stash gold), the last revive failed ("try again in N min"), or one is running. While `queued` the button is "Cancel revive". V: `TestPlayerReviveGoldShortSaysHowMuch`, `TestPlayerReviveAliveNoneAndUnknownAreOff`, `TestPlayerReviveCapKeepsTheButtonOff`.

## API (`internal/watch/mercapi.go`)
- `GET /api/merc` (viewer): `{schema: "watch.merc/1", state: alive|dead|none, price, gold, need, button {action request|cancel|none, label, enabled, reason}, request {state, reason, at}}`. `none` also covers "not read yet".
- `POST /api/merc/revive {}` and `POST /api/merc/cancel {}` on the answer listener with the gates of the repair routes (origin list, token or passcode cookie, JSON, small body, unknown fields refused). A refusal is 409 with `code` and `error` (`state`, `queued`, `none`, `off`). 503 `off` until the desk is bound; with the merc policy off the adapter answers `state: none` and code `off`.
- Goldens: `internal/watch/testdata/merc_*_{Paladin,Sorceress}.json`. V: `TestMercRoutesGatesAndGoldens`.

## Files and interfaces others must know
`cmd/d2bot/mercplayer.go` (new): `mercGear` wraps the inventory desk as `bot.PlayerGear` (passes `NextKind`, `ErrandsWaiting`, `ErrandKind` through) and `mercAPIAdapter`. Mounted by `rig.Fun.Merc = mercAPI()` (`cmd/d2bot/ladder.go`), `watch.Fun.Merc` (`internal/watch/funwire.go`), wrapped in `playerGear` (`cmd/d2bot/invwire.go`). B5 (merc item verbs) reads the same merc state: it owns `cmd/d2bot/mercwire.go` lines other than the `pr`, `lastSnap`, `snapSrc` fields and `NeedsTrip`'s `noteSnap` line.

## Live checklist (nothing ran live)
1. Hero in the field, merc dead, enough gold (price x mult + reserve): the button reads "Revive mercenary", enabled. Press it: the bot walks home at a safe moment, the log shows `merc: revived`, `GET /api/merc` shows `state: alive`, request `done`.
2. Short of gold: the button is off with "need N more gold"; the log line `merc: revive skipped (gold) ... need=` matches N plus carried.
3. Alive merc or none hired: off with the reason.
4. Press, then cancel while the hero is still walking or fighting: request `cancelled`, no trip.
5. A fight on the way: the request waits and runs after; a chicken or a portal flee during the walk cancels the walk and the request stays queued.
6. Make the revive fail (stand where the hireling cannot be reached, or take the gold away during the dialog): request `failed` with the reason, the button off for 5 minutes with "try again in N min".
7. Restart the bot with a queued request: it is forgotten (button enabled again).
8. Hourly cap: after two paid revives the button is off with "the hourly revive limit is reached".
9. The quote: with the real quote read from the NPC menu the `need` in the reply equals the one in the log.
