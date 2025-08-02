variable "cert_manager_version" {
  description = "Version of cert-manager Helm chart to install"
  type        = string
  default     = "v1.15.3"
}

variable "email" {
  description = "Email address for Let's Encrypt registration"
  type        = string
  default     = "support@a5c.ai"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "cluster_issuer_name" {
  description = "Name of the cluster issuer to create"
  type        = string
  default     = "letsencrypt-prod"
}

variable "staging_cluster_issuer_name" {
  description = "Name of the staging cluster issuer to create"
  type        = string
  default     = "letsencrypt-staging"
}

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}