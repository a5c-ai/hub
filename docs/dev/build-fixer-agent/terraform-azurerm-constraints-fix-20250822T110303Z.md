Hi tmuskal

## Build failure analysis started: Terraform provider version conflict (azurerm)

### Description
A workflow run failed on main: Infrastructure Deployment (https://github.com/a5c-ai/hub/actions/runs/17153472693). The failure occurs during `terraform init` in `terraform/environments/development` with the following error:

Error: Failed to query available provider packages
Could not retrieve the list of available versions for provider hashicorp/azurerm: no available releases match the given constraints ~> 3.0, ~> 4.0

Root cause: multiple Terraform modules pin `azurerm` to `~> 3.0` while the environment `main.tf` requires `~> 4.0`, producing conflicting combined constraints `~> 3.0, ~> 4.0`.

### Plan
- Relax azurerm version in all affected modules to a range compatible with both v3 and v4: `>= 3.0, < 5.0`.
- Keep environment definitions at `~> 4.0` to standardize on v4.
- Open a PR with these changes and link the failing run.

### Progress
- Branch created: fix/terraform-azurerm-version-constraint
- Next: update module version constraints and push.

By: build-fixer-agent (agent+build-fixer-agent@a5c.ai) - https://a5c.ai/agents/build-fixer-agent
