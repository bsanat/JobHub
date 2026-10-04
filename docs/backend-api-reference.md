---
title: "Backend и API"
description: "Подтвержденные API endpoints, auth requirements, request/response fields and source files."
---

# Backend и API

Base path: `/api/v1`. Frontend production calls the same domain through `/api/v1`; local frontend defaults to `http://localhost:8080/api/v1`.

## Auth and session

| Method | Path | Auth | Request | Response | Errors | Source |
| --- | --- | --- | --- | --- | --- | --- |
| `POST` | `/auth/register` | Guest, trusted origin | `full_name`, `email`, `password`, `role`, optional `company_name` | `user`, optional `company`, `csrf_token`, `expires_at`; sets session and CSRF cookies | `400`, `403`, `409`, `500` | `backend/internal/handlers/auth.go`, `auth.Service.Register` |
| `POST` | `/auth/login` | Guest, trusted origin | `email`, `password` | Session response; sets cookies | `400`, `401`, `403`, `500` | `auth.go`, `auth.Service.Login` |
| `GET` | `/auth/me` | Session cookie | none | Session response | `401` | `auth.go`, `auth.Service.Authenticate` |
| `POST` | `/auth/logout` | Session + CSRF + trusted origin | none | `204` | `403`, `500` | `auth.go`, `auth.Service.Logout` |
| `GET` | `/employer/me` | `employer` | none | `user`, `company` | `401`, `403` | `router.go` |
| `GET` | `/admin/me` | `admin` | none | `user` | `401`, `403` | `router.go` |

## Public jobs and companies

| Method | Path | Auth | Query/body | Response | Uses |
| --- | --- | --- | --- | --- | --- |
| `GET` | `/health` | none | none | health JSON | `Health` handler |
| `GET` | `/jobs` | none | `q`, `city`, `preferred_city`, `work_mode`, `salary_min`, `currency`, `experience`, `employment`, `date_posted`, `sort`, `page`, `page_size` | `items`, `page`, `page_size`, `total`, `total_pages` | `jobs.Reader.Search`, `backend/sql/queries/jobs.sql` |
| `GET` | `/jobs/:id` | none | path `id` | full vacancy DTO | `jobs.Reader.Get` |
| `GET` | `/companies` | none | `q`, `page`, `page_size` | company search DTO | `companies.Store.Search` |
| `GET` | `/companies/:id` | none | path `id` | public company DTO | `companies.Store.GetPublic` |
| `GET` | `/companies/:id/jobs` | none | `page`, `page_size` | public company jobs | `companies.Store.ListPublicJobs` |

Public job list validates enum and pagination input. Invalid query values return `400 VALIDATION_ERROR`; missing jobs/companies return `404`.

## Candidate endpoints

All candidate endpoints require `job_seeker`; mutations also require trusted origin and `X-CSRF-Token`.

| Method | Path | Request | Response/data change | Source |
| --- | --- | --- | --- | --- |
| `GET` | `/profile` | none | Candidate profile | `candidates.go`, `Store.GetProfile` |
| `PATCH` | `/profile` | profile fields, nested work/education/languages | Updates profile | `validateProfile`, `Store.UpdateProfile` |
| `POST` | `/profile/resume` | multipart `file` PDF | Creates active resume metadata and stores file | `Service.UploadResume` |
| `GET` | `/profile/resume` | none | Active resume metadata | `Store.GetActiveResume` |
| `GET` | `/profile/resume/content` | none | PDF blob inline | `Service.OpenOwnResume` |
| `DELETE` | `/profile/resume` | none | Deletes/retires own resume | `Service.DeleteOwnResume` |
| `GET` | `/saved-jobs` | none | Saved jobs list | `Store.ListSavedJobs` |
| `PUT` | `/jobs/:id/saved` | none | Saves job | `Store.SaveJob` |
| `DELETE` | `/jobs/:id/saved` | none | Removes saved job | `Store.UnsaveJob` |
| `POST` | `/jobs/:id/applications` | `resume_id`, optional `message` up to 3000 chars | Creates application; duplicate returns `409 APPLICATION_EXISTS` | `Store.CreateApplication` |
| `GET` | `/applications` | none | Candidate applications | `Store.ListCandidateApplications` |
| `PATCH` | `/applications/:id/withdraw` | none | Withdraws application | `Store.WithdrawApplication` |
| `GET` | `/companies/:id/follow` | none | `{ following }` | `companies.Store.IsFollowing` |
| `GET` | `/companies/following` | none | followed companies | `companies.Store.ListFollowed` |
| `PUT` | `/companies/:id/follow` | none | follows company | `companies.Store.Follow` |
| `DELETE` | `/companies/:id/follow` | none | unfollows company | `companies.Store.Unfollow` |

## Employer endpoints

All employer endpoints require `employer`; mutations require trusted origin and CSRF.

| Method | Path | Request | Response/data change | Source |
| --- | --- | --- | --- | --- |
| `GET` | `/employer/jobs` | none | Employer-owned vacancies | `employer_jobs.go`, `ListForEmployer` |
| `GET` | `/employer/jobs/:id` | none | One employer-owned vacancy | `GetForEmployer` |
| `POST` | `/jobs` | native vacancy fields | Creates native vacancy | `CreateForEmployer` |
| `PATCH` | `/jobs/:id` | native fields or `publication_status` | Updates vacancy or transitions status | `UpdateForEmployer`, `TransitionForEmployer` |
| `DELETE` | `/jobs/:id` | none | Deletes employer-owned vacancy | `DeleteForEmployer` |
| `GET` | `/employer/company` | none | Employer company profile | `companies.Store.GetForEmployer` |
| `PATCH` | `/employer/company` | `name`, `description`, `website_url`, `logo_url`, `industry`, `city` | Updates company profile | `UpdateForEmployer` |
| `GET` | `/employer/jobs/:id/applications` | none | Applications for employer job | `ListEmployerApplications` |
| `GET` | `/employer/applications/:id/resume` | none | Resume file for application | `employerResumeContent` |
| `PATCH` | `/employer/applications/:id` | status excluding `sent` and `withdrawn` | Updates application status | `UpdateEmployerApplicationStatus` |

## Internal publisher endpoint

| Method | Path | Auth | Behavior |
| --- | --- | --- | --- |
| `POST` | `/internal/telegram-publish` | Header `X-JobHub-Publisher-Secret`, configured by `TELEGRAM_WEBHOOK_SECRET` | Runs one Telegram publisher batch and returns `checked`, `sent`, `failed`, `unknown` |

This route is registered only if `TELEGRAM_WEBHOOK_SECRET` is present and valid during API startup.

