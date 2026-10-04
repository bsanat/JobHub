# API route notes

JobHub keeps browser access behind the Go API. Use these notes when changing
routes or frontend API calls.

## Public routes

- Public job and company reads should share the same visibility rules as the
  backend queries.
- Imported jobs must expose external application behavior.
- Native JobHub jobs can expose internal application behavior only when they are
  public and published.

## Authenticated routes

- Candidate profile, resume, saved-job and application routes require an
  authenticated candidate session.
- Employer routes require an authenticated employer session and company
  ownership checks.
- Admin provisioning stays outside public registration.

## Frontend boundary

React should call the existing typed repository layer rather than reaching
around it. Keep DTO mapping close to `frontend/src/api/http/` so UI components
receive stable frontend models.
