 variable "namespace" {
   description = "Kubernetes namespace for the runner controller and runners"
   type        = string
   default     = "actions-runner-system"
 }

 variable "chart_version" {
   description = "Version of the actions-runner-controller Helm chart"
   type        = string
   default     = "0.24.2"
 }

 variable "github_token" {
   description = "GitHub token with permissions to register and manage runners"
   type        = string
   sensitive   = true
 }

 variable "github_owner" {
   description = "GitHub organization or user for the repository"
   type        = string
 }

 variable "github_repository" {
   description = "GitHub repository name for self-hosted runners"
   type        = string
 }

 variable "runner_deployment_name" {
   description = "Name for the RunnerDeployment resource"
   type        = string
   default     = "runner"
 }

 variable "runner_replicas" {
   description = "Number of runner pods to maintain"
   type        = number
   default     = 2
 }

 variable "runner_labels" {
   description = "Additional labels to assign to runner pods"
   type        = map(string)
   default     = {}
 }
