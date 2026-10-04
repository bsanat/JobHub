---
title: "Каталог страниц сайта"
description: "Подробная документация всех основных страниц JobHub: route, component, данные, API, таблицы и кнопки."
---

# Каталог страниц сайта

Эта страница описывает фактические React routes from `frontend/src/App.tsx`.
API calls are verified through `frontend/src/api/http/*` and backend handlers.

## Главная страница

| Поле | Значение |
| --- | --- |
| URL | `/` |
| Component | `HomePage` |
| Пользователи | Гость, соискатель, работодатель, админ |
| Данные | Новые публичные вакансии, популярные компании/направления из frontend constants and API results |
| Frontend load | `repositories.jobs.search({ page, pageSize: 100, sort: "newest" })` |
| API | `GET /api/v1/jobs` |
| Database | `jobhub.jobs`, `jobhub.companies`, `jobhub.job_sources`; visibility through `jobhub.job_is_public` |
| Code | `frontend/src/pages/HomePage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Hero search submit | no direct write; navigates to `/jobs?...` | none | Opens catalog with query/city params | `HomePage`, `JobSearchForm` |
| Opportunity card click | none | none | Opens `/jobs` or filtered jobs URL | `HomePage` |
| Company card click | none | none | Opens `/jobs?q=<company>` | `HomePage` |
| Preference card click | none | none | Opens `/jobs` with selected filter | `HomePage` |
| CTA “all jobs” | none | none | Opens `/jobs` | `HomePage` |
| Save vacancy on card | `PUT /api/v1/jobs/:id/saved` | Inserts `saved_jobs` | Saves vacancy for candidate | `JobSaveButton` |
| Open vacancy card | `GET /api/v1/jobs/:id` after navigation | none | Opens detail page | `HomePage` |

## Вакансии: поиск и каталог

| Поле | Значение |
| --- | --- |
| URL | `/jobs` |
| Component | `JobsPage` |
| Пользователи | Гость, все роли |
| Данные | Public vacancy cards, filters and pagination state from URL/search params |
| Frontend load | `repositories.jobs.search(search)` |
| API | `GET /api/v1/jobs` |
| Database | `jobs`, `companies`, `job_sources`; public visibility function/query |
| Code | `frontend/src/pages/JobsPage.tsx`, `frontend/src/components/jobs/JobSearchForm.tsx`, `JobFilters.tsx`, `MarketplaceJobRow.tsx` |

| Кнопка немесе действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Search submit | `GET /api/v1/jobs?q=...` | none | Refresh list by keyword | `JobSearchForm` |
| Open mobile filters | none | none | Opens drawer/modal with filters | `JobsPage` |
| Change filters | `GET /api/v1/jobs?...` after state update | none | Refresh list by city/work mode/salary/etc. | `JobFilters` |
| Active filter chip remove | `GET /api/v1/jobs?...` | none | Removes one filter and reloads list | `JobsPage` |
| Clear all | `GET /api/v1/jobs` | none | Resets filters | `JobsPage` |
| Open vacancy row | `GET /api/v1/jobs/:id` after navigation | none | Opens vacancy detail | `MarketplaceJobRow` |
| Save/unsave | `PUT` or `DELETE /api/v1/jobs/:id/saved` | `saved_jobs` insert/delete | Updates saved state for candidate | `JobSaveButton` |

## Страница вакансии

| Поле | Значение |
| --- | --- |
| URL | `/jobs/:jobId` |
| Component | `JobDetailPage` |
| Пользователи | Гость; candidate actions require `job_seeker` |
| Данные | Full vacancy DTO: title, company, description, requirements, salary, source, application method |
| API | `GET /api/v1/jobs/:id`; optional save/apply endpoints |
| Database | `jobs`, `companies`, `saved_jobs`, `applications`, `application_status_events` |
| Code | `frontend/src/pages/JobDetailPage.tsx`, `ApplyAction.tsx`, `JobSaveButton.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Back | none | none | Browser route returns to previous/catalog route | `JobDetailPage` |
| Retry | `GET /api/v1/jobs/:id` | none | Reloads detail | `JobDetailPage` |
| Company link | `GET /api/v1/companies/:id` after navigation | none | Opens company profile | `JobDetailPage` |
| Save | `PUT /api/v1/jobs/:id/saved` | `saved_jobs` insert | Vacancy saved | `JobSaveButton` |
| Unsave | `DELETE /api/v1/jobs/:id/saved` | `saved_jobs` delete | Vacancy removed from saved | `JobSaveButton` |
| Apply internal | `POST /api/v1/jobs/:id/applications` | `applications`, status event | Creates candidate application | `ApplyAction` |
| Open external application | none in JobHub | none | Opens external CTA URL | `ApplyAction` |

