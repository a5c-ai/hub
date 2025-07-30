variable "server_name" {
  description = "The name of the PostgreSQL server"
  type        = string
}

variable "location" {
  description = "The Azure location where the PostgreSQL server should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
}

variable "postgresql_version" {
  description = "The version of PostgreSQL to use"
  type        = string
  default     = "15"
}

variable "delegated_subnet_id" {
  description = "The ID of the delegated subnet for the PostgreSQL server"
  type        = string
}

variable "vnet_id" {
  description = "The ID of the virtual network"
  type        = string
}

variable "admin_username" {
  description = "The administrator username for the PostgreSQL server"
  type        = string
  default     = "hub_admin"
}

variable "admin_password" {
  description = "The administrator password for the PostgreSQL server"
  type        = string
  default     = null
  sensitive   = true
}

variable "storage_mb" {
  description = "The storage size in MB for the PostgreSQL server"
  type        = number
  default     = 32768
}

variable "sku_name" {
  description = "The SKU name for the PostgreSQL server"
  type        = string
  default     = "GP_Standard_D2s_v3"
}

variable "backup_retention_days" {
  description = "The number of days to retain backups"
  type        = number
  default     = 7
}

variable "geo_redundant_backup_enabled" {
  description = "Enable geo-redundant backups"
  type        = bool
  default     = false
}

variable "high_availability_mode" {
  description = "The high availability mode for the PostgreSQL server"
  type        = string
  default     = null
  validation {
    condition = var.high_availability_mode == null || try(contains(["ZoneRedundant", "SameZone"], var.high_availability_mode), false)
    error_message = "High availability mode must be either 'ZoneRedundant' or 'SameZone'."
  }
}

variable "standby_availability_zone" {
  description = "The availability zone for the standby server"
  type        = string
  default     = null
}

variable "maintenance_window" {
  description = "The maintenance window configuration"
  type = object({
    day_of_week  = number
    start_hour   = number
    start_minute = number
  })
  default = null
}

variable "database_name" {
  description = "The name of the main database"
  type        = string
  default     = "hub"
}

variable "additional_databases" {
  description = "List of additional databases to create"
  type        = list(string)
  default     = []
}

variable "database_collation" {
  description = "The collation for the database"
  type        = string
  default     = "en_US.utf8"
}

variable "database_charset" {
  description = "The charset for the database"
  type        = string
  default     = "utf8"
}

variable "log_statement" {
  description = "PostgreSQL log_statement setting"
  type        = string
  default     = "ddl"
}

variable "log_min_duration_statement" {
  description = "PostgreSQL log_min_duration_statement setting in milliseconds"
  type        = string
  default     = "1000"
}

variable "shared_preload_libraries" {
  description = "PostgreSQL shared_preload_libraries setting"
  type        = string
  default     = "pg_stat_statements"
}

variable "max_connections" {
  description = "PostgreSQL max_connections setting"
  type        = string
  default     = "200"
}

variable "log_analytics_workspace_id" {
  description = "The ID of the Log Analytics workspace for diagnostic settings"
  type        = string
  default     = null
}

variable "public_network_access_enabled" {
  description = "Enable public network access to the PostgreSQL server. When true, the server will not use virtual network delegation."
  type        = bool
  default     = false
}

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}
