# Public job API imports

The one-shot command `backend/cmd/import-public-jobs` fetches on the Go backend.
It never runs as part of the API server or a background worker. No migration or
import was run during implementation. Existing native vacancies use their
existing authoring and internal application path.

## Sources and permissions

| Source | API behavior | Rights and attribution |
| --- | --- | --- |
| [Arbeitnow](https://www.arbeitnow.com/api/job-board-api) | Paginated `data` and `links.next`; job URL, description, location, remote flag, job types. This is an international feed, not a Kazakhstan feed. | [Terms §11](https://www.arbeitnow.com/terms) require a link back to Arbeitnow.com and allow API access to be revoked. JobHub links to the original vacancy and the source homepage. Terms do not clearly grant unrestricted third-party commercial republication; review before enabling public display. |
| [Greenhouse](https://docs.greenhouse.io/job-board.html) | Public GET for a **known board token**, all published posts in one list with `content=true`, optional `meta.total`; no global company discovery. The list exposes `updated_at`, not a reliable publication date. | API documentation describes building a company's own careers page. Its [MSA](https://www.greenhouse.com/master-subscription-agreement) governs customer use, not a general syndication licence to JobHub. Confirm employer and platform rights for JobHub display. |
| [Lever](https://github.com/lever/postings-api/blob/master/README.md) | Public GET for a **known site slug**, with `skip`/`limit`; no global company discovery. Hosted application URL is used. `createdAt` is not treated as publication date. | API documentation describes a company's own jobs site. Its [terms](https://www.lever.co/legal/terms-of-service) do not establish JobHub's third-party redistribution rights. Confirm those rights before enabling display. |

The application never claims these employers registered on JobHub. Import-created
companies use existing linking logic and are not automatically verified. Public
external-job responses omit linked company IDs, logos, and verified badges so an
imported listing does not suggest a JobHub employer account or endorsement.
The frontend may show a locally sourced official company logo when the company
name matches a reviewed asset; this does not indicate employer verification.
Provider HTML is converted to safe Markdown text (paragraphs, headings and lists)
before storage and rendered without raw HTML. Existing flat descriptions gain
structure on their next successful import update. Remote status is only mapped
from an explicit provider field; locations and
eligibility restrictions stay in the provider's original description. A remote
badge is **not** proof that Kazakhstan-based applicants are eligible. Missing
salary, location, and publication dates remain absent.

The vacancy catalog offers Kazakhstan and international filters. Kazakhstan
means a source explicitly marked for the KZ market or a recognized Kazakhstan
city. The international group also contains records with unclassified location;
it does not mean worldwide work eligibility. Both filters run in the backend,
so counts and pagination remain consistent.

## Configuration

Create a server-side JSON file outside the public web root:

```json
{
  "enabled": false,
  "arbeitnow": true,
  "greenhouse": [],
  "lever": [],
  "max_pages": 10
}
```

Change `enabled` to `true` for an intentional run. Greenhouse and Lever
entries must come from a verified employer board, with the board's real
identifier and company name supplied by the operator. Example shape only:

```json
{
  "greenhouse": [{"id": "<verified-board-token>", "company": "<company-name>"}],
  "lever": [{"id": "<verified-site-slug>", "company": "<company-name>"}]
}
```

The placeholders are **not** working identifiers. Do not copy them into a live
allowlist. Source IDs are `arbeitnow:global`, `greenhouse:<token>`, and
`lever:<slug>`. A new source is registered disabled and with
`production_permissions_confirmed=false`. For Arbeitnow only, the optional
`PUBLIC_IMPORT_ACTIVATE_ARBEITNOW=true` switch activates public display after
a successful production `--once` run with at least one normalized job. This
requires a confirmed permission review. Greenhouse and Lever remain disabled
until their own rights and board IDs are confirmed.

## Commands

From `backend/`, with Go installed:

```sh
go run ./cmd/import-public-jobs --config /absolute/path/public-imports.json --dry-run
```

Dry-run fetches and normalizes but makes **no database writes or Telegram
actions**; it does not need database variables. It still makes public API GET
requests. It returns JSON reports with samples and completeness status.

For a deliberately selected target after reviewing `APP_ENV`, the database
hostname, schema, source permissions, and the write scope:

```sh
APP_ENV=development PUBLIC_IMPORT_TARGET=development PUBLIC_IMPORT_WRITE_ENABLED=true \
PUBLIC_IMPORT_DATABASE_HOST='<reviewed-hostname>' \
PUBLIC_IMPORT_DATABASE_USER='<reviewed-username>' \
DATABASE_URL='<reviewed-target-url>' \
go run ./cmd/import-public-jobs --config /absolute/path/public-imports.json --once
```

`--once` does one pass, then exits. It neither migrates nor starts a worker.
Never set the target variable by copying an uninspected staging/production
connection string. A Supabase PostgreSQL URL works; Docker is not required.
Use `PGX_QUERY_EXEC_MODE=exec` for a transaction pooler on port 6543.
For production, `PUBLIC_IMPORT_DATABASE_USER` must exactly match the username
in the reviewed Production connection URL. Supabase staging and Production can
share a pooler hostname, so the hostname check alone does not identify a
project. Do not copy a staging URL into this setting.
Staging/production also require
`PUBLIC_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED=true`, after separate rights
review; importing alone still does not enable a source for public display.

Each provider request has a 20-second timeout, at most three attempts for
transport/429/5xx errors, and a short capped `Retry-After` wait. Arbeitnow
follows only validated same-host page links. Lever fetches `skip`/`limit`
pages. A page cap, malformed record, failed page, or incomplete Greenhouse
`meta.total` makes the snapshot incomplete. Existing records are upserted by
unique `(source, external_id)`; a complete successful run alone archives
missing records in the same transaction as its upserts. A failed or partial run
does not archive missing jobs. A provider may be too large for the configured
page/Greenhouse record cap; inspect the completeness report before treating it
as a full snapshot. Greenhouse has no documented list pagination; the one-shot
importer caps one board at 5,000 records and 20 MiB of JSON, failing safely if
the board exceeds either cap.

## Telegram after deployment

The Telegram publisher can now include imported Arbeitnow, Greenhouse, and
Lever jobs. A deployment by itself does **not** import or broadcast jobs.
For each source, first record the redistribution permission decision. The
Arbeitnow activation switch above enables its source after a successful import;
Greenhouse and Lever require a separate, approved operator activation. The
source must have a public, approved, active job. The existing Telegram
webhook/cron (or publisher CLI) then picks up eligible **new** unsent jobs;
`(job_id, channel_id)` prevents duplicate sends. Newly registered sources
remain disabled and unapproved, and demo/unapproved sources remain excluded.

For public API sources, the Telegram publisher requires an explicit publication
baseline: `job_sources.updated_at` must be later than `created_at`, and only
jobs with `first_seen_at` after that baseline are eligible. The one-time
Arbeitnow activation above sets this timestamp **after** the initial import,
so the initial batch is visible on the site but not broadcast to Telegram.
Subsequent scheduled imports do not move the baseline. If Greenhouse or Lever
is activated manually later, set `updated_at=now()` in the same source update
to establish its baseline; do not enable the source before this publisher
protection is deployed.

Telegram messages identify the source, link to the original vacancy, and link
to the JobHub detail page. Arbeitnow messages also link directly to
Arbeitnow.com as required by its API terms. Remote vacancies include a note to
check country and application restrictions at the original posting. Configure
the existing publisher channel and deployment settings as described in
[`telegram-publisher-webhook.md`](telegram-publisher-webhook.md). Neither the
importer nor deployment starts a persistent worker automatically. Verify the
publisher with a dry-run and inspect the first live message before relying on
the proposed schedule above.

**Schedule:** configured sources every 6 hours with a
one-shot CLI under GitHub Actions, with alerts
on failures and incomplete snapshots. Keep per-source request counts, revisit
terms and employer consent, and honor removal requests before enabling display.

The repository includes a six-hour GitHub Actions schedule in
`.github/workflows/public-job-imports.yml`; it is inert until the repository
variable `PUBLIC_IMPORT_ENABLED=true` is set on `main`. Configure these
GitHub Actions values only after inspecting the target Production database:

| Kind | Name | Purpose |
| --- | --- | --- |
| Secret | `PUBLIC_IMPORT_CONFIG_JSON` | JSON config above with `enabled=true` and only verified board IDs |
| Secret | `PUBLIC_IMPORT_DATABASE_URL` | Inspected Production Supabase PostgreSQL URL |
| Variable | `PUBLIC_IMPORT_DATABASE_HOST` | Exact hostname of that inspected URL |
| Variable | `PUBLIC_IMPORT_DATABASE_USER` | Exact username of that inspected Production URL |
| Variable | `PUBLIC_IMPORT_PRODUCTION_PERMISSIONS_CONFIRMED` | `true` after source rights review |
| Variable | `PUBLIC_IMPORT_ACTIVATE_ARBEITNOW` | `true` to show Arbeitnow jobs after a successful first import and establish the Telegram baseline |
| Variable | `PUBLIC_IMPORT_ENABLED` | `true` to activate scheduled/manual workflow runs |

The workflow runs `--once`, never a resident worker. With the optional
Arbeitnow activation switch, it enables site display only after a successful
import and sets the Telegram cutoff; it does not send Telegram messages itself.
GitHub Actions schedules may be delayed, so this polling interval is not a
guarantee of immediate publication after deployment.
