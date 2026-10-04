---
title: "Пользовательские сценарии"
description: "Подтвержденные frontend/backend flows: поиск, сохранение, компании, вакансии работодателя and Telegram publishing."
---

# Пользовательские сценарии

## Поиск вакансий

```mermaid
sequenceDiagram
  actor User
  participant FE as JobsPage
  participant API as GET /api/v1/jobs
  participant DB as jobhub.jobs
  User->>FE: вводит запрос, фильтры, сортировку
  FE->>API: q, city, work_mode, salary_min, sort, page
  API->>DB: Search public jobs
  DB-->>API: page of visible jobs
  API-->>FE: items, page, total
  FE-->>User: карточки вакансий
```

Компоненты: `JobsPage`, `JobSearchForm`, `JobFilters`, `SourceAwareJobRow`. Backend validates query parameters in `backend/internal/handlers/jobs.go`. Public visibility is enforced by store queries and the database function `jobhub.job_is_public`.

## Вход и сохранение вакансии

```mermaid
sequenceDiagram
  actor Candidate
  participant FE as LoginPage / JobSaveButton
  participant Auth as /api/v1/auth/login
  participant Save as /api/v1/jobs/:id/saved
  participant DB as users, sessions, saved_jobs
  Candidate->>FE: login
  FE->>Auth: email + password
  Auth->>DB: validate user, create session
  Auth-->>FE: session + CSRF cookie
  Candidate->>FE: save vacancy
  FE->>Save: PUT with CSRF
  Save->>DB: insert saved_jobs
```

Save/unsave requires `job_seeker` role and CSRF header for mutations. If the user is a guest, frontend behavior depends on `JobSaveButton`; backend will reject unauthenticated requests.

## Компания и подписка

```mermaid
sequenceDiagram
  actor Candidate
  participant FE as CompanyProfilePage
  participant API as companies endpoints
  participant DB as companies, company_follows
  Candidate->>FE: opens /companies/:id
  FE->>API: GET company and public jobs
  API->>DB: read company and visible jobs
  Candidate->>FE: follow
  FE->>API: PUT /companies/:id/follow
  API->>DB: insert company_follows
```

Public company page is available to guests. Follow state and follow/unfollow are confirmed only for authenticated `job_seeker`.

## Черновик, редактирование and публикация вакансии

```mermaid
sequenceDiagram
  actor Employer
  participant FE as EmployerVacancyFormPage
  participant API as employer job endpoints
  participant DB as jobs
  Employer->>FE: fills vacancy form
  FE->>API: POST /api/v1/jobs
  API->>DB: create native job owned by employer company
  Employer->>FE: edits fields
  FE->>API: PATCH /api/v1/jobs/:id
  API->>DB: update owned job
  Employer->>FE: publish
  FE->>API: PATCH publication_status=published
  API->>DB: transition status if allowed
```

Required backend fields for create/update include title, category, description, responsibilities, requirements, city, work mode, employment type and experience level. Transitions accept `draft`, `published`, `paused`, `closed`; invalid transitions return `409 INVALID_JOB_TRANSITION`.

## Telegram publishing

```mermaid
sequenceDiagram
  participant DB as Supabase/Postgres
  participant Hook as pg_net/pg_cron or GitHub Actions
  participant API as POST /api/v1/internal/telegram-publish
  participant Pub as telegrampublisher
  participant TG as Telegram Bot API
  DB->>Hook: job becomes public or retry cron fires
  Hook->>API: POST with X-JobHub-Publisher-Secret
  API->>Pub: ProcessOnce
  Pub->>DB: claim eligible jobs in telegram_publications
  Pub->>TG: send formatted HTML message
  TG-->>Pub: message_id or error
  Pub->>DB: mark sent/failed/pending
```

Confirmed publishing conditions are documented in [Telegram publisher](/docs/telegram-publisher-reference). Continuous work on Vercel is not guaranteed by an always-running API process: code explicitly disables background interval publisher when `VERCEL=1`; production-style execution is through webhook, Supabase SQL and/or GitHub Actions.

