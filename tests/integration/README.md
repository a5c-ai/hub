Integration Tests

This directory contains integration tests for the Hub backend and related services.

## Prerequisites

- Docker and Docker Compose installed.
- Services defined in `docker-compose.yml` are running:

```bash
docker-compose up -d postgres elasticsearch
```

For API integration tests, the backend server must also be running (e.g., on port 8080).

## Running Integration Tests

By default, integration tests are excluded and can be run with the `integration` build tag:

```bash
go test -tags=integration ./tests/integration/...
```

Ensure that services are healthy before running tests; tests will retry connections briefly but require services to be accessible.

## Planned Integration Test Categories

The following integration tests are planned to cover key service and workflow interactions:

- **Repository Data Flow Tests**: verify git operations result in correct database persistence and metadata synchronization.
- **User & Authentication Integration Tests**: validate authentication flows, session management, and permission enforcement.
- **Issue & PR Workflow Integration Tests**: ensure issue creation, PR workflows, and CI/CD triggers operate end-to-end.
