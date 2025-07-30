variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
  default     = "East US"
}

variable "owner" {
  description = "The owner of the resources"
  type        = string
  default     = "hub-staging-team"
}

variable "availability_zones" {
  description = "The availability zones to use"
  type        = list(string)
  default     = ["1", "2"]
}

# Networking Variables
variable "vnet_address_space" {
  description = "The address space of the virtual network"
  type        = list(string)
  default     = ["10.2.0.0/16"]
}

variable "aks_subnet_cidr" {
  description = "The CIDR block for the AKS subnet"
  type        = string
  default     = "10.2.1.0/24"
}

variable "database_subnet_cidr" {
  description = "The CIDR block for the database subnet"
  type        = string
  default     = "10.2.2.0/24"
}

variable "private_endpoints_subnet_cidr" {
  description = "The CIDR block for the private endpoints subnet"
  type        = string
  default     = "10.2.3.0/24"
}

variable "appgw_subnet_cidr" {
  description = "The CIDR block for the application gateway subnet"
  type        = string
  default     = "10.2.4.0/24"
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
  default     = "GP_Standard_D2s_v3"
}

variable "postgresql_storage_mb" {
  description = "The storage size in MB for the PostgreSQL server"
  type        = number
  default     = 131072  # 128GB
}

variable "postgresql_backup_retention_days" {
  description = "The number of days to retain backups"
  type        = number
  default     = 14
}

variable "postgresql_high_availability_mode" {
  description = "The high availability mode for the PostgreSQL server"
  type        = string
  default     = "SameZone"
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
  default     = "Standard_D2s_v5"
}

variable "aks_min_node_count" {
  description = "The minimum number of nodes in the AKS cluster"
  type        = number
  default     = 2
}

variable "aks_max_node_count" {
  description = "The maximum number of nodes in the AKS cluster"
  type        = number
  default     = 6
}

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

# Application Gateway Variables
variable "application_gateway_capacity" {
  description = "The capacity for the Application Gateway"
  type        = number
  default     = 2
}

# Monitoring Variables
variable "enable_grafana" {
  description = "Enable Azure Managed Grafana"
  type        = bool
  default     = false
}

variable "alert_email_receivers" {
  description = "List of email receivers for alerts"
  type = list(object({
    name          = string
    email_address = string
  }))
  default = []
}
