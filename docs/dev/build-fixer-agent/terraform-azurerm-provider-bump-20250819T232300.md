Context

- Trigger: workflow_run (Infrastructure Deployment) failed on main
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17006233328
- Root cause: Terraform provider constraint conflict for hashicorp/azurerm
  - Error: "no available releases match the given constraints ~> 3.0, ~> 4.0"
  - Cause: root environment requires azurerm "~> 4.0" while multiple modules still require "~> 3.0"

Plan

- Update all Terraform modules that pin azurerm to "~> 3.0" and bump them to "~> 4.0"
- Keep other providers as-is
- Do not modify workflow files
- Verify via static checks and rely on CI to run terraform init during the workflow

Changes

- Updated versions.tf in modules: aks, keyvault, monitoring, networking, postgresql, resource_group, storage, security to require azurerm "~> 4.0"

Notes

- This aligns modules with the recent bump to v4 at the environment level to support AKS upgrade_settings
- No functional changes to resource definitions were made in this pass

Verification

- Local terraform execution is not available in this environment; CI Infrastructure Deployment workflow will run terraform init and validate the provider resolution

