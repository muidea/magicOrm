#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${GOCACHE:-/tmp/magicorm-gocache}"
export GOFLAGS="${GOFLAGS:--mod=mod}"

LOCAL_TEST_PATTERN='^(TestLocal.*|TestReferenceLocal|TestSimpleLocal|TestConstraintLocal|TestComposeLocal|TestUpdateRelation.*)$'

go test -run "$LOCAL_TEST_PATTERN" -count 1 ./test "$@"
