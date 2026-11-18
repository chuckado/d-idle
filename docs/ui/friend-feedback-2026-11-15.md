# Friend feedback, release 0.1.3 (2026-11-15): what each remark became

> status: current. UI and docs only; the bot's behaviour is unchanged. Tags: V verified (named test or picture), S secondary, U unverified.

| remark | what it was | what changed |
|---|---|---|
| (a) menus "too zoomed in", the Skills page | The Skills board scales by card width only (aspect 300:504): on a wide window three cards of about 600 px, on a phone one card of full width, so a tree was 600 to 1000 px tall. Fonts were not the cause (clamped tokens, V read). | `.sk-board` is capped at 22rem and centred, `.sk-trees` at 78rem (internal/watch/web/skills.css). Pictures: `tools/uishots.sh --skills` (V, see the card result). |
| (b) "do I need to drink potions" | not explained anywhere | Help card (see (d2) below): the bot drinks from the belt and rebuys in town (docs/potions.md, docs/gear-policy.md: potions unchanged). |
| (c) "A first for this hero" | The `t0.first` catalogue point (data in `internal/decide/firsts_act1.json`) carried the placeholder question "A first for this hero: the question is in the situation." and the placeholder answers "the first option / the second option"; the watch card used the placeholder instead of the first's own question and answers (`Sit` fields `question`, `yes`, `no`). It is an optional question with a default, never a requirement. | `Hub.BuildCard` words a first as "First time for this hero: <what happened and the question>", the answers are the first's own labels, the footer says "Nothing is required of you. If you do nothing, the hero will <default>" (`internal/watch/cardtext.go`, `views.go`). Catalogue text now reads "First time for this hero (optional, the hero has a default)." V: `TestFirstCardIsInformation`. |
| (d) see gear drops | The pickups were events (`goals.EvDrop`, emitted by the looter when an item reaches the pack) kept only for the return screen. | New read-only `GET /api/drops` (`internal/watch/dropsfeed.go`) and an **Items** tab, first and default in the log drawer: "picked up <name> (<area>)" in the quality colour, newest first. U: what the bot did with the item afterwards (kept, sold, stashed) is not in any event, so a row says only that it was picked up. V: `TestDropsFeedAndItemsTab`, `testdata/logpanel` goldens. |
| (e) "was I supposed to equip?" | answered by the Help card | Yes: the bot never equips (user decision 2026-11-10, docs/gear-policy.md); it only suggests on the Inventory page. |
| (f) "what drives the bot" | not explained | Help card: hand-written Go rules reading the game state, no trained model, no AI service, offline. V: stdlib-only module (`go.mod` has no dependencies) and no model or service client in `internal/` or `cmd/` (grep for the common names finds only the leak scanner's test). |

## The Help card

`Help` in the "..." menu opens the onboarding drawer's Help entries; the text card "What does the bot do for me?" is in `data/onboarding/en.json` (`help.what`), drawn by `web/panels/onboarding/panel.js`. It states, with the sources: potions (docs/potions.md), safety (reflex layer: docs/survival/README.md), gear is the player's (docs/gear-policy.md; unequipping or equipping by hand is respected because the bot makes no gear decision), items by loot rules and the player's marks (docs/lootrules.md), what drives it.

## Fit pictures

Fit pages added to `tools/uishots.sh --fit`: `demo-drawer-items` (the Items tab) and `demo-help` (the Help card), at every full-screen, window and phone size. The pictures of the Skills page: `tools/uishots.sh --skills [--sizes 'WxH ...']` (fake trees from `tools/uishots/fixtures_skills.go`).
