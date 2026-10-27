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
  # e810b29: nginx Host=$proxy_host (no 508) and the button still says "Save Activity".
  # sha-596edde… is public but its front loops /api on Render. Do not use it.
  # Override with -var when promoting a newer sha. Avoid a sha whose front button
  # was intentionally renamed for the TP7 red-e2e evidence.
  default = "sha-e810b299b20b96f6ae2d3f99d1c762ea0c49fb18"
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
