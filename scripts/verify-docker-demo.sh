#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

KEEP_RUNNING=0
if [[ "${1:-}" == "--keep-running" ]]; then
	KEEP_RUNNING=1
fi

if ! command -v docker >/dev/null 2>&1; then
	echo "docker is required to verify the Docker demo path" >&2
	exit 127
fi

if ! docker compose version >/dev/null 2>&1; then
	echo "docker compose is required to verify the Docker demo path" >&2
	exit 127
fi

cleanup() {
	if [[ "$KEEP_RUNNING" != "1" ]]; then
		docker compose down --remove-orphans >/dev/null 2>&1 || true
	fi
}
trap cleanup EXIT

mkdir -p backend/data

docker compose up --build -d

python3 - <<'PY'
import json
import sys
import time
import urllib.request

checks = [
    ("backend health", "http://127.0.0.1:8080/health", lambda body: json.loads(body).get("status") == "ok"),
    ("backend podcast list", "http://127.0.0.1:8080/api/podcasts", lambda body: isinstance(json.loads(body).get("items"), list)),
    ("frontend health", "http://127.0.0.1:8081/health", lambda body: body.strip() == "ok"),
]

for label, url, predicate in checks:
    deadline = time.time() + 60
    last_error = None
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=2) as response:
                body = response.read().decode("utf-8")
            if predicate(body):
                print(f"verified {label}: {url}")
                break
            last_error = f"unexpected body: {body}"
        except Exception as exc:
            last_error = str(exc)
        time.sleep(1)
    else:
        print(f"timed out verifying {label} at {url}: {last_error}", file=sys.stderr)
        sys.exit(1)
PY

docker compose ps

if [[ "$KEEP_RUNNING" == "1" ]]; then
	echo "Docker demo is running:"
	echo "- backend: http://localhost:8080"
	echo "- frontend: http://localhost:8081"
else
	echo "Docker demo verification passed"
fi
