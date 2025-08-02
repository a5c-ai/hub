variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
  default     = "West US 3"
}

variable "backup_location" {
  description = "The Azure location for backup resources"
  type        = string
  default     = "West US 3"
}

variable "owner" {
  description = "The owner of the resources"
  type        = string
  default     = "hub-production-team"
}

variable "cost_center" {
  description = "The cost center for the resources"
  type        = string
  default     = "engineering"
}

variable "availability_zones" {
  description = "The availability zones to use"
  type        = list(string)
  default     = ["1", "2", "3"]
}

# Networking Variables
variable "vnet_address_space" {
  description = "The address space of the virtual network"
  type        = list(string)
  default     = ["10.1.0.0/16"]
}

variable "aks_subnet_cidr" {
  description = "The CIDR block for the AKS subnet"
  type        = string
  default     = "10.1.1.0/24"
}

variable "database_subnet_cidr" {
  description = "The CIDR block for the database subnet"
  type        = string
  default     = "10.1.2.0/24"
}

variable "private_endpoints_subnet_cidr" {
  description = "The CIDR block for the private endpoints subnet"
  type        = string
  default     = "10.1.3.0/24"
}

variable "appgw_subnet_cidr" {
  description = "The CIDR block for the application gateway subnet"
  type        = string
  default     = "10.1.4.0/24"
}

variable "admin_source_address_prefix" {
  description = "Source address prefix for admin access"
  type        = string
  default     = "10.0.0.0/8"
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
  default     = "ZRS"
}

variable "lifecycle_cool_after_days" {
  description = "Days after which to move blobs to cool storage"
  type        = number
  default     = 30
}

variable "lifecycle_archive_after_days" {
  description = "Days after which to move blobs to archive storage"
  type        = number
  default     = 90
}

variable "lifecycle_delete_after_days" {
  description = "Days after which to delete blobs"
  type        = number
  default     = 2555  # 7 years
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
  default     = "GP_Standard_D4s_v3"
}

variable "postgresql_storage_mb" {
  description = "The storage size in MB for the PostgreSQL server"
  type        = number
  default     = 1048576  # 1TB
}

variable "postgresql_backup_retention_days" {
  description = "The number of days to retain backups"
  type        = number
  default     = 35
}

variable "postgresql_geo_redundant_backup_enabled" {
  description = "Enable geo-redundant backups"
  type        = bool
  default     = true
}

variable "postgresql_high_availability_mode" {
  description = "The high availability mode for the PostgreSQL server"
  type        = string
  default     = "ZoneRedundant"
}

variable "postgresql_standby_availability_zone" {
  description = "The availability zone for the standby server"
  type        = string
  default     = "2"
}

# AKS Variables
variable "kubernetes_version" {
  description = "The version of Kubernetes to use"
  type        = string
  default     = "1.29"
}

variable "aks_node_count" {
  description = "The number of nodes in the AKS cluster"
  type        = number
  default     = 3
}

variable "aks_vm_size" {
  description = "The size of the virtual machines in the AKS cluster"
  type        = string
  default     = "Standard_D4s_v5"
}

variable "aks_min_node_count" {
  description = "The minimum number of nodes in the AKS cluster"
  type        = number
  default     = 3
}

variable "aks_max_node_count" {
  description = "The maximum number of nodes in the AKS cluster"
  type        = number
  default     = 20
}

variable "create_worker_node_pool" {
  description = "Create an additional worker node pool"
  type        = bool
  default     = true
}

variable "worker_vm_size" {
  description = "The size of the virtual machines in the worker node pool"
  type        = string
  default     = "Standard_D8s_v5"
}

variable "worker_node_count" {
  description = "The number of nodes in the worker node pool"
  type        = number
  default     = 3
}

variable "worker_min_node_count" {
  description = "The minimum number of nodes in the worker node pool"
  type        = number
  default     = 2
}

variable "worker_max_node_count" {
  description = "The maximum number of nodes in the worker node pool"
  type        = number
  default     = 10
}

# Application Gateway Variables
variable "application_gateway_capacity" {
  description = "The capacity for the Application Gateway"
  type        = number
  default     = 3
}

variable "waf_file_upload_limit_mb" {
  description = "The file upload limit in MB for the Web Application Firewall"
  type        = number
  default     = 500
}

variable "waf_max_request_body_size_kb" {
  description = "The maximum request body size in KB for the Web Application Firewall"
  type        = number
  default     = 512
}

variable "waf_rate_limit_threshold" {
  description = "The rate limit threshold for the Web Application Firewall"
  type        = number
  default     = 300
}

variable "ssl_certificate_data" {
  description = "The SSL certificate data in base64 format"
  type        = string
  default     = null
  sensitive   = true
}

variable "ssl_certificate_password" {
  description = "The SSL certificate password"
  type        = string
  default     = null
  sensitive   = true
}

# Monitoring Variables
variable "log_retention_days" {
  description = "The number of days to retain logs"
  type        = number
  default     = 90
}

variable "log_analytics_daily_quota_gb" {
  description = "The daily quota for the Log Analytics workspace in GB"
  type        = number
  default     = 50
}

variable "application_insights_retention_days" {
  description = "The retention period for Application Insights data"
  type        = number
  default     = 730  # 2 years
}

variable "enable_grafana" {
  description = "Enable Azure Managed Grafana"
  type        = bool
  default     = true
}

variable "alert_email_receivers" {
  description = "List of email receivers for alerts"
  type = list(object({
    name          = string
    email_address = string
  }))
  default = []
}

# GitHub Runner Variables
variable "github_token" {
  description = "GitHub token for self-hosted runner registration"
  type        = string
  default     = ""
}

variable "github_owner" {
  description = "GitHub organization or user"
  type        = string
  default     = ""
}

variable "github_repository" {
  description = "GitHub repository name"
  type        = string
  default     = ""
}

variable "runner_deployment_name" {
  description = "Name for the RunnerDeployment resource"
  type        = string
  default     = "runner"
}

variable "runner_replicas" {
  description = "Number of runners to maintain"
  type        = number
  default     = 2
}

variable "runner_labels" {
  description = "Labels to apply to runner pods"
  type        = map(string)
  default     = {}
}

# DNS Configuration
variable "public_dns_zone_name" {
  description = "The name of the public DNS zone"
  type        = string
  default     = ""
}

variable "alert_sms_receivers" {
  description = "List of SMS receivers for alerts"
  type = list(object({
    name         = string
    country_code = string
    phone_number = string
  }))
  default = []
}

variable "alert_webhook_receivers" {
  description = "List of webhook receivers for alerts"
  type = list(object({
    name        = string
    service_uri = string
  }))
  default = []
}

# AGIC Configuration
variable "create_agic_role_assignments" {
  description = "Create AGIC role assignments via Terraform. Set to false if managed externally."
  type        = bool
  default     = true
}
