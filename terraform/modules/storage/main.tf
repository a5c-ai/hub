resource "azurerm_storage_account" "main" {
  name                     = var.storage_account_name
  resource_group_name      = var.resource_group_name
  location                = var.location
  account_tier            = var.account_tier
  account_replication_type = var.replication_type
  account_kind            = "StorageV2"
  access_tier             = var.access_tier

  min_tls_version                 = "TLS1_2"
  allow_nested_items_to_be_public = false
  shared_access_key_enabled       = var.shared_access_key_enabled
  public_network_access_enabled   = var.public_network_access_enabled

  blob_properties {
    versioning_enabled       = var.versioning_enabled
    change_feed_enabled      = var.change_feed_enabled
    last_access_time_enabled = var.last_access_time_enabled

    delete_retention_policy {
      days = var.blob_retention_days
    }

    container_delete_retention_policy {
      days = var.container_retention_days
    }

    dynamic "cors_rule" {
      for_each = var.cors_rules
      content {
        allowed_origins    = cors_rule.value.allowed_origins
        allowed_methods    = cors_rule.value.allowed_methods
        allowed_headers    = cors_rule.value.allowed_headers
        exposed_headers    = cors_rule.value.exposed_headers
        max_age_in_seconds = cors_rule.value.max_age_in_seconds
      }
    }
  }

  network_rules {
    default_action             = var.network_rules_default_action
    bypass                     = var.network_rules_bypass
    ip_rules                   = var.allowed_ip_addresses
    virtual_network_subnet_ids = var.allowed_subnet_ids
  }

  tags = var.tags
}

resource "azurerm_storage_container" "repositories" {
  name                  = "repositories"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"

  metadata = {
    purpose = "git-repositories"
  }
}

resource "azurerm_storage_container" "artifacts" {
  name                  = "artifacts"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"

  metadata = {
    purpose = "build-artifacts"
  }
}

resource "azurerm_storage_container" "packages" {
  name                  = "packages"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"

  metadata = {
    purpose = "package-registry"
  }
}

resource "azurerm_storage_container" "backups" {
  name                  = "backups"
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"

  metadata = {
    purpose = "system-backups"
  }
}

resource "azurerm_storage_container" "additional" {
  for_each              = toset(var.additional_containers)
  name                  = each.value
  storage_account_id    = azurerm_storage_account.main.id
  container_access_type = "private"
}

resource "azurerm_storage_management_policy" "main" {
  count              = var.enable_lifecycle_policy ? 1 : 0
  storage_account_id = azurerm_storage_account.main.id

  rule {
    name    = "default-lifecycle"
    enabled = true
    filters {
      prefix_match = ["repositories/", "artifacts/", "packages/"]
      blob_types   = ["blockBlob"]
    }
    actions {
      base_blob {
        tier_to_cool_after_days_since_modification_greater_than    = var.lifecycle_cool_after_days
        tier_to_archive_after_days_since_modification_greater_than = var.lifecycle_archive_after_days
        delete_after_days_since_modification_greater_than          = var.lifecycle_delete_after_days
      }
      snapshot {
        delete_after_days_since_creation_greater_than = var.lifecycle_snapshot_delete_after_days
      }
      version {
        delete_after_days_since_creation = var.lifecycle_version_delete_after_days
      }
    }
  }

  rule {
    name    = "backup-lifecycle"
    enabled = true
    filters {
      prefix_match = ["backups/"]
      blob_types   = ["blockBlob"]
    }
    actions {
      base_blob {
        tier_to_archive_after_days_since_modification_greater_than = var.backup_archive_after_days
        delete_after_days_since_modification_greater_than          = var.backup_delete_after_days
      }
    }
  }
}

resource "azurerm_private_endpoint" "storage" {
  count               = var.enable_private_endpoint ? 1 : 0
  name                = "${var.storage_account_name}-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id

  private_service_connection {
    name                           = "${var.storage_account_name}-psc"
    private_connection_resource_id = azurerm_storage_account.main.id
    subresource_names              = ["blob"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "default"
    private_dns_zone_ids = [var.storage_private_dns_zone_id]
  }

  tags = var.tags
}

resource "azurerm_monitor_diagnostic_setting" "storage" {
  name                       = "${var.storage_account_name}-diagnostics"
  target_resource_id         = "${azurerm_storage_account.main.id}/blobServices/default"
  log_analytics_workspace_id = var.log_analytics_workspace_id

  enabled_log {
    category = "StorageRead"
  }

  enabled_log {
    category = "StorageWrite"
  }

  enabled_log {
    category = "StorageDelete"
  }

  enabled_metric {
    category = "AllMetrics"
  }
}
