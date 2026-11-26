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

## In-town check (entry reflex1b)
The 2026-11-26 02:51 death: a trip in its chores stage, the hero in Cold Plains with 12 hostiles, the shop chore walking 21 s toward a town NPC.
- `town.Chores` verifies "in town" before every chore step and `openNPC` before every NPC walk; outside a town they return `town.ErrNotInTown` and the rest of the chores are skipped.
- The guard aborts a running chore after 2 s outside a town (`choreGrace`, cause `errLeftTown`); it is not a failure, the next step restarts the trip from the way home (`advanceTrip`, stage 0).
- Tests: `TestChoresOutsideTownAreAbortedAndTheTripGoesHome` (bot, fakes, Sorceress and Paladin; fails by hanging when the guard rule is removed [V mutation]); `TestChoresAbortWhenTheHeroIsOutsideTown` (town sim rig, needs the game tables and skips without them [U: never run here]).
