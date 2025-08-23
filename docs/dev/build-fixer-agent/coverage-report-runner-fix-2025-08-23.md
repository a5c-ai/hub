# Coverage Report workflow failing on self-hosted runner

## Context
- Workflow: `.github/workflows/coverage-report.yml`
- Current `runs-on`: `hub-dev-runners` (self-hosted)
- Multiple runs on `main` conclude as failure with the job itself `cancelled`, suggesting runner unavailability.

## Plan
- Copy workflow to `.github_workflows/coverage-report.yml` (per repo policy).
- Change `runs-on` to `ubuntu-latest` so tests execute on GitHub-hosted runners.
- Keep steps identical; no functionality is removed.
- Validate scripts locally where reasonable and create PR.

## Rationale
- Coverage job does not require Docker or bespoke infra.
- Using GitHub-hosted runners stabilizes coverage until self-hosted runners are healthy.

## Links
- Example failed run: https://github.com/a5c-ai/hub/actions/runs/17160876591

## Verification
- Confirm script `scripts/test.sh --no-e2e` does unit + frontend tests with coverage and no DB container.
- CI will use `actions/setup-go` and `actions/setup-node@v22` to ensure versions.

## Results
- Added `.github_workflows/coverage-report.yml` with `runs-on: ubuntu-latest` to avoid self-hosted runner unavailability.
- Opened draft PR #750 and updated description with analysis and verification steps.

## Next Steps
- After approval, a maintainer can move the file to `.github/workflows/`.
- Alternatively, restore availability for `hub-dev-runners` and close this PR if not needed.
