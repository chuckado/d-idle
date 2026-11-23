#!/usr/bin/env bash
# Drive quests through the operator control of a running container (docs/operator.md): status -> start (or retry when
# parked) -> wait -> on a park back off and retry, up to --max-tries. Talks only to `docker exec <container> d2bot ops`;
# it holds no token and needs none. Writes a markdown log to stdout.
# Usage: tools/questloop.sh <container> <quest|--all> [--max-tries 3] [--wait 600] [--backoff 65] [--bin PATH]
# Exit: 0 every wanted quest done, 1 a quest exhausted its tries (evidence in the log), 2 usage.
# Test hooks: QUESTLOOP_SLEEP (default sleep), D2BOT_BIN (same as --bin).
set -u
usage() { sed -n '2,8p' "$0" | sed 's/^# \{0,1\}//' >&2; exit 2; }
[ $# -ge 2 ] || usage
container=$1; target=$2; shift 2
tries=3; wait_s=600; backoff=65; bin=${D2BOT_BIN:-}
while [ $# -gt 0 ]; do
  case $1 in
    --max-tries) tries=${2:?}; shift 2 ;;
    --wait) wait_s=${2:?}; shift 2 ;;
    --backoff) backoff=${2:?}; shift 2 ;;
    --bin) bin=${2:?}; shift 2 ;;
    *) usage ;;
  esac
done
sleeper=${QUESTLOOP_SLEEP:-sleep}

# Find the binary inside the container: the current release link, else the newest release, else PATH.
if [ -z "$bin" ]; then
  for c in /data/didle/current/bin/d2bot; do
    docker exec "$container" test -x "$c" 2>/dev/null && bin=$c && break
  done
fi
if [ -z "$bin" ]; then
  bin=$(docker exec "$container" sh -c 'ls -d /data/didle/releases/*/bin/d2bot 2>/dev/null | sort | tail -1' 2>/dev/null)
fi
bin=${bin:-d2bot}
ops() { docker exec "$container" "$bin" ops quest "$@"; }

log() { printf '%s\n' "$*"; }
status() { ops status 2>&1; }
# rows of one quest: "chunk<TAB>state"
rows_of() { printf '%s\n' "$1" | awk -F'\t' -v q="$2" '{ split($1, a, "/"); if (a[1] == q && NF >= 2) print a[2] "\t" $2 }'; }
kv() { printf '%s\n' "$1" | grep -o "$2=[^ ]*" | head -1 | cut -d= -f2-; }

# run_quest QUEST: 0 done, 1 exhausted.
run_quest() {
  local q=$1 try=1 iter=0 st rows verb out res outcome cool=0
  log "## $q"
  while [ $iter -lt 60 ]; do
    iter=$((iter + 1))
    st=$(status) || { log "- status failed: $(printf '%s' "$st" | tail -1)"; return 1; }
    rows=$(rows_of "$st" "$q")
    if [ -z "$rows" ]; then log "- no such quest on the board"; return 1; fi
    if ! printf '%s\n' "$rows" | grep -qv -P '\t(done|skipped)$'; then log "- DONE ($q)"; return 0; fi
    if printf '%s\n' "$rows" | grep -q -P '\t(accepted|in_progress|awaiting_handin)$'; then
      verb=""
    elif printf '%s\n' "$rows" | grep -q -P '\tblocked$'; then
      verb=retry
    else
      verb=start
    fi
    if [ -n "$verb" ]; then
      out=$(ops "$verb" "$q" --why "questloop try $try/$tries") ; res=$?
      log "- try $try/$tries: $verb -> $out"
      if [ $res -ne 0 ]; then
        if [ "$(kv "$out" code)" = cooldown ] && [ $cool -lt 5 ]; then
          cool=$((cool + 1)); log "  cooling down ${backoff}s"; $sleeper "$backoff"; continue
        fi
        try=$((try + 1))
        [ $try -gt $tries ] && { log "- FAILED ($q): refused, $out"; return 1; }
        $sleeper "$backoff"; continue
      fi
    fi
    out=$(ops wait "$q" --timeout "$wait_s"); res=$?
    log "- wait: \`$out\`"
    outcome=$(kv "$out" outcome)
    if [ "$outcome" = finished ]; then continue; fi   # the next chunk, if any, starts on the next pass
    try=$((try + 1))
    if [ $try -gt $tries ]; then
      log "- FAILED ($q) after $tries tries: $outcome, last step $(kv "$out" step), reason $(kv "$out" code) $(printf '%s' "$out" | sed -n 's/.*text=\(".*"\)$/\1/p')"
      return 1
    fi
    log "  backing off ${backoff}s"; $sleeper "$backoff"
  done
  log "- FAILED ($q): too many passes"; return 1
}

log "# questloop $container $target"
if [ "$target" = --all ]; then
  st=$(status) || { log "status failed"; exit 1; }
  quests=$(printf '%s\n' "$st" | awk -F'\t' 'NF >= 2 { split($1, a, "/"); if (!(a[1] in s)) { s[a[1]] = 1; print a[1] } }')
  [ -n "$quests" ] || { log "the board lists no quests"; exit 1; }
  for q in $quests; do
    rows=$(rows_of "$st" "$q")
    if ! printf '%s\n' "$rows" | grep -qv -P '\t(done|skipped)$'; then log "## $q"; log "- already done"; continue; fi
    if ! printf '%s\n' "$rows" | grep -v -P '\t(done|skipped)$' | grep -qv -P '\tlocked$'; then log "## $q"; log "- skipped: locked (needs an earlier quest)"; continue; fi
    run_quest "$q" || exit 1
  done
  log "ALL DONE"
  exit 0
fi
run_quest "$target" && { log "DONE"; exit 0; }
exit 1
