# Quest board: quests as events the player takes part in

> status: design proposal (2026-11-16, card questdesign, TASK-575), not built; waits for the user's answers (section 9). Docs only. Direction: memory `d2-quest-design-2026-11-16` (user, 2026-11-16). Tags: V = read in this tree (file named), I = my inference or design choice, U = unverified (game fact or behaviour not checked live). Classic 1.14d only.
> It replaces, when built: the Journey chapters and strip (docs/campaign.md), the quest-given-by-planner driving (internal/task Campaign plan), and the "Pay and go" charter on the Journey sheet. It keeps: the quest handlers (internal/task, internal/quest), the evidence tiers (docs/progression/quests.md), the survival margin (docs/survival/design-2026-11-14.md), the cockpit frame (docs/ui/redesign-2026-11-14/J-second-attempt.md), the overlay policy (docs/ui-overlays.md).

## 0. What is wrong today, and the one idea

Today the hero levels, and a planner decides when a "chapter" runs (V docs/campaign.md sections 3 to 5: spine, bands, Auto/Ask/Skip/Hold intents, `level.N` steps titled "Level up to N for <chapter>"). The player sees "chapter 3 of 9" (V J-second-attempt.md: Journey strip cell), a dotted path, a band word and Go now / Hold; quests start without the player and a price ("charter", V docs/tally-strip.md) appears when an order is above range. Nothing is an event: no start, no end, nothing to watch, nothing to hand in, and the only direction is a level number.

The one idea: **the quest is the unit of play.** A quest has an offer, a tap to start, a live run the player watches, a stop at the NPC, a tap to hand in, a payoff. Between quests the hero is in a mode the player picked. Everything else (levels, zones, farm, loot) is what the hero does to get ready for the next offer or in the gaps.

The loop:

```
 (mode: train / idle / farm) --ready--> OFFER --tap Start--> ACCEPTED --> IN PROGRESS (steps, watched)
        ^                                                                     |
        |                                                          bot walks to the NPC and STOPS
        |                                                                     v
   next offer <---- DONE (payoff) <---- tap Hand in <---- AWAITING HAND-IN
```

## 1. The model

### 1.1 Quest definitions are data

One file per act and difficulty class: `data/quests/<act>.json`, schema `quests.def/1`, loaded and validated by a new `internal/questboard` (I). The existing `data/quests/act1-hints.json` and `act1-givers.json` (V docs/quest-hints.md, docs/quest-givers.md) are merged into it (they already carry giver, level, available-when and lines per quest); the old files become generated views or are removed (section 5). The spine `data/campaign/normal.json` (V docs/campaign.md) is retired; its `levels`, `checks`, `reward` and `why` fields move here.

A quest row:

