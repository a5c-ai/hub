# Coverage Report workflow failing due to unavailable self-hosted runners

Context:
- Workflow: `.github/workflows/coverage-report.yml`
- Run: https://github.com/a5c-ai/hub/actions/runs/17161383602
- Head SHA: 1b76672f4084cfd8ebe04ce2495a5011016abf05

Observation:
- The job `coverage` had labels `["hub-dev-runners"]`, `runner_name` was empty, `conclusion` was `cancelled`.
- Timestamps show it waited ~23 hours (started: 2025-08-22T17:05:43Z, completed: 2025-08-23T16:15:44Z) indicating no matching runner picked it up.

Classification: Category 2 – Build infrastructure issue (self-hosted runner unavailable).

Local verification steps performed:
1. Installed frontend deps: `cd frontend && npm ci --legacy-peer-deps`.
2. Ran `./scripts/test.sh --no-e2e` with `COVERAGE=true` locally (Ubuntu environment with Docker available).
3. Result: Tests and build completed successfully; generated `coverage.out` and `frontend/coverage/lcov.info`.

Proposed Fix:
- Provide a workflow variant that uses GitHub-hosted runners to avoid blocking on self-hosted capacity.
- Added `.github_workflows/coverage-report.yml` that mirrors the original workflow but sets `runs-on: ubuntu-latest`.
- Maintainers can move it to `.github/workflows/` or update the existing workflow accordingly.

Notes:
- This change preserves all steps (`actions/setup-go`, `actions/setup-node`, test script, and artifact upload).
- No changes were made to active workflows directly per repository contribution guidelines.

By: [build-fixer-agent](https://app.a5c.ai/a5c/agents/development/build-fixer-agent)

