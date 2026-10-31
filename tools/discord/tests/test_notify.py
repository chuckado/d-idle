import http.server
import io
import json
import os
import sys
import tempfile
import threading
import unittest
from contextlib import redirect_stderr, redirect_stdout

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import common as dc
import notify

# No test may read the real ~/.local/share/d2idle: point the data dir at an empty temp dir for this process.
_HOME = tempfile.TemporaryDirectory()
os.environ[dc.HOME_ENV] = _HOME.name

TOKEN = "tok" + "En-Abc123xyz789"


class Mock:
    def __init__(self, script=None):
        self.script = list(script or [])
        self.requests = []
        outer = self

        class H(http.server.BaseHTTPRequestHandler):
            def log_message(self, *a):
                pass

            def do_POST(self):
                body = json.loads(self.rfile.read(int(self.headers.get("Content-Length") or 0)))
                outer.requests.append((self.path, body))
                code, obj = outer.script.pop(0) if outer.script else (204, None)
                raw = json.dumps(obj).encode() if obj is not None else b""
                self.send_response(code)
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)

        self.srv = http.server.HTTPServer(("127.0.0.1", 0), H)
        self.base = "http://127.0.0.1:%d/api/webhooks/" % self.srv.server_port
        self.url = self.base + "123456789/" + TOKEN
        threading.Thread(target=self.srv.serve_forever, daemon=True).start()

    def close(self):
        self.srv.shutdown()
        self.srv.server_close()


class PostTests(unittest.TestCase):
    def setUp(self):
        self.m = Mock()
        self.addCleanup(self.m.close)
        self.sleeps = []
        self.env = {notify.env_name("progress"): self.m.url}

    def post(self, **kw):
        kw.setdefault("title", "Release")
        kw.setdefault("text", "Soak 2 h passed, 0 deaths.")
        return notify.post("progress", env=self.env, store=dc.MemoryStore(), sleep=self.sleeps.append,
                           base_ok=(self.m.base,), **kw)

    def test_ok_payload(self):
        r = self.post(url="https://github.com/example/repo/releases/tag/v1")
        self.assertTrue(r.ok, r.error)
        path, body = self.m.requests[0]
        self.assertTrue(path.endswith(TOKEN))
        self.assertEqual(body["allowed_mentions"], {"parse": []})
        self.assertEqual(body["embeds"][0]["title"], "Release")

    def test_truncation(self):
        p = notify.build_payload("t" * 999, "d" * 9999)
        e = p["embeds"][0]
        self.assertEqual(len(e["title"]), 256)
        self.assertEqual(len(e["description"]), 4096)
        self.assertTrue(e["description"].endswith("..."))

    def test_429_honoured_then_ok(self):
        self.m.script = [(429, {"retry_after": 2.5}), (429, {"retry_after": 99}), (204, None)]
        r = self.post()
        self.assertTrue(r.ok)
        self.assertEqual(r.attempts, 3)
        self.assertEqual(self.sleeps, [2.5, notify.MAX_RETRY_AFTER])

    def test_429_gives_up_after_three(self):
        self.m.script = [(429, {"retry_after": 0.1})] * 5
        r = self.post()
        self.assertFalse(r.ok)
        self.assertEqual(len(self.m.requests), 3)
        self.assertNotIn(TOKEN, r.error)

    def test_client_error_no_retry_no_leak(self):
        self.m.script = [(404, {"message": TOKEN})]
        r = self.post()
        self.assertFalse(r.ok)
        self.assertEqual(len(self.m.requests), 1)
        self.assertNotIn(TOKEN, r.error)

    def test_server_error_retries(self):
        self.m.script = [(502, None), (204, None)]
        self.assertTrue(self.post().ok)

    def test_network_error_never_raises(self):
        self.m.close()
        r = self.post()
        self.assertFalse(r.ok)
        self.assertNotIn(TOKEN, r.error)

    def test_missing_webhook(self):
        r = notify.post("progress", "t", "x", env={}, store=dc.MemoryStore())
        self.assertFalse(r.ok)
        self.assertIn("init", r.error)

    def test_keyring_used(self):
        store = dc.MemoryStore({notify.account_for("progress"): self.m.url})
        r = notify.post("progress", "t", "x", env={}, store=store, base_ok=(self.m.base,), sleep=self.sleeps.append)
        self.assertTrue(r.ok, r.error)

    def test_non_discord_url_refused(self):
        r = notify.post("progress", "t", "x", env=self.env, store=dc.MemoryStore())
        self.assertFalse(r.ok)
        self.assertEqual(self.m.requests, [])

    def test_bad_channel_and_empty(self):
        self.assertTrue(notify.post("../x", "t", "x", env=self.env).refused)
        self.assertTrue(self.post(title="", text="").refused)


