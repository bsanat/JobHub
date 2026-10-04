# Testing matrix

Use this matrix to choose the smallest useful verification set for a JobHub
change, then add broader checks when the blast radius grows.

| Change type | Focused check | Broader check |
| --- | --- | --- |
| Route component or page copy | affected route in browser | frontend lint, typecheck and build |
| API repository or DTO mapper | mapper/unit coverage where present | frontend typecheck and build |
| Backend handler | handler package tests | backend race tests, vet and build |
| Domain service logic | owning package tests | backend race tests |
| SQL query | query-level tests when present | sqlc generation and backend tests |
| Migration | local migration up path | schema-dependent package tests |
| Provider adapter | provider package tests | ingestion/service tests |
| Documentation only | markdown review and `git diff --check` | no runtime check needed |

Always call out skipped checks in the PR body or handoff. Do not claim a test
passed when it was not run.
