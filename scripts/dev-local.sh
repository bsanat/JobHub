#!/bin/sh
set -eu

mode=${1:-full}
script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(dirname "$script_dir")

if [ "$mode" != "full" ] && [ "$mode" != "backend" ]; then
	echo "Usage: $0 [full|backend]" >&2
	exit 1
fi

for command_name in docker go psql pg_isready lsof; do
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "Required command is missing: $command_name" >&2
		exit 1
	fi
done
if [ "$mode" = "full" ] && ! command -v npm >/dev/null 2>&1; then
	echo "Required command is missing: npm" >&2
	exit 1
fi

env_file_value() {
	key=$1
	if [ ! -f "$repo_dir/.env" ]; then
		return 1
	fi
	line=$(grep -E "^[[:space:]]*$key=" "$repo_dir/.env" | tail -n 1 || true)
	if [ -z "$line" ]; then
		return 1
	fi
	value=${line#*=}
	case "$value" in
		\"*\")
			value=${value#\"}
			value=${value%\"}
			;;
		\'*\')
			value=${value#\'}
			value=${value%\'}
			;;
	esac
	printf '%s\n' "$value"
}

config_value() {
	key=$1
	fallback=$2
	current=$(printenv "$key" 2>/dev/null || true)
	if [ -n "$current" ]; then
		printf '%s\n' "$current"
		return
	fi
	file_value=$(env_file_value "$key" || true)
	if [ -n "$file_value" ]; then
		printf '%s\n' "$file_value"
		return
	fi
	printf '%s\n' "$fallback"
}

database_name=$(config_value JOBHUB_POC_DATABASE jobhub_ats_poc_careers_20260925)
export POSTGRES_USER=$(config_value POSTGRES_USER jobhub)
export POSTGRES_PASSWORD=$(config_value POSTGRES_PASSWORD jobhub)
export POSTGRES_DB=$database_name
export POSTGRES_PORT=$(config_value POSTGRES_PORT 5432)
database_url="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${database_name}?sslmode=disable"

case "$database_name" in
	*[!A-Za-z0-9_]*)
		echo "JOBHUB_POC_DATABASE must contain only letters, numbers, and underscores." >&2
		exit 1
		;;
esac

psql_cmd() {
	PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$@"
}

wait_for_postgres() {
	attempt=0
	while [ "$attempt" -lt 30 ]; do
		if PGPASSWORD=$POSTGRES_PASSWORD pg_isready -h localhost -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres >/dev/null 2>&1; then
			return 0
		fi
		attempt=$((attempt + 1))
		sleep 1
	done
	echo "Local PostgreSQL did not become ready on localhost:$POSTGRES_PORT." >&2
	exit 1
}

ensure_database() {
	exists=$(psql_cmd -d postgres -X -Atqc "SELECT 1 FROM pg_database WHERE datname = '$database_name'" || true)
	if [ "$exists" != "1" ]; then
		echo "Creating local development database: $database_name"
		psql_cmd -d postgres -X -v ON_ERROR_STOP=1 -c "CREATE DATABASE $database_name" >/dev/null
	fi
}

run_migrations() {
	echo "Applying local migrations..."
	(
		cd "$repo_dir"
		docker compose --profile local run --rm migrate
	)
}

seed_demo_data() {
	echo "Seeding demo companies and vacancies..."
	(
		cd "$repo_dir/backend"
		APP_ENV=development \
			DATABASE_URL=$database_url \
			PGX_QUERY_EXEC_MODE=cache_statement \
			go run ./cmd/seed-demo
	)
}

assert_demo_data() {
	job_count=$(psql_cmd -d "$database_name" -X -Atqc "SELECT count(*) FROM jobhub.jobs j WHERE jobhub.job_is_public(j)")
	company_count=$(psql_cmd -d "$database_name" -X -Atqc "SELECT count(*) FROM jobhub.companies c WHERE c.status = 'active'")
	if [ "$job_count" -lt 1 ] || [ "$company_count" -lt 1 ]; then
		echo "Local demo data is incomplete: visible_jobs=$job_count companies=$company_count" >&2
		exit 1
	fi
}

if lsof -nP -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then
	echo "Port 8080 is already in use. Stop the existing backend first." >&2
	exit 1
fi

if [ "$mode" = "full" ] && lsof -nP -iTCP:5173 -sTCP:LISTEN >/dev/null 2>&1; then
	echo "Port 5173 is already in use. Stop the existing frontend first." >&2
	exit 1
fi

(
	cd "$repo_dir"
	docker compose --profile local up -d postgres
)
wait_for_postgres
ensure_database
run_migrations
seed_demo_data

export APP_ENV=development
export DATABASE_URL=$database_url
export FRONTEND_ORIGIN=http://localhost:5173
export VITE_API_URL=http://localhost:8080/api/v1
export PGX_QUERY_EXEC_MODE=cache_statement

assert_demo_data
echo "Starting JobHub with $job_count visible vacancies and $company_count companies from $database_name."
if [ "$mode" = "full" ]; then
	echo "Frontend: http://localhost:5173"
fi
echo "Backend:  http://localhost:8080"

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
		exec npm run dev
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
