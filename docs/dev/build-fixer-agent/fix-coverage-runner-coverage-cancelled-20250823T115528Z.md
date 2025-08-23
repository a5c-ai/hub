# Fix: Coverage workflow failing/cancelled due to runner + sqlite prereqs

## Context
- Coverage Report workflow failed/cancelled on self-hosted label `hub-dev-runners`.
- Recent infra changes updated runner image and ARC config.
- Our Go tests depend on `github.com/mattn/go-sqlite3` which requires `libsqlite3-dev` and CGO.
- `scripts/test.sh` tries to `sudo apt-get install libsqlite3-dev gcc` in CI, which fails on non-root, sudo-less runners.

## Plan
- Add `libsqlite3-dev` (and `pkg-config`) to `runners/Dockerfile` so runners have headers preinstalled.
- Make `scripts/test.sh` detect presence of sqlite headers and skip installation; avoid unconditional `sudo`.
- Open PR; reference failing run and describe verification steps.

## Notes
- Multiple latest runs show `cancelled` with no steps: likely ARC runner availability/label issue as well. This PR ensures that once runners are available, coverage won’t fail on sqlite prereqs.
