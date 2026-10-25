#!/usr/bin/env bash
# Own one disposable instance; cleanup uses its exact ID, never a name prefix.
set -euo pipefail
cd "$(dirname "$0")/.."
fixture_dir=$(mktemp -d)
chmod 700 "$fixture_dir"
fixture_id=''
cleanup() {
  local result=$?
  trap - EXIT
  if [[ -n "$fixture_id" ]]; then
    if ! docker rm -fv "$fixture_id" >/dev/null; then result=1; fi
    if docker inspect "$fixture_id" >/dev/null 2>&1; then result=1; fi
  fi
  rm -rf "$fixture_dir"
  exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
python3 - "$fixture_dir/secret" <<'PY'
import pathlib,secrets,sys
p=pathlib.Path(sys.argv[1]); p.write_text('POSTGRES_PASSWORD='+secrets.token_hex(24)+'\n'); p.chmod(0o600)
PY
fixture_id=$(docker run -d --label magicorm.fixture=capability-closeout --env-file "$fixture_dir/secret" -e POSTGRES_DB=magicorm_test_closeout -e POSTGRES_USER=ormtest -p 127.0.0.1::5432 postgres:17.7-alpine3.22)
for attempt in {1..60}; do
  if docker exec "$fixture_id" psql -h 127.0.0.1 -v ON_ERROR_STOP=1 -U ormtest -d magicorm_test_closeout -c 'SELECT 1' >/dev/null 2>&1; then break; fi
  sleep 0.5
done
docker exec "$fixture_id" psql -v ON_ERROR_STOP=1 -U ormtest -d magicorm_test_closeout -c 'CREATE SCHEMA integration' >/dev/null
fixture_port=$(docker port "$fixture_id" 5432/tcp)
fixture_password=$(cut -d= -f2 "$fixture_dir/secret")
export MAGICORM_INTEGRATION_ALLOW=1
export MAGICORM_POSTGRES_SERVER="$fixture_port" MAGICORM_POSTGRES_DATABASE=magicorm_test_closeout MAGICORM_POSTGRES_SCHEMA=integration MAGICORM_POSTGRES_USER=ormtest MAGICORM_POSTGRES_PASSWORD="$fixture_password"
export MAGICORM_SCHEMA_TEST_POSTGRES_DSN="postgres://ormtest:$fixture_password@$fixture_port/magicorm_test_closeout?sslmode=disable"
export MAGICORM_INLINE_REFERENCE_TEST_ALLOW=isolated MAGICORM_INLINE_REFERENCE_TEST_SERVER="$fixture_port" MAGICORM_INLINE_REFERENCE_TEST_DATABASE=magicorm_test_closeout MAGICORM_INLINE_REFERENCE_TEST_SCHEMA=integration MAGICORM_INLINE_REFERENCE_TEST_USER=ormtest MAGICORM_INLINE_REFERENCE_TEST_PASSWORD="$fixture_password"
export MAGICORM_DATETIME_TEST_ALLOW=isolated MAGICORM_DATETIME_TEST_SERVER="$fixture_port" MAGICORM_DATETIME_TEST_DATABASE=magicorm_test_closeout MAGICORM_DATETIME_TEST_SCHEMA=integration MAGICORM_DATETIME_TEST_USER=ormtest MAGICORM_DATETIME_TEST_PASSWORD="$fixture_password"
printf 'Disposable fixture ID: %s; endpoint: %s\n' "$fixture_id" "$fixture_port"
# A supplied command can validate consumers against this same isolated target.
if (( $# )); then "$@"; else go test ./... -count=1; fi
