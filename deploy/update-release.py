#!/usr/bin/python3
"""Install a verified release for the existing Linux amd64 systemd service."""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import sqlite3
import subprocess
import tempfile
import time
from urllib.request import Request, urlopen
from urllib.error import HTTPError

API = "https://api.github.com/repos/seyshee/svyaz/releases/latest"
BINARY = Path("/usr/local/bin/svyaz")
STATE = Path("/var/lib/svyaz-updater")
ASSET = "svyaz-linux-amd64"

def healthy():
    for _ in range(15):
        try:
            with urlopen("http://127.0.0.1:8080/api.php?a=me", timeout=2):
                pass
        except HTTPError as error:
            if error.code == 401 and json.load(error).get("setup") is False:
                return True
        except OSError:
            pass
        time.sleep(1)
    return False

def update():
    with urlopen(Request(API, headers={"Accept": "application/vnd.github+json"}), timeout=20) as response:
        release = json.load(response)
    tag = release["tag_name"]
    if not re.fullmatch(r"v\d+\.\d+\.\d+", tag):
        raise ValueError("Expected stable version tag")
    asset = next(a for a in release["assets"] if a["name"] == ASSET)
    digest = asset.get("digest") or ""
    if not re.fullmatch("sha256:[a-f0-9]{64}", digest):
        raise ValueError("Missing GitHub asset digest")
    marker = STATE / "installed"
    if marker.exists() and marker.read_text() == tag + " " + digest:
        return
    url = f"https://github.com/seyshee/svyaz/releases/download/{tag}/{ASSET}"
    with tempfile.TemporaryDirectory(dir=BINARY.parent) as stage:
        candidate = Path(stage) / ASSET
        with urlopen(url, timeout=60) as source, candidate.open("wb") as target:
            shutil.copyfileobj(source, target)
        if hashlib.sha256(candidate.read_bytes()).hexdigest() != digest[7:]:
            raise ValueError("GitHub digest mismatch")
        candidate.chmod(0o755)
        db = Path("/var/lib/svyaz/chat.sqlite")
        backup = Path(tempfile.mkdtemp(prefix="before-" + tag + "-", dir=STATE))
        shutil.copy2(BINARY, backup / "svyaz")
        with sqlite3.connect(db.as_uri() + "?mode=ro", uri=True) as src, sqlite3.connect(backup / "chat.sqlite") as dst:
            src.backup(dst)
            if dst.execute("pragma quick_check").fetchone() != ("ok",):
                raise RuntimeError("SQLite backup failed")
        os.replace(candidate, BINARY)
        try:
            subprocess.run(["systemctl", "restart", "svyaz"], check=True)
            if not healthy():
                raise RuntimeError("Chat health check failed")
        except Exception:
            shutil.copy2(backup / "svyaz", candidate)
            os.replace(candidate, BINARY)
            subprocess.run(["systemctl", "restart", "svyaz"], check=True)
            raise
        marker.write_text(tag + " " + digest)
        print("Installed", tag, flush=True)

if __name__ == "__main__":
    os.umask(0o077)
    update()
