# Job visibility rules

JobHub has two job families with different public visibility rules: native
JobHub vacancies and imported external vacancies.

## Native vacancies

Native vacancies are public only when the owning company and vacancy state allow
display. Draft, paused, closed, deleted or expired vacancies must stay hidden
from public job search.

## Imported vacancies

Imported vacancies require:

- an enabled source;
- production permission when running in production mode;
- a public source URL;
- source freshness that has not expired;
- external application behavior.

## Shared expectations

List, count and detail queries should use the same visibility predicate. Do not
make a job visible in detail view if it would be hidden from the list.
