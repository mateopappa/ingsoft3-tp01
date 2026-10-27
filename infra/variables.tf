variable "image_repo_api" {
  type        = string
  description = "GHCR image name for the API (no tag)."
  default     = "ghcr.io/mateopappa/ingsoft3-tp01-backend"
}

variable "image_repo_front" {
  type        = string
  description = "GHCR image name for the frontend (no tag)."
  default     = "ghcr.io/mateopappa/ingsoft3-tp01-frontend"
}

variable "image_tag" {
  type        = string
  description = "Immutable image tag from the TP7 pipeline (sha-<40-char-commit>)."
  # b468ca1 = v7.0.0 / PROD: Host $proxy_host, Neon pooler fix, Save Activity.
  # sha-596edde… loops /api (508). Override with -var when promoting a newer sha.
  default = "sha-b468ca1da77ef75dc407513b40b9d1b143d820d9"
}

variable "region" {
  type        = string
  description = "Render free-tier region."
  default     = "oregon"

  validation {
    condition     = contains(["frankfurt", "ohio", "oregon", "singapore", "virginia"], var.region)
    error_message = "region must be one of: frankfurt, ohio, oregon, singapore, virginia."
  }
}
