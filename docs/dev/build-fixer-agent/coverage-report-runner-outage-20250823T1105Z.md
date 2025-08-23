# Coverage Report – Runner outage analysis and proposal

## Context
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17154500217
- Workflow: `.github/workflows/coverage-report.yml`
- Annotation: "The job was not acquired by Runner of type scale-set even after multiple attempts"
- Current `runs-on`: `hub-dev-runners` (custom scale-set label)

## Diagnosis (Category 2 – infra)
Self-hosted scale-set runner is unavailable, leaving jobs queued and failing. Multiple recent runs are stuck in `queued` state.

## Proposal
Provide a fallback variant of the workflow that uses GitHub-hosted `ubuntu-latest` to keep coverage running while the scale-set is restored. Per repo rules, add this under `.github_workflows/` for a maintainer to promote to `.github/workflows/`.

## Verification Steps
- No code changes; only workflow runner target.
- When promoted, job should schedule immediately on `ubuntu-latest`.
- No functional differences in the steps.

## Links
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17154500217
- Recent queued runs: `gh run list --workflow "Coverage Report" --limit 5`

By: build-fixer-agent(https://app.a5c.ai/a5c/agents/development/build-fixer-agent)
