# Terraform Security Module

This module provisions an Azure Application Gateway with Web Application Firewall (WAF) policy,
including support for configurable rate limiting rules.

## Usage Example

```hcl
module "security" {
  source = "../../modules/security"

  application_gateway_name      = "appgw-demo"
  resource_group_name           = "rg-demo"
  location                      = "eastus"
  application_gateway_subnet_id = azurerm_subnet.appgw.id

  enable_waf                    = true
  waf_mode                      = "Prevention"
  waf_rule_set_version          = "3.2"
  waf_file_upload_limit_mb      = 100
  waf_max_request_body_size_kb  = 128

  # Rate limiting configuration
  waf_rate_limit_threshold           = 100
  waf_rate_limit_duration_in_minutes = 1
  waf_rate_limit_match_variable      = "RemoteAddr"
  waf_rate_limit_selector_match_operator = "IPMatch"
  waf_rate_limit_selector            = ""
  waf_rate_limit_match_values        = ["*"]
  waf_rate_limit_group_by_keys       = ["RemoteAddr"]

  ssl_certificate_data         = var.ssl_certificate_data
  ssl_certificate_password     = var.ssl_certificate_password
  log_analytics_workspace_id   = var.log_analytics_workspace_id
  tags                         = var.tags
}
```