## Компании: список

| Поле | Значение |
| --- | --- |
| URL | `/companies` |
| Component | `CompaniesPage` |
| Пользователи | Гость, все роли |
| Данные | Company search result |
| API | `GET /api/v1/companies?q=&page=&page_size=20` |
| Database | `companies`, possibly public job counters from store query |
| Code | `frontend/src/pages/CompaniesPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Search companies | `GET /api/v1/companies?q=...` | none | Updates list | `CompaniesPage` |
| Open profile | `GET /api/v1/companies/:id` after navigation | none | Opens company page | `CompaniesPage` |

## Профиль компании

| Поле | Значение |
| --- | --- |
| URL | `/companies/:companyId` |
| Component | `CompanyProfilePage` |
| Пользователи | Гость; follow actions require `job_seeker` |
| Данные | Company detail, public vacancies, candidate follow state |
| API | `GET /companies/:id`, `GET /companies/:id/jobs`, `GET/PUT/DELETE /companies/:id/follow` |
| Database | `companies`, `jobs`, `company_follows` |
| Code | `frontend/src/pages/CompanyProfilePage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Back to companies | none | none | Opens `/companies` | `CompanyProfilePage` |
| Follow | `PUT /api/v1/companies/:id/follow` | Insert `company_follows` | Company appears in saved companies | `CompanyProfilePage` |
| Unfollow | `DELETE /api/v1/companies/:id/follow` | Delete `company_follows` | Company removed from saved companies | `CompanyProfilePage` |
| Show/hide jobs | maybe `GET /api/v1/companies/:id/jobs` if not loaded | none | Expands/collapses vacancy list | `CompanyProfilePage` |
| Load more jobs | `GET /api/v1/companies/:id/jobs?page=n` | none | Appends public jobs | `CompanyProfilePage` |
| Open company job | `GET /api/v1/jobs/:id` after navigation | none | Opens vacancy detail | `CompanyProfilePage` |

## Login

| Поле | Значение |
| --- | --- |
| URL | `/login` |
| Component | `LoginPage` |
| Пользователи | Guest |
| Данные | Email/password form |
| API | `POST /api/v1/auth/login` |
| Database | `users`, `sessions` |
| Code | `frontend/src/pages/LoginPage.tsx`, `frontend/src/app/AuthProvider.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Show/hide password | none | none | Toggles password input type | `LoginPage` |
| Submit login | `POST /api/v1/auth/login` | Creates session, sets cookies | Redirects to safe `returnTo` or default route | `LoginPage`, `HttpAuthRepository` |
| Register link | none | none | Opens `/register?returnTo=...` | `LoginPage` |

## Register

| Поле | Значение |
| --- | --- |
| URL | `/register` |
| Component | `RegisterPage` |
| Пользователи | Guest |
| Данные | Full name, email, password, role, optional company name for employer |
| API | `POST /api/v1/auth/register` |
| Database | `users`, `sessions`; for employer also company/membership through auth service |
| Code | `frontend/src/pages/RegisterPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Choose job seeker role | none | none | Form role becomes `job_seeker` | `RegisterPage` |
| Choose employer role | none | none | Form role becomes `employer`, company field is relevant | `RegisterPage` |
| Show/hide password | none | none | Toggles password input type | `RegisterPage` |
| Submit registration | `POST /api/v1/auth/register` | Creates user/session and related role records | Employer redirects to `/employer`; job seeker to `returnTo` | `RegisterPage`, `HttpAuthRepository` |
| Login link | none | none | Opens `/login?returnTo=...` | `RegisterPage` |

