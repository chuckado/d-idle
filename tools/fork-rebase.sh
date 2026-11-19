#!/bin/sh
# Fork rebase routine (docs/process/fork-rebase.md). Host-free: the checkout and the remote names come from the
# environment or local.env (see tools/env.sh); builder and sandbox steps go through the private ops repo (tools/ops.sh).
#   tools/fork-rebase.sh measure [PREV_BRANCH]          drift of PREV (default: newest didle/rebase-* on the private remote) vs upstream main
#   tools/fork-rebase.sh rebase  [PREV_BRANCH] [NEW]    new worktree + branch NEW (default didle/rebase-<today>), rebase --onto upstream main;
#                                                       stops on a conflict (exit 3) naming the files; resolve in the new worktree and `git rebase --continue`
#   tools/fork-rebase.sh push [NEW]                     push NEW to the private remote (never forced)
#   tools/fork-rebase.sh build|test|image [NEW]         builder / sandbox steps: delegated to the ops repo (tools/fork-rebase-builder.sh)
# Env: D2_FORK_CHECKOUT (a git checkout of the fork with both remotes), D2_FORK_UPSTREAM (remote name, default upstream),
#      D2_FORK_PRIVATE (private remote name, default origin), D2_FORK_WT_DIR (where new worktrees go, default <checkout>/..).
# Exit: 0 ok, 2 usage or missing setup, 3 conflict (resolve, then continue), 4 refused (branch exists, rewrite attempt).
here=$(cd "$(dirname "$0")" && pwd)
. "$here/env.sh"
CK=${D2_FORK_CHECKOUT:-}
UP=${D2_FORK_UPSTREAM:-upstream}
PR=${D2_FORK_PRIVATE:-origin}
[ -n "$CK" ] && [ -d "$CK/.git" -o -f "$CK/.git" ] || { echo "fork-rebase: set D2_FORK_CHECKOUT to a checkout of the private fork (local.env)" >&2; exit 2; }
WT=${D2_FORK_WT_DIR:-$(dirname "$CK")}
g() { git -C "$CK" "$@"; }
cmd=${1:-}; [ $# -gt 0 ] && shift

latest_prev() { g for-each-ref --format='%(refname:short)' "refs/remotes/$PR/didle/rebase-*" | sed "s#^$PR/##" | sort | tail -1; }
fetch() { g fetch -q "$UP" && g fetch -q "$PR" || { echo "fork-rebase: fetch failed ($UP / $PR)" >&2; exit 2; }; }

case $cmd in
measure)
	fetch
	prev=${1:-$(latest_prev)}; [ -n "$prev" ] || { echo "fork-rebase: no previous didle/rebase-* branch" >&2; exit 2; }
	b="$PR/$prev"; u="$UP/main"
	mb=$(g merge-base "$b" "$u")
	echo "previous branch: $prev ($(g rev-parse --short "$b"))"
	echo "merge base: $(g log -1 --format='%h %cs %s' "$mb")"
	echo "upstream behind: $(g rev-list --count "$b..$u") commits ($(g rev-list --count --no-merges "$b..$u") non-merge), $(g diff --shortstat "$mb" "$u")"
	echo "ours ahead: $(g rev-list --count --no-merges "$u..$b") commits"
	t=$(mktemp -d); g diff --name-only "$mb" "$b" | sort >"$t/ours"; g diff --name-only "$mb" "$u" | sort >"$t/theirs"
	echo "files touched on both sides (predicted conflicts, text conflicts are fewer):"; comm -12 "$t/ours" "$t/theirs" | sed 's/^/  /'
	echo "upstream commits on the devctl layer:"; g log --no-merges --format='  %h %s' "$mb..$u" -- apps/d2d/devctl docs/control_channel.md | cut -c1-120
	echo "upstream merged PRs:"; g log --first-parent --format='  %h %s' "$mb..$u" | cut -c1-110
	rm -rf "$t" ;;
rebase)
	fetch
	prev=${1:-$(latest_prev)}; new=${2:-didle/rebase-$(date +%Y%m%d)}
	[ -n "$prev" ] || { echo "fork-rebase: no previous branch" >&2; exit 2; }
	[ "$prev" != "$new" ] || { echo "fork-rebase: new branch equals the previous one; never rewrite a previous branch" >&2; exit 4; }
	g rev-parse -q --verify "refs/heads/$new" >/dev/null && { echo "fork-rebase: local branch $new exists" >&2; exit 4; }
	g rev-parse -q --verify "refs/remotes/$PR/$new" >/dev/null && { echo "fork-rebase: $new exists on the private remote" >&2; exit 4; }
	dir="$WT/wt-$(echo "$new" | tr / -)"
	mb=$(g merge-base "$PR/$prev" "$UP/main")
	g worktree add -q "$dir" -b "$new" "$PR/$prev" || exit 2
	git -C "$dir" config --unset branch."$new".remote 2>/dev/null
	echo "worktree $dir, branch $new from $prev, rebasing onto $UP/main (old base $(echo "$mb" | cut -c1-9))"
	if GIT_EDITOR=true git -C "$dir" -c user.name="${GIT_AUTHOR_NAME:-Claude}" -c user.email="${GIT_AUTHOR_EMAIL:-noreply@anthropic.com}" rebase --onto "$UP/main" "$mb" >"$dir/.rebase.log" 2>&1; then
		rm -f "$dir/.rebase.log"
		echo "rebase clean: $(git -C "$dir" rev-list --count "$UP/main..HEAD") commits on $UP/main, tip $(git -C "$dir" rev-parse --short HEAD)"
		echo "next: tools/fork-rebase.sh build $new ; test ; push ; image"
	else
		echo "CONFLICT in $dir:"; git -C "$dir" diff --name-only --diff-filter=U | sed 's/^/  /'
		echo "resolve (keep both sides' intent, docs/process/fork-rebase.md), then in $dir:"
		echo "  git add <files> && GIT_EDITOR=true git rebase --continue    (repeat per conflicting commit)"
		echo "then: tools/fork-rebase.sh build $new"
		exit 3
	fi ;;
push)
	new=${1:-$(g branch --show-current)}
	case $new in didle/rebase-*) ;; *) echo "fork-rebase: refusing to push $new (not didle/rebase-*)" >&2; exit 4 ;; esac
	g push "$PR" "$new:$new" ;;   # no --force: a rewrite attempt is rejected by the remote
build|test|image)
	new=${1:-}; [ -n "$new" ] || { echo "usage: fork-rebase.sh $cmd <branch>" >&2; exit 2; }
	exec "$here/ops.sh" run fork-rebase-builder.sh "$cmd" "$new" ;;
*) sed -n 2,12p "$0"; exit 2 ;;
esac
