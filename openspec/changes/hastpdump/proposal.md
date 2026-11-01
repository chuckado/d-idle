## Why
docs/soakstall.md: HasTP flips per area for an unknown reason; the missing evidence is an inventory and UnitStats dump at area entry.

## What Changes
Rate-limited `inv-tp-dump` INFO line from the HasTP code path (internal/bot/tpdump.go, one call in `hasPortalSource`, one Runner field), offline test, docs/hastpdump.md drill.

## Impact
No behaviour change. Overlap (declared): internal/bot/resource.go and runner.go are touched by one line each (hook), owned by main.