## Account

| Поле | Значение |
| --- | --- |
| URL | `/account` |
| Component | `AccountPage` |
| Пользователи | Authenticated user |
| Данные | Current session from `AuthProvider` |
| API | Session is loaded through `GET /api/v1/auth/me`; logout uses `POST /api/v1/auth/logout` |
| Database | `sessions` |
| Code | `frontend/src/pages/AccountPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Logout | `POST /api/v1/auth/logout` | Revokes session | Redirects to home | `AccountPage`, `AuthProvider` |
| Go to employer workspace | none | none | Opens `/employer` | `AccountPage` |
| Go to admin | none | none | Opens `/admin` | `AccountPage` |
| Job seeker account route | none | none | Redirects to `/profile#account` | `AccountPage` |

## Candidate profile

| Поле | Значение |
| --- | --- |
| URL | `/profile` |
| Component | `ProfilePage` |
| Пользователи | `job_seeker` |
| Данные | Candidate profile, skills, contacts, work experience, education, languages |
| API | `GET /api/v1/profile`, `PATCH /api/v1/profile` |
| Database | `candidate_profiles`, `work_experience`, `education`, `candidate_languages`, skills-related tables |
| Code | `frontend/src/pages/ProfilePage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Edit | none | none | Enables form fields | `ProfilePage` |
| Add/remove work experience | none until save | frontend form state | Updates form rows locally | `ProfilePage` |
| Add/remove education | none until save | frontend form state | Updates form rows locally | `ProfilePage` |
| Add/remove language | none until save | frontend form state | Updates form rows locally | `ProfilePage` |
| Cancel | none | none | Returns to read mode and resets form | `ProfilePage` |
| Save profile | `PATCH /api/v1/profile` | Updates candidate profile tables | Shows saved profile | `ProfilePage`, `HttpCandidateRepository` |

## Saved jobs

| Поле | Значение |
| --- | --- |
| URL | `/saved` |
| Component | `SavedJobsPage` |
| Пользователи | `job_seeker` |
| Данные | Candidate saved vacancies |
| API | `GET /api/v1/saved-jobs`, `DELETE /api/v1/jobs/:id/saved` |
| Database | `saved_jobs`, joined public job data |
| Code | `frontend/src/pages/SavedJobsPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Browse jobs empty state | none | none | Opens `/jobs` | `SavedJobsPage` |
| Open saved job | `GET /api/v1/jobs/:id` after navigation | none | Opens detail page | `SavedJobsPage` |
| Remove saved job | `DELETE /api/v1/jobs/:id/saved` | Deletes `saved_jobs` row | Removes row from list | `SavedJobsPage` |

## Saved companies

| Поле | Значение |
| --- | --- |
| URL | `/saved-companies` |
| Component | `SavedCompaniesPage` |
| Пользователи | `job_seeker` |
| Данные | Followed companies |
| API | `GET /api/v1/companies/following`, `DELETE /api/v1/companies/:id/follow` |
| Database | `company_follows`, `companies` |
| Code | `frontend/src/pages/SavedCompaniesPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Browse companies empty state | none | none | Opens `/companies` | `SavedCompaniesPage` |
| Open company profile | `GET /api/v1/companies/:id` after navigation | none | Opens company profile | `SavedCompaniesPage` |
| Unfollow | `DELETE /api/v1/companies/:id/follow` | Deletes `company_follows` | Removes company from saved list | `SavedCompaniesPage` |

## Candidate applications

| Поле | Значение |
| --- | --- |
| URL | `/applications` |
| Component | `ApplicationsPage` |
| Пользователи | `job_seeker` |
| Данные | Candidate applications with statuses and job/company info |
| API | `GET /api/v1/applications`, `PATCH /api/v1/applications/:id/withdraw` |
| Database | `applications`, `application_status_events`, jobs/companies joins |
| Code | `frontend/src/pages/ApplicationsPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Browse jobs empty state | none | none | Opens `/jobs` | `ApplicationsPage` |
| Open job | `GET /api/v1/jobs/:id` after navigation | none | Opens vacancy detail | `ApplicationsPage` |
| Withdraw | `PATCH /api/v1/applications/:id/withdraw` | Updates application status | Status becomes `withdrawn` | `ApplicationsPage` |

