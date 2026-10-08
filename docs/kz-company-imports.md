# Kazakhstan company vacancies

The one-shot Go command imports only explicitly allowlisted, Kazakhstan-eligible
vacancies. It reuses the existing career-site adapters, job upsert, company
linking and ingestion-run logic. It does not run with the API server or a worker.
All 30 sources in `backend/config/kz-company-sources.json` are **disabled** and
`permissions_confirmed=false`. No live import, migration or database operation
was performed while adding this pipeline.

## Source register

"Supported" means an adapter exists in code, **not** that access or display
rights are approved. Its source is still disabled. "Review" means no safe,
complete public record contract is established. "Blocked" means access rules
could not be confirmed. Exact robots and terms observations are in the JSON
source entry; they are review notes, not legal permission.

| Company | Official career page | Technology | Status |
| --- | --- | --- | --- |
| Kaspi.kz | [Kaspi Job](https://job.kaspi.kz/) | JS; `/api/web/` disallowed by robots | Review |
| Halyk Bank | [Careers](https://halykbank.kz/about/career/vacancies) | Static HTML; career path disallowed by robots | Blocked |
| Air Astana | [Careers](https://job.airastana.com/) | Official API | Supported, disabled |
| KazMunayGas | [Work portal](https://work.kmg.kz/) | JS / hosted portal | Review |
| Kazatomprom | [HR portal](https://hrekap.kazatomprom.kz/) | JS | Review |
| KEGOC | [Corporate site](https://www.kegoc.kz/) | Samruk Qyzmet portal reported; exact career URL unverified | Review |
| Kcell | [Careers](https://jobs.kcell.kz/) | Official API | Supported, disabled |
| Beeline Kazakhstan | [People](https://people.beeline.kz/) | JS | Review |
| Tele2/Altel Kazakhstan | [Careers](https://job.tele2.kz/) | Static HTML, detail contract unverified | Review |
| Freedom Bank Kazakhstan | [Corporate site](https://www.bankffin.kz/) | Career page unverified | Review |
| Freedom Finance Kazakhstan | — | Employer identity and career page unverified | Review |
| ForteBank | [Careers](https://career.forte.kz/ru) | JS | Review |
| Bank CenterCredit | [Careers](https://www.bcc.kz/en/career/) | JS | Review |
| Jusan Bank / Alatau City Bank | [Current careers](https://alataucitybank.kz/en/career/vacancy) | JS | Review |
| Eurasian Bank | [Corporate site](https://eubank.kz/) | Career page unverified | Review |
| BI Group | [Careers](https://bi.group/ru/career) | JS; robots review returned 403 | Blocked |
| Bazis-A | [Vacancies](https://bazis.kz/vacancy) | Static HTML; per-job ID/URL unverified | Review |
| Kazakhtelecom | [Corporate site](https://telecom.kz/) | Career page unverified | Review |
| Transtelecom | [Careers](https://ttc.kz/ru/About/career/) | Static HTML; per-job URLs unverified | Review |
| Astana Hub | [Organization-filtered vacancies](https://astanahub.com/en/vacancy/?organization=ast) | Marketplace / JS; employer filter requires verification | Review |
| BTS Digital | [Careers](https://btsdigital.kz/kz/career) | Static HTML; no reliable job records | Review |
| Chocofamily | [Corporate site](https://chocofamily.kz/) | Current vacancies path unavailable | Review |
| Aviata | [Corporate site](https://aviata.kz/) | Career page unverified | Review |
| Technodom | [Vacancies](https://hr.technodom.kz/vacancies/) | Static HTML | Supported, disabled |
| Sulpak | [Career map](https://job.sulpak.kz/WorkMap/WorkMap) | JS; detail contract unverified | Review |
| Magnum | [HR site](https://hr.magnum.kz/) | JS; full listing unverified | Review |
| Small | [City jobs page](https://small.kz/ru/aksu/job) | JS; not a nationwide index | Review |
| Yandex Qazaqstan | [Corporate site](https://yandex.kz/company/ru/) | Kazakhstan-specific career feed unverified | Review |
| Kolesa Group | [Jobs](https://kolesa.group/career/job) | Static HTML | Supported, disabled |
| Qazaqstan Investment Corporation | [Careers](https://qic.kz/ru/o-nas/o-nas/karera/) | Hosted Baiterek portal; employer filter unverified | Review |

No Greenhouse, Lever, Workday, SmartRecruiters, JSON-LD or sitemap board ID was
invented. Five adapters exist in code, but Halyk is blocked by robots; four
remain selectable for a read-only review. The static
adapters have bounded requests; an incomplete fetch cannot deactivate old jobs.
Their source's publishing permission remains a separate decision.

## Safe operation

From `backend/`, an offline fixture run needs neither network nor database:

```sh
go run ./cmd/import-kz-company-jobs --config ./config/kz-company-sources.json --source halyk --dry-run --fixture ./cmd/import-kz-company-jobs/testdata/halyk-fixture.json
```

The fixture is synthetic, not a real Halyk vacancy. It is a JSON array of
existing `jobs.ImportedJob` fields. It bypasses
the allowlist enable switch only for offline evaluation. A live `--dry-run`
requires the top-level `enabled` flag and that source's `enabled` flag after
the operator confirms the terms, robots rules, and adapter completeness.
It requests public source data but never writes to the database or Telegram.
Use `--compare-db` only when a read-only comparison is desired. Set
`KZ_IMPORT_READ_DATABASE_URL`, `APP_ENV`, `KZ_IMPORT_TARGET`,
`KZ_IMPORT_DATABASE_HOST`, and `KZ_IMPORT_DATABASE_USER` for the inspected
database; the command rejects a host or user mismatch and starts a read-only
transaction. Without comparison, the report gives candidate and skipped counts
but cannot claim exact inserts, updates or deactivations. A partial snapshot
always previews zero deactivations.

For a specifically selected supported source with reviewed robots rules,
`--dry-run --probe --source <id>` fetches source data while keeping the
allowlist disabled and without claiming redistribution permission. It cannot
run on a blocked source or be combined with `--once`.

`--once` additionally needs `KZ_IMPORT_WRITE_ENABLED=true` and `DATABASE_URL`
with the same inspected identity. Staging/production also require
`KZ_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED=true`. These are explicit operator
actions; none were performed here. It upserts by `(source, external_id)`;
new company records use existing linking and are not verified automatically.
The source is registered disabled/unapproved in the DB. Public visibility
requires a separate source approval and activation review using the existing
`jobhub.job_sources` controls. The `kzcompany` provider is excluded from the
automatic Telegram publisher even after activation.

The prepared manual GitHub Actions workflow is
`.github/workflows/kz-company-imports.yml`. Its `probe` mode has no database
secret and makes no writes. Its `once` mode creates a temporary allowlist
enabling only the selected reviewed source, then uses the existing
`PUBLIC_IMPORT_DATABASE_URL` secret only after `KZ_IMPORT_ENABLED=true` and
`KZ_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED=true` are set as repository
variables, and the inspected `PUBLIC_IMPORT_DATABASE_HOST` and
`PUBLIC_IMPORT_DATABASE_USER` variables match the secret URL. The workflow
has no schedule and is not active until committed to the default branch.
After a successful non-empty import it explicitly activates that source for
public display; the Telegram publisher still excludes `kzcompany`.

## Read-only source review, 2026-10-09

The official [Halyk robots.txt](https://halykbank.kz/robots.txt) disallows
`/about/career/*`. Halyk is blocked even though a static adapter exists.
The [Air Astana robots.txt](https://job.airastana.com/robots.txt),
[Technodom robots.txt](https://hr.technodom.kz/robots.txt), and
[Kolesa robots.txt](https://kolesa.group/robots.txt) did not disallow the
reviewed public vacancy paths. Kcell's career host returned HTTP 404 for
`robots.txt`; this is not affirmative redistribution permission. No clear
third-party syndication licence was found for these four hosts, so their
`permissions_confirmed` and `enabled` flags remain false.

An authorized live `--dry-run --probe` returned the following candidate counts
without database or Telegram access:

| Source | Fetched | Kazakhstan candidates | Complete snapshot |
| --- | ---: | ---: | --- |
| Air Astana | 36 | 36 | No: API has no reliable total/pagination marker |
| Kcell | 39 | 39 | Yes |
| Technodom | 39 | 38 | No: one detail lacked usable location or failed |
| Kolesa Group | 2 | 2 | Yes |

These are observations at review time, not imported vacancy counts. Import
against the locally configured database was not attempted: its credentials
identify the prior staging project rather than the Production project.

To add a company, verify its first-party career URL and record contract,
robots/terms and redistribution permission; implement or reuse a bounded
adapter that reports snapshot completeness; add an allowlist entry with a
unique ID; test offline fixtures for Kazakhstan eligibility, dedupe, URLs and
partial failures; then explicitly enable only that source. Native JobHub jobs
and their application path remain separate. Imported jobs link to the original
HTTPS application URL, preserve available salary/date/language, and do not
infer missing fields. Worldwide remote alone is not Kazakhstan eligibility.

Proposed schedule after permission and production review: a manually triggered
`--once` per source, then at most daily per source after observing stable
pagination and rate limits. Do not start a persistent worker automatically.
