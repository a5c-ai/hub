# Build Fixer: Coverage Report Runner Fix Proposal

## Context
- Failed workflow: Coverage Report
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17155611400
- Current runner label: `hub-dev-runners`
- Symptom: Job cancelled; logs not retrievable via API (likely no available self-hosted runner)

## Plan
- Propose switching this workflow to `ubuntu-latest` to avoid dependency on the self-hosted pool for coverage-only jobs
- Keep steps identical (Go + Node setup; run `./scripts/test.sh --no-e2e`; upload coverage)
- Open an infrastructure issue to restore `hub-dev-runners` capacity

## Rationale
Coverage and unit tests do not require self-hosted capabilities. Using GitHub-hosted runners increases reliability. Self-hosted pool can remain for integration jobs needing services or custom environment.

## Verification
- CI should be able to execute the workflow on `ubuntu-latest`
- Script already installs `libsqlite3-dev` and `gcc` only when `CI=true` which is compatible with ubuntu-latest

## Links
- Workflow file: `.github/workflows/coverage-report.yml`
- Related scripts: `scripts/test.sh`, `scripts/build.sh`

