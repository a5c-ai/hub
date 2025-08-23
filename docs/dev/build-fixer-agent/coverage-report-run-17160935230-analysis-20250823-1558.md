# Build Fixer: Coverage Report run failure (17160935230)

By: build-fixer-agent (https://app.a5c.ai/a5c/agents/development/build-fixer-agent)

## Context
- Repo: a5c-ai/hub
- Workflow: `.github/workflows/coverage-report.yml` (name: Coverage Report)
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17160935230
- Head commit: `70743a7e97281026e7a88a89bf1d4563173b8a81`
- Trigger: push to `main`

## Observation
- `gh run view` shows a single job `coverage` with conclusion `cancelled` and no steps executed. The run’s overall conclusion is `failure`.
- Workflow uses a self-hosted runner label: `runs-on: hub-dev-runners`.
- No job logs are available, suggesting the job never started (likely no eligible runner was available or it was auto-cancelled after queueing).

## Likely Root Cause (Category 2: infra/workflow)
Self-hosted runner label `hub-dev-runners` had no available/online capacity for this workflow at the time, causing the job to be cancelled without executing steps.

## Proposed Fix
Provide a GitHub-hosted fallback for the Coverage job to avoid being blocked on self-hosted capacity. Minimal change: run on `ubuntu-latest` (sufficient for Go + Node coverage run, E2E disabled by `--no-e2e`).

Per repo rules, do not edit files under `.github/workflows/` directly. Instead, add a proposed workflow under `.github_workflows/coverage-report.yml` with `runs-on: ubuntu-latest` for maintainers to move.

## Plan
1. Add proposed workflow copy under `.github_workflows/coverage-report.yml` using `ubuntu-latest`.
2. Validate locally that `scripts/test.sh --no-e2e --coverage` is compatible with GitHub-hosted environment assumptions.
3. Open a draft PR documenting the issue, linking the failed run, and requesting review.

## Notes
- The test script installs `libsqlite3-dev` and `gcc` only when `CI=true`; GitHub-hosted `ubuntu-latest` supports `sudo apt-get`, so this path should work in CI. E2E remains disabled in coverage workflow.

