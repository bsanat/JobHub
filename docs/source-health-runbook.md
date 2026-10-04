# Source health runbook

Use source health checks to understand imported vacancy freshness without
changing public visibility by hand.

## Signals to inspect

- latest successful run per source;
- failed run count and sanitized error codes;
- fetched, normalized, inserted, updated and skipped counts;
- stale or hidden imported job counts;
- request budget used by the provider adapter.

## Safe response

- Failed runs should not refresh job `last_seen_at` or `fresh_until` values.
- Partial ranked searches should not be treated as complete snapshots.
- Missing jobs should be reconciled only when a provider supplies a complete,
  successful snapshot and the source policy allows it.

Keep public display decisions in shared backend predicates, not manual SQL edits.