## Resume and preferences

| Поле | Значение |
| --- | --- |
| URL | `/resume`; `/preferences` redirects to `/resume#preferences` |
| Component | `ResumePage`, `JobPreferencesSection` |
| Пользователи | `job_seeker` |
| Данные | Candidate profile and active resume metadata |
| API | `GET /api/v1/profile`, `POST/GET/DELETE /api/v1/profile/resume`, `GET /profile/resume/content`, `PATCH /profile` for preferences |
| Database | `candidate_profiles`, `resumes`; storage backend for PDF |
| Code | `frontend/src/pages/ResumePage.tsx`, `JobPreferencesPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Upload/select PDF | `POST /api/v1/profile/resume` | Creates active resume and stores PDF | Resume metadata appears | `ResumePage` |
| Download resume | `GET /api/v1/profile/resume/content` | none | Browser downloads/opens PDF blob | `ResumePage` |
| Replace resume | `POST /api/v1/profile/resume` | Retires/replaces active resume per service behavior | New resume metadata appears | `ResumePage` |
| Delete resume | `DELETE /api/v1/profile/resume` | Deletes/retires active resume | Resume removed | `ResumePage` |
| Edit preferences | none | none | Opens preferences form | `JobPreferencesSection` |
| Save preferences | `PATCH /api/v1/profile` | Updates preference fields in profile | Preferences saved | `JobPreferencesSection` |

## Employer dashboard

| Поле | Значение |
| --- | --- |
| URL | `/employer` |
| Component | `EmployerPage` |
| Пользователи | `employer` |
| Данные | Static dashboard actions; session is checked by `ProtectedRoute` |
| API | no page-specific API confirmed |
| Database | none directly from page |
| Code | `frontend/src/pages/EmployerPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Manage vacancies | none | none | Opens `/employer/vacancies` | `EmployerPage` |
| Create vacancy | none | none | Opens `/employer/vacancies/new` | `EmployerPage` |

## Employer vacancies

| Поле | Значение |
| --- | --- |
| URL | `/employer/vacancies` |
| Component | `EmployerVacanciesPage` |
| Пользователи | `employer` |
| Данные | Employer-owned jobs |
| API | `GET /api/v1/employer/jobs`, `PATCH /api/v1/jobs/:id`, `DELETE /api/v1/jobs/:id` |
| Database | `jobs` |
| Code | `frontend/src/pages/EmployerVacanciesPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Create vacancy | none | none | Opens vacancy form | `EmployerVacanciesPage` |
| View published vacancy | `GET /api/v1/jobs/:id` after navigation | none | Opens public detail | `EmployerVacanciesPage` |
| Applicants | `GET /api/v1/employer/jobs/:id/applications` after navigation | none | Opens applicants page | `EmployerVacanciesPage` |
| Edit | `GET /api/v1/employer/jobs/:id` after navigation | none | Opens edit form | `EmployerVacanciesPage` |
| Publish | `PATCH /api/v1/jobs/:id` with `publication_status=published` | Updates `jobs.publication_status` | Vacancy may become public and eligible for Telegram | `EmployerVacanciesPage` |
| Pause | `PATCH /api/v1/jobs/:id` with `publication_status=paused` | Updates status | Vacancy stops being public | `EmployerVacanciesPage` |
| Close | `PATCH /api/v1/jobs/:id` with `publication_status=closed` | Updates status | Vacancy is closed | `EmployerVacanciesPage` |
| Delete | `DELETE /api/v1/jobs/:id` | Deletes/removes owned vacancy per store behavior | Row disappears from employer list | `EmployerVacanciesPage` |

## Employer vacancy create/edit form

| Поле | Значение |
| --- | --- |
| URL | `/employer/vacancies/new`, `/employer/vacancies/:jobId/edit` |
| Component | `EmployerVacancyFormPage` |
| Пользователи | `employer` |
| Данные | Native vacancy form fields |
| API | New: `POST /api/v1/jobs`; edit: `GET /api/v1/employer/jobs/:id`, `PATCH /api/v1/jobs/:id` |
| Database | `jobs` |
| Code | `frontend/src/pages/EmployerVacancyFormPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Back | none | none | Opens `/employer/vacancies` | `EmployerVacancyFormPage` |
| Save draft | `POST /api/v1/jobs` or `PATCH /api/v1/jobs/:id` | Creates/updates job; may keep draft intent | Redirects to employer vacancies after save | `EmployerVacancyFormPage` |
| Publish | `POST/PATCH` then optional `PATCH publication_status=published` | Creates/updates and publishes vacancy | Vacancy appears as published if transition allowed | `EmployerVacancyFormPage` |
| Form field edits | none until submit | frontend form state | Changes local form values | `EmployerVacancyFormPage` |

