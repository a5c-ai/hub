output "cluster_id" {
  description = "The ID of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.id
}

output "cluster_name" {
  description = "The name of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.name
}

output "cluster_fqdn" {
  description = "The FQDN of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.fqdn
}

output "kube_config" {
  description = "The kubeconfig for the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.kube_config
  sensitive   = true
}

output "kube_config_raw" {
  description = "The raw kubeconfig for the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.kube_config_raw
  sensitive   = true
}

output "kubelet_identity" {
  description = "The kubelet identity of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.kubelet_identity
}

output "oidc_issuer_url" {
  description = "The OIDC issuer URL for the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.oidc_issuer_url
}

output "log_analytics_workspace_id" {
  description = "The ID of the log analytics workspace"
  value       = azurerm_log_analytics_workspace.aks.id
}

output "cluster_identity_principal_id" {
  description = "The principal ID of the cluster identity"
  value       = azurerm_user_assigned_identity.aks.principal_id
}

output "cluster_identity_client_id" {
  description = "The client ID of the cluster identity"
  value       = azurerm_user_assigned_identity.aks.client_id
}

output "ingress_application_gateway" {
  description = "The Application Gateway Ingress Controller configuration"
  value       = var.enable_application_gateway_ingress ? azurerm_kubernetes_cluster.main.ingress_application_gateway : null
}

output "agic_identity_client_id" {
  description = "The client ID of the AGIC identity"
  value       = var.enable_application_gateway_ingress ? azurerm_kubernetes_cluster.main.ingress_application_gateway[0].ingress_application_gateway_identity[0].client_id : null
}

output "agic_identity_object_id" {
  description = "The object ID of the AGIC identity"
  value       = var.enable_application_gateway_ingress ? azurerm_kubernetes_cluster.main.ingress_application_gateway[0].ingress_application_gateway_identity[0].object_id : null
}