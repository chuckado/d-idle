#!/usr/bin/env bash
# Offline test of tools/questloop.sh with a fake `docker` shim and a fake clock (no container, no game).
set -u
cd "$(dirname "$0")/.."
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
cat > "$tmp/bin/docker" <<'S'
#!/usr/bin/env bash
# docker exec <container> <bin> ops quest <args>  (also: test -x / sh -c)
shift; shift
case $1 in test) exit 0 ;; sh) exit 0 ;; esac
shift; shift; shift
verb=$1
echo "$*" >> "$FAKE/calls"
case $verb in
  status) cat "$FAKE/status.$(cat "$FAKE/phase")" ;;
  handin) echo ok=true code= state=awaiting_handin reason=\"x\"; echo 2 > "$FAKE/phase" ;;
  start|retry) echo ok=true code= state=accepted reason=\"x\"; echo 1 > "$FAKE/phase" ;;
  wait)
    n=$(( $(cat "$FAKE/n") + 1 )); echo $n > "$FAKE/n"
    echo $(( $(cat "$FAKE/clock") + 300 )) > "$FAKE/clock"
    r=$(sed -n "${n}p" "$FAKE/waits"); [ -n "$r" ] || r=$(tail -1 "$FAKE/waits")
    case $r in
      finished) cat "$FAKE/after" > "$FAKE/phase"; echo 2 > "$FAKE/after"; echo 'RESULT outcome=finished quest=a chunk=c1 state=done step= progress=9 code= text=""' ;;
      blocked) echo 3 > "$FAKE/phase"; echo 'RESULT outcome=blocked quest=a chunk=c1 state=blocked step=walk progress=4 code=gave_up text="stuck"'; exit 1 ;;
      timeout:*) IFS=: read -r _ st pr <<<"$r"; echo "RESULT outcome=timeout quest=a chunk=c1 state=in_progress step=$st progress=$pr code= text=\"\""; exit 2 ;;
    esac ;;
esac
S
cat > "$tmp/sleep.sh" <<'S'
#!/usr/bin/env bash
echo $(( $(cat "$FAKE/clock") + $1 )) > "$FAKE/clock"
echo "$1" >> "$FAKE/sleeps"
S
chmod +x "$tmp/bin/docker" "$tmp/sleep.sh"
export PATH="$tmp/bin:$PATH" FAKE="$tmp" QUESTLOOP_SLEEP="$tmp/sleep.sh" QUESTLOOP_NOW="cat $tmp/clock"
hdr='board enabled=true difficulty=normal why=""'
printf '%s\na/c1\tready\tstep=\tcode=\t\nb/c1\tlocked\tstep=\tcode=\t\n' "$hdr" > "$tmp/status.0"
printf '%s\na/c1\tin_progress\tstep=walk\tcode=\t\nb/c1\tlocked\tstep=\tcode=\t\n' "$hdr" > "$tmp/status.1"
printf '%s\na/c1\tdone\tstep=\tcode=\t\nb/c1\tlocked\tstep=\tcode=\t\n' "$hdr" > "$tmp/status.2"
printf '%s\na/c1\tblocked\tstep=walk\tcode=gave_up\tstuck\nb/c1\tlocked\tstep=\tcode=\t\n' "$hdr" > "$tmp/status.3"
printf '%s\na/c1\tawaiting_handin\tstep=\tcode=\t\nb/c1\tin_progress\tstep=walk\tcode=\t\n' "$hdr" > "$tmp/status.4"
printf '%s\na/c1\tawaiting_handin\tstep=\tcode=\t\nb/c1\tdone\tstep=\tcode=\t\n' "$hdr" > "$tmp/status.5"
fail=0
check() { if ! grep -q -- "$2" <<<"$1"; then echo "FAIL: $3 (missing: $2)"; echo "$1"; fail=1; fi; }
nocheck() { if grep -q -- "$2" <<<"$1"; then echo "FAIL: $3 (unexpected: $2)"; fail=1; fi; }
reset() { echo 2 > "$tmp/after"; echo "$1" > "$tmp/phase"; : > "$tmp/calls"; : > "$tmp/sleeps"; echo 0 > "$tmp/n"; echo 1000000 > "$tmp/clock"; printf '%s\n' "${@:2}" > "$tmp/waits"; }
run() { tools/questloop.sh box "$@" --bin /x/d2bot --deadline @2000000; }

reset 0 finished
out=$(run a); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: success exit $rc"; fail=1; }
check "$out" "DONE" success
check "$(cat "$tmp/calls")" "start a --why questloop try 1/3" "start with a reason"

reset 0 blocked finished
out=$(run a); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: retry exit $rc"; fail=1; }
check "$(cat "$tmp/calls")" "retry a --why questloop try 2/3" "retry after a park"

# tries used up does not end the run: back off and go on with the same quest
reset 0 blocked blocked finished
out=$(run a --max-tries 2); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: exhausted-then-done exit $rc"; echo "$out"; fail=1; }
check "$out" "tries used up" "tries backoff"
check "$(cat "$tmp/sleeps")" "^1200$" "20 minute backoff"

# slow progress is not a failed try, and a restart on a running chunk does not call start
reset 1 timeout:walk:1 timeout:smith:2 timeout:chicken:3 finished
out=$(run a); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: progress exit $rc"; echo "$out"; fail=1; }
nocheck "$(cat "$tmp/calls")" "^start" "idempotent restart"
nocheck "$(cat "$tmp/calls")" "^retry" "no retry on progress"
check "$out" "step=chicken" "heartbeat line"

# a real stall with a small budget fails with evidence
reset 1 timeout:walk:1
out=$(run a --stall 2 --budget 3000); rc=$?
[ $rc -eq 1 ] || { echo "FAIL: stall exit $rc"; fail=1; }
check "$out" "FAILED (a): chunk budget" "budget"
check "$out" "no progress for 2 windows" "stall counted"

# the deadline stops the run
reset 1 timeout:walk:1
out=$(tools/questloop.sh box a --bin /x/d2bot --deadline @1000100); rc=$?
[ $rc -eq 1 ] || { echo "FAIL: deadline exit $rc"; fail=1; }
check "$out" "DEADLINE" "deadline"

# --all: a done, b locked is skipped, finishes when nothing is startable
reset 0 finished
out=$(run --all); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: --all exit $rc"; echo "$out"; fail=1; }
check "$out" "NOTHING MORE STARTABLE" "--all ends"
check "$out" "ALL DONE" "--all"

# another quest runs: wait on it first, then hand in (never start) the awaiting one
reset 4 finished; echo 5 > "$tmp/after"
out=$(run a); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: handin exit $rc"; echo "$out"; fail=1; }
check "$out" "waiting on b" "decision line"
calls=$(cat "$tmp/calls")
check "$calls" "^wait b" "waits on the active quest"
check "$calls" "^handin a" "handin verb"
nocheck "$calls" "^start" "no start"
[ "$(grep -n '^wait b' "$tmp/calls" | head -1 | cut -d: -f1)" -lt "$(grep -n '^handin a' "$tmp/calls" | head -1 | cut -d: -f1)" ] || { echo "FAIL: handin before the active quest finished"; fail=1; }

[ $fail -eq 0 ] && echo "questloop_test: ok"
exit $fail
