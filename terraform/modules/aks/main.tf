data "azurerm_client_config" "current" {}

resource "azurerm_user_assigned_identity" "aks" {
  name                = "${var.cluster_name}-identity"
  resource_group_name = var.resource_group_name
  location            = var.location

  tags = var.tags
}

resource "azurerm_role_assignment" "aks_network_contributor" {
  scope                = var.vnet_id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_user_assigned_identity.aks.principal_id
}

resource "azurerm_log_analytics_workspace" "aks" {
  name                = "${var.cluster_name}-logs"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = var.tags
}

resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = var.location
  resource_group_name = var.resource_group_name
  dns_prefix          = var.dns_prefix
  kubernetes_version  = var.kubernetes_version

  default_node_pool {
    name                = "default"
    node_count          = var.node_count
    vm_size             = var.vm_size
    zones               = var.availability_zones
    enable_auto_scaling = var.enable_auto_scaling
    min_count           = var.enable_auto_scaling ? var.min_node_count : null
    max_count           = var.enable_auto_scaling ? var.max_node_count : null
    os_disk_size_gb    = var.os_disk_size_gb
    vnet_subnet_id     = var.subnet_id

    upgrade_settings {
      max_surge = var.max_surge
    }

    tags = var.tags
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aks.id]
  }

  network_profile {
    network_plugin     = "azure"
    network_policy     = "azure"
    service_cidr       = var.service_cidr
    dns_service_ip     = var.dns_service_ip
    load_balancer_sku  = "standard"
  }

  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.aks.id
  }

  azure_policy_enabled = true

  key_vault_secrets_provider {
    secret_rotation_enabled  = true
    secret_rotation_interval = "2m"
  }

  maintenance_window_auto_upgrade {
    frequency   = "Weekly"
    interval    = 1
    duration    = 4
    day_of_week = "Sunday"
    start_time  = "00:00"
    utc_offset  = "+00:00"
  }

  maintenance_window_node_os {
    frequency   = "Weekly"
    interval    = 1
    duration    = 4
    day_of_week = "Sunday"
    start_time  = "00:00"
    utc_offset  = "+00:00"
  }

  auto_scaler_profile {
    balance_similar_node_groups      = false
    expander                         = "random"
    max_graceful_termination_sec     = "600"
    max_node_provisioning_time       = "15m"
    max_unready_nodes                = 3
    max_unready_percentage           = 45
    new_pod_scale_up_delay           = "10s"
    scale_down_delay_after_add       = "10m"
    scale_down_delay_after_delete    = "10s"
    scale_down_delay_after_failure   = "3m"
    scan_interval                    = "10s"
    scale_down_unneeded              = "10m"
    scale_down_unready               = "20m"
    scale_down_utilization_threshold = "0.5"
    empty_bulk_delete_max            = "10"
    skip_nodes_with_local_storage    = true
    skip_nodes_with_system_pods      = true
  }

  tags = var.tags

  depends_on = [
    azurerm_role_assignment.aks_network_contributor,
    azurerm_log_analytics_workspace.aks
  ]
}

resource "azurerm_kubernetes_cluster_node_pool" "worker" {
  count                 = var.create_worker_node_pool ? 1 : 0
  name                  = "worker"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = var.worker_vm_size
  node_count            = var.worker_node_count
  zones                 = var.availability_zones

  min_count            = var.enable_auto_scaling ? var.worker_min_node_count : null
  max_count            = var.enable_auto_scaling ? var.worker_max_node_count : null
  os_disk_size_gb      = var.os_disk_size_gb
  vnet_subnet_id       = var.subnet_id

  node_labels = {
    "nodepool-type" = "worker"
    "environment"   = var.environment
  }

  node_taints = var.worker_node_taints

  upgrade_settings {
    max_surge = var.max_surge
  }

  tags = var.tags
}
