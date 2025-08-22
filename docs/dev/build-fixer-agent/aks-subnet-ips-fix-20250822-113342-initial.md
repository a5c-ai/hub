Title: Fix AKS apply failure due to InsufficientSubnetSize by reducing max_node_count in development

Context: Workflow run 17154014625 for Infrastructure Deployment failed on Terraform Apply with InsufficientSubnetSize for AKS default node pool. Error indicates pre-allocated IPs 290 > available 251 in AKS subnet 10.0.1.0/24. Current module sets max_node_count=10, causing high IP preallocation.

Plan:
- Reduce AKS module max_node_count for development environment from 10 to 8 to fit /24 subnet capacity.
- Keep autoscaling enabled (min=1).
- Validate Terraform syntax locally.
- Open PR with details and link to failed run.


Results:
- Updated terraform/environments/development/main.tf to set max_node_count=8
- Created draft PR #733 with context and links to failing run
- Rationale: fit IP pre-allocation within /24 subnet and avoid InsufficientSubnetSize

Next Steps:
- After merge, rerun Infrastructure Deployment and verify AKS Apply succeeds
- Optionally further tune upgrade_settings.max_surge if needed

## Results
- Updated terraform/environments/development/main.tf: max_node_count set to 6
- Pushed PR https://github.com/a5c-ai/hub/pull/734 and requested review
- Local terraform validate unavailable; CI should confirm fix

