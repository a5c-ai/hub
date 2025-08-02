 variable "controller_namespace" {
  description = "Kubernetes namespace for the runner controller"
  type        = string
  default     = "arc-systems"
}

variable "runners_namespace" {
  description = "Kubernetes namespace for the runner scale sets"
  type        = string
  default     = "arc-runners"
}

variable "controller_chart_version" {
  description = "Version of the gha-runner-scale-set-controller Helm chart"
  type        = string
  default     = "0.9.3"
}

variable "runner_set_chart_version" {
  description = "Version of the gha-runner-scale-set Helm chart"
  type        = string
  default     = "0.9.3"
}

variable "github_config_url" {
  description = "GitHub organization URL (e.g., https://github.com/my-org)"
  type        = string
}

variable "github_app_id" {
  description = "GitHub App ID for authentication"
  type        = string
}

variable "github_app_installation_id" {
  description = "GitHub App Installation ID"
  type        = string
}

variable "github_app_private_key" {
  description = "GitHub App private key (PEM format)"
  type        = string
  sensitive   = true
}

variable "runner_scale_set_name" {
  description = "Name for the runner scale set (this will be the label used in runs-on)"
  type        = string
  default     = "arc-runner-set"
}

variable "min_runners" {
  description = "Minimum number of runner replicas"
  type        = number
  default     = 0
}

variable "max_runners" {
  description = "Maximum number of runner replicas"
  type        = number
  default     = 10
}

variable "runner_group" {
  description = "Runner group name (optional)"
  type        = string
  default     = "default"
}

variable "runner_labels" {
  description = "Additional labels to assign to runners"
  type        = list(string)
  default     = []
}

variable "container_mode" {
  description = "Container mode for runners (dind or kubernetes)"
  type        = string
  default     = "dind"
  validation {
    condition     = contains(["dind", "kubernetes"], var.container_mode)
    error_message = "Container mode must be either 'dind' or 'kubernetes'."
  }
}

variable "runner_image" {
  description = "Custom runner image (optional)"
  type        = string
  default     = null
}
