---
title: "Глоссарий"
description: "Основные термины JobHub для новых разработчиков."
---

# Глоссарий

| Термин | Значение |
| --- | --- |
| Native vacancy | Vacancy created directly by a JobHub employer; source is `jobhub` |
| Imported vacancy | Vacancy imported from an external provider/source |
| Public vacancy | Vacancy that passes `jobhub.job_is_public` |
| Employer | User with role `employer`; manages company and vacancies |
| Job seeker | User with role `job_seeker`; saves jobs, uploads resume and applies |
| Admin | User with role `admin`; currently confirmed only for route/session check |
| CSRF token | Token returned in auth session and sent as `X-CSRF-Token` for mutations |
| Publication status | Vacancy lifecycle: `draft`, `published`, `paused`, `closed` |
| Moderation status | Vacancy moderation field; Telegram requires approved/public visibility through SQL logic |
| Source permission | `job_sources.production_permissions_confirmed`; used by Telegram eligibility |
| Telegram publication | Row in `telegram_publications` preventing duplicate channel posts |
| Pending publication | Claimed Telegram row not yet marked sent or failed |
| Unknown delivery | Telegram send outcome uncertain; publisher leaves row pending |
| Supabase | Managed PostgreSQL and Storage target used by production-like deployments |
| Vercel service | One service inside Vercel multi-service project (`frontend`, `backend`) |
| Mintlify | Documentation site generator used by this repo through `docs.json` |

