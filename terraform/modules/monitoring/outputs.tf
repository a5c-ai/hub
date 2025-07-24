output "log_analytics_workspace_id" {
  description = "The ID of the Log Analytics workspace"
  value       = azurerm_log_analytics_workspace.main.id
}

output "log_analytics_workspace_name" {
  description = "The name of the Log Analytics workspace"
  value       = azurerm_log_analytics_workspace.main.name
}

output "log_analytics_workspace_primary_shared_key" {
  description = "The primary shared key of the Log Analytics workspace"
  value       = azurerm_log_analytics_workspace.main.primary_shared_key
  sensitive   = true
}

output "log_analytics_workspace_secondary_shared_key" {
  description = "The secondary shared key of the Log Analytics workspace"
  value       = azurerm_log_analytics_workspace.main.secondary_shared_key
  sensitive   = true
}

output "application_insights_id" {
  description = "The ID of the Application Insights instance"
  value       = azurerm_application_insights.main.id
}

output "application_insights_name" {
  description = "The name of the Application Insights instance"
  value       = azurerm_application_insights.main.name
}

output "application_insights_instrumentation_key" {
  description = "The instrumentation key of the Application Insights instance"
  value       = azurerm_application_insights.main.instrumentation_key
  sensitive   = true
}

output "application_insights_connection_string" {
  description = "The connection string of the Application Insights instance"
  value       = azurerm_application_insights.main.connection_string
  sensitive   = true
}

output "action_group_id" {
  description = "The ID of the action group"
  value       = azurerm_monitor_action_group.main.id
}

output "action_group_name" {
  description = "The name of the action group"
  value       = azurerm_monitor_action_group.main.name
}

output "grafana_id" {
  description = "The ID of the Grafana instance"
  value       = var.enable_grafana ? azurerm_dashboard_grafana.main[0].id : null
}

output "grafana_endpoint" {
  description = "The endpoint of the Grafana instance"
  value       = var.enable_grafana ? azurerm_dashboard_grafana.main[0].endpoint : null
}

output "grafana_identity" {
  description = "The identity of the Grafana instance"
  value       = var.enable_grafana ? azurerm_dashboard_grafana.main[0].identity : null
}

output "data_collection_rule_id" {
  description = "The ID of the data collection rule"
  value       = var.enable_data_collection_rule ? azurerm_monitor_data_collection_rule.main[0].id : null
}