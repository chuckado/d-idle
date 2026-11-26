#!/usr/bin/env python3
"""Flip a dark knob on a sandbox, watch 24 h, revert on a hard invariant breach (docs/process/dark-launch.md, card autorevert1).

  flip.py <instance> <knob>=<value> --why TEXT [--window 24h] [--dry-run]   flip now, open the window
  flip.py watch [--interval 60] [--once]                                    poll every open window; revert or close
  flip.py status                                                            list open windows

Stdlib only. Never prints tokens. No host is hard-coded: the transport is a command prefix from the environment or
local.env (via tools/env.sh):
  D2_FLIP_RUN      shell prefix that runs a command on the sandbox host (default: labrun $D2_LABRUN_HOST --)
  D2_FLIP_TUNING   tuning file inside the container, {instance} is replaced (default /data/instances/{instance}/state/tuning.json)
  D2_FLIP_BIN      the bot binary inside the container (default /data/didle/current/bin/d2bot)
  D2_FLIP_STATE    where open windows are kept (default ~/.cache/d2flip)
The container is named like the instance. Refuses while D2BOT_RUN is set (never in a measured run) and when the
instance already has an open window.
"""
import argparse
import json
import os
import shlex
import subprocess
import sys
import time
from datetime import datetime, timedelta, timezone

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
LEDGER = os.path.join(ROOT, "data", "darklaunch.json")
HARD_MISS_LIMIT = 10  # consecutive unreadable polls before the window counts as a breach


class FlipError(Exception):
    pass


def now_utc():
    return datetime.now(timezone.utc)


def iso(t):
    return t.strftime("%Y-%m-%dT%H:%M:%SZ")


def parse_iso(s):
    return datetime.strptime(s, "%Y-%m-%dT%H:%M:%SZ").replace(tzinfo=timezone.utc)


def parse_window(s):
    n, unit = s[:-1], s[-1:]
    if not n.isdigit() or unit not in "smhd":
        raise FlipError(f"bad --window {s!r} (use 90s, 30m, 24h, 1d)")
    return int(n) * {"s": 1, "m": 60, "h": 3600, "d": 86400}[unit]


def load_env():
    """Host values come from the environment, then local.env through tools/env.sh (the public convention)."""
    env = dict(os.environ)
    try:
        out = subprocess.run(["sh", "-c", '. "$1/tools/env.sh" && env', "sh", ROOT], capture_output=True, text=True, timeout=20, env=env).stdout
        for line in out.splitlines():
            k, _, v = line.partition("=")
            if k.startswith("D2_"):
                env.setdefault(k, v)
    except (OSError, subprocess.SubprocessError):
        pass
    return env


class Sandbox:
    """The four things the tool needs from a sandbox. Tests replace this class with a fake."""

    def __init__(self, instance, env):
        self.instance = instance
        run = env.get("D2_FLIP_RUN")
        if run is None:
            host = env.get("D2_LABRUN_HOST")
            if not host:
                raise FlipError("set D2_FLIP_RUN or D2_LABRUN_HOST in local.env")
            run = f"labrun {host} --"
        self.run = shlex.split(run)
        self.path = env.get("D2_FLIP_TUNING", "/data/instances/{instance}/state/tuning.json").format(instance=instance)
        self.bin = env.get("D2_FLIP_BIN", "/data/didle/current/bin/d2bot")

    def _x(self, args, stdin=None, timeout=120):
        r = subprocess.run(self.run + args, input=stdin, capture_output=True, text=True, timeout=timeout)
        if r.returncode != 0:
            raise FlipError(f"{' '.join(args[:3])} failed: {r.stderr.strip()[:200]}")
        return r.stdout

    def read_tuning(self):
        try:
            return self._x(["docker", "exec", self.instance, "cat", self.path])
        except FlipError:
            return None  # no file yet: the profile applies

    def write_tuning(self, text):
        """Atomic: a temp file and a rename, so the knob is never half set."""
        self._x(["docker", "exec", "-i", self.instance, "sh", "-c", f'cat > "{self.path}.new" && mv "{self.path}.new" "{self.path}"'], stdin=text)

    def remove_tuning(self):
        self._x(["docker", "exec", self.instance, "rm", "-f", self.path])

    def restart(self):
        self._x(["docker", "restart", self.instance], timeout=300)

    def status(self):
        """The decision-ring view (`ops status --why`); its `invariants` list has the counters."""
        return json.loads(self._x(["docker", "exec", self.instance, self.bin, "ops", "status", "--why"], timeout=30))