| field | meaning | example (Act 1) |
|---|---|---|
| `id` | the task-layer quest id and save index (V quest ids in docs/progression/quests.md section 2) | `sisters_burial` (index 2) |
| `title`, `blurb` | player text: what it is, one sentence | "Blood Raven" / "Kashya's rogues are lost; kill their captain" |
| `act`, `difficulties` | where it exists; some are skipped by policy per difficulty (V quests.md section 4: Den, Cain, Bird, Blade, Tools are time-skippable in NM/Hell) | `[normal, nm, hell]` |
| `kind` | `story` (required for the next act: Andariel, Duriel, Mephisto, Diablo, Baal and the Req quests of docs/goals/execution.md section 3), `optional` (reward quest), `unlock` (act hop, Cain's town visit) | `optional` |
| `requires` | quest ids or chunk ids that must be done (hard, game-enforced) | `[den_of_evil]` (soft, see below) |
| `softAfter` | recommended order, never a lock | |
| `givers` | NPC to accept from, NPC to hand in to, with the dialog pick (V `internal/talk/tables.go` rows) | accept Kashya `Talk`, hand in Kashya `reward` |
| `chunks[]` | the pieces, section 1.2 | |
| `items` | required quest items and where they come from (pickit must take them: V quests.md act 4 row, loot has a `quest` field) | |
| `boss` | the encounter that gets a readiness word (section 3) | `blood_raven` |
| `rewards[]` | what lands at hand-in: skill points, stat points, item, sockets, merc, act access; marked `repeats: true/false` per difficulty | Den: 1 skill point |
| `meta` | meta currency points per quest and per chunk, per difficulty (section 7) | `{normal: 5, nm: 8, hell: 12}` U |
| `minutes` | expected duration range, for the offer card ("about 10 min") | `[8, 15]` I |
| `risk` | risk class: `routine`, `boss`, `hazard` (Council, Duriel poison, Andariel). Decides which survival encounter is the gate and whether auto-start may ever be offered | `boss` |
| `autoStart` | `allowed` or `never` (Rite of Passage and Seven Tombs style fights, and anything with `hazard`, default `never` until the user opts in with eyes open) | I |
| `done` | evidence rule (section 1.4) | |

Class and build specifics (V AGENTS.md "Build flexibility"): nothing about a class in Go. A `levels` row of the old spine is replaced by the survival margin of the `boss` encounter (V survival design section 7: "A chapter's band is computed from the margin of its boss encounter"); until S-B to S-D agree, a `legacyFloor` column per class family shows beside it (V survival design 7, the same transitional rule) and an unknown class is an error naming it (V the givers table already does this).

### 1.2 Chunks

A quest with more than one deliverable is cut into chunks. A chunk is the unit of Start: it has its own offer, its own watch, and a stop. Boundaries are where the game itself pauses the work (a hand-in, a portal, a town trip) or where a failure would make the player lose the least.

| field | meaning |
|---|---|
| `chunk.id`, `title` | "Fetch the Cube" |
| `order` | `after: [chunk ids]` or `any` (order does not matter, offered in parallel) |
| `ends` | `handin` (stops at an NPC, tap Hand in), `item` (stops in town holding the item, no NPC; the board says "Done: you now have the Staff of Kings"), `auto` (the game completes it, e.g. a kill that sets the bit: stops after the kill and the loot; no hand-in tap, the payoff is the toast) |
| `steps[]` | the handler steps shown on the strip (section 4.2): names for the player, mapped to existing handler steps (V `internal/quest/steps.go`, hops) |
| `minutes`, `risk`, `meta` | as above, per chunk |

Examples (I; the game facts are V docs/goals/execution.md section 3 unless U):

| quest | chunks | order |
|---|---|---|
| Den of Evil | 1 Talk to Akara and clear the Den (ends: handin at Akara) | one chunk |
| Search for Cain | A Get the Scroll of Inifuss (Dark Wood tree), B Read it with Akara, C Touch the stones and free Cain (Tristram), D Hand in Cain | A then B then C then D (hard chain) |
| Horadric Staff (Act 2) | A Cube (Halls of the Dead 3), B Viper Amulet (Claw Viper Temple, tainted sun), C Staff of Kings (Maggot Lair), D Assemble in the Cube and hand in to Cain | A, B, C `any`; D after all three |
| Khalim's Will (Act 3) | A Eye (Spider Cavern), B Heart (Kurast sewers), C Brain (Flayer Dungeon), D Flail (a Council member), E Cube them and smash the Compelling Orb at Travincal | A to D `any`; E after all |
| Rite of Passage | one chunk, never split: leaving the area resets the fight (V execution.md Act 5 row) | single |
| Terror's End | A Open the seals and kill the Vizier, de Seis, Infector (seal chunk), B Diablo | A then B (the seals reset if the game is left, U: do not split a half-done seal run) |

The Cain case the brief names: Cain's chain and the Staff and Will pieces are chunked by piece, as above; "quests that must be done together" means the board shows them as one quest card with a chunk list, never as unrelated cards (section 5.2).

### 1.3 States

Per quest and per chunk (a quest with one chunk shows the chunk state):

| state | meaning | entered when | evidence tier needed |
|---|---|---|---|
| `locked` | a hard prerequisite is not done | start, or difficulty change | T1/T2 of the prerequisite (Done on Save or Memory evidence only, V quests.md rule "skip only on T1/T2"); T3 never unlocks |
| `available` | prerequisites met, but the hero is not ready | prerequisites met and readiness word is `too hard`, or quest needs an item the hero lacks | T1/T2 for prerequisites; readiness is the model, not a quest bit |
| `ready` (the offer) | prerequisites met and readiness word is `safe` or `risky` | the board recomputes (every tick of the world model, and at town arrival) | as above |
| `accepted` | the player tapped Start | the tap (POST start) | the player's tap is the only evidence; stored as a plan record, not as progress (1.5) |
| `in progress` (step k) | handler running | `accepted` and handler started | step k is derived each tick from T1/T2 plus the inventory plus where the hero is (V quests.md "never persist handler progress") |
| `awaiting hand-in` | the objective is met and the hero stands at the NPC | the handler's last field step finished and the hero arrived; for `ends: handin` | T2/T1 "pending" bit if known (V `quest.Interpret` Pending, bit 1: U), else T3: the reward item or kill event (T3 is enough to STOP and ask for the tap, never to skip) |
| `done` | reward landed | the tap, then the handler's reward step | Done is `Confirmed` only on T1/T2; until the save or memory says so the card shows done with "confirming" (V `Result.Confirmed` in docs/quest.md) and is corrected at the next Save and Exit |
| `skipped` | the player chose to skip, or policy skips it in this difficulty | the player's Skip, or the difficulty file | the player's choice (a plan record); counts as done for soft ordering, never for hard prerequisites |
| `blocked` | a step failed 3 times, or a wall refuses (V docs/goals/execution.md section 5 and planner.md rule 5) | executor reports failure | Needs-you row (section 4.6) |

Done-by-someone-else: if T1/T2 says Done for a quest the board never ran (the character was played by hand, or an old save), the card is `done` and quiet: no payoff event, no meta award (section 7).

### 1.4 Evidence per transition

The detection matrix lives in docs/progression/quests.md section 3 (V); the board adds no new detection. Per transition:

- `locked -> available`: T1/T2 on the prerequisite quests, nothing else.
- `ready` offer: model only (readiness), never a quest bit.
- `in progress` step: the step is the first handler step whose own check fails (the handlers are idempotent, V execution.md section 1). T3 may advance the on-screen step label but a step marked complete by T3 alone is recheckable and never persists.
- `awaiting hand-in`: the handler reports "at NPC, nothing left to do in the field".
- `done`: T2 (immediate) else T1 after Save and Exit else T3 shown as "confirming". Tier of the strongest evidence is recorded on the event (V `QuestState.Evidence`).
- `failed`: Rescue on Mount Arreat can be failed by the game (V execution.md Act 5 row, quests.md `Failed`): state `blocked` with reason "failed in this game", no retry offered, next difficulty card is fresh.

### 1.5 What is persisted

Never handler progress. Persisted per hero (a small store beside the existing campaign state, I): `accepted` (quest or chunk id, accepted-at, mode that was active), `skipped`, `autoStart` opt-ins, and the per-hero "offered/seen" flags for payoff toasts. Everything else is derived. After any restart the board is recomputed from T1/T2, inventory and the plan records, so a restart can only lose the live strip position, never the state.

## 2. The loop and edge cases

### 2.1 The run

1. Offer: board shows the card (section 3.3 for the words). Nothing starts.
2. Tap Start: a plan record `accepted` is written; the executor takes over from the mode loop. Orders from the player still come first (V docs/campaign.md section 4, the priority ladder: safety, then player orders, then the plan); safety walls stay (V: Monastery road guard, tower guard, kit gates, chicken rules are not removed).
3. The bot does the chunk's steps, with the strip live. When the objective is met it travels to the hand-in NPC and stops next to him. State `awaiting hand-in`; a Needs-you row "Hand in to Akara" (not an alarm: tone `event`, section 4.6).
4. Tap Hand in: the hero talks, takes the reward (V Akara reward flow: the first click pays, dialog table in docs/quest.md). Payoff (section 4.4). Next offer appears or the mode resumes.

If the player never taps Hand in, the hero stays at the NPC and idles in town (I); it is safe in town. After `quests.handinIdleMin` (knob, default 30 min, I) the row escalates to a push notification only if the player has push enabled; the bot does not hand in by itself unless auto-start opted in for that quest (2.6).

### 2.2 Death, restart, reset

- Death mid-quest: the hero revives per the existing death policy; the quest stays `accepted`. The board shows "interrupted: died at Blood Raven" and the step restarts from the derived state (T1/T2 plus inventory). If the death count for that boss in this attempt reaches the retry cap (knob `quests.retryMax`, default 2 deaths, I) the quest becomes `blocked` with a Needs-you row "Blood Raven killed me twice: [Try again] [Train first] [Skip]". Training first sets the mode to train and clears the accepted record; the quest returns to `available` or `ready` by the normal rule.
- Process or game restart: derive from T1/T2 plus inventory; the `accepted` record survives, so the hero resumes after the restart without a new tap (the player already consented to this run). A restart that lasts longer than `quests.resumeWindowMin` (I, default 60) marks the card "paused" and asks again with a Needs-you row, because the player may have walked away.
- Player pause or Reset: pause stops the strip at the current step; Reset (V the Reset confirm in docs/ui-overlays.md) clears nothing in the board except the live run; accepted stays until the player cancels the quest from the card ("Stop this quest", back to `ready`).
- Player edits the order list or sends "Town": the hero obeys (priority ladder). The quest remains accepted and resumes when the order ends, unless the order was "Stop this quest".
- The bot leaves the game (V `launch.Flow` Cycler, new game, crash, "Save and Exit"): nothing is lost because nothing was persisted but plan records. A boss that resets on leaving (Ancients, seals) restarts at its chunk start (the chunk boundary exists to bound that loss).
- Save lag: T1 lags; "done" via T3 is shown but not counted as done for unlocking later quests until T1/T2 confirms (V quests.md rule). A Save and Exit is requested at the hand-in payoff moment when T2 is not available, so the next offer is not stuck (I; the lab already writes saves on exit).

### 2.3 A step fails

Handler failures (V execution.md section 5): item not found, NPC dialog option not found, wrong stone, boss reset. Policy:

| failure | board behaviour |
|---|---|
| step fails once or twice | retried silently; the strip shows "again" |
| fails 3 times | `blocked`; Needs-you row with the reason in words and [Try again] [Do something else] [Skip this quest] |
| a wall refuses (kit missing, no merc, route guard) | not `blocked`: the card goes back to `available` with the reason in words ("needs 3 antidotes: buying them"), because it is a readiness fact, not a failure (V survival design 7: kit is part of the margin) |
| a quest item is lost | step re-entered (V execution.md section 5); if unrecoverable, `blocked` "item gone: skip or restart the chunk" |
| quest failed by the game | `blocked`, no retry (2.4) |

### 2.4 Gold, charter and costs

Quest costs are real gold the hero spends (identify, imbue, a Larzuk socket, antidotes), not a toll. The charter ("Pay and go", V docs/tally-strip.md "Sink: charters") exists to price orders above range; with quests as events there is no pay-to-go: a quest is offered when ready, so a charter has nothing to do on the board. Plan: the charter is retired from the board and from quest starts. Whether it stays for zone orders (orders sheet) is a separate decision (open question 5). A quest that needs gold (imbue, Cain identify, antidotes, merc hire) lists it on the card ("costs about 500 gold at Charsi") and the readiness word includes "needs gold" as a reason; the hero spends only the listed cost, the board does not hide it.

### 2.5 Quests the player never starts

Do they gate the Acts? Only `story` quests do, and only because the game does (V execution.md section 3 "Req" column: Andariel, Duriel/Horadric Staff, Khalim's Will/Mephisto, Diablo, Prison of Ice, Nihlathak, Rite of Passage, Baal). `optional` quests never gate anything. Consequences:

- A story quest the player never starts blocks the act hop. The board shows the next act as "locked until: Andariel" in the offer list; it is never started by itself (default). There is no hidden planner driving the hero toward it.
- A hero in `train` mode levels toward the readiness of the next unstarted story quest (the "next quest" is the lowest-order available story quest, then the optional quest with the best reward per risk). Train never starts the quest.
- With the opt-in auto-start (2.6) the bot can continue unattended; the bot "must still be able to progress when the user opts in" because the engine for it exists (the handlers and executor are unchanged), only the tap is replaced by a stored consent.

### 2.6 Opt-in auto-start

Per quest (and, I recommend, per difficulty: opting in for Normal does not carry to Nightmare). Default off. Stored as a plan record `autoStart: {quest, difficulty, since}`. Semantics:

- An auto-started quest still stops at the NPC and waits for the tap if the player is present (a client connected in the last N minutes, `quests.presenceMin` default 10, I). When the player is away the bot also hands in itself. The payoff event is kept for the return screen ("While you were away: finished Den of Evil, handed in at 14:20").
- Auto-start only fires when the offer is `safe` (never `risky` or `too hard`) and the quest's `autoStart` is `allowed` (1.1). `risky` offers wait for the tap even with opt-in on (I; open question 4 asks whether this should loosen).
- Changing the opt-in is one tap on the card (section 4.5); a chunk of a quest inherits the quest's setting unless set per chunk.
- Auto-start never skips a hard prerequisite and never starts two quests at once.

### 2.7 Difficulty changes, multiplayer

- Difficulty: each difficulty has its own quest bits (V quests.md section 4). A new difficulty is a fresh board: all quests of that difficulty start `locked` or `available` per the same rule, optional-skipped ones per the difficulty file. The mode loop shows "Nightmare: Act 1" and the quest list resets without losing the Normal history (the done list is kept per difficulty, section 4.4).
- Resist penalty after the switch is in the survival margin, so readiness words fall by themselves (V survival design section 7 and docs/survival wiring).
- Multiplayer is out of scope (the bot runs offline single player; V AGENTS.md).

### 2.8 Cain and "must be done together"

Search for Cain's Tristram part, the Staff pieces and Khalim's Will pieces are chunked (1.2). The board offers the chunks in the order the data says, as one card with a chunk list, the offer for each chunk having its own readiness word. A piece that leaves the hero holding a quest item is protected: the item cannot be sold or stashed away by the loot rules while its quest is `accepted` or any chunk is `done` but the quest is not (I: a loot rule `keep: quest item`, V loot has a `quest` match field per quests.md).

## 3. Readiness

### 3.1 What "ready" means

An offer is `ready` when all hold (I, built from V surviving pieces):

1. Prerequisites done (T1/T2).
2. Encounter margin (survival model, V docs/survival/design-2026-11-14.md sections 5 and 7): the margin of the chunk's `boss` encounter and of the zone on the way, against the player's risk budget (Cautious, Balanced, Bold). Words: `safe` = above the dial's enter line (comfortable band), `risky` = between the Bold line and the enter line (ready band), `too hard` = below the Bold line. The three bands are exactly the old too early/ready/comfortable bands of V docs/campaign.md renamed for the player.
3. Kit: items are part of the margin, but "can the hero afford the cost" (gold, scrolls, antidotes) is a separate yes or no in words.
4. Merc: counted inside the margin (alive adds to the margin, V survival design 6.4); the card says it when it is the swing factor.
5. Level is NOT a gate. It is a term in the margin; the card never says a level number as a reason ("you need 12" is gone) except where the game enforces one (Rite of Passage 20/40/60, V quests.md). Where the game enforces a level the card says so as a fact.

### 3.2 The words, never a number

The board shows the word `safe`, `risky` or `too hard` and one reason sentence built from the weakest margin term (V survival design 7 already produces: "Andariel: margin 0.9 (needs 1.5). Poison takes 45 percent ..." which the board strips of numbers). Reason templates (I, closed list of codes in one file, like the survival `why` codes):

| word | example reasons |
|---|---|
| safe | "You out-fight her comfortably." "You have a living Rogue and a full belt." |
| risky | "Close: one more level or a full belt makes it safe." "Poison would hurt: three antidotes would help." |
| too hard | "She would likely kill you. Train first." "The Rogue is dead: revive him." (a fixable cause is named first) |

No number, no percent, no margin value, no "needs 1.5" on the board. The API carries the margin for tools (debug view), the UI never renders it (open question 8 asks if a "Details" popover may). Fit rule check: reasons are one line ellipsis with the full text in the popover (V J-second-attempt section 2 rule "the reason is cut with an ellipsis; the full text is in the popover").

### 3.3 What an offer shows

Name; why the bot thinks it is ready (the reason sentence); the word; what it unlocks ("Act 2 and a free Rogue"); what it pays (reward and meta); expected time; costs; the chunk it would start. A `too hard` quest is listed, greyed, with its reason and a "Train toward it" button that sets the mode to train and pins the target (the player can make the unreachable a goal; this is the "direction" the user wanted).

### 3.4 Modes between quests

Three modes, always visible on the strip and on the board (section 4.7):

| mode | the hero does | exits |
|---|---|---|
| Train | levels toward the next offer: runs the survival zone chooser (V zone model: best zone whose margin clears the budget) with the target set to raise the target quest's margin; the bot also buys the listed kit. When the quest becomes `ready` the board raises the offer (toast "Blood Raven is ready"), the mode stays Train until the player starts or changes it | the offer appears; the player taps Start, or switches mode |
| Idle in town | stands in town (V Rogue Encampment idle), no field work, no XP; does chores that the town loop does anyway (repair, potions, identify, sell per loot rules). The offer list is kept current | switch mode, tap Start |
| Farm | the zone model's farm loop (V survival 5.3 zone choice) with the objective loot/gold (rune zones, gold-rich zones; the loot rules keep their effect, V docs/lootrules.md, and the zone model's `lambda` term raises weight on gold); it does not target the next quest's margin | switch mode, tap Start |

Interaction rules:

- The mode is the default behaviour whenever there is no accepted quest. Accepting a quest suspends the mode; finishing it (or stopping it) returns to the mode.
- Train is the default mode for a new character (so the offers come), unless the player picked otherwise (open question 3).
- The zone model and the farm loop are unchanged: the mode only sets their objective (xp toward the target, or gold/loot, or none). Safety rules and walls stay (V survival design). The Orders model (V C-orders, J) remains for player orders: an explicit Go-to order beats the mode, as before.
- The "plan row" in Orders (V "then the plan: level 10 ...") becomes the mode line ("Train toward Blood Raven").
- Loot rules (V docs/lootrules.md, loot-rules-ux.md): unchanged in all modes; Farm only raises the number of drops the rules act on. The planned loot vault (separate design) is a consumer of Farm mode, not part of this design.

## 4. UI

Respect: the cockpit frame (fixed budgets, no content-driven height, V J-second-attempt.md), the overlay policy (one transient overlay at a time, Escape closes the last, top-layer menu; V docs/ui-overlays.md), no number margins, text wraps in sheets, one accent use per region, the supported sizes (1920x1080, 1440x900, 1366x768; windows 1914x930, 1904x1000, 1536x730, 1280x620; phone 390x844).

### 4.1 Slot map (what replaces Journey)

Cockpit template slots stay (`band`, `stage`, `strip`, `rail`, `railWide`, `phoneTabs`; V J-second-attempt.md section 3). Changes (I):

| today | becomes | notes |
|---|---|---|
| `journey-strip` tile in `strip` (nine chapters, "chapter 3 of 9", Go now, Hold) | `quest-strip`: three-state tile. No quest: the offer teaser (the best offer, its word, [Start]). Running: the quest strip (4.2). Awaiting: the Hand in button | same 112 px cell; the old tile stays in the free-grid layouts only until retired |
| `journey` panel and `journey-chip` | retired | the chip in the band becomes the needs-count only |
| (new) | `quest-board` sheet | opened from the quest-strip tile ("All quests") and the "..." menu; a sheet over the rail, never over the picture (overlay policy) |
| Orders "plan row" | the mode line | one line, a Mode switch lives on the quest-strip and the board header |
| Needs you | gains quest rows | 4.6 |
| `phoneTabs` Progress | `[progress, quest-strip, health]` | and a fourth tab `Quests` (the board) on the phone: Orders, Map, Quests, Progress; no page scroll |

Tile names: `quest-strip` (slot), `quest-board` (sheet panel `fits: "sheet"`), `quest-mode` (the mode switch, a control inside the strip and the board header, not a separate tile), `quest-toast` (the payoff, uses the existing toast layer).

### 4.2 The quest strip on the video card

Under the picture (strip row, 112 px, or 92 on short windows), left to right, when a quest runs. No chapter count, no percent bar; the steps are words.

```
1440x900 strip cell (about 456 wide of the rail-aligned strip; shown with the picture's width)
+------------------------------------------------------------------------+
| BLOOD RAVEN   Kashya's rogues                              [Stop] [Mode]|
|  (1) Go to the Burial Grounds  (2) Kill Blood Raven  (3) Hand in        |
|        done                         NOW: kill her        next: Kashya   |
|  now: Burial Grounds, Blood Raven's tomb        you can watch above     |
+------------------------------------------------------------------------+
```

Fields: quest title; chunk title; the step dots with the current one highlighted (one accent: the current step); "now" line = area and what it is doing (from the status sentence); "next" = the next NPC or area; [Stop] cancels the run (a confirm, state back to `ready`); the mode switch on the right. Steps come from the chunk's `steps[]`; more than 4 steps show "2 of 5" with the current one named (words, no bar).

At 1366x768 (92 px strip): the dots row and the now line; the chunk subtitle drops. At 390x844: the same cell sits in the Quests tab above the board list, a two-line header; the band stays two lines.

States of the same cell:

```
Offer (no quest running):  [ BLOOD RAVEN  safe  "You out-fight her."   unlocks: free Rogue ]  [Start] [All quests]
Awaiting:                  [ Kashya is waiting: hand in Blood Raven ]  [Hand in]  (the picture shows the hero at Kashya)
```

### 4.3 The Quest board sheet

Desktop (1914x930 window, sheet over the rail's lower two thirds and the strip; the picture stays visible):

```
+------------------------------------------------------------ QUESTS ---------- Act I  Normal  [Train|Idle|Farm] [x] +
| NOW     Blood Raven  safe        step 2 of 3: kill her        [Stop]                                              |
+--------------------------------------------------------------------------------------------------------------------+
| READY                                                                                                              |
|  Blood Raven      safe    "You out-fight her."                 pays a free Rogue   about 10 min   [Start]  auto o  |
|  Find Cain  (3 parts)     risky  "Close: a full belt makes it safe"   pays the Cain visit   about 25 min [Start] o |
| NOT YET                                                                                                            |
|  Andariel         too hard "She would likely kill you. Train first"    unlocks Act II        [Train toward it]    |
|  Countess  (optional)     too hard "Poison and a long dive"             pays runes                                 |
| DONE   Den of Evil (1 skill point)   Cold Plains   Tools of the Trade                                             |
| SKIPPED   Charsi imbue [Undo]                                                                                       |
+--------------------------------------------------------------------------------------------------------------------+
```

Rules: one list, four groups (Ready, Not yet, Done, Skipped); a quest with chunks opens an inline list of its parts, each with its own word and Start; rows are one line with ellipsis, full text in the popover; the "auto" cell is the opt-in toggle (4.5). The sheet scrolls inside; the header (act, difficulty, mode) is sticky; "Pay and go" is gone (the sticky element is the Now row). At 1366x768 the same sheet with one-line rows and no pays column (the pays text moves into the popover). At 1280x620 the Done and Skipped groups collapse to counts.

Phone 390x844 (tab Quests, full screen, no sheet):

```
+----------------------------+
| Act I Normal  [Train v]    |
| NOW Blood Raven  safe      |
| step 2: kill her  [Stop]   |
+----------------------------+
| READY                      |
| Blood Raven   safe         |
|  You out-fight her.        |
|  free Rogue, about 10 min  |
|  [Start]        auto: off  |
| Find Cain (3 parts) risky  |
|  [Start part 1]            |
| NOT YET                    |
| Andariel  too hard         |
|  [Train toward it]         |
| DONE (3)  SKIPPED (1)      |
+----------------------------+
```

### 4.4 The Hand in moment and payoff

When the hero stops at the NPC the strip cell turns to the Awaiting state and a Needs-you row (event tone) appears. Tap [Hand in]:

1. The strip cell changes to "Handing in..." for the duration of the dialog (the picture shows it live).
2. Payoff: a toast on the toast layer (z 30, V ui-overlays layer scale): "Den of Evil done. +1 skill point. +5 meta" (the reward line, in words, from `rewards[]`; the meta line only where it paid and only when the player has the meta wallet; section 7).
3. The tile: the quest row in the board flips to Done with the reward line; the strip cell shows the next offer immediately (the best `ready` quest), or the mode line when none.
4. The Chronicle gets a line (V the Chronicle already gets quest events) and the return screen lists it for a player who was away.

If the reward needs a choice (imbue item, Larzuk socket target, Anya's item) the same Hand in opens a question card first (V `t1.quest.imbue` question key pattern in docs/quest-hints.md), then the payoff.

```
toast (bottom center of the picture's edge, not over the picture centre):
 +-----------------------------------------------+
 |  Done: Den of Evil   +1 skill point   +5 meta |
 +-----------------------------------------------+
```

### 4.5 Opt-in auto-start toggle per quest

A one-tap switch on each Ready/Not yet row (the "auto" cell in 4.3), a tooltip "When you are away, start this by itself when it is safe". Turning it on shows a one-line confirm in the row ("Starts by itself when safe and you are away. [Keep] [Cancel]"). A quest with `autoStart: never` shows the cell greyed with the reason ("always needs you"). Off by default. The state is part of the API (`auto`).

### 4.6 Needs-you rows

Needs you takes 0 to 3 one-line rows (V J-second-attempt.md section 2). New rows:

| row | tone | buttons |
|---|---|---|
| "Kashya is waiting: hand in Blood Raven" | event (accent, not alarm) | [Hand in] |
| "Blood Raven is ready (safe)" | event | [Start] [Later] |
| "Blood Raven killed you twice" | blocked | [Try again] [Train first] [Skip] |
| "Step failed: could not find the Cairn Stone" | blocked | [Try again] [Do something else] [Skip] |
| "Quest paused: you were away" | event | [Resume] [Stop] |

An offer row is raised once per offer (not repeated after Later); more than three: "+N". The same events feed the push channel when enabled.

### 4.7 The mode switch

A three-segment control (Train | Idle | Farm), in the strip cell (right) and the board header, one tap each, with the current mode named in the status sentence ("TRAINING toward Blood Raven", "IN TOWN, idle", "FARMING Cold Plains for loot"). No confirm. On the phone the control sits in the Quests tab header. A mode change while a quest runs is allowed and takes effect after the quest (the control shows "after this quest").

### 4.8 Fit and overlay checks

The quest-board sheet registers with `overlay.js` like the zone sheet (V ui-overlays.md policy 1 to 3): one overlay at a time, Escape closes it and returns focus. The sheet text wraps (`overflow-wrap: anywhere`), the header is sticky and never above the band. `tools/uishots.sh --fit` pages to add (I): `quest-board-*` per supported size, `quest-strip-{offer,running,awaiting}`, `quest-toast`; the must-fail fixtures reuse the overlay set. No numeric margin, no "chapter n of m" string anywhere (a text check in the fit run, I).

## 5. Migration (retire plan)

### 5.1 What is retired

| piece | where (V) | plan |
|---|---|---|
| Journey strip, journey, journey-chip panels | `internal/watch/web/panels/journey*`, `watch/campaign.go`, tests under `internal/watch/testdata/campaign` | removed from the cockpit slots when the quest-strip ships (S5); panels stay for the free-grid layouts one release, then deleted |
| The spine, bands, intents, decision | `internal/campaign` (spine.go, band.go, next.go, controller.go, view.go), `data/campaign/normal.json`, docs/campaign.md | the band function is replaced by the margin words; intents Auto/Ask/Skip/Hold map to: Start (tap), `auto` opt-in, Skip, Stop; the package is kept until S4, then archived; docs/campaign.md gets a `> status: superseded` banner pointing here (V AGENTS.md doc rule) and is archived with a stub |
| Planner auto-driving of quests | `task.Campaign.WithGuide`, `internal/task/guided.go`, `internal/bot/guided.go` | the executor runs a quest only when the board hands it an accepted record; the planner stays for the modes (train/farm zone choice) and for the Act hop mechanics. In Free mode the bot does not drive quests at all |
| Quest-givers visit, hints in town | `internal/task/givers.go`, `internal/bot/givers.go` (V docs/quest-givers.md) | folded into the board's accept step: the bot no longer visits givers unprompted; it visits when the player starts that quest. The hint lines become the offer's reason/blurb |
| Charter / "Pay and go" on the Journey sheet | `internal/currency/tallyrows.go` charter sink, `watch/tallystrip.go`, `proposals.go` | removed from quests; stays for zone orders only if question 5 says so |
| API `/api/campaign`, `/api/campaign/intent` | `internal/watch/campaign.go` | kept as a compatibility read for one release (empty-chapter mapping), then removed |
| Old quest-driven proposals ("Go for Blood Raven? [Go][Not yet]" Needs-you rows from the proposals path) | `internal/watch/proposals.go` | replaced by the offer rows (4.6) |

### 5.2 What is kept

`internal/quest` (state reader, dialog tables, hops, handlers), `internal/task` handlers and executor, the evidence tiers, `questmem` live reader, safety walls and kit gates, Orders, the zone model, loot rules, tally strip credits (a `quest done` credit of 10 stays; it keys on `quest/<quest>/<difficulty>`, V tally-strip.md).

### 5.3 API sketch

New routes in `internal/watch/quests.go` (I; viewer side GET, answer listener POST like the campaign routes, V campaign.go):

| route | body or reply |
|---|---|
| `GET /api/quests` | schema `quests/1`: `{difficulty, act, mode, now: {quest, chunk, step, steps[], area, next, state}, offers: [{id, title, parts: [...], word: safe/risky/too_hard, reason: {code, text}, unlocks, pays: [..], minutes, costs, auto: {allowed, on}}], done: [...], skipped: [...], needsYou: [...]}`. No numeric margin in the player fields; a `debug` block (margin terms) only under the debug flag |
| `POST /api/quests/<id>/start` | `{chunk?}` -> 200 accepted; 409 with `{code: "locked"|"not_ready"|"busy"}` if another quest runs |
| `POST /api/quests/<id>/handin` | only valid in `awaiting`; question answers in the body for choice rewards |
| `POST /api/quests/<id>/auto` | `{on: true|false, difficulty}` |
| `POST /api/quests/<id>/stop` | back to `ready` |
| `POST /api/quests/<id>/skip` and `/unskip` | sets/clears the plan record |
| `POST /api/mode` | `{mode: train|idle|farm}` |

Every POST is idempotent by (quest, chunk, difficulty): a second start while running returns the current state; a second hand-in after done returns 200 with `already: true` and no second payoff or award.

### 5.4 Knobs (I; cited as `group.key`, defaults to be calibrated)

`quests.handinIdleMin` (30), `quests.retryMax` (2), `quests.resumeWindowMin` (60), `quests.presenceMin` (10), `quests.offerRepeat` (once per offer), `quests.autoStartMinWord` (`safe`), `quests.defaultMode` (`train`). They go in the code knob table so docs/knobs.md generates them (V "knob defaults live in docs/knobs.md generated from code").

## 6. Slices

Sizes S/M/L as in the backlog scheme. Each slice is one PR family, offline-testable on `mem.Fake`, two builds in tests (Paladin and Sorceress, V Build flexibility rule). Dependencies are real code dependencies.

| id | slice | size | depends | live verification |
|---|---|---|---|---|
| S1 | `internal/questboard`: definitions schema, loader, states, derive-from-evidence function, plan-record store; act 1 and 2 data (merging hints/givers) | M | none | none (offline); tests for each transition and restart (derive twice, same result) |
| S2 | Readiness words: the margin-to-word mapping, reason codes, `legacyFloor` side by side; needs the survival margin function (S-B) | M | S1, survival S-B | none until S-B lands; then a table of zone and boss words checked against the logged outcomes in docs/survival (Cold Plains clvl 4 to 5 safe, Black Marsh clvl 12 without a merc too hard, V survival design 5.4 calibration targets) |
| S3 | Executor hand-off: accepted record starts the chunk handler; stop at the NPC; hand-in action; retry, blocked, resume; remove the planner's quest driving behind a flag `quests.board` | L | S1 | a drill per Act 1 quest on a throwaway character (one hidden instance, `tools/live.sh`): Den of Evil, Blood Raven (start, watch, stop at Kashya, hand in), writes after start 0 |
| S4 | API and stores: GET/POST routes above, idempotency, tests | M | S1, S3 | curl against a stub, then on the live instance with S3's drill |
| S5 | Watch UI: `quest-strip` slot panel, `quest-board` sheet, mode switch, toast, Needs-you rows, phone tab; uishots fit pages and must-fail fixtures | L | S4 | fit renders at the 8 sizes; one live look by the user |
| S6 | Modes: train target pin, idle in town, farm objective in the zone model; interaction with Orders | M | S1, S3, zone model | a 30 min run per mode on a hidden instance; check the zone choice change and no field work in idle |
| S7 | Chunked quests: Cain chain, Staff pieces, Khalim's pieces, Terror's End seals; data plus handlers per chunk (handlers largely exist as quest steps, V `internal/task/act2*.go`, `act3*.go`) | L | S3 | per-chunk drills in Acts 2 and 3 (hard to reach: needs a character at Act 2/3, lab-created with the cheat ledger and tagged CHEATED for evidence use, V private rules) |
| S8 | Opt-in auto-start and away behaviour | S | S3, S5 | one overnight style run with one opt-in quest, writes after start 0 |
| S9 | Retire: remove Journey panels, spine, charter on quests, compatibility API, archive docs | M | S5, S7 | grep for removed strings in the fit run |
| S10 | Meta currency hooks wired (needs the wallet from the vault design) | S | S1, S4, vault wallet | ledger replay test; one live hand-in awards once |

Order: S1 -> S2/S3 -> S4 -> S5 -> S6 -> S7 -> S8 -> S9 (S10 when the wallet exists).

## 7. Meta currency hooks

User add (2026-11-16): an account-level meta currency (persists across heroes and runs, not hero gold) that upgrades things like the loot vault capacity (data-driven tiers). The wallet and its tiers are designed in the vault doc (agent/vaultdesign); this section is only the hook where quests earn it.

- Where the award shows: the Hand in payoff (4.4): the toast carries "+N meta" and the quest row in the board shows the amount under the reward line. Chunk completions with `ends: handin` or `auto` pay at their own payoff; a chunk with no payoff of its own (`ends: item`) pays when the chunk is done, toast "Cube found, +N meta".
- Which quests pay how much: data columns in the quest rows (1.1): `meta` per quest and per chunk, per difficulty, e.g. `"meta": {"normal": 5, "nm": 8, "hell": 12}` (values U, placeholders; to be tuned). The sum of chunk awards plus the quest award is the quest's total; rule I: story quests and boss chunks pay more than optional chores; Nightmare and Hell pay more because they are harder and slower. Meta is a column in the data, not a number in Go.
- Idempotent: an award is a row in an account ledger keyed `quest/<quest>/<chunk or "-">/<difficulty>/<character-or-account>`. Plan: the key is account-wide per (quest, chunk, difficulty), so a quest earns its meta ONCE per account per difficulty, not once per hero (open question 7: should each hero earn it again?). The award is computed from the evidence: when the board derives `done` for a quest from T1 or T2 (not T3), it asks the ledger to credit that key; the ledger refuses duplicates. So a restart, a re-derive, a repeated tap or a second POST handin cannot double-award; a quest done by hand before the board existed is derived `done` but is recorded with `source: backfill`, `amount 0` (no award), unless the user wants backfill (question 7). T3-only completion shows the toast "confirming" with no award until T1/T2 confirms, then the award lands; the toast then repeats once as "+N meta confirmed" only if the player is present (I).
- Reuse of the existing pattern: the same ledger approach as the tally strip (V docs/tally-strip.md: keys `quest/<quest>/<difficulty>`, replay books nothing, firsts and quests account-wide). The meta wallet should reuse that ledger or its key rules; the vault design decides whether it is a new source row in `registry.json` or a new ledger (not decided here).
- API: `GET /api/quests` includes per offer `pays: [{kind: "meta", amount: N}]`; `handin` reply includes `awarded: [{kind, amount, key}]` and `already: true` on a repeat. If the wallet is absent (before the vault design lands) the field is omitted and the hook is inert (the slice S10 is after the wallet).
- Skipping never awards. Auto-start while away awards at the same payoff (it appears in "while you were away").

## 8. Faithful-to-the-game limits (U, to verify live in S3/S7)

All of these are tagged U and are not changed by this design: the quest bit meanings (Pending bit 1; V live only for 0x0001/0x8001, 0x000c, 0x0010, 0x0004, 0x0008, 0x1000, V docs/progression/quests.md section 2); the exact dialog rows of Kashya, Charsi, Warriv, Cain (V docs/quest.md live checks list, not run); chunk boundaries inside Terror's End (whether leaving mid-seals resets the seals, U); whether Den of Evil's free respec row disappears (V docs/quest.md open U). The board shows what it derives; an unknown row degrades to `blocked` with a clear reason, never a guess.

## 9. Open questions for the user (8)

1. **Does a Start tap cover a whole multi-part quest or only the current chunk?** Recommendation: only the current chunk; the next chunk raises its own offer immediately, so a three-minute chunk costs one tap and a 30 minute risky chain never runs on one tap. A per-quest "run all parts" toggle can come later.
2. **Where does the board live: a sheet over the rail, or a permanent tile?** Recommendation: the strip tile is permanent (offer, run, awaiting); the full board is a sheet. The rail stays for Needs you, Orders and the map.
3. **Default mode for a new hero.** Recommendation: Train (offers appear); Idle and Farm are one tap away; the choice is remembered per hero.
4. **Should auto-start ever fire on a `risky` offer?** Recommendation: never by default (only `safe`); an advanced per-quest setting "also when risky" later, because a risky boss while the player is away is how a hero dies unattended.
5. **Keep the charter for zone orders, or retire it entirely?** Recommendation: retire it from quests now; leave zone-order pricing alone until the user says (the user disliked "Pay and go" on the Journey sheet; the Orders pricing was not mentioned).
6. **Hand-in: must the player tap it when present, even for quests with no choice (Den skill point)?** Recommendation: yes, the user asked for "taps HAND IN and watches the reward land"; auto-start with the player away is the only case the bot hands in itself.
7. **Meta currency: once per account per quest per difficulty, or once per hero? Backfill for quests already done?** Recommendation: once per account per difficulty (prevents farming new characters for points); no backfill award for old saves. Earn rates (the `meta` column) need a first guess from the user: recommendation as a starting point, a first Normal story quest about 5 and an optional chore about 2, so a full Normal playthrough lands in the low hundreds; tune after the vault tiers have prices.
8. **May the board's popover ever show the margin number (a Details view) for the player who wants it?** Recommendation: no on the board (the user said words only); the debug view keeps the numbers for us.
