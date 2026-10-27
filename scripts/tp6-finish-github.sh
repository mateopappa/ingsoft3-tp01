#!/usr/bin/env bash
# Completa vars + secrets de hooks en GitHub (correr como mateopappa).
# Prereq: scripts/tp6-wire.out.env con las 4 URLs (ya generadas).
# Los Deploy Hooks NO salen por API: Settings → Deploy Hook de cada servicio.
set -euo pipefail
REPO="${REPO:-mateopappa/ingsoft3-tp01}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck disable=SC1091
source "$ROOT/scripts/tp6-wire.out.env"

need() { [[ -n "${!1:-}" ]] || { echo "Falta $1 en tp6-wire.out.env"; exit 1; }; }
need QA_API_URL; need QA_FRONT_URL; need PROD_API_URL; need PROD_FRONT_URL

gh variable set QA_API_URL --repo "$REPO" --body "$QA_API_URL"
gh variable set QA_FRONT_URL --repo "$REPO" --body "$QA_FRONT_URL"
gh variable set PROD_API_URL --repo "$REPO" --body "$PROD_API_URL"
gh variable set PROD_FRONT_URL --repo "$REPO" --body "$PROD_FRONT_URL"
echo "vars OK"

echo
echo "Pegá cada Deploy Hook (Dashboard → Service → Settings → Deploy Hook):"
echo "  api-qa:   https://dashboard.render.com/web/${RENDER_API_QA_ID}"
echo "  front-qa: https://dashboard.render.com/web/${RENDER_FRONT_QA_ID}"
echo "  api-prod: https://dashboard.render.com/web/${RENDER_API_PROD_ID}"
echo "  front-prod: https://dashboard.render.com/web/${RENDER_FRONT_PROD_ID}"
echo

read -r -p "HOOK api-qa: " HOOK_API_QA
read -r -p "HOOK front-qa: " HOOK_FRONT_QA
read -r -p "HOOK api-prod: " HOOK_API_PROD
read -r -p "HOOK front-prod: " HOOK_FRONT_PROD

printf '%s' "$HOOK_API_QA" | gh secret set RENDER_HOOK_API_QA --env qa --repo "$REPO"
printf '%s' "$HOOK_FRONT_QA" | gh secret set RENDER_HOOK_FRONT_QA --env qa --repo "$REPO"
printf '%s' "$HOOK_API_PROD" | gh secret set RENDER_HOOK_API_PROD --env production --repo "$REPO"
printf '%s' "$HOOK_FRONT_PROD" | gh secret set RENDER_HOOK_FRONT_PROD --env production --repo "$REPO"
echo "secrets OK"
echo "Listo. Empty commit a main o re-run Actions para disparar deploy-qa → approve prod."
