 output "controller_release_name" {
  description = "Name of the Helm release for the ARC controller"
  value       = helm_release.arc_controller.name
}

output "runner_set_release_name" {
  description = "Name of the Helm release for the runner scale set"
  value       = helm_release.arc_runner_set.name
}

output "controller_namespace" {
  description = "Namespace where the ARC controller is deployed"
  value       = var.controller_namespace
}

output "runners_namespace" {
  description = "Namespace where the runner scale set is deployed"
  value       = var.runners_namespace
}

output "runner_scale_set_name" {
  description = "Name of the runner scale set (use this in 'runs-on' in GitHub workflows)"
  value       = var.runner_scale_set_name
}

output "github_config_url" {
  description = "GitHub organization URL configured for the runners"
  value       = var.github_config_url
}

output "github_secret_name" {
  description = "Name of the Kubernetes secret containing GitHub credentials"
  value       = kubernetes_secret.github_secret.metadata[0].name
}
