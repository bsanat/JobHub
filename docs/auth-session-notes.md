# Auth and session notes

JobHub sessions are server-managed and should stay opaque to the frontend.

## Session model

- Login and registration create HttpOnly cookie sessions.
- Mutation requests use the CSRF header expected by the backend.
- Session tokens are stored server-side only as hashes.
- Logout should clear the browser cookie and invalidate the server-side session.

## Role boundaries

- Candidate-only routes must reject employer-only sessions.
- Employer-only routes must reject candidate-only sessions.
- Employer company access requires membership or ownership validation in the
  backend.
- Admin users are provisioned through backend tooling, not public signup.

When auth behavior changes, update handler tests and avoid exposing token values
in logs or frontend state.
