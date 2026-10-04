---
title: "Тестирование и устранение ошибок"
description: "Подтвержденные команды проверки, known risks, troubleshooting and screenshot placeholders."
---

# Тестирование и устранение ошибок

## Confirmed checks

| Area | Command | Notes |
| --- | --- | --- |
| Frontend build | `cd frontend && npm run build` | Runs TypeScript build and Vite build |
| Frontend typecheck | `cd frontend && npm run typecheck` | TypeScript only |
| Frontend lint | `cd frontend && npm run lint` | ESLint with `--max-warnings=0` |
| Backend tests | `cd backend && go test ./...` | May need database-related env for integration tests |
| Diff whitespace | `git diff --check` | Run before handoff |
| Mintlify config syntax | `node -e "JSON.parse(require('fs').readFileSync('docs.json','utf8'))"` | Checks JSON only |

Do not point test database variables at staging or production Supabase.

## Troubleshooting

| Symptom | Likely source | What to check |
| --- | --- | --- |
| Frontend cannot reach API locally | `VITE_API_URL`, backend port | `frontend/src/lib/env.ts`, backend running on `8080` |
| `CSRF_INVALID` | Missing or stale CSRF token | `AuthProvider`, cookies, `X-CSRF-Token` header |
| Employer mutation rejected | Role or origin | User role, `FRONTEND_ORIGIN`, trusted origin middleware |
| Saved job fails | User not `job_seeker` or missing CSRF | `RequireRole(job_seeker)`, `HttpCandidateRepository.saveJob` |
| Telegram does not send | Publisher config or eligibility | `TELEGRAM_*`, `JOBHUB_PUBLIC_BASE_URL`, `telegram_publications`, source permission flags |
| Publisher works locally but not on Vercel interval | Expected by code | `telegram_publisher.go` disables background loop when `VERCEL=1`; use webhook/Supabase/GitHub Actions |
| Remote migration risk | Wrong env target | Inspect destination safely without printing secrets |

## Screenshots

Placeholders live in `docs/assets/screenshots.md`. Required screenshots should be captured without personal data:

- Home page / vacancy catalog.
- Job detail page.
- Company profile page.
- Candidate saved jobs.
- Candidate profile/resume page with fake data.
- Employer vacancy form.
- Employer applicants page with fake data.
- Telegram channel example with masked/non-sensitive channel metadata.
- Mintlify docs homepage.

Do not add screenshots that show tokens, emails of real users, private resumes, database URLs or admin-only production data.

