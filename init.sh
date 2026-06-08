#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DATA_DIR="$BACKEND_DIR/data/init-smoke"
BACKEND_SMOKE_DB="$BACKEND_DATA_DIR/podcast.db"
BACKEND_SMOKE_STORAGE="$BACKEND_DATA_DIR/storage"
BACKEND_SMOKE_LOG="$BACKEND_DATA_DIR/server.log"
GO_CACHE_DIR="$BACKEND_DIR/.gocache"

server_pid=""

cleanup() {
	if [ -n "$server_pid" ] && kill -0 "$server_pid" 2>/dev/null; then
		kill "$server_pid" 2>/dev/null || true
		wait "$server_pid" 2>/dev/null || true
	fi
}
trap cleanup EXIT

choose_port() {
	python3 - <<'PY'
import socket

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
    sock.bind(("127.0.0.1", 0))
    print(sock.getsockname()[1])
PY
}

wait_for_json() {
	local url="$1"
	local expected_key="$2"
	local expected_value="$3"
	python3 - "$url" "$expected_key" "$expected_value" <<'PY'
import json
import sys
import time
import urllib.request

url, expected_key, expected_value = sys.argv[1:4]
deadline = time.time() + 20
last_error = None

while time.time() < deadline:
    try:
        with urllib.request.urlopen(url, timeout=1) as response:
            body = response.read().decode("utf-8")
        data = json.loads(body)
        if str(data.get(expected_key)) == expected_value:
            sys.exit(0)
        last_error = f"unexpected response from {url}: {body}"
    except Exception as exc:
        last_error = str(exc)
    time.sleep(0.25)

print(f"timed out waiting for {url}: {last_error}", file=sys.stderr)
sys.exit(1)
PY
}

verify_json_array_field() {
	local url="$1"
	local field="$2"
	python3 - "$url" "$field" <<'PY'
import json
import sys
import urllib.request

url, field = sys.argv[1:3]
with urllib.request.urlopen(url, timeout=5) as response:
    data = json.loads(response.read().decode("utf-8"))

if not isinstance(data.get(field), list):
    print(f"expected JSON field {field!r} to be an array from {url}: {data}", file=sys.stderr)
    sys.exit(1)
PY
}

echo "== Harness verification =="
"$ROOT_DIR/.agent-harness/scripts/init.sh" "$@"

if [[ "${HARNESS_SKIP_TEST_LAYERS:-}" == "1" ]]; then
	echo "skip project recovery layers: HARNESS_SKIP_TEST_LAYERS=1"
	echo "init verification passed"
	exit 0
fi

echo "== Backend dependencies and tests =="
mkdir -p "$GO_CACHE_DIR"
(
	cd "$BACKEND_DIR"
	GOCACHE="$GO_CACHE_DIR" go test ./...
)

echo "== Frontend dependencies =="
(
	cd "$FRONTEND_DIR"
	if [ ! -d node_modules ]; then
		npm ci
	else
		echo "node_modules present; npm scripts will verify installed packages"
	fi
)

echo "== Frontend tests and build =="
(
	cd "$FRONTEND_DIR"
	npm test
	npm run build
)

echo "== Local backend smoke =="
rm -rf "$BACKEND_DATA_DIR"
mkdir -p "$BACKEND_SMOKE_STORAGE"
smoke_port="$(choose_port)"
(
	cd "$BACKEND_DIR"
	env \
		PORT="$smoke_port" \
		SQLITE_PATH="$BACKEND_SMOKE_DB" \
		LOCAL_STORAGE_PATH="$BACKEND_SMOKE_STORAGE" \
		OPENAI_API_KEY= \
		TRANSCRIBE_URL= \
		SUMMARIZE_URL= \
		GOCACHE="$GO_CACHE_DIR" \
		go run ./cmd/server
) >"$BACKEND_SMOKE_LOG" 2>&1 &
server_pid="$!"

wait_for_json "http://127.0.0.1:$smoke_port/health" "status" "ok"
verify_json_array_field "http://127.0.0.1:$smoke_port/api/podcasts" "items"

echo "init verification passed"
