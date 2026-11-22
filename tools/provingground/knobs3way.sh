#!/bin/bash
# A/B/C runs of one scenario with knob sets (env D2SCEN_KNOBS), serial, on the builder long lane (card mapverify).
#   tools/provingground/knobs3way.sh <scenario-id> <class> "<seeds>" [<label>=<knobs> ...]
set -u
id=$1 cls=$2 seeds=$3; shift 3
[ $# -gt 0 ] || set -- off= grid=nav.fullGrid=1 gridwarp=nav.fullGrid=1,nav.warpRead=1
tag=$(date +%H%M%S)
for seed in $seeds; do for c in "$@"; do
  lab=${c%%=*} knobs=${c#*=}
  echo "=== $id $cls $seed $lab [$knobs]"
  D2SCEN_KNOBS=$knobs tools/provingground/run.sh -id "$id" -class "$cls" -seed "$seed" -parallel 1 -out "provingground-out/k3-$tag/$lab-$cls-$seed" 2>&1 | tail -3
done; done
