#!/usr/bin/env bash
set -euo pipefail

# Code coverage generation

COVERAGE_DIR="${COVERAGE_DIR:-coverage}"
COVERAGE_SCOPE="${COVERAGE_SCOPE:-unit}"
export GOCACHE="${GOCACHE:-/tmp/magicorm-gocache}"
export GOFLAGS="${GOFLAGS:--mod=mod}"

if [[ "$COVERAGE_SCOPE" == "all" ]]; then
    PKG_LIST=$(go list ./... | rg -v '/demo$|/example$')
else
    # 默认统计可在当前环境稳定执行的核心单元包，排除数据库集成测试包。
    PKG_LIST=$(go list ./... | rg -v '/test$|/demo$|/example$')
fi

mkdir -p "$COVERAGE_DIR"

COVERAGE_FILE="${COVERAGE_DIR}/coverage.out"
COVERPKG=$(echo "$PKG_LIST" | paste -sd, -)

go test \
    -covermode=count \
    -coverpkg="$COVERPKG" \
    -coverprofile="$COVERAGE_FILE" \
    $PKG_LIST

go tool cover -func="$COVERAGE_FILE"

if [[ "${1:-}" == "html" ]]; then
    go tool cover -html="$COVERAGE_FILE" -o coverage.html
fi
