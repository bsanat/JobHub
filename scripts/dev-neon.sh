#!/bin/sh
set -eu

mode=${1:-full}
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")

if [ "$mode" != "full" ] && [ "$mode" != "backend" ]; then
	echo "Usage: $0 [full|backend]" >&2
	exit 1
fi

for command_name in go lsof; do
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "Required command is missing: $command_name" >&2
		exit 1
	fi
done
if [ "$mode" = "full" ] && ! command -v npm >/dev/null 2>&1; then
	echo "Required command is missing: npm" >&2
	exit 1
fi
if [ ! -f "$repo_dir/.env" ]; then
	echo "Create .env with DATABASE_URL before running Neon development." >&2
	exit 1
fi

set -a
. "$repo_dir/.env"
set +a

if [ -z "${DATABASE_URL:-}" ]; then
	echo "DATABASE_URL is required in .env." >&2
	exit 1
fi

export APP_ENV=${APP_ENV:-development}
export PORT=${PORT:-8080}
export FRONTEND_ORIGIN=${FRONTEND_ORIGIN:-http://localhost:5173}
export VITE_API_URL=${VITE_API_URL:-http://localhost:${PORT}/api/v1}
export VITE_USE_MOCKS=${VITE_USE_MOCKS:-false}
export TELEGRAM_PUBLISHER_ENABLED=${TELEGRAM_PUBLISHER_ENABLED:-false}

if lsof -nP -iTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
	echo "Port $PORT is already in use. Stop the existing backend first." >&2
	exit 1
fi

if [ "$mode" = "full" ]; then
	frontend_port=${FRONTEND_PORT:-5173}
	if lsof -nP -iTCP:"$frontend_port" -sTCP:LISTEN >/dev/null 2>&1; then
		echo "Port $frontend_port is already in use. Stop the existing frontend first." >&2
		exit 1
	fi
fi

echo "Starting JobHub with Neon database from .env."
if [ "$mode" = "full" ]; then
	echo "Frontend: $FRONTEND_ORIGIN"
fi
echo "Backend:  http://localhost:${PORT}"
echo "Telegram publisher: ${TELEGRAM_PUBLISHER_ENABLED}"

backend_pid=
frontend_pid=

cleanup() {
	trap - INT TERM EXIT
	[ -z "$frontend_pid" ] || kill "$frontend_pid" 2>/dev/null || true
	[ -z "$backend_pid" ] || kill "$backend_pid" 2>/dev/null || true
	wait 2>/dev/null || true
}
trap cleanup INT TERM EXIT

(
	cd "$repo_dir/backend"
	exec go run ./cmd/api
) &
backend_pid=$!

if [ "$mode" = "full" ]; then
	(
		cd "$repo_dir/frontend"
		if [ ! -d node_modules ]; then
			npm ci
		fi
		exec npm run dev -- --host 0.0.0.0
	) &
	frontend_pid=$!
fi

if [ "$mode" = "backend" ]; then
	wait "$backend_pid"
	exit $?
fi

while kill -0 "$backend_pid" 2>/dev/null && kill -0 "$frontend_pid" 2>/dev/null; do
	sleep 1
done

echo "A JobHub development process stopped." >&2
exit 1
