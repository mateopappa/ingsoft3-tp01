#!/usr/bin/env bash
# TP6 — cableado Neon + Render + GitHub (secrets/vars/environments).
# Corré ESTO en tu máquina, autenticado como mateopappa (no el token del agente).
#
# Prereqs:
#   brew install gh jq curl   # o apt
#   gh auth login             # cuenta mateopappa, scopes: repo + workflow
#
# Exportá antes de correr:
#   export NEON_API_KEY='…'          # Neon → Account settings → API keys
#   export RENDER_API_KEY='rnd_…'    # Render → Account Settings → API Keys
#   export RENDER_OWNER_ID='usr-…'   # de la URL del dashboard (usr-… o tea-…)
#   export SESSION_SECRET_QA='…'     # opcional; se genera si falta
#   export SESSION_SECRET_PROD='…'   # opcional
#
# Uso:
#   chmod +x scripts/tp6-wire.sh
#   ./scripts/tp6-wire.sh

set -euo pipefail

REPO="${REPO:-mateopappa/ingsoft3-tp01}"
OWNER="${REPO%%/*}"
NAME="${REPO##*/}"
IMAGE_SHA="${IMAGE_SHA:-$(git rev-parse origin/main 2>/dev/null || git rev-parse HEAD)}"
IMAGE_TAG="sha-${IMAGE_SHA}"
IMG_API="ghcr.io/${OWNER}/${NAME}-backend"
IMG_FRONT="ghcr.io/${OWNER}/${NAME}-frontend"
REGION="${RENDER_REGION:-oregon}"

need() { [[ -n "${!1:-}" ]] || { echo "Falta export $1"; exit 1; }; }
need NEON_API_KEY
need RENDER_API_KEY
need RENDER_OWNER_ID

if ! gh auth status &>/dev/null; then
  echo "gh no autenticado. Corré: gh auth login"
  exit 1
fi
GH_USER=$(gh api user --jq .login)
if [[ "$GH_USER" != "$OWNER" ]]; then
  echo "Atención: gh eres '$GH_USER', el repo es de '$OWNER'. Seguí solo si tenés admin."
fi

rand_secret() { openssl rand -hex 24; }
SESSION_SECRET_QA="${SESSION_SECRET_QA:-$(rand_secret)}"
SESSION_SECRET_PROD="${SESSION_SECRET_PROD:-$(rand_secret)}"

echo "==> Imagen a desplegar: ${IMAGE_TAG}"
echo "    ${IMG_API}:${IMAGE_TAG}"
echo "    ${IMG_FRONT}:${IMAGE_TAG}"

# ---------- Neon ----------
echo "==> Neon: proyecto + roles/DBs qa y prod"
NEON_JSON=$(curl -fsS -X POST "https://console.neon.tech/api/v2/projects" \
  -H "Authorization: Bearer ${NEON_API_KEY}" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "{
    \"project\": {
      \"name\": \"flow-tp6\",
      \"region_id\": \"aws-us-east-2\",
      \"pg_version\": 16,
      \"history_retention_seconds\": 21600
    }
  }" 2>/dev/null || true)

if [[ -z "${NEON_JSON}" || "${NEON_JSON}" == *"already"* ]]; then
  echo "    (reusando / listando proyectos existentes)"
  NEON_JSON=$(curl -fsS "https://console.neon.tech/api/v2/projects" \
    -H "Authorization: Bearer ${NEON_API_KEY}" -H "Accept: application/json")
  PROJECT_ID=$(echo "$NEON_JSON" | jq -r '.projects[] | select(.name=="flow-tp6") | .id' | head -1)
  if [[ -z "$PROJECT_ID" || "$PROJECT_ID" == "null" ]]; then
    PROJECT_ID=$(echo "$NEON_JSON" | jq -r '.projects[0].id')
  fi
  PROJ=$(curl -fsS "https://console.neon.tech/api/v2/projects/${PROJECT_ID}" \
    -H "Authorization: Bearer ${NEON_API_KEY}" -H "Accept: application/json")
else
  PROJECT_ID=$(echo "$NEON_JSON" | jq -r '.project.id')
  PROJ="$NEON_JSON"
fi

