terraform {
  required_version = ">= 1.5"

  required_providers {
    neon = {
      source  = "kislerdm/neon"
      version = "0.18.0"
    }
    render = {
      source  = "render-oss/render"
      version = "1.9.1"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }
}

# Credentials from the environment — never written in HCL:
#   NEON_API_KEY, RENDER_API_KEY, RENDER_OWNER_ID
provider "neon" {}
provider "render" {}

# --- Neon project for PREPROD only (QA/PROD stay click-ops, untouched) ---

resource "neon_project" "preprod" {
  name = "flow-preprod"

  # Free tier max history retention is 21600s (6h); provider default is 86400.
  history_retention_seconds = 21600
}

resource "neon_role" "preprod" {
  project_id = neon_project.preprod.id
  branch_id  = neon_project.preprod.default_branch_id
  name       = "app_preprod_owner"
}

resource "neon_database" "preprod" {
  project_id = neon_project.preprod.id
  branch_id  = neon_project.preprod.default_branch_id
  name       = "app_preprod"
  owner_name = neon_role.preprod.name
}

# Session cookie secret for Flow — derived, not copied from another env.
resource "random_password" "session" {
  length  = 48
  special = false
}

locals {
  # Postgres URI Flow reads as DATABASE_URL. Built from Neon outputs — never pasted.
  database_url = format(
    "postgres://%s:%s@%s/%s?sslmode=require",
    neon_role.preprod.name,
    neon_role.preprod.password,
    neon_project.preprod.database_host,
    neon_database.preprod.name,
  )

  # Render may return a trailing slash; nginx BACKEND_URL must not have one.
  api_url = trimsuffix(render_web_service.api.url, "/")
}

resource "render_web_service" "api" {
  name   = "flow-api-preprod"
  plan   = "free"
  region = var.region

  runtime_source = {
    image = {
      image_url = var.image_repo_api
      tag       = var.image_tag
    }
  }

  env_vars = {
    DATABASE_URL   = { value = local.database_url }
    SESSION_SECRET = { value = random_password.session.result }
    AUTO_MIGRATE   = { value = "true" }
    COOKIE_SECURE  = { value = "true" }
    APP_ENV        = { value = "preprod" }
    APP_PORT       = { value = "8080" }
  }
}

resource "render_web_service" "front" {
  name   = "flow-front-preprod"
  plan   = "free"
  region = var.region

  runtime_source = {
    image = {
      image_url = var.image_repo_front
      tag       = var.image_tag
    }
  }

  env_vars = {
    BACKEND_URL  = { value = local.api_url }
    DNS_RESOLVER = { value = "8.8.8.8" }
  }
}
