output "db_host" {
  description = "Neon host for preprod (not a secret)."
  value       = neon_project.preprod.database_host
}

output "database" {
  description = "Neon database name for preprod."
  value       = neon_database.preprod.name
}

output "api_url" {
  description = "Public URL of the preprod API (no trailing slash)."
  value       = local.api_url
}

output "front_url" {
  description = "Public URL of the preprod frontend."
  value       = render_web_service.front.url
}

output "image_tag" {
  description = "Image tag currently declared for preprod services."
  value       = var.image_tag
}
