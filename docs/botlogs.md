# Countable guard lines

Soak extractors (tools/ops gate judge, d2dsoak) count guard events by `msg=` in the bot log (slog text format). One line per event, never repeated for the same event; behaviour is unchanged by the logging. Status: offline tests only (internal/bot/botlogs_test.go), U live.

| msg | when (exactly once) | attrs |
|---|---|---|
| `explore-giveup` | the second failed explore in a quiet area with no kill in between; the area is then marked cleared. Quest and boss runs never emit. The older human line `area abandoned: exploring failed with no kill` still follows. | `area`, `failures` (2), `kills` (kills this game so far) |
| `tp-debounce-hold` | a run of false portal-source looks ended in a true look before `tpDebounceSpan` (2 virtual minutes): the debounce kept the verdict `true`. Written when the run ends. | `verdict` (true), `raw` (false: value of the dropped-out read), `looks` (false looks in the run), `held_s` (virtual seconds from first false look to the true one) |
| `tp-debounce-flip` | the false run reached the span and the verdict became false. Written on that look. | `verdict` (false), `raw` (false), `looks`, `held_s` (>= 120) |
| `kitgate-trip` | a finished town trip whose reason holds `kit: ` (counted by `logTrip`, next to `trip budget: trip done`). | `level`, `missing` (text after `kit: `), `gold` (after the trip), `trip` (number since the kit was last complete) |
| `kitgate-skip` | from KitGateLevel (clvl 8) up: the cap `kitTripCap` of kit trips is reached with the kit still incomplete; once per stretch, reset by a complete kit. Never below clvl 8, where the old hold applies. | `trips`, `missing`, `gold`, `level` |
| `inv-tp-dump` | unchanged (docs/hastpdump.md). | |

Counting: `grep -c 'msg=explore-giveup '`. A hold line is a debounce that saved a false verdict; a flip line is a verdict change that the debounce let through. Not every false run is logged: a run still open when the log ends has no line.
