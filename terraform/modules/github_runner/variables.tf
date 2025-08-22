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
  default     = "0.12.1"
}

variable "runner_set_chart_version" {
  description = "Version of the gha-runner-scale-set Helm chart"
  type        = string
  default     = "0.12.1"
}

variable "github_config_url" {
  description = "GitHub organization URL (e.g., https://github.com/my-org)"
  type        = string
}

variable "auth_method" {
  description = "Authentication method: 'app' for GitHub App or 'token' for GitHub token"
  type        = string
  default     = "app"
  validation {
    condition     = contains(["app", "token"], var.auth_method)
    error_message = "Auth method must be either 'app' or 'token'."
  }
}

# GitHub App Authentication (when auth_method = "app")
variable "github_app_id" {
  description = "GitHub App ID for authentication"
  type        = string
  default     = ""
}

variable "github_app_installation_id" {
  description = "GitHub App Installation ID"
  type        = string
  default     = ""
}

variable "github_app_private_key" {
  description = "GitHub App private key (PEM format)"
  type        = string
  sensitive   = true
  default     = ""
}

# GitHub Token Authentication (when auth_method = "token")
variable "github_token" {
  description = "GitHub token with permissions to register and manage runners"
  type        = string
  sensitive   = true
  default     = ""
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
  default     = "kubernetes"
  validation {
    condition     = contains(["dind", "kubernetes"], var.container_mode)
    error_message = "Container mode must be either 'dind' or 'kubernetes'."
  }
}

variable "runner_image" {
  description = "Custom runner image (optional)"
  type        = string
  default     = "ghcr.io/actions/actions-runner:2.328.0"
}

variable "storage_class_name" {
  description = "Storage class name for ephemeral volumes"
  type        = string
  default     = ""  # Empty string to use cluster default
}

variable "ephemeral_storage_size" {
  description = "Storage size for ephemeral workspace PVC"
  type        = string
  default     = "10Gi"
}

variable "enable_init_container" {
  description = "Enable init container to install prerequisites dynamically"
  type        = bool
  default     = true
}

variable "runner_node_selector" {
  description = "Node selector labels to schedule GitHub runner pods on specific node pools"
  type        = map(string)
  default     = {}
}

variable "use_pvc_for_work_volume" {
  description = "Whether to provision a PVC for the runner work volume (kubernetes mode). If false, uses emptyDir."
  type        = bool
  default     = false
}