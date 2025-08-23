# Build Fix: Coverage workflow runner fallback

## Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17155438313
- Workflow: `.github/workflows/coverage-report.yml`
- Error: `The job was not acquired by Runner of type scale-set even after multiple attempts`
- Current `runs-on`: `hub-dev-runners` (ARC scale-set label)

## Root Cause
Runner scale-set did not acquire the job, causing the job to fail before any steps executed. This is an infrastructure label/availability issue, not a project code/test failure.

## Plan
- Add a safe fallback for `runs-on` using a repository/org variable so jobs can run on GitHub-hosted runners when the ARC scale-set is unavailable.
- Do not edit `.github/workflows` directly; add a mirrored workflow under `.github_workflows/` per repo guidelines, for maintainers to move.

## Proposed Change
- Copy `coverage-report.yml` to `.github_workflows/coverage-report.yml` and set:
  ```yaml
  runs-on: ${{ vars.A5C_RUNNER_COVERAGE || 'ubuntu-latest' }}
  ```
  Maintainers can set `A5C_RUNNER_COVERAGE=hub-dev-runners` at repo/org level to keep using ARC; otherwise it falls back to GitHub-hosted.

## Verification
- Local dry run: Validate that test script exists and can run without E2E.
- The actual runner acquisition fix validates once the maintainer moves the workflow file into `.github/workflows/` or updates the existing one accordingly.

## Links
- Original workflow: `.github/workflows/coverage-report.yml`
- Failing run: https://github.com/a5c-ai/hub/actions/runs/17155438313

By: [build-fixer-agent](https://app.a5c.ai/a5c/agents/development/build-fixer-agent)

## Local Verification Results
- Go unit tests: PASSED locally (`go test ./...`)
- Frontend unit tests: PASSED locally (`npm ci && npm run test:ci`)
- E2E tests: not executed as not relevant to coverage infra fix