## Employer applicants

| Поле | Значение |
| --- | --- |
| URL | `/employer/vacancies/:jobId/applicants` |
| Component | `EmployerApplicantsPage` |
| Пользователи | `employer` |
| Данные | Applications for one employer-owned job |
| API | `GET /api/v1/employer/jobs/:id/applications`, `PATCH /api/v1/employer/applications/:id`, `GET /api/v1/employer/applications/:id/resume` |
| Database | `applications`, `application_status_events`, `resumes` |
| Code | `frontend/src/pages/EmployerApplicantsPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Back to vacancies | none | none | Opens employer vacancies list | `EmployerApplicantsPage` |
| Change status | `PATCH /api/v1/employer/applications/:id` | Updates application status | Applicant row shows new status | `EmployerApplicantsPage` |
| Download resume | `GET /api/v1/employer/applications/:id/resume` | none | Downloads resume snapshot | `EmployerApplicantsPage` |

## Employer company

| Поле | Значение |
| --- | --- |
| URL | `/employer/company` |
| Component | `EmployerCompanyPage` |
| Пользователи | `employer` |
| Данные | Employer company profile fields |
| API | `GET /api/v1/employer/company`, `PATCH /api/v1/employer/company` |
| Database | `companies`, `company_memberships` for ownership |
| Code | `frontend/src/pages/EmployerCompanyPage.tsx` |

| Кнопка или действие | Запрос API | Изменение данных | Ожидаемый результат | Код |
| --- | --- | --- | --- | --- |
| Save company profile | `PATCH /api/v1/employer/company` | Updates company profile fields | Employer and public company profile show updated data | `EmployerCompanyPage` |
| Retry load | `GET /api/v1/employer/company` | none | Reloads company data | `EmployerCompanyPage` |

## Admin

| Поле | Значение |
| --- | --- |
| URL | `/admin` |
| Component | `AdminPage` |
| Пользователи | `admin` |
| Данные | Static admin placeholder/dashboard content; exact admin operations are not confirmed |
| API | `ProtectedRoute` uses auth session; backend confirms `GET /api/v1/admin/me` |
| Database | `users`, `sessions` for role/session |
| Code | `frontend/src/pages/AdminPage.tsx` |

Admin CRUD actions are **не подтверждено** in inspected frontend/backend code.

## Developer UI and redirects

| URL | Behavior |
| --- | --- |
| `/dev/ui` | `ComponentShowcasePage`, development UI showcase; calls `GET /api/v1/jobs` for sample jobs |
| `/dev/components` | Redirects to `/dev/ui` |
| `/workspace` | Protected job seeker redirect to `/profile` |
| `/preferences` | Protected job seeker redirect to `/resume#preferences` |
| `*` | `NotFoundPage`, links to `/` and `/jobs` |

