terraform {
  required_version = ">= 1.5"
  
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
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
    # resource_group_name  = "tfstate"
    # storage_account_name = "tfstatexxxxx"
    # container_name       = "tfstate"
    # key                  = "development.terraform.tfstate"
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
  environment = "development"
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
  base_name           = replace(local.resource_prefix, "-", "")
  # Truncate for key vault (max 24 chars total: kv-<base_name>v2)
  # Truncate for key vault (max 24 chars total: kv-<base_name>v2), preserve region suffix
  kv_base_name        = "${substr(local.base_name, 0, 18)}${substr(local.base_name, length(local.base_name) - 1, 1)}"
  # Truncate for storage account (max 24 chars total: st<base_name>v2), preserve region suffix
  storage_base_name   = "${substr(local.base_name, 0, 19)}${substr(local.base_name, length(local.base_name) - 1, 1)}"
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
  
  tags                 = local.common_tags
  public_dns_zone_name = var.public_dns_zone_name
}

# Key Vault
module "keyvault" {
  source = "../../modules/keyvault"
  
  key_vault_name                   = "kv-${local.kv_base_name}v2"
  location                        = local.location
  resource_group_name             = module.resource_group.name
  allowed_subnet_ids              = [module.networking.aks_subnet_id, module.networking.private_endpoints_subnet_id]
  enable_private_endpoint         = var.enable_private_endpoints
  private_endpoint_subnet_id      = module.networking.private_endpoints_subnet_id
  keyvault_private_dns_zone_id    = module.networking.keyvault_private_dns_zone_id
  log_analytics_workspace_id      = module.monitoring.log_analytics_workspace_id
  # Development settings - enable public access for Terraform deployment
  public_network_access_enabled   = true
  network_acls_default_action     = "Allow"
  
  # Development secrets
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
  
  storage_account_name           = "st${local.storage_base_name}v2"
  location                      = local.location
  resource_group_name           = module.resource_group.name
  account_tier                  = var.storage_account_tier
  replication_type              = var.storage_replication_type
  allowed_subnet_ids            = [module.networking.aks_subnet_id]
  enable_private_endpoint       = var.enable_private_endpoints
  private_endpoint_subnet_id    = module.networking.private_endpoints_subnet_id
  storage_private_dns_zone_id   = module.networking.storage_private_dns_zone_id
  log_analytics_workspace_id    = module.monitoring.log_analytics_workspace_id
  
  # Development lifecycle settings (shorter retention)
  lifecycle_cool_after_days     = 7
  lifecycle_archive_after_days  = 30
  lifecycle_delete_after_days   = 90
  
  # Development network settings - allow broader access for easier development
  network_rules_default_action  = "Allow"
  public_network_access_enabled = true
  
  tags = local.common_tags
  
  depends_on = [
    module.networking
  ]
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
  
  server_name                   = "psql-${local.resource_prefix}-v2"
  location                     = local.location
  resource_group_name          = module.resource_group.name
  delegated_subnet_id          = module.networking.database_subnet_id
  vnet_id                      = module.networking.vnet_id
  public_network_access_enabled = true
  postgresql_version           = var.postgresql_version
  sku_name                     = var.postgresql_sku_name
  storage_mb                   = var.postgresql_storage_mb
  backup_retention_days        = var.postgresql_backup_retention_days
  log_analytics_workspace_id   = module.monitoring.log_analytics_workspace_id
  
  # Development database settings
  additional_databases = ["hub_test"]
  
  tags = local.common_tags
  
  depends_on = [
    module.networking
  ]
}

# Monitoring
module "monitoring" {
  source = "../../modules/monitoring"
  
  log_analytics_workspace_name    = "log-${local.resource_prefix}-v2"
  application_insights_name       = "appi-${local.resource_prefix}-v2"
  location                       = local.location
  resource_group_name            = module.resource_group.name
  action_group_name              = "ag-${local.resource_prefix}-v2"
  action_group_short_name        = "hub-dev"
  resource_prefix                = local.resource_prefix
  
  # Development monitoring settings
  log_retention_days             = 30
  daily_quota_gb                 = 5
  enable_default_alerts          = false  # Disable alerts in development
  enable_grafana                 = false  # No Grafana needed in development
  
  email_receivers = var.alert_email_receivers
  
  tags = local.common_tags
}

# AKS Cluster
module "aks" {
  source = "../../modules/aks"
  
  cluster_name                = "aks-${local.resource_prefix}-v2"
  location                   = local.location
  resource_group_name        = module.resource_group.name
  dns_prefix                 = "hub-dev"
  kubernetes_version         = var.kubernetes_version
  subnet_id                  = module.networking.aks_subnet_id
  vnet_id                    = module.networking.vnet_id
  
  # Development cluster settings (smaller, cost-effective)
  node_count                 = var.aks_node_count
  vm_size                    = var.aks_vm_size
  min_node_count            = 1
  max_node_count            = 3
  availability_zones        = []  
  enable_auto_scaling       = true
  
  # Development environment
  environment               = local.environment
  create_worker_node_pool   = false  # No worker pool in development
  
  tags = local.common_tags
}

# Security (Application Gateway)
module "security" {
  source = "../../modules/security"
  
  application_gateway_name      = "appgw-${local.resource_prefix}-v2"
  location                     = local.location
  resource_group_name          = module.resource_group.name
  application_gateway_subnet_id = module.networking.application_gateway_subnet_id
  
  # Development security settings
  application_gateway_capacity  = 1  # Minimum capacity for cost savings
  availability_zones           = [] 
  enable_waf                   = true
  waf_mode                     = "Detection"  # Less restrictive for development
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

# Validation for GitHub Runner configuration
locals {
  github_config_provided = var.github_config_url != "" && var.github_app_id != "" && var.github_app_installation_id != "" && var.github_app_private_key != ""
}

# GitHub Actions Runner Controller (ARC)
module "github_runner" {
  count  = var.enable_github_runners ? 1 : 0
  source = "../../modules/github_runner"

  # Validate that GitHub configuration is provided when runners are enabled
  lifecycle {
    precondition {
      condition     = !var.enable_github_runners || local.github_config_provided
      error_message = <<-EOT
        GitHub runner configuration is incomplete. When enable_github_runners = true, you must provide:
        - github_config_url (e.g., "https://github.com/your-org")
        - github_app_id (numeric string)
        - github_app_installation_id (numeric string)
        - github_app_private_key (PEM format)
        
        Please check your terraform.tfvars file and ensure all GitHub App values are set.
      EOT
    }
  }

  # GitHub App Configuration
  github_config_url           = var.github_config_url
  github_app_id              = var.github_app_id
  github_app_installation_id = var.github_app_installation_id
  github_app_private_key     = var.github_app_private_key

  # Runner Configuration
  runner_scale_set_name = var.runner_scale_set_name
  min_runners          = var.runner_min_replicas
  max_runners          = var.runner_max_replicas
  container_mode       = var.runner_container_mode
  runner_labels        = var.runner_labels

  # Development-specific settings
  controller_namespace = "arc-systems"
  runners_namespace   = "arc-runners"

  depends_on = [
    module.aks,
    data.azurerm_kubernetes_cluster.cluster
  ]
}

resource "azurerm_role_assignment" "aks_storage_blob_data_contributor" {
  scope                = module.storage.storage_account_id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = module.aks.cluster_identity_principal_id
}
