# Resume storage notes

Candidate resumes are private files and should never be served as static
frontend assets.

## Local development

- Docker uses the private `resume_data` volume.
- Manual backend runs use `RESUME_STORAGE_DIR`.
- Test fixtures should avoid real resume content.

## Access rules

- Candidates can read and replace their own active resume.
- Employers can read resume snapshots only through applications for vacancies
  owned by their company.
- Existing application snapshots should not change when a candidate uploads a
  new active resume.

## Production reminders

Use a private persistent disk or object storage location. Do not expose resume
paths through public URLs, frontend assets or logs.
