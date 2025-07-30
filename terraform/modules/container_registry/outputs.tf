output "registry_id" {
  description = "The ID of the Container Registry"
  value       = azurerm_container_registry.main.id
}

output "registry_name" {
  description = "The name of the Container Registry"
  value       = azurerm_container_registry.main.name
}

output "login_server" {
  description = "The login server for the Container Registry"
  value       = azurerm_container_registry.main.login_server
}
