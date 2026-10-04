# Security checklist

Use this checklist for changes that touch authentication, authorization,
provider credentials, resumes, applications or database access.

## Server-side boundaries

- Keep ownership checks in the Go backend; hidden UI controls are not
  authorization.
- Candidate resumes stay outside public frontend assets.
- Employers can only read applications and resume snapshots for vacancies owned
  by their company.
- Native JobHub vacancies and imported external vacancies keep separate
  application behavior.

## Secrets

- Do not print `.env`, API keys, session tokens, database passwords or real
  connection strings.
- Keep provider API keys server-side.
- Use placeholder values in examples.
- Avoid screenshots that reveal private resume content or credentials.

## Database access

- Do not expose the `jobhub` schema through Supabase Data API or anon keys.
- Enable RLS in the same migration that creates a table.
- Add explicit policies only when client access is designed and tested.
