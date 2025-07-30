output "server_id" {
  description = "The ID of the PostgreSQL server"
  value       = azurerm_postgresql_flexible_server.main.id
}

output "server_name" {
  description = "The name of the PostgreSQL server"
  value       = azurerm_postgresql_flexible_server.main.name
}

output "server_fqdn" {
  description = "The FQDN of the PostgreSQL server"
  value       = azurerm_postgresql_flexible_server.main.fqdn
}

output "admin_username" {
  description = "The administrator username"
  value       = azurerm_postgresql_flexible_server.main.administrator_login
}

output "admin_password" {
  description = "The administrator password"
  value       = var.admin_password != null ? var.admin_password : random_password.admin_password[0].result
  sensitive   = true
}

output "database_name" {
  description = "The name of the main database"
  value       = azurerm_postgresql_flexible_server_database.hub.name
}

output "connection_string" {
  description = "The connection string for the PostgreSQL server"
  value       = "postgresql://${azurerm_postgresql_flexible_server.main.administrator_login}:${var.admin_password != null ? var.admin_password : random_password.admin_password[0].result}@${azurerm_postgresql_flexible_server.main.fqdn}:5432/${azurerm_postgresql_flexible_server_database.hub.name}?sslmode=require"
  sensitive   = true
}

output "private_dns_zone_id" {
  description = "The ID of the private DNS zone"
  value       = var.public_network_access_enabled ? null : azurerm_private_dns_zone.postgresql[0].id
}
