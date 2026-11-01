## Why

Gate 1 rerun (bot f4ca264, d2d lab2-freezeresume) failed the belt criterion: a 6 virtual minute window with no healing potion outside town (limit 5). The window sits at clvl 7 in the 15 minute recovery cooldown; evidence and cause in docs/fixbelt.md.

## What Changes

- internal/resource/recovery.go: `KitLow` and `RecoveryInput.LowKit`: healing potions under `recoveryPotionFloor` with an unaffordable kit start a recovery inside the cooldown.
- internal/bot/broke.go passes it; unit tests in internal/resource/recovery_test.go and internal/bot/area_ban_test.go.
- docs/fixbelt.md: evidence, cause, live recipe.

## Impact

No spec delta (the soak-gates belt requirement is unchanged). Awaits a live proof (a Gate 1 soak). Overlap (declared): FIXPROGRESS owns the town portal scroll and walk-to-town logic; this card does not touch them.
