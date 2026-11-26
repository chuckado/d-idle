# faultsoak (TASK-979, P12) design notes, from the quest drill agent

> status: design only (2026-11-26). Proposed text for `docs/design/faultsoak-p12.md`; nothing here was run. Tags: V verified in code or a run, U unverified.

Oracle: the invariant monitor (`docs/invariants.md`) through `cmd/d2scenario` (`cmd/d2scenario/invariants.json`: hard rows turn a run RED, soft rows are logged). Every injection is a lab verb or a runner-side wrapper, goes through the ledger (ADR 0014 DIAGNOSIS), happens only in a soak row (never a measured run) and is time-stamped in the result so a hit can be tied to its fault.

| fault | hook today | missing |
|---|---|---|
| zero-life reload (the live Questy bug: reload at life 0) | V `debug stat` life = 0 on the standing hero (setup verb of `internal/scenario/setup.go`); expected invariant `life-zero-alive` or `potion-not-drunk`, fixed by qd26 | none; one row per class |
| empty belt / zero purse | V `debug stat` gold, belt items by setup (`pre.belt`, `pre.items`), `debug paygold` | a mid-run verb that removes belt items (U: `debug itemat` only reads) |
| drop life mid-fight | V `debug stat` life, `debug mon set`, `debug spawn` crowds | none |
| missing NPC | V `debug kill` on an NPC class (U: whether the unit leaves the table, not only dies) | a clean `debug unit remove <class>` (fork) |
| stale portal | V `debug portals` reads portals | a close/expire verb (fork); a bot-side stand-in is to take the portal after `debug goto` away |
| failed portal cast | V setup without scroll or tome (as the Smith row) | none |
| control socket timeout (14 in the quest log) | none | a runner-side proxy on the unix socket (`d2dctl.Options` wrapper) that delays or drops one reply per N; lives in `cmd/d2scenario`, no fork work |
| cancel storm | V `d2dctl` cancel paths | a driver that sends cancel on a timer; same proxy |
| dropped packets | none | not a lab verb in the fork (U); skip, the proxy covers what the bot sees |

Rows: one scenario per fault, `soak.fault.<name>`, class Paladin and Sorceress, report-only until a night run shows no flakes (`-repeat 3`); each hit becomes an incident scenario (`tools/provingground/incidents/`). Order: zero-life reload first (a hook and a fix exist), then belt, NPC, proxy faults.
