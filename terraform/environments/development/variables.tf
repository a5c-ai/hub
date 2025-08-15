variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
  default     = "West US 3"
}

variable "owner" {
  description = "The owner of the resources"
  type        = string
  default     = "hub-team"
}

# Networking Variables
variable "vnet_address_space" {
  description = "The address space of the virtual network"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "aks_subnet_cidr" {
  description = "The CIDR block for the AKS subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "database_subnet_cidr" {
  description = "The CIDR block for the database subnet"
  type        = string
  default     = "10.0.2.0/24"
}

variable "private_endpoints_subnet_cidr" {
  description = "The CIDR block for the private endpoints subnet"
  type        = string
  default     = "10.0.3.0/24"
}

variable "appgw_subnet_cidr" {
  description = "The CIDR block for the application gateway subnet"
  type        = string
  default     = "10.0.4.0/24"
}

variable "admin_source_address_prefix" {
  description = "Source address prefix for admin access"
  type        = string
  default     = "*"
}

variable "enable_private_endpoints" {
  description = "Enable private endpoints for services"
  type        = bool
  default     = true
}

# Storage Variables
variable "storage_account_tier" {
  description = "The tier of the storage account"
  type        = string
  default     = "Standard"
}

variable "storage_replication_type" {
  description = "The replication type of the storage account"
  type        = string
  default     = "LRS"
}

# PostgreSQL Variables
variable "postgresql_version" {
  description = "The version of PostgreSQL to use"
  type        = string
  default     = "15"
}

variable "postgresql_sku_name" {
  description = "The SKU name for the PostgreSQL server"
  type        = string
  default     = "B_Standard_B1ms"
}

variable "postgresql_storage_mb" {
  description = "The storage size in MB for the PostgreSQL server"
  type        = number
  default     = 32768
}

variable "postgresql_backup_retention_days" {
  description = "The number of days to retain backups"
  type        = number
  default     = 7
}

# AKS Variables
variable "kubernetes_version" {
  description = "The version of Kubernetes to use"
  type        = string
  default     = "1.30"
}

variable "aks_node_count" {
  description = "The number of nodes in the AKS cluster"
  type        = number
  default     = 2
}

variable "aks_vm_size" {
  description = "The size of the virtual machines in the AKS cluster"
  type        = string
  default     = "Standard_D2s_v5"
}

# Worker Node Pool Configuration (optional)
variable "create_worker_node_pool" {
  description = "Create an additional worker node pool"
  type        = bool
  default     = false
}

variable "worker_vm_size" {
  description = "The size of the virtual machines in the worker node pool"
  type        = string
  default     = "Standard_D4s_v5"
}

variable "worker_node_count" {
  description = "The number of nodes in the worker node pool"
  type        = number
  default     = 2
}

variable "worker_min_node_count" {
  description = "The minimum number of nodes in the worker node pool"
  type        = number
  default     = 1
}

variable "worker_max_node_count" {
  description = "The maximum number of nodes in the worker node pool"
  type        = number
  default     = 4
}

# Monitoring Variables
variable "alert_email_receivers" {
  description = "List of email receivers for alerts"
  type = list(object({
    name          = string
    email_address = string
  }))
  default = []
}

# GitHub Runner Variables (Modern ARC)
variable "enable_github_runners" {
  description = "Enable GitHub Actions Runner Controller (ARC)"
  type        = bool
  default     = true
}

variable "github_config_url" {
  description = "GitHub organization URL (e.g., https://github.com/your-org)"
  type        = string
  default     = ""
  
  validation {
    condition = can(regex("^https://github\\.com/[a-zA-Z0-9._-]+$", var.github_config_url)) || var.github_config_url == ""
    error_message = "GitHub config URL must be a valid GitHub organization URL (e.g., https://github.com/your-org)."
  }
}

variable "github_auth_method" {
  description = "Authentication method: 'app' for GitHub App or 'token' for GitHub token"
  type        = string
  default     = "token"
  validation {
    condition     = contains(["app", "token"], var.github_auth_method)
    error_message = "Auth method must be either 'app' or 'token'."
  }
}

variable "github_token" {
  description = "GitHub token with permissions to register and manage runners (when using token auth)"
  type        = string
  sensitive   = true
  default     = ""
}

variable "runner_scale_set_name" {
  description = "Name for the runner scale set (used in 'runs-on' in workflows)"
  type        = string
  default     = "hub-dev-runners"
}

variable "runner_min_replicas" {
  description = "Minimum number of runners"
  type        = number
  default     = 0
}

variable "runner_max_replicas" {
  description = "Maximum number of runners"
  type        = number
  default     = 5
}

variable "runner_container_mode" {
  description = "Container mode for runners (dind or kubernetes)"
  type        = string
  default     = "kubernetes"
}

variable "runner_labels" {
  description = "Additional labels to assign to runners"
  type        = list(string)
  default     = ["development"]
}

# DNS Configuration
variable "public_dns_zone_name" {
  description = "The name of the public DNS zone"
  type        = string
  default     = ""
}

# AGIC Role Assignment Configuration
variable "create_agic_role_assignments" {
  description = "Whether to create AGIC role assignments (set to false if they already exist)"
  type        = bool
  default     = false
}

# Runner Image Configuration
variable "runner_image" {
  description = "Custom Docker image for GitHub Actions runners (optional)"
  type        = string
  default     = null
}

# Ingress NGINX Controller
variable "ingress_nginx_controller_enabled" {
  description = "Enable provisioning of the ingress-nginx controller"
  type        = bool
  default     = true
}

variable "ingress_nginx_controller_release_name" {
  description = "Helm release name for the ingress-nginx controller"
  type        = string
  default     = "ingress-nginx"
}

variable "ingress_nginx_controller_chart_version" {
  description = "Version of the ingress-nginx Helm chart"
  type        = string
  default     = ""
}

variable "ingress_nginx_controller_namespace" {
  description = "Namespace to install ingress-nginx controller into"
  type        = string
  default     = "ingress-nginx"
}

variable "ingress_nginx_controller_values" {
  description = "Overrides for the ingress-nginx Helm chart values"
  type        = map(string)
  default     = {}
}
