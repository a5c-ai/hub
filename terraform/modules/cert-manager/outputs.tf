output "namespace" {
  description = "The namespace where cert-manager is installed"
  value       = "cert-manager"
}

output "helm_release_name" {
  description = "The name of the cert-manager Helm release"
  value       = helm_release.cert_manager.name
}

output "helm_release_status" {
  description = "The status of the cert-manager Helm release"
  value       = helm_release.cert_manager.status
}

output "cluster_issuer_name" {
  description = "The name of the production cluster issuer"
  value       = var.cluster_issuer_name
}

output "staging_cluster_issuer_name" {
  description = "The name of the staging cluster issuer"
  value       = var.staging_cluster_issuer_name
}