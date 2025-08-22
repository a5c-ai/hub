Started analysis of failed workflow run 17155438322.

Plan:
- Investigate Terraform Apply failure in Infrastructure Deployment.
- Root cause appears to be invalid AutoscalingRunnerSet patch due to missing volumeClaimTemplate.spec.
- Implement fix in terraform/modules/github_runner to always provide valid PVC spec in kubernetes containerMode to satisfy CRD validation.
- Verify Terraform validates locally.
- Open PR linking to failing run.

