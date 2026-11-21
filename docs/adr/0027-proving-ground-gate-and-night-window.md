---
status: accepted
date: 2026-11-21
decision-makers: user (popups, Fable review 2026-11-21, Proving Ground and gate)
---
# 0027: Proving Ground gate: assertion digest, tiers, quotas, ledger, quarantine; standing night window

Source: user popups, 2026-11-21, on [proving-ground-gate.md](../design/fable-review-2026-11-21/proving-ground-gate.md) section 8 (Q1 to Q10) and [the catalogue](../design/fable-review-2026-11-21/proving-ground-gate-catalog.md). Builds on [ADR 0020](0020-proving-ground-and-stat-skill-agency.md) (Proving Ground first as merge gate) and [ADR 0023](0023-fidelity-by-rule-differentials.md).

## Context

Two runs of one seed differ for reasons that are ours (wall-clock tapper against a virtual clock); the merge gate has a long queueing tail; verdicts are scattered over probe docs.

## Decision

1. The gating bar is the verdict plus an ASSERTION digest (only the facts a scenario asserts). The full state digest is a drift report only (Q1).
2. Tiers T0 to T4. T0 and T1 (no game) block every merge. The T2 smoke set is REPORT-ONLY now and blocks after 7 clean days; the user decides then (Q2).
3. Drop `-count=1` from the merge gate after a 20-merge measurement; an uncached whole-repo `-race` run runs nightly (Q3).
4. STANDING NIGHT ALLOWANCE on the second worker host: 23:00 to 07:00 EST daily, at most 3 runs, niced, hidden instances only, the sandbox container untouched, for T3/T4 and the nightly race run. This window is the ONLY time the second worker host may be used; outside it it is not allowed (the earlier one-off ended 2026-11-20 09:00 EST) (Q4). Heavy work otherwise runs on the second worker host.
5. Agent quotas on the builder: 1 running plus 1 waiting job in the main lane, 2 in the long lane, no whole-repo run from a branch (Q5).
6. The results ledger lives on the board host, with one generated status doc in the repo (Q6).
7. Auto-quarantine: a verdict flip on unchanged binaries quarantines a scenario at once; a person releases it, never puts it in (Q7).
8. Release needs at least one EARNED act run; CHEATED scenarios are excluded from release notes, from the release after card 16 (Q8).
9. The long lane has two slots behind the memory guard (Q9).
10. Probe docs stop carrying verdict tables once the results page exists; a doc cites a run id (Q10).
11. Cards: 16, determinism proof and results ledger first, pause after each group.

## Consequences

- A scenario gates only when `stable`; setup tags EARNED, SETUP, CHEATED, IMPOSSIBLE are stored per run.
- Night runs use no more than 3 runs and must leave the sandbox container alone.
- Any other use of that host needs a fresh user go.
