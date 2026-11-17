#!/bin/sh
# On the builder: pictures of the watch page with a long quest caption and the lossless picture at the phone and desktop sizes, in /tmp/vf7.
set -u
rm -rf /tmp/vf7; mkdir -p /tmp/vf7
tools/uishots.sh --char --url "/watch?first=0&theme=dark&needs=0&quests=runlong&pal=1" --sizes "412x780 360x640 390x844 844x390 1366x768 1914x930" --tag vc --out /tmp/vf7
ls -la /tmp/vf7
