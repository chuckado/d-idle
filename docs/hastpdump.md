# inv-tp-dump: why HasTP differs per area

Context: docs/soakstall.md (Gate 1: `HasTP` false in Cold Plains, true in Blood Moor, cause UNKNOWN).

## What was added (internal/bot/tpdump.go, hook in `hasPortalSource`)
One INFO line `inv-tp-dump`, emitted from the exact code path of `resInput().HasTP` (`combat.PortalSource` with the bot's catalog adapter and `Read.UnitStats`), on the first look, on every area change and on every HasTP change; nothing otherwise. No behaviour change. Offline test: `TestInvTPDump` (internal/bot/resource_portal_test.go). Run: `go test -p 1 -run TestInvTPDump ./internal/bot`. Not verified live.

Fields: `why` (first|area|verdict), `area`, `hastp`, `decided_by` (class and addr of the item PortalSource returned, or "none"), `statsReader` (false = no UnitStats source, a tome then counts as usable), `portal_items` (every portal scroll/tome in ANY container: kind, class, `where`, addr, `counted` = our PortalQty, `raw70` = UnitStats stat 70 as read: a number, `no70` stat list without stat 70, `err:...` read failed, `nostats`), `items` (all `class@where`).
Caveat: the town's `scrollCount` also applies `portalEmpty` (an override outside this dump); a mismatch between this dump and the shop log is itself a finding.

## Drill (next soak log; grep `inv-tp-dump`)
Compare the line at entry to Cold Plains (hastp=false) with the one at Blood Moor (hastp=true):
1. Stale or wrong tome count: same addr, `counted` differs between areas while `raw70` is the same, or `raw70` changes between two lines with no purchase/cast between. Points at our computation or a stale cache.
2. Wrong Where filter: the scroll/tome shows `where` other than `inventory` (cube, stash, belt) in the false line, or the item moved container between the lines. PortalSource only accepts ContInventory.
3. UnitStats difference (for example after a restart or per area): `raw70=err:` or `no70` for a tome that is plainly full, or an `addr` that changes (stale pointer). Then the stat read, not the inventory, flips the verdict.
4. Scroll vs tome: the line after "bought Scroll of Town Portal" lacks any `scroll` entry in `portal_items` (purchase not in the snapshot yet: timing) while a tome is present with `counted=0`; or a scroll is present with `where` not inventory.
5. No portal item at all in `portal_items` in the false line but one in the true line: the snapshot item list itself differs per area (compare `items`), i.e. a d2d inventory dump issue.
Record the lines under docs/soakstall.md Missing data and decide; each hypothesis is U until a line shows it.
