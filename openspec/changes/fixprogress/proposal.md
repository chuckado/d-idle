## Why

Gate 1 rerun (docs/soaks/gate1-rerun.md on agent/gate1) failed the progress criterion: 43 virtual minutes with no kill and no XP (limit 10). The log shows a trip loop, not a stuck walk: walk to town, shop buys nothing with 600 gold, walk back, field check asks for a scroll again.

## What Changes

- internal/town: the game's word "no portal source (empty)" (ToTown finds no usable scroll or charged tome) is remembered (`NotePortalEmpty`) and outranks the tome stat read in the next visit, so the shop restocks against it. Two such visits in a row silence the "no town portal scroll" trip for FutileMinutes (the watchdog), then it may try again.
- Tests: internal/town/fixprogress_test.go (sim world).
- docs/fixprogress.md: timeline, evidence, cause tags, live recipe.

## Impact

No spec delta. Awaits a live proof (docs/fixprogress.md, section Live recipe).

## Overlap (declared)

- m2-tripwp owns internal/town/trip.go: this card adds two lines there (set the flag on ErrNoPortal, clear it when a portal arrives).
- goldfloor owns internal/town/shop.go: this card changes one case of scrollCount there (a tome counts 0 while the flag is set).