def hard_counts(view):
    """-> (since, {name: count}) of the hard counters; FlipError when the view has none (the bot predates the surface)."""
    inv = view.get("invariants")
    if inv is None:
        raise FlipError("status has no invariants list (bot too old)")
    hard = {c["name"]: c.get("count", 0) for c in inv if c.get("hard")}
    since = next((c.get("since", "") for c in inv), "")
    if not hard:
        raise FlipError("no hard counters yet (bot not ticking)")
    return since, hard


class Store:
    """One JSON file per open window."""

    def __init__(self, d):
        self.d = d
        os.makedirs(d, exist_ok=True)

    def path(self, instance):
        return os.path.join(self.d, instance + ".json")

    def get(self, instance):
        try:
            with open(self.path(instance)) as f:
                return json.load(f)
        except FileNotFoundError:
            return None

    def put(self, instance, w):
        tmp = self.path(instance) + ".new"
        with open(tmp, "w") as f:
            json.dump(w, f, indent=1)
        os.replace(tmp, self.path(instance))

    def drop(self, instance):
        try:
            os.remove(self.path(instance))
        except FileNotFoundError:
            pass

    def all(self):
        return sorted(n[:-5] for n in os.listdir(self.d) if n.endswith(".json"))


class Ledger:
    """data/darklaunch.json rows. Statuses used here: trial (window open), dark (reverted), enabled (window closed clean)."""

    def __init__(self, path=LEDGER):
        self.path = path

    def _load(self):
        with open(self.path) as f:
            return json.load(f)

    def row(self, name):
        for it in self._load()["items"]:
            if it["name"] == name:
                return dict(it)
        return None

    def update(self, name, **kv):
        d = self._load()
        for it in d["items"]:
            if it["name"] == name:
                it.update(kv)
                break
        else:
            raise FlipError(f"{name}: no row in {os.path.basename(self.path)}")
        tmp = self.path + ".new"
        with open(tmp, "w") as f:
            json.dump(d, f, indent=1, ensure_ascii=False)
            f.write("\n")
        os.replace(tmp, self.path)


