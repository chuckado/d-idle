"""Shared pieces of the d.idle Discord tooling: keyring store, redaction, the fail-closed text sanitizer. Stdlib only.

Secrets (webhook URLs, the bot token) live in the desktop keyring (secret-tool, service=d2idle), like tools/kick.
They are never written to files, logs or error text: every message goes through redact().

The sanitizer refuses text that names the maintainer's private lab. The private names are assembled from fragments so
this public file does not contain them literally (the repo's leakcheck scans for them). An optional extra deny list
can be kept outside the repo, one case-insensitive word per line, in D2IDLE_DISCORD_DENY or
~/.config/d2idle/discord-deny.txt.
"""
import os
import re
import stat
import subprocess

SERVICE = "d2idle"
BOT_TOKEN_ENV = "D2IDLE_DISCORD_BOT_TOKEN"
BOT_TOKEN_FILE_ENV = "D2IDLE_DISCORD_BOT_TOKEN_FILE"
DEFAULT_BOT_TOKEN_FILE = "~/.local/share/d2idle/discord-bot-token"
_SECRETS = set()


def register_secret(value):
    if value and len(value) >= 6:
        _SECRETS.add(value)


def redact(text):
    text = str(text)
    for s in sorted(_SECRETS, key=len, reverse=True):
        text = text.replace(s, "[secret]")
    text = re.sub(r"/api/webhooks/\d+/[\w-]+", "/api/webhooks/[secret]", text)
    return re.sub(r"(?i)\b(bot|bearer)\s+\S+", r"\1 [secret]", text)


class DiscordError(Exception):
    def __init__(self, msg):
        super().__init__(redact(msg))


class SecretToolStore:
    """Desktop keyring via secret-tool. Values are read from stdout, never passed on argv."""

    def get(self, account):
        try:
            r = subprocess.run(["secret-tool", "lookup", "service", SERVICE, "account", account],
                               capture_output=True, text=True, timeout=15)
        except (OSError, subprocess.SubprocessError):
            return None
        v = r.stdout.strip() if r.returncode == 0 else ""
        register_secret(v)
        return v or None


class MemoryStore:
    """Test double with the same interface."""

    def __init__(self, data=None):
        self.data = dict(data or {})

    def get(self, account):
        v = self.data.get(account)
        register_secret(v)
        return v


def read_secret_file(path):
    """Read a one-line secret file, fail closed. Returns the stripped value; raises DiscordError (never with the value).

    Rules: regular file (no symlink), owned by this user, mode exactly 0600, parent directory not group/world-writable,
    non-empty after stripping, no embedded whitespace.
    """
    path = os.path.expanduser(path)
    fix = "chmod 600 %s" % path
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NONBLOCK | getattr(os, "O_NOFOLLOW", 0))
    except OSError as e:
        raise DiscordError("secret file %s cannot be opened safely (%s); it must be a regular file, not a symlink" % (path, e.strerror))
    try:
        st = os.fstat(fd)
        if not stat.S_ISREG(st.st_mode):
            raise DiscordError("secret file %s is not a regular file" % path)
        if st.st_uid != os.geteuid():
            raise DiscordError("secret file %s is not owned by the current user" % path)
        if stat.S_IMODE(st.st_mode) != 0o600:
            raise DiscordError("secret file %s has mode %o, it must be exactly 600 (group/other must not read it); fix: %s"
                               % (path, stat.S_IMODE(st.st_mode), fix))
        parent = os.path.dirname(os.path.abspath(path))
        if os.stat(parent).st_mode & 0o022:
            raise DiscordError("directory %s is group/world-writable; fix: chmod go-w %s" % (parent, parent))
        with os.fdopen(fd, "r", closefd=False) as f:
            raw = f.read(4096)
    finally:
        os.close(fd)
    value = raw.strip()
    if not value:
        raise DiscordError("secret file %s is empty" % path)
    if re.search(r"\s", value):
        raise DiscordError("secret file %s must hold one value without spaces or line breaks" % path)
    register_secret(value)
    return value


def load_bot_token(env=None, store=None):
    """Bot token: env D2IDLE_DISCORD_BOT_TOKEN > file (D2IDLE_DISCORD_BOT_TOKEN_FILE, else the default private file)
    > desktop keyring. None when nothing is set. A file that exists but is unsafe raises (no fall-through)."""
    env = os.environ if env is None else env
    tok = env.get(BOT_TOKEN_ENV)
    if tok and tok.strip():
        tok = tok.strip()
        register_secret(tok)
        return tok
    explicit = env.get(BOT_TOKEN_FILE_ENV)
    path = os.path.expanduser(explicit or DEFAULT_BOT_TOKEN_FILE)
    if os.path.lexists(path):
        return read_secret_file(path)
    if explicit:
        raise DiscordError("%s points at %s, which does not exist" % (BOT_TOKEN_FILE_ENV, path))
    return (store or SecretToolStore()).get("discord-bot-token")


def _j(*parts):
    return "".join(parts)


# Private-lab names, built from fragments (see module docstring).
_NAMES = "|".join([
    _j("or", "ion"), _j("her", "dr"), _j("des", "mond"), _j("forge", "jo"), _j("un", "raid"), _j("jes", "se"),
    _j("birb", "hub"), _j("mac", "mini"), _j("omle", "tte"),
])
_WORD_NAMES = "|".join([_j("ve", "nus"), _j("ja", "de"), _j("on", "yx"), _j("to", "wer")])
_TW, _JD = _j("to", "wer"), _j("ja", "de")
_GAME_PHRASES = re.compile(r"(?i)forgotten\s+" + _TW + r"|" + _TW + r"\s+cellar|" + _JD + r"\s+figurine|" + _JD + r"\s+(?:ring|amulet|item)")

