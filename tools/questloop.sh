#!/usr/bin/env bash
# Drive quests through the operator control of a running container (docs/operator.md). Talks only to
# `docker exec <container> d2bot ops`; it holds no token and needs none. Writes a markdown log to stdout.
# Usage: tools/questloop.sh <container> <quest|--all> [--max-tries 3] [--wait 300] [--stall 6] [--backoff 65]
#        [--budget 10800] [--tries-backoff 1200] [--deadline 07:30] [--bin PATH]
# Per quest: status -> start (retry when parked; NOTHING when a chunk already runs, so a restart is safe) -> wait in
# --wait second windows (one heartbeat line each). A window that ends with moved progress (step or state changed) is
# slow progress, not a failed try; --stall windows without any move, a park, a cancel or a refusal count as a try.
# After --max-tries tries it backs off --tries-backoff seconds and goes on with the SAME quest until the per-chunk
# --budget (seconds) or the --deadline (HH:MM local, or @EPOCH) passes. A chunk done moves on; --all continues with
# the next startable quest in board order and finishes when none is left; locked quests are skipped with a line.
# Exit: 0 everything wanted is done, 1 something failed or the deadline passed (evidence in the log), 2 usage.
# Test hooks: QUESTLOOP_SLEEP (command, default sleep), QUESTLOOP_NOW (command printing epoch, default date +%s),
# D2BOT_BIN (same as --bin). Hero deaths are not exposed by the board, so they are not a stop condition here.
set -u
usage() { sed -n '2,15p' "$0" | sed 's/^# \{0,1\}//' >&2; exit 2; }
[ $# -ge 2 ] || usage
container=$1; target=$2; shift 2
tries=3; wait_s=300; stall_max=6; backoff=65; budget=10800; tries_backoff=1200; deadline_s=07:30; bin=${D2BOT_BIN:-}
while [ $# -gt 0 ]; do
  case $1 in
    --max-tries) tries=${2:?}; shift 2 ;;
    --wait) wait_s=${2:?}; shift 2 ;;
    --stall) stall_max=${2:?}; shift 2 ;;
    --backoff) backoff=${2:?}; shift 2 ;;
    --budget) budget=${2:?}; shift 2 ;;
    --tries-backoff) tries_backoff=${2:?}; shift 2 ;;
    --deadline) deadline_s=${2:?}; shift 2 ;;
    --bin) bin=${2:?}; shift 2 ;;
    *) usage ;;
  esac
done
sleeper=${QUESTLOOP_SLEEP:-sleep}
now() { ${QUESTLOOP_NOW:-date +%s}; }
t_start=$(now)
case $deadline_s in
  @*) deadline=${deadline_s#@} ;;
  *) deadline=$(date -d "$deadline_s" +%s 2>/dev/null) || usage
     [ "$deadline" -le "$t_start" ] && deadline=$((deadline + 86400)) ;;
esac

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

