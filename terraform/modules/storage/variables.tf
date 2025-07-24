variable "storage_account_name" {
  description = "The name of the storage account"
  type        = string
}

variable "location" {
  description = "The Azure location where the storage account should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
}

variable "account_tier" {
  description = "The tier of the storage account"
  type        = string
  default     = "Standard"
}

variable "replication_type" {
  description = "The replication type of the storage account"
  type        = string
  default     = "LRS"
}

variable "access_tier" {
  description = "The access tier of the storage account"
  type        = string
  default     = "Hot"
}

variable "shared_access_key_enabled" {
  description = "Enable shared access key for the storage account"
  type        = bool
  default     = true
}

variable "public_network_access_enabled" {
  description = "Enable public network access to the storage account"
  type        = bool
  default     = false
}

variable "versioning_enabled" {
  description = "Enable versioning for blobs"
  type        = bool
  default     = true
}

variable "change_feed_enabled" {
  description = "Enable change feed for blobs"
  type        = bool
  default     = true
}

variable "last_access_time_enabled" {
  description = "Enable last access time tracking"
  type        = bool
  default     = true
}

variable "blob_retention_days" {
  description = "The number of days to retain deleted blobs"
  type        = number
  default     = 7
}

variable "container_retention_days" {
  description = "The number of days to retain deleted containers"
  type        = number
  default     = 7
}

variable "cors_rules" {
  description = "CORS rules for the storage account"
  type = list(object({
    allowed_origins    = list(string)
    allowed_methods    = list(string)
    allowed_headers    = list(string)
    exposed_headers    = list(string)
    max_age_in_seconds = number
  }))
  default = []
}

variable "network_rules_default_action" {
  description = "The default action for network rules"
  type        = string
  default     = "Deny"
}

variable "network_rules_bypass" {
  description = "The bypass rules for network rules"
  type        = list(string)
  default     = ["AzureServices"]
}

variable "allowed_ip_addresses" {
  description = "List of allowed IP addresses"
  type        = list(string)
  default     = []
}

variable "allowed_subnet_ids" {
  description = "List of allowed subnet IDs"
  type        = list(string)
  default     = []
}

variable "additional_containers" {
  description = "List of additional containers to create"
  type        = list(string)
  default     = []
}

variable "enable_lifecycle_policy" {
  description = "Enable lifecycle management policy"
  type        = bool
  default     = true
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
  default     = 365
}

variable "lifecycle_snapshot_delete_after_days" {
  description = "Days after which to delete blob snapshots"
  type        = number
  default     = 30
}

variable "lifecycle_version_delete_after_days" {
  description = "Days after which to delete blob versions"
  type        = number
  default     = 30
}

variable "backup_archive_after_days" {
  description = "Days after which to archive backup blobs"
  type        = number
  default     = 7
}

variable "backup_delete_after_days" {
  description = "Days after which to delete backup blobs"
  type        = number
  default     = 90
}

variable "enable_private_endpoint" {
  description = "Enable private endpoint for the storage account"
  type        = bool
  default     = true
}

variable "private_endpoint_subnet_id" {
  description = "The subnet ID for the private endpoint"
  type        = string
  default     = null
}

variable "storage_private_dns_zone_id" {
  description = "The ID of the storage private DNS zone"
  type        = string
  default     = null
}

variable "log_analytics_workspace_id" {
  description = "The ID of the Log Analytics workspace for diagnostic settings"
  type        = string
  default     = null
}

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}