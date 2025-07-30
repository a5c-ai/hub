 # GitHub Runner Controller Module

 This module installs the GitHub Actions Runner Controller on an existing AKS cluster using the Helm provider,
 and defines a RunnerDeployment resource to manage self-hosted runners for a specific GitHub repository.

 ## Requirements

 - The AKS cluster must already exist and be configured via the Kubernetes provider.
 - The `helm` and `kubernetes` Terraform providers must be configured in the root module.

 ## Usage

```hcl
module "github_runner" {
  source                 = "../../modules/github_runner"
  github_token           = var.github_token
  github_owner           = var.github_owner
  github_repository      = var.github_repository
  runner_deployment_name = var.runner_deployment_name
  runner_replicas        = var.runner_replicas
  runner_labels          = var.runner_labels
}
```

 Ensure that `TF_VAR_github_token`, `TF_VAR_github_owner`, and `TF_VAR_github_repository` are provided via environment variables
 or in a `terraform.tfvars` file.
