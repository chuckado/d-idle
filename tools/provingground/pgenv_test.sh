#!/bin/bash
# Offline test of pgenv.sh: the ledger environment is derived from the worktree name and the host env file, and caller values win.
d=$(mktemp -d); trap 'rm -rf "$d"' EXIT; fail=0
here=$(cd "$(dirname "$0")" && pwd)
mkdir -p "$d/code/d2lod-wt/pgx/sub" "$d/plain"
echo 'PG_LEDGER_CMD="cat >> '"$d"'/spool"' > "$d/env"
t() { (cd "$1" && shift && env -i PATH="$PATH" "$@" bash -c ". $here/pgenv.sh; echo \"\$PG_EVENT|\$PG_BRANCH|\$PG_JOB|\$PG_LEDGER_CMD\""); }
[ "$(t "$d/code/d2lod-wt/pgx/sub" PG_ENV="$d/env" LABRUN_JOB=j1)" = "manual|agent/pgx|j1|cat >> $d/spool" ] && echo "ok   derived from the worktree name, job and env file" || { echo "FAIL derive: $(t "$d/code/d2lod-wt/pgx/sub" PG_ENV="$d/env" LABRUN_JOB=j1)"; fail=1; }
[ "$(t "$d/plain" PG_ENV=/nonexistent)" = "manual|||" ] && echo "ok   outside a worktree and without env file nothing is invented" || { echo "FAIL plain"; fail=1; }
[ "$(t "$d/code/d2lod-wt/pgx" PG_ENV="$d/env" PG_EVENT=nightly PG_BRANCH=main PG_JOB=n7)" = "nightly|main|n7|cat >> $d/spool" ] && echo "ok   caller values win" || { echo "FAIL caller"; fail=1; }
exit $fail
