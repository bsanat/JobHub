---
title: "Frontend"
description: "Основные страницы React-приложения, маршруты, API-вызовы и пользовательские действия."
---

# Frontend

React routes are defined in `frontend/src/App.tsx`. Protected pages use `ProtectedRoute` with optional role checks.

## Страницы

| URL | Component | Кто использует | Данные и API |
| --- | --- | --- | --- |
| `/` | `HomePage` | Гость, все пользователи | Public jobs via `repositories.jobs.search` → `GET /api/v1/jobs` |
| `/jobs` | `JobsPage` | Гость, все пользователи | Search/filter/sort public jobs → `GET /api/v1/jobs` |
| `/jobs/:jobId` | `JobDetailPage` | Гость, соискатель | Vacancy detail → `GET /api/v1/jobs/:id`; save/apply components may call candidate API |
| `/companies` | `CompaniesPage` | Гость, все пользователи | Company search → `GET /api/v1/companies` |
| `/companies/:companyId` | `CompanyProfilePage` | Гость, соискатель | Company detail, public jobs, follow state for candidate |
| `/login` | `LoginPage` | Гость | `POST /api/v1/auth/login` |
| `/register` | `RegisterPage` | Гость | `POST /api/v1/auth/register` |
| `/account` | `AccountPage` | Authenticated | Session info through `AuthProvider` and `GET /api/v1/auth/me` |
| `/profile` | `ProfilePage` | `job_seeker` | Candidate profile → `GET/PATCH /api/v1/profile` |
| `/saved` | `SavedJobsPage` | `job_seeker` | Saved jobs → `GET /api/v1/saved-jobs` |
| `/saved-companies` | `SavedCompaniesPage` | `job_seeker` | Followed companies → `GET /api/v1/companies/following` |
| `/applications` | `ApplicationsPage` | `job_seeker` | Candidate applications → `GET /api/v1/applications` |
| `/resume` | `ResumePage` | `job_seeker` | Resume metadata/content → `/api/v1/profile/resume...` |
| `/employer` | `EmployerPage` | `employer` | Employer dashboard; data source depends on child queries |
| `/employer/vacancies` | `EmployerVacanciesPage` | `employer` | Employer jobs → `GET /api/v1/employer/jobs` |
| `/employer/vacancies/new` | `EmployerVacancyFormPage` | `employer` | Create native vacancy → `POST /api/v1/jobs` |
| `/employer/vacancies/:jobId/edit` | `EmployerVacancyFormPage` | `employer` | Load/update/transition vacancy → `GET /api/v1/employer/jobs/:id`, `PATCH /api/v1/jobs/:id` |
| `/employer/vacancies/:jobId/applicants` | `EmployerApplicantsPage` | `employer` | Applications → `GET /api/v1/employer/jobs/:id/applications` |
| `/employer/company` | `EmployerCompanyPage` | `employer` | Company profile → `GET/PATCH /api/v1/employer/company` |
| `/admin` | `AdminPage` | `admin` | Protected frontend route; only `/api/v1/admin/me` is confirmed in backend |
| `/dev/ui` | `ComponentShowcasePage` | Developer | UI showcase; not a production user flow |

## Buttons and actions

