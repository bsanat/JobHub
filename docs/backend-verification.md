# Backend verification

Backend changes should be checked close to the touched package first, then with
the broader Go commands when the change is ready.

## Routine checks

```sh
cd backend
go test -race ./...
go vet ./...
go build ./...
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
```

Run the sqlc command after changing SQL queries, sqlc configuration or database
types used by generated code.

## Focused checks

- Handler changes should include handler tests for the affected route.
- Service logic changes should include package tests for success and failure
  paths.
- Authorization changes should prove forbidden cross-user or cross-company
  access is rejected server-side.
- Provider changes should cover normalization, request-budget behavior and safe
  error handling without logging secrets.

Never point destructive or migration tests at staging or production databases.
