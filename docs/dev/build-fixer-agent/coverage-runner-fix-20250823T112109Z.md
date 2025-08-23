Hi team

## Coverage Runner Fix - Switch to ubuntu-latest

### Description
The Coverage Report workflow run failed with overall conclusion "failure" while its sole job "coverage" shows conclusion "cancelled" and no steps executed. The job targeted self-hosted label `hub-dev-runners` and appears to have been cancelled immediately (likely due to runner unavailability or concurrency preemption). Link to failing run: https://github.com/a5c-ai/hub/actions/runs/17154825154

To improve reliability and avoid dependency on self-hosted runner availability for this non-privileged coverage job, propose switching it to GitHub-hosted `ubuntu-latest`.

### Plan
- Copy `.github/workflows/coverage-report.yml` to `.github_workflows/coverage-report.yml`
- Change `runs-on` from `hub-dev-runners` to `ubuntu-latest`
- Leave all steps intact
- Open PR with context and links
- After merge (or moved by maintainer to `.github/workflows/`), the workflow should run consistently

### Notes
- No functionality is disabled; only the runner is changed.
- The test script may install `libsqlite3-dev` and `gcc` when `CI=true`, both are available on `ubuntu-latest` via apt.

By: build-fixer-agent (https://app.a5c.ai/a5c/agents/development/build-fixer-agent)

### Results
- Created branch and PR: https://github.com/a5c-ai/hub/pull/752
- Logged analysis on the failing commit.
