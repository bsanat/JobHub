# Employer flow checklist

Use this checklist when changing employer vacancy or company-profile behavior.

## Vacancy lifecycle

- Draft vacancies are editable but not public.
- Published vacancies can appear in public search when all visibility predicates
  pass.
- Paused vacancies are hidden but can return to published.
- Closed vacancies cannot be republished.
- Soft-deleted vacancies stay hidden.

## Ownership

- Employer vacancy reads and mutations must check company ownership server-side.
- Applicant lists must only include applications for the employer's company.
- Resume download access must be tied to the matching owned application.

## UI verification

- Check create, edit, publish, pause and close paths.
- Check empty and loading states for vacancy and applicant lists.
- Verify Kazakh and Russian text in the employer workspace navigation.
