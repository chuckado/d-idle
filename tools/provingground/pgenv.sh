# Sourced by run.sh on the builder: the environment that turns the results ledger on (card pgledgeron, docs/process/pg-ledger-ops.md).
# Host config (the ingest command's address) is never in the repo: it comes from an env file on the host, $PG_ENV (default /data/pg/env),
# which may set PG_LEDGER_CMD. Everything else is derived here and only fills what the caller has not set.
[ -r "${PG_ENV:-/data/pg/env}" ] && . "${PG_ENV:-/data/pg/env}"
case "$PWD" in
*/d2lod-wt/*) _w=${PWD#*/d2lod-wt/}; : "${PG_BRANCH:=agent/${_w%%/*}}" ;; # the mirror keeps the agent's worktree name
esac
: "${PG_EVENT:=manual}"            # manual | merge | nightly | smoke: the scheduler sets it, a hand run is `manual`
: "${PG_JOB:=${LABRUN_JOB:-}}"
export PG_EVENT PG_BRANCH PG_JOB PG_CARD PG_LANE PG_LEDGER_CMD
