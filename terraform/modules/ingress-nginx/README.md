# Ingress NGINX Terraform module for AKS

This module installs the ingress-nginx controller into an AKS cluster via Helm.

## Example Usage

```hcl
module "ingress_nginx" {
  source       = "../../modules/ingress-nginx"

  # Enable or disable installation
  enabled       = var.ingress_nginx_controller_enabled

  # Release and chart settings
  release_name  = var.ingress_nginx_controller_release_name
  chart_version = var.ingress_nginx_controller_chart_version

  # Kubernetes namespace for the controller
  namespace     = var.ingress_nginx_controller_namespace

  # Custom chart values (flat map of key/value overrides)
  values        = var.ingress_nginx_controller_values
}
```
