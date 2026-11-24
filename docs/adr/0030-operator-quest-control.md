---
status: accepted
date: 2026-11-23
decision-makers: user
---
# 0030: Operator quest control without the viewer passcode (local socket, quest verbs only)

Source: user approval, 2026-11-23, of the operator quest control built in card opsquest (docs/operator.md; code `internal/ops`, `cmd/d2bot/ops*.go`).

## Context

Whoever runs the container (the maintainer or an agent) needs to start, retry, cancel and hand in quests, and switch auto, overnight, without the viewer passcode and without opening a network route to the viewer.

## Decision

The container operator may press the quest board verbs without the viewer passcode over a local socket, with these binding guard rails:

1. Off by default: the socket exists only with `D2BOT_OPERATOR=1` in the environment (env only, no knob, no flag).
2. Never in a measured run: with `D2BOT_RUN` set no socket is created and a later request is refused (`measured_run`).
3. A unix socket in the instance state directory, mode 0600, never on the network, not exposed through the viewer, a port or the answer listener.
4. Quest verbs only: `start`, `retry`, `cancel`, `handin`, `auto`, `status`. The allowlist is in code (`internal/ops/ops.go`) and tested; every other verb is refused.
5. The same `questboard.Desk` functions and refusals as the viewer's POST routes (busy, locked, not_ready, paused).
6. Audited: every verb the desk ran is a line in the hero's quest audit ring with `remote:"operator"`; a required reason (`--why`, 1 to 200 characters) is recorded.
7. Rate limited in the bot: 6 mutating verbs per quest per hour, 30 per hour for the whole board, and a 60 s cooldown after a `retry` for `retry` and `start` of that quest (`cancel` still works).

## Amended 2026-11-24: owner try-anyway start

Source: user decision 2026-11-24 (popup): the owner may start ONE not-ready quest chunk deliberately, and the board does not keep the hero from it.

- Owner only: the operator verb `start --try-anyway` (the socket, with `--why`) or the viewer's `Try anyway (risk)` button (same passcode and origin rules as Start, and a second tap that states the risk, `confirm`). No other verb, no auto-start, no chain.
- Audited: the audit line has its own path (`operator/try-anyway`, or the viewer's `/api/quests/<id>/try`), the stated reason and the readiness refusal it skipped (`override: try anyway ... (refused: ...)`); the bot logs `override: try anyway`.
- One shot: it covers one chunk run. A park, a death, a finish, a stop or a restart ends it; the chunk returns to the hold it had. It counts against the same rate limits (6 per quest per hour).
- Readiness words only: it skips the board word `too hard for now`, the level and life floors (Barracks, road, tower, Andariel and the other boss gates, the boss script floor) and the unattended Andariel hold. It never skips the kit gates (potions, portal, mana, antidotes), the death-spiral guards, or any combat reaction: drink, retreat, chicken, town portal and the death handling run as always.
- Never in a measured run (`D2BOT_RUN` set refuses it, as every write). A death in a try is counted as a death only; the readiness ledger is not struck further.

## Amended 2026-11-24 (user decision): agents may use the owner try-anyway start in the sandbox

Source: user decision 2026-11-24, on the pipeline audit row 16 ([pipeline-audit-2026-11-24.md](../process/pipeline-audit-2026-11-24.md)); amends the "Owner only" line of the amendment above.

- An agent running the sandbox hero may press `start --try-anyway` (with `--why`) through the operator socket. The viewer's button stays the owner's.
- Nothing else changes: audited with its own path and the skipped readiness refusal, one shot (ends at a park, death, finish, stop or restart), readiness words and floors only (never the kit gates, the death-spiral guards or combat reactions), rate limits unchanged (6 per quest per hour, 30 per board per hour), never in a measured run.
- The agent states in its report which quest it tried and why; no chain of try-anyway starts and no use outside the sandbox container.

## Consequences

- No new power: anyone with `docker exec` on the host already owns the container, its state and its save.
- Limits live in process memory and reset at a bot restart; the viewer's history shows operator lines (only the audit tells them apart).
- Threat model and usage: docs/operator.md; overnight driver `tools/questloop.sh`.
