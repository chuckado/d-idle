#!/bin/bash
# usage (ON THE BUILDER, inside a labrun job, one job at a time): wave.sh DIR HOURS SPEED CLASS ARM:SEED...
# One wave of the survival A/B for the 45 minute builder lane: runs the listed (arm, seed) pairs of one class at once
# (abq.sh, AB_MAX = all of them) in DIR ($HOME/d2d-lab/<DIR>, bin/ holds d2bot.survab, prep.abmerc, d2d.lab, d2dslot.sh,
# agentrun.sh), logs to DIR/ab-<class>/<arm>/<seed>.log, and prints the abreport-ready layout. Build the bot first:
#   go build -o $HOME/d2d-lab/DIR/bin/d2bot.survab ./cmd/d2bot
set -u
[ $# -ge 5 ] || { echo "usage: wave.sh DIR HOURS SPEED CLASS ARM:SEED..." >&2; exit 64; }
D=$1 H=$2 SP=$3 CL=$4; shift 4
AB_DIR=$HOME/d2d-lab/$D
export AB_DIR AB_CLASS=$CL AB_SPEED=$SP AB_MAX=$# D2D_BIN=$AB_DIR/bin/d2d.lab
here=$(cd "$(dirname "$0")" && pwd)
cd "$AB_DIR" || exit 2
cp -f "$here/abrun.sh" "$here/abq.sh" "$AB_DIR/"
echo "wave $D class $CL speed $SP hours $H: $* $(date)"
nice -n 15 ./abq.sh "$H" "$@"
echo "wave done $(date)"
