Title: Fix Terraform azurerm provider version alignment

Date: 2025-08-22

Context:
- Workflow run 17153733274 (Infrastructure Deployment) failed on Terraform Init due to conflicting provider constraints for hashicorp/azurerm (~> 3.0 and ~> 4.0).
- Root environment files specify azurerm ~> 4.0, modules still pin ~> 3.0.

Plan:
- Update azurerm required_providers version in Terraform modules to "~> 4.0" to align with environment.
- Validate terraform init locally in development environment.
- Open PR linking the failed run and verification steps.
