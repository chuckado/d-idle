---
status: accepted
date: 2026-11-02
decision-makers: user
---
# 0014: Memory may be read and written; the bot may never depend on writes

## Context and Problem Statement

AGENTS.md said "memory is READ-ONLY" except logged test cheats. The lab now needs more than cheats: a lockstep clock and frame stepper, tracing hooks, diagnosis pokes and fixture capture. The user's words (2026-11-02): "We can read/edit memory whenever we want, but we have to make sure we aren't using memory editing as a crutch."

## Decision Outcome

Reading is always allowed. Writing is allowed for four purposes: SETUP, INSTRUMENTATION (observe or control time and stepping without changing how game state evolves), DIAGNOSIS and FIXTURE capture, always through the logged tool with a reason and a ledger entry.

The CRUTCH RULE: the bot's play never depends on memory writes. In a measured run (gate, soak, baseline, fidelity) there is no gameplay-affecting write after the start; setup writes come before and are declared. Otherwise the run is CHEATED, reported and never counted. Each soak or gate record carries `writes after start: N`; a feature is done only when its live drill passes with writes disabled. Injected code is INSTRUMENTATION only and needs the user's go per hook set.

## Consequences

* Enforcement: `d2cheat runstart` and `d2cheat ledger --since-run` (card memrules); the gate judge in the private repo should read the same count.
* Delta spec: lab-rules requirements "Memory access" and "Crutch rule" (card memrules).
* Unchanged: agents never launch or close games, never create or delete characters, own one instance, no Battle.net.
