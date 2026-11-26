"""Shared loader for the dark-launch ledger (data/darklaunch.json) and docs/knobs.md."""
import json
import re
from datetime import date

STATUSES = ("dark", "trial", "default-on", "retired")
OPEN = ("dark", "trial")
MARKER = "[dark-launch]"
FIELDS = ("name", "built_by", "dark_since", "why_off", "enable_criteria", "owner", "review_by", "status", "criteria_met")
ROW = re.compile(r"^\| ([a-zA-Z]+\.[A-Za-z0-9]+) \| ([^|]*) \|.*$")


def load_ledger(path):
    with open(path) as f:
        return json.load(f)["items"]


def knob_rows(path):
    """key -> (default, full row text) for every row of the generated knobs table."""
    out = {}
    with open(path) as f:
        for line in f:
            m = ROW.match(line.rstrip("\n"))
            if m:
                out[m.group(1)] = (m.group(2).strip(), line)
    return out


def parse_date(s):
    return date.fromisoformat(s)
