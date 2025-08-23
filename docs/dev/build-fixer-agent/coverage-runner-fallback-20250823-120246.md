# CI Fix: Coverage Report runner fallback

- Target workflow: `.github/workflows/coverage-report.yml`
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17155755554
- Symptom: Single job `coverage` shows `cancelled` immediately; no steps/logs recorded. Run conclusion `failure`.
- Likely cause: self-hosted runner label `hub-dev-runners` unavailable/misconfigured following recent `terraform/github_runner` changes.

## Plan
1. Provide a safe fallback using GitHub-hosted runner for Coverage.
2. Do not edit `.github/workflows/`; add fixed copy under `.github_workflows/` per policy.
3. Keep all steps identical; only change `runs-on` to `ubuntu-latest`.
4. Open PR with context, label it `build` and `bug`.
5. After merge/move by privileged maintainer, verify subsequent runs succeed.

## Notes
- The test script `scripts/test.sh --no-e2e` does not require Docker services and is compatible with `ubuntu-latest`.
- Script ensures `libsqlite3-dev` is present via `apt-get` when `CI=true`.
