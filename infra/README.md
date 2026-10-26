# infra/ — preprod as code (TP8)

Declares a **new** `preprod` environment (Neon project + Render api/front running GHCR images).
**Does not touch** QA or production.

## Prerequisites

```bash
export NEON_API_KEY="…"
export RENDER_API_KEY="…"
export RENDER_OWNER_ID="usr-…"   # or tea-… from the Render dashboard URL
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

# When idle (Render free-tier quota is shared):
terraform destroy
```

State lives locally as `terraform.tfstate` (gitignored). Never commit it.

## Changing the image

Edit `var.image_tag` (or pass `-var='image_tag=sha-…'`). On the free Render plan,
updates in place fail — recreate with:

```bash
terraform apply -replace=render_web_service.api -replace=render_web_service.front
# or: terraform destroy && terraform apply
```
