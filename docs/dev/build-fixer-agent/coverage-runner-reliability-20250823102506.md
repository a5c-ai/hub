# Coverage Report: Runner Reliability Fix Proposal

## Context
- Workflow: `.github/workflows/coverage-report.yml`
- Failing Run: https://github.com/a5c-ai/hub/actions/runs/17153733262
- Branch: `main`
- Symptom: Job `coverage` concluded `cancelled` with no logs on self-hosted label `hub-dev-runners` (likely capacity/unavailability).

## Plan
1. Propose switching Coverage job to `ubuntu-latest` to decouple from self-hosted pool.
2. Harden `scripts/test.sh` to gracefully handle missing `sudo/apt` on varied runners.
3. Open PR with rationale and links, labelled `build`, `bug`.

## Verification
- Local static verification of script changes.
- CI should exercise on GitHub-hosted runners once adopted.

## Notes
- Per repo policy, do not edit `.github/workflows/` directly. Add proposed workflow change under `.github_workflows/` for later promotion.
