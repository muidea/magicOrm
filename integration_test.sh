#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${GOCACHE:-/tmp/magicorm-gocache}"
export GOFLAGS="${GOFLAGS:--mod=mod}"

go test -count 1 ./test "$@"
