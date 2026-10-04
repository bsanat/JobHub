---
title: "История изменений"
description: "Краткая история важных изменений проекта и документации."
---

# История изменений

## 2026-10-04

- Documentation translated to Russian and connected to Mintlify.
- `docs.json` added for Mintlify navigation.
- `index.md` added as Mintlify landing page to avoid root 404.
- Developer documentation expanded with factual pages for project overview, onboarding, architecture, frontend routes, backend API, database, roles, Telegram publisher, testing, deployment, glossary and change history.

## Existing project milestones from repo state

- Public vacancy catalog and company pages are implemented.
- Candidate profile, saved jobs, saved companies, resume and applications flows are implemented in frontend and backend.
- Employer vacancy management and company profile management are implemented.
- Telegram publisher has duplicate protection, retry handling, status tracking and webhook/GitHub Actions/Supabase SQL support.
- Vercel multi-service config routes `/api/...` to Go backend and other paths to Vite frontend.

## Not confirmed

- Current live production environment variables and Supabase secrets are not visible from repo and are intentionally not documented.
- Full admin management features beyond route/session check are not confirmed.
- Preview environment database separation is not confirmed.

