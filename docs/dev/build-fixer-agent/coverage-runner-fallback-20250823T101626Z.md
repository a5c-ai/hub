# Coverage workflow failing due to self-hosted runner not acquired

## Context
- Workflow: .github/workflows/coverage-report.yml
- Run: https://github.com/a5c-ai/hub/actions/runs/17153472685
- Conclusion: failure (job not acquired by scale-set runner "hub-dev-runners")

## Plan
- Propose fallback to `ubuntu-latest` in mirrored workflow under `.github_workflows/`
- Open infra issue to restore `hub-dev-runners` health

## Actions Taken
- Analyzed logs with gh; confirmed runner acquisition failure.
- Prepared PR with fallback workflow copy using hosted runner.

## Verification
- Script `scripts/test.sh --no-e2e` is compatible with hosted Ubuntu (uses sudo apt-get only when CI=true; coverage job sets CI=true; hosted runners support sudo apt-get).

## Follow-up
- Once infra is healthy, revert to using the self-hosted label.
