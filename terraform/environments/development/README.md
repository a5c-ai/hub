# Terraform Development Environment

This directory contains Terraform configuration for the development environment.

## Prerequisites

Copy `terraform.tfvars.example` to `terraform.tfvars` and customize the values as needed.

## Import Existing PostgreSQL Server

If you have an existing PostgreSQL flexible server that was created outside of Terraform,
you can import it into the Terraform state to avoid name conflicts:

```bash
terraform import 'module.postgresql.azurerm_postgresql_flexible_server.main' \
  '/subscriptions/<SUBSCRIPTION_ID>/resourceGroups/<RESOURCE_GROUP_NAME>/providers/Microsoft.DBForPostgreSQL/flexibleServers/${var.server_name}'
```

Replace `<SUBSCRIPTION_ID>` and `<RESOURCE_GROUP_NAME>` with the appropriate values for your environment.
