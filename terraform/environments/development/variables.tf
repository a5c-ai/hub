variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
  default     = "East US"
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
  default     = "1.29"
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

# Monitoring Variables
variable "alert_email_receivers" {
  description = "List of email receivers for alerts"
  type = list(object({
    name          = string
    email_address = string
  }))
  default = []
}
