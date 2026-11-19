#!/bin/bash
# CAIN-L3 ON THE BUILDER (main lane via labrun): the real d2bot (ladder + watch like the sandbox, goals campaign:normal_act1)
# on one d2d lab slot, a prepared throwaway Sorceress (SETUP via tools/ab/prep, ledgered, CHEATED: mechanics only). poll.py taps Start
# and Hand in on the board for search_for_cain. usage: run.sh SEED CLVL WALLMIN [SLOT]
export PATH=$PATH:/opt/go/bin
W=/data/d2d-lab/cainl3; cd "$W" || exit 2
S=${1:-11} CLVL=${2:-18} WALLMIN=${3:-38} SLOT=${4:-3}
SRC=$(cd "$(dirname "$0")" && pwd)
export D2D_AGENT=cainl3
exec "$W/bin/d2dslot.sh" run -n $SLOT -w 600 -t $((WALLMIN*60+300)) -- bash -c '
cd '$W'; S='$S'; export D2BOT_INSTANCE=cl3 D2D_RUN=on
export D2D_BIN=$HOME/d2d-lab/survhold/bin/d2d.lab D2D_DATA=$HOME/d2d-lab/data
rm -rf data-run; mkdir -p logs data-run/instances/cl3/state; inst=data-run/instances/cl3; home=$inst/d2d
./bin/prep.cainl3 --home "$PWD/$home" --name Cainy-$S --seed $S --class 4 --clvl '$CLVL' --gold 3000 --ledger "$PWD/$inst/prep-ledger.jsonl" > logs/prep.log 2>&1 || { echo prep failed; exit 3; }
head -c 24 /dev/urandom | od -An -tx1 | tr -d " \n" > data-run/wtoken; chmod 600 data-run/wtoken
WP=$((9300+D2D_SLOT*10))
nice -n 10 ./bin/d2bot.cainl3 --backend d2d --managed --d2d-class -1 --char Cainy-$S --seed $S \
 --toggle autoloot=off,autoparty=off --d2d-args "--lockstep" --d2d-linkwait 60s \
 --gamedir $D2D_DATA --data '$W'/data-run --d2d-sock $D2D_SOCK --addr $D2D_BOT_ADDR \
 --ladder --watch --watch-addr 127.0.0.1:$WP --watch-answer-addr 127.0.0.1:$((WP+1)) --watch-token-file '$W'/data-run/wtoken \
 --goals campaign:normal_act1 --run-for '$WALLMIN'm > logs/bot.log 2>&1 &
BOT=$!
sleep 20
WTOKEN_FILE='$W'/data-run/wtoken python3 -I '$SRC'/poll.py http://127.0.0.1:$WP http://127.0.0.1:$((WP+1)) '$((WALLMIN-3))' > logs/poll.log 2>&1
echo "poll exit $?" >> logs/poll.log
kill -TERM $BOT 2>/dev/null; for i in $(seq 1 40); do kill -0 $BOT 2>/dev/null || break; sleep 1; done
echo "bot exited"; ./bin/d2cheat ledger 2>&1 | tail -3 > logs/ledger.txt'
