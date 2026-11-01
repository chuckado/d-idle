"""Tiny Messages API client shared by pr_review.py and haiku_tick.py (stdlib only).

The key is only ever sent as the x-api-key header. It is never printed, logged or put in an exception text.
"""
import json
import urllib.request

API_URL = "https://api.anthropic.com/v1/messages"


def post_json(url, headers, body, timeout=60):
    """Default HTTP layer: returns (status, parsed_json). Tests pass a fake with the same signature."""
    req = urllib.request.Request(url, data=json.dumps(body).encode(), headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=timeout) as r:
            return r.status, json.loads(r.read() or b"{}")
    except urllib.error.HTTPError as e:
        try:
            return e.code, json.loads(e.read() or b"{}")
        except ValueError:
            return e.code, {}


def message(key, model, system, user, max_tokens, http=post_json):
    """One Messages API call. Returns the text. Raises RuntimeError with a key-free message on failure."""
    headers = {"x-api-key": key, "anthropic-version": "2023-06-01", "content-type": "application/json"}
    body = {"model": model, "max_tokens": max_tokens, "system": system,
            "messages": [{"role": "user", "content": user}]}
    status, data = http(API_URL, headers, body)
    if status != 200:
        kind = (data.get("error") or {}).get("type", "unknown") if isinstance(data, dict) else "unknown"
        raise RuntimeError("anthropic api http %s (%s)" % (status, kind))
    return "".join(b.get("text", "") for b in data.get("content", []) if b.get("type") == "text").strip()
