# Example terraform.tfvars file for development environment
# Copy this file to terraform.tfvars and customize the values

# Basic Configuration
location = "West US 3"
owner    = "hub-development-team"

# Networking Configuration
vnet_address_space              = ["10.0.0.0/16"]
aks_subnet_cidr                = "10.0.1.0/24"
database_subnet_cidr           = "10.0.2.0/24"
private_endpoints_subnet_cidr  = "10.0.3.0/24"
appgw_subnet_cidr             = "10.0.4.0/24"
admin_source_address_prefix    = "*"  # Allow all IPs for development (change to your IP for security)
enable_private_endpoints       = false  # Disable private endpoints for simpler development setup
public_dns_zone_name           = ""  # Optional: set if you have a DNS zone

# Storage Configuration
storage_account_tier      = "Standard"
storage_replication_type  = "LRS"

# PostgreSQL Configuration
postgresql_version                = "15"
postgresql_sku_name              = "B_Standard_B1ms"
postgresql_storage_mb            = 32768
postgresql_backup_retention_days = 7

# AKS Configuration
kubernetes_version = "1.30"
aks_node_count    = 2
# Note: Add worker node pool for GitHub runners (bigger and stronger nodes)
create_worker_node_pool = true
worker_vm_size         = "Standard_D4s_v5"
worker_node_count      = 2
worker_min_node_count  = 1
worker_max_node_count  = 4

# Monitoring Configuration
alert_email_receivers = [
  {
    name          = "dev-team"
    email_address = "dev@a5c.ai"  # Replace with your actual email
  }
]

# GitHub Actions Runner Controller (ARC) Configuration
enable_github_runners        = true
github_config_url           = "https://github.com/a5c-ai"  # Replace with your GitHub organization URL
github_auth_method          = "token"  # Use token authentication (simpler for CI/CD)
# github_token will be provided via environment variable TF_VAR_github_token
# Force secret recreation: 2025-08-02

# Runner Configuration
runner_scale_set_name   = "hub-dev-runners"
runner_min_replicas     = 0

# Ingress NGINX Controller health probe for Azure LB
ingress_nginx_controller_values = {
  "controller.service.annotations.service\\.beta\\.kubernetes\\.io/azure-load-balancer-health-probe-port"         = "10254"
  "controller.service.annotations.service\\.beta\\.kubernetes\\.io/azure-load-balancer-health-probe-request-path" = "/healthz"
}

runner_max_replicas     = 20
runner_container_mode   = "kubernetes"  # Use Kubernetes mode (more efficient than dind)
runner_labels          = ["development", "linux"]

# Custom Runner Image with Prerequisites
runner_image = "acrhubdevelopmentwestus3.azurecr.io/hub/github-runner:e9332fe"  # Fixed ACR name - uncomment when custom image is ready

# AGIC Configuration - Enable for Application Gateway management
create_agic_role_assignments = false
