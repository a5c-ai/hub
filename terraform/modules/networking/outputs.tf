output "vnet_id" {
  description = "The ID of the virtual network"
  value       = azurerm_virtual_network.main.id
}

output "vnet_name" {
  description = "The name of the virtual network"
  value       = azurerm_virtual_network.main.name
}

output "aks_subnet_id" {
  description = "The ID of the AKS subnet"
  value       = azurerm_subnet.aks.id
}

output "database_subnet_id" {
  description = "The ID of the database subnet"
  value       = azurerm_subnet.database.id
}

output "private_endpoints_subnet_id" {
  description = "The ID of the private endpoints subnet"
  value       = azurerm_subnet.private_endpoints.id
}

output "application_gateway_subnet_id" {
  description = "The ID of the application gateway subnet"
  value       = azurerm_subnet.application_gateway.id
}

output "keyvault_private_dns_zone_id" {
  description = "The ID of the Key Vault private DNS zone"
  value       = azurerm_private_dns_zone.keyvault.id
}

output "storage_private_dns_zone_id" {
  description = "The ID of the storage private DNS zone"
  value       = azurerm_private_dns_zone.storage.id
}

output "aks_nsg_id" {
  description = "The ID of the AKS network security group"
  value       = azurerm_network_security_group.aks.id
}

output "database_nsg_id" {
  description = "The ID of the database network security group"
  value       = azurerm_network_security_group.database.id
}