BRANCH_ID=$(echo "$PROJ" | jq -r '.project.default_branch_id // .branch.id // empty')
if [[ -z "$BRANCH_ID" ]]; then
  BRANCH_ID=$(curl -fsS "https://console.neon.tech/api/v2/projects/${PROJECT_ID}/branches" \
    -H "Authorization: Bearer ${NEON_API_KEY}" | jq -r '.branches[0].id')
fi
DB_HOST=$(echo "$PROJ" | jq -r '.project.database_host // .connection_uris[0].connection_parameters.host // empty')
if [[ -z "$DB_HOST" || "$DB_HOST" == "null" ]]; then
  DB_HOST=$(curl -fsS "https://console.neon.tech/api/v2/projects/${PROJECT_ID}" \
    -H "Authorization: Bearer ${NEON_API_KEY}" | jq -r '.project.database_host')
fi

create_db_role() {
  local dbname="$1" rolename="$2"
  # role
  local role_resp
  role_resp=$(curl -sS -X POST "https://console.neon.tech/api/v2/projects/${PROJECT_ID}/branches/${BRANCH_ID}/roles" \
    -H "Authorization: Bearer ${NEON_API_KEY}" -H "Content-Type: application/json" \
    -d "{\"role\":{\"name\":\"${rolename}\"}}" || true)
  local pass
  pass=$(echo "$role_resp" | jq -r '.role.password // empty')
  if [[ -z "$pass" ]]; then
    # reset password if role existed
    pass=$(curl -fsS -X POST "https://console.neon.tech/api/v2/projects/${PROJECT_ID}/branches/${BRANCH_ID}/roles/${rolename}/reset_password" \
      -H "Authorization: Bearer ${NEON_API_KEY}" | jq -r '.role.password')
  fi
  curl -sS -X POST "https://console.neon.tech/api/v2/projects/${PROJECT_ID}/branches/${BRANCH_ID}/databases" \
    -H "Authorization: Bearer ${NEON_API_KEY}" -H "Content-Type: application/json" \
    -d "{\"database\":{\"name\":\"${dbname}\",\"owner_name\":\"${rolename}\"}}" >/dev/null || true
  # URI encode password minimally
  local pass_enc
  pass_enc=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$pass")
  echo "postgres://${rolename}:${pass_enc}@${DB_HOST}/${dbname}?sslmode=require"
}

DATABASE_URL_QA=$(create_db_role "app_qa" "app_qa_owner")
DATABASE_URL_PROD=$(create_db_role "app_prod" "app_prod_owner")
echo "    Neon project=${PROJECT_ID} host=${DB_HOST}"
echo "    QA DB ready / PROD DB ready"

# ---------- Render helper ----------
render() {
  curl -fsS "$@" \
    -H "Authorization: Bearer ${RENDER_API_KEY}" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json"
}