| Страница/компонент | Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Ссылка на код |
| --- | --- | --- | --- | --- | --- |
| `JobsPage` / `JobSearchForm`, `JobFilters` | Search, filters, sort, pagination | `GET /api/v1/jobs` | No write | Refetches public vacancy list | `frontend/src/pages/JobsPage.tsx`, `frontend/src/components/jobs/*` |
| `JobDetailPage` | Retry load | `GET /api/v1/jobs/:id` | No write | Reloads vacancy detail | `frontend/src/pages/JobDetailPage.tsx` |
| `JobSaveButton` | Save vacancy | `PUT /api/v1/jobs/:id/saved` | Inserts `saved_jobs` row | Vacancy appears in saved list | `frontend/src/components/jobs/JobSaveButton.tsx` |
| `JobSaveButton` | Remove saved vacancy | `DELETE /api/v1/jobs/:id/saved` | Deletes `saved_jobs` row | Vacancy removed from saved list | `frontend/src/components/jobs/JobSaveButton.tsx` |
| `ApplyAction` | Apply to native vacancy | `POST /api/v1/jobs/:id/applications` | Inserts `applications` and status event | Application becomes visible in candidate and employer lists | `frontend/src/components/jobs/ApplyAction.tsx` |
| `ApplyAction` | Open external vacancy | No JobHub API write | No JobHub write | Browser opens external CTA URL | `frontend/src/components/jobs/ApplyAction.tsx` |
| `CompanyProfilePage` | Load more vacancies | `GET /api/v1/companies/:id/jobs` | No write | Appends next page of jobs | `frontend/src/pages/CompanyProfilePage.tsx` |
| `CompanyProfilePage` | Follow company | `PUT /api/v1/companies/:id/follow` | Inserts `company_follows` row | Company appears in saved companies | `frontend/src/pages/CompanyProfilePage.tsx` |
| `CompanyProfilePage` | Unfollow company | `DELETE /api/v1/companies/:id/follow` | Deletes `company_follows` row | Company removed from followed list | `frontend/src/pages/CompanyProfilePage.tsx` |
| `LoginPage` | Login | `POST /api/v1/auth/login` | Creates session row; sets cookies | User session available in `AuthProvider` | `frontend/src/pages/LoginPage.tsx` |
| `RegisterPage` | Register job seeker/employer | `POST /api/v1/auth/register` | Creates user; employer role also creates/links company membership | User is logged in | `frontend/src/pages/RegisterPage.tsx` |
| `AuthProvider` / nav | Logout | `POST /api/v1/auth/logout` | Revokes session | User becomes guest | `frontend/src/app/AuthProvider.tsx` |
| `ProfilePage` | Save profile | `PATCH /api/v1/profile` | Updates candidate profile and nested profile data | Profile form shows saved values | `frontend/src/pages/ProfilePage.tsx` |
| `ResumePage` | Upload PDF | `POST /api/v1/profile/resume` | Creates active resume metadata and stores file | Resume metadata shown | `frontend/src/pages/ResumePage.tsx` |
| `ResumePage` | Delete resume | `DELETE /api/v1/profile/resume` | Retires/deletes active resume | Resume removed from profile | `frontend/src/pages/ResumePage.tsx` |
| `ApplicationsPage` | Withdraw application | `PATCH /api/v1/applications/:id/withdraw` | Updates application status | Application status becomes withdrawn | `frontend/src/pages/ApplicationsPage.tsx` |
| `EmployerVacancyFormPage` | Save draft/new vacancy | `POST /api/v1/jobs` | Creates native job, default status from store | Employer sees vacancy in list | `frontend/src/pages/EmployerVacancyFormPage.tsx` |
| `EmployerVacancyFormPage` | Save edits | `PATCH /api/v1/jobs/:id` | Updates owned native job fields | Updated vacancy is returned | `frontend/src/pages/EmployerVacancyFormPage.tsx` |
| `EmployerVacancyFormPage` / list | Publish/pause/close | `PATCH /api/v1/jobs/:id` with `publication_status` | Changes publication status if transition allowed | Public visibility may change | `frontend/src/pages/EmployerVacancyFormPage.tsx`, `EmployerVacanciesPage.tsx` |
| `EmployerVacanciesPage` | Delete vacancy | `DELETE /api/v1/jobs/:id` | Deletes or marks employer-owned vacancy per store behavior | Vacancy disappears from employer list | `frontend/src/pages/EmployerVacanciesPage.tsx` |
| `EmployerApplicantsPage` | Change applicant status | `PATCH /api/v1/employer/applications/:id` | Updates application status | Applicant list shows new status | `frontend/src/pages/EmployerApplicantsPage.tsx` |
| `EmployerApplicantsPage` | Download resume | `GET /api/v1/employer/applications/:id/resume` | No write | Downloads application resume snapshot | `frontend/src/pages/EmployerApplicantsPage.tsx` |
| `EmployerCompanyPage` | Save company profile | `PATCH /api/v1/employer/company` | Updates employer company profile | Public company page changes | `frontend/src/pages/EmployerCompanyPage.tsx` |

Some actions are frontend-only: opening filters, changing form fields, switching local UI sections, retry buttons before they call `refetch`, and redirects such as `/workspace` → `/profile`.

## Storybook

Storybook is not configured in the repository. There is already a lightweight developer route `/dev/ui` through `ComponentShowcasePage`. For the current project documentation goal, Mintlify plus `/dev/ui` is enough. Add Storybook later only if UI component states need independent visual review outside the app.

