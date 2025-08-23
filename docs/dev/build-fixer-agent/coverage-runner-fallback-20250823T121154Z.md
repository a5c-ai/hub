# Build Fix: Coverage Report runner fallback

## Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17155937196 (Coverage Report)
- Branch: main
- Head commit: 93a95a5266460526a672acee13bebe6169487e1a
- Symptom: job `coverage` cancelled immediately; no logs available
- Likely cause: self-hosted runner label `hub-dev-runners` unavailable or not provisioning; recent runner infra changes around Kubernetes mode may affect jobs without `job.container`

## Plan
1. Provide workflow copy under `.github_workflows/coverage-report.yml` using `ubuntu-latest`.
2. Keep steps identical; only change `runs-on` and add notes.
3. Open draft PR for SRE to migrate the workflow file to `.github/workflows/`.

## Verification
- Syntax validated locally.
- Script `scripts/test.sh` is CI-friendly on Ubuntu runners (uses apt-get and docker as needed; E2E disabled in coverage workflow).

## Links
- Original workflow: `.github/workflows/coverage-report.yml`
- New copy for review: `.github_workflows/coverage-report.yml`
