#!/bin/bash
# usage (ON THE WORKER): abq.sh HOURS ARM:SEED... ; runs the listed (arm, seed) pairs through abrun.sh with at most AB_MAX
# (default 2) abrun runs alive on this host, counted over all drivers. For queues that ab.sh (every seed, both arms)
# cannot express: a retuned margin arm against an existing legacy arm, or the missing runs of a half-finished batch.
[ $# -ge 2 ] || { echo "usage: abq.sh HOURS ARM:SEED..." >&2; exit 64; }
here=$(cd "$(dirname "$0")" && pwd); H=$1; shift
alive() { pgrep -fc "^(\./)?bin/(d2bot.survab --backend|prep.abmerc --home)"; }  # a run is its prep, then its bot
for j in "$@"; do
  while [ "$(alive)" -ge "${AB_MAX:-2}" ]; do sleep 20; done
  "$here/abrun.sh" "${j%%:*}" "${j##*:}" "$H" &
  sleep 8
done
wait
