data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "main" {
  name                       = var.key_vault_name
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = data.azurerm_client_config.current.tenant_id
  sku_name                   = var.sku_name
  soft_delete_retention_days = var.soft_delete_retention_days
  purge_protection_enabled   = var.purge_protection_enabled

  enabled_for_deployment          = var.enabled_for_deployment
  enabled_for_disk_encryption     = var.enabled_for_disk_encryption
  enabled_for_template_deployment = var.enabled_for_template_deployment
  enable_rbac_authorization       = var.enable_rbac_authorization

  public_network_access_enabled = var.public_network_access_enabled

  network_acls {
    default_action             = var.network_acls_default_action
    bypass                     = var.network_acls_bypass
    ip_rules                   = var.allowed_ip_addresses
    virtual_network_subnet_ids = var.allowed_subnet_ids
  }

  tags = var.tags
}

resource "azurerm_key_vault_access_policy" "admin" {
  count        = var.enable_rbac_authorization ? 0 : 1
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = var.admin_object_id != null ? var.admin_object_id : data.azurerm_client_config.current.object_id

  key_permissions = [
    "Backup", "Create", "Decrypt", "Delete", "Encrypt", "Get", "Import",
    "List", "Purge", "Recover", "Restore", "Sign", "UnwrapKey", "Update",
    "Verify", "WrapKey", "Release", "Rotate", "GetRotationPolicy", "SetRotationPolicy"
  ]

  secret_permissions = [
    "Backup", "Delete", "Get", "List", "Purge", "Recover", "Restore", "Set"
  ]

  certificate_permissions = [
    "Backup", "Create", "Delete", "DeleteIssuers", "Get", "GetIssuers",
    "Import", "List", "ListIssuers", "ManageContacts", "ManageIssuers",
    "Purge", "Recover", "Restore", "SetIssuers", "Update"
  ]
}

resource "azurerm_key_vault_access_policy" "aks" {
  count        = var.enable_rbac_authorization ? 0 : (var.aks_principal_id != null ? 1 : 0)
  key_vault_id = azurerm_key_vault.main.id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = var.aks_principal_id

  secret_permissions = [
    "Get", "List"
  ]

  certificate_permissions = [
    "Get", "List"
  ]
}

resource "azurerm_role_assignment" "keyvault_admin" {
  count                = var.enable_rbac_authorization ? 1 : 0
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Administrator"
  principal_id         = var.admin_object_id != null ? var.admin_object_id : data.azurerm_client_config.current.object_id
}

resource "azurerm_role_assignment" "keyvault_secrets_user" {
  count                = var.enable_rbac_authorization && var.aks_principal_id != null ? 1 : 0
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = var.aks_principal_id
}

resource "azurerm_key_vault_secret" "secrets" {
  for_each        = var.secrets
  name            = each.key
  value           = each.value.value
  key_vault_id    = azurerm_key_vault.main.id
  content_type    = each.value.content_type
  expiration_date = each.value.expiration_date

  tags = merge(var.tags, each.value.tags)

  depends_on = [
    azurerm_key_vault_access_policy.admin,
    azurerm_role_assignment.keyvault_admin
  ]
}

resource "azurerm_key_vault_key" "keys" {
  for_each     = var.keys
  name         = each.key
  key_vault_id = azurerm_key_vault.main.id
  key_type     = each.value.key_type
  key_size     = each.value.key_size
  key_opts     = each.value.key_opts

  dynamic "rotation_policy" {
    for_each = each.value.rotation_policy != null ? [each.value.rotation_policy] : []
    content {
      automatic {
        time_before_expiry = rotation_policy.value.time_before_expiry
      }
      expire_after         = rotation_policy.value.expire_after
      notify_before_expiry = rotation_policy.value.notify_before_expiry
    }
  }

  tags = merge(var.tags, each.value.tags)

  depends_on = [
    azurerm_key_vault_access_policy.admin,
    azurerm_role_assignment.keyvault_admin
  ]
}

resource "azurerm_private_endpoint" "keyvault" {
  count               = var.enable_private_endpoint ? 1 : 0
  name                = "${var.key_vault_name}-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id

  private_service_connection {
    name                           = "${var.key_vault_name}-psc"
    private_connection_resource_id = azurerm_key_vault.main.id
    subresource_names              = ["vault"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "default"
    private_dns_zone_ids = [var.keyvault_private_dns_zone_id]
  }

  tags = var.tags
}

resource "azurerm_monitor_diagnostic_setting" "keyvault" {
  count                      = var.log_analytics_workspace_id != null ? 1 : 0
  name                       = "${var.key_vault_name}-diagnostics"
  target_resource_id         = azurerm_key_vault.main.id
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "AuditEvent"
  }

  enabled_log {
    category = "AzurePolicyEvaluationDetails"
  }

  metric {
    category = "AllMetrics"
    enabled  = true
  }
}