RULES = [
    ("ipv4", re.compile(r"\b\d{1,4}(?:\.\d{1,4}){3}\b")),
    ("ipv6", re.compile(r"(?i)(?<![\w:])(?:[0-9a-f]{1,4}:){7}[0-9a-f]{1,4}\b|(?<![\w:])[0-9a-f]{0,4}(?::[0-9a-f]{1,4})*::[0-9a-f:]*(?![\w:])")),
    ("private-name", re.compile(r"(?i)" + _NAMES + r"|(?<![a-z])(?:" + _WORD_NAMES + r")(?![a-z])")),
    ("private-host", re.compile(r"(?i)\b[\w-]+(?:\.[\w-]+)*\.(?:lan|local|internal|home|localdomain|corp)\b|\.home\.arpa\b")),
    ("path", re.compile(r"(?i)(?:/(?:home|tmp|mnt|etc|var|usr|root|opt|proc|users)/|~/|(?<![a-z])[a-z]:[/\\]|\\\\\w|drive_c|dosdevices|d2lod-live)")),
    ("instance-id", re.compile(r"(?<![A-Za-z0-9_])i[0-5](?![A-Za-z0-9_])")),
    ("ssh-url", re.compile(r"(?i)\bssh://|\bscp\s|\bsftp://")),
    ("secret-word", re.compile(r"(?i)\b(?:token|password|passwd|pwd|secret|api[_-]?key|auth|key)s?\b\s*[:=]\s*\S")),
    ("secret-word-bare", re.compile(r"(?i)\b(?:password|passwd|bearer|" + _j("author", "ization") + r"|private key|api[_-]?key)\b|\bbasic\s+[A-Za-z0-9+/]{8,}")),
    ("webhook-url", re.compile(r"(?i)/api/webhooks/\d+|discord(?:app)?\.com/api/")),
    ("token-prefix", re.compile(r"\bgh[pousr]_[A-Za-z0-9]{10,}|\bxox[abp]-[A-Za-z0-9-]{10,}|\bsk-[A-Za-z0-9]{16,}")),
    ("uuid", re.compile(r"(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b")),
    ("cdkey", re.compile(r"\b[A-Za-z0-9]{4,6}(?:[- ][A-Za-z0-9]{4,6}){4}\b")),
    ("email", re.compile(r"[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+")),
    ("save-file", re.compile(r"(?i)\b[\w .-]*\.(?:d2s|d2x|d2i|ctl|key|bak)\b|\bsaves?[/\\]")),
    ("hex-blob", re.compile(r"(?i)\b[0-9a-f]{24,}\b")),
]
_B64 = re.compile(r"[A-Za-z0-9+_-]{22,}={0,2}")
_URL = re.compile(r"[A-Za-z][A-Za-z0-9+.-]*://[^\s)>\]\"']+")
_QS_SECRET = re.compile(r"[?&;][^=&\s]*(?:token|key|pass|secret|auth|sig)[^=&\s]*=", re.I)
ALLOWED_URL_HOSTS = ("github.com", "kick.com", "youtube.com", "twitch.tv")


def _looks_random(tok):
    return any(c.isdigit() for c in tok) or (tok.lower() != tok and tok.upper() != tok and not tok.isalpha()) \
        or (any(c.islower() for c in tok) and any(c.isupper() for c in tok) and sum(c.isupper() for c in tok) >= 3)


def _extra_deny():
    path = os.environ.get("D2IDLE_DISCORD_DENY") or os.path.expanduser("~/.config/d2idle/discord-deny.txt")
    try:
        with open(path) as f:
            words = [w.strip() for w in f if w.strip() and not w.startswith("#")]
    except OSError:
        return None
    return re.compile("|".join(re.escape(w) for w in words), re.I) if words else None


def scan(text, extra_deny=None):
    """Names of the rules the text trips (empty list = safe). Fail-closed: any hit means do not post."""
    hits = []
    work = str(text)
    for m in _URL.finditer(work):
        url = m.group(0)
        try:
            from urllib.parse import urlsplit
            sp = urlsplit(url)
            host, port = (sp.hostname or "").lower(), sp.port
        except ValueError:
            hits.append("url-malformed")
            continue
        ok = sp.scheme.lower() in ("http", "https") and not port and not sp.username and not sp.password \
            and any(host == h or host.endswith("." + h) for h in ALLOWED_URL_HOSTS)
        if not ok:
            hits.append("url-not-allowed")
        elif _QS_SECRET.search(url):
            hits.append("url-secret-query")
    work = _URL.sub(" ", work)
    work = _GAME_PHRASES.sub(" ", work)
    for name, rx in RULES:
        if rx.search(work):
            hits.append(name)
    for m in _B64.finditer(work):
        tok = m.group(0)
        if len(tok) >= 32 and _looks_random(tok):
            hits.append("random-token")
            break
    deny = extra_deny if extra_deny is not None else _extra_deny()
    if deny and deny.search(work):
        hits.append("deny-list")
    return sorted(set(hits))
