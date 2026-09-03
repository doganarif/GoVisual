# Contributing to GoVisual

Thanks for helping improve GoVisual. Keep changes focused, add tests for behavior changes, and update generated dashboard assets when the UI changes.

## Prerequisites

- Go 1.24 for the core, MCP, and storage modules; Go 1.25 for telemetry and examples (or Go 1.25 for the full repository)
- Node.js 22 and npm for dashboard work
- Docker for PostgreSQL, Redis, MongoDB, and integration examples
- CGO and SQLite development headers for the SQLite module

## Set up a checkout

```bash
git clone https://github.com/YOUR_GITHUB_USER/GoVisual.git
cd GoVisual
git remote add upstream https://github.com/doganarif/GoVisual.git
go mod download
```

Use a short semantic branch name such as `fix/replay-validation`, `feat/request-filter`, or `docs/dashboard-guide`.

## Run the tests

Test the core v2 module:

```bash
go test ./...
```

The storage backends, telemetry package, MCP server, and examples are separate Go modules. Run them from their module directories:

```bash
for module in store/postgres store/redis store/mongodb store/sqlite telemetry mcp cmd/examples; do
	(cd "$module" && go test ./...)
done
```

PostgreSQL, Redis, and MongoDB tests need running services and the same connection variables used by CI:

```bash
export PG_CONN='postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable'
export REDIS_CONN='redis://localhost:6379/0'
export MONGO_URI='mongodb://root:root@localhost:27017'
```

Build all examples with:

```bash
(cd cmd/examples && go build ./...)
```

Before opening a pull request, format changed Go files and run static checks:

```bash
gofmt -w path/to/changed.go
go vet ./...
```

## Dashboard development

The dashboard source is in `internal/dashboard/ui`. Its production JavaScript and CSS are committed under `internal/dashboard/static` because the Go binary embeds them.

Install exactly the locked dependency set, type-check, and rebuild:

```bash
cd internal/dashboard/ui
npm ci
npm run typecheck
npm run build
```

Commit `package-lock.json` whenever dependencies change. After any UI change, include the regenerated `internal/dashboard/static/dashboard.js` and `internal/dashboard/static/styles.css`. CI rebuilds both files and fails when the committed assets differ.

For local watch mode:

```bash
npm run dev
```

## Adding a storage backend

1. Implement the `store.Store` interface from `store/store.go`.
2. Put the backend in its own module under `store/<name>`.
3. Reuse `store/storetest` for contract coverage.
4. Test persistence, ordering, capacity, cleanup, and schema migration behavior.
5. Document installation and configuration in [storage-backends.md](storage-backends.md).

## Pull requests

- Explain the user-visible problem and the chosen behavior.
- Add focused regression tests.
- Update documentation and generated assets where applicable.
- Keep unrelated refactors out of the change.
- Ensure the core module, affected submodules, dashboard checks, and example build pass.

By contributing, you agree that your contribution is licensed under the project's MIT license.
