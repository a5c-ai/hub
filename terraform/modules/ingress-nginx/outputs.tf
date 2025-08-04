output "namespace" {
  description = "The namespace where ingress-nginx controller is installed"
  value       = var.namespace
}

output "helm_release_name" {
  description = "The name of the ingress-nginx Helm release"
  value       = helm_release.controller[0].name
}

output "helm_release_status" {
  description = "The status of the ingress-nginx Helm release"
  value       = helm_release.controller[0].status
}
