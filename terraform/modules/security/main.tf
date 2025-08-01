resource "azurerm_public_ip" "appgw" {
  name                = "${var.application_gateway_name}-pip"
  resource_group_name = var.resource_group_name
  location            = var.location
  allocation_method   = "Static"
  sku                 = "Standard"
  zones               = var.availability_zones

  tags = var.tags
}

resource "azurerm_web_application_firewall_policy" "main" {
  count               = var.enable_waf ? 1 : 0
  name                = "${var.application_gateway_name}-wafpolicy"
  resource_group_name = var.resource_group_name
  location            = var.location

  policy_settings {
    enabled                     = true
    mode                        = var.waf_mode
    request_body_check          = true
    file_upload_limit_in_mb     = var.waf_file_upload_limit_mb
    max_request_body_size_in_kb = var.waf_max_request_body_size_kb
  }

  managed_rules {
    managed_rule_set {
      type    = "OWASP"
      version = var.waf_rule_set_version
    }

    dynamic "exclusion" {
      for_each = var.waf_exclusions
      content {
        match_variable          = exclusion.value.match_variable
        selector_match_operator = exclusion.value.selector_match_operator
        selector                = exclusion.value.selector
      }
    }
  }

  # Rate limiting custom rules
  # Implements configurable rate limit thresholds and match conditions
  dynamic "custom_rules" {
    for_each = var.enable_waf && var.waf_rate_limit_threshold > 0 ? [1] : []
    content {
      name      = "ratelimit"
      priority  = 100
      rule_type = "RateLimitRule"
      action    = "Block"

      rate_limit_threshold  = var.waf_rate_limit_threshold
      group_by_user_session = true

      match_conditions {
        match_variables {
          variable_name = "RemoteAddr"
        }
        operator           = "IPMatch"
        negation_condition = false
        match_values       = ["*"]
      }
    }
  }


  tags = var.tags
}

resource "azurerm_application_gateway" "main" {
  name                = var.application_gateway_name
  resource_group_name = var.resource_group_name
  location            = var.location

  sku {
    name     = var.application_gateway_sku_name
    tier     = var.application_gateway_sku_tier
    capacity = var.application_gateway_capacity
  }

  zones = var.availability_zones

  gateway_ip_configuration {
    name      = "my-gateway-ip-configuration"
    subnet_id = var.application_gateway_subnet_id
  }

  frontend_port {
    name = "${var.application_gateway_name}-feport-80"
    port = 80
  }

  frontend_port {
    name = "${var.application_gateway_name}-feport-443"
    port = 443
  }

  frontend_ip_configuration {
    name                 = "${var.application_gateway_name}-feip"
    public_ip_address_id = azurerm_public_ip.appgw.id
  }

  backend_address_pool {
    name = "${var.application_gateway_name}-beap"
  }

  backend_http_settings {
    name                  = "${var.application_gateway_name}-be-htst"
    cookie_based_affinity = "Disabled"
    path                  = "/path1/"
    port                  = 80
    protocol              = "Http"
    request_timeout       = 60
    probe_name           = "${var.application_gateway_name}-probe"
  }

  backend_http_settings {
    name                  = "${var.application_gateway_name}-be-htst-https"
    cookie_based_affinity = "Disabled"
    path                  = "/"
    port                  = 443
    protocol              = "Https"
    request_timeout       = 60
    probe_name           = "${var.application_gateway_name}-probe-https"
  }

  http_listener {
    name                           = "${var.application_gateway_name}-httplstn"
    frontend_ip_configuration_name = "${var.application_gateway_name}-feip"
    frontend_port_name             = "${var.application_gateway_name}-feport-80"
    protocol                       = "Http"
  }

  dynamic "http_listener" {
    for_each = var.ssl_certificate_data != null ? [1] : []
    content {
      name                           = "${var.application_gateway_name}-httplstn-https"
      frontend_ip_configuration_name = "${var.application_gateway_name}-feip"
      frontend_port_name             = "${var.application_gateway_name}-feport-443"
      protocol                       = "Https"
      ssl_certificate_name           = "${var.application_gateway_name}-ssl-cert"
    }
  }

  request_routing_rule {
    name                       = "${var.application_gateway_name}-rqrt"
    rule_type                  = "Basic"
    http_listener_name         = "${var.application_gateway_name}-httplstn"
    backend_address_pool_name  = "${var.application_gateway_name}-beap"
    backend_http_settings_name = "${var.application_gateway_name}-be-htst"
    priority                   = 100
  }

  dynamic "request_routing_rule" {
    for_each = var.ssl_certificate_data != null ? [1] : []
    content {
      name                       = "${var.application_gateway_name}-rqrt-https"
      rule_type                  = "Basic"
      http_listener_name         = "${var.application_gateway_name}-httplstn-https"
      backend_address_pool_name  = "${var.application_gateway_name}-beap"
      backend_http_settings_name = "${var.application_gateway_name}-be-htst-https"
      priority                   = 200
    }
  }

  probe {
    name                = "${var.application_gateway_name}-probe"
    protocol            = "Http"
    path                = var.health_probe_path
    host                = var.health_probe_host
    interval            = 30
    timeout             = 30
    unhealthy_threshold = 3
  }

  probe {
    name                = "${var.application_gateway_name}-probe-https"
    protocol            = "Https"
    path                = var.health_probe_path
    host                = var.health_probe_host
    interval            = 30
    timeout             = 30
    unhealthy_threshold = 3
  }

  dynamic "ssl_certificate" {
    for_each = var.ssl_certificate_data != null ? [1] : []
    content {
      name     = "${var.application_gateway_name}-ssl-cert"
      data     = var.ssl_certificate_data
      password = var.ssl_certificate_password
    }
  }

  firewall_policy_id = var.enable_waf ? azurerm_web_application_firewall_policy.main[0].id : null

  tags = var.tags
}

resource "azurerm_network_security_group" "appgw" {
  name                = "${var.application_gateway_name}-nsg"
  location            = var.location
  resource_group_name = var.resource_group_name

  security_rule {
    name                       = "AllowGatewayManager"
    priority                   = 1000
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "*"
    source_port_range          = "*"
    destination_port_range     = "65200-65535"
    source_address_prefix      = "GatewayManager"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowHTTP"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "AllowHTTPS"
    priority                   = 1002
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  tags = var.tags
}

resource "azurerm_monitor_diagnostic_setting" "appgw" {
  name                       = "${var.application_gateway_name}-diagnostics"
  target_resource_id         = azurerm_application_gateway.main.id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "ApplicationGatewayAccessLog"
  }

  enabled_log {
    category = "ApplicationGatewayPerformanceLog"
  }

  enabled_log {
    category = "ApplicationGatewayFirewallLog"
  }

  metric {
    category = "AllMetrics"
    enabled  = true
  }
}