def j(*p):
    return "".join(p)


class SanitizerTests(unittest.TestCase):
    BAD = {
        "home path": "saved in /home/" + "someone/x",
        "tilde path": "see ~/code/x",
        "windows path": "C:\\Users\\x",
        "lab name": "built on " + j("or", "ion"),
        "lab name 2": j("her", "dr") + " crashed",
        "person": j("jes", "se") + " said",
        "lan host": "http-check host.lan.example",
        "private domain": "x." + j("birb", "hub") + ".com",
        "ip": "connect " + j("10", ".0.", "0.5"),
        "ip2": j("192.", "168.1.4"),
        "token": "key=" + "a" * 12,
        "ghp": j("gh", "p_") + "A" * 24,
        "long token": "x" * 10 + "Ab3" * 12,
        "webhook": "https://discord.com/api/webhooks/1/abc",
        "bearer": "bearer " + "x" * 12,
        "email": "a" + "@" + "b.example",
        "save": "char.d2s",
        "instance": "ran on i3",
        "bad url": "see http://evil.example/x",
        "url secret": "https://github.com/x?token=abc",
    }

    def test_bad_refused(self):
        for name, text in self.BAD.items():
            self.assertTrue(dc.scan(text, extra_deny=False), name)
            r = notify.post("progress", "t", text, env={notify.env_name("progress"): "x"}, store=dc.MemoryStore())
            self.assertTrue(r.refused, name)

    def test_title_and_url_checked(self):
        self.assertTrue(notify.post("progress", j("or", "ion"), "x", env={}).refused)
        self.assertTrue(notify.post("progress", "t", "x", url="http://github.com/x", env={}).refused)

    def test_good_passes(self):
        for t in ("Soak run: 2 h, 0 deaths, 41 xp/min (verified)", "Release v0.3: see docs/quickstart.md",
                  "https://github.com/example/repo/pull/12 merged", "Act 1 Normal cleared",
                  "the Forgotten " + dc._TW.title() + " is a game area"):
            self.assertEqual(dc.scan(t, extra_deny=False), [], t)

    def test_extra_deny_file(self):
        p = os.path.join(os.environ.get("TMPDIR", "/tmp"), "dcdeny%d.txt" % os.getpid())
        with open(p, "w") as f:
            f.write("# c\nsecretproject\n")
        self.addCleanup(os.unlink, p)
        os.environ["D2IDLE_DISCORD_DENY"] = p
        self.addCleanup(os.environ.pop, "D2IDLE_DISCORD_DENY")
        self.assertEqual(dc.scan("about SecretProject"), ["deny-list"])


class CliTests(unittest.TestCase):
    def run_cli(self, *argv):
        out, err = io.StringIO(), io.StringIO()
        with redirect_stdout(out), redirect_stderr(err):
            code = notify.main(list(argv))
        return code, out.getvalue(), err.getvalue()

    def test_dry_run_prints_payload(self):
        code, out, _ = self.run_cli("post", "--channel", "progress", "--title", "T", "--text", "hello", "--dry-run")
        self.assertEqual(code, 0)
        self.assertEqual(json.loads(out)["embeds"][0]["description"], "hello")

    def test_dry_run_refuses(self):
        code, out, err = self.run_cli("post", "--channel", "progress", "--text", "at /home/" + "x/y", "--dry-run")
        self.assertEqual(code, 2)
        self.assertEqual(out, "")

    def test_init(self):
        code, out, _ = self.run_cli("init", "--channel", "progress")
        self.assertEqual(code, 0)
        self.assertIn("secret-tool store", out)
        self.assertIn("D2IDLE_DISCORD_WEBHOOK_PROGRESS", out)
        self.assertIn('--label="d.idle Discord progress webhook"', out)


if __name__ == "__main__":
    unittest.main()
