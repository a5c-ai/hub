resource "azurerm_log_analytics_workspace" "main" {
  name                = var.log_analytics_workspace_name
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = var.log_analytics_sku
  retention_in_days   = var.log_retention_days
  daily_quota_gb      = var.daily_quota_gb

  tags = var.tags
}

resource "azurerm_application_insights" "main" {
  name                = var.application_insights_name
  location            = var.location
  resource_group_name = var.resource_group_name
  workspace_id        = azurerm_log_analytics_workspace.main.id
  application_type    = var.application_type
  retention_in_days   = var.application_insights_retention_days
  daily_data_cap_in_gb = var.application_insights_daily_data_cap

  tags = var.tags
}

resource "azurerm_log_analytics_solution" "container_insights" {
  solution_name         = "ContainerInsights"
  location              = var.location
  resource_group_name   = var.resource_group_name
  workspace_resource_id = azurerm_log_analytics_workspace.main.id
  workspace_name        = azurerm_log_analytics_workspace.main.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/ContainerInsights"
  }

  tags = var.tags
}

resource "azurerm_log_analytics_solution" "security_center_free" {
  count                 = var.enable_security_center ? 1 : 0
  solution_name         = "Security"
  location              = var.location
  resource_group_name   = var.resource_group_name
  workspace_resource_id = azurerm_log_analytics_workspace.main.id
  workspace_name        = azurerm_log_analytics_workspace.main.name

  plan {
    publisher = "Microsoft"
    product   = "OMSGallery/Security"
  }

  tags = var.tags
}

resource "azurerm_monitor_action_group" "main" {
  name                = var.action_group_name
  resource_group_name = var.resource_group_name
  short_name          = var.action_group_short_name

  dynamic "email_receiver" {
    for_each = var.email_receivers
    content {
      name          = email_receiver.value.name
      email_address = email_receiver.value.email_address
    }
  }

  dynamic "sms_receiver" {
    for_each = var.sms_receivers
    content {
      name         = sms_receiver.value.name
      country_code = sms_receiver.value.country_code
      phone_number = sms_receiver.value.phone_number
    }
  }

  dynamic "webhook_receiver" {
    for_each = var.webhook_receivers
    content {
      name        = webhook_receiver.value.name
      service_uri = webhook_receiver.value.service_uri
    }
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "cpu_usage" {
  count               = var.enable_default_alerts ? 1 : 0
  name                = "${var.resource_prefix}-cpu-usage-alert"
  resource_group_name = var.resource_group_name
  scopes              = var.alert_scopes
  description         = "Alert when CPU usage exceeds threshold"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"

  criteria {
    metric_namespace = "Microsoft.ContainerService/managedClusters"
    metric_name      = "node_cpu_usage_percentage"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = var.cpu_threshold
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "memory_usage" {
  count               = var.enable_default_alerts ? 1 : 0
  name                = "${var.resource_prefix}-memory-usage-alert"
  resource_group_name = var.resource_group_name
  scopes              = var.alert_scopes
  description         = "Alert when memory usage exceeds threshold"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"

  criteria {
    metric_namespace = "Microsoft.ContainerService/managedClusters"
    metric_name      = "node_memory_working_set_percentage"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = var.memory_threshold
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  tags = var.tags
}

resource "azurerm_monitor_metric_alert" "disk_usage" {
  count               = var.enable_default_alerts ? 1 : 0
  name                = "${var.resource_prefix}-disk-usage-alert"
  resource_group_name = var.resource_group_name
  scopes              = var.alert_scopes
  description         = "Alert when disk usage exceeds threshold"
  severity            = 2
  frequency           = "PT5M"
  window_size         = "PT15M"

  criteria {
    metric_namespace = "Microsoft.ContainerService/managedClusters"
    metric_name      = "node_disk_usage_percentage"
    aggregation      = "Average"
    operator         = "GreaterThan"
    threshold        = var.disk_threshold
  }

  action {
    action_group_id = azurerm_monitor_action_group.main.id
  }

  tags = var.tags
}

resource "azurerm_monitor_scheduled_query_rules_alert_v2" "pod_restart" {
  count               = var.enable_default_alerts ? 1 : 0
  name                = "${var.resource_prefix}-pod-restart-alert"
  resource_group_name = var.resource_group_name
  location            = var.location
  evaluation_frequency = "PT5M"
  window_duration      = "PT15M"
  scopes               = [azurerm_log_analytics_workspace.main.id]
  severity             = 3
  criteria {
    query                   = <<-QUERY
      KubePodInventory
      | where TimeGenerated > ago(15m)
      | where RestartCount > 5
      | summarize count() by Name
    QUERY
    time_aggregation_method = "Count"
    threshold               = 1
    operator                = "GreaterThan"
  }

  action {
    action_groups = [azurerm_monitor_action_group.main.id]
  }

  tags = var.tags
}

resource "azurerm_dashboard_grafana" "main" {
  count                             = var.enable_grafana ? 1 : 0
  name                              = var.grafana_name
  resource_group_name               = var.resource_group_name
  location                          = var.location
  api_key_enabled                   = var.grafana_api_key_enabled
  deterministic_outbound_ip_enabled = var.grafana_deterministic_outbound_ip_enabled
  public_network_access_enabled     = var.grafana_public_network_access_enabled
  grafana_major_version             = "10"

  identity {
    type = "SystemAssigned"
  }

  tags = var.tags
}

resource "azurerm_role_assignment" "grafana_monitoring_reader" {
  count                = var.enable_grafana ? 1 : 0
  scope                = "/subscriptions/${data.azurerm_client_config.current.subscription_id}"
  role_definition_name = "Monitoring Reader"
  principal_id         = azurerm_dashboard_grafana.main[0].identity[0].principal_id
}

data "azurerm_client_config" "current" {}

resource "azurerm_monitor_data_collection_rule" "main" {
  count               = var.enable_data_collection_rule ? 1 : 0
  name                = "${var.resource_prefix}-dcr"
  resource_group_name = var.resource_group_name
  location            = var.location

  destinations {
    log_analytics {
      workspace_resource_id = azurerm_log_analytics_workspace.main.id
      name                  = "destination-log"
    }
  }

  data_flow {
    streams      = ["Microsoft-ContainerLog", "Microsoft-ContainerLogV2"]
    destinations = ["destination-log"]
  }

  data_sources {
    syslog {
      facility_names = ["*"]
      log_levels     = ["*"]
      name           = "datasource-syslog"
      streams        = ["Microsoft-Syslog"]
    }

    performance_counter {
      name                          = "datasource-perfcounter"
      sampling_frequency_in_seconds = 60
      streams                       = ["Microsoft-Perf"]
      counter_specifiers            = [
        "\\Processor Information(_Total)\\% Processor Time",
        "\\Processor Information(_Total)\\% Privileged Time",
        "\\Processor Information(_Total)\\% User Time",
        "\\Memory\\Available Bytes",
        "\\Memory\\% Committed Bytes In Use",
        "\\Process(_Total)\\Working Set",
        "\\Process(_Total)\\Working Set - Private"
      ]
    }
  }

  tags = var.tags
}