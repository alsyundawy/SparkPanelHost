#!/usr/bin/env python3
"""
Auto-updater untuk manifes App Installer AetherPanel v7.0.
Mengambil versi terbaru dari sumber upstream resmi dengan host allowlist,
menghitung digest SHA-256 secara langsung dari arsip resmi, dan menulis file YAML secara atomik.
"""

import hashlib
import json
import os
import re
import ssl
import sys
import tempfile
import urllib.parse
import urllib.request
from datetime import UTC, datetime
from pathlib import Path

import yaml

USER_AGENT = (
    "AetherPanel-Manifest-Updater/7.0 (+https://github.com/aetherpanel/aetherpanel)"
)
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN", "")
RECIPES_DIR = Path(__file__).resolve().parent.parent / "recipes"

ALLOWED_HOSTS = {
    "api.wordpress.org",
    "wordpress.org",
    "downloads.wordpress.org",
    "api.github.com",
    "github.com",
    "codeload.github.com",
    "objects.githubusercontent.com",
    "release-assets.githubusercontent.com",
}

SSL_CTX = ssl.create_default_context()


def validate_url(url: str) -> None:
    parsed = urllib.parse.urlparse(url)
    if parsed.scheme != "https" or parsed.hostname not in ALLOWED_HOSTS:
        raise ValueError(f"Blocked untrusted upstream URL: {url}")


def http_get_json(url: str) -> dict:
    validate_url(url)
    headers = {"User-Agent": USER_AGENT, "Accept": "application/json"}
    parsed_url = urllib.parse.urlparse(url)
    if GITHUB_TOKEN and parsed_url.hostname == "api.github.com":
        headers["Authorization"] = f"Bearer {GITHUB_TOKEN}"
    req = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(req, timeout=30, context=SSL_CTX) as resp:  # nosec B310
        return json.loads(resp.read().decode("utf-8"))


def compute_remote_sha256(url: str) -> str:
    validate_url(url)
    headers = {"User-Agent": USER_AGENT}
    req = urllib.request.Request(url, headers=headers)
    hasher = hashlib.sha256()
    with urllib.request.urlopen(
        req, timeout=120, context=SSL_CTX
    ) as resp:  # nosec B310
        validate_url(resp.geturl())
        while True:
            chunk = resp.read(1024 * 1024)
            if not chunk:
                break
            hasher.update(chunk)
    return hasher.hexdigest()


def fetch_wordpress_latest() -> tuple[str, str]:
    data = http_get_json("https://api.wordpress.org/core/version-check/1.7/")
    offer = data["offers"][0]
    version = offer["current"]
    download_url = f"https://wordpress.org/wordpress-{version}.tar.gz"
    return version, download_url


def fetch_github_release(repo: str, asset_regex: str | None = None) -> tuple[str, str]:
    data = http_get_json(f"https://api.github.com/repos/{repo}/releases/latest")
    tag = data["tag_name"]
    version = re.sub(r"^(v|V|release-)", "", tag)

    if asset_regex:
        pattern = re.compile(asset_regex)
        for asset in data.get("assets", []):
            if pattern.search(asset["name"]):
                return version, asset["browser_download_url"]
        raise RuntimeError(f"Asset matching '{asset_regex}' not found in {repo}@{tag}")

    return version, data["tarball_url"]


def resolve_upstream(upstream: dict) -> tuple[str, str]:
    provider = upstream.get("provider")
    if provider == "wordpress_org":
        return fetch_wordpress_latest()
    if provider == "github_release":
        return fetch_github_release(
            repo=upstream["repo"],
            asset_regex=upstream.get("asset_regex"),
        )
    raise ValueError(f"Unsupported upstream provider: {provider}")


def atomic_write_yaml(target_path: Path, data: dict) -> None:
    fd, tmp_name = tempfile.mkstemp(dir=target_path.parent, suffix=".tmp")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as tmp_file:
            yaml.safe_dump(data, tmp_file, sort_keys=False, allow_unicode=True)
        os.replace(tmp_name, target_path)
    except Exception:
        if os.path.exists(tmp_name):
            os.unlink(tmp_name)
        raise


def process_recipe(yaml_path: Path) -> bool:
    with yaml_path.open("r", encoding="utf-8") as f:
        recipe = yaml.safe_load(f)

    if not recipe or "upstream" not in recipe:
        return False

    app_id = recipe.get("id", yaml_path.stem)
    current_version = str(recipe.get("version", ""))

    try:
        latest_version, download_url = resolve_upstream(recipe["upstream"])
    except Exception as exc:  # noqa: BLE001
        print(f"[WARN] Gagal memeriksa upstream untuk {app_id}: {exc}", file=sys.stderr)
        return False

    if latest_version == current_version:
        print(f"[OK]   {app_id} sudah versi terbaru ({current_version})")
        return False

    print(f"[INFO] Memperbarui {app_id}: {current_version} -> {latest_version}")
    sha256_digest = compute_remote_sha256(download_url)

    recipe["version"] = latest_version
    recipe["release_date"] = datetime.now(UTC).strftime("%Y-%m-%d")
    recipe["source"]["url"] = download_url
    recipe["source"]["sha256"] = sha256_digest

    atomic_write_yaml(yaml_path, recipe)
    print(
        f"[DONE] {app_id} diperbarui ke {latest_version} (SHA256: {sha256_digest[:12]}...)"
    )
    return True


def main() -> None:
    updated_count = 0
    for yaml_file in sorted(RECIPES_DIR.rglob("*.yaml")):
        if process_recipe(yaml_file):
            updated_count += 1

    print(f"\nTotal manifes diperbarui: {updated_count}")


if __name__ == "__main__":
    main()