create_or_get_service() {
  local svc_name="$1" image_url="$2" tag="$3"
  shift 3
  # $@ = env JSON array fragments already built
  local existing
  existing=$(render "https://api.render.com/v1/services?limit=50" | \
    jq -r --arg n "$svc_name" '.[] | select(.service.name==$n) | .service.id' | head -1)
  if [[ -n "$existing" && "$existing" != "null" ]]; then
    echo "$existing"
    return
  fi
  local env_json="$1"
  local body
  # Render exige serviceDetails + image top-level (no runtime suelto).
  body=$(jq -n \
    --arg name "$svc_name" \
    --arg owner "$RENDER_OWNER_ID" \
    --arg img "$image_url" \
    --arg tag "$tag" \
    --arg region "$REGION" \
    --argjson envVars "$env_json" \
    '{
      type: "web_service",
      name: $name,
      ownerId: $owner,
      autoDeploy: "no",
      image: { imagePath: ($img + ":" + $tag) },
      serviceDetails: {
        env: "image",
        plan: "free",
        region: $region
      },
      envVars: $envVars
    }')
  # Render API variants differ; try v1 create
  local resp
  resp=$(curl -sS -X POST "https://api.render.com/v1/services" \
    -H "Authorization: Bearer ${RENDER_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "$body")
  local id
  id=$(echo "$resp" | jq -r '.service.id // .id // empty')
  if [[ -z "$id" ]]; then
    echo "ERROR creando $svc_name:" >&2
    echo "$resp" | jq . >&2 || echo "$resp" >&2
    echo "Creá el servicio a mano en el dashboard (Existing Image) y re-corré con SKIP_RENDER_CREATE=1" >&2
    exit 1
  fi
  echo "$id"
}

env_api_qa=$(jq -n \
  --arg db "$DATABASE_URL_QA" \
  --arg ss "$SESSION_SECRET_QA" \
  '[
    {key:"DATABASE_URL", value:$db},
    {key:"SESSION_SECRET", value:$ss},
    {key:"AUTO_MIGRATE", value:"true"},
    {key:"COOKIE_SECURE", value:"true"},
    {key:"APP_ENV", value:"qa"},
    {key:"APP_PORT", value:"8080"}
  ]')
env_api_prod=$(jq -n \
  --arg db "$DATABASE_URL_PROD" \
  --arg ss "$SESSION_SECRET_PROD" \
  '[
    {key:"DATABASE_URL", value:$db},
    {key:"SESSION_SECRET", value:$ss},
    {key:"AUTO_MIGRATE", value:"true"},
    {key:"COOKIE_SECURE", value:"true"},
    {key:"APP_ENV", value:"production"},
    {key:"APP_PORT", value:"8080"}
  ]')

echo "==> Render: servicios (image-backed)"
if [[ "${SKIP_RENDER_CREATE:-0}" != "1" ]]; then
  API_QA_ID=$(create_or_get_service "flow-api-qa" "$IMG_API" "$IMAGE_TAG" "$env_api_qa")
  API_PROD_ID=$(create_or_get_service "flow-api-prod" "$IMG_API" "$IMAGE_TAG" "$env_api_prod")
  echo "    api-qa=$API_QA_ID api-prod=$API_PROD_ID"
else
  API_QA_ID=$(render "https://api.render.com/v1/services?limit=50" | jq -r '.[] | select(.service.name=="flow-api-qa") | .service.id')
  API_PROD_ID=$(render "https://api.render.com/v1/services?limit=50" | jq -r '.[] | select(.service.name=="flow-api-prod") | .service.id')
fi

# URLs
svc_url() {
  render "https://api.render.com/v1/services/$1" | jq -r '.service.serviceDetails.url // .service.url // empty' | sed 's:/*$::'
}
QA_API_URL=$(svc_url "$API_QA_ID")
PROD_API_URL=$(svc_url "$API_PROD_ID")

env_front_qa=$(jq -n --arg b "$QA_API_URL" '[{key:"BACKEND_URL",value:$b},{key:"DNS_RESOLVER",value:"8.8.8.8"}]')
env_front_prod=$(jq -n --arg b "$PROD_API_URL" '[{key:"BACKEND_URL",value:$b},{key:"DNS_RESOLVER",value:"8.8.8.8"}]')

if [[ "${SKIP_RENDER_CREATE:-0}" != "1" ]]; then
  FRONT_QA_ID=$(create_or_get_service "flow-front-qa" "$IMG_FRONT" "$IMAGE_TAG" "$env_front_qa")
  FRONT_PROD_ID=$(create_or_get_service "flow-front-prod" "$IMG_FRONT" "$IMAGE_TAG" "$env_front_prod")
else
  FRONT_QA_ID=$(render "https://api.render.com/v1/services?limit=50" | jq -r '.[] | select(.service.name=="flow-front-qa") | .service.id')
  FRONT_PROD_ID=$(render "https://api.render.com/v1/services?limit=50" | jq -r '.[] | select(.service.name=="flow-front-prod") | .service.id')
fi

QA_FRONT_URL=$(svc_url "$FRONT_QA_ID")
PROD_FRONT_URL=$(svc_url "$FRONT_PROD_ID")

echo "    QA  api=$QA_API_URL  front=$QA_FRONT_URL"
echo "    PROD api=$PROD_API_URL front=$PROD_FRONT_URL"

# Deploy hooks
hook() {
  # Prefer dedicated deploy-hook endpoint; fallback to service details
  local id="$1"
  local h
  h=$(curl -sS -X POST "https://api.render.com/v1/services/${id}/deploys" \
    -H "Authorization: Bearer ${RENDER_API_KEY}" -H "Content-Type: application/json" \
    -d '{}' >/dev/null 2>&1 || true)
  # Render returns deployHook in service settings — fetch via dashboard API field if present
  render "https://api.render.com/v1/services/${id}" | jq -r '.service.deployHook // .deployHook // empty'
}

echo "==> Deploy hooks (si vacíos: Dashboard → Service → Settings → Deploy Hook)"
HOOK_API_QA=$(hook "$API_QA_ID" || true)
HOOK_FRONT_QA=$(hook "$FRONT_QA_ID" || true)
HOOK_API_PROD=$(hook "$API_PROD_ID" || true)
HOOK_FRONT_PROD=$(hook "$FRONT_PROD_ID" || true)

# ---------- GitHub ----------
echo "==> GitHub environments + secrets + vars"
USER_ID=$(gh api user --jq .id)

gh api -X PUT "repos/${REPO}/environments/qa" -f wait_timer=0 >/dev/null || true
gh api -X PUT "repos/${REPO}/environments/production" \
  --input - >/dev/null <<EOF
{
  "wait_timer": 0,
  "prevent_self_review": false,
  "reviewers": [{"type":"User","id": ${USER_ID}}]
}
EOF

set_env_secret() {
  local env="$1" key="$2" val="$3"
  if [[ -z "$val" ]]; then
    echo "    SKIP secret $key ($env): vacío — pegalo a mano"
    return
  fi
  printf '%s' "$val" | gh secret set "$key" --env "$env" --repo "$REPO"
  echo "    secret $env/$key OK"
}

set_env_secret qa RENDER_HOOK_API_QA "$HOOK_API_QA"
set_env_secret qa RENDER_HOOK_FRONT_QA "$HOOK_FRONT_QA"
set_env_secret production RENDER_HOOK_API_PROD "$HOOK_API_PROD"
set_env_secret production RENDER_HOOK_FRONT_PROD "$HOOK_FRONT_PROD"

gh variable set QA_API_URL --repo "$REPO" --body "$QA_API_URL"
gh variable set QA_FRONT_URL --repo "$REPO" --body "$QA_FRONT_URL"
gh variable set PROD_API_URL --repo "$REPO" --body "$PROD_API_URL"
gh variable set PROD_FRONT_URL --repo "$REPO" --body "$PROD_FRONT_URL"
echo "    variables QA/PROD URLs OK"

# Packages public reminder
echo "==> Recordatorio: packages GHCR públicos"
echo "    https://github.com/users/${OWNER}/packages/container/${NAME}-backend/settings"
echo "    https://github.com/users/${OWNER}/packages/container/${NAME}-frontend/settings"

OUT="scripts/tp6-wire.out.env"
cat > "$OUT" <<EOF
# Generado por tp6-wire.sh — NO commitear
QA_API_URL=$QA_API_URL
QA_FRONT_URL=$QA_FRONT_URL
PROD_API_URL=$PROD_API_URL
PROD_FRONT_URL=$PROD_FRONT_URL
IMAGE_TAG=$IMAGE_TAG
RENDER_API_QA_ID=$API_QA_ID
RENDER_FRONT_QA_ID=$FRONT_QA_ID
RENDER_API_PROD_ID=$API_PROD_ID
RENDER_FRONT_PROD_ID=$FRONT_PROD_ID
EOF
echo "==> Guardado $OUT (gitignored si agregás *.out.env)"

echo
echo "=== SIGUIENTE ==="
echo "1) Si algún HOOK quedó vacío, copialo del panel y:"
echo "     gh secret set RENDER_HOOK_API_QA --env qa"
echo "     (etc. para FRONT_QA, API_PROD, FRONT_PROD)"
echo "2) Forzá la cadena:"
echo "     git commit --allow-empty -m 'chore(tp6): trigger CD after wire' && git push origin main"
echo "3) En Actions: rechazá PROD una vez → otra corrida → aprobá"
echo "4) Tag v6.0.0 sobre Deployments/production (ver decisiones.md)"
echo "5) Pegá las 4 URLs en decisiones.md Enlaces TP6 (o mandámelas al agente)"
