variable "registry_name" {
  description = "The name of the Azure Container Registry"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group for the registry"
  type        = string
}

variable "location" {
  description = "The Azure location where the registry should be created"
  type        = string
}

variable "sku" {
  description = "The SKU of the Azure Container Registry"
  type        = string
  default     = "Standard"
}

variable "admin_enabled" {
  description = "Enable the admin user on the registry"
  type        = bool
  default     = false
}

variable "tags" {
  description = "A mapping of tags to assign to the registry"
  type        = map(string)
  default     = {}
}
