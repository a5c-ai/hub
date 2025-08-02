 # GitHub Actions Runner Controller (ARC) Module

This module installs the official GitHub Actions Runner Controller on a Kubernetes cluster using Helm. It creates organization-level self-hosted runners that can be used across all repositories in your GitHub organization.

## Features

- **Organization-level runners**: Not tied to specific repositories
- **Modern GitHub ARC**: Uses the official GitHub Actions Runner Controller
- **GitHub App authentication**: Secure authentication using GitHub Apps
- **Auto-scaling**: Runners scale up/down based on demand
- **Container modes**: Supports both Docker-in-Docker (dind) and Kubernetes modes
- **Customizable**: Configurable runner images, labels, and scaling parameters

## Architecture

This module deploys two main components:

1. **ARC Controller** (`gha-runner-scale-set-controller`): Manages the overall system in the `arc-systems` namespace
2. **Runner Scale Set** (`gha-runner-scale-set`): Creates and manages actual runners in the `arc-runners` namespace

## Requirements

- Kubernetes cluster (EKS, GKE, AKS, or any other)
- `helm` and `kubernetes` Terraform providers configured
- GitHub organization with admin permissions
- GitHub App created and installed

## GitHub App Setup

Before using this module, you need to create a GitHub App:

### 1. Create a GitHub App

Go to your GitHub organization settings:

```text
Settings → Developer settings → GitHub Apps → New GitHub App
```

Fill in the form:

- **GitHub App name**: `ARC-Runners-[YOUR-ORG]`
- **Homepage URL**: `https://github.com/actions/actions-runner-controller`
- **Webhook**: Uncheck "Active" (no webhook needed)

### 2. Set Permissions

**Repository permissions:**

- Administration: Read & write
- Metadata: Read

**Organization permissions:**

- Self-hosted runners: Read & write

### 3. Install the App

After creation:

- Note the **App ID**
- Generate and download a **Private Key**
- Click "Install App" and install it on your organization
- Note the **Installation ID** from the URL

## Usage

```hcl
module "github_runner" {
  source = "../../modules/github_runner"

  # GitHub Configuration
  github_config_url           = "https://github.com/my-org"
  github_app_id              = "123456"
  github_app_installation_id = "12345678"
  github_app_private_key     = file("path/to/private-key.pem")

  # Runner Configuration
  runner_scale_set_name = "my-org-runners"
  min_runners          = 0
  max_runners          = 10
  container_mode       = "dind"

  # Optional: Custom runner image
  runner_image = "ghcr.io/my-org/custom-runner:latest"
  
  # Optional: Runner labels
  runner_labels = ["gpu", "large"]
}
```

## Using the Runners in Workflows

After deployment, use the runner scale set name in your GitHub Actions workflows:

```yaml
name: My Workflow
on: [push]

jobs:
  build:
    runs-on: my-org-runners  # Use the runner_scale_set_name
    steps:
      - uses: actions/checkout@v4
      - run: echo "Running on self-hosted ARC runner!"
```

## Variables

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| `github_config_url` | GitHub organization URL | `string` | n/a | yes |
| `github_app_id` | GitHub App ID | `string` | n/a | yes |
| `github_app_installation_id` | GitHub App Installation ID | `string` | n/a | yes |
| `github_app_private_key` | GitHub App private key (PEM format) | `string` | n/a | yes |
| `runner_scale_set_name` | Name for the runner scale set (used in runs-on) | `string` | `"arc-runner-set"` | no |
| `controller_namespace` | Namespace for the ARC controller | `string` | `"arc-systems"` | no |
| `runners_namespace` | Namespace for the runner scale sets | `string` | `"arc-runners"` | no |
| `min_runners` | Minimum number of runners | `number` | `0` | no |
| `max_runners` | Maximum number of runners | `number` | `10` | no |
| `container_mode` | Container mode (dind or kubernetes) | `string` | `"dind"` | no |
| `runner_image` | Custom runner image | `string` | `null` | no |
| `runner_labels` | Additional labels for runners | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| `controller_release_name` | Name of the ARC controller Helm release |
| `runner_set_release_name` | Name of the runner scale set Helm release |
| `runner_scale_set_name` | Name to use in `runs-on` in workflows |
| `controller_namespace` | Namespace where controller is deployed |
| `runners_namespace` | Namespace where runners are deployed |

## Migration from Old Version

If you're migrating from the old summerwind/actions-runner-controller:

1. **Remove old resources** first (they use different APIs)
2. **Create GitHub App** instead of using personal access tokens
3. **Update module call** with new variables
4. **Update workflows** to use the new runner scale set name

## Troubleshooting

### Check controller status

```bash
kubectl get pods -n arc-systems
kubectl logs -n arc-systems deployment/arc-controller-gha-rs-controller
```

### Check runner scale set status

```bash
kubectl get pods -n arc-runners
kubectl get runnerscaleset -n arc-runners
```

### Check GitHub integration

- Verify GitHub App permissions
- Check the Installation ID in the GitHub organization settings
- Ensure the private key is correctly formatted
