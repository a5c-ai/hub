variable "enabled" {
  description = "Whether to install the ingress-nginx controller"
  type        = bool
  default     = true
}

variable "release_name" {
  description = "Name of the Helm release for ingress-nginx"
  type        = string
  default     = "ingress-nginx"
}

variable "chart_repository" {
  description = "Helm repository for the ingress-nginx chart"
  type        = string
  default     = "https://kubernetes.github.io/ingress-nginx"
}

variable "chart_name" {
  description = "Name of the ingress-nginx chart"
  type        = string
  default     = "ingress-nginx"
}

variable "chart_version" {
  description = "Version of the ingress-nginx Helm chart to install"
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Kubernetes namespace into which to install the ingress-nginx controller"
  type        = string
  default     = "ingress-nginx"
}

variable "create_namespace" {
  description = "Whether to create the namespace if it does not exist"
  type        = bool
  default     = true
}

variable "values" {
  description = "Flat map of chart values to override for ingress-nginx controller"
  type        = map(string)
  default     = {}
}

variable "timeout" {
  description = "Timeout in seconds for the Helm release"
  type        = number
  default     = 300
}

variable "wait_duration" {
  description = "Duration to wait after Helm install for controller readiness (e.g., '60s')"
  type        = string
  default     = "60s"
}
