output "application_gateway_id" {
  description = "The ID of the Application Gateway"
  value       = azurerm_application_gateway.main.id
}

output "application_gateway_name" {
  description = "The name of the Application Gateway"
  value       = azurerm_application_gateway.main.name
}

output "public_ip_address" {
  description = "The public IP address of the Application Gateway"
  value       = azurerm_public_ip.appgw.ip_address
}

output "public_ip_fqdn" {
  description = "The FQDN of the public IP address"
  value       = azurerm_public_ip.appgw.fqdn
}

output "backend_address_pool_id" {
  description = "The ID of the backend address pool"
  value       = [for bap in azurerm_application_gateway.main.backend_address_pool : bap.id][0]
}

output "waf_policy_id" {
  description = "The ID of the WAF policy"
  value       = var.enable_waf ? azurerm_web_application_firewall_policy.main[0].id : null
}

output "network_security_group_id" {
  description = "The ID of the network security group"
  value       = azurerm_network_security_group.appgw.id
}
