# D. Layout presets, migration and rollout

> status: sections 2 and 4 are SUPERSEDED by [J-second-attempt.md](J-second-attempt.md) (2026-11-14): the default view is the cockpit, a template layout, not a free-grid preset, and the fallback rollout is withdrawn. Sections 1 and 3 (what the shell keeps, the migration table) stand. Earlier status: accepted by the user's answers of 2026-11-14; sections 1 and 4 were rewritten for answer 4 (Standard is the default from S1) and answer 3 (a new hero starts on Simple). The drafts in [layouts/](layouts/) are `watch.layout/1` files written and checked by [layouts/gen.py](layouts/gen.py) (V: `python3 -I docs/ui/redesign-2026-11-14/layouts/gen.py` prints 0 findings: schema, spans inside 12 columns, no overlap among shown entries, unique order, known ids). They name six panels that do not exist yet, so the editor loads them only after slice S1 adds those ids to `reserved` in `web/shell.json` (U until then). Nothing in `internal/watch/web/` was changed.

## 1. The shell stays

Unchanged: `watch.layout/1`, 12 columns, 48 px rows, the four breakpoints (phone under 640, tablet under 1024, laptop under 1440, wide above), fit rows, the editor (`?edit=1`), `web/layout.json`, `web/layouts/{mockup,live,command}.json`, the alias, compat and reserved blocks of `web/shell.json`, every panel id and every pre-shell card id.

Four small additions are needed. Each is additive and ignored by an older shell:

| Addition | Why | Compatibility |
|---|---|---|
| Layout key `status: ["statusline", "hero-meters", "needs-count"]` (top level, optional) | The status band's panels come from `web/shell.json` today, so a preset cannot choose the new band. With the key, a layout does; without it, the `shell.json` list applies as now | `ParseLayout` ignores unknown keys (V, design system 15.3). `tools/uilint` is strict for repo files and learns the key in S1 |
| Panel manifest key `optIn: true` | The shell packs every panel a layout does not name after the placed ones. Without this key, a user with a saved layout would get six new tiles appended to it the day they ship. An `optIn` panel is mounted only when the layout in force names it | A new optional manifest key; `watch.panel/1` readers ignore it |
| Layout entry key `fallback: "<old id>"` (and the same on a `status` entry written as `{id, fallback}`) | Standard becomes the default with S1, before the new panels exist. An entry whose `id` is not a mounted panel is replaced, in the same cell, by the panel or card its `fallback` names; when the new panel ships, the entry resolves to it and the old tile is hidden as the preset says. Without a `fallback` an unknown id is skipped as today | An older shell ignores the key and skips the reserved id, which leaves a gap but no error |
| `web/layouts/{simple,standard,stream,debug}.json` and `?layout=<name>` for them | New files next to the three that exist | `?layout=` already loads a preset by name (V, design system 15.5) |

The presets switch old tiles off with the per-entry `hidden: true` that the editor already writes for a tile the user turned off (V `web/layout.js`), so they need no new mechanism: every existing id is named in every preset, either placed or hidden.

## 2. The presets

All four share the status band `statusline | hero-meters | needs-count | Inventory | Skills | ...`.

### Standard (the proposed default), 1440 px and wider

```
+ status --------------------------------------------------------------------------------------------+
| d.idle  * FARMING Cold Plains, for level 10, to unlock Blood Raven   life ===- 36  mana ==== 105  (2) Inventory Skills ... |
+----------------------------------------------+----------------------+----------------------+
| video                                  6 col | map (minimap)  3 col | needs-you      3 col |
|                                              |  Layers  + -         |  ! Weapon broken     |
|                                              |                      |  ? Go for Blood Raven|
|                                              +----------------------+----------------------+
|                                              | zone-map             | orders               |
|                                              |  the act graph       |  Go to  Town  Repair |
|                                              |  pins 1 2, route,    |  Pause  More         |
|                                              |  level bands         |  > Farm until repair |
+----------------------------------------------+                      |  1 Identify all      |
| progress   lvl 7 ===--- 10  25 min  xp/h ~^  |                      |  2 Cold Plains 30 min|
|            chapter 3 of 9  gold/h            |                      |  plan: ...           |
+----------------------------------------------+                      +----------------------+
| journey (as built)                           |                      | health  o o o o o    |
+-----------------------+----------------------+                      +----------------------+
| cards                 | hero-vitals          |                      | log  Now Journal ... |
+-----------------------+----------------------+----------------------+----------------------+
```

