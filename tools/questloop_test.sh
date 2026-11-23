#!/usr/bin/env bash
# Offline test of tools/questloop.sh with a fake `docker` shim (no container, no game).
set -u
cd "$(dirname "$0")/.."
tmp=$(mktemp -d); trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
cat > "$tmp/bin/docker" <<'S'
#!/usr/bin/env bash
# docker exec <container> <bin> ops quest <args>  (also: test -x / sh -c)
shift  # exec
shift  # container
case $1 in test) exit 0 ;; sh) exit 0 ;; esac
shift  # bin
shift; shift # ops quest
verb=$1
echo "$*" >> "$FAKE/calls"
n=$(cat "$FAKE/n" 2>/dev/null || echo 0)
case $verb in
  status) cat "$FAKE/status.$(cat "$FAKE/phase")" ;;
  start|retry) echo ok=true code= state=accepted reason=\"x\"; echo 1 > "$FAKE/phase" ;;
  wait)
    n=$((n + 1)); echo $n > "$FAKE/n"
    r=$(sed -n "${n}p" "$FAKE/waits")
    case $r in
      finished) echo 0 > "$FAKE/phase"; echo 2 > "$FAKE/phase"; echo 'RESULT outcome=finished quest=a chunk=c1 state=done step= code= text=""'; exit 0 ;;
      blocked) echo 3 > "$FAKE/phase"; echo 'RESULT outcome=blocked quest=a chunk=c1 state=blocked step=walk code=gave_up text="stuck"'; exit 1 ;;
    esac ;;
esac
S
chmod +x "$tmp/bin/docker"
export PATH="$tmp/bin:$PATH" FAKE="$tmp" QUESTLOOP_SLEEP=true
printf 'board enabled=true difficulty=normal why=""\na/c1\tready\tstep=\tcode=\t\nb/c1\tlocked\tstep=\tcode=\t\n' > "$tmp/status.0"
printf 'board enabled=true difficulty=normal why=""\na/c1\tin_progress\tstep=walk\tcode=\t\nb/c1\tlocked\tstep=\tcode=\t\n' > "$tmp/status.1"
printf 'board enabled=true difficulty=normal why=""\na/c1\tdone\tstep=\tcode=\t\nb/c1\tready\tstep=\tcode=\t\n' > "$tmp/status.2"
printf 'board enabled=true difficulty=normal why=""\na/c1\tblocked\tstep=walk\tcode=gave_up\tstuck\nb/c1\tlocked\tstep=\tcode=\t\n' > "$tmp/status.3"
fail=0
check() { if ! grep -q -- "$2" <<<"$1"; then echo "FAIL: $3 (missing: $2)"; echo "$1"; fail=1; fi; }

# 1. success on the first try
echo 0 > "$tmp/phase"; : > "$tmp/calls"; echo 0 > "$tmp/n"; printf 'finished\n' > "$tmp/waits"
out=$(tools/questloop.sh box a --bin /x/d2bot); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: success exit $rc"; fail=1; }
check "$out" "DONE" success
check "$(cat "$tmp/calls")" "start a --why questloop try 1/3" "start with a reason"

# 2. park, retry, then success
echo 0 > "$tmp/phase"; : > "$tmp/calls"; echo 0 > "$tmp/n"; printf 'blocked\nfinished\n' > "$tmp/waits"
out=$(tools/questloop.sh box a --bin /x/d2bot); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: retry exit $rc"; fail=1; }
check "$(cat "$tmp/calls")" "retry a --why questloop try 2/3" "retry after a park"

# 3. exhausts tries: exit 1 with evidence
echo 0 > "$tmp/phase"; : > "$tmp/calls"; echo 0 > "$tmp/n"; printf 'blocked\nblocked\n' > "$tmp/waits"
out=$(tools/questloop.sh box a --bin /x/d2bot --max-tries 2); rc=$?
[ $rc -eq 1 ] || { echo "FAIL: exhausted exit $rc"; fail=1; }
check "$out" "FAILED (a) after 2 tries: blocked, last step walk, reason gave_up" "evidence"

# 4. --all: a done, b locked is skipped
echo 0 > "$tmp/phase"; : > "$tmp/calls"; echo 0 > "$tmp/n"; printf 'finished\n' > "$tmp/waits"
out=$(tools/questloop.sh box --all --bin /x/d2bot); rc=$?
[ $rc -eq 0 ] || { echo "FAIL: --all exit $rc"; echo "$out"; fail=1; }
check "$out" "skipped: locked" "--all skips locked"
check "$out" "ALL DONE" "--all"

[ $fail -eq 0 ] && echo "questloop_test: ok"
exit $fail
