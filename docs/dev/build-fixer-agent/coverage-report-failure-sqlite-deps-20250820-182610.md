# CI Coverage Report Failure - SQLite build deps on runners

Context:
- Failing workflow: Coverage Report
- Recent runs: failing on main branch push
- Suspected cause: scripts/test.sh attempts to install libsqlite3-dev via `sudo apt-get` unconditionally in CI, which fails on some self-hosted runner images (non-Debian/Ubuntu or no sudo)

Plan:
- Make scripts/test.sh robust: detect available package manager (apt-get, apk, dnf, yum, zypper) and install appropriate sqlite build dependencies; gracefully warn/continue if installation isn't possible.
- Keep CGO_ENABLED=1 for tests using go-sqlite3.
- Verify locally by running scripts/test.sh with coverage and ensure coverage artifacts are generated.
- Open PR with details and link to failed runs.

Changes to implement:
- Update scripts/test.sh CI dependency installation block to be distro-aware and sudo-aware.

Verification:
- Ran `bash scripts/test.sh --no-e2e --coverage` locally; coverage.out and frontend/coverage/lcov.info generated successfully.

