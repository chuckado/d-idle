# Safety latency and cancellation latency

> status: current (card reflex1a, 2026-11-26). Source: docs/research/fable-logic-review-2026-11-25.md sections 3.1 and 3.3, proposal P3.

## Safety latency (measured)
The time from a safety verdict (the guard saw `Fighter.Safety` return an action, `internal/bot/safetylat.go`) to the input leaving the bot.
- Drinks (life, rejuvenation, mana): to the end of `Fighter.Act`, which is the key press.
- Portal, chicken, crowd: to the start of `Fighter.Act`. The cast that follows is the action's own work; what the measurement isolates is the wait for the cancelled action to return.
- A send without a guard verdict (paused-fight and hazard paths) is not counted.

Where it shows:
- Minute line, only in a minute that sent a safety action: `safetyN`, `safetyLatAvgMs`, `safetyLatMaxMs`, `safetySlow` (sends over 300 ms), `reflexDrinks` (drinks the guard sent itself, entry reflex1c).
- Death autopsy: `safetyLatMs` (last send in the 60 s trail window) and a summary clause "last safety action left N ms after its verdict".

Numbers: none yet from a live run [U]. The offline test pins the arithmetic (a 400 ms action gives 400 ms).

## Cancellation latency (T6)
`internal/combat/cancellat_test.go` cancels a running action on the simulator's virtual clock and bounds the virtual time to its return (1500 ms), on a Sorceress and a Barbarian [V: both pass; the test fails with 40 s and 175 s of virtual time when the context checks of `Fighter.sleep` and the fight loop are removed].

| action | covered | result |
|---|---|---|
| combat `Clear` | yes | returns within the bound |
| combat `Kill` | yes | returns within the bound |
| combat `Buff`, `Act` (portal, chicken) | no | U |
| nav `GoTo`, `ExitTo`, `Explore`, `TakeWaypoint` | no (needs the nav simulator) | U |
| town `Chores`, `ToTown`, `BackToField`, `Rebuy` | no | U |
| loot `PickUp` | no | U |

A blocking input call (a real socket write to the helper) is not modelled: the bound is for the loop's own checks of the context.
