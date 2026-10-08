# Kazakhstan company vacancies

The one-shot Go command imports only explicitly allowlisted, Kazakhstan-eligible
vacancies. It reuses the existing career-site adapters, job upsert, company
linking and ingestion-run logic. It does not run with the API server or a worker.
All 30 sources in `backend/config/kz-company-sources.json` are **disabled** and
`permissions_confirmed=false`. Production activation is handled only by an
explicit, guarded GitHub Actions run; the checked-in configuration stays off.

On 2026-10-09, with the operator's permission, the manual workflow imported
36 Air Astana, 39 Kcell, 38 Technodom and 2 Kolesa Group records. Air Astana
and Kcell were already served by older official-careers importers, so the new
records overlap. Read-only comparison found 14 exact original-URL matches at
Air Astana and 36 at Kcell. The `reconcile-preview` workflow mode shows exact
matches and unique new records. `reconcile` retires only the matched records,
preserving the 25 unique new vacancies. Future `once` runs automatically
reconcile Air Astana and Kcell after importing, so the two official adapters
can continue to supply unique records without leaving duplicates public.
Both reconciliation runs succeeded: 50 duplicate records were retired, leaving
65 distinct new active vacancies across the four sources. The public
Kazakhstan listing showed 158 jobs immediately afterward, including 22 from
the new Air Astana source, 3 from the new Kcell source, 38 Technodom and 2
Kolesa Group jobs. These counts are a dated snapshot, not a promised feed size.
The 2026-10-09 follow-up added 10 Bank CenterCredit and one Sulpak vacancy;
the Kazakhstan listing then showed 169 jobs. A complete scheduled six-source
run updated existing records without inserting copies or deactivating jobs.

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
| Air Astana | [Careers](https://job.airastana.com/) | Official API | Existing importer active; new pipeline adds unique records |
| KazMunayGas | [Official career guidance](https://www.kmg.kz/ru/career/) → [Samruk Qyzmet employer page](https://qsamruk.kz/company/kazmunaygas) | Official hosted portal HTML | Supported, disabled by default; six initial public jobs |
| Kazatomprom | [HR portal](https://hrekap.kazatomprom.kz/) | JS | Review |
| KEGOC | [Official career link](https://www.kegoc.kz/ru/chairmans-blog/) → [Samruk Qyzmet employer page](https://qsamruk.kz/company/32) | Official hosted portal HTML | Supported, disabled by default; six initial public jobs |
| Kcell | [Careers](https://jobs.kcell.kz/) | Official API | Existing importer active; new pipeline adds unique records |
| Beeline Kazakhstan | [People](https://people.beeline.kz/) | JS | Review |
| Tele2/Altel Kazakhstan | [Careers](https://job.tele2.kz/) | Static HTML, detail contract unverified | Review |
| Freedom Bank Kazakhstan | [Corporate site](https://www.bankffin.kz/) | Career page unverified | Review |
| Freedom Finance Kazakhstan | — | Employer identity and career page unverified | Review |
| ForteBank | [Careers](https://career.forte.kz/ru) | JS | Review |
| Bank CenterCredit | [Official vacancies](https://www.bcc.kz/en/career/vacancies/) | Server-rendered HTML, numeric vacancy URLs | Supported, disabled by default |
| Alatau City Bank (formerly Jusan Bank) | [Official vacancies](https://alataucitybank.kz/career/vacancy) | Server-embedded vacancy JSON | Supported, disabled by default; six initial public jobs |
| Eurasian Bank | [Corporate site](https://eubank.kz/) | Career page unverified | Review |
| BI Group | [Careers](https://bi.group/ru/career) | JS; robots review returned 403 | Blocked |
| Bazis-A | [Vacancies](https://bazis.kz/vacancy) | Static HTML; per-job ID/URL unverified | Review |
| Kazakhtelecom | [Samruk Qyzmet employer page](https://qsamruk.kz/company/jsc-kazakhtelecom) | Official hosted portal HTML | Supported, disabled by default; six initial public jobs |
| Transtelecom | [Careers](https://ttc.kz/ru/About/career/) | Static HTML; per-job URLs unverified | Review |
| Astana Hub | [Official careers](https://astanahub.com/ru/pages/h/careers) → [open employer vacancies](https://astanahub.com/ru/vacancy/?company_id=10658&opened=True) | Official marketplace HTML, employer-filtered | Supported, disabled by default; six initial public jobs |
| BTS Digital | [Careers](https://btsdigital.kz/kz/career) | Static HTML; no reliable job records | Review |
| Chocofamily | [Corporate site](https://chocofamily.kz/) | Current vacancies path unavailable | Review |
| Aviata | [Corporate site](https://aviata.kz/) | Career page unverified | Review |
| Technodom | [Vacancies](https://hr.technodom.kz/vacancies/) | Static HTML | Supported, disabled |
| Sulpak | [Office vacancies](https://job.sulpak.kz/WorkInCompany/Office) | Server-rendered HTML, stable detail URLs | Supported, disabled by default; office roles only |
| Magnum | [Office vacancies](https://hr.magnum.kz/office/) | Static HTML, stable first-party detail URLs | Supported, disabled by default; six initial public jobs |
| Small | [City jobs page](https://small.kz/ru/aksu/job) | JS; not a nationwide index | Review |
| Yandex Qazaqstan | [Corporate site](https://yandex.kz/company/ru/) | Kazakhstan-specific career feed unverified | Review |
| Kolesa Group | [Jobs](https://kolesa.group/career/job) | Static HTML | Supported, disabled |
| Qazaqstan Investment Corporation | [Official company listing](https://jobs.baiterek.gov.kz/kompanii/?ELEMENT_ID=757) | Baiterek-hosted HTML, numeric detail URLs | Supported, disabled by default; six initial public jobs at most |

No Greenhouse, Lever, Workday, SmartRecruiters, JSON-LD or sitemap board ID was
invented. Ten adapters exist in code, but Halyk is blocked by robots; nine
remain selectable for a read-only review. The static
adapters have bounded requests; an incomplete fetch cannot deactivate old jobs.
Their source's publishing permission remains a separate decision.

The official [Magnum office listing](https://hr.magnum.kz/office/) exposes
first-party vacancy detail URLs. Its [robots.txt](https://hr.magnum.kz/robots.txt)
allows the listing and detail paths. A read-only probe on 2026-10-09 found
nine Kazakhstan office candidates. Once a Production connection is available,
the first authorized import will show at most six; older initial records remain
hidden, while genuinely new jobs discovered later are published automatically.
The source remains disabled until that import and activation succeed.

The [QIC career page](https://qic.kz/ru/o-nas/o-nas/karera/) links to its
[company listing on the Baiterek jobs portal](https://jobs.baiterek.gov.kz/kompanii/?ELEMENT_ID=757).
The portal's robots.txt returned HTTP 404; no published disallow rule was
found. A read-only probe on 2026-10-09 found one Kazakhstan vacancy with a
first-party detail URL. The adapter accepts only numeric vacancy detail paths
from the QIC company listing and treats every sync as partial, so a missing
listing cannot deactivate an existing vacancy. It remains disabled pending a
successful authorized Production import and activation.

The [Alatau City Bank vacancy listing](https://alataucitybank.kz/career/vacancy)
embeds vacancy IDs, descriptions, city and available salary in its own HTML.
The [official robots.txt](https://alataucitybank.kz/robots.txt) allows that
path. A read-only probe on 2026-10-09 found 75 Kazakhstan vacancies; the
first authorized import would display six, with later new IDs displayed
automatically. The adapter accepts only cities observed on the bank's official
Kazakhstan listing, excludes unknown/foreign locations, strips external HTML
to formatted text, and does not invent a publication date. Official detail
URLs were verified to render the vacancy in a browser. The bank's former
Jusan identity must be reviewed against any existing JobHub company record
before Production activation. Its source stays disabled until then.

The [Samruk Qyzmet portal](https://qsamruk.kz/) is an official external
recruiting channel identified by [KazMunayGas](https://www.kmg.kz/ru/career/)
and [KEGOC](https://www.kegoc.kz/ru/chairmans-blog/); Kazakhtelecom's
published employee report also identifies it as a recruiting platform.
Its [robots.txt](https://qsamruk.kz/robots.txt) allows public pages. The
portal's [site agreement](https://qsamruk.kz/privacy-policy) addresses user
data rather than granting a general redistribution license; the operator
reports permission for these employers, and every imported vacancy retains
an active link to its original QSamruk detail page. The adapter reads only
the direct employer's first open-vacancy section, removes duplicate IDs
listed under multiple categories, and confirms the exact employer on each
detail page. Subsidiary, closed and location-unknown roles are excluded.
The 2026-10-09 read-only probes found three KEGOC, four KazMunayGas and 18
Kazakhtelecom candidates; one additional Kazakhtelecom title contained
copyright boilerplate and was skipped. These are partial snapshots: absent
jobs are never automatically deactivated, and the first authorized import
can display at most six per employer. The source configuration remains off.

The [Astana Hub career page](https://astanahub.com/ru/pages/h/careers)
links to its vacancy marketplace. Its [robots.txt](https://astanahub.com/robots.txt)
allows public `/ru/vacancy/` pages but disallows `/api/`, so the adapter uses
HTML only. The verified `company_id=10658&opened=True` filter showed six open
fund vacancies on 2026-10-09. Each detail is checked against the fund's
company profile; member companies' marketplace posts are excluded. The
original application URL and source-provided publication timestamp are
preserved. Pagination makes the snapshot partial, so absent jobs are never
deactivated. The source remains disabled until an authorized Production
import and activation.

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
actions. It upserts by `(source, external_id)`;
new company records use existing linking and are not verified automatically.
The source is registered disabled/unapproved in the DB. Public visibility
requires a separate source approval and activation review using the existing
`jobhub.job_sources` controls. Approved, active company vacancies are eligible
for the Telegram publisher after activation. The publisher records each
vacancy/channel pair so scheduled imports do not repost unchanged jobs.

The prepared manual GitHub Actions workflow is
`.github/workflows/kz-company-imports.yml`. Its `probe` mode has no database
secret and makes no writes. Its `once` mode creates a temporary allowlist
enabling only the selected reviewed source, then uses the existing
`PUBLIC_IMPORT_DATABASE_URL` secret only after `KZ_IMPORT_ENABLED=true` and
`KZ_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED=true` are set as repository
variables, and the inspected `PUBLIC_IMPORT_DATABASE_HOST` and
`PUBLIC_IMPORT_DATABASE_USER` variables match the secret URL. The workflow
has no schedule and runs only on manual dispatch from the default branch.
After a successful non-empty import it explicitly activates that source for
public display and makes its approved vacancies eligible for Telegram.
For Air Astana or Kcell, choose `reconcile-preview` first, inspect the exact
match counts, then `reconcile` to retire only URL-matched new-source jobs.
Their unique new jobs, the established importer, and original application URLs
remain active. Every later `once` run for these two sources reconciles again.

The separate `.github/workflows/kz-company-scheduled-import.yml` polls the
explicit `KZ_IMPORT_AUTO_SOURCES` repository-variable list every five minutes.
`KZ_IMPORT_AUTO_ENABLED=true` currently enables the six reviewed sources.
The workflow uses the same Production host/user identity and permission gates,
imports each approved source once, and reconciles Air Astana and Kcell URL
duplicates. The importer also skips URLs already public through their
established career sources, avoiding transient duplicate listings during a
scheduled run. A failed source does not prevent later sources in that run from
updating. After a successful run, the workflow invokes the Telegram publisher;
the publication table prevents reposts. The manual `once` workflow does the
same. This is **five-minute polling**, not instant source notifications; GitHub
may delay scheduled starts. Neither the API server nor a persistent worker
starts an importer automatically.

The read-only probes on 2026-10-09 found 10 Kazakhstan vacancies on the
[Bank CenterCredit official listing](https://www.bcc.kz/en/career/vacancies/)
and one location-explicit office vacancy on
[Sulpak's official office page](https://job.sulpak.kz/WorkInCompany/Office).
Two other Sulpak office details did not state a location and were excluded.
Both adapters treat their page as a partial snapshot, so unseen jobs are never
deactivated on that basis. The BCC [robots.txt](https://www.bcc.kz/robots.txt)
allows the reviewed career paths; the Sulpak career host returned 404 for
`robots.txt`, which does not establish a prohibition or grant redistribution
rights. The operator confirmed publication rights, while checked-in source
flags remain off by default.

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

New sources should be added to the five-minute workflow only after checking
their load, pagination and rate limits. GitHub Actions may delay a scheduled
run, so no fixed delivery time is promised.
