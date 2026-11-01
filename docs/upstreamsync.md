# Upstream sync after #152 and #153 merged (2026-11-01)

> status: current (card upstreamsync, 2026-11-01). Read-only measurement plus local rebases on the worker host; nothing pushed, nothing posted, no PR comment, the private fork not rebased.

Upstream main moved from 72132b99 to `7c3a50af` (merge of #153 on top of the merge of #152; 13 commits, 15 files). Files changed by #152 and #153 together: `components/game/world.hpp`, `components/game/world/commands.cpp`, `components/game/world/setup.cpp`, `components/rules/rules.hpp`, `tests/game/test_game.cpp`, `tests/test_rules.cpp`, `tests/pad_use_d2d.py`, `tools/emu/README.md`, `tools/emu/price_model.py`, `price_oracle.py`, `price_terms.py`, `wp_refuse.py`, `docs/research/re/act1-end.md`, `store.md`, `waypoint.md`.

## 1. Mergeability against 7c3a50af

Method: `git merge-tree --write-tree` of the branch head against upstream/main (textual only, no checkout changed). "Overlap" = files the branch changes that #152 / #153 also changed.

| Branch (worker host checkout) | Head | Merge | Overlap with #152 / #153 |
|---|---|---|---|
| #155 `fix/static-field` (as pushed: `pr/cr155`, `fix/cr155`) | 8c56e96a | clean | none |
| #156 `fix/walk-velocity` (`pr/walk4`, `fix/walk-velocity-v4`) | 3f418788 | CONFLICT `tools/emu/README.md` only (one hunk: both sides added a row to the same oracle table; keep both rows). `components/game/world.hpp` auto-merges | `world.hpp` (#153), `tools/emu/README.md` |
| #157 `fix/boss-stats` (`pr/buildmods`) | cf85634a | clean | none |
| #158 `fix/nexthit-justhit` (`pr/nexthit2`, `fix/nexthit-v2`) | 6b3964fa | clean | none |
| `fix/armor-hit` (`pr/armorhit`) | d5bd19c1 | clean | none |
| `fix/freeze-resume` (`pr/fresume`) | efacc7f4 | clean | none |
| bundle `fix/frozen-armor-and-resume-2c` (`pr/bundlear`) | f9e38b15 | clean | none |
| `fix/curse-immune` (`pr/skresistfix`) | 557d14e3 | clean | none |
| `fix/vendor-prices-followon-v2` (`pr/price4`, stacked on #152's old head 8a45cfc9, now in main) | da80d0ab | clean | it extends the code #152 changed (`rules.hpp`), no textual conflict |
| `fix/freeze-rules-v2` (`pr/freeze2`) | 409903a6 | clean | none |
| `fix/missile-roll-top` (`pr/skmiss`) | 23e2baec | clean | none |
| `fix/skills-doc-cast`, `fix/lut-gholein-preset-rng`, `fix/lightning-sequence`, `fix/missile-damage-fine-v2` | e310a1c9, 4bbf7f37, 2036fd3f, 8a3cb678 | conflicts (old pre-split branches; `fight.cpp` deleted upstream) | not from #152 / #153; stale, already merged under other heads or superseded |

Only #156 needs a hand edit, and that edit is docs-only. The merge preview for #155, #157, #158 is clean, but each PR still needs its own CI after the maintainer's next merge; none of them touches what #152 / #153 changed.

### Rebased copies (new branches, new worktrees, originals untouched)

Cherry-picks (`-x`) of the original commits onto 7c3a50af; no conflicts. Checks per tree: `tools/pr_guard.py upstream/main HEAD` and per commit, `test_pr_guard.py` and `tools/emu/test_tooling.py`, Release build of the tests, `test_rules` rc 0, `test_game` OK. The differentials (oracle runs, 0 mismatches) were NOT re-run: neither claim touches a file #152 / #153 changed, so the counts recorded in the proofs stay the evidence; an author re-run goes with the push card.

| New branch | Worktree (worker host) | Commits | pr_guard | Tests |
|---|---|---|---|---|
| `fix/armor-hit-rebased` | `~/d2d-lab/pr/armorhit-rebased` | 2a28a088, d5b03195 | 0 FAIL / 0 WARN (range and each) | test_rules rc 0, test_game OK |
| `fix/freeze-resume-rebased` | `~/d2d-lab/pr/fresume-rebased` | 1bdf489c, 5d70d63c | 0 / 0 | test_rules rc 0, test_game OK |
| `fix/frozen-armor-and-resume-2c-rebased` | `~/d2d-lab/pr/bundlear-rebased` | 5ca03331, 931cc65b | 0 / 0 | test_rules rc 0, test_game OK |
| `fix/curse-immune-rebased` | `~/d2d-lab/pr/skresistfix-rebased` | 4132071c, 16d6663a | 0 / 0 | test_rules rc 0, test_game OK |

Not rebased: #155, #157, #158 (open PR heads; a rebase would change the PR head and the maintainer has not asked), #156 (conflict, see above), `price4` (stacked follow-on, rebase recipe in verify-price2.md now applies: `git rebase --onto upstream/main 8a45cfc9`; merge preview clean). Test targets of the pr_guard-built trees: `~/d2d-lab/pr/upstreamsync-out/<tree>/`.

## 2. The private fork (the fork checkout on the worker host, private remote)

Fork branches are a linear stack, each on the previous: `lab-debug-skills` (6 commits over 03f42254) < `lab-nexthit` (+3) < `lab-armorhit` (+2) < `lab-freezeresume` (+2). Base is 03f42254, older than the old upstream main 72132b99, so the fork also lacks #147 and #149.

| Fork branch | Head | Merge-tree against 7c3a50af | Against the OLD main 72132b99 |
|---|---|---|---|
| `lab-debug-skills` | d9b96c2b | CONFLICT `apps/d2d/town/update.cpp`, `components/rules/rules.hpp` | same two files |
| `lab-nexthit` | 2c6c4a1a | same two files (from the base commit) | same |
| `lab-armorhit` | 7c73b881 | same | same |
| `lab-freezeresume` | efe7bc36 | same | same |
| `lab` (old lab6 line, base e36922ea) | 07453d15 | conflicts (`CMakeLists.txt`, `devctl_verbs.hpp` deleted upstream, more); retired by `lab-port-148` | not examined |

The conflicts come from the lab click-modifier commit `3a629371` (shift/ctrl click on pack items, `rules::belt_slot`), which collides with upstream's merged #147 / #149 (`belt_box`, `auto_belt`, `click.cmd`, the shift flag). They are NOT caused by #152 or #153: the same two files conflict against 72132b99. The nexthit, armorhit and freezeresume commits themselves apply on top of whatever the first commit resolves to; the merge-tree conflict listing stops at the stack base.

PLAN (a later card does it under the standing fork authorisation; do not rebase here):

1. `lab-debug-skills` owns the stack base: rebase onto 7c3a50af, resolving `3a629371` first. Keep `update.cpp`'s upstream `click.cmd` / shift handling and add only the part upstream lacks (ctrl-click sell to the open store: `cmd::Sell`); drop the lab `belt_slot` and shift-to-belt code in favour of upstream `belt_box` / `auto_belt` (superseded, section 3). In `rules.hpp` take upstream's belt block whole and re-add nothing. Check `test_rules.cpp` (3 lines in the same commit). Then the debug-missiles and debug-skills commits (no known conflicts).
2. `lab-nexthit` onto the new `lab-debug-skills`: its three commits equal #158's claim and tooling (the PR is still open); when #158 merges, drop them and keep the stack lean. Until then they stay: the lab binary needs NextHit.
3. `lab-armorhit` onto `lab-nexthit`: two commits equal `fix/armor-hit` (rebased copy 2a28a088, d5b03195 is clean on the new main).
4. `lab-freezeresume` onto `lab-armorhit`: two commits equal `fix/freeze-resume`.
5. Rebuild the lab binary from `lab-freezeresume` (head of the stack), re-run the lab checks (`dump vendor` prices change because #152 is now in the base: new prices are the intended ones), then push all four branches (the rebase rewrites them, so the push is a force push on the fork's own branches; the standing rule is no force-push, so the card needs the user's explicit word for that) and update docs/d2d-fork.md.
6. Order of ownership afterwards: each PR that merges upstream is dropped from the stack at the next rebase (#158 first, then the armor / resume pair when opened); `lab-debug-skills` stays private for good (debug verbs are not offered upstream).

## 3. Superseded local fixes and fork patches

| Item | Evidence | Action |
|---|---|---|
| Fork vendor-price or waypoint patch | none exists: `git diff 03f42254..lab-freezeresume` contains no pricing or waypoint rule change; the only price touch is a caller `rules::item_price(...)` in the `dump vendor` verb (commit 7abb81d1 `dump vendor section`, in `3a629371`), whose signature #152 did not change | nothing to drop; vendor dump now shows #152's prices |
| Fork branch `fix-vendor-prices` (4884b1d5 tome / quiver by quantity) | #127 merged long before; not a lab patch | none |
| Lab `belt_slot` / shift-click belt in `3a629371` | upstream #147 / #149 merged `belt_box` / `auto_belt` (conflict hunks in `rules.hpp` and `update.cpp`) | drop at the fork rebase, keep ctrl-click sell |
| Bot `WpCooldownSec` / `waitCooldown` (`internal/env/tuning_nav.go`, `internal/nav/waypoint.go`) | #153 merged: single-player waypoint use has no arrival cooldown in game.exe; BACKLOG.md:90 already marks it unnecessary once merged and in the lab binary | becomes removable only after the lab binary is rebuilt on 7c3a50af (step 5); a bot card, not this one |
| Local candidate branches `fix/vendor-prices`, `fix/vendor-prices-followon` (v1), `fix/waypoint-hostile-window` (`pr/price`, `price2`, `price3`, `wp`) | #152, #153 merged; `wp` head 5315e66a is in main | can be deleted by the owner of the PR checkout |

## 4. Outdated docs (file:line, proposed edit)

Nothing below is edited by this card (other cards own the files). Lines are as of origin/main 95bdbf5.

| File:line | Now says | Proposed edit |
|---|---|---|
| docs/research/pr-bundle-queue.md:9 | two open PRs by exception: #149 stable, price PR (#152) | "#152 and #153 MERGED upstream 2026-11-01 by the maintainer; #155 to #158 open" |
| docs/research/pr-bundle-queue.md:19 | waypoint: OPEN #153 | "MERGED 2026-11-01 (maintainer)" |
| docs/research/pr-bundle-queue.md:20 | vendor prices: OPEN #152 | "MERGED 2026-11-01 (maintainer)" |
| docs/research/pr-bundle-queue.md:22 | price follow-on: "#152 must merge first" | "#152 merged; rebase recipe applies; merge preview vs 7c3a50af clean (card upstreamsync)" |
| docs/BACKLOG.md:70, 74 | #153 open | add "MERGED 2026-11-01" |
| docs/BACKLOG.md:76 | price PR OPENED #152, two open PRs exception | add "MERGED 2026-11-01" |
| docs/BACKLOG.md:90 | waypoint PR READY (local) ... "once merged upstream" | "merged 2026-11-01; WpCooldownSec removal waits for the lab binary on 7c3a50af" |
| docs/BACKLOG.md:99, 151 | #153 review watch; "one open bundle" | mark #153 merged; drop the CodeRabbit watch line |
| docs/decomp-queue.md:22, 24 (items 3 and 5), 29, 60 | local branch / PR candidate status for waypoint and prices, follow-on "pending #152" | "merged upstream 2026-11-01 (#153, #152)"; follow-on: "rebase now possible" |
| docs/primitives.md:35, 37, 134 (P23) | failure list "10 s refusal after level change"; drills show the 10 s second use | note: game.exe single player has no arrival cooldown; d2d follows after #153 (lab binary on 7c3a50af); drill rows stay as history |
| docs/waypoints.md:20 | bot waits 10 s after the level change | add "d2d needs it only until the lab binary has #153" |
| docs/fixwp.md:9 | W3 untouched: d2d refuses for 10 s after any arrival | add "fixed upstream by #153 (merged 2026-11-01)" |
| docs/research/waypoint-trace.md:46, 65 | W3 "right on d2d today" | same banner |
| docs/research/pr-wp-draft.md:3, 56, 57 | DRAFT / "joins the next bundle after #149 / #152 merge" | status banner "MERGED as #153" |
| docs/research/pr-wp-evidence-pack.md:3, 21 | POSTED as #153 | add "MERGED 2026-11-01" |
| docs/research/pr-price-draft.md:3, pr-price-evidence-pack.md:3, pr-price-proof.md:3 | DRAFT / current, nothing posted | add "MERGED as #152 on 2026-11-01" |
| docs/research/pr-price2-draft.md:3, pr-price2-evidence-pack.md:3, pr-price2-proof.md:3, verify-price2.md:48, 66 | follow-on "stacked on #152" / "does NOT apply today: #152 is not merged" | "#152 merged; rebase onto upstream/main is clean (merge-tree 2026-11-01)" |
| docs/research/pr-curseimmune-proof.md:62, 76 | merged-against list names #152 and #153 as open and base 72132b99 | "base now 7c3a50af; clean, rebased copy fix/curse-immune-rebased" |
| docs/research/pr-mods-proof.md:17; pr-mods-draft-v2.md:61; pr-nexthit-verifier.md:81 | "shares no file with #152, #153" | still true; add that #152 / #153 are merged |
| docs/research/pr-walk-evidence-pack.md:64 | "main unchanged at 72132b99, #152 #153 still open" | history line, leave; add a posted line "main now 7c3a50af, #156 needs one README row merge" |
| docs/d2d-fork.md (fork base / lab6 paragraph near :227) | base 32768ddb | point at this doc's rebase plan |
| docs/review-strategy.md:115 and README.md:17, 30, docs/capabilities.md:12 | gate ladder / capability lines | checked: NOT outdated (no PR or waypoint status in them) |

Count: 22 rows (about 45 line references); `review-strategy.md`, README and capabilities need nothing.

## After #157 / #158 (upstream main 52362d52)

The maintainer also merged #157 and #158. Merge-tree (textual) of the four rebased copies against 52362d52: `fix/armor-hit-rebased`, `fix/freeze-resume-rebased`, `fix/frozen-armor-and-resume-2c-rebased` and `fix/curse-immune-rebased` all merge clean (rc 0, no conflicts). Builds and tests were not re-run on the new head. Still open on our side: #155, #156 (README row conflict, section 1). The fork plan (section 2) drops the lab-nexthit commits now, because #158 is merged.

## Shared-file lines for the merger

- docs/BACKLOG.md: "upstreamsync (card upstreamsync, 2026-11-01): the maintainer merged #152 (vendor prices) and #153 (waypoint cooldown) on 2026-11-01; main is 7c3a50af. Merge preview clean for #155, #157, #158 and the armor / resume / curse-immune candidates (rebased copies fix/armor-hit-rebased, fix/freeze-resume-rebased, fix/frozen-armor-and-resume-2c-rebased, fix/curse-immune-rebased: pr_guard 0/0, tests green); #156 conflicts in one README row; fork stack needs a hand merge of the lab click-modifier commit (plan in docs/upstreamsync.md); bot WpCooldownSec removable once the lab binary is on 7c3a50af."
- docs/INDEX.md: "[upstreamsync.md](upstreamsync.md) mergeability of our PRs and candidates against upstream main after #152 / #153, fork rebase plan, outdated-docs list."
- CHANGELOG.md: "agent/upstreamsync: mergeability sweep after #152 and #153 merged upstream, four clean rebased copies of local candidates, fork rebase plan, outdated-docs list. Offline, nothing pushed."
- docs/research/pr-bundle-queue.md: "#152 and #153 MERGED upstream 2026-11-01 by the maintainer (new main 7c3a50af)." plus the rows in section 4.
