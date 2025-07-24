variable "application_gateway_name" {
  description = "The name of the Application Gateway"
  type        = string
}

variable "location" {
  description = "The Azure location where resources should be created"
  type        = string
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
}

variable "application_gateway_subnet_id" {
  description = "The subnet ID for the Application Gateway"
  type        = string
}

variable "availability_zones" {
  description = "The availability zones for the Application Gateway"
  type        = list(string)
  default     = ["1", "2", "3"]
}

variable "application_gateway_sku_name" {
  description = "The SKU name for the Application Gateway"
  type        = string
  default     = "WAF_v2"
}

variable "application_gateway_sku_tier" {
  description = "The SKU tier for the Application Gateway"
  type        = string
  default     = "WAF_v2"
}

variable "application_gateway_capacity" {
  description = "The capacity for the Application Gateway"
  type        = number
  default     = 2
}

variable "enable_waf" {
  description = "Enable Web Application Firewall"
  type        = bool
  default     = true
}

variable "waf_mode" {
  description = "The mode for the Web Application Firewall"
  type        = string
  default     = "Prevention"
  validation {
    condition     = contains(["Detection", "Prevention"], var.waf_mode)
    error_message = "WAF mode must be either 'Detection' or 'Prevention'."
  }
}

variable "waf_rule_set_version" {
  description = "The rule set version for the Web Application Firewall"
  type        = string
  default     = "3.2"
}

variable "waf_file_upload_limit_mb" {
  description = "The file upload limit in MB for the Web Application Firewall"
  type        = number
  default     = 100
}

variable "waf_max_request_body_size_kb" {
  description = "The maximum request body size in KB for the Web Application Firewall"
  type        = number
  default     = 128
}

variable "waf_rate_limit_threshold" {
  description = "The rate limit threshold for the Web Application Firewall"
  type        = number
  default     = 100
}

variable "waf_exclusions" {
  description = "List of WAF exclusions"
  type = list(object({
    match_variable          = string
    selector_match_operator = string
    selector                = string
  }))
  default = []
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

variable "health_probe_path" {
  description = "The path for health probes"
  type        = string
  default     = "/health"
}

variable "health_probe_host" {
  description = "The host for health probes"
  type        = string
  default     = "127.0.0.1"
}

variable "backend_address_pools" {
  description = "List of backend address pools"
  type = list(object({
    name         = string
    ip_addresses = list(string)
    fqdns        = list(string)
  }))
  default = []
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