# run_quest QUEST: 0 done, 1 failed (budget), 3 skipped (locked), 4 deadline.
run_quest() {
  local q=$1 try=1 st rows verb out res outcome t0 stall=0 lastp="" cool=0 p ocode
  t0=$(now)
  log "## $q"
  while :; do
    if [ "$(now)" -ge "$deadline" ]; then log "- DEADLINE passed, stopping"; return 4; fi
    if [ $(( $(now) - t0 )) -ge "$budget" ]; then log "- FAILED ($q): chunk budget of ${budget}s used up"; return 1; fi
    st=$(status) || { log "- status failed: $(printf '%s' "$st" | tail -1)"; $sleeper "$backoff"; continue; }
    rows=$(rows_of "$st" "$q")
    if [ -z "$rows" ]; then log "- no such quest on the board"; return 3; fi
    if ! printf '%s\n' "$rows" | grep -qv -P '\t(done|skipped)$'; then log "- DONE ($q)"; return 0; fi
    if ! printf '%s\n' "$rows" | grep -v -P '\t(done|skipped)$' | grep -qv -P '\t(locked|unavailable)$'; then
      log "- skipped: locked or unavailable"; return 3
    fi
    if printf '%s\n' "$rows" | grep -q -P '\t(accepted|in_progress|awaiting_handin)$'; then
      verb=""
    elif printf '%s\n' "$rows" | grep -q -P '\tblocked$'; then
      verb=retry
    else
      verb=start
    fi
    if [ -n "$verb" ]; then
      out=$(ops "$verb" "$q" --why "questloop try $try/$tries"); res=$?
      log "- try $try/$tries: $verb -> $out"
      if [ $res -ne 0 ]; then
        ocode=$(kv "$out" code)
        case $ocode in
          cooldown) $sleeper "$backoff"; continue ;;
          rate_limited) log "  rate limited, waiting ${tries_backoff}s"; $sleeper "$tries_backoff"; continue ;;
          not_ready|locked|unavailable) log "- skipped: refused ($ocode)"; return 3 ;;
        esac
        try=$((try + 1)); $sleeper "$backoff"
        if [ $try -gt $tries ]; then log "  tries used up, backing off ${tries_backoff}s and going on"; $sleeper "$tries_backoff"; try=1; fi
        continue
      fi
    fi
    out=$(ops wait "$q" --timeout "$wait_s"); res=$?
    outcome=$(kv "$out" outcome)
    log "- [$(( $(now) - t0 ))s, try $try/$tries] $outcome step=$(kv "$out" step) chunk=$(kv "$out" chunk) progress=$(kv "$out" progress)"
    case $outcome in
      finished) try=1; stall=0; lastp=""; t0=$(now); continue ;;
      timeout)
        p="$(kv "$out" step)/$(kv "$out" progress)"
        if [ "$p" != "$lastp" ]; then lastp=$p; stall=0; continue; fi
        stall=$((stall + 1))
        [ $stall -lt $stall_max ] && continue
        log "  no progress for $stall windows: counts as a failed try"; stall=0 ;;
      *) log "  \`$out\`" ;;
    esac
    try=$((try + 1))
    if [ $try -gt $tries ]; then log "  tries used up (last step $(kv "$out" step), reason $(kv "$out" code)), backing off ${tries_backoff}s and going on"; $sleeper "$tries_backoff"; try=1
    else $sleeper "$backoff"; fi
  done
}

log "# questloop $container $target"
if [ "$target" = --all ]; then
  failed=" "; failures=0
  while :; do
    st=$(status) || { log "status failed"; exit 1; }
    quests=$(printf '%s\n' "$st" | awk -F'\t' 'NF >= 2 { split($1, a, "/"); if (!(a[1] in s)) { s[a[1]] = 1; print a[1] } }')
    [ -n "$quests" ] || { log "the board lists no quests"; exit 1; }
    moved=0
    for q in $quests; do
      case $failed in *" $q "*) continue ;; esac
      rows=$(rows_of "$st" "$q")
      printf '%s\n' "$rows" | grep -qv -P '\t(done|skipped)$' || continue
      printf '%s\n' "$rows" | grep -v -P '\t(done|skipped)$' | grep -qv -P '\t(locked|unavailable)$' || continue
      run_quest "$q"; rc=$?
      case $rc in
        0) moved=1; break ;;
        4) log "STOPPED at the deadline"; exit 1 ;;
        *) failed="$failed$q "; failures=$((failures + 1)) ;;
      esac
    done
    [ $moved -eq 1 ] || break
  done
  log "NOTHING MORE STARTABLE ($failures quest(s) failed)"
  [ $failures -eq 0 ] && { log "ALL DONE"; exit 0; }
  exit 1
fi
run_quest "$target"; rc=$?
[ $rc -eq 0 ] && { log "DONE"; exit 0; }
exit 1
