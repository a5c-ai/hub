variable "key_vault_name" {
  description = "The name of the Key Vault"
  type        = string
}

variable "location" {
  description = "The Azure location where the Key Vault should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
}

variable "sku_name" {
  description = "The SKU name for the Key Vault"
  type        = string
  default     = "standard"
  validation {
    condition     = contains(["standard", "premium"], var.sku_name)
    error_message = "SKU name must be either 'standard' or 'premium'."
  }
}

variable "soft_delete_retention_days" {
  description = "The number of days to retain soft-deleted items"
  type        = number
  default     = 7
  validation {
    condition     = var.soft_delete_retention_days >= 7 && var.soft_delete_retention_days <= 90
    error_message = "Soft delete retention days must be between 7 and 90."
  }
}

variable "purge_protection_enabled" {
  description = "Enable purge protection for the Key Vault"
  type        = bool
  default     = true
}

variable "enabled_for_deployment" {
  description = "Enable the Key Vault for VM deployment"
  type        = bool
  default     = false
}

variable "enabled_for_disk_encryption" {
  description = "Enable the Key Vault for disk encryption"
  type        = bool
  default     = true
}

variable "enabled_for_template_deployment" {
  description = "Enable the Key Vault for template deployment"
  type        = bool
  default     = false
}

variable "enable_rbac_authorization" {
  description = "Enable RBAC authorization for the Key Vault"
  type        = bool
  default     = true
}

variable "public_network_access_enabled" {
  description = "Enable public network access to the Key Vault"
  type        = bool
  default     = false
}

variable "network_acls_default_action" {
  description = "The default action for network ACLs"
  type        = string
  default     = "Deny"
  validation {
    condition     = contains(["Allow", "Deny"], var.network_acls_default_action)
    error_message = "Network ACLs default action must be either 'Allow' or 'Deny'."
  }
}

variable "network_acls_bypass" {
  description = "The bypass setting for network ACLs"
  type        = string
  default     = "AzureServices"
  validation {
    condition     = contains(["AzureServices", "None"], var.network_acls_bypass)
    error_message = "Network ACLs bypass must be either 'AzureServices' or 'None'."
  }
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

variable "admin_object_id" {
  description = "The object ID of the Key Vault administrator"
  type        = string
  default     = null
}

variable "aks_principal_id" {
  description = "The principal ID of the AKS cluster identity"
  type        = string
  default     = null
}

variable "secrets" {
  description = "Map of secrets to create in the Key Vault"
  type = map(object({
    value           = string
    content_type    = optional(string)
    expiration_date = optional(string)
    tags            = optional(map(string), {})
  }))
  default = {}
}

variable "keys" {
  description = "Map of keys to create in the Key Vault"
  type = map(object({
    key_type = string
    key_size = optional(number)
    key_opts = list(string)
    rotation_policy = optional(object({
      expire_after         = string
      notify_before_expiry = string
      time_before_expiry   = string
    }))
    tags = optional(map(string), {})
  }))
  default = {}
}

variable "enable_private_endpoint" {
  description = "Enable private endpoint for the Key Vault"
  type        = bool
  default     = true
}

variable "private_endpoint_subnet_id" {
  description = "The subnet ID for the private endpoint"
  type        = string
  default     = null
}

variable "keyvault_private_dns_zone_id" {
  description = "The ID of the Key Vault private DNS zone"
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