Positions (col, row, colSpan, rowSpan; `onboarding` is 1, 1, 12, 3 and takes no cell once hidden, so everything rises three rows):

| Panel | wide and laptop | Why there |
|---|---|---|
| `video` | 1, 4, 6, 11 | the picture is first and never below the fold |
| `map` | 7, 4, 3, 6 | the minimap belongs to the picture |
| `needs-you` | 10, 4, 3, 3 | top right, where the eye lands after the sentence; hidden when empty, and Orders rises |
| `orders` | 10, 7, 3, 7 | next to the act map it takes places from |
| `zone-map` | 7, 10, 3, 12 | the graph is portrait (356 by 636 units, V `ordersmap` `box`), three columns fit it |
| `health` | 10, 14, 3, 2 | one row of lights |
| `progress` | 1, 15, 6, 3 | under the picture |
| `journey` | 1, 18, 6, 4 | as built |
| `log` | 10, 16, 3, 6 | last |
| `cards`, `hero-vitals` | 1, 22, 3, 3 and 4, 22, 3, 4 | second screen |

### Simple

`video` 8 columns, then one right column: `needs-you`, `orders`, `health`, `progress`; `journey` full width underneath. No minimap, no act map tile (it opens from "Go to"). Density `comfortable`. For a friend's first week and for the Electron viewer in a small window.

### Stream

`video` 9 columns by 16 rows; right column `map`, `progress`, `orders`, `log`. No `needs-you`, no onboarding, no buttons that matter to a viewer (the verb row hides when the page has no answer origin, as writes are refused anyway). Density `comfortable`. It is a layout for a second monitor or a capture; `?mode=overlay` stays the OBS strip.

### Debug

The new top (video, a 6 column `map` with all its controls, Orders, act map, log) and then every tile that exists today at its present size: `hero-vitals`, `why`, `activity`, `gear`, `tally`, `currency`, `now`, `orders-map`, `queue`, `thoughts`, `journal`, `chronicle`, `quests`, `cards`, `bestiary`, `atlas`, `locked`, `rules`, `diary`, `alerts`, `spend`. Nothing is hidden. This is where the map's timeline, replay and exports stay visible.

### Narrow widths

Tablet (640 to 1023, which covers the 900 px case): flow by `order`; half-width pairs share a row. Standard: video; needs-you; orders beside zone-map; health beside progress; map beside log; journey; cards beside hero-vitals. Phone: one column in the order video, needs-you, orders, health, progress, map, log, journey. The sticky band plus the picture plus the first Orders row fit an 844 px screen (band 56, picture 293 at 4:3, verb row 52, two order rows 88).

## 3. Migration table (old layouts keep loading)

Rule: **no id is removed and no existing alias changes its target.** A saved layout or one of the three old presets renders exactly as today, because every tile it names is still mounted when named.

