# Park retry after a new release (2026-11-21)

Evidence (V, sandbox hero, deploy 25): a chunk parked by an older buggy release (`gave up after 3 tries: the step
kept failing`) stayed parked after the fixed release started, because a park survives a restart (Q5).

## Rule
- Only parks caused by a failing step: reason code `gave_up`. (`game_failed` is the game failing the quest: Retry
  refuses it by design, so it is not touched. `died_too_often`, `hold_timeout`, `prereq_*`, `quest_locked`,
  `unsupported`, `regress`, `handin_failed` stay as decided.)
- The park stores the bot release id (`Run.Release`, plan file v1, additive, omitempty). The id is
  `buildinfo.Current()` (the ldflags stamp), overridden by `D2BOT_RELEASE_ID` (scenarios fake a new release).
- When the board first loads in a process and a `gave_up` park carries another release id (an older file has none,
  which counts as another), the run is retried once through `SM.Retry`, with `Run.Release` set to the new id and
  `RetriesAfterRelease` incremented, and the task executor's group unparked. Log line: `quest board: <chunk> park
  cleared after release <old> -> <new>: retrying once`.
- If the step fails again the park is stored with the new id, so the next start of the same release does not retry.

## Tests
`go test ./internal/questboard -run TestSMRelease` (Paladin and Sorceress): retry on release change, same release
stays, hold_timeout stays, second failure parks again with no third attempt, field round trip.
Scenario `board.park-retry-after-release` (pending): runner field `trigger.restart_new_release`.

## Proposed ADR 0025 addendum
"A park caused by a failing step (gave_up) records the bot release that stored it. On the first board load of a
different release the chunk is retried once (history kept in retries_after_release); a second failure parks it again
under the new release. Parks that need the player stay until the player acts (Q5)."

## Live checklist (not run)
RED then green on a lab slot with `D2BOT_RELEASE_ID` unset vs set; Tower builder LONG lane, determinism pair.
