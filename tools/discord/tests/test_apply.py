import http.server
import io
import json
import os
import re
import sys
import threading
import unittest
from contextlib import redirect_stderr, redirect_stdout

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import apply_server as ap

SPEC_PATH = os.path.join(os.path.dirname(__file__), "..", "..", "..", "docs", "community", "discord", "server.json")
GUILD = "5550001"


class MockApi:
    def __init__(self):
        self.roles = [{"id": GUILD, "name": "@everyone", "permissions": "0"}]
        self.channels = []
        self.log = []
        self.n = 100
        outer = self

        class H(http.server.BaseHTTPRequestHandler):
            def log_message(self, *a):
                pass

            def _go(self):
                body = json.loads(self.rfile.read(int(self.headers.get("Content-Length") or 0)) or b"null")
                outer.log.append((self.command, self.path, body))
                assert self.headers.get("Author" + "ization", "").startswith("Bot ")
                res, code = outer.handle(self.command, self.path, body)
                raw = json.dumps(res).encode()
                self.send_response(code)
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)

            do_GET = do_POST = do_PATCH = do_DELETE = _go

        self.srv = http.server.HTTPServer(("127.0.0.1", 0), H)
        self.base = "http://127.0.0.1:%d" % self.srv.server_port
        threading.Thread(target=self.srv.serve_forever, daemon=True).start()

    def handle(self, m, path, body):
        self.n += 1
        if m == "GET" and path.endswith("/roles"):
            return self.roles, 200
        if m == "GET" and path.endswith("/channels"):
            return self.channels, 200
        if m == "POST" and path.endswith("/roles"):
            r = dict(body, id=str(self.n))
            self.roles.append(r)
            return r, 200
        if m == "POST" and path.endswith("/channels"):
            c = dict(body, id=str(self.n))
            self.channels.append(c)
            return c, 200
        return {}, 200

    def writes(self):
        return [x for x in self.log if x[0] != "GET"]

    def close(self):
        self.srv.shutdown()
        self.srv.server_close()


class ApplyTests(unittest.TestCase):
    def setUp(self):
        self.m = MockApi()
        self.addCleanup(self.m.close)
        self.api = ap.Api("bot-token-123456", base=self.m.base, sleep=lambda s: None)
        with open(SPEC_PATH) as f:
            self.spec = json.load(f)

    def run_plan(self, **kw):
        lines = []
        acts = ap.plan_and_apply(self.api, GUILD, self.spec, out=lines.append, **kw)
        return acts, lines

    def test_spec_valid(self):
        names = [c["name"] for cat in self.spec["categories"] for c in cat["channels"]]
        self.assertEqual(names, ["welcome", "rules", "announcements", "progress", "dev", "d2d-upstream", "help",
                                 "bug-reports", "general", "show-your-runs", "stream"])
        for cat in self.spec["categories"]:
            for c in cat["channels"]:
                self.assertLessEqual(len(c["topic"]), 100, c["name"])
        for r in self.spec["roles"]:
            ap.perm_bits(r["permissions"])
        self.assertEqual([r["name"] for r in self.spec["roles"]], ["owner", "contributor", "member"])

    def test_dry_run_changes_nothing(self):
        acts, _ = self.run_plan()
        self.assertEqual(len(acts), 3 + 5 + 11)
        self.assertEqual(self.m.writes(), [])

    def test_apply_creates_then_idempotent(self):
        acts, _ = self.run_plan(apply=True)
        self.assertEqual(len(self.m.writes()), len(acts))
        prog = next(c for c in self.m.channels if c["name"] == "progress")
        self.assertEqual(prog["permission_overwrites"][0]["id"], GUILD)
        self.assertEqual(int(prog["permission_overwrites"][0]["deny"]) & (1 << 11), 1 << 11)
        cat = next(c for c in self.m.channels if c["name"] == "PROJECT")
        self.assertEqual(prog["parent_id"], cat["id"])
        owner = next(r for r in self.m.roles if r["name"] == "owner")
        self.assertEqual(int(owner["permissions"]), 1 << 3)
        before = len(self.m.writes())
        acts2, lines = self.run_plan(apply=True)
        self.assertEqual(acts2, [])
        self.assertEqual(len(self.m.writes()), before)
        self.assertIn("Nothing to do", lines[-1])

    def test_only_missing_created_never_deletes(self):
        self.m.roles.append({"id": "7", "name": "member", "permissions": "0"})
        self.m.channels.append({"id": "8", "name": "general", "type": 0})
        acts, _ = self.run_plan(apply=True)
        self.assertFalse(any("role member" in a or "#general" in a for a in acts))
        self.assertFalse([w for w in self.m.writes() if w[0] in ("DELETE", "PATCH")])

    def test_settings_only_with_flag(self):
        self.run_plan(apply=True)
        self.assertEqual([w for w in self.m.writes() if w[0] == "PATCH"], [])
        self.run_plan(apply=True, settings=True)
        patches = [w for w in self.m.writes() if w[0] == "PATCH"]
        self.assertEqual(len(patches), 2)
        self.assertEqual(patches[0][2]["verification_level"], 2)

    def test_429_retry(self):
        calls = {"n": 0}
        real = self.m.handle

        def flaky(m, path, body):
            calls["n"] += 1
            return ({"retry_after": 0.5}, 429) if calls["n"] == 1 else real(m, path, body)
        self.m.handle = flaky
        self.assertTrue(self.api.call("GET", "/guilds/%s/roles" % GUILD))

    def test_cli_dry_run_default(self):
        out, err = io.StringIO(), io.StringIO()
        with redirect_stdout(out), redirect_stderr(err):
            code = ap.main(["--guild", GUILD], api=self.api)
        self.assertEqual(code, 0)
        self.assertIn("Dry run", out.getvalue())
        self.assertEqual(self.m.writes(), [])

    def test_cli_bad_guild_and_error_redacted(self):
        err = io.StringIO()
        with redirect_stderr(err):
            self.assertEqual(ap.main(["--guild", "abc"]), 2)
        self.m.close()
        err = io.StringIO()
        with redirect_stderr(err), redirect_stdout(io.StringIO()):
            self.assertEqual(ap.main(["--guild", GUILD], api=self.api), 1)
        self.assertNotIn("bot-token-123456", err.getvalue())


if __name__ == "__main__":
    unittest.main()
