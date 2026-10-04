# Dependency update notes

Treat dependency updates as behavior changes unless they are clearly isolated.

## Frontend

- Keep `package.json` and `package-lock.json` together.
- Run lint, typecheck and build after updating frontend packages.
- Check routes that use upgraded UI, router, form or query libraries.
- Avoid combining package updates with unrelated feature work.

## Backend

- Keep `go.mod` and `go.sum` together.
- Run package tests plus `go test -race ./...`, `go vet ./...` and
  `go build ./...`.
- Check provider clients after HTTP, JSON or test-library changes.

## Review notes

Document why the dependency changed, whether it is security-driven and which
runtime behavior was checked.
