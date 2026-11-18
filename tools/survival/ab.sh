#!/bin/bash
# usage (ON THE WORKER): ab.sh HOURS SEED... ; interleaved arms per seed (legacy then margin side by side), at most AB_MAX (default 2)
# of mine at once (one slot each, about one core each). Then: go run ./tools/survival/abreport ~/d2d-lab/survab/ab
[ $# -ge 2 ] || { echo "usage: ab.sh HOURS SEED..." >&2; exit 64; }
here=$(cd "$(dirname "$0")" && pwd); H=$1; shift
n=0
for s in "$@"; do for arm in legacy margin; do
  [ $n -lt ${AB_MAX:-2} ] || { wait -n; n=$((n-1)); }
  "$here/abrun.sh" $arm $s $H &
  n=$((n+1)); sleep 5
done; done
wait
