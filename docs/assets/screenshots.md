---
title: "Скриншоты"
description: "Placeholder page for safe screenshots required by JobHub documentation."
---

# Скриншоты

Скриншоты пока не добавлены намеренно. Перед добавлением убедитесь, что на них нет:

- real user emails or names;
- tokens, API keys, database URLs;
- private resumes or uploaded PDF content;
- private Telegram channel/admin metadata;
- production-only admin data.

## Needed screenshots

| Screenshot | Suggested route | Data requirement |
| --- | --- | --- |
| Home/catalog | `/` or `/jobs` | Demo or fake vacancies |
| Job detail | `/jobs/:jobId` | Public test vacancy |
| Company profile | `/companies/:companyId` | Public test company |
| Candidate saved jobs | `/saved` | Test job seeker account |
| Candidate resume page | `/resume` | Fake PDF metadata only |
| Employer vacancy form | `/employer/vacancies/new` | Test employer account |
| Employer applicants | `/employer/vacancies/:jobId/applicants` | Fake candidate data |
| Telegram post example | Telegram channel screenshot | Mask private metadata if needed |
| Mintlify homepage | docs site root | No secrets |

