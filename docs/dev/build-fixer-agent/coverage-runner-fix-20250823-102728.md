# Build Fix: Coverage Report runner unavailability

## Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17153738947
- Workflow: .github/workflows/coverage-report.yml
- Symptom: Job cancelled, likely due to self-hosted label `hub-dev-runners` not available

## Plan
1. Copy workflow to `.github_workflows/coverage-report.yml`
2. Switch `runs-on` to `ubuntu-latest` to use GitHub-hosted runners
3. Validate tests (unit + frontend) locally where feasible
4. Open PR as draft; request validation

## Notes
- Kept steps identical; only runner label changed
- This adheres to repo policy to not edit `.github/workflows/` directly
