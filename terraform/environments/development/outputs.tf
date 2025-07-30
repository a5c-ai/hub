output "resource_group_name" {
  description = "The name of the resource group"
  value       = module.resource_group.name
}

output "aks_cluster_name" {
  description = "The name of the AKS cluster"
  value       = module.aks.cluster_name
}

output "aks_cluster_fqdn" {
  description = "The FQDN of the AKS cluster"
  value       = module.aks.cluster_fqdn
}

output "postgresql_fqdn" {
  description = "The FQDN of the PostgreSQL server"
  value       = module.postgresql.server_fqdn
}

output "postgresql_admin_username" {
  description = "The administrator username for PostgreSQL"
  value       = module.postgresql.admin_username
}

output "storage_account_name" {
  description = "The name of the storage account"
  value       = module.storage.storage_account_name
}

output "storage_primary_blob_endpoint" {
  description = "The primary blob endpoint of the storage account"
  value       = module.storage.primary_blob_endpoint
}

output "key_vault_name" {
  description = "The name of the Key Vault"
  value       = module.keyvault.key_vault_name
}

output "key_vault_uri" {
  description = "The URI of the Key Vault"
  value       = module.keyvault.key_vault_uri
}

output "application_gateway_public_ip" {
  description = "The public IP address of the Application Gateway"
  value       = module.security.public_ip_address
}

output "log_analytics_workspace_name" {
  description = "The name of the Log Analytics workspace"
  value       = module.monitoring.log_analytics_workspace_name
}

output "application_insights_name" {
  description = "The name of the Application Insights instance"
  value       = module.monitoring.application_insights_name
}

# Sensitive outputs
output "postgresql_admin_password" {
  description = "The administrator password for PostgreSQL"
  value       = module.postgresql.admin_password
  sensitive   = true
}

output "kube_config" {
  description = "The kubeconfig for the AKS cluster"
  value       = module.aks.kube_config_raw
  sensitive   = true
}

output "storage_connection_string" {
  description = "The connection string for the storage account"
  value       = module.storage.primary_connection_string
  sensitive   = true
}

output "application_insights_instrumentation_key" {
  description = "The instrumentation key for Application Insights"
  value       = module.monitoring.application_insights_instrumentation_key
  sensitive   = true
}

output "application_insights_connection_string" {
  description = "The connection string for Application Insights"
  value       = module.monitoring.application_insights_connection_string
  sensitive   = true
}

output "container_registry_login_server" {
  description = "The login server for the Container Registry"
  value       = module.container_registry.login_server
}
