#!/bin/sh
# SKGROUND2: summarise `test_sk_ground --blizz` lines (stdin): shard count, offsets outside game.exe's -6..5, histogram, frames
grep '^S' | awk '{x[$3]++; y[$4]++; if($3>5||$3<-6||$4>5||$4<-6) bad++; n++; f[$2]++}
END{printf "shards %d, outside game.exe -6..5: %d (%.1f%%)\n", n, bad+0, 100*bad/n;
for(k=-8;k<=8;k++) printf "%d:%d/%d ", k, x[k], y[k]; print " (x/y histogram)";
c=0; for(k in f) c++; printf "distinct shard frames %d (a cast has 25)\n", c}'
