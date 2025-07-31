terraform {
  required_version = ">= 1.5"
  
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }

  backend "azurerm" {
    # Backend configuration will be provided during terraform init
    # Example:
    # resource_group_name  = "tfstate-prod"
    # storage_account_name = "tfstateprodxxxxx"
    # container_name       = "tfstate"
    # key                  = "production.terraform.tfstate"
  }
}

provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy    = false
      recover_soft_deleted_key_vaults = true
    }
    resource_group {
      prevent_deletion_if_contains_resources = true
    }
  }
}

locals {
  environment = "production"
  location    = var.location
  
  common_tags = {
    Environment = local.environment
    Project     = "hub"
    ManagedBy   = "terraform"
    Owner       = var.owner
    CostCenter  = var.cost_center
    Backup      = "required"
  }

  # Naming convention: {service}-hub-{environment}-{region}
  resource_prefix = "hub-${local.environment}-${replace(lower(local.location), " ", "")}"
  
  # Base name without hyphens for services with length constraints
  base_name = replace(local.resource_prefix, "-", "")
}

# Resource Group
module "resource_group" {
  source = "../../modules/resource_group"
  
  name     = "rg-${local.resource_prefix}"
  location = local.location
  tags     = local.common_tags
}

# Networking
module "networking" {
  source = "../../modules/networking"
  
  vnet_name                        = "vnet-${local.resource_prefix}"
  location                        = local.location
  resource_group_name             = module.resource_group.name
  address_space                   = var.vnet_address_space
  aks_subnet_cidr                 = var.aks_subnet_cidr
  database_subnet_cidr            = var.database_subnet_cidr
  private_endpoints_subnet_cidr   = var.private_endpoints_subnet_cidr
  appgw_subnet_cidr              = var.appgw_subnet_cidr
  admin_source_address_prefix     = var.admin_source_address_prefix
  
  tags = local.common_tags
}

# Key Vault
module "keyvault" {
  source = "../../modules/keyvault"
  
  key_vault_name                   = "kv-${replace(local.resource_prefix, "-", "")}"
  location                        = local.location
  resource_group_name             = module.resource_group.name
  allowed_subnet_ids              = [module.networking.aks_subnet_id, module.networking.private_endpoints_subnet_id]
  enable_private_endpoint         = true
  private_endpoint_subnet_id      = module.networking.private_endpoints_subnet_id
  keyvault_private_dns_zone_id    = module.networking.keyvault_private_dns_zone_id
  log_analytics_workspace_id      = module.monitoring.log_analytics_workspace_id
  public_network_access_enabled   = true
  purge_protection_enabled        = true
  soft_delete_retention_days      = 90
  
  # Production secrets
  secrets = {
    "database-password" = {
      value        = module.postgresql.admin_password
      content_type = "password"
      tags         = { "Environment" = "production" }
    }
    "storage-connection-string" = {
      value        = module.storage.primary_connection_string
      content_type = "connection-string"
      tags         = { "Environment" = "production" }
    }
  }
  
  # Production encryption keys
  keys = {
    "hub-encryption-key" = {
      key_type = "RSA"
      key_size = 4096
      key_opts = ["decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"]
      rotation_policy = {
        expire_after         = "P2Y"
        notify_before_expiry = "P30D"
        time_before_expiry   = "P30D"
      }
      tags = { "Purpose" = "encryption" }
    }
  }
  
  tags = local.common_tags
}

# Storage Account
module "storage" {
  source = "../../modules/storage"
  
  storage_account_name           = "st${replace(local.resource_prefix, "-", "")}"
  location                      = local.location
  resource_group_name           = module.resource_group.name
  account_tier                  = var.storage_account_tier
  replication_type              = var.storage_replication_type
  allowed_subnet_ids            = [module.networking.aks_subnet_id]
  enable_private_endpoint       = true
  private_endpoint_subnet_id    = module.networking.private_endpoints_subnet_id
  storage_private_dns_zone_id   = module.networking.storage_private_dns_zone_id
  public_network_access_enabled = true
  log_analytics_workspace_id    = module.monitoring.log_analytics_workspace_id
  
