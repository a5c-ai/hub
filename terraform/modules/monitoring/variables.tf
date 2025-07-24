variable "log_analytics_workspace_name" {
  description = "The name of the Log Analytics workspace"
  type        = string
}

variable "application_insights_name" {
  description = "The name of the Application Insights instance"
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

variable "log_analytics_sku" {
  description = "The SKU of the Log Analytics workspace"
  type        = string
  default     = "PerGB2018"
}

variable "log_retention_days" {
  description = "The number of days to retain logs"
  type        = number
  default     = 30
}

variable "daily_quota_gb" {
  description = "The daily quota for the Log Analytics workspace in GB"
  type        = number
  default     = -1
}

variable "application_type" {
  description = "The type of Application Insights to create"
  type        = string
  default     = "web"
}

variable "application_insights_retention_days" {
  description = "The retention period for Application Insights data"
  type        = number
  default     = 90
}

variable "application_insights_daily_data_cap" {
  description = "The daily data volume cap for Application Insights in GB"
  type        = number
  default     = 100
}

variable "enable_security_center" {
  description = "Enable Security Center solution"
  type        = bool
  default     = true
}

variable "action_group_name" {
  description = "The name of the action group"
  type        = string
}

variable "action_group_short_name" {
  description = "The short name of the action group"
  type        = string
}

variable "email_receivers" {
  description = "List of email receivers for alerts"
  type = list(object({
    name          = string
    email_address = string
  }))
  default = []
}

variable "sms_receivers" {
  description = "List of SMS receivers for alerts"
  type = list(object({
    name         = string
    country_code = string
    phone_number = string
  }))
  default = []
}

variable "webhook_receivers" {
  description = "List of webhook receivers for alerts"
  type = list(object({
    name        = string
    service_uri = string
  }))
  default = []
}

variable "enable_default_alerts" {
  description = "Enable default metric alerts"
  type        = bool
  default     = true
}

variable "resource_prefix" {
  description = "Prefix for alert resource names"
  type        = string
}

variable "alert_scopes" {
  description = "List of resource IDs to monitor"
  type        = list(string)
  default     = []
}

variable "cpu_threshold" {
  description = "CPU usage threshold for alerts"
  type        = number
  default     = 80
}

variable "memory_threshold" {
  description = "Memory usage threshold for alerts"
  type        = number
  default     = 80
}

variable "disk_threshold" {
  description = "Disk usage threshold for alerts"
  type        = number
  default     = 80
}

variable "enable_grafana" {
  description = "Enable Azure Managed Grafana"
  type        = bool
  default     = false
}

variable "grafana_name" {
  description = "The name of the Grafana instance"
  type        = string
  default     = null
}

variable "grafana_api_key_enabled" {
  description = "Enable API key authentication for Grafana"
  type        = bool
  default     = false
}

variable "grafana_deterministic_outbound_ip_enabled" {
  description = "Enable deterministic outbound IP for Grafana"
  type        = bool
  default     = false
}

variable "grafana_public_network_access_enabled" {
  description = "Enable public network access for Grafana"
  type        = bool
  default     = true
}

variable "enable_data_collection_rule" {
  description = "Enable data collection rule"
  type        = bool
  default     = true
}

variable "tags" {
  description = "A mapping of tags to assign to the resources"
  type        = map(string)
  default     = {}
}