#!/bin/bash
# usage (ON THE HEAD BOX, repo root): wavelane.sh [--lane long|main] [--dry-run] DIR HOURS SPEED CLASS ARM:SEED...
# Starts one survival A/B wave (tools/survival/wave.sh) on the builder through labrun. The default lane is `long`
# (150 minute cap, beside the main lane; the lane docs are in the private ops repo). The 45 minute main lane cut a
# 90 virtual minute run at 45 real minutes, so the launcher first checks the estimate against the lane:
#   real minutes = HOURS*60/SPEED + 10 (prep, load, report; U, from the survhold2 wave), at most 80 percent of the cap.
# HOURS is virtual hours (1.5 = 90 virtual minutes). Detached with a phone push; follow with `labrun watch JOB`.
set -eu
lane=long dry=
while [ $# -gt 0 ]; do
  case $1 in
    --lane) lane=${2:-}; shift 2 ;;
    --dry-run) dry=--dry-run; shift ;;
    *) break ;;
  esac
done
[ $# -ge 5 ] || { echo "usage: wavelane.sh [--lane long|main] [--dry-run] DIR HOURS SPEED CLASS ARM:SEED..." >&2; exit 64; }
case $lane in long) cap=150 ;; main) cap=45 ;; *) echo "wavelane: lane must be long or main" >&2; exit 64 ;; esac
H=$2 SP=$3
est=$(awk -v h="$H" -v s="$SP" 'BEGIN { if (s <= 0 || h <= 0) exit 1; printf "%d", h * 60 / s + 10 + 0.999 }') \
  || { echo "wavelane: HOURS and SPEED must be positive numbers" >&2; exit 64; }
if [ $(( est * 100 )) -gt $(( cap * 80 )) ]; then
  echo "wavelane: about $est real minutes needed (HOURS $H at SPEED $SP) but the $lane lane allows $cap; use a higher SPEED, fewer HOURS or --lane long" >&2
  exit 65
fi
[ -x tools/survival/wave.sh ] || { echo "wavelane: tools/survival/wave.sh missing (run from the repo root)" >&2; exit 2; }
echo "wavelane: lane $lane (cap $cap min), estimate $est real minutes"
exec labrun builder --lane "$lane" --detach --notify $dry -- tools/survival/wave.sh "$@"
