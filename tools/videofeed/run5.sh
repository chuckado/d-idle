#!/bin/sh
# On the builder: pictures of the watch page with the stream monitor on (pfdebug=1) at three window sizes, in /tmp/vf6.
set -u
rm -rf /tmp/vf6; mkdir -p /tmp/vf6
tools/uishots.sh --char --url "/watch?first=0&theme=dark&needs=2&pal=1&pfdebug=1" --sizes "1075x700 1914x930 390x844" --tag stats --out /tmp/vf6
ls -la /tmp/vf6
