# Lane policy on the builder: quotas, merge priority, two long slots (card mglanes, TASK-749)

> status: current, rolled out 2026-11-21 (warn-only for one day, then enforced). Design: `docs/design/fable-review-2026-11-21/proving-ground-gate.md` section 4.5; owner decisions Q5 and Q9 of ADR 0027. The code lives in the private ops repo (builder lane wrapper and the builder status tool); this page describes it by role, with the numbers an agent needs.

## What an agent sees

| rule | value | what happens over it |
|---|---|---|
| main lane quota | at most 1 running and 1 waiting job per agent | the launcher exits 75 and prints the queue; do offline work and retry |
| long lane quota | at most 2 jobs per agent (running plus waiting) | exit 75 |
| whole-repo `go vet ./...` or `go test ./...` from an agent branch | refused on the main lane | exit 75 with a pointer: the merge runs it, test your packages (`labrun BUILDER --gate` is the one sanctioned whole-repo job from a branch) |
| wait budget | an agent job queued longer than 20 minutes is cancelled | exit 3 |
| merge priority | a merge job on the main lane goes ahead of every queued non-merge job; it never stops a running job | an agent job simply starts later |
| load guard | a merge gate does not start while the builder 1-minute load is above 3 times the container CPU count, for at most 10 minutes; then it runs and is marked contended | a `contended` mark in the lane state, a line in the job log |
| long lane slots | two jobs side by side behind the memory guard; the second slot starts only with 4500 MB available while the first is busy (the first still needs 2500 MB) | the job waits, then exit 5 after 30 minutes without memory |

Who is who: the agent is the name of the worktree under `d2lod-wt/` (the builder mirror keeps the path), the merge class is the merge worktree or `LABRUN_CLASS=merge`; every other directory is class other (no quota, no refusal, still yields to a queued merge). A merge also holds priority for 2 minutes after it ends, so the second builder call of one code merge cannot be overtaken.

Override for the coordinator: put `BLDR_ALLOW_WHOLE=1` in front of the command text (`labrun BUILDER -- sh -c 'BLDR_ALLOW_WHOLE=1 go test ./...'`). Agents do not use it.

## Rollout and reversal

- Warn-only first: until the epoch in the builder's lane policy file (`ENFORCE_AFTER`, set to 24 hours after the install) every refusal is only written as a `would-refuse` line to the lane policy log and the job runs. Merge priority and the two long slots are active from the install (they refuse nothing).
- Enforcement start: the epoch passes; no restart needed. Reversal: delete the policy file (back to warn-only) or set `MAIN_MAX`, `LONG_MAX`, `AGENT_WAIT`, `LOAD_FACTOR`, `LOAD_WAIT`, `MERGE_GRACE` in it. The previous lane wrapper stays in the ops repo history; `labrun builder-install` of an older commit restores it.
- Tags: V = ran: `tools/tests/test_lockrun.sh` of the ops repo (14 checks: refusal warn-only and enforced, package test allowed, override, other directory, merge class, quota, other agent, priority order, load guard, two long slots, memory gate of the second slot); a live main-lane job from an agent worktree queued behind a real merge and the queue showed both classes (V, the lane state file). U: behaviour under many simultaneous agents; the quota counts live processes, so a job killed by hand frees its place at once.

## Not done here

The merge gate still calls the builder twice per code merge (card mgonejob); the two-minute merge grace covers the gap until then.
