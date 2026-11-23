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

## Consequences

- No new power: anyone with `docker exec` on the host already owns the container, its state and its save.
- Limits live in process memory and reset at a bot restart; the viewer's history shows operator lines (only the audit tells them apart).
- Threat model and usage: docs/operator.md; overnight driver `tools/questloop.sh`.
