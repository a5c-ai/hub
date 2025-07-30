 output "helm_release_name" {
   description = "Name of the Helm release for actions-runner-controller"
   value       = helm_release.controller.name
 }

 output "runner_deployment_name" {
   description = "Name of the RunnerDeployment resource"
   value       = var.runner_deployment_name
 }

 output "runner_namespace" {
   description = "Namespace where runner controller and runners are deployed"
   value       = var.namespace
 }
