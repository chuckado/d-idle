## Why

docs/primitives-review.md F4 and F5: the d2d portal verb clicked an empty Tome of Town Portal (939 `no portal after the click` lines in navsoak3, 81 of 246 fallbacks right after a shop visit with 0 charges), and the kit check counted an empty tome as a portal, so trips and field runs started without one.

## What Changes

- internal/combat/portal.go: one shared pick `PortalSource` / `PortalQty` (scroll or tome with charges; an empty tome is nothing). Combat's picker, the d2d town portal, the bot kit check and the boss view all call it.
- cmd/d2bot/d2dwire.go: `d2dPortal` uses it; no source means no click and an immediate `ErrNoPortal`; a failed click logs what was clicked and the portal objects seen.
- internal/bot/resource.go, internal/combat/resource.go, cmd/d2bot/boss.go: `HasTP` is true only with a portal source.
- Delta: backend-contract ADDED `The portal verb picks a source with charges`, `HasTP counts charges`. Shared by both backends (d2d tome charge = dump item qty; Game.exe = stat 70); the click bug itself is d2d-only.