class Flipper:
    def __init__(self, make_sandbox, store, ledger, clock=now_utc, sleep=time.sleep, out=print, env=None):
        self.make, self.store, self.ledger, self.clock, self.sleep, self.out = make_sandbox, store, ledger, clock, sleep, out
        self.env = os.environ if env is None else env

    # --- flip ---------------------------------------------------------------------------------------------
    def flip(self, instance, knob, value, why, window_s):
        if self.env.get("D2BOT_RUN"):
            raise FlipError("D2BOT_RUN is set: never flip in a measured run")
        if not why.strip():
            raise FlipError("--why is required")
        if self.store.get(instance):
            raise FlipError(f"{instance} already has an open flip window (one per sandbox)")
        row = self.ledger.row(knob)
        if row is None or row["status"] not in ("dark", "trial"):
            raise FlipError(f"{knob}: needs a dark row in data/darklaunch.json")
        sb = self.make(instance)
        before = sb.read_tuning()
        state = json.loads(before) if before else {}
        if not isinstance(state, dict):
            raise FlipError("tuning file is not an object")
        # Baseline: the counters of the running child before the flip. A flip during a breach is refused.
        since0, base0 = hard_counts(sb.status())
        ts = self.clock()
        tag = "bak-flip-" + ts.strftime("%Y%m%dT%H%M%SZ")
        old = (state.get("overrides") or {}).get(knob)
        state.setdefault("overrides", {})[knob] = value
        new_text = json.dumps(state, indent=1)
        w = {"instance": instance, "knob": knob, "old": old, "new": value, "why": why, "start": iso(ts), "window_s": window_s,
             "end": iso(ts + timedelta(seconds=window_s)), "tag": tag, "backup": before, "pending": "", "misses": 0,
             "baseline": base0, "since": since0}
        self.store.put(instance, w)  # the backup is on disk before the knob moves
        try:
            sb.write_tuning(new_text)
            sb.restart()
        except Exception as e:
            self.out(f"FLIP FAILED {instance} {knob}: {e}; restoring the backup")
            self._restore(sb, w)  # best effort; on failure the window stays 'pending' and watch retries
            if w["pending"]:
                self.store.put(instance, w)
                raise FlipError(f"flip failed and the restore is pending ({w['pending']}): run flip.py watch --once") from e
            self.store.drop(instance)
            raise FlipError(f"flip failed, tuning restored: {e}") from e
        try:  # the restart reset the counters: take the baseline of the new child
            w["since"], w["baseline"] = hard_counts(sb.status())
        except Exception:
            w["since"], w["baseline"] = "", {}
        self.store.put(instance, w)
        self.ledger.update(knob, status="trial", trial_started=iso(ts), trial_instance=instance, trial_value=value,
                           trial_window=f"{window_s}s", trial_why=why)
        self.out(f"FLIP {instance} {knob}={value} (was {old}) window {window_s}s until {w['end']} backup {tag}")
        return w

    # --- restore / revert ---------------------------------------------------------------------------------
    def _restore(self, sb, w):
        """Put the backup back and restart. Sets w['pending'] to the step that failed ('' when done)."""
        w["pending"] = ""
        if w["pending"] == "" and not w.get("restored"):
            try:
                if w["backup"] is None:
                    sb.remove_tuning()
                else:
                    sb.write_tuning(w["backup"])
                w["restored"] = True
            except Exception as e:
                w["pending"] = f"write: {e}"
                return
        try:
            sb.restart()
        except Exception as e:
            w["pending"] = f"restart: {e}"

    def revert(self, w, reason):
        sb = self.make(w["instance"])
        w["revert_reason"] = w.get("revert_reason") or reason
        self._restore(sb, w)
        if w["pending"]:
            self.store.put(w["instance"], w)
            self.out(f"REVERT FAILED {w['instance']} {w['knob']}: {w['pending']} (reason: {w['revert_reason']}); will retry")
            return False
        self.ledger.update(w["knob"], status="dark", revert_reason=w["revert_reason"], reverted_at=iso(self.clock()))
        self.store.drop(w["instance"])
        self.out(f"REVERT {w['instance']} {w['knob']}: {w['revert_reason']}; tuning restored from {w['tag']}, child restarted, row dark")
        return True

    def close(self, w):
        self.ledger.update(w["knob"], status="enabled", enabled_at=iso(self.clock()))
        self.store.drop(w["instance"])  # the backup tag goes with the state file
        self.out(f"CLOSE {w['instance']} {w['knob']}: no hard breach for {w['window_s']}s; row enabled, backup {w['tag']} removed")

    # --- watch --------------------------------------------------------------------------------------------
    def poll(self, instance):
        """One poll of one window. Returns 'open', 'closed' or 'reverted'."""
        w = self.store.get(instance)
        if w is None:
            return "closed"
        if w.get("revert_reason") and not w.get("done"):  # a revert that failed earlier
            return "reverted" if self.revert(w, w["revert_reason"]) else "open"
        sb = self.make(instance)
        breach = ""
        try:
            since, counts = hard_counts(sb.status())
            w["misses"] = 0
            if since != w["since"]:  # the child restarted: its counters begin at zero again
                w["since"], w["baseline"] = since, {}
            over = {n: c - w["baseline"].get(n, 0) for n, c in counts.items() if c > w["baseline"].get(n, 0)}
            if over:
                breach = "hard invariant breach: " + ",".join(f"{n}+{d}" for n, d in sorted(over.items()))
        except Exception as e:
            w["misses"] = w.get("misses", 0) + 1
            if w["misses"] >= HARD_MISS_LIMIT:
                breach = f"status unreadable for {w['misses']} polls: {e}"
        if breach:
            return "reverted" if self.revert(w, breach) else "open"
        if self.clock() >= parse_iso(w["end"]):
            self.close(w)
            return "closed"
        self.store.put(instance, w)
        return "open"

    def watch(self, interval=60, once=False):
        while True:
            names = self.store.all()
            states = [self.poll(n) for n in names]
            if once or "open" not in states:
                return states
            self.sleep(interval)


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("args", nargs="+", help="<instance> <knob>=<value> | watch | status")
    ap.add_argument("--why", default="")
    ap.add_argument("--window", default="24h")
    ap.add_argument("--interval", type=int, default=60)
    ap.add_argument("--once", action="store_true")
    a = ap.parse_args(argv)
    env = load_env()
    store = Store(env.get("D2_FLIP_STATE") or os.path.expanduser("~/.cache/d2flip"))
    f = Flipper(lambda inst: Sandbox(inst, env), store, Ledger(), env=env)
    try:
        if a.args[0] == "watch":
            f.watch(a.interval, a.once)
        elif a.args[0] == "status":
            for n in store.all():
                w = store.get(n)
                print(f"{n} {w['knob']}={w['new']} until {w['end']} pending={w['pending'] or '-'}")
        else:
            if len(a.args) != 2 or "=" not in a.args[1]:
                raise FlipError("usage: flip.py <instance> <knob>=<value> --why TEXT [--window 24h]")
            knob, _, v = a.args[1].partition("=")
            try:
                val = float(v)
            except ValueError:
                raise FlipError(f"value {v!r} is not a number (tuning overrides are numbers)")
            f.flip(a.args[0], knob, val, a.why, parse_window(a.window))
    except FlipError as e:
        print(f"flip: {e}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
