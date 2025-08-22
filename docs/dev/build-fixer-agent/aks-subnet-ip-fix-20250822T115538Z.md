# Fix dev AKS InsufficientSubnetSize during rotation

## Context
- Failed workflow run: https://github.com/a5c-ai/hub/actions/runs/17154500216
- Error from Terraform Apply on azurerm_kubernetes_cluster: InsufficientSubnetSize. Pre-allocated IPs 290 > available 251 in 10.0.1.0/24.

## Plan
- Reduce AKS default node pool autoscaler max_node_count further in terraform/environments/development/main.tf.
- Keep min_node_count=1, enable_auto_scaling=true, availability_zones=[] for cost.
- Validate Terraform configuration locally.

## Rationale
Lowering max_node_count reduces IP pre-allocation, keeping total under the /24 subnet capacity during temporary pool rotation.

## Verification Steps
- terraform init -backend=false
- terraform validate
- (CI) Re-run Infrastructure Deployment to confirm apply succeeds.

