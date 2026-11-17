# ETHER: the account meta wallet of the Ethereal Vault

> status: built offline 2026-11-17 (agents vaultv7 and vaultv7b, slices V7 and V7b of [design/loot-vault-2026-11-16.md](design/loot-vault-2026-11-16.md) section 2.6). Wired into the bot behind `loot.vault` (default 0) in V7b; NOT run live. Tags: **V** verified by a test in this repo, **U** unverified or a placeholder number.

User decisions 2026-11-16: the currency is called ETHER; it is account-level (survives heroes and runs, not hero gold); it is earned from quest hand-ins, "a first" rows, boss kills, milestones and a small capped kill trickle, NOT from pressing Keep; it is spent on vault capacity first (auto-pickup range and filter slots later, as more rows of the same kind).

## Why a new ledger file and not a `tally` entry kind

The `tally` ledger (`internal/currency`, [tally-strip.md](tally-strip.md)) is hash-chained, registry-driven and its events mode has an allow-list of kinds. ETHER needs a refund kind and tier ownership, which would touch the registry, `Tally.Credit` and the events allow-list that other agents own. So ETHER reuses the tally RULES (key per observation, replay books nothing, firsts and quests account-wide, a daily cap, a spend that cannot go negative) in a small separate store: `<profile currency dir>/ether.jsonl` (`ether.FileName`), one fsynced line per entry, stdlib only. If the user later wants one ledger, the entry shape maps onto a `currency.Entry` (kind, amount, key, source) and the earn table onto `tallyrows.json`.

## Ledger

`internal/ether`. Entry: `seq, ts, kind (credit|spend|refund), amount, source, key, step, note`. Everything is derived by replay at `Open`: the balance, the owned capacity steps (a `spend` with a `step` owns it, a `refund` of it releases it) and the per-step refund count. Nothing is stored on its own, so a crash cannot leave a tier and the ledger disagreeing.

- Keys are unique over the whole file; a duplicate key is refused at write and rejected as corrupt at load.
- A torn last line (a crash mid write) is dropped and the file cut back to the last whole line (`Repaired()` reports the bytes). Damage anywhere else, a seq gap, a duplicate key or a spend over the balance makes `Open` fail with `ErrCorrupt` naming the line (a hand edited wallet is a cheat; there is no import). Nothing is rewritten in that case.
- A failed write cuts the file back to its previous size and leaves the state unchanged.
- One process owns the file (mutex, no file lock). V (`TestCrashTornTail`, `TestCorruptMiddleRefused`).

## Earn sources (data: `data/vault/ether.json`)

