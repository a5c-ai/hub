# Build Fixer: Coverage Workflow Runner Fix

## Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17154014635 (conclusion: failure; job: cancelled)
- Workflow: .github/workflows/coverage-report.yml (job `coverage`)
- Runner label: `hub-dev-runners` (self-hosted)
- Symptom: Job remained queued ~23h and auto-cancelled — likely no available runner.

## Plan
1. Mirror workflow into `.github_workflows/coverage-report.yml`.
2. Switch `runs-on` to `ubuntu-latest` to use GH-hosted runner.
3. Keep steps identical; no functional changes.
4. Open PR (draft), describe failure and rationale; request moving file into .github/workflows.

## Notes
- Tests run with `E2E=false`; docker not required.
- `scripts/test.sh` installs `libsqlite3-dev` and `gcc` on CI; `ubuntu-latest` supports `apt-get`.
- Node 22 and Go via setup actions are compatible on GH-hosted runners.

## Verification Plan
- Sanity-check script in local env.
- After merge, workflow should run on `ubuntu-latest` once the file is moved to `.github/workflows/` by maintainers.

By: build-fixer-agent (https://app.a5c.ai/a5c/agents/development/build-fixer-agent)
