---
title: "Роли пользователей и проверка прав"
description: "Гость, соискатель, работодатель and admin: frontend route guards and backend authorization."
---

# Роли пользователей и проверка прав

Roles are defined in frontend models and backend auth code as:

- `job_seeker`
- `employer`
- `admin`

User statuses include `active` and `suspended`.

## Frontend protection

`frontend/src/App.tsx` uses `ProtectedRoute`:

| Route group | Required role |
| --- | --- |
| `/account` | Authenticated user |
| `/profile`, `/saved`, `/saved-companies`, `/applications`, `/resume` | `job_seeker` |
| `/employer`, `/employer/vacancies`, `/employer/company`, employer applicants pages | `employer` |
| `/admin` | `admin` |

Frontend guards are UX protection only. They do not replace backend authorization.

## Backend protection

Backend route groups use:

- `RequireRole(authService, role)` for session and role checks.
- `RequireTrustedOrigin(origin)` for state-changing requests.
- `RequireCSRF(authService)` for mutations.

Confirmed examples:

| Role | Allowed backend capabilities |
| --- | --- |
| Guest | Public jobs and companies, register, login, health |
| `job_seeker` | Candidate profile, resume, saved jobs, applications, company follow/unfollow |
| `employer` | Employer jobs, employer company profile, employer applications and resume access for applicants |
| `admin` | `/api/v1/admin/me`; additional admin operations are **не подтверждено** |

## Cookies and CSRF

Auth handlers set:

- HttpOnly session cookie.
- Non-HttpOnly CSRF cookie.

Mutations send `X-CSRF-Token`. In production, cookies are marked `Secure` because `authHandler.secure` is true when `APP_ENV=production`.

## Suspended users

Auth error mapping includes `ErrSuspended` → `403 ACCOUNT_SUSPENDED`. Exact UI handling for suspended users should be verified in `AuthProvider` and page behavior before documenting as a complete product flow.

