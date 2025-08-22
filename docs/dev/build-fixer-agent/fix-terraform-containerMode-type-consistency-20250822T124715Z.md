Title: Fix Terraform containerMode type inconsistency in github_runner module

Summary
- Addressed Terraform validate failure in Infrastructure Deployment workflow.
- Error: true/false expressions must have consistent types due to containerMode map branching.

Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17155611384
- Commit: 719b9b8c66d62f7a5aa026d1a5731d890f5f6ebc
- File: terraform/modules/github_runner/main.tf

Plan
- Replace ternary that returns differently shaped objects with a single object having consistent keys.
- Always include kubernetesModeWorkVolumeClaim; use null when not applicable.
- Use local.volume_spec for DRY volume claim definition.

Change
- containerMode now:
  - type = var.container_mode
  - kubernetesModeWorkVolumeClaim = var.container_mode == "kubernetes" ? local.volume_spec : null

Verification
- Could not run terraform validate locally (terraform not available in container), but logic matches Terraform typing rules and should satisfy validation.
- Will rely on CI to validate.

By: build-fixer-agent (agent+build-fixer-agent@a5c.ai) - https://a5c.ai/agents/build-fixer-agent
