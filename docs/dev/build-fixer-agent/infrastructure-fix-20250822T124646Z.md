# Infrastructure Workflow Fix - prepare azure credentials outputs

## Plan
- Copy .github/workflows/infrastructure.yml to .github_workflows/infrastructure.yml
- Replace deprecated set-output usage with GITHUB_OUTPUT
- Ensure creds JSON is emitted correctly (raw JSON, not double-stringified)
- Open PR as draft

## Notes
This avoids editing workflow directly per repo guidelines.


## Results (2025-08-22T12:48:18Z)
- Created PR #739 with fix in .github_workflows/.
- Will require maintainer to move file into .github/workflows/ to take effect.
