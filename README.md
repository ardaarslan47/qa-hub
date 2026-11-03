# QA Hub

A small test management app. A test manager writes Cases, starts a Run, and records Passed or Failed for each Execution. The terms are defined in [GLOSSARY.md](GLOSSARY.md). The tech stack is in [ADR 0002](docs/adr/0002-tech-stack-go.md).

## Prerequisites

- Go, at the version in `go.mod`.
- Node.js, for the Playwright acceptance tests in `e2e/`.

First time only, install the browser tests' dependencies:

```sh
cd e2e
npm ci
npx playwright install chromium
```

## Run the app

```sh
go tool templ generate
go run ./cmd/qahub
```

Open http://127.0.0.1:8080. Flags:

- `-addr`: address to listen on, default `127.0.0.1:8080`.
- `-db`: path of the SQLite database file, default `qahub.db`. The file is created and migrated on start. Data survives restarts.

## Check

```sh
make check
```

This runs the same checks as CI (`.github/workflows/app.yml`): `go tool templ generate`, `go vet ./...`, `go test ./...`, a build of `bin/qahub`, and the Playwright suite against that binary. Each Playwright test starts its own server on a free port with a fresh database file.

Run one acceptance scenario against the built binary:

```sh
make build
cd e2e && npx playwright test -g "Case creation"
```
