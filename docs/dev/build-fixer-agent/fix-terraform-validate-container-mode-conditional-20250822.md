Title: Fix terraform validate failure in github_runner module (containerMode conditional types)

Date: 2025-08-22

Context:
- Failed workflow run: https://github.com/a5c-ai/hub/actions/runs/17160935234
- Job: Infrastructure Deployment / terraform
- Failing step: Terraform Validate

Error summary from logs:
- Error: Inconsistent conditional result types in modules/github_runner/main.tf
- Location: resource "helm_release" "arc_runner_set" -> values.containerMode conditional
- Message: The 'true' value includes object attribute "kubernetesMode", which is absent in the 'false' value.

Root cause:
- Terraform 1.5 requires consistent types in conditional expressions. The previous implementation used a conditional that returned a populated object for kubernetes mode and an empty object `{}` otherwise. This caused a type mismatch during `terraform validate`.

Changes made:
- Updated terraform/modules/github_runner/main.tf:
  - Replaced the `merge(..., condition ? { ... } : {})` pattern for `containerMode` with an object that sets attributes using attribute-level conditionals, returning `null` when not applicable to maintain consistent typing.
  - Adjusted `template.spec` construction to use attribute-level conditionals for `nodeSelector` and `containers`, setting them to `null` when not provided to avoid similar type inconsistencies.

Verification steps executed:
- Retrieved logs for the failing run using `gh run view` and captured the error.
- Local static review of the updated expressions for consistent conditional result types.
- Note: Terraform binary is not available in this environment, so `terraform validate` could not be executed locally. The change should satisfy Terraform's type checker by avoiding mixed object schemas across conditionals.

Next steps:
- Open PR and let CI run the Infrastructure Deployment workflow to verify `terraform validate` passes.

By: build-fixer-agent (https://app.a5c.ai/a5c/agents/development/build-fixer-agent)

