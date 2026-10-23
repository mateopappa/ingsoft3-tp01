#!/bin/sh
# TP5 quality gate for Go: measure statement coverage on business-logic files
# (*/service.go) and fail the build when below COVERAGE_THRESHOLD.
#
# Go's `go test -cover` measures statements only — there is no native branch
# coverage. That limitation is documented in decisiones.md.

set -eu

OUT_DIR="${OUT_DIR:-/out}"
THRESHOLD="${COVERAGE_THRESHOLD:-70}"
COVERPKG="${COVERPKG:-flow/internal/activities,flow/internal/categories,flow/internal/timer,flow/internal/dashboard,flow/internal/auth}"
TEST_PKGS="${TEST_PKGS:-./internal/activities/ ./internal/categories/ ./internal/timer/ ./internal/dashboard/ ./internal/auth/}"

mkdir -p "$OUT_DIR"

echo "==> Running unit tests with coverage (coverpkg=$COVERPKG)"
# shellcheck disable=SC2086
go test -count=1 -coverprofile="$OUT_DIR/coverage.raw.out" -covermode=count -coverpkg="$COVERPKG" $TEST_PKGS

echo "==> Filtering profile to business logic (*service.go) — excluding handlers, repos, models, main"
# Keep the mode line + any entries whose path ends with service.go
awk '
  NR==1 { print; next }
  $0 ~ /service\.go:/ { print }
' "$OUT_DIR/coverage.raw.out" > "$OUT_DIR/coverage.out"

LINES=$(wc -l < "$OUT_DIR/coverage.out" | tr -d ' ')
if [ "$LINES" -lt 2 ]; then
  echo "FAIL: filtered coverage profile is empty — nothing measured"
  exit 1
fi

go tool cover -func="$OUT_DIR/coverage.out" | tee "$OUT_DIR/coverage.txt"
go tool cover -html="$OUT_DIR/coverage.out" -o "$OUT_DIR/coverage.html"

TOTAL=$(go tool cover -func="$OUT_DIR/coverage.out" | awk '/total:/ { gsub(/%/, "", $3); print $3 }')

# Markdown summary for GitHub Actions step summary (written by the workflow too)
{
  echo "## Backend coverage (Go statement coverage)"
  echo ""
  echo "| Metric | Value |"
  echo "|---|---|"
  echo "| Statement coverage (service.go) | ${TOTAL}% |"
  echo "| Branch coverage | N/A — \`go test -cover\` measures statements only |"
  echo "| Threshold | ${THRESHOLD}% (statements) |"
  echo "| Scope | \`*service.go\` in activities, categories, timer, dashboard, auth |"
  echo ""
  echo "### Per-function"
  echo '```'
  go tool cover -func="$OUT_DIR/coverage.out"
  echo '```'
} > "$OUT_DIR/SummaryGithub.md"

echo ""
echo "Statement coverage (total): ${TOTAL}% — threshold: ${THRESHOLD}%"
echo "Branch coverage: N/A (Go cover measures statements only)"

awk -v t="$TOTAL" -v m="$THRESHOLD" 'BEGIN {
  if ((t + 0) < (m + 0)) {
    printf "FAIL: coverage %.1f%% is below threshold %s%%\n", t, m
    exit 1
  }
  printf "OK: coverage %.1f%% meets threshold %s%%\n", t, m
  exit 0
}'
