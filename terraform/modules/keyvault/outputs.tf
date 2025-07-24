output "key_vault_id" {
  description = "The ID of the Key Vault"
  value       = azurerm_key_vault.main.id
}

output "key_vault_name" {
  description = "The name of the Key Vault"
  value       = azurerm_key_vault.main.name
}

output "key_vault_uri" {
  description = "The URI of the Key Vault"
  value       = azurerm_key_vault.main.vault_uri
}

output "key_vault_tenant_id" {
  description = "The tenant ID of the Key Vault"
  value       = azurerm_key_vault.main.tenant_id
}

output "secret_ids" {
  description = "The IDs of the secrets"
  value       = { for k, v in azurerm_key_vault_secret.secrets : k => v.id }
}

output "secret_versions" {
  description = "The versions of the secrets"
  value       = { for k, v in azurerm_key_vault_secret.secrets : k => v.version }
}

output "key_ids" {
  description = "The IDs of the keys"
  value       = { for k, v in azurerm_key_vault_key.keys : k => v.id }
}

output "key_versions" {
  description = "The versions of the keys"
  value       = { for k, v in azurerm_key_vault_key.keys : k => v.version }
}

output "private_endpoint_id" {
  description = "The ID of the private endpoint"
  value       = var.enable_private_endpoint ? azurerm_private_endpoint.keyvault[0].id : null
}

output "private_endpoint_ip" {
  description = "The private IP address of the private endpoint"
  value       = var.enable_private_endpoint ? azurerm_private_endpoint.keyvault[0].private_service_connection[0].private_ip_address : null
}