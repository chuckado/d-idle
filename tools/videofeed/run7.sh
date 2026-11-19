#!/bin/sh
# On the builder: pictures of the long-caption page at phone and desktop sizes (/tmp/vf8), then the fit check on the video pages
# and the must-fail defect page.
set -u
rm -rf /tmp/vf8; mkdir -p /tmp/vf8
tools/uishots.sh --char --url "/watch?first=0&theme=dark&needs=0&quests=runlong&pal=1" --sizes "412x780 360x640 390x844 844x390 1366x768 1914x930" --tag vc --out /tmp/vf8
for o in quest-runlong quest-runlong-light viddefect-ctl default-live-pal quest-running; do
	echo "== fit $o"
	tools/uishots.sh --fit --only "$o" 2>&1 | grep -E "^(ok|FAIL|fit)" | cut -c1-200 | head -n 40
done
ls /tmp/vf8
