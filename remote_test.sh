#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${GOCACHE:-/tmp/magicorm-gocache}"
export GOFLAGS="${GOFLAGS:--mod=mod}"

REMOTE_TEST_PATTERN='^(TestRemote.*|TestReferenceRemote|TestSimpleRemote|TestComposeRemote|TestConstraintRemote)$'

go test -run "$REMOTE_TEST_PATTERN" -count 1 ./test "$@"
