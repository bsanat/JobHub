# Candidate flow checklist

Use this checklist when changing candidate profile, saved-job, resume or
application behavior.

## Profile and resume

- Candidate profile updates should preserve existing optional data unless the
  user intentionally clears it.
- PDF resumes remain private and are served through the backend.
- Resume size limits and content-type validation should be enforced server-side.

## Saved jobs and applications

- Saved jobs are idempotent.
- A candidate can submit only one application per vacancy.
- Applications snapshot the active resume version at submission time.
- Replacing the active resume must not change already submitted applications.
- External vacancies keep their external CTA and do not create internal
  applications.

## UI verification

- Check profile completeness, resume upload, saved jobs and applications pages.
- Verify empty states in Kazakh and Russian.
