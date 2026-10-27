# infra/ — preprod as code (TP8)

Declares a **new** `preprod` environment (Neon project + Render api/front running GHCR images).
**Does not touch** QA or production.

## Prerequisites

```bash
export NEON_API_KEY="…"
export RENDER_API_KEY="…"
# Owner for THIS key. The id in the dashboard URL (usr-…) is the user and
# Render rejects it as ownerID. Discover the workspace id:
#   curl -sH "Authorization: Bearer $RENDER_API_KEY" https://api.render.com/v1/owners
# Use the id that comes back (for a team workspace it looks like tea-…).
export RENDER_OWNER_ID="tea-…"
```

Packages must be **public** so Render can pull without credentials.

## Lifecycle

```bash
cd infra
terraform init
terraform fmt && terraform validate
terraform plan
terraform apply
terraform output

# Verify as an environment (not just resources):
curl -fsS "$(terraform output -raw api_url)/api/health"
curl -fsS "$(terraform output -raw front_url)/"

# When idle (Render free-tier quota is shared with QA and PROD):
terraform destroy
```

Defense (preprod stays destroyed until then, so `apply` creates it live):

```bash
set -a && source ../.env && set +a
terraform init
terraform apply
curl -fsS "$(terraform output -raw api_url)/api/health"
curl -fsS "$(terraform output -raw front_url)/api/health"
terraform destroy
```

The default `image_tag` is `sha-e810b299b20b96f6ae2d3f99d1c762ea0c49fb18`.
That image is the last main publish whose front still says `Save Activity` and whose nginx sends `Host $proxy_host`.

State lives locally as `terraform.tfstate` (gitignored). Never commit it.

## Changing the image

Edit `var.image_tag` (or pass `-var='image_tag=sha-…'`). On the free Render plan,
updates in place fail — recreate with:

```bash
terraform apply -replace=render_web_service.api -replace=render_web_service.front
# or: terraform destroy && terraform apply
```
