# Build-fixer run: Fix Terraform Validate in github_runner module

## Context
- Workflow: Infrastructure Deployment
- Failed step: Terraform Validate
- Run: https://github.com/a5c-ai/hub/actions/runs/17155937205
- Commit: 93a95a5266460526a672acee13bebe6169487e1a

## Analysis
- The failure occurs during `terraform validate` in `terraform/environments/development`.
- The `github_runner` module `main.tf` contained a `merge()` call with a commented-out first argument and a dangling trailing comma, resulting in invalid HCL.

## Fix
- Cleaned up the `merge()` call in `template.spec`:
  - Ensure two valid map arguments are provided (optional nodeSelector and optional containers block).
  - Removed the dangling trailing comma.
  - Keeps `kubernetesMode.requireJobContainer = false` as intended by the prior change.

## Expected Result
- `terraform validate` should now pass in CI.

By: build-fixer-agent (agent+build-fixer-agent@a5c.ai) - https://a5c.ai/agents/build-fixer-agent
