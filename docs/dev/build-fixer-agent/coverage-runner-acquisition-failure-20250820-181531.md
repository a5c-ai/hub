Title: Coverage workflow failing due to self-hosted runner acquisition issue

Summary
- Workflow: Coverage Report (.github/workflows/coverage-report.yml)
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17006237007
- Failure cause: Job was not acquired by the configured self-hosted scale-set runner label (hub-dev-runners)
- Category: 2 - Test framework / build infrastructure issue

Details
- GitHub Actions reported: "The job was not acquired by Runner of type scale-set even after multiple attempts" for job ID 48426506372.
- The workflow currently uses: `runs-on: hub-dev-runners`.
- When no matching self-hosted runner is online or capacity is exhausted, jobs wait and eventually fail with the above message.

Local verification
- Ran the same steps locally with COVERAGE=true and E2E=false using scripts/test.sh.
- Backend (Go) unit tests passed and produced coverage.out.
- Frontend (Next.js) builds successfully; no unit tests are currently present and Jest exits successfully.
- Script artifacts observed: coverage.out (Go), frontend build output.

Recommendations
1) Infra fix (preferred): Ensure the self-hosted runner scale set labeled `hub-dev-runners` is online with sufficient capacity. Validate the label and group configuration in the repository or organization runner settings.
2) Workflow mitigation (optional): Switch the Coverage Report workflow to use `ubuntu-latest` (GitHub-hosted) or provide an alternative workflow that uses GitHub-hosted runners when self-hosted capacity is unavailable.

Proposed change in this PR
- Added a reference workflow file at `.github_workflows/coverage-report-ubuntu.yml` that mirrors the current coverage workflow but uses `runs-on: ubuntu-latest`. This file is not active by default (placed under `.github_workflows/` as per repo rules). It can be promoted to `.github/workflows/` by a maintainer if desired.

Links
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17006237007
- Head commit of failed run: 30f18b98031540e553a9abc33f40b89096373c33

Verification steps (performed locally)
- Executed: `bash scripts/test.sh --no-e2e --coverage`
- Confirmed Go tests pass and coverage file is generated.
- Confirmed frontend builds successfully; Jest exits cleanly.

Notes
- No code changes were necessary; the failure is purely infrastructure-related.