  # Production lifecycle settings
  lifecycle_cool_after_days     = var.lifecycle_cool_after_days
  lifecycle_archive_after_days  = var.lifecycle_archive_after_days
  lifecycle_delete_after_days   = var.lifecycle_delete_after_days
  
  # Enhanced backup retention for production
  backup_archive_after_days     = 7
  backup_delete_after_days      = 2555  # 7 years
  
  tags = local.common_tags
}

# Container Registry
module "container_registry" {
  source              = "../../modules/container_registry"

  registry_name       = "acr${local.base_name}"
  resource_group_name = module.resource_group.name
  location            = local.location
  tags                = local.common_tags
}

# PostgreSQL
module "postgresql" {
  source = "../../modules/postgresql"
  
  server_name                   = "psql-${local.resource_prefix}"
  location                     = "West US 3"
  resource_group_name          = module.resource_group.name
  delegated_subnet_id          = module.networking.database_subnet_id
  vnet_id                      = module.networking.vnet_id
  postgresql_version           = var.postgresql_version
  sku_name                     = var.postgresql_sku_name
  storage_mb                   = var.postgresql_storage_mb
  backup_retention_days        = var.postgresql_backup_retention_days
  geo_redundant_backup_enabled = var.postgresql_geo_redundant_backup_enabled
  high_availability_mode       = var.postgresql_high_availability_mode
  standby_availability_zone    = var.postgresql_standby_availability_zone
  log_analytics_workspace_id   = module.monitoring.log_analytics_workspace_id
  
  # Production maintenance window (Sunday 2 AM)
  maintenance_window = {
    day_of_week  = 0
    start_hour   = 2
    start_minute = 0
  }
  
  tags = local.common_tags
}

# Monitoring
module "monitoring" {
  source = "../../modules/monitoring"
  
  log_analytics_workspace_name    = "log-${local.resource_prefix}"
  application_insights_name       = "appi-${local.resource_prefix}"
  location                       = local.location
  resource_group_name            = module.resource_group.name
  action_group_name              = "ag-${local.resource_prefix}"
  action_group_short_name        = "hub-prod"
  resource_prefix                = local.resource_prefix
  
  # Production monitoring settings
  log_retention_days             = var.log_retention_days
  daily_quota_gb                 = var.log_analytics_daily_quota_gb
  application_insights_retention_days = var.application_insights_retention_days
  enable_default_alerts          = true
  enable_grafana                 = var.enable_grafana
  grafana_name                   = var.enable_grafana ? "grafana-${local.resource_prefix}" : null
  enable_security_center         = true
  
  # Production alerting
  alert_scopes = [module.aks.cluster_id]
  cpu_threshold    = 70
  memory_threshold = 75
  disk_threshold   = 80
  
  email_receivers    = var.alert_email_receivers
  sms_receivers      = var.alert_sms_receivers
  webhook_receivers  = var.alert_webhook_receivers
  
  tags = local.common_tags
}

# AKS Cluster
module "aks" {
  source = "../../modules/aks"
  
  cluster_name                = "aks-${local.resource_prefix}"
  location                   = local.location
  resource_group_name        = module.resource_group.name
  dns_prefix                 = "hub-prod"
  kubernetes_version         = var.kubernetes_version
  subnet_id                  = module.networking.aks_subnet_id
  vnet_id                    = module.networking.vnet_id
  
  # Production cluster settings (high availability)
  node_count                 = var.aks_node_count
  vm_size                    = var.aks_vm_size
  min_node_count            = var.aks_min_node_count
  max_node_count            = var.aks_max_node_count
  availability_zones        = var.availability_zones
  enable_auto_scaling       = true
  