| source | key | pays | note |
|---|---|---|---|
| `quest` | `quest/<quest>/<difficulty>` | the caller's amount, 0..100 (`maxHook`) | HOOK for the Quest board (S10, built): sub key `<quest>/<chunk or ->`, the `meta` column value for that difficulty. Account-wide: once per quest (or chunk) and difficulty; the board settles it in `Exec.HandIn`, see [questboard.md](questboard.md) section Ether |
| `first` | `first/<hero>/<kind>/<name>` | 5 U | "a first for this hero": the first kill of a monster kind (`monster`) and the first visit to a field area (`area`); a second hero earns its own |
| `boss` | `boss/<name>/<difficulty>` | 10 / 15 / 25 U | normal / nm / hell; once per account (a second hero does not earn it again) |
| `milestone` | `milestone/<id>` | per row U | `clvl10, clvl20, clvl30, clvl50, diff_nm, diff_hell`; an id not in the table is refused. `clvlN` pays when a hero reaches level N or more (the thresholds are the table's keys); `diff_nm` and `diff_hell` pay when a hero is playing that difficulty (reached, not cleared). Once per account |
| `kills` | `kills/<local day>/<k>` | 1 U per `per` (25 U) kills | capped at 10 per LOCAL day (`dailyCap`); `Wallet.Trickle` books the next free `<k>` of today, so a restart never rewrites a key and the cap holds; the unpaid kill count of a run is not saved (a restart starts the count again) |

`Wallet.Earn(source, sub, difficulty, hookAmount)` returns `{credited, key, already, capped}`. A repeated key pays nothing (`already`), so a journal replay books nothing (V `TestEarnIdempotentAndReplay`). Unknown source, missing or unknown difficulty, over-limit hook amount: an error naming it (V `TestEarnErrors`). Pressing Keep never earns. The Quest board hand-in (S10) is wired: it calls `Earn("quest", "<quest>/<chunk or ->", <difficulty>, <meta column>)` and reads `Paid(key)` for the pay chips (V `TestQuestLedgerOverWallet`, `TestEtherAwardOnceAndRepeat`).

## Capacity tiers and spending

Tiers are `data/vault/capacity.json` (base 400, steps v1..v6 add 200, 300, 300, 400, 400, 400, price 20, 40, 80, 160, 320, 640, max 2400; all U placeholders). The wallet implements `vault.Account` (`OwnedSteps`) and, through `Pay()`, `vault.Wallet` (`Balance`, `Spend(key, n)` idempotent per key).

- `Upgrade(step)`: steps are bought in table order; `""` means the next. An owned step is a no-op (`already`: a double click books once). A short balance returns `ShortError{Need, Have}` (HTTP 402), nothing booked. Key `vault/<step>#<n>` where n counts earlier refunds of the step, so buy, refund, buy again works.
- `Downgrade(step)`: refunds only the LAST owned step, in full (the price paid then), and only while the active hero's vault count fits the lower tier (`Options.Content`); else `FitError{Count, Capacity}` (409 `in-use`) and nothing changes. Shrinking never deletes anything. Refund of a step not owned is a no-op, of another step than the last is out of order.
- The refund check reads the ACTIVE hero's content; capacity itself is account-wide. V `TestUpgradeSpendRefundRules`, `TestHeroesShareWalletRefundReadsActiveHero`, `TestAllStepsThenNothingToBuy`; balance never negative and replay equals state over random sequences V `TestPropertyBalanceNeverNegativeAndReplays`.

## API (`internal/watch/etherapi.go`, `EtherAPI{W}`)

- `GET /api/ether` (`EtherAPI.Get()`, GET only, schema `watch.ether/1`): `currency, balance, chip ("Ether 35"), owned[], capacity, next {id, add, price, canAfford, short, capacity}, downgrade {available, step, refund, reason}, sources [{id, label, dailyCap, earnedToday}], earned {source: total}, recent[] (20 newest ledger entries)`.
- `POST /api/vault/upgrade` `{"step":"v1"}` (step optional) and `POST /api/vault/downgrade` (`Upgrade()` / `Downgrade()`, mount on the answer listener like `POST /api/orders`). Reply `{ok, changed, already, step, price, balance, capacity}`; refusals carry `code` and `error`: 402 `insufficient` (+ `need`), 409 `in-use` (+ `count`, `fitsTier`), 409 `out-of-order`, `maxed`, `nothing-to-refund`, 400 `bad-body`.
- Contract for the V4 page (additive): its `VaultUpgrade` placeholder maps as `available = next.canAfford`, `next = {id, add, price}`, `Owned = Wallet.OwnedSteps`; the Upgrade button POSTs `/api/vault/upgrade`; the header and Hero sheet chip is `chip`; the capacity meter takes `capacity`. Field names of `GET /api/vault` are untouched.

## Wiring (V7b, `cmd/d2bot/etherwire.go`)

- One wallet per ACCOUNT: `<currency dir>/ether.jsonl` of the profile (`etherFor`, one `etherHome` per directory, shared by the recorder, the page, the routes and the earn hook). A Character Select switch restarts the child and the next one opens the same ledger; two accounts have two wallets (V `TestEtherFeedEventsRestartAndAccounts`, `TestEtherCapacityAndHeroSwitch`).
- Behind `loot.vault`: with the knob at 0 the wallet is not opened for earning, spending or the routes (`Wallet()` is nil, GET says `enabled: false`, the POSTs answer 409 `off`, nothing is earned). The capacity of a ledger that exists on disk is still read, so a read only page does not show a false over capacity.
- Capacity: `vaultHome` gives every store `Options.Capacity` = the wallet's `Capacity(0)` (late bound, open stores see an upgrade at once); `VaultDesk.Owned` is the wallet's owned steps; the refund rule's `Content` is the ACTIVE hero's record count. `Wallet.Capacity` takes no lock on purpose: the store calls it under its own lock and the wallet reads the store's count under its own, and a locked read was a lock order inversion (V `TestEtherNoDeadlockStoreAndWallet`).
- Routes: `Fun.Ether` mounts `GET /api/ether` on the viewer and `POST /api/vault/upgrade`, `/downgrade` on the answer listener behind the same gate as the other answer routes (token, origin list, JSON only, small body with known fields; V `TestEtherAPIGateAndOff`). The V7 routes were ungated; V7b adds `AnswerOptions`. The body is `{}` or `{"step":"v1"}`.
- Earn: `startEther` subscribes to the goal manager and `etherFeed` maps events to `ether.Earner` calls (kill, boss kill with the hero's difficulty, field area change, level up). Keys carry what makes them once: see the table. Replaying a journal or restarting books nothing twice (V `TestEarnerNoDoubleCreditAcrossRestartAndHeroes`, `TestEtherFeedEventsRestartAndAccounts`); the rules table is V `TestEarnerRulesTable`; the trickle cap and the local day roll are V `TestEarnerKillTrickleCapAndDay`.
- Hero tile: `GET /api/hero` carries `ether` ("Ether 35") only while the wallet is on (`Hub.SetEtherSource`); the Hero panel draws it as a link chip to the Vault.

## The page

`/watch/vault` reads `GET /api/ether` next to `GET /api/vault` (on load, after every press and on the poll). With a wallet: the header shows the balance chip and a real Upgrade button ("Upgrade +200 for 20 Ether"); a press opens the confirm sheet (price, balance before and after, new capacity), Pay posts `{"step": <next>}` and the page then reloads both GETs and says what they say ("Capacity is now 600. 15 Ether left."). A short balance keeps the button `aria-disabled` and shows "Need N more Ether" (a press says it on a phone). A Refund button appears once a step is owned; while the vault does not fit the lower tier it is `aria-disabled` with the server's reason, otherwise its sheet shows the refund and the capacity. Refusals use the texts of `vault.json` with the numbers of the answer (402 `insufficient`, 409 `in-use`, `out-of-order`, `maxed`, `nothing-to-refund`, the gate codes); an unknown code falls back to the server sentence, never to a generic line. Without a wallet the old disabled placeholder stands with the reason "needs the vault switched on". The sheet is registered with the overlay manager. V node test `TestVaultEtherUIJS`; fit pages `vault-eth*`, `default-ether`, must-fail `vaultdefect-ethenabled`.

## Not done

- The Quest board hand-in call (S10) and the `meta` column; the live drill with a real bot (the earn rules need real kill and area events: whether the bot's journal carries `Monster` for every kill and `Area` for every change is U until a run).
- Auto-pickup range and filter slots as more rows of the same kind (later).
- A hero that is not the active one cannot be refunded against its own content (the rule reads the active hero only; capacity itself is account wide).

Open: U every amount and price; whether the quest `meta` column should also feed `first`; the user may want ETHER shown under another name (the name is the `currency` string of the table).
