# Release readiness

Use this lightweight checklist before tagging or deploying an MVP increment.

## Code health

- Backend tests, vet, build and sqlc generation complete.
- Frontend lint, typecheck and build complete.
- `git diff --check` reports no whitespace issues.
- Generated files are up to date when their sources changed.

## Product checks

- Candidate and employer happy paths still work locally.
- Public job search and job detail pages handle native and imported vacancies.
- Company catalog and company profile pages load with empty and populated data.
- Kazakh and Russian locales both render the touched flows.

## Operational checks

- No secrets or real connection strings are in commits.
- Remote migration version is checked before running staging migrations.
- Provider quota is not consumed without explicit authorization.
