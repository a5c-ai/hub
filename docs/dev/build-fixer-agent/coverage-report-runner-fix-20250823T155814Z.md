# Coverage Report: Runner availability fix

## Context
The Coverage Report workflow is failing due to the job being scheduled on a self-hosted runner label `hub-dev-runners` that currently has no available runners, causing the job to be immediately cancelled.

## Plan
- Propose switching `runs-on` to GitHub-hosted `ubuntu-latest` to avoid self-hosted dependency.
- Keep all steps identical.
- Submit change under `.github_workflows/` as per repo policy.
- Open draft PR for review.

## Links
- Failed run: https://github.com/a5c-ai/hub/actions/runs/17160948703
- Commit: 3de4c2665500c94221f63d933c0b29e04ec4ac52

