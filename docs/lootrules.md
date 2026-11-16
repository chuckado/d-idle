# Loot rules: one engine for every item decision (TASK-421)

> status: design plus the PROTOTYPE (`internal/lootrules/`) WIRED IN SHADOW since wave 1 (cards lootwire-L2 to L4, branch agent/lootwire1): the bot still decides with today's code; the Starter preset also judges every ground pickup, tidy and sell decision and the differences are only logged and counted (section 11, knob `bot.lootShadow`, `GET /api/lootrules/shadow`). Wave 2a (cards lootwire2, section 12) adds the per-stage switch `loot.rulesEngine` (0 off, 1 ground, 2 ground and tidy; the sell chore stays on today's code), the rule store with `GET/PUT /api/lootrules` and the Loot Rules sheet; the default is OFF. The player-facing editor is designed in parallel in `docs/ui/redesign-2026-11-14/loot-rules-ux.md` (branch agent/uidesign2); section 8 maps its section 5 needs to this engine. Tags: V verified (how), S secondary (a doc, the SME, memory of public tools), U unverified, EXP veteran-player advice (SME consult 2026-11-14). Nothing here ran against a live game.

## 0. Summary

Today four tables decide about items and they disagree at the edges (docs/sellskip.md section 1): the pickit `internal/loot/pickit_starter.json` with its poor-mode and empty-slot overrides in `evaluate`, the tidy policy `internal/tidy/policy.json`, the sell chore's own `sellable` test, and the shop and identify chores. The engine replaces the decision part of all of them with ONE ordered, versioned, JSON rule set, compiled once per hero, asked at three places: the ground (pickup), the pack (keep, stash, sell, drop, identify, ask) and the stash. The chores keep doing the clicking; they ask the engine what to do.

The shipped `Starter` preset is today's pickit (ground rules) and today's tidy policy (pack rules) written in the new format, so the default behaviour does not change (V for the pack side: `TestStarterPackEqualsTidyPolicy` compares 432 decisions with `tidy.Decide`, three classes, two pack fill levels, two stash fill levels; the ground side is translated by hand and spot-checked, the full table test against the real pickit needs the game tables and is card L2).

## 1. The item fact model

A rule sees an `Item` (what the thing is), a `Ctx` (the situation) and, for requirement rules, two `Actor`s (the hero and the mercenary). Every field has a source and a place where it is unreadable; an unreadable field makes a condition UNKNOWN, never false (section 2.2).

