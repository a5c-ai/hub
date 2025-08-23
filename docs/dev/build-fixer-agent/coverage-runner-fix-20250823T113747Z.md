# Coverage workflow runner fix (initial)

- Trigger: workflow_run failure for Coverage Report (run id 17155198058)
- Observation: Job `coverage` cancelled; no runner assigned (self-hosted label `hub-dev-runners`).
- Plan: Provide a workflow variant under `.github_workflows/` that runs on `ubuntu-latest` to avoid self-hosted runner unavailability. Maintainers can move it into `.github/workflows/`.

## Steps
- Copy `.github/workflows/coverage-report.yml` to `.github_workflows/coverage-report.yml`.
- Change `runs-on` to `ubuntu-latest`.
- Keep all steps identical.
- PR as draft with context and verification notes.

