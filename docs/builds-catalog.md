# Build catalogue (build text and stars)

> status: current (card buildcatalog). One file, `internal/plandata/build-catalog.json`, loaded by `internal/plandata/catalog.go`; the UI (card E6) reads it.

- A build is one real build, not one plan table: a Hammerdin owns several tables (`tables[]`). Its name, identity, summary, 3 pros, 3 cons and seven ratings live only here.
- Add a build: add an entry (copy one), list its plan tables (file name without `build-` and `.json`), write all seven ratings with `source: estimate`. A new plan table with no entry and no `uncovered` reason fails `TestBuildCatalog`.
- Limits (phone card): identity 80, summary 280, pro/con 90, rating note 100 characters; stars 1 to 5. The loader errors on any unknown class or table, never falls back.
- `estimate` = a veteran's opinion (the d2-sme draft, all unverified). `measured` = turned into stars from a named scenario, with the number in `note`. Never from a CHEATED character.
- To replace an estimate: run the scenario at least 3 times, read the number against the `thresholds` block in the same file (SME ballpark, unverified, expected to be re-set after the first measured runs), set `stars`, `source: measured` and write the number and run id in `note`.
- Small card: Defense, the two highest of the other six, then the lowest of the remaining four; ties go to the earlier key in the order leveling, farmingItems, farmingXP, bossing, defense, clearSpeed, idle. The detail view shows all seven.
- Accessibility text: `Defense: 4 of 5 stars, estimate` (Defense adds `(Normal 5, Hell 4)` where the file has `defenseByDifficulty`).
- `build-catalog.json` shares the `build-` prefix with the plan tables; `plandata.TableIDs` and `buildfile.PlanNames` skip it.

## The Build tab (card buildcards, review E6)

- `GET /api/hero/build` carries `cards[]` for the hero's class (`internal/points/buildcards.go`): one `BuildCard` per catalogue build, keyed by the catalogue id, the small-card ratings (each with `text`) and `variants[]` = the plan tables behind it that the leveler can load (the name is what `POST /api/hero/build` takes). `cardsError` names an unknown class; there is no fallback to another class. `GET /api/hero/build?detail=<id>` is one card's detail (all seven ratings, summary, pros and cons, `defenseByDifficulty`, variants with needs and the first levels); another class or an unknown id is a 404 naming it.
- The page draws it with `web/buildcards.js` (`docs/ui-design-system.md` section 4.7). Auto-allocation off (`agency.autoAlloc`, default off) = browsing only: Choose is disabled with the visible line "Auto-allocation is off".
- Tests: `TestCardsThreeClasses` and friends (`internal/points/buildcards_test.go`), `TestBuildAPICardsAndDetail`, `TestBuildCardsDOM` (golden `internal/watch/testdata/buildcards_dom.txt`), fit pages `builds-list` and `builds-detail`. Pictures: `docs/ui/build-cards/`.

## Variants in plain words (card buildchoose)

Each catalogue build lists `variants[]` (one per plan table: `table`, a plain `label`, a one-line `why`, optional `fromLevel`) and a `defaultVariant`. Contract (`TestCatalogVariants`, `validateVariants`): exactly one default, it is one of the build's tables and has `fromLevel` 0; label and why are non-empty and never show a table id; an unknown build or table is an error naming it. The API variants carry `label`, `why`, `default` and `recommended`: the default for the hero's level is the variant with the highest `fromLevel` not above the level among the supported tables (so the full 90-level plan is the pick from level 30), else the catalogue default, else the first supported plan. An experimental plan is never pre-selected while the build has a supported one (`TestCatalogDefaultVariantIsSupported`); a build whose plans are all experimental (today: Lightning Sorceress as loaded in the test, Poison Nova, Barbarian Frenzy, Tornado Druid) has a default but nothing marked recommended, and the Choose sheet warns inline. U: the bot does NOT switch tables by itself when the hero reaches level 30 (no such code exists, V grep of internal/leveler, points, d2bot); the sheet therefore never promises it, and the pick follows the level only at the moment the player opens the sheet.