  # Worker node pool for production workloads
  create_worker_node_pool   = var.create_worker_node_pool
  worker_vm_size           = var.worker_vm_size
  worker_node_count        = var.worker_node_count
  worker_min_node_count    = var.worker_min_node_count
  worker_max_node_count    = var.worker_max_node_count
  
  environment               = local.environment
  log_retention_days        = var.log_retention_days
  
  tags = local.common_tags
}

# Security (Application Gateway)
module "security" {
  source = "../../modules/security"
  
  application_gateway_name      = "appgw-${local.resource_prefix}"
  location                     = local.location
  resource_group_name          = module.resource_group.name
  application_gateway_subnet_id = module.networking.application_gateway_subnet_id
  
  # Production security settings
  application_gateway_capacity  = var.application_gateway_capacity
  availability_zones           = var.availability_zones
  enable_waf                   = true
  waf_mode                     = "Prevention"
  waf_rule_set_version         = "3.2"
  waf_file_upload_limit_mb     = var.waf_file_upload_limit_mb
  waf_max_request_body_size_kb = var.waf_max_request_body_size_kb
  waf_rate_limit_threshold     = var.waf_rate_limit_threshold
  
  ssl_certificate_data         = var.ssl_certificate_data
  ssl_certificate_password     = var.ssl_certificate_password
  
  log_analytics_workspace_id   = module.monitoring.log_analytics_workspace_id
  
  tags = local.common_tags
}

# Role Assignments
resource "azurerm_role_assignment" "aks_keyvault_secrets_user" {
  scope                = module.keyvault.key_vault_id
  role_definition_name = "Key Vault Secrets User"
  principal_id         = module.aks.cluster_identity_principal_id

}

# GitHub Runner Controller and RunnerDeployment
data "azurerm_kubernetes_cluster" "cluster" {
  name                = module.aks.cluster_name
  resource_group_name = module.resource_group.name
}

provider "kubernetes" {
  host                   = try(data.azurerm_kubernetes_cluster.cluster.kube_config[0].host, "")
  client_certificate     = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].client_certificate), "")
  client_key             = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].client_key), "")
  cluster_ca_certificate = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].cluster_ca_certificate), "")
}

provider "helm" {
  kubernetes {
    host                   = try(data.azurerm_kubernetes_cluster.cluster.kube_config[0].host, "")
    client_certificate     = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].client_certificate), "")
    client_key             = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].client_key), "")
    cluster_ca_certificate = try(base64decode(data.azurerm_kubernetes_cluster.cluster.kube_config[0].cluster_ca_certificate), "")
  }
}

module "github_runner" {
  source                 = "../../modules/github_runner"
  github_token           = var.github_token
  github_owner           = var.github_owner
  github_repository      = var.github_repository
  runner_deployment_name = var.runner_deployment_name
  runner_replicas        = var.runner_replicas
  runner_labels          = var.runner_labels
}

resource "azurerm_role_assignment" "aks_storage_blob_data_contributor" {
  scope                = module.storage.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = module.aks.cluster_identity_principal_id
}

# Backup and Disaster Recovery (example with additional storage for backups)
module "backup_storage" {
  source = "../../modules/storage"
  
  storage_account_name           = "stbak${replace(local.resource_prefix, "-", "")}"
  location                      = var.backup_location
  resource_group_name           = module.resource_group.name
  account_tier                  = "Standard"
  replication_type              = "GRS"
  allowed_subnet_ids            = [module.networking.aks_subnet_id]
  enable_private_endpoint       = true
  private_endpoint_subnet_id    = module.networking.private_endpoints_subnet_id
  storage_private_dns_zone_id   = module.networking.storage_private_dns_zone_id
  log_analytics_workspace_id    = module.monitoring.log_analytics_workspace_id
  
  # Backup-specific containers
  additional_containers = ["database-backups", "system-backups"]
  
  # Long-term retention for backups
  backup_archive_after_days     = 1
  backup_delete_after_days      = 2555  # 7 years
  
  tags = merge(local.common_tags, {
    Purpose = "backup"
  })
}