| Old id | Kind today | In the new presets | In an old layout | Later, if the tile is ever retired |
|---|---|---|---|---|
| `hero-status`, `hero-detail`, `whyline`, `hero-now`, `voice` | alias (V `shell.json`) | resolve as today | unchanged | keep the alias |
| `hero-vitals` | panel | Standard (second screen), Debug; hidden in Simple and Stream | unchanged | stays (it is the Hero sheet) |
| `why` | panel | hidden; content is the `statusline` popover | unchanged | manifest `mergedInto: "statusline"` + alias `why -> statusline` |
| `activity` | panel | hidden; content is the first Orders row | unchanged | `mergedInto: "orders"` + alias |
| `gear` | panel | hidden; the alert is a `needs-you` row, the chips the Gear light | unchanged | `mergedInto: "health"` + alias |
| `tally`, `currency` | panel | hidden; a line and a popover of `progress` | unchanged | `mergedInto: "progress"` + alias |
| `journey-chip`, `alerts-count`, `hero-status` (status band) | panel | not in the new band (`status` key) | unchanged (old band) | stay as stubs |
| `now`, `queue` | compat card | hidden | unchanged | alias `-> orders`, removed from `compat`, the `data-lid` section stays in `watch.html` as an empty stub |
| `orders-map` | compat card | hidden (the same data draws `zone-map`) | unchanged | alias `-> zone-map` |
| `atlas` | compat card | Debug only | unchanged | alias `-> zone-map` |
| `thoughts`, `journal`, `chronicle`, `diary`, `alerts` | compat card | hidden; tabs of `log` | unchanged | alias `-> log` |
| `rules`, `spend`, `quests`, `bestiary`, `locked`, `cards` | compat card | Debug (and `cards` in Standard) | unchanged | stay |
| `risk` | reserved | - | skipped as today | stays reserved |
| new: `statusline`, `needs-count`, `needs-you`, `orders`, `health`, `progress`, `zone-map`, `log` | - | placed | not mounted (`optIn`) | - |

The "later" column is slice S8, the last one, started only on the user's word (answer 8). It is written down so the retirements, if the user ever wants them, follow the promise: stub directory, `mergedInto`, alias, and a new compat fixture. Tests to add with the presets: `testdata/compat/layout-v1-{simple,standard,stream,debug}.json` (once shipped they are only added to), and one test that loads today's `layout-v1-{mockup,live,command}.json` plus `web/layout.json` under the new shell and asserts the same placed ids and cells as before.

## 4. Rollout (user answers 3, 4, 8)

1. **With S1, Standard is the default for everyone without a stored layout.** `web/layout.json` becomes the Standard file; today's file moves unchanged to `web/layouts/classic.json` and stays loadable with `?layout=classic` and from the menu. A stored per-profile layout is never read differently and never rewritten.
2. **Standard is progressive.** At S1 none of the six new tiles exists, so each of their cells shows its fallback:

   | Cell (new tile) | Shows until the tile ships | Ships in |
   |---|---|---|
   | `statusline` (band) | `hero-status` | S2 |
   | `needs-count` (band) | `alerts-count` | S3 |
   | `needs-you` | `gear` (the repair alert and its button) | S3 |
   | `health` | `why` | S3 |
   | `progress` | `tally` | S3 |
   | `orders` | `now` (the order, Cancel, Come home) | S4 |
   | `zone-map` | `orders-map` (the act graph that takes orders today) | S5 |
   | `log` | `thoughts` (open questions and the feed) | S6 |

   A fallback wins over the preset's own `hidden: true` for that old tile while it is standing in. Simple keeps the act map as a tile under the Journey strip until S5 (when "Go to" can open it as a sheet and S5 hides the entry), because otherwise Simple would have no way to give a zone order.

   So the day S1 merges, the default page is: picture, minimap, gear, the current order, the act map, why, tally, thoughts, journey, cards, hero. Every command that works today still has its button on the first two screens; nothing is blank. Each later slice swaps one cell. A test in S1 asserts that Standard resolved against today's panel list has no unknown id, no overlap and no empty cell.
3. **A new hero starts on Simple** (no stored layout and the onboarding card not yet dismissed); the same fallbacks apply. When the onboarding card is dismissed or done the page offers Standard once.
4. **Going back is one click.** The "..." menu of the band has "Layout: Classic / Simple / Standard / Stream / Debug", kept per browser like the theme. S1 also shows a one-time line under the band for users who had the old default: "The layout changed. Classic is in the menu."
5. **Stream and Debug** are opt-in from the menu or `?layout=`.
6. **Old tiles are retired last** (slice S8, only after the user says the new layouts have proven themselves): `mergedInto` and alias per the table in section 3, the stub stays, a compat fixture is added. Until then every old tile stays mounted for Classic, Debug and stored layouts.
7. The Electron viewer wraps the same page and gets all of it with no change (S `apps/desktop-electron`, ADR 0020).
