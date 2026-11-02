## Why

Gate 1 pass (docs/soaks/gate1-20261102.md, Review 3) could not count three guards from the logs: the explore give-up, the tome debounce and the kit gate. Only the ping-pong guard had a countable line.

## What Changes

One structured INFO line per event, rate-limited by construction (once per event), behaviour unchanged: `explore-giveup`, `tp-debounce-hold`, `tp-debounce-flip`, `kitgate-trip`; `kitgate-skip` is verified by a test in its documented condition. Formats: docs/botlogs.md.

## Impact

internal/bot (goal.go, resource.go, tpdebounce.go and tests), docs/botlogs.md. Needs agent/kitgatelevel (merged into this branch, the kitgate-skip code lives there).

Overlap (declared): goal.go and resource.go are owned by the archived and open soak cards (kitgatelevel, tpglitch, soakstall); the edits are one log call each. kitgatelevel is merged into this branch, so merge kitgatelevel first or together.
