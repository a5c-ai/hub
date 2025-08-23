# Build Fix: Coverage Report runner unavailability

## Context
- Workflow: .github/workflows/coverage-report.yml ("Coverage Report")
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17153795009
- Symptom: Jobs queued/cancelled due to unavailable self-hosted label `hub-dev-runners`.

## Plan
- Copy workflow to `.github_workflows/coverage-report.yml` per repo policy.
- Switch `runs-on` to GitHub-hosted `ubuntu-24.04` to unblock.
- Keep steps identical; no functionality removed.
- Validate locally: run Go unit tests; sanity-check frontend config.

## Notes
- E2E disabled in this workflow; no Docker needs on GH-hosted runner.
- test.sh installs `libsqlite3-dev gcc` in CI; available on `ubuntu-24.04`.
