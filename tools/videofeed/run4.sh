#!/bin/sh
# On the builder: pictures of the watch page with the lossless feed (pal=1) and with MJPEG at several window sizes, in /tmp/vf5.
set -u
rm -rf /tmp/vf5; mkdir -p /tmp/vf5
tools/uishots.sh --char --url "/watch?first=0&theme=dark&needs=2&pal=1" --sizes "1366x768 1075x700 1920x1080 390x844 1024x768" --tag pal --out /tmp/vf5
tools/uishots.sh --char --url "/watch?first=0&theme=dark&needs=2" --sizes "1075x700 390x844" --tag mjpeg --out /tmp/vf5
ls -la /tmp/vf5