| field | ground | pack | stash | source (V = read in code) |
|---|---|---|---|---|
| code, type and ancestors, name | yes (label parse on d2d) | yes | yes | `d2.Unit.Class` -> `data.ItemBase` (V `internal/loot/loot.go base`); ancestors through `data.DB.IsA` (V `isAny`); d2d ground: label text only (V docs/d2d-loot.md) |
| quality | GUESSED from the label (U for magic, rare, unique, set labels; docs/d2d-loot.md "Quality guess") | yes (`Item.Quality`) | yes | V `d2.Item.Quality` 1..8 |
| ilvl | NO on d2d ground (0); yes on the Wine path | yes (`Item.ILvl`, V +0x2C) | yes | flag `HILvl` marks it unreadable |
| identified, ethereal | identified: unreadable on d2d ground; ethereal likewise | yes (`Item.Flags`, V) | yes | `HIdentified` |
| sockets | no on ground | yes (stat 194 in the pickit, V rules) | yes | |
| stat values (life, fcr, resists, skills ...) | NO | only when identified (an unidentified magic or better item shows none, S: game rule; the engine also treats `!identified && quality >= magic` as hidden) | same | `d2.Stats` via `UnitStats` (V `loot.Looter.stats`); the display name table is `stats.go` (ids S from the pickit, others U) |
| sell value | no | `loot.SellValue` (V, a lower bound for identified magic and better: docs say so in `SellValue`) | same | flag `HValue` on ground |
| required level, strength, dexterity | yes (base row) | yes | yes | V `data.ItemBase.LevelReq/ReqStr/ReqDex` (used by `wearsBetter`) |
| class restriction | yes | yes | yes | V `data.ItemType.Class` (`wearsBetter` compares it with the hero's class code) |
| hireling types that can wear it | yes | yes | yes | U: from Hireling.txt plus the item type tree; `internal/merc` reads the merc; no function exists yet, card L2 builds it |
| durability, max durability, broken | no | yes (stats, U which ids) | yes | adapter fills `Durability`, `MaxDur` |
| rune number, gem grade, base tier, max sockets | yes (from the base row) | yes | yes | data tables (U: derived by the adapter; the rune number is the code order) |
| stackable, quest, start item, qty | yes | yes | yes | V `data.ItemBase.Quest`, `FlagStartItem`; qty: gold amount from the label (V docs/d2d-loot.md) or the stack stat 70 |
| mark (keep, sell) | n/a | yes (`invdesk` marks, V docs/inventory-actions.md) | yes | |
| place | the container: ground, pack, stash; equipment, belt, cube, cursor and vendor are NEVER decided |

Ctx: hero class and build (data keys), difficulty, act, clvl, gold, pack fullness percent, stash fullness percent, `Needs(kind)` (the belt, scroll, tome, key, gold, gem, charm-room, runeword and empty-slot needs the looter has today: `loot.needsOf`), `Owned(item)` (how many of the base the hero holds). Actors: `Hero{Class, Level, Str, Dex}` and `Merc{Type, Level, Str, Dex}`, nil merc = none, level 0 = not read yet.

Unidentified rules, explicit: a condition on `stat:*` of an unidentified magic-or-better item is UNKNOWN. A rule that is UNKNOWN is skipped, and the decision carries `Pending: true` ("a rule above could not be judged; identify and ask again"). A rule may say `"onUnknown": "match"` to take the item anyway (the Starter ground rules "charms to id" and "jewelry to id" do: pick it up, identify later). The `identify` action spends a scroll only where a rule says so (docs/inventory-actions.md: identify stays a player command unless a rule asks; the Starter has no identify rule, matching today).

## 2. The rule language

### 2.1 The file (`lootrules/1`, JSON, `internal/lootrules/types.go`)

```
{ "schema": "lootrules/1", "seq": 12, "name": "Starter", "comment": "...",
  "limits":   { "maxActions": 12, "stashMaxPct": 90 },
  "groups":   { "my_rings": { "types": ["ring"], "codes": ["rin*"] } },        // adds to the built-ins
  "defaults": { "ground": "ignore", "pack": "ignore" },                         // the fixed last row of each stage
  "ack":      [ "drop-valuable" ],                                               // the player's explicit acknowledgements
  "rules": [ { "id": "t-runes", "name": "runes", "scope": ["pack"], "classes": ["Amazon"], "builds": [],
               "when": <condition>, "action": "stash", "prio": 90, "tier": "S",
               "reason": "rune", "confirm": false, "onUnknown": "skip", "pickup": true, "off": false } ] }
```

Six lines of language: (1) an ordered list, first match wins, an explicit default per scope; (2) a rule is `when` (a condition tree) plus an `action` and a `reason` shown in the "why" tooltip; (3) conditions are `all`, `any`, `not` or a leaf `{f, op, v}`; (4) `scope` picks the stage (ground, pack, stash; none = ground and pack) and `classes` and `builds` are the per-hero overlays (data, no class literal in Go); (5) three-valued logic: unreadable fields make a rule UNKNOWN, not false; (6) safety floors that no rule can change, and `confirm` plus the set's `ack` for the one thing a player may unlock (dropping or selling valuables).

Strictness: unknown keys are refused (a typo never silently disables a rule), `id` is unique and stable across edits and reorders, `reason` is required.

### 2.2 Conditions

Leaf fields (the editor reads the list from `lootrules.Fields()`, 58 entries with label, unit, allowed values, where hidden; V `TestSchemaPreviewDiffShadow` checks the schema and the parser agree):
- item: `quality`, `type` (with ancestors), `group` (built-in: runes, gems, charms, jewels, jewelry, circlets, potions, supplies, gold, gear, javelins; own groups in `groups`), `code` (globs), `name` (contains), `ilvl`, `reqlvl`, `reqstr`, `reqdex`, `sockets`, `maxsockets`, `value`, `qty`, `runeno`, `grade`, `basetier`, `owned`, `durpct`, `identified`, `ethereal`, `stackable`, `quest`, `broken`, `stat:<name>` (life, mana, str, dex, vit, enr, fire/cold/light/poisres, fcr, fhr, frw, ias, mf, gf, allskills, classskills, skilltab, ed, and the derived `stat:ressum`).
- requirements (user addition 2026-11-14): `usable_by eq hero|merc` with an optional `plus {lvl, str, dex}` ("usable after +5 strength or +2 levels"), `hero_usable_from` and `merc_usable_from` (the lowest level at which the wearer can wear it: 0 now, the required level when only the level stands in the way, 999 when the class, hireling type, strength or dexterity does), `class_restricted in [classes]`, `reqlvl`, `reqstr`, `reqdex` for plain comparisons. "Unusable now but usable at clvl 30" is `hero_usable_from le 30` after a `usable_by` rule. A merc can wear an item only when its `MercTypes` list names his type (data from Hireling.txt, U); an unread actor (level 0) answers UNKNOWN, so a rule never says "cannot wear it" because a number has not arrived yet.
- context: `class`, `build`, `difficulty`, `act`, `clvl`, `gold`, `packfull`, `stashfull`, `need` (the looter's needs).
- ops: numbers `eq ne ge gt le lt`; bools `eq`; enums `eq ne in`; lists `in`; `name contains`. "Rare or better" is expanded by the editor into `quality in [...]`.

Evaluation is Kleene logic (true, false, unknown): `all` is false if any child is false, else unknown if any is unknown; `any` mirrors it; `not` keeps unknown. V `TestUnidentifiedIsExplicit` (an unidentified ring is identified, not sold; `not(unknown)` stays unknown; ground `onUnknown`).

### 2.3 Actions

`pickup` (ground only), `keep`, `keepmerc` ("keep for the mercenary": a keep the sell chore and tidy leave alone; the player equips him), `stash`, `sell`, `drop`, `identify`, `ask` (raises a Needs-you row and leaves the item until answered), `ignore`. There is NO equip or unequip action: the vocabulary itself guarantees the filter never makes an equipment decision (docs/gear-policy.md; V `TestInvariants`: `"equip"` does not parse). A rule with `keep`, `stash` or `keepmerc` on the ground means "pick it up" (the same rule serves both stages, as the pickit does today); `sell` and `drop` on the ground mean "leave it" unless the rule says `pickup: true` (the poor-mode junk-gear rule does). The decision then carries `Fate` (the rule's own action) and `Tier`, so the strictness knob (tier floor) and the sell-pickup in poor mode keep working in card L2.

### 2.4 Order of evaluation, marks and commands

1. Place: anything but ground, pack, stash is `ignore` (equipment, belt, cube, cursor, vendor).
2. Player marks win: keep -> ignore ("Kept: you marked it keep"), sell -> ignore ("Will sell: you marked it"; the sell chore owns marked sales). A player command (equip, trash, identify, stash move from the inventory menu) is not a decision of the engine; it runs before the rules, through the queue of docs/inventory-actions.md, and the engine never overrides it.
3. Start kit -> keep. Quest item -> pickup on the ground, keep in the pack.
4. The rules of the scope, in order, first match. Then the scope default.
5. Floors on the answer: stash action while the stash is at `stashMaxPct` -> keep; `sell` or `drop` of tomes, scrolls or keys -> keep; `drop` of anything but low, normal or superior quality (or of unreadable quality), and `sell` or `drop` of runes, gems, charms or jewels -> ignore unless the winning rule has `confirm: true` AND the set lists `"drop-valuable"` in `ack`. V `TestInvariants` (Paladin context), `TestStarterPackEqualsTidyPolicy` (the same floors as tidy's `dropSafe`, `neverSell`).
6. Plan: `Compiled.Plan` takes the pack, returns identify, stash (best value first), sell, drop (cheapest first), at most `limits.maxActions` (12 = `town.tidyMaxActions`), ties by item id, so one run is the same as the next (V `TestPlanCapOrderAndDeterminism`).

Determinism: no map iteration and no clock in an evaluation. Performance: `Compile(rs, class, build)` validates once, drops disabled rules and other classes' rules and turns every condition into a closure; a decision is a linear walk without parsing. V `BenchmarkStarterGround`: 40 ground items per iteration in about 65 microseconds (about 1.6 microseconds per item, one allocation per item for the why text), Starter, offline box. The ground scan each tick is therefore negligible; the only per-item cost the card L2 adds is building the `Item` from the unit. A decision cache keyed by (unit id, quality, identified flag, `Ctx.Fingerprint()` when `Compiled.UsesActors`) re-evaluates when the hero levels up or stats change (fingerprint changes: V `TestRequirementsHeroMercAndLater`).

### 2.5 Safety invariants (code, not data; `Invariants()` lists them for the editor's grey block)

place; marks; start kit; quest items; tomes, scrolls and keys never sold or dropped; dropping magic or better items (or anything of unreadable quality), and selling or dropping runes, gems, charms and jewels, needs a confirmed rule and the acknowledgement (selling a magic or rare item stays an ordinary rule, as the tidy `sellable` row is today: V `TestInvariants` shows a unique sells under a plain sell-all rule; the preview warns); stash-full floor; the action cap per pass; no equip action exists. Also in the town code, unchanged: abort on a fight, the 120 s pass cap, two failed drops end the pass, the cursor is never left holding an item (docs/tidy.md).

## 3. What exists today and what it becomes

| today | where | becomes |
|---|---|---|
| pickit rules (quest, runes, gems, uniques, potions "needed", charms with stats, jewelry, runeword bases, junk) | `internal/loot/pickit_starter.json`, `rules.go` | Starter ground rules `g-*` (translated 1:1 by name and order) |
| `needed` (potion, scroll, tome, key, gold, gem, runeword, charmroom) | `loot.needsOf` | `need eq <kind>`; `Ctx.Needs` stays the loot package's function |
| stat conditions `{stat id, min, shift}` | pickit | `stat:<name> ge N` (id table S, section 1) |
| tier S/A/B/C and strictness | pickit | `tier` kept on the rule; strictness is a tier floor applied by L2 |
| `fillsEmptySlot` (empty weapon or body slot) | `loot.EvaluateWith` | rule `g-empty-slot`, `need eq emptyslot` (keeps a wearable weapon or armor; picking up is not equipping) |
| poor mode (gold < 300: junk gear worth >= 20 sells) | `EvaluateWith`, `poorSellable` | rule `g-poor-gear` (`gold lt 300`, `pickup: true`) |
| `WearBetter` | `ctxFrom` | stays OFF: gear is the player's (docs/sellskip.md); not translated, the `usable_by` rules replace its purpose ("keep what I could wear") as data |
| tidy rows, per-class overrides, `dropSafe`, marks, start kit | `internal/tidy` | Starter pack rules `t-*` and `amazon-javelins` (`classes: ["Amazon"]`), floors in the engine |
| the sell chore's `sellable` and `needSell` | `internal/town` | asks `Plan(..., Sell)` (L3), ending the "bot counts N, town finds 0" disagreement of docs/sellskip.md |
| `Valuables` (what to stash before a dangerous trip) | `loot/valuables.go` | `stash` decisions plus the `value` field (L3) |
| the drop-pickup loop guard | `tidy.Drops`, `Looter.SetIgnore` | unchanged; the engine only adds that a rule that drops puts the item on the ignore list (section 7) |

Behaviour changes by default: NONE intended. Known deltas, all to be shown in the L2 table test: (a) a unique or set weapon or armor that fills an empty slot: today `fillsEmptySlot` turns Stash into Keep, the translation answers Stash (U, rare); (b) the engine adds invariants that were only in tidy (equipment never touched) to the ground stage where they cannot matter; (c) `WearBetter` stays off as it already is by default.

## 4. Migration and the Starter preset

Starter = the file `internal/lootrules/starter.json`, generated from the two sources (the 12 tidy rows plus the Amazon override (13 pack rules, tidy names as `name`) and 33 ground rules translating the pickit rows plus its empty-slot and poor-mode overrides) and kept as data. Exact translation of today's tidy rows (V by the table test):

| tidy row | rule `when` | action |
|---|---|---|
| tomes and scrolls | `code in [tbk ibk tsc isc key]` | keep |
| potions | `group potions` | keep |
| runes | `group runes` | stash |
| gems | `group gems` (type prefixes gem, skul) | stash |
| charms, jewels | `group charms`, `group jewels` | stash |
| unique and set | `quality in [unique set]` | stash |
| rare jewelry | `type in [ring amul circ] and quality in [rare crafted]` | stash |
| magic jewelry | `type in [ring amul] and quality in [magic]` | stash |
| socketed bases | `sockets ge 4 and quality in [normal superior]` | stash |
| pure junk | `value lt 20 and packfull ge 85 and quality in [low normal superior]` | drop |
| sellable | `value ge 20` | sell |
| Amazon override | `group javelins`, `classes ["Amazon"]`, placed first | keep |

`limits` carry `tidyMaxActions` (12) and `stashMaxPct` (90). The knobs `town.tidyEnabled`, `town.tidyMaxActions`, `town.tidyMinPackPct` stay as the switch and the cap; `loot.*` knobs that are conditions today (`poorGoldBelow` 300, `poorMinSell` 20, `valuableMinGold`) become the numbers inside Starter and the knobs are read once to fill them when a hero has none (L6 does the lift, so the migration is data, not code). The embedded `policy.json` and `pickit_starter.json` stay until L2 and L3 are merged and the table tests pass; then they become one-line stubs pointing here.

### Preset defaults from the SME (EXP, 2026-11-14)
All EXP. Starter does NOT adopt them (it is today's behaviour); they become the other presets in card L6 and the preview shows the difference first.
- Always pick up in Normal Act 1-2: all potions and rejuvenations, town portal and identify scrolls up to a cap (20 and 10), ALL runes, magic-or-better rings and amulets, rare circlets, uniques and sets of any base, all charms, white 3 to 4 socket-capable runeword bases. Gold floor about 20 x clvl per pile (conflicts with docs/goldprio.md, where the user wants gold prioritised: keep today's bar, offer the SME bar as a preset knob). Ignore: plain white no-socket items, magic items under about 40 gold sale value, javelins and arrows for non-ranged builds, keys (a leveling bot has no use), skulls and low gems beyond a stack cap.
- Runes: keep every rune, stash, never sell; grades El to Ort for the starter runewords, Pul (rune 20) and up are the upgrade path. Gems: chipped and flawed are junk unless a recipe needs them; keep regular and flawless of Ruby, Topaz, Skull for sockets and every perfect gem.
- Charms: stash +20 life, +1 class skill, all resists with +3 each, FCR 20 on a grand charm, MF 7+ on a small charm; early (clvl < 25) only life 15+ or a single resist 7+. Rings and amulets: +skills, FCR 10+ on a ring, life 20+, all resists 10+ (sum), MF 20+. Jewels: 15% IAS or good resist and life. Uniques and sets: stash anything with use; keep the first 3 to 4 set pieces; sell low-value uniques with no function.
- Sell: white gear, magic and rare items under about 400 gold at clvl < 20 (threshold rises with level). Stash: keepers above. Drop: only when the pack is full AND the item is white junk or an unneeded potion, in the order arrows, junk potions, cheapest white, magic under the threshold. Never drop: unique, set, rune, gem, jewel, charm, quest items, runeword bases with 3+ sockets, identified rares, rejuvenation. Sockets: keep 3 to 4 socket armor and shields, 4 socket weapons; ethereal only when the base has 3+ sockets.
- Identify: scroll 100 gold; spend one on rings, amulets, jewels, circlets or a usable base, not on plain magic items; unidentified rares: carry rings, amulets, circlets, boots, gloves, belts up to about 6 slots, sell the rest.
- Traps (become tests and preview warnings): picking up everything; dropping unidentified rares and uniques the filter cannot see into (so: the floor blocks dropping anything magic or better, and `Pending` defers a sale, section 7); selling charms, jewels, runes without a stat check; dropping a low gem a cube recipe needs; fetching gold behind a pack (the fight and danger guards of docs/goldprio.md, unchanged); the pickup-drop loop; selling a base you hoped to socket; "sells for little" is not "useless" (Stone of Jordan sells low, is worth a lot: so a sell rule on `quality unique` is never in Starter).
- Safety defaults the SME asks for, all adopted: caps per pass (here 12 stash plus drop; the SME says 10 pickups, 10 sells, 5 drops: L6 sets per-kind caps in `limits`, U), confirmation to drop or sell valuables with default OFF (the `ack`), a never-drop list (the invariants), one log line per decision with item, quality and rule id (L2, L3), keep 2 to 3 free pack cells (today `loot.freeCellsKeep` 2).

## 5. Unlock design (ladder pattern: `data/unlocks/ladder.json`, `docs/unlock-ladder.md`, `internal/watch/unlockorderqueue.go`)

Pattern: `ui` group rows with a trigger, optional `alt` triggers and a `price` (tally) when `tally_spend` is among them; slot rows (`orderqueue`, `orderqueue2`, `orderqueue3`) open features in steps and a small embedded table (`orderqueue.json`) lists them. The same shape fits here: `data/unlocks/lootrules.json`-style table not needed, three ladder rows suffice and the engine reads `Allowed(id)`.

| row | gives | trigger proposal (the user's call) | alt |
|---|---|---|---|
| (no row) | Starter, fixed, read-only view of the active rules with the reasons ("this is what the hero does and why"), the item tooltip's why line | always | |
| `lootrules_presets` | choose among the shipped presets (Starter, Runes and gems, Sell junk, Keep everything, Veteran), with the difference preview | `item_quality` qualities `[rare]` (first rare item found, early in Act 1, reuses the idreveal trigger kind) | `level` min 12 |
| `lootrules` | the full editor: add, edit, reorder, switch off, import, export, undo, own groups | `quest_stage` `search_for_cain` min 3 (Deckard Cain rescued) | `tally_spend`, price 60 |
| `lootrules_ack` (optional) | the `drop-valuable` acknowledgement and `confirm` on rules that drop or sell valuables | never automatic: a separate toggle in the editor with a typed confirmation, no ladder row | |

The UX page suggests "sold twenty items" (`seen` event `sold_20`): no emitter exists (U), so it is listed as an alternative for a later card, not the trigger. Reveal text for `lootrules_presets`: "Your hero follows a loot list. Pick another one that suits how you play." for `lootrules`: "Write your own loot rules: what to pick up, keep, stash, sell or drop, and why."

What "locked" means: the engine always runs the active set (Starter until the player picks otherwise); the API serves `GET /api/lootrules` in every state so the page can show the active rules read-only with the `Why` of every rule; a PUT or a preset change answers 403 `locked` with the teaser. The player learns the language before owning it. No behaviour is held back by the lock.

## 6. API for the UI (additive, versioned: `watch.lootrules/1`)

Same gate as the inventory routes (origin list, token or passcode cookie, JSON only, small body, unknown fields refused). All per hero.

| route | body | answer |
|---|---|---|
| `GET /api/lootrules` | | `{schema, seq, locked: none\|presets\|editor, active: <RuleSet>, protected: Invariants(), presets: [{id, name, description}], stats: {rules, shadowed:[...]}}` |
| `PUT /api/lootrules` | a RuleSet (seq of the one it was based on) | 200 `{seq}` or 400 `{errors:[{path, msg, line, col}]}` (syntax errors carry line and column, field errors the path `rules[3].when.all[1].v`); 409 `seq` when the stored set changed since; 403 `locked` |
| `GET /api/lootrules/schema` | | `{fields: Fields(), actions: {ground, pack, stash}, groups, classes, builds, qualities, itemTypes}`: the editor builds its pickers from it |
| `POST /api/lootrules/preview` | `{rules: <RuleSet, may be unsaved>, scope: ground\|pack, sample: drops\|pack\|stash\|both}` | `{counts:[{rule, count, ids}], decisions:{id:{action, rule, why}}, changes:[{id, from, to}], shadowed:[{rule, by}], warnings:[...]}`: `Preview` over the sample, `Diff` against the saved set, `Shadowed` for "never reached"; a valuable that a `sell` or `drop` rule would take gets a warning |
| `POST /api/lootrules/preset` | `{id}` | like PUT after the preview diff has been shown |
| `POST /api/lootrules/undo` | | restores the previous saved set (one step, kept in the state file) |
| `GET /api/lootrules/export`, `POST /api/lootrules/import` | the canonical JSON (`Marshal`) | import validates and previews, it does not apply |
| `GET /api/lootrules/why?id=<item>` | | the `Decision` and the `Explain` trace of every rule for one item |

The inventory view (`invdesk.ItemView`) keeps its additive `tidy` field and gains `rule` (id) and `why`; `ViewSchema` bump is the invdesk owner's call (as in docs/tidy.md). Versioning: `schema: "lootrules/1"` is checked on read; a newer major is refused with the version named; unknown keys are refused so a file cannot lose rules silently; a migration table `v1 -> v2` is the place for later changes. Storage: `lootrules-<profile>.json` beside the other state files, atomic write, one previous copy for undo, `seq` incremented on every save (the page reads once and on the change event).

### The recent-drops ring buffer (card L7)
Preview over the ground needs the last N items seen on the ground. `loot.Looter.ObserveGround` already sees each ground unit once per tick (docs/lootmemory.md, "Remember"); L7 adds a ring of 100 entries `{id, area, time, Item facts as filled by the adapter, the decision taken and the rule id, what happened (picked, left, failed)}` next to the memory, in memory only plus a JSON dump on shutdown (so a restart does not empty the preview). Unidentified items keep their `Hidden` mask so the preview can say "decided before identification". About 100 bytes of facts per entry plus name text: under 30 KB.

## 7. Risks and what answers them

| risk | answer |
|---|---|
| A bad rule drops or sells good items | floors in code (section 2.5); `confirm` plus `ack` for dropping magic or better and for selling or dropping runes, gems, charms, jewels; the preview warns before saving; the cap per pass; the town log line per decision with rule id; undo of the last save; marks always win |
| Rule UNKNOWN on an unidentified item, then dropped | an UNKNOWN condition never matches, and dropping magic or better is blocked by the floor whatever the rules say; for sell, the town chore (L3) does not sell a `Pending` item that can still be identified, and the plan lists `identify` first |
| Drop-pickup loop (the ground rules pick up what the pack rules drop) | every executed drop goes to `tidy.Drops` and `Looter.SetIgnore` as today; the preview shows a "ground picks it, pack drops it" warning (a static check: a drop rule whose `when` implies a ground pickup rule) in L4; the loop test of docs/tidy.md stays |
| Walking into danger for a pickup | unchanged: the fight and danger guards of docs/goldprio.md sit above the looter; the engine only answers what is wanted |
| Hero and merc numbers change (level up, stat points) | `Ctx.Fingerprint()` changes, decision caches clear when `UsesActors` |
| The editor produces a rule set the engine rejects | one validator (`Validate`) used by PUT, import, preview and `Compile`; errors carry paths |
| Performance | section 2.4, about 1.6 microseconds per item |
| Two readers disagree (sell chore, tooltip, pickup) | one `Compiled` per hero, built by one owner (the bot wiring in L2) and handed to the looter, tidy, the town sell chore and the inventory view |
| Per-class or per-build data goes missing | an unknown class has no overlay rules and the common rules apply (a table lookup, not a Sorceress fallback; AGENTS.md "Build flexibility"); L6 tests presets with Sorceress and one other class and a merc |
| Gear policy | no equip action; `keepmerc` only keeps; `usable_by` reads, never writes |

## 8. How this meets the UX page (loot-rules-ux.md section 5)

| UX need | here |
|---|---|
| schema with types, values, labels, units | `Fields()`, `ActionsFor()`, `GET /api/lootrules/schema` |
| validate with rule and field | `Validate`, `Error{Path, Msg, Line, Col}` |
| evaluate one item, return action, rule id, reason | `Decide`, `GET /why` |
| dry run over snapshots, per-rule counts and ids, difference to saved | `Preview`, `Diff` |
| never-reached analysis | `Shadowed` (conservative: an unconditional or identical earlier rule) |
| protected set as data | `Invariants()` |
| presets as data in the same format | `starter.json` plus L6 files |
| versioned format, forward tolerance, migration | `schema`, unknown keys refused, migration table; from today's tables via Starter |
| stable rule ids | `Rule.ID` required and unique |
| per-hero storage and `seq` | `RuleSet.Seq`, per-profile file |
| two stages, a fixed last row | `scope` plus `defaults.ground` and `defaults.pack`; the editor writes every rule with ONE scope, an unscoped rule (hand-written or imported) shows as "both stages" |
| "X or better" quality | expanded to `quality in [...]` by the editor |
| ask me | action `ask` (Needs-you row; L4 plumbs the question) |
| usable by hero or merc, after +N | `usable_by` with `plus`, `hero_usable_from`, `merc_usable_from`, `class_restricted` |
| quantity already owned | `owned` (`Ctx.Owned`) |
| drop log | L7 |
| ladder row `lootrules` | section 5; the sold-20 trigger needs an emitter (U) |

## 9. Wiring plan (cards, none started; sizes in agent hours, S 1 to 2, M 3 to 5, L 6 to 10)

| card | what | depends on | size |
|---|---|---|---|
| L1 | engine and tests: THIS branch (`internal/lootrules`, Starter, table test, benchmark). To harden: fuzz the parser, golden Starter file | none | done, S to harden |
| L2 | pickup wiring: `internal/loot` builds `Item`/`Ctx` from `d2.Unit` and the d2d ground label (`Enrich`), `Looter` asks the engine behind the knob `loot.rulesEngine` (default 0, then 1 after the table test passes against the real pickit with game tables: `D2_MPQ_DIR=$HOME/d2d-lab/data`, two classes); hireling-type data; `Needs` and strictness tier floor; poor mode and empty slot as rules | L1 | L (about 8) |
| L3 | tidy, sell, stash chores ask `Plan`: `internal/tidy` becomes a thin adapter over the engine, `town` sell chore uses `Plan(Sell)`, `Valuables`, `identify` plan step, `keepmerc` and `ask` plumbing; delete `needSell` disagreement | L1, L2 | L (about 8) |
| L4 | API: store, routes of section 6, locked states, undo, preview with sample sources, loop-warning check | L1, L7 (for ground preview) | M |
| L5 | UI: the Loot rules tab (Fable's slice, loot-rules-ux.md) | L4 | L, theirs |
| L6 | presets and SME defaults as data files (Runes and gems, Sell junk, Keep everything, Veteran), the three ladder rows, knob lift into Starter numbers, per-kind caps in `limits` | L1 (L3 for live check) | M |
| L7 | recent-drops ring buffer in `loot.Looter` (section 6), JSON dump, the sample source for preview | none (parallel with L1) | S |

Order: L1 -> (L2 and L7 in parallel) -> L3 -> L4 -> L6 -> L5. The first live check is L2's: a run with the engine on and the Starter, comparing `loot:` pickups and `tidy:` plans with the old paths over the same hour (zero differences expected except delta (a) of section 3).

## 10. The prototype (what is in `internal/lootrules/`)

`types.go` (rule set, rule, condition, item, ctx, decision), `parse.go` (strict parse, `Validate` with paths, line and column for syntax errors), `compile.go` (three-valued closures), `requirements.go` (usable_by, usable_from, fingerprint), `eval.go` (`Compile`, `Decide`, floors, `Plan`, `Explain`), `preview.go` (`Preview`, `Diff`, `Shadowed`), `schema.go` (`Fields`, `ActionsFor`, `Invariants`), `groups.go`, `stats.go`, `starter.json`/`starter.go`. Stdlib only; the only non-test import of another package is none (the test imports `internal/tidy`). Tests V (offline, `go test -p 1 ./internal/lootrules`): parse errors by path and line, round trip, first match with all/any/not, unidentified and `onUnknown`, ground fate to pickup, invariants (equipment, marks, start kit, quest, tome, rune, unique, drop gate with and without `confirm` and `ack`, no equip), plan cap order and determinism, class and build overlay (Amazon and others, unknown class), requirements (Paladin and Sorceress heroes, a Desert Guard and a Rogue Scout merc, level 22 vs 24, class orb, unread actor stays pending, fingerprint), durability, schema/preview/diff/shadow, ask, owned, rune number, the Starter against tidy (432 decisions), Starter ground spot checks. Benchmark above. Not done (U): the ground half against the real pickit with game tables (L2), the id-to-name stat table against live stat ids, `MercTypes` data, the API.

## Proposed text for the guarded files (not written, per the guard rule)
- `GLOSSARY.md`: "Loot rule set: the ordered, versioned list of rules (docs/lootrules.md) that decides pickup, keep, stash, sell, drop, identify for every item; the Starter preset is today's behaviour."
- `docs/decisions.md` (index line, for the coordinator or the user): "Item decisions move to one rules engine (`internal/lootrules`); the bot never gains an equip action; dropping or selling valuables needs a confirmed rule plus an acknowledgement (2026-11-14, proposal, not a decision yet)."

## 11. Wave 1: the shadow stage (cards lootwire-L2, L3, L4)

Tags: V = offline test named; U = not verified live (no game ran). Nothing in this wave changes behaviour: the engine is asked after the old code has decided, the answer is compared and logged.

### What is wired
- L2 hireling wear (`internal/merc/wear.go`, data in `wear.json`): `merc.WearTypes(typeChain, twoHanded)` lists the hireling kinds (`rogue desert ironwolf barbarian`) that can wear an item, `merc.CanWear(kind, ...)` asks one (an unknown kind is an error naming it), `merc.TypeChain(db, type)` walks the item type tree. The table is `merc.Allowed` written as data keyed by item type (V `TestWearTableMatchesAllowed`, `TestWearTypesTwoHirelings`); `merc.WTypeGaps(hirelingRow)` cross-checks the weapon types against the Hireling.txt `WType1/2` columns (U: not run against the real table in this wave). Javelins, axes, two-handed swords and Druid pelts stay refused (U, as in the merc policy).
- L3 adapter (`internal/loot/rulesadapt.go`): `Looter.Rules()` gives a `RulesAdapter` with `Item(unit, ItemOpts)` (code, name, type chain, quality, ilvl, requirements, class restriction, merc types, rune number, gem grade, base tier, sockets, durability, quantity, vendor value, stat map, mark), `ActorsOf(snapshot)` (hero from the player, merc from `Snapshot.Merc` plus Hireling.txt; a merc whose type or numbers are unread has level 0, so requirement rules answer UNKNOWN) and `Ctx(snapshot, build, packPct, stashPct)` (class, difficulty, act, level, gold, fill, `Owned`, `Needs` = the looter's own `needsOf` plus `emptyslot`). Unreadable fields stay hidden, never invented: a d2d ground label (item level 0) gets `HILvl|HIdentified`, plus `HValue` when the quality is magic or better; an unreadable stat list gets `HStats`; an unmapped quality gets `HQuality`. Stat ids: S from the pickit and the player stat list for life, resists, FCR, FHR, MF, skills, sockets; U for mana, FRW, IAS, GF, ED, durability (72, 73), and the gem grade from the code letters. V `TestAdapterItemFacts`, `TestAdapterBaseDerivedFields`, `TestAdapterActorsTwoHeroesAndMerc`, `TestAdapterCtxNeedsAndOwned` (a synthetic item table, two classes, two hirelings).
- L4 shadow (`internal/loot/lootshadow.go`): one `Shadow` per process (`loot.DefaultShadow`, Starter preset) fed from three decision points through `Looter.ShadowPass(snapshot, packPct, stashPct)`: `ground` (`Looter.candidates`, the verdict after the strictness, minimum priority, sell-value and small-gold-pile filters, before the pack-room check), `tidy` (`town.tidyPlan`, `tidy.Decide` per pack item) and `sell` (`town.sellableIf`, sold or not). The ground rules side applies today's strictness knob to the rule's tier (`ShadowPass.SetGround`), because the old pickit filters by tier after the verdict and the engine only carries the tier. Every disagreement is one line `loot shadow: <scope> <item> today=<action> rules=<action> rule=<id> reason=<reason>` (dedup by scope, area, item code and quality; at most 20 lines a minute, the rest only counted), and the counters are served read-only at `GET /api/lootrules/shadow` (`watch.lootshadow/1`: totals per scope, top 10 rules by disagreements with an example line, the last 50 lines, average and maximum cost in microseconds, error count). The knob `bot.lootShadow` (0 off, 1 on, default 1) is log only: `Observe` recovers from a panic and counts it, the engine never acts and the reflex layer is not touched. Stat lists are read only where a rule can use them (gold, magic or better, bases that can take four sockets), and on the ground the read EvaluateWith makes anyway is reused (one stat read per ground item, as before).
- Latency (V `TestShadowBudget`, the offline box): about 6 to 9 microseconds per decision including the adapter, 420 fixture decisions per run, the budget is 1 ms; the per-pass context (hero, merc, needs, owned counts) is built once per snapshot.

### Shadow findings (offline, nothing live)
The comparison runs over a synthetic fixture set (every base of a small item table in eight qualities, identified or not, with and without affixes, two classes, pack and stash fill levels, poor and rich). The repository's recorded item fixtures hold five items (`internal/d2dsrc/testdata/dump_items_late.json`: a worn wand, a belt potion, a scroll, a stash cap, a cursor javelin) and decide nothing the table does not; the sandbox logs were not read.
1. Pack stage (tidy) is EQUAL: 0 disagreements in 5040 decisions with the Starter (V `TestShadowStarterPackStageNoDisagreements`, marks, never-sell items and pack fill 40 and 90 included).
2. Ground stage is EQUAL once the strictness knob is applied: 0 in 1680 (V `TestShadowStarterGroundStageEquivalence`, memory-style items) and 0 in 1680 on d2d labels (V `TestShadowStarterGroundLabelEquivalence`: no item level, identified flag or stat list). Without the strictness fold the same comparison shows 175 disagreements, all "today picks up, rules ignore": the old low-gold mode (strictness 1 below `loot.lowGoldBelow`) takes tier-B sell rules (magic and rare items, low gems), which the engine's ground rule leaves on the floor. Finding: wave 2 must carry the strictness knob (a tier floor) into the ground stage or those pickups silently disappear. What stays outside the shadow because it is not a rule: the pack-room test, the value-per-cell test in low-gold mode are applied by `candidates` before the observation, so a skipped item counts as "today ignore" while the rules would say pickup; this is the real class of ground disagreements to expect live (U: no fixture models it).
3. Sell stage disagrees, as docs/sellskip.md section 1 predicted: 250 of 840 decisions (V `TestShadowSellStageFindings`, Sorceress, pack 40 and 90). The sell chore sells what the pickit verdict says (sell, or drop while the pack is full and the tier is not S); the Starter pack rules sell what the tidy table calls sellable (value 20 or more). Kinds: 114 today keeps, rules sell (`t-sell`: sellable gear the pickit wants, tier S or A); 72 today sells, rules keep (`default`: value under 20 and no rule); 30 today sells, rules stash (`t-gems`: surplus low gems); 24 today sells, rules would drop or stash (`t-junk`); 10 today sells, rules stash (`t-rare-jewelry` 8, `t-magic-jewelry` 2: sold unidentified or surplus jewelry). Finding: switching the sell chore to `Plan(Sell)` changes behaviour in exactly these five kinds; wave 2 decides each (the engine is the better written table, but the chore also has the identify, rejuvenation cap and spare-surplus rules, which are not rules yet).
4. Not testable offline: how often the real game produces these kinds; the stat ids marked U in the adapter; `WTypeGaps` against the real Hireling.txt. Live check list for the worker with the game tables: run `go test ./internal/town -run TestTidyAndSellFeedTheLootShadow` and the loot tests with `D2_MPQ_DIR` set (they skip without game tables), then a short soak and read `GET /api/lootrules/shadow` and the `loot shadow:` lines.

### Wave 2 needs
Per stage switch-over behind `loot.rulesEngine` (ground first: strictness tier floor, small-gold and per-cell sell filters as rules or limits; then tidy, which is already equal; the sell chore last, after deciding the five kinds above), a decision cache keyed by `Ctx.Fingerprint()`, the persistent store and API of section 6, and the Loot Rules tab (card L1 of section 9). Section 9's card names L2 to L7 are the design numbering; the wave-1 cards are lootwire-L2 (hireling wear), lootwire-L3 (adapter) and lootwire-L4 (shadow stage).

## 12. Wave 2a: the switch, the store, the sheet (card lootwire2)

Tags: V = offline test named; U = not verified live (no game ran). Everything is additive and OFF by default (`loot.rulesEngine` 0).

### 12.1 The knob `loot.rulesEngine` (V, `internal/loot/rulesengine.go`, `rulesengine_test.go`)
| value | what decides |
|---|---|
| 0 (default) | today's code everywhere; the shadow stage (`bot.lootShadow`) only compares |
| 1 ground | the engine decides what to pick up from the ground |
| 2 ground and tidy | plus the stash and drop decisions of the tidy chore |
| 3 | RESERVED for the sell chore: needs a user decision on the five sell kinds below, NOT implemented, behaves like 2 |

- Ground: `Looter.candidates` asks `ShadowPass.GroundVerdict`, which turns the engine decision into today's `Verdict` (a `pickup` is Keep with the rule's tier and prio, an `ignore` whose fate is `sell` is Sell, everything else is Drop = leave). Every filter after it stays today's code and applies to the engine verdict unchanged: the strictness tier floor (`loot.pickStrictness`, the old low-gold mode), the minimum priority, the per-cell value filter of low-gold mode, the small-gold-pile filter, the pack-room test and the by-name override. So the "strictness knob carried into the ground stage" of the wave 1 finding is exactly `passes()`, not a copy of it. V `TestEngineGroundStageEqualsToday`: 6720 decisions (the fixture set, two classes, poor and rich, strictness 0 to 3), 0 differ from today's pickit.
- Tidy: `town.tidyPlan` asks `ShadowPass.TidyDecision` per pack item and builds the same plan (stash first, best value first, then drops, cheapest first, capped by `town.tidyMaxActions`) with `planDecided`. V `TestEngineTidyStageEqualsToday` (840 decisions, two classes). The tooltip text (`TidyWhyFor`) and the count line of `tidyPass` still come from the tidy table (the same answers with the Starter).
- The shadow keeps running at every value (old code against the active rule set), so a switched stage still logs `loot shadow:` lines when someone edits the rules away from today's behaviour. Sell is always compared, never switched.
- Disagreement guard (V `TestEngineFailureFallsBackAndLogsOnce`): the engine is asked for ONE decision at a time; a compile error (unknown schema, a rule set that does not compile for the hero) or a panic inside it is recovered, that decision falls back to today's answer, and one `loot rules engine failed; today's code decides this item` line per distinct message is logged. An item is never lost to an engine bug; the worst case is today's behaviour.
- `SetActiveRules(rs)` hands the player's set to the engine and the shadow (the store calls it at start and after every save).

### 12.2 The store and the API (V, `internal/lootrules/store.go`, `options.go`, `plain.go`, `internal/watch/lootrulesapi.go`, `cmd/d2bot/lootruleswire.go`)
- File `lootrules-<profile>.json` beside the other state files: atomic write (temp file, fsync, rename), the previous version copied to `.bak` first, `seq` incremented on every save, no secrets. A missing file is the Starter preset; a damaged one is moved to `.bad`, the Starter is used and `loadError` tells the page (a bad edit can never stop the bot).
- `GET /api/lootrules` (viewer, read only, `watch.lootrules/1`): `{seq, active, plain:{ruleId: sentence}, protected, presets, classes, options, shadowed, recent, stage, loadError}`.
- `PUT /api/lootrules` (answer listener, same gates as the other control POSTs: origin list, token, JSON only; body limit 256 KB because a set is bigger than an answer): 200 `{ok, seq}`; 400 `{errors:[{path, msg, line, col}]}` (syntax errors carry line and column, field errors the path `rules[3].action`); 409 `{seq}` when the stored set changed since the client read it; 413, 415, 401, 403 as elsewhere.
- Presets: `starter` (default) and `class:<Name>` for every hero class (the Starter with only that class's rules and the class tags removed, so the player can edit it freely; the class list comes from `d2.Class`, no list in the store). PUT with `preset` set (and optional `options`) rebuilds the rules from the preset and the knobs and ignores the rules in the body; `preset: "custom"` takes the rules as sent (reorder, switch off). V `TestPresetsPerClass`, `TestStorePutOptionsRebuildsFromPreset`.
- The four knobs (`Options`, each edits named rules of the preset, so the result is an ordinary rule set): gear picked up only to sell (`g-magic-rare`: magic and better, rare and better, never), lowest rune kept (`g-runes`, `t-runes` get `runeno >= N`), keep socketed bases (`g-rw2..4`, `t-sockets` off), smallest gold pile (`g-spare-gold` gets `qty >= N`; a pile the bot needs gold for is still taken by `g-gold`). V `TestCustomizeOptionsEditNamedRules` (two classes).
- `Describe(rule)` writes the plain-language sentence of every rule ("On the ground: sell when gold on hand less than 300 and ...").
- "Why was this kept, sold or left": the shadow keeps the last 40 decisions with the matching rule id and reason (`Shadow.Recent`, repeats of the same item and answer collapsed); the sheet lists them newest first and marks the ones that differ from today's code.
- Not done in this wave (U): per-hero store (one file per profile for now), `POST /api/lootrules/preview|preset|undo|import|export|why` of section 6 (the `.bak` copy is there for undo), the recent-drops ring buffer of card L7 (the recent list comes from the shadow, which sees ground items only when the pickup runs).

### 12.3 The Loot Rules sheet (V node test and fit pages, U on a real phone)
`web/lootrulesui.js`, opened from the "..." menu (`Loot rules...`, event `d2:lootrules-ask`), registered with the overlay manager (one overlay at a time, Escape returns focus, docs/ui-overlays.md). Starter mode: the preset list and the four knobs, one Apply. Custom mode: every rule in plain language with an on/off box and Up/Down, one Save (it makes the set Custom). A refusal is shown in words with its line or field. The stage line says in plain words what the knob is doing now. Tests: `TestLootRulesUIJS` (stub DOM: nothing is sent before a press, the PUT bodies, the refusal text, Escape), fit pages `lootrules-*` at every fit size plus 412x780 and 360x640 (`tools/uishots.sh --fit`), fixture `tools/uishots/fixtures_lootrules.go`. Pictures at 1366x768, 412x780 and 360x640, dark and light, Starter and Custom, were looked at (not stored).

### 12.4 The five sell kinds: what the user decides (RECOMMENDED default each)
The sell chore sells what the pickit verdict says; the Starter pack rules sell what the tidy table calls sellable. They disagree in exactly five kinds (wave 1, 250 of 840 fixture decisions). Plain words, with the recommendation; `loot.rulesEngine` 3 stays unimplemented until these are answered.
1. Good gear the bot picked up on purpose (114 of 250; rule `t-sell`). Today the sell chore KEEPS it because the pickit wants it (an upgrade, a base worth a socket); the tidy table would sell it for its gold value. RECOMMENDED: keep (today's behaviour). Selling something the bot walked to and picked up as "good" a minute ago is a contradiction; the rules' `t-sell` gets a "not a pickit keeper" condition.
2. Cheap junk under 20 gold with no rule (72, `default`). Today it is SOLD at the next vendor visit; the rules leave it in the pack. RECOMMENDED: sell (today's behaviour), because every cell it holds is a cell a better item cannot use and the gold is real for a poor hero. A rule "sell anything worth at least 1 gold at a vendor visit" would express it.
3. Surplus low gems (30, `t-gems`). Today they are SOLD; the rules stash them. RECOMMENDED: sell the surplus (today's), stash only the gems the build or the cube still needs (the existing `need gem` condition). The stash is small and fills with chipped gems otherwise.
4. Plain junk the tidy table would drop (24, `t-junk`). Today it is SOLD in town; the rules would drop it on the ground or stash it. RECOMMENDED: sell in town, drop only when the pack is full away from town (the tidy drop rule keeps that case).
5. Jewelry (10: rare 8, magic 2, `t-rare-jewelry`, `t-magic-jewelry`). Today unidentified jewelry is SOLD as is (no identify tome early), surplus too; the rules stash it. RECOMMENDED: sell while the hero has no Tome of Identify and no scrolls (today's reason), stash once it can identify; rare jewelry is the one kind worth the trouble.
Question for the user: accept all five recommendations (then the sell stage is a rule rewrite plus value-0 checks, one more wave), or choose per kind.

### 12.5 Live checklist (not run)
1. Set `loot.rulesEngine` 1 on a throwaway hero, play 15 minutes, read `GET /api/lootrules/shadow` and the `loot shadow:` lines: the ground disagreement count must stay near the wave 1 number (0 with the Starter on plain drops).
2. Set it to 2, open the stash, check one tidy pass stashes runes, gems and charms as before (`tidy: plan` log line).
3. Edit a knob in the sheet (lowest rune 5), see El stay on the floor, set it back; PUT a broken set by hand and see the 400 with the line.
4. Delete the state file and check the Starter comes back; corrupt it and check `.bad` appears and the page shows `loadError`.
