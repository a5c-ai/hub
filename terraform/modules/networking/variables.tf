variable "vnet_name" {
  description = "The name of the virtual network"
  type        = string
}

variable "address_space" {
  description = "The address space of the virtual network"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
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

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}