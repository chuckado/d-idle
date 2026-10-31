## Why

DRILLS5 findings F3 (a broke melee character rests instead of drinking mana) and F4 (the first `Fighter.Clear` after a position jump calls every monster unreachable). F3 turned out to be the intended economy rule; F4 is a stale map in the navigator.

## What Changes

- `nav.Navigator.Sync` (observe once, advance the map tracker) and the optional `combat.Syncer`; `Fighter.Clear` and `Kill` call it first.
- A test pinning the poor-melee potion rule (no code change).
- docs/fixmisc.md: findings, changes, verified vs unverified, live proof recipe.
