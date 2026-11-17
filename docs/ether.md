# ETHER: the account meta wallet of the Ethereal Vault

> status: built offline 2026-11-17 (agent vaultv7, slice V7 of [design/loot-vault-2026-11-16.md](design/loot-vault-2026-11-16.md) section 2.6). NOT wired into the bot and NOT run live. Tags: **V** verified by a test in this repo, **U** unverified or a placeholder number.

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
| `quest` | `quest/<quest>/<difficulty>` | the caller's amount, 0..100 (`maxHook`) | HOOK for the Quest board (S10): the `meta` column value for that difficulty. Account-wide: once per quest and difficulty |
| `first` | `first/<kind>/<name>` | 5 U | the "a first" rows of `tally_firsts` |
| `boss` | `boss/<name>/<difficulty>` | 10 / 15 / 25 U | normal / nm / hell |
| `milestone` | `milestone/<id>` | per row U | `clvl10, clvl20, clvl30, clvl50, diff_nm, diff_hell`; an id not in the table is refused |
| `kills` | `kills/<batch id>` | 1 U | capped at 10 per UTC day (`dailyCap`); the cap clips, a capped call books nothing |

`Wallet.Earn(source, sub, difficulty, hookAmount)` returns `{credited, key, already, capped}`. A repeated key pays nothing (`already`), so a journal replay books nothing (V `TestEarnIdempotentAndReplay`). Unknown source, missing or unknown difficulty, over-limit hook amount: an error naming it (V `TestEarnErrors`). No call site exists yet: the Quest board hand-in (S10), the tally event hook and the bot's milestone events call `Earn`; this slice only documents them. Pressing Keep never earns.

## Capacity tiers and spending

Tiers are `data/vault/capacity.json` (base 400, steps v1..v6 add 200, 300, 300, 400, 400, 400, price 20, 40, 80, 160, 320, 640, max 2400; all U placeholders). The wallet implements `vault.Account` (`OwnedSteps`) and, through `Pay()`, `vault.Wallet` (`Balance`, `Spend(key, n)` idempotent per key).

- `Upgrade(step)`: steps are bought in table order; `""` means the next. An owned step is a no-op (`already`: a double click books once). A short balance returns `ShortError{Need, Have}` (HTTP 402), nothing booked. Key `vault/<step>#<n>` where n counts earlier refunds of the step, so buy, refund, buy again works.
- `Downgrade(step)`: refunds only the LAST owned step, in full (the price paid then), and only while the active hero's vault count fits the lower tier (`Options.Content`); else `FitError{Count, Capacity}` (409 `in-use`) and nothing changes. Shrinking never deletes anything. Refund of a step not owned is a no-op, of another step than the last is out of order.
- The refund check reads the ACTIVE hero's content; capacity itself is account-wide. V `TestUpgradeSpendRefundRules`, `TestHeroesShareWalletRefundReadsActiveHero`, `TestAllStepsThenNothingToBuy`; balance never negative and replay equals state over random sequences V `TestPropertyBalanceNeverNegativeAndReplays`.

## API (`internal/watch/etherapi.go`, `EtherAPI{W}`)

- `GET /api/ether` (`EtherAPI.Get()`, GET only, schema `watch.ether/1`): `currency, balance, chip ("Ether 35"), owned[], capacity, next {id, add, price, canAfford, short, capacity}, downgrade {available, step, refund, reason}, sources [{id, label, dailyCap, earnedToday}], earned {source: total}, recent[] (20 newest ledger entries)`.
- `POST /api/vault/upgrade` `{"step":"v1"}` (step optional) and `POST /api/vault/downgrade` (`Upgrade()` / `Downgrade()`, mount on the answer listener like `POST /api/orders`). Reply `{ok, changed, already, step, price, balance, capacity}`; refusals carry `code` and `error`: 402 `insufficient` (+ `need`), 409 `in-use` (+ `count`, `fitsTier`), 409 `out-of-order`, `maxed`, `nothing-to-refund`, 400 `bad-body`.
- Contract for the V4 page (additive): its `VaultUpgrade` placeholder maps as `available = next.canAfford`, `next = {id, add, price}`, `Owned = Wallet.OwnedSteps`; the Upgrade button POSTs `/api/vault/upgrade`; the header and Hero sheet chip is `chip`; the capacity meter takes `capacity`. Field names of `GET /api/vault` are untouched.
- Wiring still to do (V8 or the V4 merge): open the wallet in `cmd/d2bot` next to the tally ledger (`<currency dir>/ether.jsonl`), pass `Content` from the active vault store, `VaultStore` capacity `Options.Capacity = wallet.Capacity(override)`, mount the three routes. Not done here to keep the slice additive.

Open: U every amount and price; whether the quest `meta` column should also feed `first`; the user may want ETHER shown under another name (the name is the `currency` string of the table).
