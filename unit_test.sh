#!/usr/bin/env bash
set -euo pipefail
export GOCACHE="${GOCACHE:-/tmp/magicorm-gocache}"
# Integration tests remain visible as skips in the default full package set.
unset MAGICORM_INTEGRATION_ALLOW MAGICORM_DATETIME_TEST_ALLOW MAGICORM_INLINE_REFERENCE_TEST_ALLOW
go test -count=1 ./... "$@"
