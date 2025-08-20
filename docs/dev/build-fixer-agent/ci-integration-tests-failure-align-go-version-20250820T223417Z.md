# Task: Fix failing Integration Tests workflow due to Go version mismatch

Context:
- Failed workflow run: https://github.com/a5c-ai/hub/actions/runs/17084110053 (Integration Tests)
- Workflow sets up Go 1.21 via actions/setup-go@v4
- go.mod currently specifies `go 1.24.0` and the `toolchain` directive was removed in a prior PR to avoid toolchain downloads
- Without the toolchain directive, using Go 1.21 with a `go 1.24` module causes builds/tests to fail on CI

Plan:
- Align project go.mod `go` version with the workflow's configured version
- Change `go 1.24.0` to `go 1.21.0`
- Run `go mod tidy`, `go build ./...`, and `go test -short ./...` locally
- Open a PR with explanation and verification steps; link the failing run

