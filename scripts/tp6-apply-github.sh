#!/usr/bin/env bash
set -euo pipefail
REPO="${REPO:-mateopappa/ingsoft3-tp01}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/tp6-wire.out.env"
# shellcheck disable=SC1091
source "$ROOT/scripts/tp6-wire.secrets.env"

if ! gh api user --jq .login | grep -qx mateopappa; then
  echo "Necesitás gh auth login como mateopappa (ahora: $(gh api user --jq .login 2>/dev/null || echo none))"
  exit 1
fi

gh variable set QA_API_URL --repo "$REPO" --body "$QA_API_URL"
gh variable set QA_FRONT_URL --repo "$REPO" --body "$QA_FRONT_URL"
gh variable set PROD_API_URL --repo "$REPO" --body "$PROD_API_URL"
gh variable set PROD_FRONT_URL --repo "$REPO" --body "$PROD_FRONT_URL"

printf '%s' "$RENDER_HOOK_API_QA" | gh secret set RENDER_HOOK_API_QA --env qa --repo "$REPO"
printf '%s' "$RENDER_HOOK_FRONT_QA" | gh secret set RENDER_HOOK_FRONT_QA --env qa --repo "$REPO"
printf '%s' "$RENDER_HOOK_API_PROD" | gh secret set RENDER_HOOK_API_PROD --env production --repo "$REPO"
printf '%s' "$RENDER_HOOK_FRONT_PROD" | gh secret set RENDER_HOOK_FRONT_PROD --env production --repo "$REPO"

echo "OK: vars + 4 secrets. Dispará CD con empty commit a main o re-run del último workflow."
