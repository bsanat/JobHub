# Data safety notes

JobHub handles resumes, accounts, applications and provider data. These notes
summarize the default safety posture for everyday changes.

## Private user data

- Do not place resumes in public frontend directories.
- Do not include private user content in fixtures, screenshots or docs.
- Do not log raw session tokens, passwords or uploaded resume content.
- Keep candidate and employer account boundaries explicit in tests.

## Provider data

- Preserve source, external ID and source URL provenance.
- Do not assume raw company names identify the same employer.
- Do not infer salary, publication date or lifecycle state when a provider did
  not supply it.
- Hide sources in production until permissions are confirmed.

## Operational data

Use placeholder values in examples and redact real connection details from
diagnostics.
