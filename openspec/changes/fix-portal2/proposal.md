## Why

docs/research/town-portal-trace.md (DECOMPTOME) proved that d2d starts a cast on a right click of a portal scroll or tome: the portal object appears about 280 ms later, the click is silently dropped while the player is busy, and the charge is spent at the click. Of 778 "no portal opened" lines, 17 percent were `context canceled` (three tries within 10 ms, nothing clicked) and 15 percent had a tome with charges (dropped click or cut cast). fix-portal fixed the empty tome (67 percent); this card fixes the rest in the d2d portal verb.

## What Changes

- cmd/d2bot/d2dwire.go `d2dPortal` (d2d only): never tries or retries with a dead context (returns a cancelled error naming the cause); clicks only when the player mode is idle (bounded wait of 2 s, polled every 100 ms); succeeds only when portal object 59 appears near the player (a scroll leaving the pack no longer counts); re-reads the pack and charges before each try, so a cut cast retries with the next source and an empty source ends with the no-portal error; failure lines say the cause (empty, busy, cancelled, no object: dropped or cut). The log line `d2d portal: click without a portal` stays and gains `sourceSpent`. No timeout change.
- Delta: backend-contract ADDED requirements.
