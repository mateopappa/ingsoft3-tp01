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
  # Published on main by TP6/TP7 publish job; override when promoting a newer sha.
  default = "sha-596edde63b87b10552d01d59956ebcc343d2a814"
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
