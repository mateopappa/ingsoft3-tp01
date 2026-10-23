#!/bin/sh
# Runs vitest with coverage (thresholds enforced by vitest.config.js)
# and copies reports to OUT_DIR for the CI volume mount.
set -eu

OUT_DIR="${OUT_DIR:-/out}"
export OUT_DIR
mkdir -p "$OUT_DIR"

npm run test:coverage

cp -r coverage "$OUT_DIR/coverage"
cp coverage/coverage-summary.json "$OUT_DIR/coverage-summary.json"

node /app/scripts/write-summary.mjs 2>/dev/null || node "$(dirname "$0")/write-summary.mjs"
