terraform {
  required_version = ">= 1.5"
  
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0.0, < 4.0.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.9"
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
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }

  backend "azurerm" {
    # Backend configuration will be provided during terraform init
    # Example:
    # resource_group_name  = "tfstate-staging"
    # storage_account_name = "tfstatestagingxxxxx"
    # container_name       = "tfstate"
    # key                  = "staging.terraform.tfstate"
  }
}

provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
    resource_group {
      prevent_deletion_if_contains_resources = false
    }
  }
}

locals {
  environment = "staging"
  location    = var.location
  
  common_tags = {
    Environment = local.environment
    Project     = "hub"
    ManagedBy   = "terraform"
    Owner       = var.owner
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
  enable_private_endpoint         = var.enable_private_endpoints
  private_endpoint_subnet_id      = module.networking.private_endpoints_subnet_id
  keyvault_private_dns_zone_id    = module.networking.keyvault_private_dns_zone_id
  log_analytics_workspace_id      = module.monitoring.log_analytics_workspace_id
  public_network_access_enabled   = true
  
  # Staging secrets
  secrets = {
    "database-password" = {
      value        = module.postgresql.admin_password
      content_type = "password"
      tags         = {}
    }
    "storage-connection-string" = {
      value        = module.storage.primary_connection_string
      content_type = "connection-string"
      tags         = {}
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
  enable_private_endpoint       = var.enable_private_endpoints
  private_endpoint_subnet_id    = module.networking.private_endpoints_subnet_id
  storage_private_dns_zone_id   = module.networking.storage_private_dns_zone_id
  public_network_access_enabled = true
  log_analytics_workspace_id    = module.monitoring.log_analytics_workspace_id
  
  # Staging lifecycle settings (medium retention)
  lifecycle_cool_after_days     = 14
  lifecycle_archive_after_days  = 60
  lifecycle_delete_after_days   = 180
  
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
  high_availability_mode       = var.postgresql_high_availability_mode
  log_analytics_workspace_id   = module.monitoring.log_analytics_workspace_id
  
  # Staging database settings
  additional_databases = ["hub_staging_test"]
  
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
  action_group_short_name        = "hub-stg"
  resource_prefix                = local.resource_prefix
  
  # Staging monitoring settings
  log_retention_days             = 60
  daily_quota_gb                 = 10
  application_insights_retention_days = 90
  enable_default_alerts          = true
  enable_grafana                 = var.enable_grafana
  grafana_name                   = var.enable_grafana ? "grafana-${local.resource_prefix}" : null
  
  # Staging alerting (less sensitive than production)
  alert_scopes = [module.aks.cluster_id]
  cpu_threshold    = 85
  memory_threshold = 85
  disk_threshold   = 85
  
  email_receivers = var.alert_email_receivers
  
  tags = local.common_tags
}

# AKS Cluster
module "aks" {
  source = "../../modules/aks"
  
  cluster_name                = "aks-${local.resource_prefix}"
  location                   = local.location
  resource_group_name        = module.resource_group.name
  dns_prefix                 = "hub-staging"
  kubernetes_version         = var.kubernetes_version
  subnet_id                  = module.networking.aks_subnet_id
  vnet_id                    = module.networking.vnet_id
  
  # Staging cluster settings (production-like but smaller)
  node_count                 = var.aks_node_count
  vm_size                    = var.aks_vm_size
  min_node_count            = var.aks_min_node_count
  max_node_count            = var.aks_max_node_count
  availability_zones        = var.availability_zones
  enable_auto_scaling       = true
  
  # Worker node pool for staging workloads
  create_worker_node_pool   = var.create_worker_node_pool
  worker_vm_size           = var.worker_vm_size
  worker_node_count        = var.worker_node_count
  worker_min_node_count    = var.worker_min_node_count
  worker_max_node_count    = var.worker_max_node_count
  
  environment               = local.environment
  
  tags = local.common_tags
}

# Security (Application Gateway)
module "security" {
  source = "../../modules/security"
  
  application_gateway_name      = "appgw-${local.resource_prefix}"
  location                     = local.location
  resource_group_name          = module.resource_group.name
  application_gateway_subnet_id = module.networking.application_gateway_subnet_id
  
  # Staging security settings (production-like configuration)
  application_gateway_capacity  = var.application_gateway_capacity
  availability_zones           = var.availability_zones
  enable_waf                   = true
  waf_mode                     = "Detection"  # Less restrictive than production
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
