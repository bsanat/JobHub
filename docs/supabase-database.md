---
title: "Supabase и база данных"
description: "PostgreSQL schema, migrations, key tables, visibility rules and migration safety."
---

# Supabase и база данных

The repo uses PostgreSQL migrations under `backend/migrations`. The schema name is `jobhub`. Supabase is used as managed PostgreSQL in production-like setups; the Go backend connects through `DATABASE_URL`.

## Migration order

| Migration | Purpose |
| --- | --- |
| `000001_initialize` | Initial schema setup; exact content not expanded in this page |
| `000002_create_vacancy_supply` | `job_sources`, `companies`, `company_sources`, `jobs`, `ingestion_runs`; public job indexes; RLS enabled |
| `000003_create_auth_foundation` | `users`, company ownership fields, `company_memberships`, `sessions`; RLS enabled |
| `000004_create_native_vacancy_flow` | Native vacancy fields, employer/public indexes, `jobhub.job_is_public(candidate jobhub.jobs)` |
| `000005_create_candidate_profiles` | `candidate_profiles`, `work_experience`, `education`, `skills`, `candidate_skills`, `candidate_languages` |
| `000006_create_resumes_and_saved_jobs` | `resumes`, `saved_jobs`, one active resume index |
| `000007_create_applications` | `applications`, `application_status_events` |
| `000008_add_candidate_birth_date` | Adds candidate birth date |
| `000009_add_canonical_city_ids` | Canonical city IDs and validation function/index |
| `000010_create_company_profiles_and_follows` | Company profile fields and `company_follows` |
| `000011_link_imported_companies` | Normalized company linkage and unique normalized name index |
| `000012_seed_curated_companies` | Curated company seed data |
| `000013_optimize_public_job_reads` | Additional public read indexes |
| `000014_create_telegram_publications` | Publication tracking table for Telegram duplicate protection and retry state |

## Key tables

| Table | Main purpose | Used by |
| --- | --- | --- |
| `jobhub.users` | Login identity, role and status | auth service, role middleware |
| `jobhub.sessions` | Session token hashes and CSRF tokens | auth middleware |
| `jobhub.companies` | Company profile and imported/native company data | company pages, employer company settings |
| `jobhub.company_memberships` | Employer ownership/relationship to company | employer auth and vacancy ownership |
| `jobhub.job_sources` | Source registry, enabled flags and production permission flags | importers, public reads, Telegram eligibility |
| `jobhub.jobs` | Native and imported vacancies | catalog, employer dashboard, Telegram publisher |
| `jobhub.candidate_profiles` | Candidate profile core fields | profile page |
| `jobhub.resumes` | Resume metadata and storage key/path | resume upload/download, applications |
| `jobhub.saved_jobs` | Candidate saved vacancies | saved page, save buttons |
| `jobhub.company_follows` | Candidate follows company | saved companies, company follow state |
| `jobhub.applications` | Candidate applications to native vacancies | candidate and employer application pages |
| `jobhub.application_status_events` | Application status history | application workflows |
| `jobhub.telegram_publications` | Telegram publish status, attempts and message IDs | Telegram publisher |

## Public visibility

The database function `jobhub.job_is_public(candidate jobhub.jobs)` is created in migration `000004`. The exact logic lives in SQL; code uses it in public vacancy queries and Telegram eligibility. Any change to visibility must be coordinated across:

- `backend/migrations/*`
- `backend/sql/queries/jobs.sql`
- `backend/internal/jobs/postgres_store.go`
- `backend/internal/telegrampublisher/store.go`

## RLS

Migrations enable row-level security on key tables. The Go backend still performs server-side authorization checks and uses backend database credentials. Actual RLS policies are not expanded in the observed migration summaries; therefore fine-grained policy behavior is **не подтверждено** from the inspected snippets. Treat backend authorization as the confirmed protection layer.

## Data changes by user action

| Action | Tables |
| --- | --- |
| Register job seeker | `users`, `sessions`, likely `candidate_profiles` through auth/domain service |
| Register employer | `users`, `companies`, `company_memberships`, `sessions` |
| Save vacancy | `saved_jobs` |
| Follow company | `company_follows` |
| Upload resume | `resumes` and storage backend |
| Apply to native vacancy | `applications`, `application_status_events` |
| Employer creates vacancy | `jobs` with source `jobhub` |
| Employer publishes vacancy | `jobs.publication_status`, `published_at` depending on store logic |
| Telegram sends vacancy | `telegram_publications` |

## Migrations

Confirmed commands:

```sh
make migrate-local
make migrate-staging
make migrate-version
```

Safety notes:

- `make migrate-local` uses Docker Compose profile `local`.
- `make migrate-staging` calls `scripts/migrate-remote.sh up`; verify target URL before running.
- Do not run remote migrations from docs examples against production unless there is explicit approval and a verified backup/rollback plan.
- Do not paste real `DATABASE_URL` values into docs or screenshots.

