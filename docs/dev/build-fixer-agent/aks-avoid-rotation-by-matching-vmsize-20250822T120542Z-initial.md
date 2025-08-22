Hi tmuskal

## Build Fix: AKS InsufficientSubnetSize on Terraform Apply

### Description
Infrastructure Deployment run failed on Terraform Apply when modifying azurerm_kubernetes_cluster default node pool. Error:
- InsufficientSubnetSize: Pre-allocated IPs 290 exceeds IPs available 251 in Subnet 10.0.1.0/24 during creation of temporary Agent Pool "defaulttemp".
- Run: https://github.com/a5c-ai/hub/actions/runs/17154619452
- Commit: 5e0096e

Root cause: The plan simultaneously reduced max_count (10 -> 6) and changed vm_size (Standard_D2s_v5 -> Standard_D4s_v5). Changing vm_size triggers a rotation that creates a temporary node pool sized from the current configuration (max_count=10) before the new max_count takes effect, causing preallocation of ~290 IPs in a /24 subnet.

### Plan
- Stage the change: first apply the lower autoscaler max_count without changing vm_size to avoid rotation.
- Concretely: set aks_vm_size to the current cluster size (Standard_D2s_v5) in terraform/environments/development/terraform.tfvars so Apply only changes max_count (10 -> 6), not vm_size, avoiding temporary pool creation exceeding subnet capacity.
- After this passes, we can consider increasing vm_size in a separate change if still desired.

### Progress
- Prepare fix in a small PR limited to development tfvars.
- Keep workflows unchanged.

By: build-fixer-agent (agent+build-fixer-agent@a5c.ai) - https://a5c.ai/agents/build-fixer-agent
